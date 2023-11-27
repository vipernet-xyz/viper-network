package types

import (
	"fmt"
	"reflect"
	"time"

	sdk "github.com/vipernet-xyz/viper-network/types"
)

// POS params default values
const (
	// DefaultParamspace for params keeper
	DefaultTokenRewardFactor    int64 = 1000
	DefaultParamspace                 = ModuleName
	DefaultUnstakingTime              = time.Hour * 24 * 7 * 3
	DefaultMaxValidators        int64 = 100
	DefaultMinStake             int64 = 10000000000
	DefaultMaxEvidenceAge             = 60 * 2 * time.Second
	DefaultSignedBlocksWindow         = int64(10)
	DefaultDowntimeJailDuration       = 60 * 60 * time.Second
	DefaultSessionBlocktime           = 4
	DefaultProposerAllocation         = 5
	DefaultDAOAllocation              = 10
	DefaultProviderAllocation         = 5
	DefaultFishermenAllocation        = 5
	DefaultMaxChains                  = 15
	DefaultMaxJailedBlocks            = 2000
	DefaultServicerCountLock    bool  = false
	DefaultBurnActive           bool  = false
	DefaultMinPauseTime               = 60 * 10 * time.Second
	DefaultMaxFishermen               = int64(50)
	DefaultFishermenCount             = int64(1)
)

// - Keys for parameter access
var (
	KeyUnstakingTime               = []byte("UnstakingTime")
	KeyMaxValidators               = []byte("MaxValidators")
	KeyStakeDenom                  = []byte("StakeDenom")
	KeyStakeMinimum                = []byte("StakeMinimum")
	KeyMaxEvidenceAge              = []byte("MaxEvidenceAge")
	KeySignedBlocksWindow          = []byte("SignedBlocksWindow")
	KeyMinSignedPerWindow          = []byte("MinSignedPerWindow")
	KeyDowntimeJailDuration        = []byte("DowntimeJailDuration")
	KeySlashFractionDoubleSign     = []byte("SlashFractionDoubleSign")
	KeySlashFractionDowntime       = []byte("SlashFractionDowntime")
	KeyTokenRewardFactor           = []byte("TokenRewardFactor")
	KeySessionBlock                = []byte("BlocksPerSession")
	KeyDAOAllocation               = []byte("DAOAllocation")
	KeyProviderAllocation          = []byte("ProviderAllocation")
	KeyFishermenAllocation         = []byte("FishermenAllocation")
	KeyProposerAllocation          = []byte("ProposerPercentage")
	KeyMaxChains                   = []byte("MaximumChains")
	KeyMaxJailedBlocks             = []byte("MaxJailedBlocks")
	DoubleSignJailEndTime          = time.Unix(253402300799, 0) // forever
	DefaultMinSignedPerWindow      = sdk.NewDecWithPrec(5, 1)
	DefaultSlashFractionDoubleSign = sdk.NewDec(1).Quo(sdk.NewDec(1000000))
	DefaultSlashFractionDowntime   = sdk.NewDec(1).Quo(sdk.NewDec(1000000))
	ServicerCountLock              = []byte("ServicerCountLock")
	BurnActive                     = []byte("BurnActive")
	KeyMinPauseTime                = []byte("MinPauseTime")
	KeyMaxFishermen                = []byte("MaxFishermen")
	KeyFishermenCount              = []byte("FishermenCount")
	KeySlashFractionNoActivity     = []byte("SlashFractionNoActivity")
	DefaultSlashFractionNoActivity = sdk.NewDec(1).Quo(sdk.NewDec(1000000))
	KeyLatencyScoreWeight          = []byte("LatencyScoreWeight")
	DefaultLatencyScoreWeight      = sdk.NewDecWithPrec(4, 1)
	KeyAvailabilityScoreWeight     = []byte("AvailabilityScoreWeigh")
	DefaultAvailabilityScoreWeight = sdk.NewDecWithPrec(3, 1)
	KeyReliabilityScoreWeight      = []byte("ReliabilityScoreWeight")
	DefaultReliabilityScoreWeight  = sdk.NewDecWithPrec(3, 1)
	KeySlashFractionFisherman      = []byte("SlashFractionFisherman")
	DefaultSlashFractionFisherman  = sdk.NewDec(2).Quo(sdk.NewDec(100))
)

