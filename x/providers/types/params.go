package types

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"time"

	"github.com/vipernet-xyz/viper-network/types"
)

// POS params default values
const (
	// DefaultParamspace for params keeper
	DefaultParamspace                        = ModuleName
	DefaultUnstakingTime                     = time.Hour * 24 * 7 * 3
	DefaultMaxProviders                int64 = math.MaxInt64
	DefaultMinStake                    int64 = 10000
	DefaultBaseRelaysPerVIPR           int64 = 200000
	DefaultStabilityModulation         int64 = 0
	DefaultParticipationRateOn         bool  = false
	DefaultMaxChains                   int64 = 15
	DefaultMinNumServicers                   = int32(3)
	DefaultMaxNumServicers                   = int32(25)
	DefaultMaxFreeTierRelaysPerSession       = int64(5000)
)

// Keys for parameter access
var (
	KeyUnstakingTime               = []byte("ProviderUnstakingTime")
	KeyMaxProviders                = []byte("MaxProviders")
	KeyMinProviderStake            = []byte("MinimumProviderStake")
	BaseRelaysPerVIPR              = []byte("BaseRelaysPerVIPR")
	StabilityModulation            = []byte("StabilityModulation")
	ParticipationRate              = []byte("ParticipationRate")
	KeyMaximumChains               = []byte("MaximumChains")
	KeyMinNumServicers             = []byte("MinNumServicers")
	KeyMaxNumServicers             = []byte("MaxNumServicers")
	KeyMaxFreeTierRelaysPerSession = []byte("MaxFreeTierRelaysPerSession")
)

var _ types.ParamSet = (*Params)(nil)

// Params defines the high level settings for pos module
type Params struct {
	UnstakingTime               time.Duration `json:"unstaking_time" yaml:"unstaking_time"`                 // duration of unstaking
	MaxProviders                int64         `json:"max_providers" yaml:"max_providers"`                   // maximum number of providers
	MinProviderStake            int64         `json:"minimum_provider_stake" yaml:"minimum_provider_stake"` // minimum amount needed to stake as an provider
	BaseRelaysPerVIPR           int64         `json:"base_relays_per_vip" yaml:"base_relays_per_vip"`       // base relays per VIPR coin staked
	StabilityModulation         int64         `json:"stability_modulation" yaml:"stability_modulation"`     // the stability adjustment from the governance
	ParticipationRate           bool          `json:"participation_rate_on" yaml:"participation_rate_on"`   // the participation rate affects the amount minted based on staked ratio
	MaxChains                   int64         `json:"maximum_chains" yaml:"maximum_chains"`                 // the maximum number of chains an provider can stake for
	MinNumServicers             int32         `json:"minimum_number_servicers"`
	MaxNumServicers             int32         `json:"maximum_number_servicers"`
	MaxFreeTierRelaysPerSession int64         `json:"maximum_free_tier_relays_per_session"`
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
		{Key: KeyMinNumServicers, Value: p.MinNumServicers},
		{Key: KeyMaxNumServicers, Value: p.MaxNumServicers},
		{Key: KeyMaxFreeTierRelaysPerSession, Value: p.MaxFreeTierRelaysPerSession},
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return Params{
		UnstakingTime:               DefaultUnstakingTime,
		MaxProviders:                DefaultMaxProviders,
		MinProviderStake:            DefaultMinStake,
		BaseRelaysPerVIPR:           DefaultBaseRelaysPerVIPR,
		StabilityModulation:         DefaultStabilityModulation,
		ParticipationRate:           DefaultParticipationRateOn,
		MaxChains:                   DefaultMaxChains,
		MinNumServicers:             DefaultMinNumServicers,
		MaxNumServicers:             DefaultMaxNumServicers,
		MaxFreeTierRelaysPerSession: DefaultMaxFreeTierRelaysPerSession,
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
	if p.MaxFreeTierRelaysPerSession < 0 {
		return fmt.Errorf("invalid max free tier relays per session, must be above 0")
	}
	// session count constraints
	if p.MaxNumServicers > 100 || p.MaxNumServicers < 1 {
		return errors.New("invalid Max session servicer count")
	}
	if p.MinNumServicers > 100 || p.MinNumServicers < 1 {
		return errors.New("invalid Min session servicer count")
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
  MaxChains                    %d
  MinNumServicers              %d
  MaxNumServicers              %d
  MaxFreeTierRelaysPerSession  %d,`,
		p.UnstakingTime,
		p.MaxProviders,
		p.MinProviderStake,
		p.BaseRelaysPerVIPR,
		p.StabilityModulation,
		p.ParticipationRate,
		p.MaxChains,
		p.MinNumServicers,
		p.MaxNumServicers,
		p.MaxFreeTierRelaysPerSession)
}
