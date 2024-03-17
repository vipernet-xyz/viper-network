package keeper

import (
	"time"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/servicers/types"
)

// Default parameter namespace
const (
	DefaultParamspace = types.ModuleName
	//ExponentDenominator This is used as an input to the decimal power function used for
	//By calculating the exponent, it avoids any overflows when taking the CthRoot of A by ensuring
	//that the exponient is always devisable by 100 giving the effective range of
	//ServicerStakeBinExponent 0-1 in steps of 0.01.
	ExponentDenominator = 100
)

// ParamKeyTable for staking module
func ParamKeyTable() sdk.KeyTable {
	return sdk.NewKeyTable().RegisterParamSet(&types.Params{})
}

// UnStakingTime - Retrieve unstaking time param
func (k Keeper) UnStakingTime(ctx sdk.Ctx) (res time.Duration) {
	k.Paramstore.Get(ctx, types.KeyUnstakingTime, &res)
	return
}

// MaxValidators - Retrieve maximum number of validators
func (k Keeper) MaxValidators(ctx sdk.Ctx) (res int64) {
	k.Paramstore.Get(ctx, types.KeyMaxValidators, &res)
	return
}

// StakeDenom - Bondable coin denomination
func (k Keeper) StakeDenom(ctx sdk.Ctx) (res string) {
	k.Paramstore.Get(ctx, types.KeyStakeDenom, &res)
	return
}

// MinimumStake - Retrieve Minimum stake
func (k Keeper) MinimumStake(ctx sdk.Ctx) (res int64) {
	k.Paramstore.Get(ctx, types.KeyStakeMinimum, &res)
	return
}

// ProposerAllocation - Retrieve proposer allocation
func (k Keeper) ProposerAllocation(ctx sdk.Ctx) (res int64) {
	k.Paramstore.Get(ctx, types.KeyProposerAllocation, &res)
	return
}

func (k Keeper) FishermenAllocation(ctx sdk.Ctx) (res int64) {
	k.Paramstore.Get(ctx, types.KeyFishermenAllocation, &res)
	return
}

// MaxEvidenceAge - Max age for evidence
func (k Keeper) MaxEvidenceAge(ctx sdk.Ctx) (res time.Duration) {
	k.Paramstore.Get(ctx, types.KeyMaxEvidenceAge, &res)
	return
}

// SignedBlocksWindow - Sliding window for downtime slashing
func (k Keeper) SignedBlocksWindow(ctx sdk.Ctx) (res int64) {
	k.Paramstore.Get(ctx, types.KeySignedBlocksWindow, &res)
	return
}

// MinSignedPerWindow - Downtime slashing threshold
func (k Keeper) MinSignedPerWindow(ctx sdk.Ctx) (res int64) {
	var minSignedPerWindow sdk.BigDec
	k.Paramstore.Get(ctx, types.KeyMinSignedPerWindow, &minSignedPerWindow)
	signedBlocksWindow := k.SignedBlocksWindow(ctx)

	// NOTE: RoundInt64 will never panic as minSignedPerWindow is
	//       less than 1.
	return minSignedPerWindow.MulInt64(signedBlocksWindow).RoundInt64() // todo may have to be int64 .RoundInt64()
}

// DowntimeJailDuration - Downtime jail duration
func (k Keeper) DowntimeJailDuration(ctx sdk.Ctx) (res time.Duration) {
	k.Paramstore.Get(ctx, types.KeyDowntimeJailDuration, &res)
	return
}

func (k Keeper) MinPauseTime(ctx sdk.Ctx) (res time.Duration) {
	k.Paramstore.Get(ctx, types.KeyMinPauseTime, &res)
	return
}

// SlashFractionDoubleSign - Retrieve slash fraction for double signature
func (k Keeper) SlashFractionDoubleSign(ctx sdk.Ctx) (res sdk.BigDec) {
	k.Paramstore.Get(ctx, types.KeySlashFractionDoubleSign, &res)
	return
}

