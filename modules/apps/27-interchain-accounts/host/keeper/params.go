package keeper

import (
	sdk "github.com/vipernet-xyz/viper-network/types"

	"github.com/vipernet-xyz/viper-network/modules/apps/27-interchain-accounts/host/types"
)

// IsHostEnabled retrieves the host enabled boolean from the paramstore.
// True is returned if the host submodule is enabled.
func (k Keeper) IsHostEnabled(ctx sdk.Ctx) bool {
	var res bool
	k.paramSpace.Get(ctx, types.KeyHostEnabled, &res)
	return res
}

// GetAllowMessages retrieves the host enabled msg types from the paramstore
func (k Keeper) GetAllowMessages(ctx sdk.Ctx) []string {
	var res []string
	k.paramSpace.Get(ctx, types.KeyAllowMessages, &res)
	return res
}

// GetParams returns the total set of the host submodule parameters.
func (k Keeper) GetParams(ctx sdk.Ctx) types.Params {
	return types.NewParams(k.IsHostEnabled(ctx), k.GetAllowMessages(ctx))
}

// SetParams sets the total set of the host submodule parameters.
func (k Keeper) SetParams(ctx sdk.Ctx, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}
