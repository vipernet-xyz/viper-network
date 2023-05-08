package keeper

import (
	"encoding/hex"

	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	vc "github.com/vipernet-xyz/viper-network/x/vipernet/types"
)

// "AATGeneration" - Generates an provider authentication token with an provider public key hex string
// a client public key hex string, a passphrase and a keybase. The contract is that the keybase contains the provider pub key
// and the passphrase corresponds to the provider public key keypair.
func AATGeneration(providerPubKey string, clientPubKey string, key crypto.PrivateKey) (vc.AAT, sdk.Error) {
	// create the aat object
	aat := vc.AAT{
		Version:           vc.SupportedTokenVersions[0],
		ProviderPublicKey: providerPubKey,
		ClientPublicKey:   clientPubKey,
		ProviderSignature: "",
	}
	// marshal aat using json
	sig, err := key.Sign(aat.Hash())
	if err != nil {
		return vc.AAT{}, vc.NewSignatureError(vc.ModuleName, err)
	}
	// stringify the signature into hex
	aat.ProviderSignature = hex.EncodeToString(sig)
	return aat, nil
}
