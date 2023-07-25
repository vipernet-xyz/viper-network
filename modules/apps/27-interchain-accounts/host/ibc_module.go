package host

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/vipernet-xyz/viper-network/types"
	capabilitytypes "github.com/vipernet-xyz/viper-network/x/capability/types"

	ibcerrors "github.com/vipernet-xyz/viper-network/internal/errors"
	"github.com/vipernet-xyz/viper-network/modules/apps/27-interchain-accounts/host/keeper"
	"github.com/vipernet-xyz/viper-network/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/vipernet-xyz/viper-network/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/vipernet-xyz/viper-network/modules/core/04-channel/types"
	ibcexported "github.com/vipernet-xyz/viper-network/modules/core/exported"
)

// IBCModule implements the ICS26 interface for interchain accounts host chains
type IBCModule struct {
	keeper keeper.Keeper
}

// NewIBCModule creates a new IBCModule given the associated keeper
func NewIBCModule(k keeper.Keeper) IBCModule {
	return IBCModule{
		keeper: k,
	}
}

// OnChanOpenInit implements the IBCModule interface
func (im IBCModule) OnChanOpenInit(
	ctx sdk.Ctx,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version string,
) (string, error) {
	return "", errorsmod.Wrap(icatypes.ErrInvalidChannelFlow, "channel handshake must be initiated by controller chain")
}

// OnChanOpenTry implements the IBCModule interface
func (im IBCModule) OnChanOpenTry(
	ctx sdk.Ctx,
	order channeltypes.Order,
	connectionHops []string,
	portID,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (string, error) {
	if !im.keeper.IsHostEnabled(ctx) {
		return "", types.ErrHostSubModuleDisabled
	}

	return im.keeper.OnChanOpenTry(ctx, order, connectionHops, portID, channelID, chanCap, counterparty, counterpartyVersion)
}

// OnChanOpenAck implements the IBCModule interface
func (im IBCModule) OnChanOpenAck(
	ctx sdk.Ctx,
	portID,
	channelID string,
	counterpartyChannelID string,
	counterpartyVersion string,
) error {
	return errorsmod.Wrap(icatypes.ErrInvalidChannelFlow, "channel handshake must be initiated by controller chain")
}

// OnChanOpenAck implements the IBCModule interface
func (im IBCModule) OnChanOpenConfirm(
	ctx sdk.Ctx,
	portID,
	channelID string,
) error {
	if !im.keeper.IsHostEnabled(ctx) {
		return types.ErrHostSubModuleDisabled
	}

	return im.keeper.OnChanOpenConfirm(ctx, portID, channelID)
}

// OnChanCloseInit implements the IBCModule interface
func (im IBCModule) OnChanCloseInit(
	ctx sdk.Ctx,
	portID,
	channelID string,
) error {
	// Disallow user-initiated channel closing for interchain account channels
	return errorsmod.Wrap(ibcerrors.ErrInvalidRequest, "user cannot close channel")
}

// OnChanCloseConfirm implements the IBCModule interface
func (im IBCModule) OnChanCloseConfirm(
	ctx sdk.Ctx,
	portID,
	channelID string,
) error {
	return im.keeper.OnChanCloseConfirm(ctx, portID, channelID)
}

// OnRecvPacket implements the IBCModule interface
func (im IBCModule) OnRecvPacket(
	ctx sdk.Ctx,
	packet channeltypes.Packet,
	_ sdk.Address,
) ibcexported.Acknowledgement {
	logger := im.keeper.Logger(ctx)
	if !im.keeper.IsHostEnabled(ctx) {
		logger.Info("host submodule is disabled")
		return channeltypes.NewErrorAcknowledgement(types.ErrHostSubModuleDisabled)
	}

	txResponse, err := im.keeper.OnRecvPacket(ctx, packet)
	ack := channeltypes.NewResultAcknowledgement(txResponse)
	if err != nil {
		ack = channeltypes.NewErrorAcknowledgement(err)
		logger.Error(fmt.Sprintf("%s sequence %d", err.Error(), packet.Sequence))
	} else {
		logger.Info("successfully handled packet sequence: %d", packet.Sequence)
	}

	// Emit an event indicating a successful or failed acknowledgement.
	keeper.EmitAcknowledgementEvent(ctx, packet, ack, err)

	// NOTE: acknowledgement will be written synchronously during IBC handler execution.
	return ack
}

// OnAcknowledgementPacket implements the IBCModule interface
func (im IBCModule) OnAcknowledgementPacket(
	ctx sdk.Ctx,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.Address,
) error {
	return errorsmod.Wrap(icatypes.ErrInvalidChannelFlow, "cannot receive acknowledgement on a host channel end, a host chain does not send a packet over the channel")
}

// OnTimeoutPacket implements the IBCModule interface
func (im IBCModule) OnTimeoutPacket(
	ctx sdk.Ctx,
	packet channeltypes.Packet,
	relayer sdk.Address,
) error {
	return errorsmod.Wrap(icatypes.ErrInvalidChannelFlow, "cannot cause a packet timeout on a host channel end, a host chain does not send a packet over the channel")
}
