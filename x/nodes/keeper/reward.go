package keeper

import (
	"fmt"

	"github.com/vipernet-xyz/viper-network/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	appsTypes "github.com/vipernet-xyz/viper-network/x/apps/types"
	govTypes "github.com/vipernet-xyz/viper-network/x/gov/types"
	"github.com/vipernet-xyz/viper-network/x/nodes/types"
)

// RewardForRelays - Award coins to an address (will be called at the beginning of the next block)
func (k Keeper) RewardForRelays(ctx sdk.Ctx, relays sdk.BigInt, address sdk.Address, appAddress sdk.Address) sdk.BigInt {
	if k.Cdc.IsAfterNonCustodialUpgrade(ctx.BlockHeight()) {
		var found bool
		address, found = k.GetValidatorOutputAddress(ctx, address)
		if !found {
			k.Logger(ctx).Error(fmt.Sprintf("no validator found for address %s; unable to mint the relay reward...", address.String()))
			return sdk.ZeroInt()
		}
	}

	var coins sdk.BigInt

	//check if it is enabled, if so scale the rewards
	if k.Cdc.IsAfterNamedFeatureActivationHeight(ctx.BlockHeight(), codec.RSCALKey) {
		//grab stake
		validator, found := k.GetValidator(ctx, address)
		if !found {
			ctx.Logger().Error(fmt.Errorf("no validator found for address %s; at height %d\n", address.String(), ctx.BlockHeight()).Error())
			return sdk.ZeroInt()
		}

		stake := validator.GetTokens()
		//floorstake to the lowest bin multiple or take ceiling, whicherver is smaller
		flooredStake := sdk.MinInt(stake.Sub(stake.Mod(k.MinServicerStakeBinWidth(ctx))), k.MaxServicerStakeBin(ctx).Sub(k.MaxServicerStakeBin(ctx).Mod(k.MinServicerStakeBinWidth(ctx))))
		//Convert from tokens to a BIN number
		bin := flooredStake.Quo(k.MinServicerStakeBinWidth(ctx))
		//calculate the weight value, weight will be a floatng point number so cast to DEC here and then truncate back to big int
		weight := bin.ToDec().FracPow(k.ServicerStakeBinExponent(ctx), ExponentDenominator).Quo(k.ServicerStakeWeight(ctx))
		coinsDecimal := k.TokenRewardFactor(ctx).ToDec().Mul(relays.ToDec()).Mul(weight)
		//truncate back to int
		coins = coinsDecimal.TruncateInt()
	} else {
		coins = k.TokenRewardFactor(ctx).Mul(relays)
	}

	toNode, toFeeCollector := k.NodeReward(ctx, coins)
	if toNode.IsPositive() {
		k.mint(ctx, toNode, address)
	}
	if toFeeCollector.IsPositive() {
		k.mint(ctx, toFeeCollector, k.getFeePool(ctx).GetAddress())
	}
	toApp := k.AppReward(ctx, coins)
	if toApp.IsPositive() {
		k.mint(ctx, toApp, appAddress)
	}
	return toNode
}

// blockReward - Handles distribution of the collected fees
func (k Keeper) blockReward(ctx sdk.Ctx, previousProposer sdk.Address) {
	feesCollector := k.getFeePool(ctx)
	feesCollected := feesCollector.GetCoins().AmountOf(sdk.DefaultStakeDenom)
	// check for zero fees
	if feesCollected.IsZero() {
		return
	}
	// get the dao and proposer % ex DAO .08 or 8% Proposer .01 or 1%
	daoAllocation := sdk.NewDec(k.DAOAllocation(ctx))
	proposerAllocation := sdk.NewDec(k.ProposerAllocation(ctx))
	appAllocation := sdk.NewDec(k.AppAllocation(ctx))
	daoProposerAndappAllocation := daoAllocation.Add(proposerAllocation).Add(appAllocation)
	// get the new percentages based on the total. This is needed because the node (relayer) cut has already been allocated
	daoAllocation = daoAllocation.Quo(daoProposerAndappAllocation)
	// dao cut calculation truncates int ex: 1.99uvip = 1uvip
	daoCut := feesCollected.ToDec().Mul(daoAllocation).TruncateInt()
	// get the new percentages based on the total. This is needed because the node (relayer) cut has already been allocated
	appAllocation = appAllocation.Quo(daoProposerAndappAllocation)
	// app cut calculation truncates int ex: 1.99uvip = 1uvip
	appCut := feesCollected.ToDec().Mul(appAllocation).TruncateInt()
	// proposer is whatever is left
	proposerCut := feesCollected.Sub(daoCut).Sub(appCut)
	// send to the two parties
	feeAddr := feesCollector.GetAddress()
	err := k.AccountKeeper.SendCoinsFromAccountToModule(ctx, feeAddr, govTypes.DAOAccountName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultStakeDenom, daoCut)))
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("unable to send %s cut of block reward to the dao: %s, at height %d", daoCut.String(), err.Error(), ctx.BlockHeight()))
	}
	if k.Cdc.IsAfterNonCustodialUpgrade(ctx.BlockHeight()) {
		outputAddress, found := k.GetValidatorOutputAddress(ctx, previousProposer)
		if !found {
			ctx.Logger().Error(fmt.Sprintf("unable to send %s cut of block reward to the proposer: %s, with error %s, at height %d", proposerCut.String(), previousProposer, types.ErrNoValidatorForAddress(types.ModuleName), ctx.BlockHeight()))
			return
		}
		err = k.AccountKeeper.SendCoins(ctx, feeAddr, outputAddress, sdk.NewCoins(sdk.NewCoin(sdk.DefaultStakeDenom, proposerCut)))
		if err != nil {
			ctx.Logger().Error(fmt.Sprintf("unable to send %s cut of block reward to the proposer: %s, with error %s, at height %d", proposerCut.String(), previousProposer, err.Error(), ctx.BlockHeight()))
		}
		return
	}
	err = k.AccountKeeper.SendCoins(ctx, feeAddr, previousProposer, sdk.NewCoins(sdk.NewCoin(sdk.DefaultStakeDenom, proposerCut)))
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("unable to send %s cut of block reward to the proposer: %s, with error %s, at height %d", proposerCut.String(), previousProposer, err.Error(), ctx.BlockHeight()))
	}
}

