package types

import (
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/authentication/types"
	capabilitytypes "github.com/vipernet-xyz/viper-network/x/capability/types"

	channeltypes "github.com/vipernet-xyz/viper-network/modules/core/04-channel/types"
)

// AccountKeeper defines the contract required for account APIs.
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.Address
	GetAccount(sdk.Ctx, sdk.Address) types.AccountI
}

// ChannelKeeper defines the expected IBC channel keeper
type ChannelKeeper interface {
	GetChannel(ctx sdk.Ctx, srcPort, srcChan string) (channel channeltypes.Channel, found bool)
	GetPacketCommitment(ctx sdk.Ctx, portID, channelID string, sequence uint64) []byte
	GetNextSequenceSend(ctx sdk.Ctx, portID, channelID string) (uint64, bool)
}

// PortKeeper defines the expected IBC port keeper
type PortKeeper interface {
	BindPort(ctx sdk.Context, portID string) *capabilitytypes.Capability
}

// BankKeeper defines the expected bank keeper
type BankKeeper interface {
	HasBalance(ctx sdk.Context, addr sdk.Address, amt sdk.Coin) bool
	SendCoinsFromAccountToModule(ctx sdk.Ctx, senderAddr sdk.Address, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Ctx, senderModule string, recipientAddr sdk.Address, amt sdk.Coins) error
	BlockedAddr(sdk.Address) bool
	IsSendEnabledCoins(ctx sdk.Context, coins ...sdk.Coin) error
}
