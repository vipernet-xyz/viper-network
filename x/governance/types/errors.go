package types

import (
	"fmt"

	github_com_viper_network_viper_core_types "github.com/vipernet-xyz/viper-network/types"
	sdk "github.com/vipernet-xyz/viper-network/types"
	sdkerrors "github.com/vipernet-xyz/viper-network/types/errors"
)

// Param module codespace constants
const (
	CodeUnknownSubspace               sdk.CodeType = 1
	CodeSettingParameter              sdk.CodeType = 2
	CodeEmptyData                     sdk.CodeType = 3
	CodeInvalidACL                    sdk.CodeType = 4
	CodeUnauthorizedParamChange       sdk.CodeType = 5
	CodeSubspaceNotFound              sdk.CodeType = 6
	CodeUnrecognizedDAOAction         sdk.CodeType = 7
	CodeZeroValueDAOAction            sdk.CodeType = 8
	CodeZeroHeightUpgrade             sdk.CodeType = 9
	CodeEmptyVersionUpgrade           sdk.CodeType = 10
	CodeUnauthorizedHeightParamChange sdk.CodeType = 11
	CodeUnrecognizedClientType        sdk.CodeType = 12
)

func ErrZeroHeightUpgrade(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeZeroHeightUpgrade, "the upgrade Height must not be zero")
}

func ErrZeroValueDAOAction(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeZeroValueDAOAction, "dao action value must not be zero: ")
}

func ErrUnrecognizedDAOAction(codespace sdk.CodespaceType, daoActionString string) sdk.Error {
	return sdk.NewError(codespace, CodeUnrecognizedDAOAction, "unrecognized dao action: "+daoActionString)
}
func ErrUnrecognizedClientType(codespace sdk.CodespaceType, clientTypeNumber github_com_viper_network_viper_core_types.Int64) sdk.Error {
	return sdk.NewError(codespace, CodeUnrecognizedClientType, "unrecognized client type: %v", clientTypeNumber)
}
func ErrInvalidACL(codespace sdk.CodespaceType, err error) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidACL, "invalid ACL: "+err.Error())
}

func ErrSubspaceNotFound(codespace sdk.CodespaceType, subspaceName string) sdk.Error {
	return sdk.NewError(codespace, CodeSubspaceNotFound, fmt.Sprintf("the subspace %s cannot be found", subspaceName))
}

func ErrUnauthorizedParamChange(codespace sdk.CodespaceType, owner sdk.Address, param string) sdk.Error {
	return sdk.NewError(codespace, CodeUnauthorizedParamChange,
		fmt.Sprintf("the param change is unathorized: Account %s does not have the permission to change param %s", owner, param))
}

func ErrUnauthorizedHeightParamChange(codespace sdk.CodespaceType, height int64, param string) sdk.Error {
	return sdk.NewError(codespace, CodeUnauthorizedHeightParamChange,
		fmt.Sprintf("the param change is unathorized: Wait For Upgrade Height %v to change param %s", height, param))
}

// ErrUnknownSubspace returns an unknown subspace error.
func ErrUnknownSubspace(codespace sdk.CodespaceType, space string) sdk.Error {
	return sdk.NewError(codespace, CodeUnknownSubspace, fmt.Sprintf("unknown subspace %s", space))
}

// ErrSettingParameter returns an error for failing to set a parameter.
func ErrSettingParameter(codespace sdk.CodespaceType, key, subkey, value, msg string) sdk.Error {
	return sdk.NewError(codespace, CodeSettingParameter, fmt.Sprintf("error setting parameter %s on %s (%s): %s", value, key, subkey, msg))
}

// ErrEmptyChanges returns an error for empty parameter changes.
func ErrEmptyChanges(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeEmptyData, "submitted parameter changes are empty")
}

// ErrEmptySubspace returns an error for an empty subspace.
func ErrEmptySubspace(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeEmptyData, "parameter subspace is empty")
}

// ErrEmptyKey returns an error for when an empty key is given.
func ErrEmptyKey(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeEmptyData, "parameter key is empty")
}

// ErrEmptyValue returns an error for when an empty key is given.
func ErrEmptyValue(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeEmptyData, "parameter value is empty")
}

// x/gov module sentinel errors
var (
	ErrUnknownProposal       = sdkerrors.Register(ModuleName, 2, "unknown proposal")
	ErrInactiveProposal      = sdkerrors.Register(ModuleName, 3, "inactive proposal")
	ErrAlreadyActiveProposal = sdkerrors.Register(ModuleName, 4, "proposal already active")
	// Errors 5 & 6 are legacy errors related to v1beta1.Proposal.
	ErrInvalidProposalContent  = sdkerrors.Register(ModuleName, 5, "invalid proposal content")
	ErrInvalidProposalType     = sdkerrors.Register(ModuleName, 6, "invalid proposal type")
	ErrInvalidVote             = sdkerrors.Register(ModuleName, 7, "invalid vote option")
	ErrInvalidGenesis          = sdkerrors.Register(ModuleName, 8, "invalid genesis state")
	ErrNoProposalHandlerExists = sdkerrors.Register(ModuleName, 9, "no handler exists for proposal type")
	ErrUnroutableProposalMsg   = sdkerrors.Register(ModuleName, 10, "proposal message not recognized by router")
	ErrNoProposalMsgs          = sdkerrors.Register(ModuleName, 11, "no messages proposed")
	ErrInvalidProposalMsg      = sdkerrors.Register(ModuleName, 12, "invalid proposal message")
	ErrInvalidSigner           = sdkerrors.Register(ModuleName, 13, "expected gov account as only signer for proposal message")
	ErrInvalidSignalMsg        = sdkerrors.Register(ModuleName, 14, "signal message is invalid")
	ErrMetadataTooLong         = sdkerrors.Register(ModuleName, 15, "metadata too long")
	ErrMinDepositTooSmall      = sdkerrors.Register(ModuleName, 16, "minimum deposit is too small")
)
