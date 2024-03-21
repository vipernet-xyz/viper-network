package types

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/vipernet-xyz/viper-network/types"
)

// POS params default values
const (
	// DefaultParamspace for params keeper
	DefaultParamspace                 = ModuleName
	DefaultClaimSubmissionWindow      = int64(3)       // default sessions to submit a claim
	DefaultClaimExpiration            = int64(24)      // default servicers to exprie claims
	DefaultReplayAttackBurnMultiplier = int64(3)       // default replay attack burn multiplier
	DefaultMinimumNumberOfProofs      = int64(10)      // default minimum number of proofs //change back
	DefaultBlockByteSize              = int64(8000000) // default block size in bytes
	DefaultMinimumSampleRelays        = int64(2)       //change back 25
	DefaultReportCardSubmissionWindow = int64(3)
)

var (
	DefaultSupportedBlockchains   = []string{"0001", "0002"} //change back
	DefaultSupportedGeoZones      = []string{"0001"}         //change back
	KeyClaimSubmissionWindow      = []byte("ClaimSubmissionWindow")
	KeySupportedBlockchains       = []byte("SupportedBlockchains")
	KeyClaimExpiration            = []byte("ClaimExpiration")
	KeyReplayAttackBurnMultiplier = []byte("ReplayAttackBurnMultiplier")
	KeyMinimumNumberOfProofs      = []byte("MinimumNumberOfProofs")
	KeyBlockByteSize              = []byte("BlockByteSize")
	KeySupportedGeoZones          = []byte("SupportedGeoZones")
	KeyMinimumSampleRelays        = []byte("MinimumSampleRelays")
	KeyReportCardSubmissionWindow = []byte("ReportCardSubmissionWindow")
)

var _ types.ParamSet = (*Params)(nil)

// "Params" - defines the governance set, high level settings for vipernet module
type Params struct {
	ClaimSubmissionWindow      int64    `json:"proof_waiting_period"`
	SupportedBlockchains       []string `json:"supported_blockchains"`
	ClaimExpiration            int64    `json:"claim_expiration"` // per session
	ReplayAttackBurnMultiplier int64    `json:"replay_attack_burn_multiplier"`
	MinimumNumberOfProofs      int64    `json:"minimum_number_of_proofs"`
	BlockByteSize              int64    `json:"block_byte_size,omitempty"`
	SupportedGeoZones          []string `json:"supported_geo_zones"`
	MinimumSampleRelays        int64    `json:"minimum_sample_relays"`
	ReportCardSubmissionWindow int64    `json:"report_card_submission_window"`
}

// "ParamSetPairs" - returns an kv params object
// Note: Implements params.ParamSet
func (p *Params) ParamSetPairs() types.ParamSetPairs {
	return types.ParamSetPairs{
		{Key: KeyClaimSubmissionWindow, Value: &p.ClaimSubmissionWindow},
		{Key: KeySupportedBlockchains, Value: &p.SupportedBlockchains},
		{Key: KeyClaimExpiration, Value: &p.ClaimExpiration},
		{Key: KeyReplayAttackBurnMultiplier, Value: p.ReplayAttackBurnMultiplier},
		{Key: KeyMinimumNumberOfProofs, Value: p.MinimumNumberOfProofs},
		{Key: KeyBlockByteSize, Value: p.BlockByteSize},
		{Key: KeySupportedGeoZones, Value: p.SupportedGeoZones},
		{Key: KeyMinimumSampleRelays, Value: p.MinimumSampleRelays},
		{Key: KeyReportCardSubmissionWindow, Value: p.ReportCardSubmissionWindow},
	}
}

// "DefaultParams" - Returns a default set of parameters
func DefaultParams() Params {
	return Params{
		ClaimSubmissionWindow:      DefaultClaimSubmissionWindow,
		SupportedBlockchains:       DefaultSupportedBlockchains,
		ClaimExpiration:            DefaultClaimExpiration,
		ReplayAttackBurnMultiplier: DefaultReplayAttackBurnMultiplier,
		MinimumNumberOfProofs:      DefaultMinimumNumberOfProofs,
		MinimumSampleRelays:        DefaultMinimumSampleRelays,
		BlockByteSize:              DefaultBlockByteSize,
		SupportedGeoZones:          DefaultSupportedGeoZones,
		ReportCardSubmissionWindow: DefaultReportCardSubmissionWindow,
	}
}

// "Validate" - Validate a set of params
func (p Params) Validate() error {
	// claim submission window constraints
	if p.ClaimSubmissionWindow < 2 {
		return errors.New("waiting period must be at least 2 sessions")
	}
	// verify each supported blockchain
	for _, chain := range p.SupportedBlockchains {
		if err := NetworkIdentifierVerification(chain); err != nil {
			return err
		}
	}
	// verify each supported GeoZone
	for _, geoZone := range p.SupportedGeoZones {
		if err := GeoZoneIdentifierVerification(geoZone); err != nil {
			return err
		}
	}
	// ensure replay attack burn multiplier is above 0
	if p.ReplayAttackBurnMultiplier < 0 {
		return errors.New("invalid replay attack burn multiplier")
	}
	// ensure claim expiration
	if p.ClaimExpiration < 0 {
		return errors.New("invalid claim expiration")
	}
	if p.ClaimExpiration < p.ClaimSubmissionWindow {
		return errors.New("unverified Proof expiration is far too short, must be greater than Proof waiting period")
	}
	if p.ReportCardSubmissionWindow < 1 {
		return errors.New("report card submission window cannot be less than one session")
	}
	return nil
}

// "Equal" - Checks the equality of two param objects
func (p Params) Equal(p2 Params) bool {
	return reflect.DeepEqual(p, p2)
}

// "String" -  returns a human readable string representation of the parameters
func (p Params) String() string {
	return fmt.Sprintf(`Params:
  ClaimSubmissionWindow:     %d
  Supported Blockchains      %v
  ClaimExpiration            %d
  ReplayAttackBurnMultiplier %d
  BlockByteSize              %d
  Supported GeoZones         %v
  MinimumSampleRelays        %d
  ReportCardSubmissionWindow %d
`,
		p.ClaimSubmissionWindow,
		p.SupportedBlockchains,
		p.ClaimExpiration,
		p.ReplayAttackBurnMultiplier,
		p.BlockByteSize,
		p.SupportedGeoZones,
		p.MinimumSampleRelays,
		p.ReportCardSubmissionWindow)
}
