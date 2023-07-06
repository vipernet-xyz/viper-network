package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/vipernet-xyz/viper-network/codec"

	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"

	sdk "github.com/vipernet-xyz/viper-network/types"
)

// Provider represents a viper network decentralized provider. Providers stake in the network for relay throughput.
type Provider struct {
	Address                 sdk.Address      `json:"address" yaml:"address"`       // address of the provider; hex encoded in JSON
	PublicKey               crypto.PublicKey `json:"public_key" yaml:"public_key"` // the public key of the provider; hex encoded in JSON
	Jailed                  bool             `json:"jailed" yaml:"jailed"`         // has the provider been jailed from staked status?
	Status                  sdk.StakeStatus  `json:"status" yaml:"status"`         // provider status (staked/unstaking/unstaked)
	Chains                  []string         `json:"chains" yaml:"chains"`         // requested chains
	StakedTokens            sdk.BigInt       `json:"tokens" yaml:"tokens"`         // tokens staked in the network
	MaxRelays               sdk.BigInt       `json:"max_relays" yaml:"max_relays"` // maximum number of relays allowed
	GeoZones                []string         `json:"geo_zone" yaml:"geo_zone"`     //geo location
	NumServicers            int8             `json:"num_servicers" yaml:"num_servicers"`
	UnstakingCompletionTime time.Time        `json:"unstaking_time" yaml:"unstaking_time"` // if unstaking, min time for the provider to complete unstaking
}

// NewProvider - initialize a new instance of an provider
func NewProvider(addr sdk.Address, publicKey crypto.PublicKey, chains []string, tokensToStake sdk.BigInt, geoZones []string, numServicers int8) Provider {
	return Provider{
		Address:                 addr,
		PublicKey:               publicKey,
		Jailed:                  false,
		Status:                  sdk.Staked,
		Chains:                  chains,
		GeoZones:                geoZones,
		NumServicers:            numServicers,
		StakedTokens:            tokensToStake,
		UnstakingCompletionTime: time.Time{}, // zero out because status: staked
	}
}

// get the consensus-engine power
// a reduction of 10^6 from provider tokens is providerlied
func (a Provider) ConsensusPower() int64 {
	if a.IsStaked() {
		return sdk.TokensToConsensusPower(a.StakedTokens)
	}
	return 0
}

// RemoveStakedTokens removes tokens from a provider
func (a Provider) RemoveStakedTokens(tokens sdk.BigInt) (Provider, error) {
	if tokens.IsNegative() {
		return Provider{}, fmt.Errorf("should not happen: trying to remove negative tokens %v", tokens)
	}
	if a.StakedTokens.LT(tokens) {
		return Provider{}, fmt.Errorf("should not happen: only have %v tokens, trying to remove %v", a.StakedTokens, tokens)
	}
	a.StakedTokens = a.StakedTokens.Sub(tokens)
	return a, nil
}

// AddStakedTokens tokens to staked field for a provider
func (a Provider) AddStakedTokens(tokens sdk.BigInt) (Provider, error) {
	if tokens.IsNegative() {
		return Provider{}, fmt.Errorf("should not happen: trying to remove negative tokens %v", tokens)
	}
	a.StakedTokens = a.StakedTokens.Add(tokens)
	return a, nil
}

// compares the vital fields of two provider structures
func (a Provider) Equals(v2 Provider) bool {
	return a.PublicKey.Equals(v2.PublicKey) &&
		bytes.Equal(a.Address, v2.Address) &&
		a.Status.Equal(v2.Status) &&
		a.StakedTokens.Equal(v2.StakedTokens)
}

// UpdateStatus updates the staking status
func (a Provider) UpdateStatus(newStatus sdk.StakeStatus) Provider {
	a.Status = newStatus
	return a
}

func (a Provider) GetChains() []string            { return a.Chains }
func (a Provider) IsStaked() bool                 { return a.GetStatus().Equal(sdk.Staked) }
func (a Provider) IsUnstaked() bool               { return a.GetStatus().Equal(sdk.Unstaked) }
func (a Provider) IsUnstaking() bool              { return a.GetStatus().Equal(sdk.Unstaking) }
func (a Provider) IsJailed() bool                 { return a.Jailed }
func (a Provider) GetStatus() sdk.StakeStatus     { return a.Status }
func (a Provider) GetAddress() sdk.Address        { return a.Address }
func (a Provider) GetPublicKey() crypto.PublicKey { return a.PublicKey }
func (a Provider) GetTokens() sdk.BigInt          { return a.StakedTokens }
func (a Provider) GetConsensusPower() int64       { return a.ConsensusPower() }
func (a Provider) GetMaxRelays() sdk.BigInt       { return a.MaxRelays }
func (a Provider) GetGeoZones() []string          { return a.GeoZones }
func (a Provider) GetNumServicers() int8          { return a.NumServicers }

var _ codec.ProtoMarshaler = &Provider{}

func (a *Provider) Reset() {
	*a = Provider{}
}

func (a Provider) ProtoMessage() {
	p := a.ToProto()
	p.ProtoMessage()
}

func (a Provider) Marshal() ([]byte, error) {
	p := a.ToProto()
	return p.Marshal()
}

func (a Provider) MarshalTo(data []byte) (n int, err error) {
	p := a.ToProto()
	return p.MarshalTo(data)
}

func (a Provider) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	p := a.ToProto()
	return p.MarshalToSizedBuffer(dAtA)
}

func (a Provider) Size() int {
	p := a.ToProto()
	return p.Size()
}

func (a *Provider) Unmarshal(data []byte) (err error) {
	var pa ProtoProvider
	err = pa.Unmarshal(data)
	if err != nil {
		return err
	}
	*a, err = pa.FromProto()
	return
}

