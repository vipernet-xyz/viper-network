package types

import (
	"github.com/vipernet-xyz/viper-network/codec"
	"github.com/vipernet-xyz/viper-network/codec/types"
	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
)

// RegisterCodec registers concrete types on the codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterStructure(MsgProtoStake{}, "requestors/MsgProtoStake")
	cdc.RegisterStructure(MsgStake{}, "requestors/MsgRequestorStake")
	cdc.RegisterStructure(MsgBeginUnstake{}, "requestors/MsgRequestorBeginUnstake")
	cdc.RegisterStructure(MsgUnjail{}, "requestors/MsgRequestorUnjail")
	cdc.RegisterImplementation((*sdk.ProtoMsg)(nil), &MsgStake{}, &MsgBeginUnstake{}, &MsgUnjail{})
	cdc.RegisterImplementation((*sdk.Msg)(nil), &MsgStake{}, &MsgBeginUnstake{}, &MsgUnjail{})
	ModuleCdc = cdc
}

// module wide codec
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.NewCodec(types.NewInterfaceRegistry())
	RegisterCodec(ModuleCdc)
	crypto.RegisterAmino(ModuleCdc.AminoCodec().Amino)
}
