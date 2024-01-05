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

// Requestor represents a viper network decentralized requestor. Requestors stake in the network for relay throughput.
type Requestor struct {
	Address                 sdk.Address      `json:"address" yaml:"address"`       // address of the requestor; hex encoded in JSON
	PublicKey               crypto.PublicKey `json:"public_key" yaml:"public_key"` // the public key of the requestor; hex encoded in JSON
	Jailed                  bool             `json:"jailed" yaml:"jailed"`         // has the requestor been jailed from staked status?
	Status                  sdk.StakeStatus  `json:"status" yaml:"status"`         // requestor status (staked/unstaking/unstaked)
	Chains                  []string         `json:"chains" yaml:"chains"`         // requested chains
	StakedTokens            sdk.BigInt       `json:"tokens" yaml:"tokens"`         // tokens staked in the network
	MaxRelays               sdk.BigInt       `json:"max_relays" yaml:"max_relays"` // maximum number of relays allowed
	GeoZones                []string         `json:"geo_zone" yaml:"geo_zone"`     //geo location
	NumServicers            int32            `json:"num_servicers" yaml:"num_servicers"`
	UnstakingCompletionTime time.Time        `json:"unstaking_time" yaml:"unstaking_time"` // if unstaking, min time for the requestor to complete unstaking
}

