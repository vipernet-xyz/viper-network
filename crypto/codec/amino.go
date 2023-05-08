package codec

import (
	"github.com/cometbft/cometbft/crypto/sr25519"
	"github.com/vipernet-xyz/viper-network/codec"
	"github.com/vipernet-xyz/viper-network/crypto/keys/ed25519"
	kmultisig "github.com/vipernet-xyz/viper-network/crypto/keys/multisig"
	"github.com/vipernet-xyz/viper-network/crypto/keys/secp256k1"
	cryptotypes "github.com/vipernet-xyz/viper-network/crypto/types"

	"github.com/tendermint/go-amino"
	cryptoAmino "github.com/tendermint/tendermint/crypto/encoding/amino"
)

var cdc = codec.NewLegacyAminoCodec()

func init() {
	RegisterAmino(cdc.Amino)
	cryptoAmino.RegisterAmino(cdc.Amino)
}

// RegisterAmino registers all go-crypto related types in the given (amino) codec.
func RegisterAmino(cdc *amino.Codec) {
	cdc.RegisterInterface((*PublicKey)(nil), nil)
	cdc.RegisterInterface((*PrivateKey)(nil), nil)
	cdc.RegisterConcrete(Ed25519PublicKey{}, "crypto/ed25519_public_key", nil)
	cdc.RegisterConcrete(Ed25519PrivateKey{}, "crypto/ed25519_private_key", nil)
	cdc.RegisterConcrete(Secp256k1PublicKey{}, "crypto/secp256k1_public_key", nil)
	cdc.RegisterConcrete(Secp256k1PrivateKey{}, "crypto/secp256k1_private_key", nil)
	cdc.RegisterInterface((*MultiSig)(nil), nil)
	cdc.RegisterInterface((*PublicKeyMultiSig)(nil), nil)
	cdc.RegisterConcrete(PublicKeyMultiSignature{}, "crypto/public_key_multi_signature", nil)
	cdc.RegisterConcrete(MultiSignature{}, "crypto/multi_signature", nil)
}

// RegisterCrypto registers all crypto dependency types with the provided Amino
// codec.
func RegisterCrypto(cdc *codec.LegacyAmino) {
	cdc.RegisterInterface((*cryptotypes.PubKey)(nil), nil)
	cdc.RegisterConcrete(sr25519.PubKey{},
		sr25519.PubKeyName, nil)
	cdc.RegisterConcrete(&ed25519.PubKey{},
		ed25519.PubKeyName, nil)
	cdc.RegisterConcrete(&secp256k1.PubKey{},
		secp256k1.PubKeyName, nil)
	cdc.RegisterConcrete(&kmultisig.LegacyAminoPubKey{},
		kmultisig.PubKeyAminoRoute, nil)

	cdc.RegisterInterface((*cryptotypes.PrivKey)(nil), nil)
	cdc.RegisterConcrete(sr25519.PrivKey{},
		sr25519.PrivKeyName, nil)
	cdc.RegisterConcrete(&ed25519.PrivKey{}, //nolint:staticcheck
		ed25519.PrivKeyName, nil)
	cdc.RegisterConcrete(&secp256k1.PrivKey{},
		secp256k1.PrivKeyName, nil)
}
