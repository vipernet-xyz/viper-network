package types

import (
	errorsmod "cosmossdk.io/errors"
	proto "github.com/cosmos/gogoproto/proto"
	codectypes "github.com/vipernet-xyz/viper-network/codec/types"
	sdk "github.com/vipernet-xyz/viper-network/types"
	msgservice "github.com/vipernet-xyz/viper-network/types/msgservice"
	govtypes "github.com/vipernet-xyz/viper-network/x/governance/types"

	ibcerrors "github.com/vipernet-xyz/viper-network/internal/errors"
	"github.com/vipernet-xyz/viper-network/modules/core/exported"
)

// RegisterInterfaces registers the client interfaces to protobuf Any.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterInterface(
		"ibc.core.client.v1.ClientState",
		(*exported.ClientState)(nil),
	)
	registry.RegisterInterface(
		"ibc.core.client.v1.ConsensusState",
		(*exported.ConsensusState)(nil),
	)
	registry.RegisterInterface(
		"ibc.core.client.v1.Header",
		(*exported.ClientMessage)(nil),
	)
	registry.RegisterInterface(
		"ibc.core.client.v1.Height",
		(*exported.Height)(nil),
		&Height{},
	)
	registry.RegisterInterface(
		"ibc.core.client.v1.Misbehaviour",
		(*exported.ClientMessage)(nil),
	)
	registry.RegisterImplementations(
		(*govtypes.Content)(nil),
		&ClientUpdateProposal{},
		&UpgradeProposal{},
	)
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgCreateClient{},
		&MsgUpdateClient{},
		&MsgUpgradeClient{},
		&MsgSubmitMisbehaviour{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

// PackClientState constructs a new Any packed with the given client state value. It returns
// an error if the client state can't be casted to a protobuf message or if the concrete
// implemention is not registered to the protobuf codec.
func PackClientState(clientState exported.ClientState) (*codectypes.Any, error) {
	msg, ok := clientState.(proto.Message)
	if !ok {
		return nil, errorsmod.Wrapf(ibcerrors.ErrPackAny, "cannot proto marshal %T", clientState)
	}

	anyClientState, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return nil, errorsmod.Wrap(ibcerrors.ErrPackAny, err.Error())
	}

	return anyClientState, nil
}

// UnpackClientState unpacks an Any into a ClientState. It returns an error if the
// client state can't be unpacked into a ClientState.
func UnpackClientState(any *codectypes.Any) (exported.ClientState, error) {
	if any == nil {
		return nil, errorsmod.Wrap(ibcerrors.ErrUnpackAny, "protobuf Any message cannot be nil")
	}

	clientState, ok := any.GetCachedValue().(exported.ClientState)
	if !ok {
		return nil, errorsmod.Wrapf(ibcerrors.ErrUnpackAny, "cannot unpack Any into ClientState %T", any)
	}

	return clientState, nil
}

// PackConsensusState constructs a new Any packed with the given consensus state value. It returns
// an error if the consensus state can't be casted to a protobuf message or if the concrete
// implemention is not registered to the protobuf codec.
func PackConsensusState(consensusState exported.ConsensusState) (*codectypes.Any, error) {
	msg, ok := consensusState.(proto.Message)
	if !ok {
		return nil, errorsmod.Wrapf(ibcerrors.ErrPackAny, "cannot proto marshal %T", consensusState)
	}

	anyConsensusState, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return nil, errorsmod.Wrap(ibcerrors.ErrPackAny, err.Error())
	}

	return anyConsensusState, nil
}

// MustPackConsensusState calls PackConsensusState and panics on error.
func MustPackConsensusState(consensusState exported.ConsensusState) *codectypes.Any {
	anyConsensusState, err := PackConsensusState(consensusState)
	if err != nil {
		panic(err)
	}

	return anyConsensusState
}

// UnpackConsensusState unpacks an Any into a ConsensusState. It returns an error if the
// consensus state can't be unpacked into a ConsensusState.
func UnpackConsensusState(any *codectypes.Any) (exported.ConsensusState, error) {
	if any == nil {
		return nil, errorsmod.Wrap(ibcerrors.ErrUnpackAny, "protobuf Any message cannot be nil")
	}

	consensusState, ok := any.GetCachedValue().(exported.ConsensusState)
	if !ok {
		return nil, errorsmod.Wrapf(ibcerrors.ErrUnpackAny, "cannot unpack Any into ConsensusState %T", any)
	}

	return consensusState, nil
}

// PackClientMessage constructs a new Any packed with the given value. It returns
// an error if the value can't be casted to a protobuf message or if the concrete
// implemention is not registered to the protobuf codec.
func PackClientMessage(clientMessage exported.ClientMessage) (*codectypes.Any, error) {
	msg, ok := clientMessage.(proto.Message)
	if !ok {
		return nil, errorsmod.Wrapf(ibcerrors.ErrPackAny, "cannot proto marshal %T", clientMessage)
	}

	protoAny, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return nil, errorsmod.Wrap(ibcerrors.ErrPackAny, err.Error())
	}

	return protoAny, nil
}

// UnpackClientMessage unpacks an Any into a ClientMessage. It returns an error if the
// consensus state can't be unpacked into a ClientMessage.
func UnpackClientMessage(any *codectypes.Any) (exported.ClientMessage, error) {
	if any == nil {
		return nil, errorsmod.Wrap(ibcerrors.ErrUnpackAny, "protobuf Any message cannot be nil")
	}

	clientMessage, ok := any.GetCachedValue().(exported.ClientMessage)
	if !ok {
		return nil, errorsmod.Wrapf(ibcerrors.ErrUnpackAny, "cannot unpack Any into Header %T", any)
	}

	return clientMessage, nil
}