func (a Provider) ToProto() ProtoProvider {
	return ProtoProvider{
		Address:                 a.Address,
		PublicKey:               a.PublicKey.RawBytes(),
		Jailed:                  a.Jailed,
		Status:                  a.Status,
		Chains:                  a.Chains,
		StakedTokens:            a.StakedTokens,
		MaxRelays:               a.MaxRelays,
		GeoZones:                a.GeoZones,
		NumServicers:            a.NumServicers,
		UnstakingCompletionTime: a.UnstakingCompletionTime,
	}
}

func (ae ProtoProvider) FromProto() (Provider, error) {
	pk, err := crypto.NewPublicKeyBz(ae.PublicKey)
	if err != nil {
		return Provider{}, err
	}
	return Provider{
		Address:                 ae.Address,
		PublicKey:               pk,
		Jailed:                  ae.Jailed,
		Status:                  ae.Status,
		Chains:                  ae.Chains,
		StakedTokens:            ae.StakedTokens,
		MaxRelays:               ae.MaxRelays,
		GeoZones:                ae.GeoZones,
		UnstakingCompletionTime: ae.UnstakingCompletionTime,
	}, nil
}

// Providers is a slice of type provider.
type Providers []Provider

func (a Providers) String() (out string) {
	for _, val := range a {
		out += val.String() + "\n\n"
	}
	return strings.TrimSpace(out)
}

// String returns a human readable string representation of a provider.
func (a Provider) String() string {
	return fmt.Sprintf("Address:\t\t%s\nPublic Key:\t\t%s\nJailed:\t\t\t%v\nChains:\t\t\t%v\nMaxRelays:\t\t%v\nStatus:\t\t\t%s\nTokens:\t\t\t%s\nGeoZones:\t\t\t%sUnstaking Time:\t%v\n----\n",
		a.Address, a.PublicKey.RawString(), a.Jailed, a.Chains, a.MaxRelays, a.Status, a.StakedTokens, a.GeoZones, a.UnstakingCompletionTime,
	)
}

// this is a helper struct used for JSON de- and encoding only
type JSONProvider struct {
	Address                 sdk.Address     `json:"address" yaml:"address"`             // the hex address of the provider
	PublicKey               string          `json:"public_key" yaml:"public_key"`       // the hex consensus public key of the provider
	Jailed                  bool            `json:"jailed" yaml:"jailed"`               // has the provider been jailed from staked status?
	Chains                  []string        `json:"chains" yaml:"chains"`               // non native (external) blockchains needed for the provider
	MaxRelays               sdk.BigInt      `json:"max_relays" yaml:"max_relays"`       // maximum number of relays allowed for the provider
	Status                  sdk.StakeStatus `json:"status" yaml:"status"`               // provider status (staked/unstaking/unstaked)
	StakedTokens            sdk.BigInt      `json:"staked_tokens" yaml:"staked_tokens"` // how many staked tokens
	GeoZones                []string        `json:"geo_zones" yaml:"geo_zones"`
	NumServicers            int8            `json:"num_servicers" yaml:"num_servicers"`
	UnstakingCompletionTime time.Time       `json:"unstaking_time" yaml:"unstaking_time"` // if unstaking, min time for the provider to complete unstaking
}

// marshal structure into JSON encoding
func (a Providers) JSON() (out []byte, err error) {
	return json.Marshal(a)
}

// MarshalJSON marshals the provider to JSON using raw Hex for the public key
func (a Provider) MarshalJSON() ([]byte, error) {
	return ModuleCdc.MarshalJSON(JSONProvider{
		Address:                 a.Address,
		PublicKey:               a.PublicKey.RawString(),
		Jailed:                  a.Jailed,
		Status:                  a.Status,
		Chains:                  a.Chains,
		MaxRelays:               a.MaxRelays,
		StakedTokens:            a.StakedTokens,
		GeoZones:                a.GeoZones,
		UnstakingCompletionTime: a.UnstakingCompletionTime,
	})
}

// UnmarshalJSON unmarshals the provider from JSON using raw hex for the public key
func (a *Provider) UnmarshalJSON(data []byte) error {
	bv := &JSONProvider{}
	if err := ModuleCdc.UnmarshalJSON(data, bv); err != nil {
		return err
	}
	consPubKey, err := crypto.NewPublicKey(bv.PublicKey)
	if err != nil {
		return err
	}
	*a = Provider{
		Address:                 bv.Address,
		PublicKey:               consPubKey,
		Chains:                  bv.Chains,
		MaxRelays:               bv.MaxRelays,
		Jailed:                  bv.Jailed,
		StakedTokens:            bv.StakedTokens,
		Status:                  bv.Status,
		GeoZones:                bv.GeoZones,
		NumServicers:            bv.NumServicers,
		UnstakingCompletionTime: bv.UnstakingCompletionTime,
	}
	return nil
}

// marshal the provider
func MarshalProvider(cdc *codec.Codec, ctx sdk.Ctx, provider Provider) (result []byte, err error) {
	return cdc.MarshalBinaryLengthPrefixed(&provider, ctx.BlockHeight())
}

// unmarshal the provider
func UnmarshalProvider(cdc *codec.Codec, ctx sdk.Ctx, providerBytes []byte) (provider Provider, err error) {
	err = cdc.UnmarshalBinaryLengthPrefixed(providerBytes, &provider, ctx.BlockHeight())
	return
}

type ProvidersPage struct {
	Result Providers `json:"result"`
	Total  int       `json:"total_pages"`
	Page   int       `json:"page"`
}

// String returns a human readable string representation of a validator page
func (aP ProvidersPage) String() string {
	return fmt.Sprintf("Total:\t\t%d\nPage:\t\t%d\nResult:\t\t\n====\n%s\n====\n", aP.Total, aP.Page, aP.Result.String())
}
