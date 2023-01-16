package types

import (
	"fmt"
	"strings"

	sdk "github.com/vipernet-xyz/viper-network/types"
)

type CodeType = sdk.CodeType

const (
	DefaultCodespace          sdk.CodespaceType = ModuleName
	CodeInvalidPlatform       CodeType          = 101
	CodeInvalidInput          CodeType          = 103
	CodePlatformJailed        CodeType          = 104
	CodePlatformNotJailed     CodeType          = 105
	CodeMissingSelfDelegation CodeType          = 106
	CodeInvalidStatus         CodeType          = 110
	CodeMinimumStake          CodeType          = 111
	CodeNotEnoughCoins        CodeType          = 112
	CodeInvalidStakeAmount    CodeType          = 115
	CodeNoChains              CodeType          = 116
	CodeInvalidNetworkID      CodeType          = 117
	CodeTooManyChains         CodeType          = 118
	CodeMaxPlatforms          CodeType          = 119
	CodeMinimumEditStake      CodeType          = 120
)

func ErrTooManyChains(Codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(Codespace, CodeTooManyChains, "platform staking for too many chains")
}

func ErrNoChains(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeNoChains, "validator must stake with hosted blockchains")
}
func ErrNilPlatformAddr(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidInput, "platform address is nil")
}
func ErrPlatformStatus(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidStatus, "platform status is not valid")
}
func ErrNoPlatformFound(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidPlatform, "platform does not exist for that address")
}

func ErrBadStakeAmount(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidStakeAmount, "the stake amount is invalid")
}

func ErrNotEnoughCoins(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeNotEnoughCoins, "platform does not have enough coins in their account")
}

func ErrMaxPlatforms(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeMaxPlatforms, "the threshold of the amount of platforms authorized ")
}

func ErrMinimumStake(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeMinimumStake, "platform isn't staking above the minimum")
}

func ErrPlatformPubKeyExists(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidPlatform, "platform already exist for this pubkey, must use new platform pubkey")
}

func ErrPlatformPubKeyTypeNotSupported(codespace sdk.CodespaceType, keyType string, supportedTypes []string) sdk.Error {
	msg := fmt.Sprintf("platform pubkey type %s is not supported, must use %s", keyType, strings.Join(supportedTypes, ","))
	return sdk.NewError(codespace, CodeInvalidPlatform, msg)
}

func ErrNoPlatformForAddress(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidPlatform, "that address is not associated with any known platform")
}

func ErrBadPlatformAddr(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidPlatform, "platform does not exist for that address")
}

func ErrPlatformJailed(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodePlatformJailed, "platform still jailed, cannot yet be unjailed")
}

func ErrPlatformNotJailed(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodePlatformNotJailed, "platform not jailed, cannot be unjailed")
}

func ErrMissingPlatformStake(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeMissingSelfDelegation, "platform has no stake; cannot be unjailed")
}

func ErrStakeTooLow(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodePlatformNotJailed, "platform's self delegation less than min stake, cannot be unjailed")
}

func ErrInvalidNetworkIdentifier(codespace sdk.CodespaceType, err error) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidNetworkID, "the platforms network identifier is not valid: "+err.Error())
}

func ErrMinimumEditStake(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeMinimumEditStake, "platform must edit stake with a stake greater than or equal to current stake")
}
