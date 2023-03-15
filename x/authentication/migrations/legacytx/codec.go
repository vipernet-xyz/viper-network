package legacytx

import (
	"github.com/vipernet-xyz/viper-network/codec"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(StdTx{}, "viper-network/StdTx", nil)
}
