package types

import (
	"math/rand"
	"reflect"
	"testing"

	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	"github.com/vipernet-xyz/viper-network/types"
)

func TestNewQueryRequestorParams(t *testing.T) {
	type args struct {
		requestorAddr types.Address
	}
	var pub crypto.Ed25519PublicKey
	_, err := rand.Read(pub[:])
	if err != nil {
		_ = err
	}
	va := types.Address(pub.Address())

	tests := []struct {
		name string
		args args
		want QueryRequestorParams
	}{
		{"default Test", args{va}, QueryRequestorParams{Address: va}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewQueryRequestorParams(tt.args.requestorAddr); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewQueryRequestorParams() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewQueryRequestorsParams(t *testing.T) {
	type args struct {
		page  int
		limit int
	}
	tests := []struct {
		name string
		args args
		want QueryRequestorsParams
	}{
		{"Default Test", args{page: 1, limit: 1}, QueryRequestorsParams{Page: 1, Limit: 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewQueryRequestorsParams(tt.args.page, tt.args.limit); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewQueryRequestorsParams() = %v, want %v", got, tt.want)
			}
		})
	}
}
