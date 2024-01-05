package keeper

import (
	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/requestors/exported"
)

// "GetRequestor" - Retrieves an requestor from the requestor store, using the requestorKeeper (a link to the requestors module)
func (k Keeper) GetRequestor(ctx sdk.Ctx, address sdk.Address) (a exported.RequestorI, found bool) {
	a = k.requestorKeeper.Requestor(ctx, address)
	if a == nil {
		return a, false
	}
	return a, true
}

// "GetRequestorFromPublicKey" - Retrieves an requestor from the requestor store, using the requestorKeeper (a link to the requestors module)
// using a hex string public key
func (k Keeper) GetRequestorFromPublicKey(ctx sdk.Ctx, pubKey string) (requestor exported.RequestorI, found bool) {
	pk, err := crypto.NewPublicKey(pubKey)
	if err != nil {
		return nil, false
	}
	return k.GetRequestor(ctx, sdk.Address(pk.Address()))
}
