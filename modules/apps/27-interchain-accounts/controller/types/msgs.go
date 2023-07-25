package types

import (
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/vipernet-xyz/viper-network/types"

	ibcerrors "github.com/vipernet-xyz/viper-network/internal/errors"
	icatypes "github.com/vipernet-xyz/viper-network/modules/apps/27-interchain-accounts/types"
	host "github.com/vipernet-xyz/viper-network/modules/core/24-host"
)

var _ sdk.Msg1 = &MsgRegisterInterchainAccount{}

// NewMsgRegisterInterchainAccount creates a new instance of MsgRegisterInterchainAccount
func NewMsgRegisterInterchainAccount(connectionID, owner, version string) *MsgRegisterInterchainAccount {
	return &MsgRegisterInterchainAccount{
		ConnectionId: connectionID,
		Owner:        owner,
		Version:      version,
	}
}

// ValidateBasic implements sdk.Msg
func (msg MsgRegisterInterchainAccount) ValidateBasic() error {
	if err := host.ConnectionIdentifierValidator(msg.ConnectionId); err != nil {
		return errorsmod.Wrap(err, "invalid connection ID")
	}

	if strings.TrimSpace(msg.Owner) == "" {
		return errorsmod.Wrap(ibcerrors.ErrInvalidAddress, "owner address cannot be empty")
	}

	return nil
}

// GetSigners implements sdk.Msg
func (msg MsgRegisterInterchainAccount) GetSigners() []sdk.Address {
	Addr, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		panic(err)
	}

	return []sdk.Address{Addr}
}

// NewMsgSendTx creates a new instance of MsgSendTx
func NewMsgSendTx(owner, connectionID string, relativeTimeoutTimestamp uint64, packetData icatypes.InterchainAccountPacketData) *MsgSendTx {
	return &MsgSendTx{
		ConnectionId:    connectionID,
		Owner:           owner,
		RelativeTimeout: relativeTimeoutTimestamp,
		PacketData:      packetData,
	}
}

// ValidateBasic implements sdk.Msg
func (msg MsgSendTx) ValidateBasic() error {
	if err := host.ConnectionIdentifierValidator(msg.ConnectionId); err != nil {
		return errorsmod.Wrap(err, "invalid connection ID")
	}

	if strings.TrimSpace(msg.Owner) == "" {
		return errorsmod.Wrap(ibcerrors.ErrInvalidAddress, "owner address cannot be empty")
	}

	if err := msg.PacketData.ValidateBasic(); err != nil {
		return errorsmod.Wrap(err, "invalid interchain account packet data")
	}

	if msg.RelativeTimeout == 0 {
		return errorsmod.Wrap(ibcerrors.ErrInvalidRequest, "relative timeout cannot be zero")
	}

	return nil
}

// GetSigners implements sdk.Msg
func (msg MsgSendTx) GetSigners() []sdk.Address {
	Addr, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		panic(err)
	}

	return []sdk.Address{Addr}
}