// SlashFractionDowntime - Retrieve slash fraction time
func (k Keeper) SlashFractionDowntime(ctx sdk.Ctx) (res sdk.BigDec) {
	k.Paramstore.Get(ctx, types.KeySlashFractionDowntime, &res)
	return
}

// SlashFractionDowntime - Retrieve slash fraction time
func (k Keeper) SlashFractionNoActivity(ctx sdk.Ctx) (res sdk.BigDec) {
	k.Paramstore.Get(ctx, types.KeySlashFractionNoActivity, &res)
	return
}

// TokenRewardFactor - Retrieve relay token multipler
func (k Keeper) TokenRewardFactor(ctx sdk.Ctx) sdk.BigInt {
	var multiplier int64
	k.Paramstore.Get(ctx, types.KeyTokenRewardFactor, &multiplier)
	return sdk.NewInt(multiplier)
}

func (k Keeper) NodeReward01(ctx sdk.Ctx, reward sdk.BigInt) (servicerReward sdk.BigInt, feesCollected sdk.BigInt) {
	// convert reward to dec
	r := reward.ToDec()
	// get the dao, proposer, and fishermen % ex DAO .05 or 5% Proposer .01 or 1%  App .05 or 5% Fishermen .02 or 2%
	daoAllocationPercentage := sdk.NewDec(k.DAOAllocation(ctx)).QuoInt64(int64(100))             // dec percentage
	proposerAllocationPercentage := sdk.NewDec(k.ProposerAllocation(ctx)).QuoInt64(int64(100))   // dec percentage
	requestorAllocationPercentage := sdk.NewDec(k.RequestorAllocation(ctx)).QuoInt64(int64(100)) // dec percentage
	fishermenAllocationPercentage := sdk.NewDec(k.FishermenAllocation(ctx)).QuoInt64(int64(100))
	// the dao, proposer, and fishermen allocations go to the fee collector
	daoAllocation := r.Mul(daoAllocationPercentage.Add(requestorAllocationPercentage))
	proposerAllocation := r.Mul(proposerAllocationPercentage)
	fishermenAllocation := r.Mul(fishermenAllocationPercentage).TruncateInt()
	// truncate int ex 1.99 uvipr goes to 1 uvipr
	feesCollected = daoAllocation.Add(proposerAllocation).TruncateInt()
	// the rest goes to the servicer
	servicerReward = reward.Sub(feesCollected).Sub(fishermenAllocation)
	return
}

func (k Keeper) NodeReward02(ctx sdk.Ctx, reward sdk.BigInt) (servicerReward sdk.BigInt, feesCollected sdk.BigInt) {
	// convert reward to dec
	r := reward.ToDec()
	// get the dao, proposer, and fishermen % ex DAO .08 or 8% Proposer .01 or 1%  App .02 or 2% Fishermen .01 or 1%
	daoAllocationPercentage := sdk.NewDec(k.DAOAllocation(ctx)).QuoInt64(int64(100))             // dec percentage
	proposerAllocationPercentage := sdk.NewDec(k.ProposerAllocation(ctx)).QuoInt64(int64(100))   // dec percentage
	requestorAllocationPercentage := sdk.NewDec(k.RequestorAllocation(ctx)).QuoInt64(int64(100)) // dec percentage
	fishermenAllocationPercentage := sdk.NewDec(k.FishermenAllocation(ctx)).QuoInt64(int64(100))
	// the dao, proposer, and fishermen allocations go to the fee collector
	daoAllocation := r.Mul(daoAllocationPercentage)
	proposerAllocation := r.Mul(proposerAllocationPercentage)
	fishermenAllocation := r.Mul(fishermenAllocationPercentage)
	// truncate int ex 1.99 uvipr goes to 1 uvipr
	feesCollected = daoAllocation.Add(proposerAllocation).TruncateInt()
	//requestorAllocation go to the requestor
	requestorAllocation := r.Mul(requestorAllocationPercentage)
	ProvAndFish := requestorAllocation.Add(fishermenAllocation).TruncateInt()
	// the rest goes to the servicer
	servicerReward = reward.Sub(feesCollected).Sub(ProvAndFish)
	return
}

