package types

import (
	"github.com/vipernet-xyz/viper-network/codec"
	"github.com/vipernet-xyz/viper-network/codec/types"
	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/servicers/exported"
)

// Register concrete types on codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterStructure(LegacyMsgProtoStake{}, "pos/MsgProtoStake")
	cdc.RegisterStructure(LegacyMsgBeginUnstake{}, "pos/MsgBeginUnstake")
	cdc.RegisterStructure(LegacyMsgUnjail{}, "pos/MsgUnjail")
	cdc.RegisterStructure(MsgSend{}, "pos/Send")
	cdc.RegisterStructure(LegacyMsgStake{}, "pos/MsgStake")
	cdc.RegisterStructure(MsgUnjail{}, "pos/MsgUnjail")
	cdc.RegisterStructure(MsgBeginUnstake{}, "pos/MsgBeginUnstake")
	cdc.RegisterStructure(MsgProtoStake{}, "pos/MsgProtoStake")
	cdc.RegisterStructure(MsgStake{}, "pos/MsgStake")
	cdc.RegisterStructure(MsgPause{}, "pos/MsgPause")
	cdc.RegisterImplementation((*sdk.ProtoMsg)(nil), &MsgUnjail{}, &MsgBeginUnstake{}, &MsgSend{}, &MsgStake{},
		&LegacyMsgUnjail{}, &LegacyMsgBeginUnstake{}, &LegacyMsgStake{}, &MsgPause{}, &MsgUnpause{})
	cdc.RegisterImplementation((*sdk.Msg)(nil), &MsgUnjail{}, &MsgBeginUnstake{}, &MsgSend{}, &MsgStake{},
		&LegacyMsgUnjail{}, &LegacyMsgBeginUnstake{}, &LegacyMsgStake{}, &MsgPause{}, &MsgUnpause{})
	cdc.RegisterInterface("servicers/validatorI", (*exported.ValidatorI)(nil), &Validator{}, &LegacyValidator{})
	ModuleCdc = cdc
}

var ModuleCdc *codec.Codec // generic sealed codec to be used throughout this module

func init() {
	ModuleCdc = codec.NewCodec(types.NewInterfaceRegistry())
	RegisterCodec(ModuleCdc)
	crypto.RegisterAmino(ModuleCdc.AminoCodec().Amino)
	ModuleCdc.AminoCodec().Seal()
}
