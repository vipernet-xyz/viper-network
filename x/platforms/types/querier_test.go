package types

import (
	"math/rand"
	"reflect"
	"testing"

	"github.com/vipernet-xyz/viper-network/crypto"
	"github.com/vipernet-xyz/viper-network/types"
)

func TestNewQueryPlatformParams(t *testing.T) {
	type args struct {
		platformAddr types.Address
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
		want QueryPlatformParams
	}{
		{"default Test", args{va}, QueryPlatformParams{Address: va}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewQueryPlatformParams(tt.args.platformAddr); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewQueryPlatformParams() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewQueryPlatformsParams(t *testing.T) {
	type args struct {
		page  int
		limit int
	}
	tests := []struct {
		name string
		args args
		want QueryPlatformsParams
	}{
		{"Default Test", args{page: 1, limit: 1}, QueryPlatformsParams{Page: 1, Limit: 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewQueryPlatformsParams(tt.args.page, tt.args.limit); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewQueryPlatformsParams() = %v, want %v", got, tt.want)
			}
		})
	}
}
