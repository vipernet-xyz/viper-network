// Package auth nolint
// autogenerated code using github.com/rigelrozanski/multitool
// aliases generated for the following subdirectories:
// ALIASGEN: github.com/vipernet-xyz/viper-network/x/auth/types
package auth

import (
	"github.com/vipernet-xyz/viper-network/x/auth/exported"
	"github.com/vipernet-xyz/viper-network/x/auth/keeper"
	"github.com/vipernet-xyz/viper-network/x/auth/types"
)

// Const Constants
const (
	ModuleName        = types.ModuleName
	StoreKey          = types.StoreKey
	FeeCollectorName  = types.FeeCollectorName
	QuerierRoute      = types.QuerierRoute
	DefaultParamspace = types.DefaultCodespace
	QueryAccount      = types.QueryAccount
	Burner            = types.Burner
	Staking           = types.Staking
	Minter            = types.Minter
)

var (
	NewKeeper                 = keeper.NewKeeper
	NewModuleAddress          = types.NewModuleAddress
	NewBaseAccountWithAddress = types.NewBaseAccountWithAddress
	RegisterCodec             = types.RegisterCodec
	CountSubKeys              = types.CountSubKeys
	StdSignBytes              = types.StdSignBytes
	DefaultTxDecoder          = types.DefaultTxDecoder
	DefaultTxEncoder          = types.DefaultTxEncoder
	NewTxBuilder              = types.NewTxBuilder
	ModuleCdc                 = types.ModuleCdc
)

// Type exported types
type (
	GenesisState       = types.GenesisState
	Keeper             = keeper.Keeper
	Account            = exported.Account
	BaseAccount        = types.BaseAccount
	Params             = types.Params
	QueryAccountParams = types.QueryAccountParams
	ProtoStdTx         = types.ProtoStdTx
	StdTx              = types.StdTx
	StdSignDoc         = types.StdSignDoc
	StdSignature       = types.ProtoStdSignature
	TxBuilder          = types.TxBuilder
)
