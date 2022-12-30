package types

import (
	"github.com/vipernet-xyz/viper-network/codec"
	"github.com/vipernet-xyz/viper-network/codec/types"
	"github.com/vipernet-xyz/viper-network/crypto"
	sdk "github.com/vipernet-xyz/viper-network/types"
)

// RegisterCodec registers concrete types on the codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterStructure(MsgProtoStake{}, "apps/MsgProtoStake")
	cdc.RegisterStructure(MsgStake{}, "apps/MsgAppStake")
	cdc.RegisterStructure(MsgBeginUnstake{}, "apps/MsgAppBeginUnstake")
	cdc.RegisterStructure(MsgUnjail{}, "apps/MsgAppUnjail")
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
