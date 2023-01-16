package keeper

import (
	"time"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/providers/types"
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

// TokenRewardFactor - Retrieve relay token multipler
func (k Keeper) TokenRewardFactor(ctx sdk.Ctx) sdk.BigInt {
	var multiplier int64
	k.Paramstore.Get(ctx, types.KeyTokenRewardFactor, &multiplier)
	return sdk.NewInt(multiplier)
}

// MinServicerStakeBinWidth - Retrieve MinServicerStakeBinWidth
func (k Keeper) MinServicerStakeBinWidth(ctx sdk.Ctx) sdk.BigInt {
	var multiplier int64
	k.Paramstore.Get(ctx, types.KeyMinServicerStakeBinWidth, &multiplier)
	return sdk.NewInt(multiplier)
}

// ServicerStakeWeight - Retrieve ServicerStakeWeight
func (k Keeper) ServicerStakeWeight(ctx sdk.Ctx) (res sdk.BigDec) {
	k.Paramstore.Get(ctx, types.KeyServicerStakeWeight, &res)
	return
}

// MaxServicerStakeBin - Retrieve MaxServicerStakeBin
func (k Keeper) MaxServicerStakeBin(ctx sdk.Ctx) sdk.BigInt {
	var multiplier int64
	k.Paramstore.Get(ctx, types.KeyMaxServicerStakeBin, &multiplier)
	return sdk.NewInt(multiplier)
}

// ServicerStakeBinExponent - Retrieve ServicerStakeBinExponent
func (k Keeper) ServicerStakeBinExponent(ctx sdk.Ctx) (res sdk.BigDec) {
	k.Paramstore.Get(ctx, types.KeyServicerStakeBinExponent, &res)
	return
}

func (k Keeper) NodeReward(ctx sdk.Ctx, reward sdk.BigInt) (providerReward sdk.BigInt, feesCollected sdk.BigInt) {
	// convert reward to dec
	r := reward.ToDec()
	// get the dao and proposer % ex DAO .08 or 8% Proposer .01 or 1%  App .02 or 2%
	daoAllocationPercentage := sdk.NewDec(k.DAOAllocation(ctx)).QuoInt64(int64(100))           // dec percentage
	proposerAllocationPercentage := sdk.NewDec(k.ProposerAllocation(ctx)).QuoInt64(int64(100)) // dec percentage
	platformAllocationPercentage := sdk.NewDec(k.PlatformAllocation(ctx)).QuoInt64(int64(100)) // dec percentage
	// the dao and proposer allocations go to the fee collector
	daoAllocation := r.Mul(daoAllocationPercentage)
	proposerAllocation := r.Mul(proposerAllocationPercentage)
	// truncate int ex 1.99 uvipr goes to 1 uvipr
	feesCollected = daoAllocation.Add(proposerAllocation).TruncateInt()
	//platformAllocation go to the platform
	platformAllocation := r.Mul(platformAllocationPercentage).TruncateInt()
	// the rest goes to the provider
	providerReward = reward.Sub(feesCollected).Sub(platformAllocation)
	return
}

func (k Keeper) PlatformReward(ctx sdk.Ctx, reward sdk.BigInt) (platformReward sdk.BigInt) {
	// convert reward to dec
	r := reward.ToDec()
	platformAllocationPercentage := sdk.NewDec(k.PlatformAllocation(ctx)).QuoInt64(int64(100)) // dec percentage
	platformAllocation := r.Mul(platformAllocationPercentage).TruncateInt()
	platformReward = platformAllocation
	return
}

// DAOAllocation - Retrieve DAO allocation
func (k Keeper) DAOAllocation(ctx sdk.Ctx) (res int64) {
	k.Paramstore.Get(ctx, types.KeyDAOAllocation, &res)
	return
}

// PlatformAllocation - Retrieve Platform Allocation
func (k Keeper) PlatformAllocation(ctx sdk.Ctx) (res int64) {
	k.Paramstore.Get(ctx, types.KeyPlatformAllocation, &res)
	return
}

// BlocksPerSession - Retrieve blocks per session
func (k Keeper) BlocksPerSession(ctx sdk.Ctx) (res int64) {
	k.Paramstore.Get(ctx, types.KeySessionBlock, &res)
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

// GetParams - Retrieve all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Ctx) types.Params {
	return types.Params{
		TokenRewardFactor:        k.TokenRewardFactor(ctx).Int64(),
		UnstakingTime:            k.UnStakingTime(ctx),
		MaxValidators:            k.MaxValidators(ctx),
		StakeDenom:               k.StakeDenom(ctx),
		StakeMinimum:             k.MinimumStake(ctx),
		SessionBlockFrequency:    k.BlocksPerSession(ctx),
		DAOAllocation:            k.DAOAllocation(ctx),
		PlatformAllocation:       k.PlatformAllocation(ctx),
		ProposerAllocation:       k.ProposerAllocation(ctx),
		MaximumChains:            k.MaxChains(ctx),
		MaxJailedBlocks:          k.MaxJailedBlocks(ctx),
		MaxEvidenceAge:           k.MaxEvidenceAge(ctx),
		SignedBlocksWindow:       k.SignedBlocksWindow(ctx),
		MinSignedPerWindow:       sdk.NewDec(k.MinSignedPerWindow(ctx)),
		DowntimeJailDuration:     k.DowntimeJailDuration(ctx),
		SlashFractionDoubleSign:  k.SlashFractionDoubleSign(ctx),
		SlashFractionDowntime:    k.SlashFractionDowntime(ctx),
		MinServicerStakeBinWidth: k.MinServicerStakeBinWidth(ctx).Int64(),
		ServicerStakeWeight:      k.ServicerStakeWeight(ctx),
		MaxServicerStakeBin:      k.MaxServicerStakeBin(ctx).Int64(),
		ServicerStakeBinExponent: k.ServicerStakeBinExponent(ctx),
	}
}

// SetParams - Apply set of params
func (k Keeper) SetParams(ctx sdk.Ctx, params types.Params) {
	k.Paramstore.SetParamSet(ctx, &params)
}