var _ sdk.ParamSet = (*Params)(nil)

// Params defines the high level settings for pos module
type Params struct {
	TokenRewardFactor       int64         `json:"relays_to_tokens_multiplier" yaml:"relays_to_tokens_multiplier"`
	UnstakingTime           time.Duration `json:"unstaking_time" yaml:"unstaking_time"`                   // how much time must pass between the begin_unstaking_tx and the servicer going to -> unstaked status
	MaxValidators           int64         `json:"max_validators" yaml:"max_validators"`                   // maximum number of validators in the network at any given block
	StakeDenom              string        `json:"stake_denom" yaml:"stake_denom"`                         // the monetary denomination of the coins in the network `uvipr` or `uVipr` or `Wei`
	StakeMinimum            int64         `json:"stake_minimum" yaml:"stake_minimum"`                     // minimum amount of `uvipr` needed to stake in the network as a servicer
	SessionBlockFrequency   int64         `json:"session_block_frequency" yaml:"session_block_frequency"` // how many blocks are in a session (viper network unit)
	DAOAllocation           int64         `json:"dao_allocation" yaml:"dao_allocation"`
	ProviderAllocation      int64         `json:"provider_allocation" yaml:"provider_allocation"`
	ProposerAllocation      int64         `json:"proposer_allocation" yaml:"proposer_allocation"`
	FishermenAllocation     int64         `json:"fisherman_allocation" yaml:"fisherman_allocation"`
	MaximumChains           int64         `json:"maximum_chains" yaml:"maximum_chains"`
	MaxJailedBlocks         int64         `json:"max_jailed_blocks" yaml:"max_jailed_blocks"`
	MaxEvidenceAge          time.Duration `json:"max_evidence_age" yaml:"max_evidence_age"`                     // maximum age of tendermint evidence that is still valid (currently not implemented in Cosmos or Viper-Core)
	SignedBlocksWindow      int64         `json:"signed_blocks_window" yaml:"signed_blocks_window"`             // window of time in blocks (unit) used for signature verification -> specifically in not signing (missing) blocks
	MinSignedPerWindow      sdk.BigDec    `json:"min_signed_per_window" yaml:"min_signed_per_window"`           // minimum number of blocks the servicer must sign per window
	DowntimeJailDuration    time.Duration `json:"downtime_jail_duration" yaml:"downtime_jail_duration"`         // minimum amount of time servicer must spend in jail after missing blocks
	SlashFractionDoubleSign sdk.BigDec    `json:"slash_fraction_double_sign" yaml:"slash_fraction_double_sign"` // the factor of which a servicer is slashed for a double sign
	SlashFractionDowntime   sdk.BigDec    `json:"slash_fraction_downtime" yaml:"slash_fraction_downtime"`       // the factor of which a servicer is slashed for missing blocks
	ServicerCountLock       bool          `json:"servicer_count_lock" yaml:"servicer_count_lock"`
	BurnActive              bool          `json:"burn_active" yaml:"burn_active"`
	MinPauseTime            time.Duration `json:"min_pause_time" yaml:"min_pause_time"`
	MaxFishermen            int64         `json:"max_fishermen"`
	FishermenCount          int64         `json:"fishermen_count"`
	SlashFractionNoActivity sdk.BigDec    `json:"slash_fraction_noactivity" yaml:"slash_fraction_noactivity"`
	LatencyScoreWeight      sdk.BigDec    `json:"latency_score_weight" yaml:"latency_score_weight"`
	AvailabilityScoreWeight sdk.BigDec    `json:"availability_score_weight" yaml:"availability_score_weight"`
	ReliabilityScoreWeight  sdk.BigDec    `json:"reliability_score_weight" yaml:"reliability_score_weight"`
	SlashFractionFisherman  sdk.BigDec    `json:"slash_fraction_fisherman" yaml:"slash_fraction_fisherman"`
}

