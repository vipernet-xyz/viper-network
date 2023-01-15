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
	DefaultTokenRewardFactor        int64 = 1000
	DefaultParamspace                     = ModuleName
	DefaultUnstakingTime                  = time.Hour * 24 * 7 * 3
	DefaultMaxValidators            int64 = 1000
	DefaultMinStake                 int64 = 1000000
	DefaultMaxEvidenceAge                 = 60 * 2 * time.Second
	DefaultSignedBlocksWindow             = int64(10)
	DefaultDowntimeJailDuration           = 60 * 10 * time.Second
	DefaultSessionBlocktime               = 4
	DefaultProposerAllocation             = 5
	DefaultDAOAllocation                  = 10
	DefaultAppAllocation                  = 5
	DefaultMaxChains                      = 15
	DefaultMaxJailedBlocks                = 2000
	DefaultMinServicerStakeBinWidth int64 = 15000000000
	DefaultMaxServicerStakeBin      int64 = 60000000000
)

// - Keys for parameter access
var (
	KeyUnstakingTime                = []byte("UnstakingTime")
	KeyMaxValidators                = []byte("MaxValidators")
	KeyStakeDenom                   = []byte("StakeDenom")
	KeyStakeMinimum                 = []byte("StakeMinimum")
	KeyMaxEvidenceAge               = []byte("MaxEvidenceAge")
	KeySignedBlocksWindow           = []byte("SignedBlocksWindow")
	KeyMinSignedPerWindow           = []byte("MinSignedPerWindow")
	KeyDowntimeJailDuration         = []byte("DowntimeJailDuration")
	KeySlashFractionDoubleSign      = []byte("SlashFractionDoubleSign")
	KeySlashFractionDowntime        = []byte("SlashFractionDowntime")
	KeyTokenRewardFactor            = []byte("TokenRewardFactor")
	KeySessionBlock                 = []byte("BlocksPerSession")
	KeyDAOAllocation                = []byte("DAOAllocation")
	KeyAppAllocation                = []byte("AppAllocation")
	KeyProposerAllocation           = []byte("ProposerPercentage")
	KeyMaxChains                    = []byte("MaximumChains")
	KeyMaxJailedBlocks              = []byte("MaxJailedBlocks")
	KeyMinServicerStakeBinWidth     = []byte("MinServicerStakeBinWidth")
	KeyServicerStakeWeight          = []byte("ServicerStakeWeight")
	KeyMaxServicerStakeBin          = []byte("MaxServicerStakeBin")
	KeyServicerStakeBinExponent     = []byte("ServicerStakeBinExponent")
	DefaultServicerStakeWeight      = sdk.NewDec(1)
	DefaultServicerStakeBinExponent = sdk.NewDec(1)
	DoubleSignJailEndTime           = time.Unix(253402300799, 0) // forever
	DefaultMinSignedPerWindow       = sdk.NewDecWithPrec(5, 1)
	DefaultSlashFractionDoubleSign  = sdk.NewDec(1).Quo(sdk.NewDec(20))
	DefaultSlashFractionDowntime    = sdk.NewDec(1).Quo(sdk.NewDec(100))
)

var _ sdk.ParamSet = (*Params)(nil)

