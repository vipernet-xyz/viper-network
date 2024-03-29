package types

import (
	"math/rand"
	"reflect"
	"testing"
	"time"

	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	"github.com/vipernet-xyz/viper-network/types"
)

func TestAddressFromPrevStateRequestorPowerKey(t *testing.T) {
	type args struct {
		key []byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{{"sampleByteArray", args{key: []byte{0x51, 0x41, 0x33}}, []byte{0x41, 0x33}}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AddressFromKey(tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AddressFromKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeyForUnstakingRequestors(t *testing.T) {
	type args struct {
		unstakingTime time.Time
	}
	ut := time.Now()

	tests := []struct {
		name string
		args args
		want []byte
	}{
		{"sampleByteArray", args{ut}, append(UnstakingRequestorsKey, types.FormatTimeBytes(ut)...)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := KeyForUnstakingRequestors(tt.args.unstakingTime); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("KeyForUnstakingRequestors() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeyForValByAllVals(t *testing.T) {
	type args struct {
		addr types.Address
	}
	ca, _ := types.AddressFromHex("29f0a60104f3218a2cb51e6a269182d5dc271447114e342086d9c922a106a3c0")

	tests := []struct {
		name string
		args args
		want []byte
	}{
		{"sampleByteArray", args{ca}, append(AllRequestorsKey, ca.Bytes()...)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := KeyForRequestorByAllRequestors(tt.args.addr); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("KeyForValByAllVals() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeyForRequestorBurn(t *testing.T) {
	type args struct {
		address types.Address
	}
	ca, _ := types.AddressFromHex("29f0a60104f3218a2cb51e6a269182d5dc271447114e342086d9c922a106a3c0")

	tests := []struct {
		name string
		args args
		want []byte
	}{
		{"sampleByteArray", args{ca}, append(BurnRequestorKey, ca.Bytes()...)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := KeyForRequestorBurn(tt.args.address); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("KeyForRequestorBurn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeyForRequestorInStakingSet(t *testing.T) {
	type args struct {
		requestor Requestor
	}
	var pub crypto.Ed25519PublicKey
	_, err := rand.Read(pub[:])
	if err != nil {
		_ = err
	}

	operAddrInvr := types.CopyBytes(pub.Address())
	for i, b := range operAddrInvr {
		operAddrInvr[i] = ^b
	}

	tests := []struct {
		name string
		args args
		want []byte
	}{
		{"NewRequestor", args{requestor: NewRequestor(types.Address(pub.Address()), pub, []string{"0001"}, types.ZeroInt(), []string{"0001"}, 5)}, append([]byte{0x02, 0, 0, 0, 0, 0, 0, 0, 0}, operAddrInvr...)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := KeyForRequestorInStakingSet(tt.args.requestor); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("KeyForRequestorInStakingSet() = %s, want %s", got, tt.want)
			}
		})
	}
}
