package keeper

import (
	sdk "github.com/vipernet-xyz/viper-network/types"

	"github.com/vipernet-xyz/viper-network/modules/apps/27-interchain-accounts/controller/types"
)

// IsControllerEnabled retrieves the controller enabled boolean from the paramstore.
// True is returned if the controller submodule is enabled.
func (k Keeper) IsControllerEnabled(ctx sdk.Ctx) bool {
	var res bool
	k.paramSpace.Get(ctx, types.KeyControllerEnabled, &res)
	return res
}

// GetParams returns the total set of the controller submodule parameters.
func (k Keeper) GetParams(ctx sdk.Ctx) types.Params {
	return types.NewParams(k.IsControllerEnabled(ctx))
}

// SetParams sets the total set of the controller submodule parameters.
func (k Keeper) SetParams(ctx sdk.Ctx, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}
