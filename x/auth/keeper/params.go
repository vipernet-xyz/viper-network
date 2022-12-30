package keeper

import (
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/auth/types"
)

// SetParams sets the auth module's parameters.
func (k Keeper) SetParams(ctx sdk.Ctx, params types.Params) {
	k.subspace.SetParamSet(ctx, &params)
}

// GetParams gets the auth module's parameters.
func (k Keeper) GetParams(ctx sdk.Ctx) (params types.Params) {
	k.subspace.GetParamSet(ctx, &params)
	return
}
