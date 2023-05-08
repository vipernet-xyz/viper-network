package types

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"

	"github.com/vipernet-xyz/viper-network/codec/types"

	"github.com/vipernet-xyz/viper-network/codec"
	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
)

var msgProviderStake MsgStake
var msgBeginProviderUnstake MsgBeginUnstake
var msgProviderUnjail MsgUnjail
var pk crypto.Ed25519PublicKey

func init() {
	var pub crypto.Ed25519PublicKey
	_, err := rand.Read(pub[:])
	if err != nil {
		_ = err
	}

	pk = pub

	cdc := codec.NewCodec(types.NewInterfaceRegistry())
	RegisterCodec(cdc)
	crypto.RegisterAmino(cdc.AminoCodec().Amino)

	msgProviderStake = MsgStake{
		PubKey: pub,
		Chains: []string{"0001"},
		Value:  sdk.NewInt(10),
	}
	msgProviderUnjail = MsgUnjail{sdk.Address(pub.Address())}
	msgBeginProviderUnstake = MsgBeginUnstake{sdk.Address(pub.Address())}
}

func TestMsgProvider_GetSigners(t *testing.T) {
	type args struct {
		msgProviderStake MsgStake
	}
	tests := []struct {
		name string
		args
		want []sdk.Address
	}{
		{
			name: "return signers",
			args: args{msgProviderStake},
			want: []sdk.Address{sdk.Address(pk.Address())},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgProviderStake.GetSigners(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSigners() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestMsgProvider_GetSignBytes(t *testing.T) {
	type args struct {
		msgProviderStake MsgStake
	}
	res, err := ModuleCdc.MarshalJSON(&msgProviderStake)
	res = sdk.MustSortJSON(res)
	if err != nil {
		panic(err)
	}
	tests := []struct {
		name string
		args
		want []byte
	}{
		{
			name: "return signers",
			args: args{msgProviderStake},
			want: res,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgProviderStake.GetSignBytes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSigners() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestMsgProvider_Route(t *testing.T) {
	type args struct {
		msgProviderStake MsgStake
	}
	tests := []struct {
		name string
		args
		want string
	}{
		{
			name: "return signers",
			args: args{msgProviderStake},
			want: RouterKey,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgProviderStake.Route(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSigners() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestMsgProvider_Type(t *testing.T) {
	type args struct {
		msgProviderStake MsgStake
	}
	tests := []struct {
		name string
		args
		want string
	}{
		{
			name: "return signers",
			args: args{msgProviderStake},
			want: MsgProviderStakeName,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgProviderStake.Type(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSigners() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestMsgProvider_ValidateBasic(t *testing.T) {
	type args struct {
		msgProviderStake MsgStake
	}
	tests := []struct {
		name string
		args
		want sdk.Error
		msg  string
	}{
		{
			name: "errs if no Address",
			args: args{MsgStake{}},
			want: ErrNilProviderAddr(DefaultCodespace),
		},
		{
			name: "errs if no stake lower than zero",
			args: args{MsgStake{PubKey: msgProviderStake.PubKey, Value: sdk.NewInt(-1)}},
			want: ErrBadStakeAmount(DefaultCodespace),
		},
		{
			name: "errs if no native chains supported",
			args: args{MsgStake{PubKey: msgProviderStake.PubKey, Value: sdk.NewInt(1), Chains: []string{}}},
			want: ErrNoChains(DefaultCodespace),
		},
		{
			name: "returns err",
			args: args{MsgStake{PubKey: msgProviderStake.PubKey, Value: msgProviderStake.Value, Chains: []string{"aaaaaa"}}},
			want: ErrInvalidNetworkIdentifier("provider", fmt.Errorf("net id length is > 2")),
		},
		{
			name: "returns nil if valid address",
			args: args{msgProviderStake},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgProviderStake.ValidateBasic(); got != nil {
				if !reflect.DeepEqual(got.Error(), tt.want.Error()) {
					t.Errorf("ValidatorBasic() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestMsgBeginProviderUnstake_GetSigners(t *testing.T) {
	type args struct {
		msgBeginProviderUnstake MsgBeginUnstake
	}
	tests := []struct {
		name string
		args
		want []sdk.Address
	}{
		{
			name: "return signers",
			args: args{msgBeginProviderUnstake},
			want: []sdk.Address{sdk.Address(pk.Address())},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgBeginProviderUnstake.GetSigners(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSigners() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestMsgBeginProviderUnstake_GetSignBytes(t *testing.T) {
	type args struct {
		msgBeginProviderUnstake MsgBeginUnstake
	}
	res, err := ModuleCdc.MarshalJSON(&msgBeginProviderUnstake)
	if err != nil {
		panic(err)
	}
	tests := []struct {
		name string
		args
		want []byte
	}{
		{
			name: "return signers",
			args: args{msgBeginProviderUnstake},
			want: res,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgBeginProviderUnstake.GetSignBytes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSignBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestMsgBeginProviderUnstake_Route(t *testing.T) {
	type args struct {
		msgBeginProviderUnstake MsgBeginUnstake
	}
	tests := []struct {
		name string
		args
		want string
	}{
		{
			name: "return signers",
			args: args{msgBeginProviderUnstake},
			want: RouterKey,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgBeginProviderUnstake.Route(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSigners() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestMsgBeginProviderUnstake_Type(t *testing.T) {
	type args struct {
		msgBeginProviderUnstake MsgBeginUnstake
	}
	tests := []struct {
		name string
		args
		want string
	}{
		{
			name: "return signers",
			args: args{msgBeginProviderUnstake},
			want: MsgProviderUnstakeName,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgBeginProviderUnstake.Type(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSigners() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestMsgBeginProviderUnstake_ValidateBasic(t *testing.T) {
	type args struct {
		msgBeginProviderUnstake MsgBeginUnstake
	}
	tests := []struct {
		name string
		args
		want sdk.Error
		msg  string
	}{
		{
			name: "errs if no Address",
			args: args{MsgBeginUnstake{}},
			want: ErrNilProviderAddr(DefaultCodespace),
		},
		{
			name: "returns nil if valid address",
			args: args{msgBeginProviderUnstake},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgBeginProviderUnstake.ValidateBasic(); got != nil {
				if !reflect.DeepEqual(got.Error(), tt.want.Error()) {
					t.Errorf("ValidatorBasic() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestMsgProviderUnjail_Route(t *testing.T) {
	type args struct {
		msgProviderUnjail MsgUnjail
	}
	tests := []struct {
		name string
		args
		want string
	}{
		{
			name: "return signers",
			args: args{msgProviderUnjail},
			want: RouterKey,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgProviderUnjail.Route(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSigners() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestMsgProviderUnjail_Type(t *testing.T) {
	type args struct {
		msgProviderUnjail MsgUnjail
	}
	tests := []struct {
		name string
		args
		want string
	}{
		{
			name: "return signers",
			args: args{msgProviderUnjail},
			want: MsgProviderUnjailName,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgProviderUnjail.Type(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSigners() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestMsgProviderUnjail_GetSigners(t *testing.T) {
	type args struct {
		msgProviderUnjail MsgUnjail
	}
	tests := []struct {
		name string
		args
		want []sdk.Address
	}{
		{
			name: "return signers",
			args: args{msgProviderUnjail},
			want: []sdk.Address{sdk.Address(msgProviderUnjail.ProviderAddr)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgProviderUnjail.GetSigners(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSigners() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestMsgProviderUnjail_GetSignBytes(t *testing.T) {
	type args struct {
		msgProviderUnjail MsgUnjail
	}
	res, err := ModuleCdc.MarshalJSON(&msgProviderUnjail)
	if err != nil {
		panic(err)
	}
	tests := []struct {
		name string
		args
		want []byte
	}{
		{
			name: "return signers",
			args: args{msgProviderUnjail},
			want: res,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgProviderUnjail.GetSignBytes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSigners() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestMsgProviderUnjail_ValidateBasic(t *testing.T) {
	type args struct {
		msgProviderUnjail MsgUnjail
	}
	tests := []struct {
		name string
		args
		want sdk.Error
	}{
		{
			name: "errs if no Address",
			args: args{MsgUnjail{}},
			want: ErrBadProviderAddr(DefaultCodespace),
		},
		{
			name: "returns nil if valid address",
			args: args{msgProviderUnjail},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgProviderUnjail.ValidateBasic(); got != nil {
				if !reflect.DeepEqual(got.Error(), tt.want.Error()) {
					t.Errorf("GetSigners() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