// NewRequestor - initialize a new instance of an requestor
func NewRequestor(addr sdk.Address, publicKey crypto.PublicKey, chains []string, tokensToStake sdk.BigInt, geoZones []string, numServicers int32) Requestor {
	return Requestor{
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
// a reduction of 10^6 from requestor tokens is requestorlied
func (a Requestor) ConsensusPower() int64 {
	if a.IsStaked() {
		return sdk.TokensToConsensusPower(a.StakedTokens)
	}
	return 0
}

// RemoveStakedTokens removes tokens from a requestor
func (a Requestor) RemoveStakedTokens(tokens sdk.BigInt) (Requestor, error) {
	if tokens.IsNegative() {
		return Requestor{}, fmt.Errorf("should not happen: trying to remove negative tokens %v", tokens)
	}
	if a.StakedTokens.LT(tokens) {
		return Requestor{}, fmt.Errorf("should not happen: only have %v tokens, trying to remove %v", a.StakedTokens, tokens)
	}
	a.StakedTokens = a.StakedTokens.Sub(tokens)
	return a, nil
}

// AddStakedTokens tokens to staked field for a requestor
func (a Requestor) AddStakedTokens(tokens sdk.BigInt) (Requestor, error) {
	if tokens.IsNegative() {
		return Requestor{}, fmt.Errorf("should not happen: trying to remove negative tokens %v", tokens)
	}
	a.StakedTokens = a.StakedTokens.Add(tokens)
	return a, nil
}

// compares the vital fields of two requestor structures
func (a Requestor) Equals(v2 Requestor) bool {
	return a.PublicKey.Equals(v2.PublicKey) &&
		bytes.Equal(a.Address, v2.Address) &&
		a.Status.Equal(v2.Status) &&
		a.StakedTokens.Equal(v2.StakedTokens)
}

// UpdateStatus updates the staking status
func (a Requestor) UpdateStatus(newStatus sdk.StakeStatus) Requestor {
	a.Status = newStatus
	return a
}

func (a Requestor) GetChains() []string            { return a.Chains }
func (a Requestor) IsStaked() bool                 { return a.GetStatus().Equal(sdk.Staked) }
func (a Requestor) IsUnstaked() bool               { return a.GetStatus().Equal(sdk.Unstaked) }
func (a Requestor) IsUnstaking() bool              { return a.GetStatus().Equal(sdk.Unstaking) }
func (a Requestor) IsJailed() bool                 { return a.Jailed }
func (a Requestor) GetStatus() sdk.StakeStatus     { return a.Status }
func (a Requestor) GetAddress() sdk.Address        { return a.Address }
func (a Requestor) GetPublicKey() crypto.PublicKey { return a.PublicKey }
func (a Requestor) GetTokens() sdk.BigInt          { return a.StakedTokens }
func (a Requestor) GetConsensusPower() int64       { return a.ConsensusPower() }
func (a Requestor) GetMaxRelays() sdk.BigInt       { return a.MaxRelays }
func (a Requestor) GetGeoZones() []string          { return a.GeoZones }
func (a Requestor) GetNumServicers() int32         { return a.NumServicers }

var _ codec.ProtoMarshaler = &Requestor{}

func (a *Requestor) Reset() {
	*a = Requestor{}
}

func (a Requestor) ProtoMessage() {
	p := a.ToProto()
	p.ProtoMessage()
}

func (a Requestor) Marshal() ([]byte, error) {
	p := a.ToProto()
	return p.Marshal()
}

func (a Requestor) MarshalTo(data []byte) (n int, err error) {
	p := a.ToProto()
	return p.MarshalTo(data)
}

func (a Requestor) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	p := a.ToProto()
	return p.MarshalToSizedBuffer(dAtA)
}

func (a Requestor) Size() int {
	p := a.ToProto()
	return p.Size()
}

func (a *Requestor) Unmarshal(data []byte) (err error) {
	var pa ProtoRequestor
	err = pa.Unmarshal(data)
	if err != nil {
		return err
	}
	*a, err = pa.FromProto()
	return
}

func (a Requestor) ToProto() ProtoRequestor {
	return ProtoRequestor{
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

func (ae ProtoRequestor) FromProto() (Requestor, error) {
	pk, err := crypto.NewPublicKeyBz(ae.PublicKey)
	if err != nil {
		return Requestor{}, err
	}
	return Requestor{
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

// Requestors is a slice of type requestor.
type Requestors []Requestor

func (a Requestors) String() (out string) {
	for _, val := range a {
		out += val.String() + "\n\n"
	}
	return strings.TrimSpace(out)
}

// String returns a human readable string representation of a requestor.
func (a Requestor) String() string {
	return fmt.Sprintf("Address:\t\t%s\nPublic Key:\t\t%s\nJailed:\t\t\t%v\nChains:\t\t\t%v\nMaxRelays:\t\t%v\nStatus:\t\t\t%s\nTokens:\t\t\t%s\nGeoZones:\t\t\t%sUnstaking Time:\t%v\n----\n",
		a.Address, a.PublicKey.RawString(), a.Jailed, a.Chains, a.MaxRelays, a.Status, a.StakedTokens, a.GeoZones, a.UnstakingCompletionTime,
	)
}

// this is a helper struct used for JSON de- and encoding only
type JSONRequestor struct {
	Address                 sdk.Address     `json:"address" yaml:"address"`             // the hex address of the requestor
	PublicKey               string          `json:"public_key" yaml:"public_key"`       // the hex consensus public key of the requestor
	Jailed                  bool            `json:"jailed" yaml:"jailed"`               // has the requestor been jailed from staked status?
	Chains                  []string        `json:"chains" yaml:"chains"`               // non native (external) blockchains needed for the requestor
	MaxRelays               sdk.BigInt      `json:"max_relays" yaml:"max_relays"`       // maximum number of relays allowed for the requestor
	Status                  sdk.StakeStatus `json:"status" yaml:"status"`               // requestor status (staked/unstaking/unstaked)
	StakedTokens            sdk.BigInt      `json:"staked_tokens" yaml:"staked_tokens"` // how many staked tokens
	GeoZones                []string        `json:"geo_zones" yaml:"geo_zones"`
	NumServicers            int32           `json:"num_servicers" yaml:"num_servicers"`
	UnstakingCompletionTime time.Time       `json:"unstaking_time" yaml:"unstaking_time"` // if unstaking, min time for the requestor to complete unstaking
}

// marshal structure into JSON encoding
func (a Requestors) JSON() (out []byte, err error) {
	return json.Marshal(a)
}

// MarshalJSON marshals the requestor to JSON using raw Hex for the public key
func (a Requestor) MarshalJSON() ([]byte, error) {
	return ModuleCdc.MarshalJSON(JSONRequestor{
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

// UnmarshalJSON unmarshals the requestor from JSON using raw hex for the public key
func (a *Requestor) UnmarshalJSON(data []byte) error {
	bv := &JSONRequestor{}
	if err := ModuleCdc.UnmarshalJSON(data, bv); err != nil {
		return err
	}
	consPubKey, err := crypto.NewPublicKey(bv.PublicKey)
	if err != nil {
		return err
	}
	*a = Requestor{
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

// marshal the requestor
func MarshalRequestor(cdc *codec.Codec, ctx sdk.Ctx, requestor Requestor) (result []byte, err error) {
	return cdc.MarshalBinaryLengthPrefixed(&requestor)
}

// unmarshal the requestor
func UnmarshalRequestor(cdc *codec.Codec, ctx sdk.Ctx, requestorBytes []byte) (requestor Requestor, err error) {
	err = cdc.UnmarshalBinaryLengthPrefixed(requestorBytes, &requestor)
	return
}

type RequestorsPage struct {
	Result Requestors `json:"result"`
	Total  int        `json:"total_pages"`
	Page   int        `json:"page"`
}

// String returns a human readable string representation of a validator page
func (aP RequestorsPage) String() string {
	return fmt.Sprintf("Total:\t\t%d\nPage:\t\t%d\nResult:\t\t\n====\n%s\n====\n", aP.Total, aP.Page, aP.Result.String())
}
