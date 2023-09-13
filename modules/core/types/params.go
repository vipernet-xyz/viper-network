package types

import (
	fmt "fmt"
	"reflect"

	"github.com/vipernet-xyz/viper-network/types"
)

const (
	DefaultParamspace                     = ModuleName
	DefaultMaxExpectedTimePerBlock uint64 = 30000000000
)

var KeyMaxExpectedTimePerBlock = []byte("MaxExpectedTimePerBlock")

var _ types.ParamSet = (*Params)(nil)

type Params struct {
	MaxExpectedTimePerBlock uint64 `json:"max_expected_time_per_block" yaml:"max_expected_time_per_block"`
}

// Implements params.ParamSet
func (p *Params) ParamSetPairs() types.ParamSetPairs {
	return types.ParamSetPairs{
		{Key: KeyMaxExpectedTimePerBlock, Value: &p.MaxExpectedTimePerBlock},
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return Params{
		MaxExpectedTimePerBlock: DefaultMaxExpectedTimePerBlock,
	}
}

// Validate a set of params
func (p Params) Validate() error {
	if p.MaxExpectedTimePerBlock == 0 {
		return fmt.Errorf("parameter MaxExpectedTimePerBlock must be a positive integer")
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
	MaxExpectedTimePerBlock :               %d
,`,
		p.MaxExpectedTimePerBlock,
	)
}
