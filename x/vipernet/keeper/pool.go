package keeper

import (
	sdk "github.com/vipernet-xyz/viper-network/types"
)

// "StakeDenom" - Returns the stake coin denomination from the provider module
func (k Keeper) StakeDenom(ctx sdk.Ctx) (res string) {
	res = k.posKeeper.StakeDenom(ctx)
	return
}

// "GetPlatformStakedTokens" - Returns the total number of staked tokens in the platforms module
func (k Keeper) GetPlatformStakedTokens(ctx sdk.Ctx) (res sdk.BigInt) {
	res = k.platformKeeper.GetStakedTokens(ctx)
	return
}

// "GetNodeStakedTokens" - Returns the total number of staked tokens in the providers module
func (k Keeper) GetProvidersStakedTokens(ctx sdk.Ctx) (res sdk.BigInt) {
	res = k.posKeeper.GetStakedTokens(ctx)
	return
}

// "GetTotalTokens" - Returns the total number of tokens kept in any/all modules
func (k Keeper) GetTotalTokens(ctx sdk.Ctx) (res sdk.BigInt) {
	res = k.posKeeper.TotalTokens(ctx)
	return
}

// "GetTotalStakedTokens" - Returns the summation of platform staked tokens and provider staked tokens
func (k Keeper) GetTotalStakedTokens(ctx sdk.Ctx) (res sdk.BigInt) {
	res = k.GetProvidersStakedTokens(ctx).Add(k.GetPlatformStakedTokens(ctx))
	return
}
