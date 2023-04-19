package keeper

import (
	"fmt"

	sdk "github.com/vipernet-xyz/viper-network/types"
	governanceTypes "github.com/vipernet-xyz/viper-network/x/governance/types"
	providersTypes "github.com/vipernet-xyz/viper-network/x/providers/types"
	"github.com/vipernet-xyz/viper-network/x/servicers/types"
)

// RewardForRelays - Award coins to an address
func (k Keeper) RewardForRelays(ctx sdk.Ctx, relays sdk.BigInt, address sdk.Address, provider providersTypes.Provider) sdk.BigInt {
	if k.Cdc.IsAfterNonCustodialUpgrade(ctx.BlockHeight()) {
		var found bool
		address, found = k.GetValidatorOutputAddress(ctx, address)
		if !found {
			k.Logger(ctx).Error(fmt.Sprintf("no validator found for address %s; unable to mint the relay reward...", address.String()))
			return sdk.ZeroInt()
		}
	}

	var coins sdk.BigInt

	coins = k.TokenRewardFactor(ctx).Mul(relays)

	coins1 := relays.Quo((sdk.NewInt(k.providerKeeper.BaselineThroughputStakeRate(ctx)).Quo(sdk.NewInt(100))))

	toNode, toFeeCollector := k.NodeReward(ctx, coins)
	if toNode.IsPositive() {
		k.mint(ctx, toNode, address)
	}
	if toFeeCollector.IsPositive() {
		k.mint(ctx, toFeeCollector, k.getFeePool(ctx).GetAddress())
	}
	toProvider := k.ProviderReward(ctx, coins)
	if toProvider.IsPositive() {
		k.mint(ctx, toProvider, provider.Address)
	}
	p := k.providerKeeper.Provider(ctx, provider.Address)
	p1 := p.(providersTypes.Provider)
	if k.BurnActive(ctx) == true {
		k.burn(ctx, coins1, p1)
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
	// get the dao and proposer % ex DAO .1 or 10% Proposer .05 or 5%
	daoAllocation := sdk.NewDec(k.DAOAllocation(ctx))
	proposerAllocation := sdk.NewDec(k.ProposerAllocation(ctx))
	daoAndProposerAllocation := daoAllocation.Add(proposerAllocation)
	// get the new percentages based on the total. This is needed because the servicer (relayer) cut has already been allocated
	daoAllocation = daoAllocation.Quo(daoAndProposerAllocation)
	// dao cut calculation truncates int ex: 1.99uvipr = 1uvipr
	daoCut := feesCollected.ToDec().Mul(daoAllocation).TruncateInt()
	// proposer is whatever is left
	proposerCut := feesCollected.Sub(daoCut)
	// send to the two parties
	feeAddr := feesCollector.GetAddress()
	err := k.AccountKeeper.SendCoinsFromAccountToModule(ctx, feeAddr, governanceTypes.DAOAccountName, sdk.NewCoins(sdk.NewCoin(sdk.DefaultStakeDenom, daoCut)))
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

// "mint" - takes an amount and mints it to the servicer staking pool, then sends the coins to the address
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

func (k Keeper) burn(ctx sdk.Ctx, amount sdk.BigInt, provider providersTypes.Provider) (sdk.Result, sdk.Error) {
	coins := sdk.NewCoins(sdk.NewCoin(k.StakeDenom(ctx), amount))
	burnErr := k.AccountKeeper.BurnCoins(ctx, providersTypes.StakedPoolName, coins)
	if burnErr != nil {
		ctx.Logger().Error(fmt.Sprintf("unable to burn tokens, at height %d: ", ctx.BlockHeight()) + burnErr.Error())
		return burnErr.Result(), nil
	}
	// cannot decrease balance below zero
	tokensToBurn := sdk.MinInt(amount, provider.StakedTokens)
	tokensToBurn = sdk.MaxInt(tokensToBurn, sdk.ZeroInt()) // defensive.
	provider, err := provider.RemoveStakedTokens(tokensToBurn)
	if err != nil {
		return sdk.Result{}, sdk.ErrInternal(err.Error())
	}
	//reset provider relays
	provider.MaxRelays = k.providerKeeper.CalculateProviderRelays(ctx, provider)

	k.providerKeeper.SetProvider(ctx, provider)
	// if falls below minimum force burn all of the stake
	if provider.GetTokens().LT(sdk.NewInt(k.providerKeeper.MinimumStake(ctx))) {
		var err error
		if k.Cdc.IsAfterNonCustodialUpgrade(ctx.BlockHeight()) {
			err = k.providerKeeper.ForceProviderUnstake(ctx, provider)
		} else {
			err = k.providerKeeper.LegacyForceProviderUnstake(ctx, provider)
		}
		if err != nil {
			k.Logger(ctx).Error("could not force unstake: " + err.Error() + "\nfor provider " + provider.Address.String())
			return sdk.Result{}, nil
		}
	}
	logString := fmt.Sprintf("an amount of %s tokens was burned from %s", amount.String(), provider.Address.String())
	k.Logger(ctx).Info(logString)
	return sdk.Result{
		Log: logString,
	}, nil

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

// GetProviderKey - Retrieve the provider key
func (k Keeper) GetProvider(ctx sdk.Ctx) (addr sdk.Address) {
	store := ctx.KVStore(k.storeKey)
	b, _ := store.Get(providersTypes.AllProvidersKey)
	if b == nil {
		k.Logger(ctx).Error("Provider not set")
		return nil
		//os.Exit(1)
	}
	_ = k.Cdc.UnmarshalBinaryLengthPrefixed(b, &addr, ctx.BlockHeight())
	return addr

}

// SetProviderKey -  Store provider public key for this block
func (k Keeper) SetProviderKey(ctx sdk.Ctx, consAddr sdk.Address) {
	store := ctx.KVStore(k.storeKey)
	b, err := k.Cdc.MarshalBinaryLengthPrefixed(&consAddr, ctx.BlockHeight())
	if err != nil {
		panic(err)
	}
	_ = store.Set(providersTypes.AllProvidersKey, b)
}
