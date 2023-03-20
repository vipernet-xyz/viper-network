package multisig

import (
	"github.com/cometbft/cometbft/crypto/sr25519"

	"github.com/vipernet-xyz/viper-network/codec"
	"github.com/vipernet-xyz/viper-network/crypto/keys/ed25519"
	"github.com/vipernet-xyz/viper-network/crypto/keys/secp256k1"
	cryptotypes "github.com/vipernet-xyz/viper-network/crypto/types"
)

// TODO: Figure out API for others to either add their own pubkey types, or
// to make verify / marshal accept a AminoCdc.
const (
	PubKeyAminoRoute = "tendermint/PubKeyMultisigThreshold"
)

// AminoCdc is being deprecated in the SDK. But even if you need to
// use Amino for some reason, please use `codec/legacy.Cdc` instead.
var AminoCdc = codec.NewLegacyAmino()

func init() {
	AminoCdc.RegisterInterface((*cryptotypes.PubKey)(nil), nil)
	AminoCdc.RegisterConcrete(ed25519.PubKey{},
		ed25519.PubKeyName, nil)
	AminoCdc.RegisterConcrete(sr25519.PubKey{},
		sr25519.PubKeyName, nil)
	AminoCdc.RegisterConcrete(&secp256k1.PubKey{},
		secp256k1.PubKeyName, nil)
	AminoCdc.RegisterConcrete(&LegacyAminoPubKey{},
		PubKeyAminoRoute, nil)
}
