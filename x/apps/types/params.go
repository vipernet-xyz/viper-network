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
	DefaultMaxApplications     int64 = math.MaxInt64
	DefaultMinStake            int64 = 1000000
	DefaultBaseRelaysPerVIP    int64 = 200000
	DefaultStabilityModulation int64 = 0
	DefaultParticipationRateOn bool  = false
	DefaultMaxChains           int64 = 15
)

// Keys for parameter access
var (
	KeyUnstakingTime       = []byte("AppUnstakingTime")
	KeyMaxApplications     = []byte("MaxApplications")
	KeyMinApplicationStake = []byte("MinimumApplicationStake")
	BaseRelaysPerVIPR      = []byte("BaseRelaysPerVIPR")
	StabilityModulation    = []byte("StabilityModulation")
	ParticipationRate      = []byte("ParticipationRate")
	KeyMaximumChains       = []byte("MaximumChains")
)

var _ types.ParamSet = (*Params)(nil)

// Params defines the high level settings for pos module
type Params struct {
	UnstakingTime       time.Duration `json:"unstaking_time" yaml:"unstaking_time"`               // duration of unstaking
	MaxApplications     int64         `json:"max_applications" yaml:"max_applications"`           // maximum number of applications
	MinAppStake         int64         `json:"minimum_app_stake" yaml:"minimum_app_stake"`         // minimum amount needed to stake as an application
	BaseRelaysPerVIPR   int64         `json:"base_relays_per_vip" yaml:"base_relays_per_vip"`     // base relays per VIPR coin staked
	StabilityModulation int64         `json:"stability_modulation" yaml:"stability_modulation"`   // the stability adjustment from the governance
	ParticipationRate   bool          `json:"participation_rate_on" yaml:"participation_rate_on"` // the participation rate affects the amount minted based on staked ratio
	MaxChains           int64         `json:"maximum_chains" yaml:"maximum_chains"`               // the maximum number of chains an app can stake for
}

// Implements params.ParamSet
func (p *Params) ParamSetPairs() types.ParamSetPairs {
	return types.ParamSetPairs{
		{Key: KeyUnstakingTime, Value: &p.UnstakingTime},
		{Key: KeyMaxApplications, Value: &p.MaxApplications},
		{Key: KeyMinApplicationStake, Value: &p.MinAppStake},
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
		MaxApplications:     DefaultMaxApplications,
		MinAppStake:         DefaultMinStake,
		BaseRelaysPerVIPR:   DefaultBaseRelaysPerVIP,
		StabilityModulation: DefaultStabilityModulation,
		ParticipationRate:   DefaultParticipationRateOn,
		MaxChains:           DefaultMaxChains,
	}
}

// Validate a set of params
func (p Params) Validate() error {
	if p.MaxApplications == 0 {
		return fmt.Errorf("staking parameter MaxApplications must be a positive integer")
	}
	if p.MinAppStake < DefaultMinStake {
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
  Max Applications:            %d
  Minimum Stake:     	       %d
  BaseRelaysPerVIPR            %d
  Stability Adjustment         %d
  Participation Rate On        %v
  MaxChains                    %d,`,
		p.UnstakingTime,
		p.MaxApplications,
		p.MinAppStake,
		p.BaseRelaysPerVIPR,
		p.StabilityModulation,
		p.ParticipationRate,
		p.MaxChains)
}