// "mint" - takes an amount and mints it to the node staking pool, then sends the coins to the address
func (k Keeper) mint(ctx sdk.Ctx, amount sdk.BigInt, address sdk.Address) sdk.Result {
	coins := sdk.NewCoins(sdk.NewCoin(k.StakeDenom(ctx), amount))
	mintErr := k.AccountKeeper.MintCoins(ctx, types.StakedPoolName, coins)
	if mintErr != nil {
		ctx.Logger().Error(fmt.Sprintf("unable to mint tokens, at height %d: ", ctx.BlockHeight()) + mintErr.Error())
		return mintErr.Result()
	}
	sendErr := k.AccountKeeper.SendCoinsFromModuleToAccount(ctx, types.StakedPoolName, address, coins)
	if sendErr != nil {
		ctx.Logger().Error(fmt.Sprintf("unable to send tokens, at height %d: ", ctx.BlockHeight()) + sendErr.Error())
		return sendErr.Result()
	}
	logString := fmt.Sprintf("a reward of %s was minted to %s", amount.String(), address.String())
	k.Logger(ctx).Info(logString)
	return sdk.Result{
		Log: logString,
	}
}

// MintRate = (total supply * inflation rate) / (30 day avg. of daily relays * 365 days)

// GetPreviousProposer - Retrieve the proposer public key for this block
func (k Keeper) GetPreviousProposer(ctx sdk.Ctx) (addr sdk.Address) {
	store := ctx.KVStore(k.storeKey)
	b, _ := store.Get(types.ProposerKey)
	if b == nil {
		k.Logger(ctx).Error("Previous proposer not set")
		return nil
		//os.Exit(1)
	}
	_ = k.Cdc.UnmarshalBinaryLengthPrefixed(b, &addr, ctx.BlockHeight())
	return addr

}

// SetPreviousProposer -  Store proposer public key for this block
func (k Keeper) SetPreviousProposer(ctx sdk.Ctx, consAddr sdk.Address) {
	store := ctx.KVStore(k.storeKey)
	b, err := k.Cdc.MarshalBinaryLengthPrefixed(&consAddr, ctx.BlockHeight())
	if err != nil {
		panic(err)
	}
	_ = store.Set(types.ProposerKey, b)
}

// GetApplicationKey - Retrieve the application key
func (k Keeper) GetApplication(ctx sdk.Ctx) (addr sdk.Address) {
	store := ctx.KVStore(k.storeKey)
	b, _ := store.Get(appsTypes.AllApplicationsKey)
	if b == nil {
		k.Logger(ctx).Error("Application not set")
		return nil
		//os.Exit(1)
	}
	_ = k.Cdc.UnmarshalBinaryLengthPrefixed(b, &addr, ctx.BlockHeight())
	return addr

}

// SetApplicationKey -  Store application public key for this block
func (k Keeper) SetApplicationKey(ctx sdk.Ctx, consAddr sdk.Address) {
	store := ctx.KVStore(k.storeKey)
	b, err := k.Cdc.MarshalBinaryLengthPrefixed(&consAddr, ctx.BlockHeight())
	if err != nil {
		panic(err)
	}
	_ = store.Set(appsTypes.AllApplicationsKey, b)
}
