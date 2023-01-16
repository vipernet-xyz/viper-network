package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/vipernet-xyz/viper-network/codec"

	"github.com/vipernet-xyz/viper-network/crypto"

	sdk "github.com/vipernet-xyz/viper-network/types"
)

// Platform represents a viper network decentralized platform. Platforms stake in the network for relay throughput.
type Platform struct {
	Address                 sdk.Address      `json:"address" yaml:"address"`               // address of the platform; hex encoded in JSON
	PublicKey               crypto.PublicKey `json:"public_key" yaml:"public_key"`         // the public key of the platform; hex encoded in JSON
	Jailed                  bool             `json:"jailed" yaml:"jailed"`                 // has the platform been jailed from staked status?
	Status                  sdk.StakeStatus  `json:"status" yaml:"status"`                 // platform status (staked/unstaking/unstaked)
	Chains                  []string         `json:"chains" yaml:"chains"`                 // requested chains
	StakedTokens            sdk.BigInt       `json:"tokens" yaml:"tokens"`                 // tokens staked in the network
	MaxRelays               sdk.BigInt       `json:"max_relays" yaml:"max_relays"`         // maximum number of relays allowed
	UnstakingCompletionTime time.Time        `json:"unstaking_time" yaml:"unstaking_time"` // if unstaking, min time for the platform to complete unstaking
}

// NewPlatform - initialize a new instance of an platform
func NewPlatform(addr sdk.Address, publicKey crypto.PublicKey, chains []string, tokensToStake sdk.BigInt) Platform {
	return Platform{
		Address:                 addr,
		PublicKey:               publicKey,
		Jailed:                  false,
		Status:                  sdk.Staked,
		Chains:                  chains,
		StakedTokens:            tokensToStake,
		UnstakingCompletionTime: time.Time{}, // zero out because status: staked
	}
}

// get the consensus-engine power
// a reduction of 10^6 from platform tokens is platformlied
func (a Platform) ConsensusPower() int64 {
	if a.IsStaked() {
		return sdk.TokensToConsensusPower(a.StakedTokens)
	}
	return 0
}

// RemoveStakedTokens removes tokens from a platform
func (a Platform) RemoveStakedTokens(tokens sdk.BigInt) (Platform, error) {
	if tokens.IsNegative() {
		return Platform{}, fmt.Errorf("should not hplatformen: trying to remove negative tokens %v", tokens)
	}
	if a.StakedTokens.LT(tokens) {
		return Platform{}, fmt.Errorf("should not hplatformen: only have %v tokens, trying to remove %v", a.StakedTokens, tokens)
	}
	a.StakedTokens = a.StakedTokens.Sub(tokens)
	return a, nil
}

// AddStakedTokens tokens to staked field for a platform
func (a Platform) AddStakedTokens(tokens sdk.BigInt) (Platform, error) {
	if tokens.IsNegative() {
		return Platform{}, fmt.Errorf("should not hplatformen: trying to remove negative tokens %v", tokens)
	}
	a.StakedTokens = a.StakedTokens.Add(tokens)
	return a, nil
}

// compares the vital fields of two platform structures
func (a Platform) Equals(v2 Platform) bool {
	return a.PublicKey.Equals(v2.PublicKey) &&
		bytes.Equal(a.Address, v2.Address) &&
		a.Status.Equal(v2.Status) &&
		a.StakedTokens.Equal(v2.StakedTokens)
}

// UpdateStatus updates the staking status
func (a Platform) UpdateStatus(newStatus sdk.StakeStatus) Platform {
	a.Status = newStatus
	return a
}

func (a Platform) GetChains() []string            { return a.Chains }
func (a Platform) IsStaked() bool                 { return a.GetStatus().Equal(sdk.Staked) }
func (a Platform) IsUnstaked() bool               { return a.GetStatus().Equal(sdk.Unstaked) }
func (a Platform) IsUnstaking() bool              { return a.GetStatus().Equal(sdk.Unstaking) }
func (a Platform) IsJailed() bool                 { return a.Jailed }
func (a Platform) GetStatus() sdk.StakeStatus     { return a.Status }
func (a Platform) GetAddress() sdk.Address        { return a.Address }
func (a Platform) GetPublicKey() crypto.PublicKey { return a.PublicKey }
func (a Platform) GetTokens() sdk.BigInt          { return a.StakedTokens }
func (a Platform) GetConsensusPower() int64       { return a.ConsensusPower() }
func (a Platform) GetMaxRelays() sdk.BigInt       { return a.MaxRelays }

var _ codec.ProtoMarshaler = &Platform{}

func (a *Platform) Reset() {
	*a = Platform{}
}

func (a Platform) ProtoMessage() {
	p := a.ToProto()
	p.ProtoMessage()
}

func (a Platform) Marshal() ([]byte, error) {
	p := a.ToProto()
	return p.Marshal()
}

func (a Platform) MarshalTo(data []byte) (n int, err error) {
	p := a.ToProto()
	return p.MarshalTo(data)
}

func (a Platform) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	p := a.ToProto()
	return p.MarshalToSizedBuffer(dAtA)
}

func (a Platform) Size() int {
	p := a.ToProto()
	return p.Size()
}

func (a *Platform) Unmarshal(data []byte) (err error) {
	var pa ProtoPlatform
	err = pa.Unmarshal(data)
	if err != nil {
		return err
	}
	*a, err = pa.FromProto()
	return
}

