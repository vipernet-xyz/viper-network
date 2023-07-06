package types

import (
	"fmt"
	"strings"

	sdk "github.com/vipernet-xyz/viper-network/types"
)

type CodeType = sdk.CodeType

const (
	DefaultCodespace          sdk.CodespaceType = ModuleName
	CodeInvalidProvider       CodeType          = 101
	CodeInvalidInput          CodeType          = 103
	CodeProviderJailed        CodeType          = 104
	CodeProviderNotJailed     CodeType          = 105
	CodeMissingSelfDelegation CodeType          = 106
	CodeInvalidStatus         CodeType          = 110
	CodeMinimumStake          CodeType          = 111
	CodeNotEnoughCoins        CodeType          = 112
	CodeInvalidStakeAmount    CodeType          = 115
	CodeNoChains              CodeType          = 116
	CodeInvalidNetworkID      CodeType          = 117
	CodeTooManyChains         CodeType          = 118
	CodeInvalidGeoZone        CodeType          = 119
	CodeMaxProviders          CodeType          = 120
	CodeMinimumEditStake      CodeType          = 121
	CodeNoGeoZones            CodeType          = 122
	CodeNoServicers           CodeType          = 123
	CodeNumServicers          CodeType          = 124
)

func ErrTooManyChains(Codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(Codespace, CodeTooManyChains, "provider staking for too many chains")
}
func ErrNoGeoZones(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeNoGeoZones, "validator must stake with hosted geozones")
}
func ErrNoChains(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeNoChains, "validator must stake with hosted blockchains")
}
func ErrNilProviderAddr(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidInput, "provider address is nil")
}
func ErrProviderStatus(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidStatus, "provider status is not valid")
}
func ErrNoProviderFound(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidProvider, "provider does not exist for that address")
}

func ErrBadStakeAmount(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidStakeAmount, "the stake amount is invalid")
}

func ErrNoServicers(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeNoServicers, "the number of servicers is zero")
}

func ErrNotEnoughCoins(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeNotEnoughCoins, "provider does not have enough coins in their account")
}

func ErrMaxProviders(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeMaxProviders, "the threshold of the amount of providers authorized ")
}

func ErrMinimumStake(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeMinimumStake, "provider isn't staking above the minimum")
}

func ErrProviderPubKeyExists(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidProvider, "provider already exist for this pubkey, must use new provider pubkey")
}

func ErrProviderPubKeyTypeNotSupported(codespace sdk.CodespaceType, keyType string, supportedTypes []string) sdk.Error {
	msg := fmt.Sprintf("provider pubkey type %s is not supported, must use %s", keyType, strings.Join(supportedTypes, ","))
	return sdk.NewError(codespace, CodeInvalidProvider, msg)
}

func ErrNoProviderForAddress(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidProvider, "that address is not associated with any known provider")
}

func ErrBadProviderAddr(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidProvider, "provider does not exist for that address")
}

func ErrProviderJailed(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeProviderJailed, "provider still jailed, cannot yet be unjailed")
}

func ErrProviderNotJailed(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeProviderNotJailed, "provider not jailed, cannot be unjailed")
}

func ErrMissingProviderStake(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeMissingSelfDelegation, "provider has no stake; cannot be unjailed")
}

func ErrStakeTooLow(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeProviderNotJailed, "provider's self delegation less than min stake, cannot be unjailed")
}

func ErrInvalidNetworkIdentifier(codespace sdk.CodespaceType, err error) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidNetworkID, "the providers network identifier is not valid: "+err.Error())
}

func ErrInvalidGeoZoneIdentifier(codespace sdk.CodespaceType, err error) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidGeoZone, "the providers geo zone identifier is not valid: "+err.Error())
}

func ErrMinimumEditStake(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeMinimumEditStake, "provider must edit stake with a stake greater than or equal to current stake")
}

func ErrNumServicers(Codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(Codespace, CodeNumServicers, "Number of servicer's out of range")
}
