package ledger

import (
	"github.com/vipernet-xyz/viper-network/codec"
)

var cdc = codec.NewLegacyAminoCodec()

func init() {
	RegisterAmino(cdc)
}

// RegisterAmino registers all go-crypto related types in the given (amino) codec.
func RegisterAmino(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(PrivKeyLedgerSecp256k1{},
		"tendermint/PrivKeyLedgerSecp256k1", nil)
}
