package types

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/vipernet-xyz/viper-network/codec"
	"github.com/vipernet-xyz/viper-network/codec/types"
	"github.com/vipernet-xyz/viper-network/crypto"
	sdk "github.com/vipernet-xyz/viper-network/types"

	types2 "github.com/tendermint/tendermint/abci/types"
)

func TestNewProvider(t *testing.T) {
	type args struct {
		addr          sdk.Address
		pubkey        crypto.PublicKey
		tokensToStake sdk.BigInt
		chains        []string
		serviceURL    string
	}
	var pub crypto.Ed25519PublicKey
	_, err := rand.Read(pub[:])
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	tests := []struct {
		name string
		args args
		want Provider
	}{
		{"defaultProvider", args{sdk.Address(pub.Address()), pub, sdk.ZeroInt(), []string{"0001"}, "google.com"},
			Provider{
				Address:                 sdk.Address(pub.Address()),
				PublicKey:               pub,
				Jailed:                  false,
				Status:                  sdk.Staked,
				StakedTokens:            sdk.ZeroInt(),
				Chains:                  []string{"0001"},
				UnstakingCompletionTime: time.Time{}, // zero out because status: staked
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewProvider(tt.args.addr, tt.args.pubkey, tt.args.chains, tt.args.tokensToStake); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewProvider() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_AddStakedTokens(t *testing.T) {
	type fields struct {
		Address                 sdk.Address
		pubkey                  crypto.PublicKey
		Jailed                  bool
		Status                  sdk.StakeStatus
		StakedTokens            sdk.BigInt
		UnstakingCompletionTime time.Time
	}
	var pub crypto.Ed25519PublicKey
	_, err := rand.Read(pub[:])
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	type args struct {
		tokens sdk.BigInt
	}
	tests := []struct {
		name     string
		hasError bool
		fields   fields
		args     args
		want     interface{}
	}{
		{
			"Default Add Token Test",
			false,
			fields{
				Address:                 sdk.Address(pub.Address()),
				pubkey:                  pub,
				Jailed:                  false,
				Status:                  sdk.Staked,
				StakedTokens:            sdk.ZeroInt(),
				UnstakingCompletionTime: time.Time{},
			},
			args{
				tokens: sdk.NewInt(100),
			},
			Provider{
				Address:                 sdk.Address(pub.Address()),
				PublicKey:               pub,
				Jailed:                  false,
				Status:                  sdk.Staked,
				StakedTokens:            sdk.NewInt(100),
				UnstakingCompletionTime: time.Time{},
			},
		},
		{
			" hasError Add negative amount",
			true,
			fields{
				Address:                 sdk.Address(pub.Address()),
				pubkey:                  pub,
				Jailed:                  false,
				Status:                  sdk.Staked,
				StakedTokens:            sdk.ZeroInt(),
				UnstakingCompletionTime: time.Time{},
			},
			args{
				tokens: sdk.NewInt(-1),
			},
			"should not happen: trying to add negative tokens -1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Provider{
				Address:                 tt.fields.Address,
				PublicKey:               tt.fields.pubkey,
				Jailed:                  tt.fields.Jailed,
				Status:                  tt.fields.Status,
				StakedTokens:            tt.fields.StakedTokens,
				UnstakingCompletionTime: tt.fields.UnstakingCompletionTime,
			}
			switch tt.hasError {
			case true:
				_, _ = v.AddStakedTokens(tt.args.tokens)
			default:
				if got, _ := v.AddStakedTokens(tt.args.tokens); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("AddStakedTokens() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestProvider_ConsAddress(t *testing.T) {
	type fields struct {
		Address                 sdk.Address
		pubkey                  crypto.PublicKey
		Jailed                  bool
		Status                  sdk.StakeStatus
		StakedTokens            sdk.BigInt
		UnstakingCompletionTime time.Time
	}
	var pub crypto.Ed25519PublicKey
	_, err := rand.Read(pub[:])
	if err != nil {
		_ = err
	}

	tests := []struct {
		name   string
		fields fields
		want   sdk.Address
	}{
		{"Default Test", fields{
			Address:                 sdk.Address(pub.Address()),
			pubkey:                  pub,
			Jailed:                  false,
			Status:                  sdk.Staked,
			StakedTokens:            sdk.ZeroInt(),
			UnstakingCompletionTime: time.Time{},
		}, sdk.Address(pub.Address())},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Provider{
				Address:                 tt.fields.Address,
				PublicKey:               tt.fields.pubkey,
				Jailed:                  tt.fields.Jailed,
				Status:                  tt.fields.Status,
				StakedTokens:            tt.fields.StakedTokens,
				UnstakingCompletionTime: tt.fields.UnstakingCompletionTime,
			}
			if got := v.GetAddress(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_ConsensusPower(t *testing.T) {
	type fields struct {
		Address                 sdk.Address
		pubkey                  crypto.PublicKey
		Jailed                  bool
		Status                  sdk.StakeStatus
		StakedTokens            sdk.BigInt
		UnstakingCompletionTime time.Time
	}
	var pub crypto.Ed25519PublicKey
	_, err := rand.Read(pub[:])
	if err != nil {
		_ = err
	}

	tests := []struct {
		name   string
		fields fields
		want   int64
	}{
		{"Default Test / 0 power", fields{
			Address:                 sdk.Address(pub.Address()),
			pubkey:                  pub,
			Jailed:                  false,
			Status:                  sdk.Staked,
			StakedTokens:            sdk.ZeroInt(),
			UnstakingCompletionTime: time.Time{},
		}, 0},
		{"Default Test / 1 power", fields{
			Address:                 sdk.Address(pub.Address()),
			pubkey:                  pub,
			Jailed:                  false,
			Status:                  sdk.Staked,
			StakedTokens:            sdk.NewInt(1000000),
			UnstakingCompletionTime: time.Time{},
		}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Provider{
				Address:                 tt.fields.Address,
				PublicKey:               tt.fields.pubkey,
				Jailed:                  tt.fields.Jailed,
				Status:                  tt.fields.Status,
				StakedTokens:            tt.fields.StakedTokens,
				UnstakingCompletionTime: tt.fields.UnstakingCompletionTime,
			}
			if got := v.ConsensusPower(); got != tt.want {
				t.Errorf("ConsensusPower() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_Equals(t *testing.T) {
	type fields struct {
		Address                 sdk.Address
		pubkey                  crypto.PublicKey
		Jailed                  bool
		Status                  sdk.StakeStatus
		StakedTokens            sdk.BigInt
		UnstakingCompletionTime time.Time
	}
	var pub crypto.Ed25519PublicKey
	_, err := rand.Read(pub[:])
	if err != nil {
		_ = err
	}

	type args struct {
		v2 Provider
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{"Default Test Equal", fields{
			Address:                 sdk.Address(pub.Address()),
			pubkey:                  pub,
			Jailed:                  false,
			Status:                  sdk.Staked,
			StakedTokens:            sdk.ZeroInt(),
			UnstakingCompletionTime: time.Time{},
		}, args{Provider{
			Address:                 sdk.Address(pub.Address()),
			PublicKey:               pub,
			Jailed:                  false,
			Status:                  sdk.Staked,
			StakedTokens:            sdk.ZeroInt(),
			UnstakingCompletionTime: time.Time{},
		}}, true},
		{"Default Test Not Equal", fields{
			Address:                 sdk.Address(pub.Address()),
			pubkey:                  pub,
			Jailed:                  false,
			Status:                  sdk.Staked,
			StakedTokens:            sdk.ZeroInt(),
			UnstakingCompletionTime: time.Time{},
		}, args{Provider{
			Address:                 sdk.Address(pub.Address()),
			PublicKey:               pub,
			Jailed:                  false,
			Status:                  sdk.Unstaked,
			StakedTokens:            sdk.ZeroInt(),
			UnstakingCompletionTime: time.Time{},
		}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Provider{
				Address:                 tt.fields.Address,
				PublicKey:               tt.fields.pubkey,
				Jailed:                  tt.fields.Jailed,
				Status:                  tt.fields.Status,
				StakedTokens:            tt.fields.StakedTokens,
				UnstakingCompletionTime: tt.fields.UnstakingCompletionTime,
			}
			if got := v.Equals(tt.args.v2); got != tt.want {
				t.Errorf("Equals() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_GetAddress(t *testing.T) {
	type fields struct {
		Address                 sdk.Address
		pubkey                  crypto.PublicKey
		Jailed                  bool
		Status                  sdk.StakeStatus
		StakedTokens            sdk.BigInt
		UnstakingCompletionTime time.Time
	}

	var pub crypto.Ed25519PublicKey
	_, err := rand.Read(pub[:])
	if err != nil {
		_ = err
	}

	tests := []struct {
		name   string
		fields fields
		want   sdk.Address
	}{
		{"Default Test", fields{
			Address:                 sdk.Address(pub.Address()),
			pubkey:                  pub,
			Jailed:                  false,
			Status:                  sdk.Staked,
			StakedTokens:            sdk.ZeroInt(),
			UnstakingCompletionTime: time.Time{},
		}, sdk.Address(pub.Address())},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Provider{
				Address:                 tt.fields.Address,
				PublicKey:               tt.fields.pubkey,
				Jailed:                  tt.fields.Jailed,
				Status:                  tt.fields.Status,
				StakedTokens:            tt.fields.StakedTokens,
				UnstakingCompletionTime: tt.fields.UnstakingCompletionTime,
			}
			if got := v.GetAddress(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_GetConsAddr(t *testing.T) {
	type fields struct {
		Address                 sdk.Address
		pubkey                  crypto.PublicKey
		Jailed                  bool
		Status                  sdk.StakeStatus
		StakedTokens            sdk.BigInt
		UnstakingCompletionTime time.Time
	}
	var pub crypto.Ed25519PublicKey
	_, err := rand.Read(pub[:])
	if err != nil {
		_ = err
	}

	tests := []struct {
		name   string
		fields fields
		want   sdk.Address
	}{
		{"Default Test", fields{
			Address:                 sdk.Address(pub.Address()),
			pubkey:                  pub,
			Jailed:                  false,
			Status:                  sdk.Staked,
			StakedTokens:            sdk.ZeroInt(),
			UnstakingCompletionTime: time.Time{},
		}, sdk.Address(pub.Address())},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Provider{
				Address:                 tt.fields.Address,
				PublicKey:               tt.fields.pubkey,
				Jailed:                  tt.fields.Jailed,
				Status:                  tt.fields.Status,
				StakedTokens:            tt.fields.StakedTokens,
				UnstakingCompletionTime: tt.fields.UnstakingCompletionTime,
			}
			if got := v.GetAddress(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_Getpubkey(t *testing.T) {
	type fields struct {
		Address                 sdk.Address
		pubkey                  crypto.PublicKey
		Jailed                  bool
		Status                  sdk.StakeStatus
		StakedTokens            sdk.BigInt
		UnstakingCompletionTime time.Time
	}
	var pub crypto.Ed25519PublicKey
	_, err := rand.Read(pub[:])
	if err != nil {
		_ = err
	}

	tests := []struct {
		name   string
		fields fields
		want   crypto.PublicKey
	}{
		{"Default Test", fields{
			Address:                 sdk.Address(pub.Address()),
			pubkey:                  pub,
			Jailed:                  false,
			Status:                  sdk.Staked,
			StakedTokens:            sdk.ZeroInt(),
			UnstakingCompletionTime: time.Time{},
		}, pub},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Provider{
				Address:                 tt.fields.Address,
				PublicKey:               tt.fields.pubkey,
				Jailed:                  tt.fields.Jailed,
				Status:                  tt.fields.Status,
				StakedTokens:            tt.fields.StakedTokens,
				UnstakingCompletionTime: tt.fields.UnstakingCompletionTime,
			}
			if got := v.GetPublicKey(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPublicKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_GetConsensusPower(t *testing.T) {
	type fields struct {
		Address                 sdk.Address
		pubkey                  crypto.PublicKey
		Jailed                  bool
		Status                  sdk.StakeStatus
		StakedTokens            sdk.BigInt
		UnstakingCompletionTime time.Time
	}

	var pub crypto.Ed25519PublicKey
	_, err := rand.Read(pub[:])
	if err != nil {
		_ = err
	}

	tests := []struct {
		name   string
		fields fields
		want   int64
	}{
		{"Default Test", fields{
			Address:                 sdk.Address(pub.Address()),
			pubkey:                  pub,
			Jailed:                  false,
			Status:                  sdk.Staked,
			StakedTokens:            sdk.ZeroInt(),
			UnstakingCompletionTime: time.Time{},
		}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Provider{
				Address:                 tt.fields.Address,
				PublicKey:               tt.fields.pubkey,
				Jailed:                  tt.fields.Jailed,
				Status:                  tt.fields.Status,
				StakedTokens:            tt.fields.StakedTokens,
				UnstakingCompletionTime: tt.fields.UnstakingCompletionTime,
			}
			if got := v.GetConsensusPower(); got != tt.want {
				t.Errorf("GetConsensusPower() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_GetStatus(t *testing.T) {
	type fields struct {
		Address                 sdk.Address
		pubkey                  crypto.PublicKey
		Jailed                  bool
		Status                  sdk.StakeStatus
		StakedTokens            sdk.BigInt
		UnstakingCompletionTime time.Time
	}

	var pub crypto.Ed25519PublicKey
	_, err := rand.Read(pub[:])
	if err != nil {
		_ = err
	}

	tests := []struct {
		name   string
		fields fields
		want   sdk.StakeStatus
	}{
		{"Default Test", fields{
			Address:                 sdk.Address(pub.Address()),
			pubkey:                  pub,
			Jailed:                  false,
			Status:                  sdk.Staked,
			StakedTokens:            sdk.ZeroInt(),
			UnstakingCompletionTime: time.Time{},
		}, sdk.Staked},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Provider{
				Address:                 tt.fields.Address,
				PublicKey:               tt.fields.pubkey,
				Jailed:                  tt.fields.Jailed,
				Status:                  tt.fields.Status,
				StakedTokens:            tt.fields.StakedTokens,
				UnstakingCompletionTime: tt.fields.UnstakingCompletionTime,
			}
			if got := v.GetStatus(); got != tt.want {
				t.Errorf("GetStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_GetTokens(t *testing.T) {
	type fields struct {
		Address                 sdk.Address
		pubkey                  crypto.PublicKey
		Jailed                  bool
		Status                  sdk.StakeStatus
		StakedTokens            sdk.BigInt
		UnstakingCompletionTime time.Time
	}

	var pub crypto.Ed25519PublicKey
	_, err := rand.Read(pub[:])
	if err != nil {
		_ = err
	}

	tests := []struct {
		name   string
		fields fields
		want   sdk.BigInt
	}{
		{"Default Test", fields{
			Address:                 sdk.Address(pub.Address()),
			pubkey:                  pub,
			Jailed:                  false,
			Status:                  sdk.Staked,
			StakedTokens:            sdk.ZeroInt(),
			UnstakingCompletionTime: time.Time{},
		}, sdk.ZeroInt()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Provider{
				Address:                 tt.fields.Address,
				PublicKey:               tt.fields.pubkey,
				Jailed:                  tt.fields.Jailed,
				Status:                  tt.fields.Status,
				StakedTokens:            tt.fields.StakedTokens,
				UnstakingCompletionTime: tt.fields.UnstakingCompletionTime,
			}
			if got := v.GetTokens(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTokens() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_IsJailed(t *testing.T) {
	type fields struct {
		Address                 sdk.Address
		pubkey                  crypto.PublicKey
		Jailed                  bool
		Status                  sdk.StakeStatus
		StakedTokens            sdk.BigInt
		UnstakingCompletionTime time.Time
	}

	var pub crypto.Ed25519PublicKey
	_, err := rand.Read(pub[:])
	if err != nil {
		_ = err
	}

	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"Default Test", fields{
			Address:                 sdk.Address(pub.Address()),
			pubkey:                  pub,
			Jailed:                  false,
			Status:                  sdk.Staked,
			StakedTokens:            sdk.ZeroInt(),
			UnstakingCompletionTime: time.Time{},
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Provider{
				Address:                 tt.fields.Address,
				PublicKey:               tt.fields.pubkey,
				Jailed:                  tt.fields.Jailed,
				Status:                  tt.fields.Status,
				StakedTokens:            tt.fields.StakedTokens,
				UnstakingCompletionTime: tt.fields.UnstakingCompletionTime,
			}
			if got := v.IsJailed(); got != tt.want {
				t.Errorf("IsJailed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_IsStaked(t *testing.T) {
	type fields struct {
		Address                 sdk.Address
		pubkey                  crypto.PublicKey
		Jailed                  bool
		Status                  sdk.StakeStatus
		StakedTokens            sdk.BigInt
		UnstakingCompletionTime time.Time
	}

	var pub crypto.Ed25519PublicKey
	_, err := rand.Read(pub[:])
	if err != nil {
		_ = err
	}

	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"Default Test / staked true", fields{
			Address:                 sdk.Address(pub.Address()),
			pubkey:                  pub,
			Jailed:                  false,
			Status:                  sdk.Staked,
			StakedTokens:            sdk.ZeroInt(),
			UnstakingCompletionTime: time.Time{},
		}, true},
		{"Default Test / Unstaking false", fields{
			Address:                 sdk.Address(pub.Address()),
			pubkey:                  pub,
			Jailed:                  false,
			Status:                  sdk.Unstaking,
			StakedTokens:            sdk.ZeroInt(),
			UnstakingCompletionTime: time.Time{},
		}, false},
		{"Default Test / Unstaked false", fields{
			Address:                 sdk.Address(pub.Address()),
			pubkey:                  pub,
			Jailed:                  false,
			Status:                  sdk.Unstaked,
			StakedTokens:            sdk.ZeroInt(),
			UnstakingCompletionTime: time.Time{},
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Provider{
				Address:                 tt.fields.Address,
				PublicKey:               tt.fields.pubkey,
				Jailed:                  tt.fields.Jailed,
				Status:                  tt.fields.Status,
				StakedTokens:            tt.fields.StakedTokens,
				UnstakingCompletionTime: tt.fields.UnstakingCompletionTime,
			}
			if got := v.IsStaked(); got != tt.want {
				t.Errorf("IsStaked() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_IsUnstaked(t *testing.T) {
	type fields struct {
		Address                 sdk.Address
		pubkey                  crypto.PublicKey
		Jailed                  bool
		Status                  sdk.StakeStatus
		StakedTokens            sdk.BigInt
		UnstakingCompletionTime time.Time
	}

	var pub crypto.Ed25519PublicKey
	_, err := rand.Read(pub[:])
	if err != nil {
		_ = err
	}

	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"Default Test / staked false", fields{
			Address:                 sdk.Address(pub.Address()),
			pubkey:                  pub,
			Jailed:                  false,
			Status:                  sdk.Staked,
			StakedTokens:            sdk.ZeroInt(),
			UnstakingCompletionTime: time.Time{},
		}, false},
		{"Default Test / Unstaking false", fields{
			Address:                 sdk.Address(pub.Address()),
			pubkey:                  pub,
			Jailed:                  false,
			Status:                  sdk.Unstaking,
			StakedTokens:            sdk.ZeroInt(),
			UnstakingCompletionTime: time.Time{},
		}, false},
		{"Default Test / Unstaked true", fields{
			Address:                 sdk.Address(pub.Address()),
			pubkey:                  pub,
			Jailed:                  false,
			Status:                  sdk.Unstaked,
			StakedTokens:            sdk.ZeroInt(),
			UnstakingCompletionTime: time.Time{},
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Provider{
				Address:                 tt.fields.Address,
				PublicKey:               tt.fields.pubkey,
				Jailed:                  tt.fields.Jailed,
				Status:                  tt.fields.Status,
				StakedTokens:            tt.fields.StakedTokens,
				UnstakingCompletionTime: tt.fields.UnstakingCompletionTime,
			}
			if got := v.IsUnstaked(); got != tt.want {
				t.Errorf("IsUnstaked() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_IsUnstaking(t *testing.T) {
	type fields struct {
		Address                 sdk.Address
		pubkey                  crypto.PublicKey
		Jailed                  bool
		Status                  sdk.StakeStatus
		StakedTokens            sdk.BigInt
		UnstakingCompletionTime time.Time
	}

	var pub crypto.Ed25519PublicKey
	_, err := rand.Read(pub[:])
	if err != nil {
		_ = err
	}

	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"Default Test / staked false", fields{
			Address:                 sdk.Address(pub.Address()),
			pubkey:                  pub,
			Jailed:                  false,
			Status:                  sdk.Staked,
			StakedTokens:            sdk.ZeroInt(),
			UnstakingCompletionTime: time.Time{},
		}, false},
		{"Default Test / Unstaking true", fields{
			Address:                 sdk.Address(pub.Address()),
			pubkey:                  pub,
			Jailed:                  false,
			Status:                  sdk.Unstaking,
			StakedTokens:            sdk.ZeroInt(),
			UnstakingCompletionTime: time.Time{},
		}, true},
		{"Default Test / Unstaked false", fields{
			Address:                 sdk.Address(pub.Address()),
			pubkey:                  pub,
			Jailed:                  false,
			Status:                  sdk.Unstaked,
			StakedTokens:            sdk.ZeroInt(),
			UnstakingCompletionTime: time.Time{},
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Provider{
				Address:                 tt.fields.Address,
				PublicKey:               tt.fields.pubkey,
				Jailed:                  tt.fields.Jailed,
				Status:                  tt.fields.Status,
				StakedTokens:            tt.fields.StakedTokens,
				UnstakingCompletionTime: tt.fields.UnstakingCompletionTime,
			}
			if got := v.IsUnstaking(); got != tt.want {
				t.Errorf("IsUnstaking() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_PotentialConsensusPower(t *testing.T) {
	type fields struct {
		Address                 sdk.Address
		pubkey                  crypto.PublicKey
		Jailed                  bool
		Status                  sdk.StakeStatus
		StakedTokens            sdk.BigInt
		UnstakingCompletionTime time.Time
	}

	var pub crypto.Ed25519PublicKey
	_, err := rand.Read(pub[:])
	if err != nil {
		_ = err
	}

	tests := []struct {
		name   string
		fields fields
		want   int64
	}{
		{"Default Test / potential power 0", fields{
			Address:                 sdk.Address(pub.Address()),
			pubkey:                  pub,
			Jailed:                  false,
			Status:                  sdk.Staked,
			StakedTokens:            sdk.ZeroInt(),
			UnstakingCompletionTime: time.Time{},
		}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Provider{
				Address:                 tt.fields.Address,
				PublicKey:               tt.fields.pubkey,
				Jailed:                  tt.fields.Jailed,
				Status:                  tt.fields.Status,
				StakedTokens:            tt.fields.StakedTokens,
				UnstakingCompletionTime: tt.fields.UnstakingCompletionTime,
			}
			if got := v.ConsensusPower(); got != tt.want {
				t.Errorf("ConsensusPower() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_RemoveStakedTokens(t *testing.T) {
	type fields struct {
		Address                 sdk.Address
		pubkey                  crypto.PublicKey
		Jailed                  bool
		Status                  sdk.StakeStatus
		StakedTokens            sdk.BigInt
		UnstakingCompletionTime time.Time
	}

	var pub crypto.Ed25519PublicKey
	_, err := rand.Read(pub[:])
	if err != nil {
		_ = err
	}

	type args struct {
		tokens sdk.BigInt
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Provider
	}{
		{"Remove 0 tokens having 0 tokens ", fields{
			Address:                 sdk.Address(pub.Address()),
			pubkey:                  pub,
			Jailed:                  false,
			Status:                  sdk.Staked,
			StakedTokens:            sdk.ZeroInt(),
			UnstakingCompletionTime: time.Time{},
		}, args{tokens: sdk.ZeroInt()}, Provider{
			Address:                 sdk.Address(pub.Address()),
			PublicKey:               pub,
			Jailed:                  false,
			Status:                  sdk.Staked,
			StakedTokens:            sdk.ZeroInt(),
			UnstakingCompletionTime: time.Time{},
		}},
		{"Remove 99 tokens having 100 tokens ", fields{
			Address:                 sdk.Address(pub.Address()),
			pubkey:                  pub,
			Jailed:                  false,
			Status:                  sdk.Staked,
			StakedTokens:            sdk.NewInt(100),
			UnstakingCompletionTime: time.Time{},
		}, args{tokens: sdk.NewInt(99)}, Provider{
			Address:                 sdk.Address(pub.Address()),
			PublicKey:               pub,
			Jailed:                  false,
			Status:                  sdk.Staked,
			StakedTokens:            sdk.OneInt(),
			UnstakingCompletionTime: time.Time{},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Provider{
				Address:                 tt.fields.Address,
				PublicKey:               tt.fields.pubkey,
				Jailed:                  tt.fields.Jailed,
				Status:                  tt.fields.Status,
				StakedTokens:            tt.fields.StakedTokens,
				UnstakingCompletionTime: tt.fields.UnstakingCompletionTime,
			}
			if got, _ := v.RemoveStakedTokens(tt.args.tokens); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveStakedTokens() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_UpdateStatus(t *testing.T) {
	type fields struct {
		Address                 sdk.Address
		pubkey                  crypto.PublicKey
		Jailed                  bool
		Status                  sdk.StakeStatus
		StakedTokens            sdk.BigInt
		UnstakingCompletionTime time.Time
	}

	var pub crypto.Ed25519PublicKey
	_, err := rand.Read(pub[:])
	if err != nil {
		_ = err
	}

	type args struct {
		newStatus sdk.StakeStatus
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Provider
	}{
		{"Test Staked -> Unstaking", fields{
			Address:                 sdk.Address(pub.Address()),
			pubkey:                  pub,
			Jailed:                  false,
			Status:                  sdk.Staked,
			StakedTokens:            sdk.ZeroInt(),
			UnstakingCompletionTime: time.Time{},
		}, args{newStatus: sdk.Unstaking}, Provider{
			Address:                 sdk.Address(pub.Address()),
			PublicKey:               pub,
			Jailed:                  false,
			Status:                  sdk.Unstaking,
			StakedTokens:            sdk.ZeroInt(),
			UnstakingCompletionTime: time.Time{},
		}},
		{"Test Unstaking -> Unstaked", fields{
			Address:                 sdk.Address(pub.Address()),
			pubkey:                  pub,
			Jailed:                  false,
			Status:                  sdk.Unstaking,
			StakedTokens:            sdk.ZeroInt(),
			UnstakingCompletionTime: time.Time{},
		}, args{newStatus: sdk.Unstaked}, Provider{
			Address:                 sdk.Address(pub.Address()),
			PublicKey:               pub,
			Jailed:                  false,
			Status:                  sdk.Unstaked,
			StakedTokens:            sdk.ZeroInt(),
			UnstakingCompletionTime: time.Time{},
		}},
		{"Test Unstaked -> Staked", fields{
			Address:                 sdk.Address(pub.Address()),
			pubkey:                  pub,
			Jailed:                  false,
			Status:                  sdk.Unstaked,
			StakedTokens:            sdk.ZeroInt(),
			UnstakingCompletionTime: time.Time{},
		}, args{newStatus: sdk.Staked}, Provider{
			Address:                 sdk.Address(pub.Address()),
			PublicKey:               pub,
			Jailed:                  false,
			Status:                  sdk.Staked,
			StakedTokens:            sdk.ZeroInt(),
			UnstakingCompletionTime: time.Time{},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Provider{
				Address:                 tt.fields.Address,
				PublicKey:               tt.fields.pubkey,
				Jailed:                  tt.fields.Jailed,
				Status:                  tt.fields.Status,
				StakedTokens:            tt.fields.StakedTokens,
				UnstakingCompletionTime: tt.fields.UnstakingCompletionTime,
			}
			if got := v.UpdateStatus(tt.args.newStatus); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdateStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_GetChains(t *testing.T) {
	type args struct {
		addr          sdk.Address
		pubkey        crypto.PublicKey
		tokensToStake sdk.BigInt
		chains        []string
		serviceURL    string
	}
	var pub crypto.Ed25519PublicKey
	_, err := rand.Read(pub[:])
	if err != nil {
		_ = err
	}

	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			"defaultProvider",
			args{sdk.Address(pub.Address()), pub, sdk.ZeroInt(), []string{"0001"}, "google.com"},
			[]string{"0001"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewProvider(tt.args.addr, tt.args.pubkey, tt.args.chains, tt.args.tokensToStake)
			if got := provider.GetChains(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetChains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_GetMaxRelays(t *testing.T) {
	type args struct {
		addr          sdk.Address
		pubkey        crypto.PublicKey
		tokensToStake sdk.BigInt
		chains        []string
		serviceURL    string
		maxRelays     sdk.BigInt
	}
	var pub crypto.Ed25519PublicKey
	_, err := rand.Read(pub[:])
	if err != nil {
		_ = err
	}

	tests := []struct {
		name string
		args args
		want sdk.BigInt
	}{
		{
			"defaultProvider",
			args{sdk.Address(pub.Address()), pub, sdk.ZeroInt(), []string{"0001"}, "google.com", sdk.NewInt(1)},
			sdk.NewInt(1),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := Provider{
				Address:   tt.args.addr,
				PublicKey: tt.args.pubkey,
				Chains:    tt.args.chains,
				MaxRelays: tt.args.maxRelays,
			}
			if got := provider.GetMaxRelays(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMaxRelays() = %v, want %v", got, tt.want)
			}
		})
	}
}

var provider Provider
var cdc *codec.Codec

func init() {
	var pub crypto.Ed25519PublicKey
	_, err := rand.Read(pub[:])
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	cdc = codec.NewCodec(types.NewInterfaceRegistry())
	RegisterCodec(cdc)
	crypto.RegisterAmino(cdc.AminoCodec().Amino)

	provider = Provider{
		Address:                 sdk.Address(pub.Address()),
		PublicKey:               pub,
		Jailed:                  false,
		Status:                  sdk.Staked,
		StakedTokens:            sdk.NewInt(100),
		MaxRelays:               sdk.NewInt(1000),
		UnstakingCompletionTime: time.Time{},
	}
}

func TestProviderUtil_MarshalJSON(t *testing.T) {
	type args struct {
		provider Provider
		codec    *codec.Codec
	}
	hexProvider := JSONProvider{
		Address:                 provider.Address,
		PublicKey:               provider.PublicKey.RawString(),
		Jailed:                  provider.Jailed,
		Status:                  provider.Status,
		StakedTokens:            provider.StakedTokens,
		UnstakingCompletionTime: provider.UnstakingCompletionTime,
		MaxRelays:               provider.MaxRelays,
	}
	bz, _ := cdc.MarshalJSON(hexProvider)

	tests := []struct {
		name string
		args
		want []byte
	}{
		{
			name: "marshals provider",
			args: args{provider: provider, codec: cdc},
			want: bz,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := tt.args.provider.MarshalJSON(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MmashalJSON() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestProviderUtil_String(t *testing.T) {
	tests := []struct {
		name string
		args Providers
		want string
	}{
		{
			name: "serializes providerlicaitons into string",
			args: Providers{provider},
			want: fmt.Sprintf("Address:\t\t%s\nPublic Key:\t\t%s\nJailed:\t\t\t%v\nChains:\t\t\t%v\nMaxRelays:\t\t%s\nStatus:\t\t\t%s\nTokens:\t\t\t%s\nUnstaking Time:\t%v\n----\n",
				provider.Address,
				provider.PublicKey.RawString(),
				provider.Jailed,
				provider.Chains,
				provider.MaxRelays.String(),
				provider.Status,
				provider.StakedTokens,
				provider.UnstakingCompletionTime,
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.String(); got != strings.TrimSpace(fmt.Sprintf("%s\n", tt.want)) {
				t.Errorf("String() = \n%s\nwant\n%s", got, tt.want)
			}
		})
	}
}

func TestProviderUtil_JSON(t *testing.T) {
	providers := Providers{provider}
	j, _ := json.Marshal(providers)

	tests := []struct {
		name string
		args Providers
		want []byte
	}{
		{
			name: "serializes providerlicaitons into JSON",
			args: providers,
			want: j,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := tt.args.JSON(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("JSON() = %s", got)
				t.Errorf("JSON() = %s", tt.want)
			}
		})
	}
}
func TestProviderUtil_UnmarshalJSON(t *testing.T) {
	type args struct {
		provider Provider
	}
	tests := []struct {
		name string
		args
		want Provider
	}{
		{
			name: "marshals provider",
			args: args{provider: provider},
			want: provider,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			marshaled, err := tt.args.provider.MarshalJSON()
			if err != nil {
				t.Fatalf("Cannot marshal provider")
			}
			if err = tt.args.provider.UnmarshalJSON(marshaled); err != nil {
				t.Fatalf("UnmarshalObject(): returns %v but want %v", err, tt.want)
			}
			// NOTE CANNOT PERFORM DEEP EQUAL
			// Unmarshalling causes StakedTokens & MaxRelays to be
			//  assigned a new memory address overwriting the previous reference to provider
			// separate them and assert absolute value rather than deep equal

			gotStaked := tt.args.provider.StakedTokens
			wantStaked := tt.want.StakedTokens
			gotRelays := tt.args.provider.StakedTokens
			wantRelays := tt.want.StakedTokens

			tt.args.provider.StakedTokens = tt.want.StakedTokens
			tt.args.provider.MaxRelays = tt.want.MaxRelays

			if !reflect.DeepEqual(tt.args.provider, tt.want) {
				t.Errorf("got %v but want %v", tt.args.provider, tt.want)
			}
			if !gotStaked.Equal(wantStaked) {
				t.Errorf("got %v but want %v", gotStaked, wantStaked)
			}
			if !gotRelays.Equal(wantRelays) {
				t.Errorf("got %v but want %v", gotRelays, wantRelays)
			}
		})
	}
}

func TestProviderUtil_UnMarshalProvider(t *testing.T) {
	type args struct {
		provider Provider
		codec    *codec.Codec
	}
	tests := []struct {
		name string
		args
		want Provider
	}{
		{
			name: "can unmarshal provider",
			args: args{provider: provider, codec: cdc},
			want: provider,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := sdk.NewContext(nil, types2.Header{Height: 1}, false, nil)
			c.BlockHeight()
			bz, _ := MarshalProvider(tt.args.codec, c, tt.args.provider)
			unmarshaledProvider, err := UnmarshalProvider(tt.args.codec, c, bz)
			if err != nil {
				t.Fatalf("could not unmarshal provider")
			}

			if !reflect.DeepEqual(unmarshaledProvider, tt.want) {
				t.Fatalf("got %v but want %v", unmarshaledProvider, unmarshaledProvider)
			}
		})
	}
}
