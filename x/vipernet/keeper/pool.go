package keeper

import (
	sdk "github.com/vipernet-xyz/viper-network/types"
)

// "StakeDenom" - Returns the stake coin denomination from the servicer module
func (k Keeper) StakeDenom(ctx sdk.Ctx) (res string) {
	res = k.posKeeper.StakeDenom(ctx)
	return
}

// "GetProviderStakedTokens" - Returns the total number of staked tokens in the providers module
func (k Keeper) GetProviderStakedTokens(ctx sdk.Ctx) (res sdk.BigInt) {
	res = k.providerKeeper.GetStakedTokens(ctx)
	return
}

// "GetNodeStakedTokens" - Returns the total number of staked tokens in the servicers module
func (k Keeper) GetServicersStakedTokens(ctx sdk.Ctx) (res sdk.BigInt) {
	res = k.posKeeper.GetStakedTokens(ctx)
	return
}

// "GetTotalTokens" - Returns the total number of tokens kept in any/all modules
func (k Keeper) GetTotalTokens(ctx sdk.Ctx) (res sdk.BigInt) {
	res = k.posKeeper.TotalTokens(ctx)
	return
}

// "GetTotalStakedTokens" - Returns the summation of provider staked tokens and servicer staked tokens
func (k Keeper) GetTotalStakedTokens(ctx sdk.Ctx) (res sdk.BigInt) {
	res = k.GetServicersStakedTokens(ctx).Add(k.GetProviderStakedTokens(ctx))
	return
}
