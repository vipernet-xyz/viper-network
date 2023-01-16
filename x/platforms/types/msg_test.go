package types

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"

	"github.com/vipernet-xyz/viper-network/codec/types"

	"github.com/vipernet-xyz/viper-network/codec"
	"github.com/vipernet-xyz/viper-network/crypto"
	sdk "github.com/vipernet-xyz/viper-network/types"
)

var msgPlatformStake MsgStake
var msgBeginPlatformUnstake MsgBeginUnstake
var msgPlatformUnjail MsgUnjail
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

	msgPlatformStake = MsgStake{
		PubKey: pub,
		Chains: []string{"0001"},
		Value:  sdk.NewInt(10),
	}
	msgPlatformUnjail = MsgUnjail{sdk.Address(pub.Address())}
	msgBeginPlatformUnstake = MsgBeginUnstake{sdk.Address(pub.Address())}
}

func TestMsgPlatform_GetSigners(t *testing.T) {
	type args struct {
		msgPlatformStake MsgStake
	}
	tests := []struct {
		name string
		args
		want []sdk.Address
	}{
		{
			name: "return signers",
			args: args{msgPlatformStake},
			want: []sdk.Address{sdk.Address(pk.Address())},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgPlatformStake.GetSigners(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSigners() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestMsgPlatform_GetSignBytes(t *testing.T) {
	type args struct {
		msgPlatformStake MsgStake
	}
	res, err := ModuleCdc.MarshalJSON(&msgPlatformStake)
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
			args: args{msgPlatformStake},
			want: res,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgPlatformStake.GetSignBytes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSigners() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestMsgPlatform_Route(t *testing.T) {
	type args struct {
		msgPlatformStake MsgStake
	}
	tests := []struct {
		name string
		args
		want string
	}{
		{
			name: "return signers",
			args: args{msgPlatformStake},
			want: RouterKey,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgPlatformStake.Route(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSigners() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestMsgPlatform_Type(t *testing.T) {
	type args struct {
		msgPlatformStake MsgStake
	}
	tests := []struct {
		name string
		args
		want string
	}{
		{
			name: "return signers",
			args: args{msgPlatformStake},
			want: MsgPlatformStakeName,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgPlatformStake.Type(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSigners() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestMsgPlatform_ValidateBasic(t *testing.T) {
	type args struct {
		msgPlatformStake MsgStake
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
			want: ErrNilPlatformAddr(DefaultCodespace),
		},
		{
			name: "errs if no stake lower than zero",
			args: args{MsgStake{PubKey: msgPlatformStake.PubKey, Value: sdk.NewInt(-1)}},
			want: ErrBadStakeAmount(DefaultCodespace),
		},
		{
			name: "errs if no native chains supported",
			args: args{MsgStake{PubKey: msgPlatformStake.PubKey, Value: sdk.NewInt(1), Chains: []string{}}},
			want: ErrNoChains(DefaultCodespace),
		},
		{
			name: "returns err",
			args: args{MsgStake{PubKey: msgPlatformStake.PubKey, Value: msgPlatformStake.Value, Chains: []string{"aaaaaa"}}},
			want: ErrInvalidNetworkIdentifier("platform", fmt.Errorf("net id length is > 2")),
		},
		{
			name: "returns nil if valid address",
			args: args{msgPlatformStake},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgPlatformStake.ValidateBasic(); got != nil {
				if !reflect.DeepEqual(got.Error(), tt.want.Error()) {
					t.Errorf("ValidatorBasic() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestMsgBeginPlatformUnstake_GetSigners(t *testing.T) {
	type args struct {
		msgBeginPlatformUnstake MsgBeginUnstake
	}
	tests := []struct {
		name string
		args
		want []sdk.Address
	}{
		{
			name: "return signers",
			args: args{msgBeginPlatformUnstake},
			want: []sdk.Address{sdk.Address(pk.Address())},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgBeginPlatformUnstake.GetSigners(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSigners() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestMsgBeginPlatformUnstake_GetSignBytes(t *testing.T) {
	type args struct {
		msgBeginPlatformUnstake MsgBeginUnstake
	}
	res, err := ModuleCdc.MarshalJSON(&msgBeginPlatformUnstake)
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
			args: args{msgBeginPlatformUnstake},
			want: res,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgBeginPlatformUnstake.GetSignBytes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSignBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestMsgBeginPlatformUnstake_Route(t *testing.T) {
	type args struct {
		msgBeginPlatformUnstake MsgBeginUnstake
	}
	tests := []struct {
		name string
		args
		want string
	}{
		{
			name: "return signers",
			args: args{msgBeginPlatformUnstake},
			want: RouterKey,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgBeginPlatformUnstake.Route(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSigners() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestMsgBeginPlatformUnstake_Type(t *testing.T) {
	type args struct {
		msgBeginPlatformUnstake MsgBeginUnstake
	}
	tests := []struct {
		name string
		args
		want string
	}{
		{
			name: "return signers",
			args: args{msgBeginPlatformUnstake},
			want: MsgPlatformUnstakeName,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgBeginPlatformUnstake.Type(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSigners() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestMsgBeginPlatformUnstake_ValidateBasic(t *testing.T) {
	type args struct {
		msgBeginPlatformUnstake MsgBeginUnstake
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
			want: ErrNilPlatformAddr(DefaultCodespace),
		},
		{
			name: "returns nil if valid address",
			args: args{msgBeginPlatformUnstake},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgBeginPlatformUnstake.ValidateBasic(); got != nil {
				if !reflect.DeepEqual(got.Error(), tt.want.Error()) {
					t.Errorf("ValidatorBasic() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestMsgPlatformUnjail_Route(t *testing.T) {
	type args struct {
		msgPlatformUnjail MsgUnjail
	}
	tests := []struct {
		name string
		args
		want string
	}{
		{
			name: "return signers",
			args: args{msgPlatformUnjail},
			want: RouterKey,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgPlatformUnjail.Route(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSigners() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestMsgPlatformUnjail_Type(t *testing.T) {
	type args struct {
		msgPlatformUnjail MsgUnjail
	}
	tests := []struct {
		name string
		args
		want string
	}{
		{
			name: "return signers",
			args: args{msgPlatformUnjail},
			want: MsgPlatformUnjailName,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgPlatformUnjail.Type(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSigners() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestMsgPlatformUnjail_GetSigners(t *testing.T) {
	type args struct {
		msgPlatformUnjail MsgUnjail
	}
	tests := []struct {
		name string
		args
		want []sdk.Address
	}{
		{
			name: "return signers",
			args: args{msgPlatformUnjail},
			want: []sdk.Address{sdk.Address(msgPlatformUnjail.PlatformAddr)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgPlatformUnjail.GetSigners(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSigners() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestMsgPlatformUnjail_GetSignBytes(t *testing.T) {
	type args struct {
		msgPlatformUnjail MsgUnjail
	}
	res, err := ModuleCdc.MarshalJSON(&msgPlatformUnjail)
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
			args: args{msgPlatformUnjail},
			want: res,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgPlatformUnjail.GetSignBytes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSigners() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestMsgPlatformUnjail_ValidateBasic(t *testing.T) {
	type args struct {
		msgPlatformUnjail MsgUnjail
	}
	tests := []struct {
		name string
		args
		want sdk.Error
	}{
		{
			name: "errs if no Address",
			args: args{MsgUnjail{}},
			want: ErrBadPlatformAddr(DefaultCodespace),
		},
		{
			name: "returns nil if valid address",
			args: args{msgPlatformUnjail},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgPlatformUnjail.ValidateBasic(); got != nil {
				if !reflect.DeepEqual(got.Error(), tt.want.Error()) {
					t.Errorf("GetSigners() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
