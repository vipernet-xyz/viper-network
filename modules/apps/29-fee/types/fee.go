package types

import (
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/vipernet-xyz/viper-network/types"

	ibcerrors "github.com/vipernet-xyz/viper-network/internal/errors"
	channeltypes "github.com/vipernet-xyz/viper-network/modules/core/04-channel/types"
)

// NewPacketFee creates and returns a new PacketFee struct including the incentivization fees, refund address and relayers
func NewPacketFee(fee Fee, refundAddr string, relayers []string) PacketFee {
	return PacketFee{
		Fee:           fee,
		RefundAddress: refundAddr,
		Relayers:      relayers,
	}
}

// Validate performs basic stateless validation of the associated PacketFee
func (p PacketFee) Validate() error {
	_, err := sdk.AccAddressFromBech32(p.RefundAddress)
	if err != nil {
		return errorsmod.Wrap(err, "failed to convert RefundAddress into sdk.AccAddress")
	}

	// enforce relayers are not set
	if len(p.Relayers) != 0 {
		return ErrRelayersNotEmpty
	}

	if err := p.Fee.Validate(); err != nil {
		return err
	}

	return nil
}

// NewPacketFees creates and returns a new PacketFees struct including a list of type PacketFee
func NewPacketFees(packetFees []PacketFee) PacketFees {
	return PacketFees{
		PacketFees: packetFees,
	}
}

// NewIdentifiedPacketFees creates and returns a new IdentifiedPacketFees struct containing a packet ID and packet fees
func NewIdentifiedPacketFees(packetID channeltypes.PacketId, packetFees []PacketFee) IdentifiedPacketFees {
	return IdentifiedPacketFees{
		PacketId:   packetID,
		PacketFees: packetFees,
	}
}

// NewFee creates and returns a new Fee struct encapsulating the receive, acknowledgement and timeout fees as sdk.Coins
func NewFee(recvFee, ackFee, timeoutFee sdk.Coins) Fee {
	return Fee{
		RecvFee:    recvFee,
		AckFee:     ackFee,
		TimeoutFee: timeoutFee,
	}
}

// Total returns the total amount for a given Fee
func (f Fee) Total() sdk.Coins {
	a := f.RecvFee.Add1(f.AckFee...).Add1(f.TimeoutFee...)
	return a
}

// Validate asserts that each Fee is valid and all three Fees are not empty or zero
func (f Fee) Validate() error {
	var errFees []string
	if !f.AckFee.IsValid() {
		errFees = append(errFees, "ack fee invalid")
	}
	if !f.RecvFee.IsValid() {
		errFees = append(errFees, "recv fee invalid")
	}
	if !f.TimeoutFee.IsValid() {
		errFees = append(errFees, "timeout fee invalid")
	}

	if len(errFees) > 0 {
		return errorsmod.Wrapf(ibcerrors.ErrInvalidCoins, "contains invalid fees: %s", strings.Join(errFees, " , "))
	}

	// if all three fee's are zero or empty return an error
	if f.AckFee.IsZero() && f.RecvFee.IsZero() && f.TimeoutFee.IsZero() {
		return errorsmod.Wrap(ibcerrors.ErrInvalidCoins, "all fees are zero")
	}

	return nil
}
