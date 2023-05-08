package types

import (
	"math/rand"
	"reflect"
	"testing"

	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	"github.com/vipernet-xyz/viper-network/types"
)

func TestNewQueryProviderParams(t *testing.T) {
	type args struct {
		providerAddr types.Address
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
		want QueryProviderParams
	}{
		{"default Test", args{va}, QueryProviderParams{Address: va}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewQueryProviderParams(tt.args.providerAddr); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewQueryProviderParams() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewQueryProvidersParams(t *testing.T) {
	type args struct {
		page  int
		limit int
	}
	tests := []struct {
		name string
		args args
		want QueryProvidersParams
	}{
		{"Default Test", args{page: 1, limit: 1}, QueryProvidersParams{Page: 1, Limit: 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewQueryProvidersParams(tt.args.page, tt.args.limit); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewQueryProvidersParams() = %v, want %v", got, tt.want)
			}
		})
	}
}
