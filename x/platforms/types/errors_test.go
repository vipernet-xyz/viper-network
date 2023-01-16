package types

import (
	"fmt"
	"strings"
	"testing"

	sdk "github.com/vipernet-xyz/viper-network/types"
)

var codespace = sdk.CodespaceType("platform")

func TestError_ErrNoChains(t *testing.T) {
	type args struct {
		codespace sdk.CodespaceType
	}
	tests := []struct {
		name string
		args
		want sdk.Error
	}{
		{
			name: "returns error for stake on unhosted blockchain",
			args: args{codespace},
			want: sdk.NewError(codespace, sdk.CodeType(116), "validator must stake with hosted blockchains"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrNoChains(tt.args.codespace); got.Error() != tt.want.Error() {
				t.Errorf("ErrorNoChains(): returns %v but want %v", got, tt.want)
			}
		})
	}
}
func TestError_ErrNilPlatformAddr(t *testing.T) {
	type args struct {
		codespace sdk.CodespaceType
	}
	tests := []struct {
		name string
		args
		want sdk.Error
	}{
		{
			name: "returns error platform address is nil",
			args: args{codespace},
			want: sdk.NewError(codespace, sdk.CodeType(103), "platform address is nil"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrNilPlatformAddr(tt.args.codespace); got.Error() != tt.want.Error() {
				t.Errorf("ErrNilPlatformAddr(): returns %v but want %v", got, tt.want)
			}
		})
	}
}
func TestError_ErrPlatformlicaitonStatus(t *testing.T) {
	type args struct {
		codespace sdk.CodespaceType
	}
	tests := []struct {
		name string
		args
		want sdk.Error
	}{
		{
			name: "returns error status is invalid",
			args: args{codespace},
			want: sdk.NewError(codespace, sdk.CodeType(110), "platform status is not valid"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrPlatformStatus(tt.args.codespace); got.Error() != tt.want.Error() {
				t.Errorf("ErrPlatformStatus(): returns %v but want %v", got, tt.want)
			}
		})
	}
}
func TestError_ErrNoPlatformFound(t *testing.T) {
	type args struct {
		codespace sdk.CodespaceType
	}
	tests := []struct {
		name string
		args
		want sdk.Error
	}{
		{
			name: "returns error platform not found for address",
			args: args{codespace},
			want: sdk.NewError(codespace, sdk.CodeType(101), "platform does not exist for that address"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrNoPlatformFound(tt.args.codespace); got.Error() != tt.want.Error() {
				t.Errorf("ErrNoPlatformFound(): returns %v but want %v", got, tt.want)
			}
		})
	}
}
func TestError_ErrBadStakeAmount(t *testing.T) {
	type args struct {
		codespace sdk.CodespaceType
	}
	tests := []struct {
		name string
		args
		want sdk.Error
	}{
		{
			name: "returns error stake amount is invalid",
			args: args{codespace},
			want: sdk.NewError(codespace, sdk.CodeType(115), "the stake amount is invalid"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrBadStakeAmount(tt.args.codespace); got.Error() != tt.want.Error() {
				t.Errorf("ErrBadStakeAmount(): returns %v but want %v", got, tt.want)
			}
		})
	}
}
func TestError_ErrNotEnoughCoins(t *testing.T) {
	type args struct {
		codespace sdk.CodespaceType
	}
	tests := []struct {
		name string
		args
		want sdk.Error
	}{
		{
			name: "returns error platform does not have enough coins",
			args: args{codespace},
			want: sdk.NewError(codespace, sdk.CodeType(112), "platform does not have enough coins in their account"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrNotEnoughCoins(tt.args.codespace); got.Error() != tt.want.Error() {
				t.Errorf("ErrNotEnoughCoins(): returns %v but want %v", got, tt.want)
			}
		})
	}
}
func TestError_ErrPlatformlicaitonPubKeyExists(t *testing.T) {
	type args struct {
		codespace sdk.CodespaceType
	}
	tests := []struct {
		name string
		args
		want sdk.Error
	}{
		{
			name: "returns error platform already exists for public key",
			args: args{codespace},
			want: sdk.NewError(codespace, sdk.CodeType(101), "platform already exist for this pubkey, must use new platform pubkey"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrPlatformPubKeyExists(tt.args.codespace); got.Error() != tt.want.Error() {
				t.Errorf("ErrPlatformPubKeyExists(): returns %v but want %v", got, tt.want)
			}
		})
	}
}
func TestError_ErrMinimumStake(t *testing.T) {
	type args struct {
		codespace sdk.CodespaceType
	}
	tests := []struct {
		name string
		args
		want sdk.Error
	}{
		{
			name: "returns error platform staking lower than minimum",
			args: args{codespace},
			want: sdk.NewError(codespace, sdk.CodeType(111), "platform isn't staking above the minimum"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrMinimumStake(tt.args.codespace); got.Error() != tt.want.Error() {
				t.Errorf("ErrMinimumStake(): returns %v but want %v", got, tt.want)
			}
		})
	}
}
func TestError_ErrNoPlatformFoundForAddress(t *testing.T) {
	type args struct {
		codespace sdk.CodespaceType
	}
	tests := []struct {
		name string
		args
		want sdk.Error
	}{
		{
			name: "returns error platformlicaiton not found",
			args: args{codespace},
			want: sdk.NewError(codespace, sdk.CodeType(101), "that address is not associated with any known platform"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrNoPlatformForAddress(tt.args.codespace); got.Error() != tt.want.Error() {
				t.Errorf("ErrNoPlatformForAddress(): returns %v but want %v", got, tt.want)
			}
		})
	}
}
func TestError_ErrBadPlatformlicaitonAddr(t *testing.T) {
	type args struct {
		codespace sdk.CodespaceType
	}
	tests := []struct {
		name string
		args
		want sdk.Error
	}{
		{
			name: "returns error platform does not exist for address",
			args: args{codespace},
			want: sdk.NewError(codespace, sdk.CodeType(101), "platform does not exist for that address"),
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			if got := ErrBadPlatformAddr(tt.args.codespace); got.Error() != tt.want.Error() {
				t.Errorf("ErrBadPlatformlicaitonAddr(): returns %v but want %v", got, tt.want)
			}
		})
	}
}
func TestError_ErrPlatformNotJailed(t *testing.T) {
	type args struct {
		codespace sdk.CodespaceType
	}
	tests := []struct {
		name string
		args
		want sdk.Error
	}{
		{
			name: "returns error platform is not jailed",
			args: args{codespace},
			want: sdk.NewError(codespace, sdk.CodeType(105), "platform not jailed, cannot be unjailed"),
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			if got := ErrPlatformNotJailed(tt.args.codespace); got.Error() != tt.want.Error() {
				t.Errorf("ErrPlatformNotJailed(): returns %v but want %v", got, tt.want)
			}
		})
	}
}
func TestError_ErrMissingPlatformStake(t *testing.T) {
	type args struct {
		codespace sdk.CodespaceType
	}
	tests := []struct {
		name string
		args
		want sdk.Error
	}{
		{
			name: "returns error platform has no stake",
			args: args{codespace},
			want: sdk.NewError(codespace, sdk.CodeType(106), "platform has no stake; cannot be unjailed"),
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			if got := ErrMissingPlatformStake(tt.args.codespace); got.Error() != tt.want.Error() {
				t.Errorf("ErrMissingPlatformStake(): returns %v but want %v", got, tt.want)
			}
		})
	}
}
func TestError_ErrStakeTooLow(t *testing.T) {
	type args struct {
		codespace sdk.CodespaceType
	}
	tests := []struct {
		name string
		args
		want sdk.Error
	}{
		{
			name: "returns error platform stke lower than delegation",
			args: args{codespace},
			want: sdk.NewError(codespace, sdk.CodeType(105), "platform's self delegation less than min stake, cannot be unjailed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrStakeTooLow(tt.args.codespace); got.Error() != tt.want.Error() {
				t.Errorf("ErrStakeTooLow(): returns %v but want %v", got, tt.want)
			}
		})
	}
}
func TestError_ErrPlatformPubKeyTypeNotSupported(t *testing.T) {
	type args struct {
		codespace sdk.CodespaceType
		types     []string
		keyType   string
	}
	tests := []struct {
		name string
		args
		want sdk.Error
	}{
		{
			name: "returns error platform does not exist",
			args: args{codespace, []string{"ed25519", "blake2b"}, "int"},
			want: sdk.NewError(
				codespace,
				sdk.CodeType(101),
				fmt.Sprintf("platform pubkey type %s is not supported, must use %s", "int", strings.Join([]string{"ed25519", "blake2b"}, ","))),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrPlatformPubKeyTypeNotSupported(tt.args.codespace, tt.args.keyType, tt.args.types); got.Error() != tt.want.Error() {
				t.Errorf("ErrPlatformPubKeyTypeNotSupported(): returns %v but want %v", got, tt.want)
			}
		})
	}
}
