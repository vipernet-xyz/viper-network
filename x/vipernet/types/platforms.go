package types

import (
	"github.com/vipernet-xyz/viper-network/crypto"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/providers/exported"
)

// "GetProviderFromPublicKey" - Retrieves an providerlication from the provider store, using the providerKeeper (a link to the providers module)
// using a hex string public key
func GetProviderFromPublicKey(ctx sdk.Ctx, providersKeeper ProvidersKeeper, pubKey string) (provider exported.ProviderI, found bool) {
	pk, err := crypto.NewPublicKey(pubKey)
	if err != nil {
		return nil, false
	}
	return GetProvider(ctx, providersKeeper, pk.Address().Bytes())
}

// "GetProvider" - Retrieves an providerlication from the provider store, using the providerKeeper (a link to the providers module)
func GetProvider(ctx sdk.Ctx, providersKeeper ProvidersKeeper, address sdk.Address) (a exported.ProviderI, found bool) {
	a = providersKeeper.Provider(ctx, address)
	if a == nil {
		return a, false
	}
	return a, true
}