func (k Keeper) RequestorReward(ctx sdk.Ctx, reward sdk.BigInt) (requestorReward sdk.BigInt) {
	// convert reward to dec
	r := reward.ToDec()
	requestorAllocationPercentage := sdk.NewDec(k.RequestorAllocation(ctx)).QuoInt64(int64(100)) // dec percentage
	requestorAllocation := r.Mul(requestorAllocationPercentage)
	requestorReward = requestorAllocation.TruncateInt()
	return
}

func (k Keeper) FishermenReward(ctx sdk.Ctx, reward sdk.BigInt) (fishermenReward sdk.BigInt) {
	// convert reward to dec
	r := reward.ToDec()
	fishermenAllocationPercentage := sdk.NewDec(k.FishermenAllocation(ctx)).QuoInt64(int64(100)) // dec percentage
	fishermenAllocation := r.Mul(fishermenAllocationPercentage)
	// Convert the decimal result to BigInt
	fishermenReward = fishermenAllocation.TruncateInt()
	return
}

// DAOAllocation - Retrieve DAO allocation
func (k Keeper) DAOAllocation(ctx sdk.Ctx) (res int64) {
	k.Paramstore.Get(ctx, types.KeyDAOAllocation, &res)
	return
}

// RequestorAllocation - Retrieve Requestor Allocation
func (k Keeper) RequestorAllocation(ctx sdk.Ctx) (res int64) {
	k.Paramstore.Get(ctx, types.KeyRequestorAllocation, &res)
	return
}

// BlocksPerSession - Retrieve blocks per session
func (k Keeper) BlocksPerSession(ctx sdk.Ctx) (res int64) {
	k.Paramstore.Get(ctx, types.KeySessionBlock, &res)
	return
}

func (k Keeper) RelaysToTokensChainMultiplierMap(ctx sdk.Ctx) (res map[string]int64) {
	k.Paramstore.Get(ctx, types.KeyRelaysToTokensChainMultiplierMap, &res)
	if res == nil {
		res = types.DefaultRelaysToTokensChainMultiplierMap
	}
	return
}

func (k Keeper) RelaysToTokensGeoZoneMultiplierMap(ctx sdk.Ctx) (res map[string]int64) {
	k.Paramstore.Get(ctx, types.KeyRelaysToTokensGeoZoneMultiplierMap, &res)
	if res == nil {
		res = types.DefaultRelaysToTokensGeoZoneMultiplierMap
	}
	return
}

func (k Keeper) MaxChains(ctx sdk.Ctx) (res int64) {
	k.Paramstore.Get(ctx, types.KeyMaxChains, &res)
	return
}
func (k Keeper) MaxJailedBlocks(ctx sdk.Ctx) (res int64) {
	k.Paramstore.Get(ctx, types.KeyMaxJailedBlocks, &res)
	return
}

func (k Keeper) ServicerCountLock(ctx sdk.Ctx) (isOn bool) {
	k.Paramstore.Get(ctx, types.ServicerCountLock, &isOn)
	return
}

func (k Keeper) BurnActive(ctx sdk.Ctx) (isOn bool) {
	k.Paramstore.Get(ctx, types.BurnActive, &isOn)
	return
}

func (k Keeper) MaxFishermen(ctx sdk.Ctx) (res int64) {
	k.Paramstore.Get(ctx, types.KeyMaxFishermen, &res)
	return
}

func (k Keeper) FishermenCount(ctx sdk.Ctx) (res int64) {
	k.Paramstore.Get(ctx, types.KeyFishermenCount, &res)
	return
}

