package types

import (
	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/requestors/exported"
)

// "GetRequestorFromPublicKey" - Retrieves an requestor from the requestor store, using the requestorKeeper (a link to the requestors module)
// using a hex string public key
func GetRequestorFromPublicKey(ctx sdk.Ctx, requestorsKeeper RequestorsKeeper, pubKey string) (requestor exported.RequestorI, found bool) {
	pk, err := crypto.NewPublicKey(pubKey)
	if err != nil {
		return nil, false
	}
	return GetRequestor(ctx, requestorsKeeper, pk.Address().Bytes())
}

// "GetRequestor" - Retrieves an requestor from the requestor store, using the requestorKeeper (a link to the requestors module)
func GetRequestor(ctx sdk.Ctx, requestorsKeeper RequestorsKeeper, address sdk.Address) (a exported.RequestorI, found bool) {
	a = requestorsKeeper.Requestor(ctx, address)
	if a == nil {
		return a, false
	}
	return a, true
}