func (a Platform) ToProto() ProtoPlatform {
	return ProtoPlatform{
		Address:                 a.Address,
		PublicKey:               a.PublicKey.RawBytes(),
		Jailed:                  a.Jailed,
		Status:                  a.Status,
		Chains:                  a.Chains,
		StakedTokens:            a.StakedTokens,
		MaxRelays:               a.MaxRelays,
		UnstakingCompletionTime: a.UnstakingCompletionTime,
	}
}

func (ae ProtoPlatform) FromProto() (Platform, error) {
	pk, err := crypto.NewPublicKeyBz(ae.PublicKey)
	if err != nil {
		return Platform{}, err
	}
	return Platform{
		Address:                 ae.Address,
		PublicKey:               pk,
		Jailed:                  ae.Jailed,
		Status:                  ae.Status,
		Chains:                  ae.Chains,
		StakedTokens:            ae.StakedTokens,
		MaxRelays:               ae.MaxRelays,
		UnstakingCompletionTime: ae.UnstakingCompletionTime,
	}, nil
}

// Platforms is a slice of type platform.
type Platforms []Platform

func (a Platforms) String() (out string) {
	for _, val := range a {
		out += val.String() + "\n\n"
	}
	return strings.TrimSpace(out)
}

// String returns a human readable string representation of a platform.
func (a Platform) String() string {
	return fmt.Sprintf("Address:\t\t%s\nPublic Key:\t\t%s\nJailed:\t\t\t%v\nChains:\t\t\t%v\nMaxRelays:\t\t%v\nStatus:\t\t\t%s\nTokens:\t\t\t%s\nUnstaking Time:\t%v\n----\n",
		a.Address, a.PublicKey.RawString(), a.Jailed, a.Chains, a.MaxRelays, a.Status, a.StakedTokens, a.UnstakingCompletionTime,
	)
}

// this is a helper struct used for JSON de- and encoding only
type JSONPlatform struct {
	Address                 sdk.Address     `json:"address" yaml:"address"`               // the hex address of the platform
	PublicKey               string          `json:"public_key" yaml:"public_key"`         // the hex consensus public key of the platform
	Jailed                  bool            `json:"jailed" yaml:"jailed"`                 // has the platform been jailed from staked status?
	Chains                  []string        `json:"chains" yaml:"chains"`                 // non native (external) blockchains needed for the platform
	MaxRelays               sdk.BigInt      `json:"max_relays" yaml:"max_relays"`         // maximum number of relays allowed for the platform
	Status                  sdk.StakeStatus `json:"status" yaml:"status"`                 // platform status (staked/unstaking/unstaked)
	StakedTokens            sdk.BigInt      `json:"staked_tokens" yaml:"staked_tokens"`   // how many staked tokens
	UnstakingCompletionTime time.Time       `json:"unstaking_time" yaml:"unstaking_time"` // if unstaking, min time for the platform to complete unstaking
}

// marshal structure into JSON encoding
func (a Platforms) JSON() (out []byte, err error) {
	return json.Marshal(a)
}

// MarshalJSON marshals the platform to JSON using raw Hex for the public key
func (a Platform) MarshalJSON() ([]byte, error) {
	return ModuleCdc.MarshalJSON(JSONPlatform{
		Address:                 a.Address,
		PublicKey:               a.PublicKey.RawString(),
		Jailed:                  a.Jailed,
		Status:                  a.Status,
		Chains:                  a.Chains,
		MaxRelays:               a.MaxRelays,
		StakedTokens:            a.StakedTokens,
		UnstakingCompletionTime: a.UnstakingCompletionTime,
	})
}

// UnmarshalJSON unmarshals the platform from JSON using raw hex for the public key
func (a *Platform) UnmarshalJSON(data []byte) error {
	bv := &JSONPlatform{}
	if err := ModuleCdc.UnmarshalJSON(data, bv); err != nil {
		return err
	}
	consPubKey, err := crypto.NewPublicKey(bv.PublicKey)
	if err != nil {
		return err
	}
	*a = Platform{
		Address:                 bv.Address,
		PublicKey:               consPubKey,
		Chains:                  bv.Chains,
		MaxRelays:               bv.MaxRelays,
		Jailed:                  bv.Jailed,
		StakedTokens:            bv.StakedTokens,
		Status:                  bv.Status,
		UnstakingCompletionTime: bv.UnstakingCompletionTime,
	}
	return nil
}

// unmarshal the platform
func MarshalPlatform(cdc *codec.Codec, ctx sdk.Ctx, platform Platform) (result []byte, err error) {
	return cdc.MarshalBinaryLengthPrefixed(&platform, ctx.BlockHeight())
}

// unmarshal the platform
func UnmarshalPlatform(cdc *codec.Codec, ctx sdk.Ctx, platformBytes []byte) (platform Platform, err error) {
	err = cdc.UnmarshalBinaryLengthPrefixed(platformBytes, &platform, ctx.BlockHeight())
	return
}

type PlatformsPage struct {
	Result Platforms `json:"result"`
	Total  int       `json:"total_pages"`
	Page   int       `json:"page"`
}

// String returns a human readable string representation of a validator page
func (aP PlatformsPage) String() string {
	return fmt.Sprintf("Total:\t\t%d\nPage:\t\t%d\nResult:\t\t\n====\n%s\n====\n", aP.Total, aP.Page, aP.Result.String())
}
