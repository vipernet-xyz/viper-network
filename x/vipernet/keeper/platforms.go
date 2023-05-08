package keeper

import (
	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/providers/exported"
)

// "GetProvider" - Retrieves an provider from the provider store, using the providerKeeper (a link to the providers module)
func (k Keeper) GetProvider(ctx sdk.Ctx, address sdk.Address) (a exported.ProviderI, found bool) {
	a = k.providerKeeper.Provider(ctx, address)
	if a == nil {
		return a, false
	}
	return a, true
}

// "GetProviderFromPublicKey" - Retrieves an provider from the provider store, using the providerKeeper (a link to the providers module)
// using a hex string public key
func (k Keeper) GetProviderFromPublicKey(ctx sdk.Ctx, pubKey string) (provider exported.ProviderI, found bool) {
	pk, err := crypto.NewPublicKey(pubKey)
	if err != nil {
		return nil, false
	}
	return k.GetProvider(ctx, sdk.Address(pk.Address()))
}
