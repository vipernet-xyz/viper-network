package rootmulti

import (
	"github.com/vipernet-xyz/viper-network/codec"
	"github.com/vipernet-xyz/viper-network/codec/types"
)

var cdc = codec.NewCodec(types.NewInterfaceRegistry())
