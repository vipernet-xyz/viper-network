package types

import (
	"github.com/vipernet-xyz/viper-network/codec"
)

// // Register the sdk message type
//
//	func RegisterCodec(cdc *codec.Codec) {
//		amino.RegisterInterface((*ProtoMsg)(nil), nil)
//		amino.RegisterInterface((*Tx)(nil), nil)
//		proto.Register("types/msg", (*ProtoMsg)(nil))
//		proto.Register("types/tx", (*Tx)(nil))
//	}
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface("types/protoMsg", (*ProtoMsg)(nil))
	cdc.RegisterInterface("types/msg", (*Msg)(nil))
	cdc.RegisterInterface("types/tx", (*Tx)(nil))
}

// RegisterLegacyAminoCodec registers the sdk message type.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterInterface((*Msg)(nil), nil)
	cdc.RegisterInterface((*Tx)(nil), nil)
}
