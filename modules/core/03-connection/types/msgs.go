package types

import (
	errorsmod "cosmossdk.io/errors"
	codectypes "github.com/vipernet-xyz/viper-network/codec/types"
	sdk "github.com/vipernet-xyz/viper-network/types"

	ibcerrors "github.com/vipernet-xyz/viper-network/internal/errors"
	clienttypes "github.com/vipernet-xyz/viper-network/modules/core/02-client/types"
	commitmenttypes "github.com/vipernet-xyz/viper-network/modules/core/23-commitment/types"
	host "github.com/vipernet-xyz/viper-network/modules/core/24-host"
	"github.com/vipernet-xyz/viper-network/modules/core/exported"
)

var (
	_ sdk.Msg1 = &MsgConnectionOpenInit{}
	_ sdk.Msg1 = &MsgConnectionOpenConfirm{}
	_ sdk.Msg1 = &MsgConnectionOpenAck{}
	_ sdk.Msg1 = &MsgConnectionOpenTry{}

	_ codectypes.UnpackInterfacesMessage = MsgConnectionOpenTry{}
	_ codectypes.UnpackInterfacesMessage = MsgConnectionOpenAck{}
)

// NewMsgConnectionOpenInit creates a new MsgConnectionOpenInit instance. It sets the
// counterparty connection identifier to be empty.
//
//nolint:interfacer
func NewMsgConnectionOpenInit(
	clientID, counterpartyClientID string,
	counterpartyPrefix commitmenttypes.MerklePrefix,
	version *Version, delayPeriod uint64, signer string,
) *MsgConnectionOpenInit {
	// counterparty must have the same delay period
	counterparty := NewCounterparty(counterpartyClientID, "", counterpartyPrefix)
	return &MsgConnectionOpenInit{
		ClientId:     clientID,
		Counterparty: counterparty,
		Version:      version,
		DelayPeriod:  delayPeriod,
		Signer:       signer,
	}
}

// ValidateBasic implements sdk.Msg.
func (msg MsgConnectionOpenInit) ValidateBasic() error {
	if err := host.ClientIdentifierValidator(msg.ClientId); err != nil {
		return errorsmod.Wrap(err, "invalid client ID")
	}
	if msg.Counterparty.ConnectionId != "" {
		return errorsmod.Wrap(ErrInvalidCounterparty, "counterparty connection identifier must be empty")
	}

	// NOTE: Version can be nil on MsgConnectionOpenInit
	if msg.Version != nil {
		if err := ValidateVersion(msg.Version); err != nil {
			return errorsmod.Wrap(err, "basic validation of the provided version failed")
		}
	}
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return errorsmod.Wrapf(ibcerrors.ErrInvalidAddress, "string could not be parsed as address: %v", err)
	}
	return msg.Counterparty.ValidateBasic()
}

// GetSigners implements sdk.Msg
func (msg MsgConnectionOpenInit) GetSigners() []sdk.Address {
	accAddr, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.Address{accAddr}
}

// NewMsgConnectionOpenTry creates a new MsgConnectionOpenTry instance
//
//nolint:interfacer
func NewMsgConnectionOpenTry(
	clientID, counterpartyConnectionID, counterpartyClientID string,
	counterpartyClient exported.ClientState,
	counterpartyPrefix commitmenttypes.MerklePrefix,
	counterpartyVersions []*Version, delayPeriod uint64,
	proofInit, proofClient, proofConsensus []byte,
	proofHeight, consensusHeight clienttypes.Height, signer string,
) *MsgConnectionOpenTry {
	counterparty := NewCounterparty(counterpartyClientID, counterpartyConnectionID, counterpartyPrefix)
	protoAny, _ := clienttypes.PackClientState(counterpartyClient)
	return &MsgConnectionOpenTry{
		ClientId:             clientID,
		ClientState:          protoAny,
		Counterparty:         counterparty,
		CounterpartyVersions: counterpartyVersions,
		DelayPeriod:          delayPeriod,
		ProofInit:            proofInit,
		ProofClient:          proofClient,
		ProofConsensus:       proofConsensus,
		ProofHeight:          proofHeight,
		ConsensusHeight:      consensusHeight,
		Signer:               signer,
	}
}

// ValidateBasic implements sdk.Msg
func (msg MsgConnectionOpenTry) ValidateBasic() error {
	if msg.PreviousConnectionId != "" {
		return errorsmod.Wrap(ErrInvalidConnectionIdentifier, "previous connection identifier must be empty, this field has been deprecated as crossing hellos are no longer supported")
	}
	if err := host.ClientIdentifierValidator(msg.ClientId); err != nil {
		return errorsmod.Wrap(err, "invalid client ID")
	}
	// counterparty validate basic allows empty counterparty connection identifiers
	if err := host.ConnectionIdentifierValidator(msg.Counterparty.ConnectionId); err != nil {
		return errorsmod.Wrap(err, "invalid counterparty connection ID")
	}
	if msg.ClientState == nil {
		return errorsmod.Wrap(clienttypes.ErrInvalidClient, "counterparty client is nil")
	}
	clientState, err := clienttypes.UnpackClientState(msg.ClientState)
	if err != nil {
		return errorsmod.Wrapf(clienttypes.ErrInvalidClient, "unpack err: %v", err)
	}
	if err := clientState.Validate(); err != nil {
		return errorsmod.Wrap(err, "counterparty client is invalid")
	}
	if len(msg.CounterpartyVersions) == 0 {
		return errorsmod.Wrap(ibcerrors.ErrInvalidVersion, "empty counterparty versions")
	}
	for i, version := range msg.CounterpartyVersions {
		if err := ValidateVersion(version); err != nil {
			return errorsmod.Wrapf(err, "basic validation failed on version with index %d", i)
		}
	}
	if len(msg.ProofInit) == 0 {
		return errorsmod.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof init")
	}
	if len(msg.ProofClient) == 0 {
		return errorsmod.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit empty proof client")
	}
	if len(msg.ProofConsensus) == 0 {
		return errorsmod.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof of consensus state")
	}
	if msg.ConsensusHeight.IsZero() {
		return errorsmod.Wrap(ibcerrors.ErrInvalidHeight, "consensus height must be non-zero")
	}
	_, err = sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return errorsmod.Wrapf(ibcerrors.ErrInvalidAddress, "string could not be parsed as address: %v", err)
	}
	return msg.Counterparty.ValidateBasic()
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgConnectionOpenTry) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return unpacker.UnpackAny(msg.ClientState, new(exported.ClientState))
}

