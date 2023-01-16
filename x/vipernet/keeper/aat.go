package keeper

import (
	"encoding/hex"

	"github.com/vipernet-xyz/viper-network/crypto"
	sdk "github.com/vipernet-xyz/viper-network/types"
	vc "github.com/vipernet-xyz/viper-network/x/vipernet/types"
)

// "AATGeneration" - Generates an platformlication authentication token with an platformlication public key hex string
// a client public key hex string, a passphrase and a keybase. The contract is that the keybase contains the platform pub key
// and the passphrase corresponds to the platform public key keypair.
func AATGeneration(platformPubKey string, clientPubKey string, key crypto.PrivateKey) (vc.AAT, sdk.Error) {
	// create the aat object
	aat := vc.AAT{
		Version:           vc.SupportedTokenVersions[0],
		PlatformPublicKey: platformPubKey,
		ClientPublicKey:   clientPubKey,
		PlatformSignature: "",
	}
	// marshal aat using json
	sig, err := key.Sign(aat.Hash())
	if err != nil {
		return vc.AAT{}, vc.NewSignatureError(vc.ModuleName, err)
	}
	// stringify the signature into hex
	aat.PlatformSignature = hex.EncodeToString(sig)
	return aat, nil
}
