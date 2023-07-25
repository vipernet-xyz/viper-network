package keeper

import (
	"fmt"

	sdk "github.com/vipernet-xyz/viper-network/types"

	genesistypes "github.com/vipernet-xyz/viper-network/modules/apps/27-interchain-accounts/genesis/types"
	host "github.com/vipernet-xyz/viper-network/modules/core/24-host"
)

// InitGenesis initializes the interchain accounts controller application state from a provided genesis state
func InitGenesis(ctx sdk.Ctx, keeper Keeper, state genesistypes.ControllerGenesisState) {
	for _, portID := range state.Ports {
		if !keeper.IsBound(ctx, portID) {
			cap := keeper.BindPort(ctx, portID)
			if err := keeper.ClaimCapability(ctx, cap, host.PortPath(portID)); err != nil {
				panic(fmt.Sprintf("could not claim port capability: %v", err))
			}
		}
	}

	for _, ch := range state.ActiveChannels {
		keeper.SetActiveChannelID(ctx, ch.ConnectionId, ch.PortId, ch.ChannelId)

		if ch.IsMiddlewareEnabled {
			keeper.SetMiddlewareEnabled(ctx, ch.PortId, ch.ConnectionId)
		} else {
			keeper.SetMiddlewareDisabled(ctx, ch.PortId, ch.ConnectionId)
		}
	}

	for _, acc := range state.InterchainAccounts {
		keeper.SetInterchainAccountAddress(ctx, acc.ConnectionId, acc.PortId, acc.AccountAddress)
	}

	keeper.SetParams(ctx, state.Params)
}

// ExportGenesis returns the interchain accounts controller exported genesis
func ExportGenesis(ctx sdk.Ctx, keeper Keeper) genesistypes.ControllerGenesisState {
	return genesistypes.NewControllerGenesisState(
		keeper.GetAllActiveChannels(ctx),
		keeper.GetAllInterchainAccounts(ctx),
		keeper.GetAllPorts(ctx),
		keeper.GetParams(ctx),
	)
}
