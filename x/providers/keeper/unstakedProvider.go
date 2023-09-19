package keeper

import (
	"bytes"
	"fmt"
	"time"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/providers/types"
)

// SetUnstakingProvider - Store an provider address to the appropriate position in the unstaking queue
func (k Keeper) SetUnstakingProvider(ctx sdk.Ctx, val types.Provider) {
	providers := k.getUnstakingProviders(ctx, val.UnstakingCompletionTime)
	providers = append(providers, val.Address)
	k.setUnstakingProviders(ctx, val.UnstakingCompletionTime, providers)
}

// deleteUnstakingProvider - DeleteEvidence an provider address from the unstaking queue
func (k Keeper) deleteUnstakingProvider(ctx sdk.Ctx, val types.Provider) {
	providers := k.getUnstakingProviders(ctx, val.UnstakingCompletionTime)
	var newProviders []sdk.Address
	for _, addr := range providers {
		if !bytes.Equal(addr, val.Address) {
			newProviders = append(newProviders, addr)
		}
	}
	if len(newProviders) == 0 {
		k.deleteUnstakingProviders(ctx, val.UnstakingCompletionTime)
	} else {
		k.setUnstakingProviders(ctx, val.UnstakingCompletionTime, newProviders)
	}
}

// getAllUnstakingProviders - Retrieve the set of all unstaking providers with no limits
func (k Keeper) getAllUnstakingProviders(ctx sdk.Ctx) (providers []types.Provider) {
	providers = make(types.Providers, 0)
	store := ctx.KVStore(k.storeKey)
	iterator, _ := sdk.KVStorePrefixIterator(store, types.UnstakingProvidersKey)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var addrs sdk.Addresses
		err := k.Cdc.UnmarshalBinaryLengthPrefixed(iterator.Value(), &addrs, ctx.BlockHeight())
		if err != nil {
			k.Logger(ctx).Error(fmt.Errorf("could not unmarshal unstakingProviders in getAllUnstakingProviders call: %s", string(iterator.Value())).Error())
			return
		}
		for _, addr := range addrs {
			provider, found := k.GetProvider(ctx, addr)
			if !found {
				k.Logger(ctx).Error(fmt.Errorf("provider %s in unstakingSet but not found in all providers store", provider.Address).Error())
				continue
			}
			providers = append(providers, provider)
		}

	}
	return providers
}

// getUnstakingProviders - Retrieve all of the providers who will be unstaked at exactly this time
func (k Keeper) getUnstakingProviders(ctx sdk.Ctx, unstakingTime time.Time) (valAddrs sdk.Addresses) {
	store := ctx.KVStore(k.storeKey)
	bz, _ := store.Get(types.KeyForUnstakingProviders(unstakingTime))
	if bz == nil {
		return []sdk.Address{}
	}
	err := k.Cdc.UnmarshalBinaryLengthPrefixed(bz, &valAddrs, ctx.BlockHeight())
	if err != nil {
		panic(err)
	}
	return valAddrs

}

// setUnstakingProviders - Store providers in unstaking queue at a certain unstaking time
func (k Keeper) setUnstakingProviders(ctx sdk.Ctx, unstakingTime time.Time, keys sdk.Addresses) {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.Cdc.MarshalBinaryLengthPrefixed(&keys, ctx.BlockHeight())
	if err != nil {
		panic(err)
	}
	_ = store.Set(types.KeyForUnstakingProviders(unstakingTime), bz)
}

// delteUnstakingProviders - Remove all the providers for a specific unstaking time
func (k Keeper) deleteUnstakingProviders(ctx sdk.Ctx, unstakingTime time.Time) {
	store := ctx.KVStore(k.storeKey)
	_ = store.Delete(types.KeyForUnstakingProviders(unstakingTime))
}

// unstakingProvidersIterator - Retrieve an iterator for all unstaking providers up to a certain time
func (k Keeper) unstakingProvidersIterator(ctx sdk.Ctx, endTime time.Time) (sdk.Iterator, error) {
	store := ctx.KVStore(k.storeKey)
	return store.Iterator(types.UnstakingProvidersKey, sdk.InclusiveEndBytes(types.KeyForUnstakingProviders(endTime)))
}

// getMatureProviders - Retrieve a list of all the mature validators
func (k Keeper) getMatureProviders(ctx sdk.Ctx) (matureValsAddrs sdk.Addresses) {
	matureValsAddrs = make([]sdk.Address, 0)
	unstakingValsIterator, _ := k.unstakingProvidersIterator(ctx, ctx.BlockHeader().Time)
	defer unstakingValsIterator.Close()
	for ; unstakingValsIterator.Valid(); unstakingValsIterator.Next() {
		var providers sdk.Addresses
		err := k.Cdc.UnmarshalBinaryLengthPrefixed(unstakingValsIterator.Value(), &providers, ctx.BlockHeight())
		if err != nil {
			panic(err)
		}
		matureValsAddrs = append(matureValsAddrs, providers...)

	}
	return matureValsAddrs
}

// unstakeAllMatureValidators - Unstake all the unstaking providers that have finished their unstaking period
func (k Keeper) unstakeAllMatureProviders(ctx sdk.Ctx) {
	store := ctx.KVStore(k.storeKey)
	unstakingProvidersIterator, _ := k.unstakingProvidersIterator(ctx, ctx.BlockHeader().Time)
	defer unstakingProvidersIterator.Close()
	for ; unstakingProvidersIterator.Valid(); unstakingProvidersIterator.Next() {
		var unstakingVals sdk.Addresses
		err := k.Cdc.UnmarshalBinaryLengthPrefixed(unstakingProvidersIterator.Value(), &unstakingVals, ctx.BlockHeight())
		if err != nil {
			panic(err)
		}
		for _, valAddr := range unstakingVals {
			val, found := k.GetProvider(ctx, valAddr)
			if !found {
				k.Logger(ctx).Error(fmt.Errorf("provider %s, in the unstaking queue was not found", valAddr).Error())
				continue
			}
			err := k.ValidateProviderFinishUnstaking(ctx, val)
			if err != nil {
				ctx.Logger().Error(fmt.Sprintf("Could not finish unstaking mature provider at height %d: ", ctx.BlockHeight()) + err.Error())
				continue
			}
			k.FinishUnstakingProvider(ctx, val)
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeCompleteUnstaking,
					sdk.NewAttribute(types.AttributeKeyProvider, valAddr.String()),
				),
			)
			k.DeleteProvider(ctx, valAddr)

		}
		_ = store.Delete(unstakingProvidersIterator.Key())
	}
}
