package types

import (
	"fmt"
	"math"
	"reflect"
	"time"

	"github.com/vipernet-xyz/viper-network/types"
)

// POS params default values
const (
	// DefaultParamspace for params keeper
	DefaultParamspace                = ModuleName
	DefaultUnstakingTime             = time.Hour * 24 * 7 * 3
	DefaultMaxProviders        int64 = math.MaxInt64
	DefaultMinStake            int64 = 1000000
	DefaultBaseRelaysPerVIPR   int64 = 200000
	DefaultStabilityModulation int64 = 0
	DefaultParticipationRateOn bool  = false
	DefaultMaxChains           int64 = 15
)

// Keys for parameter access
var (
	KeyUnstakingTime    = []byte("ProviderUnstakingTime")
	KeyMaxProviders     = []byte("MaxProviders")
	KeyMinProviderStake = []byte("MinimumProviderStake")
	BaseRelaysPerVIPR   = []byte("BaseRelaysPerVIPR")
	StabilityModulation = []byte("StabilityModulation")
	ParticipationRate   = []byte("ParticipationRate")
	KeyMaximumChains    = []byte("MaximumChains")
)

var _ types.ParamSet = (*Params)(nil)

// Params defines the high level settings for pos module
type Params struct {
	UnstakingTime       time.Duration `json:"unstaking_time" yaml:"unstaking_time"`                 // duration of unstaking
	MaxProviders        int64         `json:"max_providers" yaml:"max_providers"`                   // maximum number of providers
	MinProviderStake    int64         `json:"minimum_provider_stake" yaml:"minimum_provider_stake"` // minimum amount needed to stake as an provider
	BaseRelaysPerVIPR   int64         `json:"base_relays_per_vip" yaml:"base_relays_per_vip"`       // base relays per VIPR coin staked
	StabilityModulation int64         `json:"stability_modulation" yaml:"stability_modulation"`     // the stability adjustment from the governance
	ParticipationRate   bool          `json:"participation_rate_on" yaml:"participation_rate_on"`   // the participation rate affects the amount minted based on staked ratio
	MaxChains           int64         `json:"maximum_chains" yaml:"maximum_chains"`                 // the maximum number of chains an provider can stake for
}

// Implements params.ParamSet
func (p *Params) ParamSetPairs() types.ParamSetPairs {
	return types.ParamSetPairs{
		{Key: KeyUnstakingTime, Value: &p.UnstakingTime},
		{Key: KeyMaxProviders, Value: &p.MaxProviders},
		{Key: KeyMinProviderStake, Value: &p.MinProviderStake},
		{Key: BaseRelaysPerVIPR, Value: &p.BaseRelaysPerVIPR},
		{Key: StabilityModulation, Value: &p.StabilityModulation},
		{Key: ParticipationRate, Value: &p.ParticipationRate},
		{Key: KeyMaximumChains, Value: &p.MaxChains},
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return Params{
		UnstakingTime:       DefaultUnstakingTime,
		MaxProviders:        DefaultMaxProviders,
		MinProviderStake:    DefaultMinStake,
		BaseRelaysPerVIPR:   DefaultBaseRelaysPerVIPR,
		StabilityModulation: DefaultStabilityModulation,
		ParticipationRate:   DefaultParticipationRateOn,
		MaxChains:           DefaultMaxChains,
	}
}

// Validate a set of params
func (p Params) Validate() error {
	if p.MaxProviders == 0 {
		return fmt.Errorf("staking parameter MaxProviders must be a positive integer")
	}
	if p.MinProviderStake < DefaultMinStake {
		return fmt.Errorf("staking parameter StakeMimimum must be a positive integer")
	}
	if p.BaseRelaysPerVIPR < 0 {
		return fmt.Errorf("invalid baseline throughput stake rate, must be above 0")
	}
	// todo
	return nil
}

// Checks the equality of two param objects
func (p Params) Equal(p2 Params) bool {
	return reflect.DeepEqual(p, p2)
}

// String returns a human readable string representation of the parameters.
func (p Params) String() string {
	return fmt.Sprintf(`Params:
  Unstaking Time:              %s
  Max Providers:               %d
  Minimum Stake:     	       %d
  BaseRelaysPerVIPR            %d
  Stability Adjustment         %d
  Participation Rate On        %v
  MaxChains                    %d,`,
		p.UnstakingTime,
		p.MaxProviders,
		p.MinProviderStake,
		p.BaseRelaysPerVIPR,
		p.StabilityModulation,
		p.ParticipationRate,
		p.MaxChains)
}
