package types

import (
	"github.com/vipernet-xyz/viper-network/codec"
	"github.com/vipernet-xyz/viper-network/codec/types"
	codectypes "github.com/vipernet-xyz/viper-network/codec/types"
	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"

	clienttypes "github.com/vipernet-xyz/viper-network/modules/core/02-client/types"
	connectiontypes "github.com/vipernet-xyz/viper-network/modules/core/03-connection/types"
	channeltypes "github.com/vipernet-xyz/viper-network/modules/core/04-channel/types"
	commitmenttypes "github.com/vipernet-xyz/viper-network/modules/core/23-commitment/types"
)

// RegisterInterfaces registers x/ibc interfaces into protobuf Any.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	clienttypes.RegisterInterfaces(registry)
	connectiontypes.RegisterInterfaces(registry)
	channeltypes.RegisterInterfaces(registry)
	commitmenttypes.RegisterInterfaces(registry)
}

// module wide codec
var ModuleCdc *codec.Codec
var amino = codec.NewLegacyAmino()
var AminoCdc = codec.NewAminoCodec(amino)

func init() {
	ModuleCdc = codec.NewCodec(types.NewInterfaceRegistry())
	crypto.RegisterAmino(ModuleCdc.AminoCodec().Amino)
}

// RegisterCodec registers concrete types on the codec
func RegisterCodec(cdc *codec.Codec) {
	ModuleCdc = cdc

}