func (k Keeper) LatencyScoreWeight(ctx sdk.Ctx) (res sdk.BigDec) {
	k.Paramstore.Get(ctx, types.KeyLatencyScoreWeight, &res)
	return
}

func (k Keeper) AvailabilityScoreWeight(ctx sdk.Ctx) (res sdk.BigDec) {
	k.Paramstore.Get(ctx, types.KeyAvailabilityScoreWeight, &res)
	return
}

func (k Keeper) ReliabilityScoreWeight(ctx sdk.Ctx) (res sdk.BigDec) {
	k.Paramstore.Get(ctx, types.KeyReliabilityScoreWeight, &res)
	return
}

func (k Keeper) SlashFractionFisherman(ctx sdk.Ctx) (res sdk.BigDec) {
	k.Paramstore.Get(ctx, types.KeySlashFractionFisherman, &res)
	return
}

func (k Keeper) MaxFreeTierRelaysPerSession(ctx sdk.Ctx) (res int64) {
	k.Paramstore.Get(ctx, types.KeyMaxFreeTierRelaysPerSession, &res)
	return
}

// GetParams - Retrieve all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Ctx) types.Params {
	return types.Params{
		TokenRewardFactor:                  k.TokenRewardFactor(ctx).Int64(),
		UnstakingTime:                      k.UnStakingTime(ctx),
		MaxValidators:                      k.MaxValidators(ctx),
		StakeDenom:                         k.StakeDenom(ctx),
		StakeMinimum:                       k.MinimumStake(ctx),
		SessionBlockFrequency:              k.BlocksPerSession(ctx),
		DAOAllocation:                      k.DAOAllocation(ctx),
		RequestorAllocation:                k.RequestorAllocation(ctx),
		ProposerAllocation:                 k.ProposerAllocation(ctx),
		FishermenAllocation:                k.FishermenAllocation(ctx),
		MaximumChains:                      k.MaxChains(ctx),
		MaxJailedBlocks:                    k.MaxJailedBlocks(ctx),
		MaxEvidenceAge:                     k.MaxEvidenceAge(ctx),
		SignedBlocksWindow:                 k.SignedBlocksWindow(ctx),
		MinSignedPerWindow:                 sdk.NewDec(k.MinSignedPerWindow(ctx)),
		DowntimeJailDuration:               k.DowntimeJailDuration(ctx),
		SlashFractionDoubleSign:            k.SlashFractionDoubleSign(ctx),
		SlashFractionDowntime:              k.SlashFractionDowntime(ctx),
		ServicerCountLock:                  k.ServicerCountLock(ctx),
		BurnActive:                         k.BurnActive(ctx),
		MinPauseTime:                       k.MinPauseTime(ctx),
		MaxFishermen:                       k.MaxFishermen(ctx),
		FishermenCount:                     k.FishermenCount(ctx),
		SlashFractionNoActivity:            k.SlashFractionNoActivity(ctx),
		LatencyScoreWeight:                 k.LatencyScoreWeight(ctx),
		AvailabilityScoreWeight:            k.AvailabilityScoreWeight(ctx),
		ReliabilityScoreWeight:             k.ReliabilityScoreWeight(ctx),
		SlashFractionFisherman:             k.SlashFractionFisherman(ctx),
		MaxFreeTierRelaysPerSession:        k.MaxFreeTierRelaysPerSession(ctx),
		RelaysToTokensChainMultiplierMap:   k.RelaysToTokensChainMultiplierMap(ctx),
		RelaysToTokensGeoZoneMultiplierMap: k.RelaysToTokensGeoZoneMultiplierMap(ctx),
	}
}

// SetParams - Apply set of params
func (k Keeper) SetParams(ctx sdk.Ctx, params types.Params) {
	k.Paramstore.SetParamSet(ctx, &params)
}

// UnbondingTime - The time duration for unbonding
func (k Keeper) UnbondingTime(ctx sdk.Ctx) time.Duration {
	return k.GetParams(ctx).UnstakingTime
}
