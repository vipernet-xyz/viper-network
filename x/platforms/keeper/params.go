package keeper

import (
	"time"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/platforms/types"
)

// Default parameter namespace
const (
	DefaultParamspace = types.ModuleName
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

// BaselineThroughputStakeRate - Retrieve base relays per VIPR
func (k Keeper) BaselineThroughputStakeRate(ctx sdk.Ctx) (base int64) {
	k.Paramstore.Get(ctx, types.BaseRelaysPerVIPR, &base)
	return
}

// ParticipationRate - Retrieve participation rate
func (k Keeper) ParticipationRate(ctx sdk.Ctx) (isOn bool) {
	k.Paramstore.Get(ctx, types.ParticipationRate, &isOn)
	return
}

// StakingAdjustment - Retrieve stability adjustment
func (k Keeper) StakingAdjustment(ctx sdk.Ctx) (adjustment int64) {
	k.Paramstore.Get(ctx, types.StabilityModulation, &adjustment)
	return
}

// MaxPlatforms - Retrieve maximum number of platforms
func (k Keeper) MaxPlatforms(ctx sdk.Ctx) (res int64) {
	k.Paramstore.Get(ctx, types.KeyMaxPlatforms, &res)
	return
}

// MinimumStake - Retrieve minimum stake
func (k Keeper) MinimumStake(ctx sdk.Ctx) (res int64) {
	k.Paramstore.Get(ctx, types.KeyMinPlatformStake, &res)
	return
}

// MaxChains - Retrieve maximum chains
func (k Keeper) MaxChains(ctx sdk.Ctx) (res int64) {
	k.Paramstore.Get(ctx, types.KeyMaximumChains, &res)
	return
}

// Get all parameteras as types.Params
func (k Keeper) GetParams(ctx sdk.Ctx) types.Params {
	return types.Params{
		UnstakingTime:       k.UnStakingTime(ctx),
		MaxPlatforms:        k.MaxPlatforms(ctx),
		MinPlatformStake:    k.MinimumStake(ctx),
		BaseRelaysPerVIPR:   k.BaselineThroughputStakeRate(ctx),
		ParticipationRate:   k.ParticipationRate(ctx),
		StabilityModulation: k.StakingAdjustment(ctx),
		MaxChains:           k.MaxChains(ctx),
	}
}

// SetParams - Platformly set of params
func (k Keeper) SetParams(ctx sdk.Ctx, params types.Params) {
	k.Paramstore.SetParamSet(ctx, &params)
}
