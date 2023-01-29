package keeper

import (
	"fmt"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/providers/exported"
	"github.com/vipernet-xyz/viper-network/x/providers/types"
)

// SetStakedProvider - Store staked provider
func (k Keeper) SetStakedProvider(ctx sdk.Ctx, provider types.Provider) {
	if provider.Jailed {
		return // jailed providers are not kept in the staking set
	}
	store := ctx.KVStore(k.storeKey)
	_ = store.Set(types.KeyForProviderInStakingSet(provider), provider.Address)
	ctx.Logger().Info("Setting Provider on Staking Set " + provider.Address.String())
}

// StakeDenom - Retrieve the denomination of coins.
func (k Keeper) StakeDenom(ctx sdk.Ctx) string {
	return k.POSKeeper.StakeDenom(ctx)
}

// deleteProviderFromStakingSet - Remove provider from staked set
func (k Keeper) deleteProviderFromStakingSet(ctx sdk.Ctx, provider types.Provider) {
	store := ctx.KVStore(k.storeKey)
	_ = store.Delete(types.KeyForProviderInStakingSet(provider))
	ctx.Logger().Info("Removing Provider From Staking Set " + provider.Address.String())
}

// removeProviderTokens - Update the staked tokens of an existing provider, update the providers power index key
func (k Keeper) removeProviderTokens(ctx sdk.Ctx, provider types.Provider, tokensToRemove sdk.BigInt) (types.Provider, error) {
	ctx.Logger().Info("Removing Provider Tokens, tokensToRemove: " + tokensToRemove.String() + " Provider Address: " + provider.Address.String())
	k.deleteProviderFromStakingSet(ctx, provider)
	provider, err := provider.RemoveStakedTokens(tokensToRemove)
	if err != nil {
		return types.Provider{}, err
	}
	k.SetProvider(ctx, provider)
	return provider, nil
}

// getStakedProviders - Retrieve the current staked providers sorted by power-rank
func (k Keeper) getStakedProviders(ctx sdk.Ctx) types.Providers {
	var providers = make(types.Providers, 0)
	iterator, _ := k.stakedProvidersIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		address := iterator.Value()
		provider, found := k.GetProvider(ctx, address)
		if !found {
			k.Logger(ctx).Error(fmt.Errorf("provider %s in staking set but not found in all providers store", address).Error())
			continue
		}
		if provider.IsStaked() {
			providers = append(providers, provider)
		}
	}
	return providers
}

// getStakedProvidersCount returns a count of the total staked providerlcations currently
func (k Keeper) getStakedProvidersCount(ctx sdk.Ctx) (count int64) {
	iterator, _ := k.stakedProvidersIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		count++
	}
	return
}

// stakedProvidersIterator - Retrieve an iterator for the current staked providers
func (k Keeper) stakedProvidersIterator(ctx sdk.Ctx) (sdk.Iterator, error) {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStoreReversePrefixIterator(store, types.StakedProvidersKey)
}

// IterateAndExecuteOverStakedProviders - Goes through the staked provider set and execute handler
func (k Keeper) IterateAndExecuteOverStakedProviders(
	ctx sdk.Ctx, fn func(index int64, provider exported.ProviderI) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator, _ := sdk.KVStoreReversePrefixIterator(store, types.StakedProvidersKey)
	defer iterator.Close()
	i := int64(0)
	for ; iterator.Valid(); iterator.Next() {
		address := iterator.Value()
		provider, found := k.GetProvider(ctx, address)
		if !found {
			k.Logger(ctx).Error(fmt.Errorf("provider %s in staking set but not found in all providers store", address).Error())
			continue
		}
		if provider.IsStaked() {
			stop := fn(i, provider) // XXX is this safe will the provider unexposed fields be able to get written to?
			if stop {
				break
			}
			i++
		}
	}
}
