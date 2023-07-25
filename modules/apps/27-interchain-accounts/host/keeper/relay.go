package keeper

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/gogoproto/proto"
	codectypes "github.com/vipernet-xyz/viper-network/codec/types"
	sdk "github.com/vipernet-xyz/viper-network/types"

	ibcerrors "github.com/vipernet-xyz/viper-network/internal/errors"
	"github.com/vipernet-xyz/viper-network/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/vipernet-xyz/viper-network/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/vipernet-xyz/viper-network/modules/core/04-channel/types"
)

// OnRecvPacket handles a given interchain accounts packet on a destination host chain.
// If the transaction is successfully executed, the transaction response bytes will be returned.
func (k Keeper) OnRecvPacket(ctx sdk.Ctx, packet channeltypes.Packet) ([]byte, error) {
	var data icatypes.InterchainAccountPacketData

	if err := icatypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		// UnmarshalJSON errors are indeterminate and therefore are not wrapped and included in failed acks
		return nil, errorsmod.Wrapf(icatypes.ErrUnknownDataType, "cannot unmarshal ICS-27 interchain account packet data")
	}

	switch data.Type {
	case icatypes.EXECUTE_TX:
		msgs, err := icatypes.DeserializeCosmosTx(k.cdc, data.Data)
		if err != nil {
			return nil, errorsmod.Wrapf(err, "failed to deserialize interchain account transaction")
		}

		txResponse, err := k.executeTx(ctx, packet.SourcePort, packet.DestinationPort, packet.DestinationChannel, msgs)
		if err != nil {
			return nil, errorsmod.Wrapf(err, "failed to execute interchain account transaction")
		}
		return txResponse, nil
	default:
		return nil, icatypes.ErrUnknownDataType
	}
}

// executeTx attempts to execute the provided transaction. It begins by authenticating the transaction signer.
// If authentication succeeds, it does basic validation of the messages before attempting to deliver each message
// into state. The state changes will only be committed if all messages in the transaction succeed. Thus the
// execution of the transaction is atomic, all state changes are reverted if a single message fails.
func (k Keeper) executeTx(ctx sdk.Ctx, sourcePort, destPort, destChannel string, msgs []sdk.Msg1) ([]byte, error) {
	channel, found := k.channelKeeper.GetChannel(ctx, destPort, destChannel)
	if !found {
		return nil, channeltypes.ErrChannelNotFound
	}

	if err := k.authenticateTx(ctx, msgs, channel.ConnectionHops[0], sourcePort); err != nil {
		return nil, err
	}

	txMsgData := &sdk.TxMsgData{
		MsgResponses: make([]*codectypes.Any, len(msgs)),
	}

	// CacheContext returns a new context with the multi-store branched into a cached storage object
	// writeCache is called only if all msgs succeed, performing state transitions atomically
	cacheCtx, writeCache := ctx.CacheContext()
	for i, msg := range msgs {
		if err := msg.ValidateBasic(); err != nil {
			return nil, err
		}

		protoAny, err := k.executeMsg(cacheCtx, msg)
		if err != nil {
			return nil, err
		}

		txMsgData.MsgResponses[i] = protoAny
	}

	writeCache()

	txResponse, err := proto.Marshal(txMsgData)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to marshal tx data")
	}

	return txResponse, nil
}

// authenticateTx ensures the provided msgs contain the correct interchain account signer address retrieved
// from state using the provided controller port identifier
func (k Keeper) authenticateTx(ctx sdk.Ctx, msgs []sdk.Msg1, connectionID, portID string) error {
	interchainAccountAddr, found := k.GetInterchainAccountAddress(ctx, connectionID, portID)
	if !found {
		return errorsmod.Wrapf(icatypes.ErrInterchainAccountNotFound, "failed to retrieve interchain account on port %s", portID)
	}

	allowMsgs := k.GetAllowMessages(ctx)
	for _, msg := range msgs {
		if !types.ContainsMsgType(allowMsgs, msg) {
			return errorsmod.Wrapf(ibcerrors.ErrUnauthorized, "message type not allowed: %s", sdk.MsgTypeURL(msg))
		}

		for _, signer := range msg.GetSigners() {
			if interchainAccountAddr != signer.String() {
				return errorsmod.Wrapf(ibcerrors.ErrUnauthorized, "unexpected signer address: expected %s, got %s", interchainAccountAddr, signer.String())
			}
		}
	}

	return nil
}

// Attempts to get the message handler from the router and if found will then execute the message.
// If the message execution is successful, the proto marshaled message response will be returned.
func (k Keeper) executeMsg(ctx sdk.Context, msg sdk.Msg1) (*codectypes.Any, error) {
	handler := k.msgRouter.Handler(msg)
	if handler == nil {
		return nil, icatypes.ErrInvalidRoute
	}

	res, err := handler(ctx, msg)
	if err != nil {
		return nil, err
	}

	// NOTE: The sdk msg handler creates a new EventManager, so events must be correctly propagated back to the current context
	ctx.EventManager().EmitEvents(res.GetEvents())

	// Each individual sdk.Result has exactly one Msg response. We aggregate here.
	msgResponse := res.MsgResponses[0]
	if msgResponse == nil {
		return nil, errorsmod.Wrapf(ibcerrors.ErrLogic, "got nil Msg response for msg %s", sdk.MsgTypeURL(msg))
	}

	return msgResponse, nil
}
