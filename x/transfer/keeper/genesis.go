package keeper

import (
	"fmt"

	sdk "github.com/vipernet-xyz/viper-network/types"

	"github.com/vipernet-xyz/viper-network/x/transfer/types"
)

// InitGenesis initializes the ibc-transfer state and binds to PortID.
func (k Keeper) InitGenesis(ctx sdk.Ctx, state types.GenesisState) {
	if state.PortId == "" {
		state.PortId = types.PortID
	}
	k.SetPort(ctx, state.PortId)

	for _, trace := range state.DenomTraces {
		k.SetDenomTrace(ctx, trace)
	}
	// Only try to bind to port if it is not already bound, since we may already own
	// port capability from capability InitGenesis
	if !k.IsBound(ctx, state.PortId) {
		// transfer module binds to the transfer port on InitChain
		// and claims the returned capability
		k.Logger(ctx).Info("Initializing port", "portID", state.PortId)
		err := k.BindPort(ctx, state.PortId)
		if err != nil {
			panic(fmt.Sprintf("could not claim port capability: %v", err))
		}
	}
	k.SetParams(ctx, state.Params)
}

// ExportGenesis exports ibc-transfer module's portID and denom trace info into its genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Ctx) *types.GenesisState {
	return &types.GenesisState{
		PortId:      k.GetPort(ctx),
		DenomTraces: k.GetAllDenomTraces(ctx),
		Params:      k.GetParams(ctx),
	}
}
