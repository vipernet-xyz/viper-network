package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/vipernet-xyz/viper-network/types"
	sdk1 "github.com/vipernet-xyz/viper-network/types"
	auth "github.com/vipernet-xyz/viper-network/x/authentication"

	ibcerrors "github.com/vipernet-xyz/viper-network/internal/errors"
	channeltypes "github.com/vipernet-xyz/viper-network/modules/core/04-channel/types"
	host "github.com/vipernet-xyz/viper-network/modules/core/24-host"
)

const gasCostPerIteration = uint64(10)

var _ auth.Authorization = &TransferAuthorization{}

// NewTransferAuthorization creates a new TransferAuthorization object.
func NewTransferAuthorization(allocations ...Allocation) *TransferAuthorization {
	return &TransferAuthorization{
		Allocations: allocations,
	}
}

// MsgTypeURL implements Authorization.MsgTypeURL.
func (a TransferAuthorization) MsgTypeURL() string {
	return sdk1.MsgTypeURL(&MsgTransfer{})
}

// Accept implements Authorization.Accept.
func (a TransferAuthorization) Accept(ctx sdk.Ctx, msg sdk.Msg1) (auth.AcceptResponse, error) {
	msgTransfer, ok := msg.(*MsgTransfer)
	if !ok {
		return auth.AcceptResponse{}, errorsmod.Wrap(ibcerrors.ErrInvalidType, "type mismatch")
	}

	for index, allocation := range a.Allocations {
		if allocation.SourceChannel == msgTransfer.SourceChannel && allocation.SourcePort == msgTransfer.SourcePort {
			limitLeft, isNegative := allocation.SpendLimit.SafeSub1(msgTransfer.Token)
			if isNegative {
				return auth.AcceptResponse{}, errorsmod.Wrapf(ibcerrors.ErrInsufficientFunds, "requested amount is more than spend limit")
			}

			if !isAllowedAddress(ctx, msgTransfer.Receiver, allocation.AllowList) {
				return auth.AcceptResponse{}, errorsmod.Wrap(ibcerrors.ErrInvalidAddress, "not allowed address for transfer")
			}

			if limitLeft.IsZero() {
				a.Allocations = append(a.Allocations[:index], a.Allocations[index+1:]...)
				if len(a.Allocations) == 0 {
					return auth.AcceptResponse{Accept: true, Delete: true}, nil
				}
				return auth.AcceptResponse{Accept: true, Delete: false, Updated: &TransferAuthorization{
					Allocations: a.Allocations,
				}}, nil
			}
			a.Allocations[index] = Allocation{
				SourcePort:    allocation.SourcePort,
				SourceChannel: allocation.SourceChannel,
				SpendLimit:    limitLeft,
				AllowList:     allocation.AllowList,
			}

			return auth.AcceptResponse{Accept: true, Delete: false, Updated: &TransferAuthorization{
				Allocations: a.Allocations,
			}}, nil
		}
	}
	return auth.AcceptResponse{}, errorsmod.Wrapf(ibcerrors.ErrNotFound, "requested port and channel allocation does not exist")
}

// ValidateBasic implements Authorization.ValidateBasic.
func (a TransferAuthorization) ValidateBasic() error {
	if len(a.Allocations) == 0 {
		return errorsmod.Wrap(ErrInvalidAuthorization, "allocations cannot be empty")
	}

	foundChannels := make(map[string]bool, 0)

	for _, allocation := range a.Allocations {
		if _, found := foundChannels[allocation.SourceChannel]; found {
			return errorsmod.Wrapf(channeltypes.ErrInvalidChannel, "duplicate source channel ID: %s", allocation.SourceChannel)
		}

		foundChannels[allocation.SourceChannel] = true

		if allocation.SpendLimit == nil {
			return errorsmod.Wrap(ibcerrors.ErrInvalidCoins, "spend limit cannot be nil")
		}

		if err := allocation.SpendLimit.Validate(); err != nil {
			return errorsmod.Wrapf(ibcerrors.ErrInvalidCoins, err.Error())
		}

		if err := host.PortIdentifierValidator(allocation.SourcePort); err != nil {
			return errorsmod.Wrap(err, "invalid source port ID")
		}

		if err := host.ChannelIdentifierValidator(allocation.SourceChannel); err != nil {
			return errorsmod.Wrap(err, "invalid source channel ID")
		}

		found := make(map[string]bool, 0)
		for i := 0; i < len(allocation.AllowList); i++ {
			if found[allocation.AllowList[i]] {
				return errorsmod.Wrapf(ErrInvalidAuthorization, "duplicate entry in allow list %s", allocation.AllowList[i])
			}
			found[allocation.AllowList[i]] = true
		}
	}

	return nil
}

// isAllowedAddress returns a boolean indicating if the receiver address is valid for transfer.
// gasCostPerIteration gas is consumed for each iteration.
func isAllowedAddress(ctx sdk.Ctx, receiver string, allowedAddrs []string) bool {
	if len(allowedAddrs) == 0 {
		return true
	}

	for _, addr := range allowedAddrs {
		ctx.GasMeter().ConsumeGas(gasCostPerIteration, "transfer authorization")
		if addr == receiver {
			return true
		}
	}
	return false
}