// GetSigners implements sdk.Msg
func (msg MsgConnectionOpenTry) GetSigners() []sdk.Address {
	accAddr, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.Address{accAddr}
}

// NewMsgConnectionOpenAck creates a new MsgConnectionOpenAck instance
//
//nolint:interfacer
func NewMsgConnectionOpenAck(
	connectionID, counterpartyConnectionID string, counterpartyClient exported.ClientState,
	proofTry, proofClient, proofConsensus []byte,
	proofHeight, consensusHeight clienttypes.Height,
	version *Version,
	signer string,
) *MsgConnectionOpenAck {
	protoAny, _ := clienttypes.PackClientState(counterpartyClient)
	return &MsgConnectionOpenAck{
		ConnectionId:             connectionID,
		CounterpartyConnectionId: counterpartyConnectionID,
		ClientState:              protoAny,
		ProofTry:                 proofTry,
		ProofClient:              proofClient,
		ProofConsensus:           proofConsensus,
		ProofHeight:              proofHeight,
		ConsensusHeight:          consensusHeight,
		Version:                  version,
		Signer:                   signer,
	}
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgConnectionOpenAck) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return unpacker.UnpackAny(msg.ClientState, new(exported.ClientState))
}

// ValidateBasic implements sdk.Msg
func (msg MsgConnectionOpenAck) ValidateBasic() error {
	if !IsValidConnectionID(msg.ConnectionId) {
		return ErrInvalidConnectionIdentifier
	}
	if err := host.ConnectionIdentifierValidator(msg.CounterpartyConnectionId); err != nil {
		return errorsmod.Wrap(err, "invalid counterparty connection ID")
	}
	if err := ValidateVersion(msg.Version); err != nil {
		return err
	}
	if msg.ClientState == nil {
		return errorsmod.Wrap(clienttypes.ErrInvalidClient, "counterparty client is nil")
	}
	clientState, err := clienttypes.UnpackClientState(msg.ClientState)
	if err != nil {
		return errorsmod.Wrapf(clienttypes.ErrInvalidClient, "unpack err: %v", err)
	}
	if err := clientState.Validate(); err != nil {
		return errorsmod.Wrap(err, "counterparty client is invalid")
	}
	if len(msg.ProofTry) == 0 {
		return errorsmod.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof try")
	}
	if len(msg.ProofClient) == 0 {
		return errorsmod.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit empty proof client")
	}
	if len(msg.ProofConsensus) == 0 {
		return errorsmod.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof of consensus state")
	}
	if msg.ConsensusHeight.IsZero() {
		return errorsmod.Wrap(ibcerrors.ErrInvalidHeight, "consensus height must be non-zero")
	}
	_, err = sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return errorsmod.Wrapf(ibcerrors.ErrInvalidAddress, "string could not be parsed as address: %v", err)
	}
	return nil
}

// GetSigners implements sdk.Msg
func (msg MsgConnectionOpenAck) GetSigners() []sdk.Address {
	accAddr, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.Address{accAddr}
}

// NewMsgConnectionOpenConfirm creates a new MsgConnectionOpenConfirm instance
//
//nolint:interfacer
func NewMsgConnectionOpenConfirm(
	connectionID string, proofAck []byte, proofHeight clienttypes.Height,
	signer string,
) *MsgConnectionOpenConfirm {
	return &MsgConnectionOpenConfirm{
		ConnectionId: connectionID,
		ProofAck:     proofAck,
		ProofHeight:  proofHeight,
		Signer:       signer,
	}
}

// ValidateBasic implements sdk.Msg
func (msg MsgConnectionOpenConfirm) ValidateBasic() error {
	if !IsValidConnectionID(msg.ConnectionId) {
		return ErrInvalidConnectionIdentifier
	}
	if len(msg.ProofAck) == 0 {
		return errorsmod.Wrap(commitmenttypes.ErrInvalidProof, "cannot submit an empty proof ack")
	}
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return errorsmod.Wrapf(ibcerrors.ErrInvalidAddress, "string could not be parsed as address: %v", err)
	}
	return nil
}

// GetSigners implements sdk.Msg
func (msg MsgConnectionOpenConfirm) GetSigners() []sdk.Address {
	accAddr, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.Address{accAddr}
}
