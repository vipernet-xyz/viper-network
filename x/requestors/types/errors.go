package types

import (
	"fmt"
	"strings"

	sdk "github.com/vipernet-xyz/viper-network/types"
)

type CodeType = sdk.CodeType

const (
	DefaultCodespace          sdk.CodespaceType = ModuleName
	CodeInvalidRequestor      CodeType          = 101
	CodeInvalidInput          CodeType          = 103
	CodeRequestorJailed       CodeType          = 104
	CodeRequestorNotJailed    CodeType          = 105
	CodeMissingSelfDelegation CodeType          = 106
	CodeInvalidStatus         CodeType          = 110
	CodeMinimumStake          CodeType          = 111
	CodeNotEnoughCoins        CodeType          = 112
	CodeInvalidStakeAmount    CodeType          = 115
	CodeNoChains              CodeType          = 116
	CodeInvalidNetworkID      CodeType          = 117
	CodeTooManyChains         CodeType          = 118
	CodeInvalidGeoZone        CodeType          = 119
	CodeMaxRequestors         CodeType          = 120
	CodeMinimumEditStake      CodeType          = 121
	CodeNoGeoZones            CodeType          = 122
	CodeNoServicers           CodeType          = 123
	CodeNumServicers          CodeType          = 124
)

func ErrTooManyChains(Codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(Codespace, CodeTooManyChains, "requestor staking for too many chains")
}
func ErrNoGeoZones(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeNoGeoZones, "validator must stake with hosted geozones")
}
func ErrNoChains(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeNoChains, "validator must stake with hosted blockchains")
}
func ErrNilRequestorAddr(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidInput, "requestor address is nil")
}
func ErrRequestorStatus(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidStatus, "requestor status is not valid")
}
func ErrNoRequestorFound(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidRequestor, "requestor does not exist for that address")
}

func ErrBadStakeAmount(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidStakeAmount, "the stake amount is invalid")
}

func ErrNoServicers(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeNoServicers, "the number of servicers is zero")
}

func ErrNotEnoughCoins(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeNotEnoughCoins, "requestor does not have enough coins in their account")
}

func ErrMaxRequestors(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeMaxRequestors, "the threshold of the amount of requestors authorized ")
}

func ErrMinimumStake(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeMinimumStake, "requestor isn't staking above the minimum")
}

func ErrRequestorPubKeyExists(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidRequestor, "requestor already exist for this pubkey, must use new requestor pubkey")
}

func ErrRequestorPubKeyTypeNotSupported(codespace sdk.CodespaceType, keyType string, supportedTypes []string) sdk.Error {
	msg := fmt.Sprintf("requestor pubkey type %s is not supported, must use %s", keyType, strings.Join(supportedTypes, ","))
	return sdk.NewError(codespace, CodeInvalidRequestor, msg)
}

func ErrNoRequestorForAddress(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidRequestor, "that address is not associated with any known requestor")
}

func ErrBadRequestorAddr(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidRequestor, "requestor does not exist for that address")
}

func ErrRequestorJailed(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeRequestorJailed, "requestor still jailed, cannot yet be unjailed")
}

func ErrRequestorNotJailed(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeRequestorNotJailed, "requestor not jailed, cannot be unjailed")
}

func ErrMissingRequestorStake(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeMissingSelfDelegation, "requestor has no stake; cannot be unjailed")
}

func ErrStakeTooLow(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeRequestorNotJailed, "requestor's self delegation less than min stake, cannot be unjailed")
}

func ErrInvalidNetworkIdentifier(codespace sdk.CodespaceType, err error) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidNetworkID, "the requestors network identifier is not valid: "+err.Error())
}

func ErrInvalidGeoZoneIdentifier(codespace sdk.CodespaceType, err error) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidGeoZone, "the requestors geo zone identifier is not valid: "+err.Error())
}

func ErrMinimumEditStake(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeMinimumEditStake, "requestor must edit stake with a stake greater than or equal to current stake")
}

func ErrNumServicers(Codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(Codespace, CodeNumServicers, "Number of servicer's out of range")
}
