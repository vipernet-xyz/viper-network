package keys

import (
	"github.com/vipernet-xyz/viper-network/crypto"

	cryptoAmino "github.com/tendermint/tendermint/crypto/encoding/amino"

	"github.com/vipernet-xyz/viper-network/codec"
)

var cdc *codec.LegacyAmino

func init() {
	cdc = codec.NewLegacyAminoCodec()
	cryptoAmino.RegisterAmino(cdc.Amino)
	crypto.RegisterAmino(cdc.Amino)
	cdc.RegisterConcrete(KeyPair{}, "crypto/keys/keypair", nil)
	cdc.Seal()
}
