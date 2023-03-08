package types

import (
	"fmt"

	paramtypes "github.com/vipernet-xyz/viper-network/types"
	//"github.com/vipernet-xyz/viper-network/x/servicers/types"
)

const (
	DefaultParamspace = ModuleName
)
const (
	// DefaultSendEnabled enabled
	DefaultSendEnabled = true
	// DefaultReceiveEnabled enabled
	DefaultReceiveEnabled = true
)

var (
	// KeySendEnabled is store's key for SendEnabled Params
	KeySendEnabled = []byte("SendEnabled")
	// KeyReceiveEnabled is store's key for ReceiveEnabled Params
	KeyReceiveEnabled = []byte("ReceiveEnabled")
	KeyStakeDenom     = []byte("StakeDenom")
)

// ParamKeyTable type declaration for parameters
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new parameter configuration for the ibc transfer module
func NewParams(enableSend, enableReceive bool) Params {
	return Params{
		SendEnabled:    enableSend,
		ReceiveEnabled: enableReceive,
	}
}

// DefaultParams is the default parameter configuration for the ibc-transfer module
func DefaultParams() Params {
	return NewParams(DefaultSendEnabled, DefaultReceiveEnabled)
}

// Validate all ibc-transfer module parameters
func (p Params) Validate() error {
	if err := validateEnabledType(p.SendEnabled); err != nil {
		return err
	}

	return validateEnabledType(p.ReceiveEnabled)
}

// ParamSetPairs implements params.ParamSet
func (p *Params) ParamSetPairs1() paramtypes.ParamSetPairs1 {
	return paramtypes.ParamSetPairs1{
		paramtypes.NewParamSetPair(KeySendEnabled, p.SendEnabled, validateEnabledType),
		paramtypes.NewParamSetPair(KeyReceiveEnabled, p.ReceiveEnabled, validateEnabledType),
	}
}

func validateEnabledType(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}
