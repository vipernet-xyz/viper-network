package keeper

import (
	"encoding/hex"

	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	vc "github.com/vipernet-xyz/viper-network/x/viper-main/types"
)

// "AATGeneration" - Generates an requestor authentication token with an requestor public key hex string
// a client public key hex string, a passphrase and a keybase. The contract is that the keybase contains the requestor pub key
// and the passphrase corresponds to the requestor public key keypair.
func AATGeneration(requestorPubKey string, clientPubKey string, key crypto.PrivateKey) (vc.AAT, sdk.Error) {
	// create the aat object
	aat := vc.AAT{
		Version:            vc.SupportedTokenVersions[0],
		RequestorPublicKey: requestorPubKey,
		ClientPublicKey:    clientPubKey,
		RequestorSignature: "",
	}
	// marshal aat using json
	sig, err := key.Sign(aat.Hash())
	if err != nil {
		return vc.AAT{}, vc.NewSignatureError(vc.ModuleName, err)
	}
	// stringify the signature into hex
	aat.RequestorSignature = hex.EncodeToString(sig)
	return aat, nil
}
