package keeper

import (
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/providers/exported"
	"github.com/vipernet-xyz/viper-network/x/providers/types"
)

// Provider - wrproviderer for GetProvider call
func (k Keeper) Provider(ctx sdk.Ctx, address sdk.Address) exported.ProviderI {
	provider, found := k.GetProvider(ctx, address)
	if !found {
		return nil
	}
	return provider
}

// AllProviders - Retrieve a list of all providers
func (k Keeper) AllProviders(ctx sdk.Ctx) (providers []exported.ProviderI) {
	store := ctx.KVStore(k.storeKey)
	iterator, _ := sdk.KVStorePrefixIterator(store, types.AllProvidersKey)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		provider, err := types.UnmarshalProvider(k.Cdc, ctx, iterator.Value())
		if err != nil {
			k.Logger(ctx).Error("couldn't unmarshal provider in AllProviders call: " + string(iterator.Value()) + "\n" + err.Error())
			continue
		}
		providers = append(providers, provider)
	}
	return providers
}
