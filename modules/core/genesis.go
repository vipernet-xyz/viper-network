package ibc

import (
	sdk "github.com/vipernet-xyz/viper-network/types"

	client "github.com/vipernet-xyz/viper-network/modules/core/02-client"
	connection "github.com/vipernet-xyz/viper-network/modules/core/03-connection"
	channel "github.com/vipernet-xyz/viper-network/modules/core/04-channel"
	"github.com/vipernet-xyz/viper-network/modules/core/keeper"
	"github.com/vipernet-xyz/viper-network/modules/core/types"
)

// InitGenesis initializes the ibc state from a provided genesis
// state.
func InitGenesis(ctx sdk.Ctx, k keeper.Keeper, gs *types.GenesisState) {
	client.InitGenesis(ctx, k.ClientKeeper, gs.ClientGenesis)
	connection.InitGenesis(ctx, k.ConnectionKeeper, gs.ConnectionGenesis)
	channel.InitGenesis(ctx, k.ChannelKeeper, gs.ChannelGenesis)
}

// ExportGenesis returns the ibc exported genesis.
func ExportGenesis(ctx sdk.Ctx, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		ClientGenesis:     client.ExportGenesis(ctx, k.ClientKeeper),
		ConnectionGenesis: connection.ExportGenesis(ctx, k.ConnectionKeeper),
		ChannelGenesis:    channel.ExportGenesis(ctx, k.ChannelKeeper),
	}
}
