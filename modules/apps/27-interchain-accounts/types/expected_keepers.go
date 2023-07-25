package types

import (
	sdk "github.com/vipernet-xyz/viper-network/types"
	authexported "github.com/vipernet-xyz/viper-network/x/authentication/exported"
	capabilitytypes "github.com/vipernet-xyz/viper-network/x/capability/types"

	channeltypes "github.com/vipernet-xyz/viper-network/modules/core/04-channel/types"
	ibcexported "github.com/vipernet-xyz/viper-network/modules/core/exported"
)

// AccountKeeper defines the expected account keeper
type AccountKeeper interface {
	NewAccount(ctx sdk.Ctx, acc authexported.Account) authexported.Account
	GetAccount(ctx sdk.Ctx, addr sdk.Address) authexported.Account
	SetAccount(ctx sdk.Ctx, acc authexported.Account)
	GetModuleAccount(ctx sdk.Ctx, name string) authexported.ModuleAccountI
	GetModuleAddress(name string) sdk.Address
}

// ChannelKeeper defines the expected IBC channel keeper
type ChannelKeeper interface {
	GetChannel(ctx sdk.Ctx, srcPort, srcChan string) (channel channeltypes.Channel, found bool)
	GetNextSequenceSend(ctx sdk.Ctx, portID, channelID string) (uint64, bool)
	GetConnection(ctx sdk.Ctx, connectionID string) (ibcexported.ConnectionI, error)
	GetAllChannelsWithPortPrefix(ctx sdk.Ctx, portPrefix string) []channeltypes.IdentifiedChannel
}

// PortKeeper defines the expected IBC port keeper
type PortKeeper interface {
	BindPort(ctx sdk.Ctx, portID string) *capabilitytypes.Capability
	IsBound(ctx sdk.Ctx, portID string) bool
}
