package types

import (
	"fmt"
	"strings"
	"testing"

	sdk "github.com/vipernet-xyz/viper-network/types"
)

var codespace = sdk.CodespaceType("provider")

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
func TestError_ErrNilProviderAddr(t *testing.T) {
	type args struct {
		codespace sdk.CodespaceType
	}
	tests := []struct {
		name string
		args
		want sdk.Error
	}{
		{
			name: "returns error provider address is nil",
			args: args{codespace},
			want: sdk.NewError(codespace, sdk.CodeType(103), "provider address is nil"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrNilProviderAddr(tt.args.codespace); got.Error() != tt.want.Error() {
				t.Errorf("ErrNilProviderAddr(): returns %v but want %v", got, tt.want)
			}
		})
	}
}
func TestError_ErrProviderStatus(t *testing.T) {
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
			want: sdk.NewError(codespace, sdk.CodeType(110), "provider status is not valid"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrProviderStatus(tt.args.codespace); got.Error() != tt.want.Error() {
				t.Errorf("ErrProviderStatus(): returns %v but want %v", got, tt.want)
			}
		})
	}
}
func TestError_ErrNoProviderFound(t *testing.T) {
	type args struct {
		codespace sdk.CodespaceType
	}
	tests := []struct {
		name string
		args
		want sdk.Error
	}{
		{
			name: "returns error provider not found for address",
			args: args{codespace},
			want: sdk.NewError(codespace, sdk.CodeType(101), "provider does not exist for that address"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrNoProviderFound(tt.args.codespace); got.Error() != tt.want.Error() {
				t.Errorf("ErrNoProviderFound(): returns %v but want %v", got, tt.want)
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
			name: "returns error provider does not have enough coins",
			args: args{codespace},
			want: sdk.NewError(codespace, sdk.CodeType(112), "provider does not have enough coins in their account"),
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
func TestError_ErrProviderPubKeyExists(t *testing.T) {
	type args struct {
		codespace sdk.CodespaceType
	}
	tests := []struct {
		name string
		args
		want sdk.Error
	}{
		{
			name: "returns error provider already exists for public key",
			args: args{codespace},
			want: sdk.NewError(codespace, sdk.CodeType(101), "provider already exist for this pubkey, must use new provider pubkey"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrProviderPubKeyExists(tt.args.codespace); got.Error() != tt.want.Error() {
				t.Errorf("ErrProviderPubKeyExists(): returns %v but want %v", got, tt.want)
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
			name: "returns error provider staking lower than minimum",
			args: args{codespace},
			want: sdk.NewError(codespace, sdk.CodeType(111), "provider isn't staking above the minimum"),
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
func TestError_ErrNoProviderFoundForAddress(t *testing.T) {
	type args struct {
		codespace sdk.CodespaceType
	}
	tests := []struct {
		name string
		args
		want sdk.Error
	}{
		{
			name: "returns error providerlicaiton not found",
			args: args{codespace},
			want: sdk.NewError(codespace, sdk.CodeType(101), "that address is not associated with any known provider"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrNoProviderForAddress(tt.args.codespace); got.Error() != tt.want.Error() {
				t.Errorf("ErrNoProviderForAddress(): returns %v but want %v", got, tt.want)
			}
		})
	}
}
func TestError_ErrBadProviderAddr(t *testing.T) {
	type args struct {
		codespace sdk.CodespaceType
	}
	tests := []struct {
		name string
		args
		want sdk.Error
	}{
		{
			name: "returns error provider does not exist for address",
			args: args{codespace},
			want: sdk.NewError(codespace, sdk.CodeType(101), "provider does not exist for that address"),
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			if got := ErrBadProviderAddr(tt.args.codespace); got.Error() != tt.want.Error() {
				t.Errorf("ErrBadProviderAddr(): returns %v but want %v", got, tt.want)
			}
		})
	}
}
func TestError_ErrProviderNotJailed(t *testing.T) {
	type args struct {
		codespace sdk.CodespaceType
	}
	tests := []struct {
		name string
		args
		want sdk.Error
	}{
		{
			name: "returns error provider is not jailed",
			args: args{codespace},
			want: sdk.NewError(codespace, sdk.CodeType(105), "provider not jailed, cannot be unjailed"),
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			if got := ErrProviderNotJailed(tt.args.codespace); got.Error() != tt.want.Error() {
				t.Errorf("ErrProviderNotJailed(): returns %v but want %v", got, tt.want)
			}
		})
	}
}
func TestError_ErrMissingProviderStake(t *testing.T) {
	type args struct {
		codespace sdk.CodespaceType
	}
	tests := []struct {
		name string
		args
		want sdk.Error
	}{
		{
			name: "returns error provider has no stake",
			args: args{codespace},
			want: sdk.NewError(codespace, sdk.CodeType(106), "provider has no stake; cannot be unjailed"),
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			if got := ErrMissingProviderStake(tt.args.codespace); got.Error() != tt.want.Error() {
				t.Errorf("ErrMissingProviderStake(): returns %v but want %v", got, tt.want)
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
			name: "returns error provider stke lower than delegation",
			args: args{codespace},
			want: sdk.NewError(codespace, sdk.CodeType(105), "provider's self delegation less than min stake, cannot be unjailed"),
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
func TestError_ErrProviderPubKeyTypeNotSupported(t *testing.T) {
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
			name: "returns error provider does not exist",
			args: args{codespace, []string{"ed25519", "blake2b"}, "int"},
			want: sdk.NewError(
				codespace,
				sdk.CodeType(101),
				fmt.Sprintf("provider pubkey type %s is not supported, must use %s", "int", strings.Join([]string{"ed25519", "blake2b"}, ","))),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrProviderPubKeyTypeNotSupported(tt.args.codespace, tt.args.keyType, tt.args.types); got.Error() != tt.want.Error() {
				t.Errorf("ErrProviderPubKeyTypeNotSupported(): returns %v but want %v", got, tt.want)
			}
		})
	}
}