// Params defines the high level settings for pos module
type Params struct {
	TokenRewardFactor        int64         `json:"relays_to_tokens_multiplier" yaml:"relays_to_tokens_multiplier"`
	UnstakingTime            time.Duration `json:"unstaking_time" yaml:"unstaking_time"`                   // how much time must pass between the begin_unstaking_tx and the node going to -> unstaked status
	MaxValidators            int64         `json:"max_validators" yaml:"max_validators"`                   // maximum number of validators in the network at any given block
	StakeDenom               string        `json:"stake_denom" yaml:"stake_denom"`                         // the monetary denomination of the coins in the network `uvipr` or `uAtom` or `Wei`
	StakeMinimum             int64         `json:"stake_minimum" yaml:"stake_minimum"`                     // minimum amount of `uvipr` needed to stake in the network as a node
	SessionBlockFrequency    int64         `json:"session_block_frequency" yaml:"session_block_frequency"` // how many blocks are in a session (viper network unit)
	DAOAllocation            int64         `json:"dao_allocation" yaml:"dao_allocation"`
	AppAllocation            int64         `json:"app_allocation" yaml:"app_allocation"`
	ProposerAllocation       int64         `json:"proposer_allocation" yaml:"proposer_allocation"`
	MaximumChains            int64         `json:"maximum_chains" yaml:"maximum_chains"`
	MaxJailedBlocks          int64         `json:"max_jailed_blocks" yaml:"max_jailed_blocks"`
	MaxEvidenceAge           time.Duration `json:"max_evidence_age" yaml:"max_evidence_age"`                     // maximum age of tendermint evidence that is still valid (currently not implemented in Cosmos or Viper-Core)
	SignedBlocksWindow       int64         `json:"signed_blocks_window" yaml:"signed_blocks_window"`             // window of time in blocks (unit) used for signature verification -> specifically in not signing (missing) blocks
	MinSignedPerWindow       sdk.BigDec    `json:"min_signed_per_window" yaml:"min_signed_per_window"`           // minimum number of blocks the node must sign per window
	DowntimeJailDuration     time.Duration `json:"downtime_jail_duration" yaml:"downtime_jail_duration"`         // minimum amount of time node must spend in jail after missing blocks
	SlashFractionDoubleSign  sdk.BigDec    `json:"slash_fraction_double_sign" yaml:"slash_fraction_double_sign"` // the factor of which a node is slashed for a double sign
	SlashFractionDowntime    sdk.BigDec    `json:"slash_fraction_downtime" yaml:"slash_fraction_downtime"`       // the factor of which a node is slashed for missing blocks
	MinServicerStakeBinWidth int64         `json:"servicer_stake_floor_multipler" yaml:"servicer_stake_floor_multipler"`
	ServicerStakeWeight      sdk.BigDec    `json:"servicer_stake_weight_multipler" yaml:"servicer_stake_weight_multipler"`
	MaxServicerStakeBin      int64         `json:"servicer_stake_weight_ceiling" yaml:"servicer_stake_weight_cieling"`
	ServicerStakeBinExponent sdk.BigDec    `json:"servicer_stake_floor_multiplier_exponent" yaml:"servicer_stake_floor_multiplier_exponent"`
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
		{Key: KeyAppAllocation, Value: &p.AppAllocation},
		{Key: KeyProposerAllocation, Value: &p.ProposerAllocation},
		{Key: KeyTokenRewardFactor, Value: &p.TokenRewardFactor},
		{Key: KeyMaxChains, Value: &p.MaximumChains},
		{Key: KeyMaxJailedBlocks, Value: &p.MaxJailedBlocks},
		{Key: KeyMinServicerStakeBinWidth, Value: &p.MinServicerStakeBinWidth},
		{Key: KeyServicerStakeWeight, Value: &p.ServicerStakeWeight},
		{Key: KeyMaxServicerStakeBin, Value: &p.MaxServicerStakeBin},
		{Key: KeyServicerStakeBinExponent, Value: &p.ServicerStakeBinExponent},
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
		AppAllocation:           DefaultAppAllocation,
		ProposerAllocation:      DefaultProposerAllocation,
		TokenRewardFactor:       DefaultTokenRewardFactor,
		MaximumChains:           DefaultMaxChains,
		MaxJailedBlocks:         DefaultMaxJailedBlocks,
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
		return fmt.Errorf("staking parameter StakeMimimum must be a positive integer")
	}
	if p.SessionBlockFrequency < 2 {
		return fmt.Errorf("session block must be greater than 1")
	}
	if p.DAOAllocation < 0 {
		return fmt.Errorf("the dao allocation must not be negative")
	}
	if p.AppAllocation < 0 {
		return fmt.Errorf("the app allocation must not be negative")
	}
	if p.ProposerAllocation < 0 {
		return fmt.Errorf("the proposer allication must not be negative")
	}
	if p.ProposerAllocation+p.DAOAllocation+p.AppAllocation > 100 {
		return fmt.Errorf("the combo of proposer allocation, dao allocation and app allocation must not be greater than 100")
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
  App allocation           %d
  Maximum Chains           %d
  Max Jailed Blocks        %d`,
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
		p.AppAllocation,
		p.MaximumChains,
		p.MaxJailedBlocks)
}
