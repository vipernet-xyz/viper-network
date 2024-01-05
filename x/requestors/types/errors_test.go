package types

import (
	"fmt"
	"strings"
	"testing"

	sdk "github.com/vipernet-xyz/viper-network/types"
)

var codespace = sdk.CodespaceType("requestor")

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
func TestError_ErrNilRequestorAddr(t *testing.T) {
	type args struct {
		codespace sdk.CodespaceType
	}
	tests := []struct {
		name string
		args
		want sdk.Error
	}{
		{
			name: "returns error requestor address is nil",
			args: args{codespace},
			want: sdk.NewError(codespace, sdk.CodeType(103), "requestor address is nil"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrNilRequestorAddr(tt.args.codespace); got.Error() != tt.want.Error() {
				t.Errorf("ErrNilRequestorAddr(): returns %v but want %v", got, tt.want)
			}
		})
	}
}
func TestError_ErrRequestorStatus(t *testing.T) {
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
			want: sdk.NewError(codespace, sdk.CodeType(110), "requestor status is not valid"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrRequestorStatus(tt.args.codespace); got.Error() != tt.want.Error() {
				t.Errorf("ErrRequestorStatus(): returns %v but want %v", got, tt.want)
			}
		})
	}
}
func TestError_ErrNoRequestorFound(t *testing.T) {
	type args struct {
		codespace sdk.CodespaceType
	}
	tests := []struct {
		name string
		args
		want sdk.Error
	}{
		{
			name: "returns error requestor not found for address",
			args: args{codespace},
			want: sdk.NewError(codespace, sdk.CodeType(101), "requestor does not exist for that address"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrNoRequestorFound(tt.args.codespace); got.Error() != tt.want.Error() {
				t.Errorf("ErrNoRequestorFound(): returns %v but want %v", got, tt.want)
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
			name: "returns error requestor does not have enough coins",
			args: args{codespace},
			want: sdk.NewError(codespace, sdk.CodeType(112), "requestor does not have enough coins in their account"),
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
func TestError_ErrRequestorPubKeyExists(t *testing.T) {
	type args struct {
		codespace sdk.CodespaceType
	}
	tests := []struct {
		name string
		args
		want sdk.Error
	}{
		{
			name: "returns error requestor already exists for public key",
			args: args{codespace},
			want: sdk.NewError(codespace, sdk.CodeType(101), "requestor already exist for this pubkey, must use new requestor pubkey"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrRequestorPubKeyExists(tt.args.codespace); got.Error() != tt.want.Error() {
				t.Errorf("ErrRequestorPubKeyExists(): returns %v but want %v", got, tt.want)
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
			name: "returns error requestor staking lower than minimum",
			args: args{codespace},
			want: sdk.NewError(codespace, sdk.CodeType(111), "requestor isn't staking above the minimum"),
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
func TestError_ErrNoRequestorFoundForAddress(t *testing.T) {
	type args struct {
		codespace sdk.CodespaceType
	}
	tests := []struct {
		name string
		args
		want sdk.Error
	}{
		{
			name: "returns error requestorlicaiton not found",
			args: args{codespace},
			want: sdk.NewError(codespace, sdk.CodeType(101), "that address is not associated with any known requestor"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrNoRequestorForAddress(tt.args.codespace); got.Error() != tt.want.Error() {
				t.Errorf("ErrNoRequestorForAddress(): returns %v but want %v", got, tt.want)
			}
		})
	}
}
func TestError_ErrBadRequestorAddr(t *testing.T) {
	type args struct {
		codespace sdk.CodespaceType
	}
	tests := []struct {
		name string
		args
		want sdk.Error
	}{
		{
			name: "returns error requestor does not exist for address",
			args: args{codespace},
			want: sdk.NewError(codespace, sdk.CodeType(101), "requestor does not exist for that address"),
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			if got := ErrBadRequestorAddr(tt.args.codespace); got.Error() != tt.want.Error() {
				t.Errorf("ErrBadRequestorAddr(): returns %v but want %v", got, tt.want)
			}
		})
	}
}
func TestError_ErrRequestorNotJailed(t *testing.T) {
	type args struct {
		codespace sdk.CodespaceType
	}
	tests := []struct {
		name string
		args
		want sdk.Error
	}{
		{
			name: "returns error requestor is not jailed",
			args: args{codespace},
			want: sdk.NewError(codespace, sdk.CodeType(105), "requestor not jailed, cannot be unjailed"),
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			if got := ErrRequestorNotJailed(tt.args.codespace); got.Error() != tt.want.Error() {
				t.Errorf("ErrRequestorNotJailed(): returns %v but want %v", got, tt.want)
			}
		})
	}
}
func TestError_ErrMissingRequestorStake(t *testing.T) {
	type args struct {
		codespace sdk.CodespaceType
	}
	tests := []struct {
		name string
		args
		want sdk.Error
	}{
		{
			name: "returns error requestor has no stake",
			args: args{codespace},
			want: sdk.NewError(codespace, sdk.CodeType(106), "requestor has no stake; cannot be unjailed"),
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			if got := ErrMissingRequestorStake(tt.args.codespace); got.Error() != tt.want.Error() {
				t.Errorf("ErrMissingRequestorStake(): returns %v but want %v", got, tt.want)
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
			name: "returns error requestor stke lower than delegation",
			args: args{codespace},
			want: sdk.NewError(codespace, sdk.CodeType(105), "requestor's self delegation less than min stake, cannot be unjailed"),
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
func TestError_ErrRequestorPubKeyTypeNotSupported(t *testing.T) {
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
			name: "returns error requestor does not exist",
			args: args{codespace, []string{"ed25519", "blake2b"}, "int"},
			want: sdk.NewError(
				codespace,
				sdk.CodeType(101),
				fmt.Sprintf("requestor pubkey type %s is not supported, must use %s", "int", strings.Join([]string{"ed25519", "blake2b"}, ","))),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrRequestorPubKeyTypeNotSupported(tt.args.codespace, tt.args.keyType, tt.args.types); got.Error() != tt.want.Error() {
				t.Errorf("ErrRequestorPubKeyTypeNotSupported(): returns %v but want %v", got, tt.want)
			}
		})
	}
}
