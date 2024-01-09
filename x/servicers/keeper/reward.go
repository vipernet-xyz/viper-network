package keeper

import (
	"fmt"

	sdk "github.com/vipernet-xyz/viper-network/types"
	governanceTypes "github.com/vipernet-xyz/viper-network/x/governance/types"
	requestorsTypes "github.com/vipernet-xyz/viper-network/x/requestors/types"
	"github.com/vipernet-xyz/viper-network/x/servicers/types"
	viperTypes "github.com/vipernet-xyz/viper-network/x/viper-main/types"
)

func (k Keeper) RewardForRelays(ctx sdk.Ctx, reportCard viperTypes.MsgSubmitReportCard, relays sdk.BigInt, address sdk.Address, requestor requestorsTypes.Requestor) sdk.BigInt {
	_, found := k.GetValidator(ctx, address)
	if !found {
		ctx.Logger().Error(fmt.Errorf("no validator found for address %s; at height %d\n", address.String(), ctx.BlockHeight()).Error())
		return sdk.ZeroInt()
	}
	address, found = k.GetValidatorOutputAddress(ctx, address)
	if !found {
		k.Logger(ctx).Error(fmt.Sprintf("no validator found for address %s; unable to mint the relay reward...", address.String()))
		return sdk.ZeroInt()
	}

	latencyScore := reportCard.Report.LatencyScore
	availabilityScore := reportCard.Report.AvailabilityScore
	reliabilityScore := reportCard.Report.ReliabilityScore

	// Ensure scores are within the valid range
	latencyScore = sdk.MinDec(latencyScore, sdk.OneDec())
	availabilityScore = sdk.MinDec(availabilityScore, sdk.OneDec())
	reliabilityScore = sdk.MinDec(reliabilityScore, sdk.OneDec())

	// Calculate the weighted average of scores
	totalScore := latencyScore.Mul(k.LatencyScoreWeight(ctx)).Add(availabilityScore.Mul(k.AvailabilityScoreWeight(ctx))).Add(reliabilityScore.Mul(k.ReliabilityScoreWeight(ctx)))

	// Calculate the reward coins based on the total score and relays
	trf, _ := sdk.NewDecFromStr(k.TokenRewardFactor(ctx).String())
	r, _ := sdk.NewDecFromStr(relays.String())
	coins := trf.Mul(r).Mul(totalScore).RoundInt()

	// Validate requestor and mint rewards accordingly
	if !k.GovKeeper.HasDiscountKey(ctx, requestor.GetAddress()) {
		toNode, toFeeCollector := k.NodeReward01(ctx, coins)
		if toNode.IsPositive() {
			k.mint(ctx, toNode, address)
		}
		if toFeeCollector.IsPositive() {
			k.mint(ctx, toFeeCollector, k.getFeePool(ctx).GetAddress())
		}
		toRequestor := k.RequestorReward(ctx, coins)
		if toRequestor.IsPositive() {
			k.mint(ctx, toRequestor, k.GovKeeper.GetDAOAccount(ctx).GetAddress())
		}
		toFishermen := k.FishermenReward(ctx, coins)
		if toFishermen.IsPositive() {
			k.mint(ctx, toFishermen, reportCard.FishermanAddress)
		}
		maxFreeTierRelays := sdk.NewInt(k.RequestorKeeper.MaxFreeTierRelaysPerSession(ctx))
		if k.BurnActive(ctx) && relays.GT(maxFreeTierRelays) {
			k.burn(ctx, coins, requestor)
		}
	} else {
		toNode, toFeeCollector := k.NodeReward02(ctx, coins)
		if toNode.IsPositive() {
			k.mint(ctx, toNode, address)
		}
		if toFeeCollector.IsPositive() {
			k.mint(ctx, toFeeCollector, k.getFeePool(ctx).GetAddress())
		}
		toRequestor := k.RequestorReward(ctx, coins)
		if toRequestor.IsPositive() {
			k.mint(ctx, toRequestor, requestor.Address)
		}
		toFishermen := k.FishermenReward(ctx, coins)
		if toFishermen.IsPositive() {
			k.mint(ctx, toFishermen, reportCard.FishermanAddress)
		}
		maxFreeTierRelays := sdk.NewInt(k.RequestorKeeper.MaxFreeTierRelaysPerSession(ctx))

		if k.BurnActive(ctx) && relays.GT(maxFreeTierRelays) {
			k.burn(ctx, coins, requestor)
		}
		return toNode
	}

	return sdk.ZeroInt()
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
	outputAddress, found := k.GetValidatorOutputAddress(ctx, previousProposer)
	if !found {
		ctx.Logger().Error(fmt.Sprintf("unable to send %s cut of block reward to the proposer: %s, with error %s, at height %d", proposerCut.String(), previousProposer, types.ErrNoValidatorForAddress(types.ModuleName), ctx.BlockHeight()))
		return
	}
	err = k.AccountKeeper.SendCoins(ctx, feeAddr, outputAddress, sdk.NewCoins(sdk.NewCoin(sdk.DefaultStakeDenom, proposerCut)))
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

// MintRate = (total supply * inflation rate) / (30 day avg. of daily relays * 365 days)

// "burn" - takes an amount and burns it
func (k Keeper) burn(ctx sdk.Ctx, amount sdk.BigInt, requestor requestorsTypes.Requestor) (sdk.Result, error) {
	// Burn coins from requestor account
	r, burnErr := requestor.RemoveStakedTokens(amount)
	if burnErr != nil {
		ctx.Logger().Error(fmt.Sprintf("unable to burn tokens, at height %d: %s", ctx.BlockHeight(), burnErr.Error()))
		return sdk.Result{}, burnErr
	}
	// Reset requestor relays
	r.MaxRelays = k.RequestorKeeper.CalculateRequestorRelays(ctx, requestor)

	// Update requestor in the store
	k.RequestorKeeper.SetRequestor(ctx, r)

	// If falls below minimum, force unstake
	if requestor.GetTokens().LT(sdk.NewInt(k.RequestorKeeper.MinimumStake(ctx))) {
		if err := k.RequestorKeeper.ForceRequestorUnstake(ctx, requestor); err != nil {
			logString := fmt.Sprintf("could not force unstake: %s for requestor %s", err.Error(), requestor.Address.String())
			k.Logger(ctx).Error(logString)
			return sdk.Result{}, sdk.ErrInternal(logString)
		}
	}

	// Log the amount of tokens burned and requestor's address
	logString := fmt.Sprintf("an amount of %s tokens was burned from %s", amount.String(), requestor.Address.String())
	k.Logger(ctx).Info(logString)

	return sdk.Result{
		Log: logString,
	}, nil
}

// GetPreviousProposer - Retrieve the proposer public key for this block
func (k Keeper) GetPreviousProposer(ctx sdk.Ctx) (addr sdk.Address) {
	store := ctx.KVStore(k.storeKey)
	b, _ := store.Get(types.ProposerKey)
	if b == nil {
		k.Logger(ctx).Error("Previous proposer not set")
		return nil
		//os.Exit(1)
	}
	_ = k.Cdc.UnmarshalBinaryLengthPrefixed(b, &addr)
	return addr

}

// SetPreviousProposer -  Store proposer public key for this block
func (k Keeper) SetPreviousProposer(ctx sdk.Ctx, consAddr sdk.Address) {
	store := ctx.KVStore(k.storeKey)
	b, err := k.Cdc.MarshalBinaryLengthPrefixed(&consAddr)
	if err != nil {
		panic(err)
	}
	_ = store.Set(types.ProposerKey, b)
}

// GetRequestorKey - Retrieve the requestor key
func (k Keeper) GetRequestor(ctx sdk.Ctx) (addr sdk.Address) {
	store := ctx.KVStore(k.storeKey)
	b, _ := store.Get(requestorsTypes.AllRequestorsKey)
	if b == nil {
		k.Logger(ctx).Error("Requestor not set")
		return nil
		//os.Exit(1)
	}
	_ = k.Cdc.UnmarshalBinaryLengthPrefixed(b, &addr)
	return addr

}

// SetRequestorKey -  Store requestor public key for this block
func (k Keeper) SetRequestorKey(ctx sdk.Ctx, consAddr sdk.Address) {
	store := ctx.KVStore(k.storeKey)
	b, err := k.Cdc.MarshalBinaryLengthPrefixed(&consAddr)
	if err != nil {
		panic(err)
	}
	_ = store.Set(requestorsTypes.AllRequestorsKey, b)
}
