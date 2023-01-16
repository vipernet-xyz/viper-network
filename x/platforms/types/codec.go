package types

import (
	"github.com/vipernet-xyz/viper-network/codec"
	"github.com/vipernet-xyz/viper-network/codec/types"
	"github.com/vipernet-xyz/viper-network/crypto"
	sdk "github.com/vipernet-xyz/viper-network/types"
)

// RegisterCodec registers concrete types on the codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterStructure(MsgProtoStake{}, "platforms/MsgProtoStake")
	cdc.RegisterStructure(MsgStake{}, "platforms/MsgPlatformStake")
	cdc.RegisterStructure(MsgBeginUnstake{}, "platforms/MsgPlatformBeginUnstake")
	cdc.RegisterStructure(MsgUnjail{}, "platforms/MsgPlatformUnjail")
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
