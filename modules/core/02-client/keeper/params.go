package keeper

import (
	sdk "github.com/vipernet-xyz/viper-network/types"

	"github.com/vipernet-xyz/viper-network/modules/core/02-client/types"
)

// GetAllowedClients retrieves the allowed clients from the paramstore
func (k Keeper) GetAllowedClients(ctx sdk.Ctx) []string {
	var res []string
	k.paramSpace.Get(ctx, types.KeyAllowedClients, &res)
	return res
}

// GetParams returns the total set of ibc-client parameters.
func (k Keeper) GetParams(ctx sdk.Ctx) types.Params {
	return types.NewParams(k.GetAllowedClients(ctx)...)
}

// SetParams sets the total set of ibc-client parameters.
func (k Keeper) SetParams(ctx sdk.Ctx, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}
