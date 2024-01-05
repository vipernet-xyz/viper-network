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

var msgRequestorStake MsgStake
var msgBeginRequestorUnstake MsgBeginUnstake
var msgRequestorUnjail MsgUnjail
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

	msgRequestorStake = MsgStake{
		PubKey:       pub,
		Chains:       []string{"0001"},
		GeoZones:     []string{"0001"},
		NumServicers: 10,
		Value:        sdk.NewInt(10),
	}
	msgRequestorUnjail = MsgUnjail{sdk.Address(pub.Address())}
	msgBeginRequestorUnstake = MsgBeginUnstake{sdk.Address(pub.Address())}
}

func TestMsgRequestor_GetSigners(t *testing.T) {
	type args struct {
		msgRequestorStake MsgStake
	}
	tests := []struct {
		name string
		args
		want []sdk.Address
	}{
		{
			name: "return signers",
			args: args{msgRequestorStake},
			want: []sdk.Address{sdk.Address(pk.Address())},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgRequestorStake.GetSigners(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSigners() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestMsgRequestor_GetSignBytes(t *testing.T) {
	type args struct {
		msgRequestorStake MsgStake
	}
	res, err := ModuleCdc.MarshalJSON(&msgRequestorStake)
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
			args: args{msgRequestorStake},
			want: res,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgRequestorStake.GetSignBytes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSigners() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestMsgRequestor_Route(t *testing.T) {
	type args struct {
		msgRequestorStake MsgStake
	}
	tests := []struct {
		name string
		args
		want string
	}{
		{
			name: "return signers",
			args: args{msgRequestorStake},
			want: RouterKey,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgRequestorStake.Route(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSigners() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestMsgRequestor_Type(t *testing.T) {
	type args struct {
		msgRequestorStake MsgStake
	}
	tests := []struct {
		name string
		args
		want string
	}{
		{
			name: "return signers",
			args: args{msgRequestorStake},
			want: MsgRequestorStakeName,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgRequestorStake.Type(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSigners() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestMsgRequestor_ValidateBasic(t *testing.T) {
	type args struct {
		msgRequestorStake MsgStake
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
			want: ErrNilRequestorAddr(DefaultCodespace),
		},
		{
			name: "errs if no stake lower than zero",
			args: args{MsgStake{PubKey: msgRequestorStake.PubKey, Value: sdk.NewInt(-1)}},
			want: ErrBadStakeAmount(DefaultCodespace),
		},
		{
			name: "errs if no native chains supported",
			args: args{MsgStake{PubKey: msgRequestorStake.PubKey, Value: sdk.NewInt(1), Chains: []string{}}},
			want: ErrNoChains(DefaultCodespace),
		},
		{
			name: "errs if no native geozone supported",
			args: args{MsgStake{PubKey: msgRequestorStake.PubKey, Value: sdk.NewInt(1), Chains: []string{}, GeoZones: []string{}}},
			want: ErrNoChains(DefaultCodespace),
		},
		{
			name: "returns err",
			args: args{MsgStake{PubKey: msgRequestorStake.PubKey, Value: msgRequestorStake.Value, Chains: []string{"aaaaaa"}, GeoZones: []string{"aaaaaa"}}},
			want: ErrInvalidNetworkIdentifier("requestor", fmt.Errorf("net id length is > 2")),
		},
		{
			name: "returns nil if valid address",
			args: args{msgRequestorStake},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.args.msgRequestorStake.ValidateBasic()
			if got != nil && tt.want != nil {
				if !reflect.DeepEqual(got.Error(), tt.want.Error()) {
					t.Errorf("ValidatorBasic() = %v, want %v", got, tt.want)
				}
			} else if (got != nil && tt.want == nil) || (got == nil && tt.want != nil) {
				t.Errorf("ValidatorBasic() = %v, want %v", got, tt.want)
			}

		})
	}
}

func TestMsgBeginRequestorUnstake_GetSigners(t *testing.T) {
	type args struct {
		msgBeginRequestorUnstake MsgBeginUnstake
	}
	tests := []struct {
		name string
		args
		want []sdk.Address
	}{
		{
			name: "return signers",
			args: args{msgBeginRequestorUnstake},
			want: []sdk.Address{sdk.Address(pk.Address())},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgBeginRequestorUnstake.GetSigners(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSigners() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestMsgBeginRequestorUnstake_GetSignBytes(t *testing.T) {
	type args struct {
		msgBeginRequestorUnstake MsgBeginUnstake
	}
	res, err := ModuleCdc.MarshalJSON(&msgBeginRequestorUnstake)
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
			args: args{msgBeginRequestorUnstake},
			want: res,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgBeginRequestorUnstake.GetSignBytes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSignBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestMsgBeginRequestorUnstake_Route(t *testing.T) {
	type args struct {
		msgBeginRequestorUnstake MsgBeginUnstake
	}
	tests := []struct {
		name string
		args
		want string
	}{
		{
			name: "return signers",
			args: args{msgBeginRequestorUnstake},
			want: RouterKey,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgBeginRequestorUnstake.Route(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSigners() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestMsgBeginRequestorUnstake_Type(t *testing.T) {
	type args struct {
		msgBeginRequestorUnstake MsgBeginUnstake
	}
	tests := []struct {
		name string
		args
		want string
	}{
		{
			name: "return signers",
			args: args{msgBeginRequestorUnstake},
			want: MsgRequestorUnstakeName,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgBeginRequestorUnstake.Type(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSigners() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestMsgBeginRequestorUnstake_ValidateBasic(t *testing.T) {
	type args struct {
		msgBeginRequestorUnstake MsgBeginUnstake
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
			want: ErrNilRequestorAddr(DefaultCodespace),
		},
		{
			name: "returns nil if valid address",
			args: args{msgBeginRequestorUnstake},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgBeginRequestorUnstake.ValidateBasic(); got != nil {
				if !reflect.DeepEqual(got.Error(), tt.want.Error()) {
					t.Errorf("ValidatorBasic() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestMsgRequestorUnjail_Route(t *testing.T) {
	type args struct {
		msgRequestorUnjail MsgUnjail
	}
	tests := []struct {
		name string
		args
		want string
	}{
		{
			name: "return signers",
			args: args{msgRequestorUnjail},
			want: RouterKey,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgRequestorUnjail.Route(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSigners() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestMsgRequestorUnjail_Type(t *testing.T) {
	type args struct {
		msgRequestorUnjail MsgUnjail
	}
	tests := []struct {
		name string
		args
		want string
	}{
		{
			name: "return signers",
			args: args{msgRequestorUnjail},
			want: MsgRequestorUnjailName,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgRequestorUnjail.Type(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSigners() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestMsgRequestorUnjail_GetSigners(t *testing.T) {
	type args struct {
		msgRequestorUnjail MsgUnjail
	}
	tests := []struct {
		name string
		args
		want []sdk.Address
	}{
		{
			name: "return signers",
			args: args{msgRequestorUnjail},
			want: []sdk.Address{sdk.Address(msgRequestorUnjail.RequestorAddr)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgRequestorUnjail.GetSigners(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSigners() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestMsgRequestorUnjail_GetSignBytes(t *testing.T) {
	type args struct {
		msgRequestorUnjail MsgUnjail
	}
	res, err := ModuleCdc.MarshalJSON(&msgRequestorUnjail)
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
			args: args{msgRequestorUnjail},
			want: res,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgRequestorUnjail.GetSignBytes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSigners() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestMsgRequestorUnjail_ValidateBasic(t *testing.T) {
	type args struct {
		msgRequestorUnjail MsgUnjail
	}
	tests := []struct {
		name string
		args
		want sdk.Error
	}{
		{
			name: "errs if no Address",
			args: args{MsgUnjail{}},
			want: ErrBadRequestorAddr(DefaultCodespace),
		},
		{
			name: "returns nil if valid address",
			args: args{msgRequestorUnjail},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.msgRequestorUnjail.ValidateBasic(); got != nil {
				if !reflect.DeepEqual(got.Error(), tt.want.Error()) {
					t.Errorf("GetSigners() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
