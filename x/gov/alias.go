// nolint
package gov

import (
	"github.com/vipernet-xyz/viper-network/x/gov/types"
)

const (
	StoreKey         = types.StoreKey
	TStoreKey        = types.TStoreKey
	DefaultCodespace = types.DefaultCodespace
	ModuleName       = types.ModuleName
	RouterKey        = types.RouterKey
)

var (
	RegisterCodec = types.RegisterCodec
	// variable aliases
	ModuleCdc = types.ModuleCdc
)
