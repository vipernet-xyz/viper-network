package types

import (
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/authentication/types"
	capabilitytypes "github.com/vipernet-xyz/viper-network/x/capability/types"

	connectiontypes "github.com/vipernet-xyz/viper-network/modules/core/03-connection/types"
	channeltypes "github.com/vipernet-xyz/viper-network/modules/core/04-channel/types"
	ibcexported "github.com/vipernet-xyz/viper-network/modules/core/exported"
)

// AccountKeeper defines the contract required for account APIs.
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.Addresses
	GetModuleAccount(ctx sdk.Ctx, name string) types.ModuleAccountI
}

// BankKeeper defines the expected bank keeper
type BankKeeper interface {
	SendCoins(ctx sdk.Ctx, fromAddr sdk.Address, toAddr sdk.Address, amt sdk.Coins) error
	MintCoins(ctx sdk.Ctx, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Ctx, moduleName string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Ctx, senderModule string, recipientAddr sdk.Address, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Ctx, senderAddr sdk.Address, recipientModule string, amt sdk.Coins) error
	BlockedAddr(addr sdk.Address) bool
	IsSendEnabledCoin(ctx sdk.Ctx, coin sdk.Coin) bool
}

// ChannelKeeper defines the expected IBC channel keeper
type ChannelKeeper interface {
	GetChannel(ctx sdk.Ctx, srcPort, srcChan string) (channel channeltypes.Channel, found bool)
	GetNextSequenceSend(ctx sdk.Ctx, portID, channelID string) (uint64, bool)
}

// ClientKeeper defines the expected IBC client keeper
type ClientKeeper interface {
	GetClientConsensusState(ctx sdk.Ctx, clientID string) (connection ibcexported.ConsensusState, found bool)
}

// ConnectionKeeper defines the expected IBC connection keeper
type ConnectionKeeper interface {
	GetConnection(ctx sdk.Ctx, connectionID string) (connection connectiontypes.ConnectionEnd, found bool)
}

// PortKeeper defines the expected IBC port keeper
type PortKeeper interface {
	BindPort(ctx sdk.Ctx, portID string) *capabilitytypes.Capability
}
