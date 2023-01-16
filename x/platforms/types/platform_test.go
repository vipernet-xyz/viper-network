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

func TestNewPlatform(t *testing.T) {
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
		want Platform
	}{
		{"defaultPlatform", args{sdk.Address(pub.Address()), pub, sdk.ZeroInt(), []string{"0001"}, "google.com"},
			Platform{
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
			if got := NewPlatform(tt.args.addr, tt.args.pubkey, tt.args.chains, tt.args.tokensToStake); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPlatform() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlatform_AddStakedTokens(t *testing.T) {
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
			Platform{
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
			"should not hplatformen: trying to add negative tokens -1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Platform{
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

func TestPlatform_ConsAddress(t *testing.T) {
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
			v := Platform{
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

func TestPlatform_ConsensusPower(t *testing.T) {
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
			v := Platform{
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

func TestPlatform_Equals(t *testing.T) {
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
		v2 Platform
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
		}, args{Platform{
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
		}, args{Platform{
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
			v := Platform{
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

func TestPlatform_GetAddress(t *testing.T) {
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
			v := Platform{
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

func TestPlatform_GetConsAddr(t *testing.T) {
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
			v := Platform{
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

func TestPlatform_Getpubkey(t *testing.T) {
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
			v := Platform{
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

func TestPlatform_GetConsensusPower(t *testing.T) {
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
			v := Platform{
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

func TestPlatform_GetStatus(t *testing.T) {
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
			v := Platform{
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

func TestPlatform_GetTokens(t *testing.T) {
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
			v := Platform{
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

func TestPlatform_IsJailed(t *testing.T) {
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
			v := Platform{
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

func TestPlatform_IsStaked(t *testing.T) {
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
			v := Platform{
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

func TestPlatform_IsUnstaked(t *testing.T) {
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
			v := Platform{
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

func TestPlatform_IsUnstaking(t *testing.T) {
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
			v := Platform{
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

func TestPlatform_PotentialConsensusPower(t *testing.T) {
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
			v := Platform{
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

func TestPlatform_RemoveStakedTokens(t *testing.T) {
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
		want   Platform
	}{
		{"Remove 0 tokens having 0 tokens ", fields{
			Address:                 sdk.Address(pub.Address()),
			pubkey:                  pub,
			Jailed:                  false,
			Status:                  sdk.Staked,
			StakedTokens:            sdk.ZeroInt(),
			UnstakingCompletionTime: time.Time{},
		}, args{tokens: sdk.ZeroInt()}, Platform{
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
		}, args{tokens: sdk.NewInt(99)}, Platform{
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
			v := Platform{
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

func TestPlatform_UpdateStatus(t *testing.T) {
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
		want   Platform
	}{
		{"Test Staked -> Unstaking", fields{
			Address:                 sdk.Address(pub.Address()),
			pubkey:                  pub,
			Jailed:                  false,
			Status:                  sdk.Staked,
			StakedTokens:            sdk.ZeroInt(),
			UnstakingCompletionTime: time.Time{},
		}, args{newStatus: sdk.Unstaking}, Platform{
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
		}, args{newStatus: sdk.Unstaked}, Platform{
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
		}, args{newStatus: sdk.Staked}, Platform{
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
			v := Platform{
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

func TestPlatform_GetChains(t *testing.T) {
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
			"defaultPlatform",
			args{sdk.Address(pub.Address()), pub, sdk.ZeroInt(), []string{"0001"}, "google.com"},
			[]string{"0001"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			platform := NewPlatform(tt.args.addr, tt.args.pubkey, tt.args.chains, tt.args.tokensToStake)
			if got := platform.GetChains(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetChains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlatform_GetMaxRelays(t *testing.T) {
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
			"defaultPlatform",
			args{sdk.Address(pub.Address()), pub, sdk.ZeroInt(), []string{"0001"}, "google.com", sdk.NewInt(1)},
			sdk.NewInt(1),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			platform := Platform{
				Address:   tt.args.addr,
				PublicKey: tt.args.pubkey,
				Chains:    tt.args.chains,
				MaxRelays: tt.args.maxRelays,
			}
			if got := platform.GetMaxRelays(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMaxRelays() = %v, want %v", got, tt.want)
			}
		})
	}
}

var platform Platform
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

	platform = Platform{
		Address:                 sdk.Address(pub.Address()),
		PublicKey:               pub,
		Jailed:                  false,
		Status:                  sdk.Staked,
		StakedTokens:            sdk.NewInt(100),
		MaxRelays:               sdk.NewInt(1000),
		UnstakingCompletionTime: time.Time{},
	}
}

func TestPlatformUtil_MarshalJSON(t *testing.T) {
	type args struct {
		platform Platform
		codec    *codec.Codec
	}
	hexPlatform := JSONPlatform{
		Address:                 platform.Address,
		PublicKey:               platform.PublicKey.RawString(),
		Jailed:                  platform.Jailed,
		Status:                  platform.Status,
		StakedTokens:            platform.StakedTokens,
		UnstakingCompletionTime: platform.UnstakingCompletionTime,
		MaxRelays:               platform.MaxRelays,
	}
	bz, _ := cdc.MarshalJSON(hexPlatform)

	tests := []struct {
		name string
		args
		want []byte
	}{
		{
			name: "marshals platform",
			args: args{platform: platform, codec: cdc},
			want: bz,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := tt.args.platform.MarshalJSON(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MmashalJSON() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestPlatformUtil_String(t *testing.T) {
	tests := []struct {
		name string
		args Platforms
		want string
	}{
		{
			name: "serializes platformlicaitons into string",
			args: Platforms{platform},
			want: fmt.Sprintf("Address:\t\t%s\nPublic Key:\t\t%s\nJailed:\t\t\t%v\nChains:\t\t\t%v\nMaxRelays:\t\t%s\nStatus:\t\t\t%s\nTokens:\t\t\t%s\nUnstaking Time:\t%v\n----\n",
				platform.Address,
				platform.PublicKey.RawString(),
				platform.Jailed,
				platform.Chains,
				platform.MaxRelays.String(),
				platform.Status,
				platform.StakedTokens,
				platform.UnstakingCompletionTime,
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

func TestPlatformUtil_JSON(t *testing.T) {
	platforms := Platforms{platform}
	j, _ := json.Marshal(platforms)

	tests := []struct {
		name string
		args Platforms
		want []byte
	}{
		{
			name: "serializes platformlicaitons into JSON",
			args: platforms,
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
func TestPlatformUtil_UnmarshalJSON(t *testing.T) {
	type args struct {
		platform Platform
	}
	tests := []struct {
		name string
		args
		want Platform
	}{
		{
			name: "marshals platform",
			args: args{platform: platform},
			want: platform,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			marshaled, err := tt.args.platform.MarshalJSON()
			if err != nil {
				t.Fatalf("Cannot marshal platform")
			}
			if err = tt.args.platform.UnmarshalJSON(marshaled); err != nil {
				t.Fatalf("UnmarshalObject(): returns %v but want %v", err, tt.want)
			}
			// NOTE CANNOT PERFORM DEEP EQUAL
			// Unmarshalling causes StakedTokens & MaxRelays to be
			//  assigned a new memory address overwriting the previous reference to platform
			// separate them and assert absolute value rather than deep equal

			gotStaked := tt.args.platform.StakedTokens
			wantStaked := tt.want.StakedTokens
			gotRelays := tt.args.platform.StakedTokens
			wantRelays := tt.want.StakedTokens

			tt.args.platform.StakedTokens = tt.want.StakedTokens
			tt.args.platform.MaxRelays = tt.want.MaxRelays

			if !reflect.DeepEqual(tt.args.platform, tt.want) {
				t.Errorf("got %v but want %v", tt.args.platform, tt.want)
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

func TestPlatformUtil_UnMarshalPlatform(t *testing.T) {
	type args struct {
		platform Platform
		codec    *codec.Codec
	}
	tests := []struct {
		name string
		args
		want Platform
	}{
		{
			name: "can unmarshal platform",
			args: args{platform: platform, codec: cdc},
			want: platform,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := sdk.NewContext(nil, types2.Header{Height: 1}, false, nil)
			c.BlockHeight()
			bz, _ := MarshalPlatform(tt.args.codec, c, tt.args.platform)
			unmarshaledPlatform, err := UnmarshalPlatform(tt.args.codec, c, bz)
			if err != nil {
				t.Fatalf("could not unmarshal platform")
			}

			if !reflect.DeepEqual(unmarshaledPlatform, tt.want) {
				t.Fatalf("got %v but want %v", unmarshaledPlatform, unmarshaledPlatform)
			}
		})
	}
}