// Implements sdk.ParamSet
func (p *Params) ParamSetPairs() sdk.ParamSetPairs {
	return sdk.ParamSetPairs{
		{Key: KeyUnstakingTime, Value: &p.UnstakingTime},
		{Key: KeyMaxValidators, Value: &p.MaxValidators},
		{Key: KeyStakeDenom, Value: &p.StakeDenom},
		{Key: KeyStakeMinimum, Value: &p.StakeMinimum},
		{Key: KeyMaxEvidenceAge, Value: &p.MaxEvidenceAge},
		{Key: KeySignedBlocksWindow, Value: &p.SignedBlocksWindow},
		{Key: KeyMinSignedPerWindow, Value: &p.MinSignedPerWindow},
		{Key: KeyDowntimeJailDuration, Value: &p.DowntimeJailDuration},
		{Key: KeySlashFractionDoubleSign, Value: &p.SlashFractionDoubleSign},
		{Key: KeySlashFractionDowntime, Value: &p.SlashFractionDowntime},
		{Key: KeySessionBlock, Value: &p.SessionBlockFrequency},
		{Key: KeyDAOAllocation, Value: &p.DAOAllocation},
		{Key: KeyProviderAllocation, Value: &p.ProviderAllocation},
		{Key: KeyProposerAllocation, Value: &p.ProposerAllocation},
		{Key: KeyFishermenAllocation, Value: &p.FishermenAllocation},
		{Key: KeyTokenRewardFactor, Value: &p.TokenRewardFactor},
		{Key: KeyMaxChains, Value: &p.MaximumChains},
		{Key: KeyMaxJailedBlocks, Value: &p.MaxJailedBlocks},
		{Key: ServicerCountLock, Value: &p.ServicerCountLock},
		{Key: BurnActive, Value: &p.BurnActive},
		{Key: KeyMinPauseTime, Value: &p.MinPauseTime},
		{Key: KeyMaxFishermen, Value: p.MaxFishermen},
		{Key: KeyFishermenCount, Value: p.FishermenCount},
		{Key: KeySlashFractionNoActivity, Value: &p.SlashFractionNoActivity},
		{Key: KeyLatencyScoreWeight, Value: &p.LatencyScoreWeight},
		{Key: KeyAvailabilityScoreWeight, Value: &p.AvailabilityScoreWeight},
		{Key: KeyReliabilityScoreWeight, Value: &p.ReliabilityScoreWeight},
		{Key: KeySlashFractionFisherman, Value: &p.SlashFractionFisherman},
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return Params{
		UnstakingTime:           DefaultUnstakingTime,
		MaxValidators:           DefaultMaxValidators,
		StakeMinimum:            DefaultMinStake,
		StakeDenom:              sdk.DefaultStakeDenom,
		MaxEvidenceAge:          DefaultMaxEvidenceAge,
		SignedBlocksWindow:      DefaultSignedBlocksWindow,
		MinSignedPerWindow:      DefaultMinSignedPerWindow,
		DowntimeJailDuration:    DefaultDowntimeJailDuration,
		SlashFractionDoubleSign: DefaultSlashFractionDoubleSign,
		SlashFractionDowntime:   DefaultSlashFractionDowntime,
		SessionBlockFrequency:   DefaultSessionBlocktime,
		DAOAllocation:           DefaultDAOAllocation,
		ProviderAllocation:      DefaultProviderAllocation,
		ProposerAllocation:      DefaultProposerAllocation,
		FishermenAllocation:     DefaultFishermenAllocation,
		TokenRewardFactor:       DefaultTokenRewardFactor,
		MaximumChains:           DefaultMaxChains,
		MaxJailedBlocks:         DefaultMaxJailedBlocks,
		ServicerCountLock:       DefaultServicerCountLock,
		MinPauseTime:            DefaultMinPauseTime,
		MaxFishermen:            DefaultMaxFishermen,
		FishermenCount:          DefaultFishermenCount,
		SlashFractionNoActivity: DefaultSlashFractionNoActivity,
		LatencyScoreWeight:      DefaultLatencyScoreWeight,
		AvailabilityScoreWeight: DefaultAvailabilityScoreWeight,
		ReliabilityScoreWeight:  DefaultReliabilityScoreWeight,
		SlashFractionFisherman:  DefaultSlashFractionFisherman,
	}
}

// validate a set of params
func (p Params) Validate() error {
	if p.StakeDenom == "" {
		return fmt.Errorf("staking parameter StakeDenom can't be an empty string")
	}
	if p.MaxValidators == 0 {
		return fmt.Errorf("staking parameter MaxValidators must be a positive integer")
	}
	if p.StakeMinimum < DefaultMinStake {
		return fmt.Errorf("staking parameter StakeMimimum must be greater the Minimum Value")
	}
	if p.SessionBlockFrequency < 2 {
		return fmt.Errorf("session block must be greater than 1")
	}
	if p.DAOAllocation < 0 {
		return fmt.Errorf("the dao allocation must not be negative")
	}
	if p.ProviderAllocation < 0 {
		return fmt.Errorf("the provider allocation must not be negative")
	}
	if p.ProposerAllocation < 0 {
		return fmt.Errorf("the proposer allication must not be negative")
	}
	if p.FishermenAllocation < 0 {
		return fmt.Errorf("the proposer allication must not be negative")
	}
	if p.ProposerAllocation+p.DAOAllocation+p.ProviderAllocation+p.FishermenAllocation > 100 {
		return fmt.Errorf("the combo of proposer allocation, dao allocation and provider allocation must not be greater than 100")
	}
	if p.MaxFishermen < 1 {
		return fmt.Errorf("max fishermen must be equal to or greater than 1")
	}
	if p.FishermenCount < 1 {
		return fmt.Errorf("fishermen count must be equal to or greater than 1")
	}
	if p.LatencyScoreWeight.RoundInt64()+p.AvailabilityScoreWeight.RoundInt64()+p.ReliabilityScoreWeight.RoundInt64() > 1 {
		return fmt.Errorf("the combo of latency score weight, availability score weight and reliability score weight must not be greater than 1")
	}
	return nil
}

// Checks the equality of two param objects
func (p Params) Equal(p2 Params) bool {
	return reflect.DeepEqual(p, p2)
}

// String returns a human readable string representation of the parameters.
func (p Params) String() string {
	return fmt.Sprintf(`Params:
  Unstaking Time:          %s
  Max Validators:          %d
  Stake Coin Denom:        %s
  Minimum Stake:     	   %d
  MaxEvidenceAge:          %s
  SignedBlocksWindow:      %d
  MinSignedPerWindow:      %s
  DowntimeJailDuration:    %s
  SlashFractionDoubleSign: %s
  SlashFractionDowntime:   %s
  BlocksPerSession         %d
  Proposer Allocation      %d
  DAO allocation           %d
  Provider allocation      %d
  Fisherman allocation     %d
  Maximum Chains           %d
  Max Jailed Blocks        %d
  Servicer Count Lock      %v 
  Burn Active              %v 
  MinPauseTime             %s
  Max Fishermen            %d
  Fishermen Count          %d
  SlashFractionNoActivity: %s
  LatencyScoreWeight       %s
  AvailabilityScoreWeight  %s
  ReliabilityScoreWeight   %s
  SlashFractionFisherman   %s`,
		p.UnstakingTime,
		p.MaxValidators,
		p.StakeDenom,
		p.StakeMinimum,
		p.MaxEvidenceAge,
		p.SignedBlocksWindow,
		p.MinSignedPerWindow,
		p.DowntimeJailDuration,
		p.SlashFractionDoubleSign,
		p.SlashFractionDowntime,
		p.SessionBlockFrequency,
		p.ProposerAllocation,
		p.DAOAllocation,
		p.ProviderAllocation,
		p.FishermenAllocation,
		p.MaximumChains,
		p.MaxJailedBlocks,
		p.ServicerCountLock,
		p.BurnActive,
		p.MinPauseTime,
		p.MaxFishermen,
		p.FishermenCount,
		p.SlashFractionNoActivity,
		p.LatencyScoreWeight,
		p.AvailabilityScoreWeight,
		p.ReliabilityScoreWeight,
		p.SlashFractionFisherman)
}
