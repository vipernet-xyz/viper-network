package types

import (
	"github.com/vipernet-xyz/viper-network/codec"
)

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface("types/protoMsg", (*ProtoMsg)(nil))
	cdc.RegisterInterface("types/msg", (*Msg)(nil))
	cdc.RegisterInterface("types/tx", (*Tx)(nil))
}
