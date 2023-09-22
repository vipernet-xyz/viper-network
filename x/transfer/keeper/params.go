package keeper

import (
	sdk "github.com/vipernet-xyz/viper-network/types"

	"github.com/vipernet-xyz/viper-network/x/transfer/types"
)

// GetSendEnabled retrieves the send enabled boolean from the paramstore
func (k Keeper) GetSendEnabled(ctx sdk.Ctx) bool {
	var res bool
	k.paramSpace.Get(ctx, types.KeySendEnabled, &res)
	return res
}

// GetReceiveEnabled retrieves the receive enabled boolean from the paramstore
func (k Keeper) GetReceiveEnabled(ctx sdk.Ctx) bool {
	var res bool
	k.paramSpace.Get(ctx, types.KeyReceiveEnabled, &res)
	return res
}

// GetParams returns the total set of ibc-transfer parameters.
func (k Keeper) GetParams(ctx sdk.Ctx) types.Params {
	return types.Params{
		SendEnabled:    k.GetSendEnabled(ctx),
		ReceiveEnabled: k.GetReceiveEnabled(ctx),
	}
}

// SetParams sets the total set of ibc-transfer parameters.
func (k Keeper) SetParams(ctx sdk.Ctx, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}
