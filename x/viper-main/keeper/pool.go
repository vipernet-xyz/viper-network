package keeper

import (
	sdk "github.com/vipernet-xyz/viper-network/types"
)

// "StakeDenom" - Returns the stake coin denomination from the servicer module
func (k Keeper) StakeDenom(ctx sdk.Ctx) (res string) {
	res = k.posKeeper.StakeDenom(ctx)
	return
}

// "GetRequestorStakedTokens" - Returns the total number of staked tokens in the requestors module
func (k Keeper) GetRequestorStakedTokens(ctx sdk.Ctx) (res sdk.BigInt) {
	res = k.requestorKeeper.GetStakedTokens(ctx)
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

// "GetTotalStakedTokens" - Returns the summation of requestor staked tokens and servicer staked tokens
func (k Keeper) GetTotalStakedTokens(ctx sdk.Ctx) (res sdk.BigInt) {
	res = k.GetServicersStakedTokens(ctx).Add(k.GetRequestorStakedTokens(ctx))
	return
}
