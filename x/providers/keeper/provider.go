package keeper

import (
	"math"
	"math/big"
	"os"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/providers/exported"
	"github.com/vipernet-xyz/viper-network/x/providers/types"
)

// GetProvider - Retrieve a single provider from the main store
func (k Keeper) GetProvider(ctx sdk.Ctx, addr sdk.Address) (provider types.Provider, found bool) {
	user, found := k.ProviderCache.GetWithCtx(ctx, addr.String())
	if found && user != nil {
		return user.(types.Provider), found
	}
	store := ctx.KVStore(k.storeKey)
	value, _ := store.Get(types.KeyForProviderByAllProviders(addr))
	if value == nil {
		return provider, false
	}
	provider, err := types.UnmarshalProvider(k.Cdc, ctx, value)
	if err != nil {
		k.Logger(ctx).Error("could not unmarshal provider from store")
		return provider, false
	}
	_ = k.ProviderCache.AddWithCtx(ctx, addr.String(), provider)
	return provider, true
}

// SetProvider - Add a single provider the main store
func (k Keeper) SetProvider(ctx sdk.Ctx, provider types.Provider) {
	store := ctx.KVStore(k.storeKey)
	bz, err := types.MarshalProvider(k.Cdc, ctx, provider)
	if err != nil {
		k.Logger(ctx).Error("could not marshal provider object", err.Error())
		os.Exit(1)
	}
	_ = store.Set(types.KeyForProviderByAllProviders(provider.Address), bz)
	ctx.Logger().Info("Setting Provider on Main Store " + provider.Address.String())
	if provider.IsUnstaking() {
		k.SetUnstakingProvider(ctx, provider)
	}
	if provider.IsStaked() && !provider.IsJailed() {
		k.SetStakedProvider(ctx, provider)
	}
	_ = k.ProviderCache.AddWithCtx(ctx, provider.Address.String(), provider)
}

func (k Keeper) SetProviders(ctx sdk.Ctx, providers types.Providers) {
	for _, provider := range providers {
		k.SetProvider(ctx, provider)
	}
}

// SetValidator - Store validator in the main store
func (k Keeper) DeleteProvider(ctx sdk.Ctx, addr sdk.Address) {
	store := ctx.KVStore(k.storeKey)
	_ = store.Delete(types.KeyForProviderByAllProviders(addr))
	k.ProviderCache.RemoveWithCtx(ctx, addr.String())
}

// GetAllProviders - Retrieve the set of all providers with no limits from the main store
func (k Keeper) GetAllProviders(ctx sdk.Ctx) (providers types.Providers) {
	providers = make([]types.Provider, 0)
	store := ctx.KVStore(k.storeKey)
	iterator, _ := sdk.KVStorePrefixIterator(store, types.AllProvidersKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		provider, err := types.UnmarshalProvider(k.Cdc, ctx, iterator.Value())
		if err != nil {
			k.Logger(ctx).Error("couldn't unmarshal provider in GetAllProviders call: " + string(iterator.Value()) + "\n" + err.Error())
			continue
		}
		providers = append(providers, provider)
	}
	return providers
}

// GetAllProvidersWithOpts - Retrieve the set of all providers with no limits from the main store
func (k Keeper) GetAllProvidersWithOpts(ctx sdk.Ctx, opts types.QueryProvidersWithOpts) (providers types.Providers) {
	providers = make([]types.Provider, 0)
	store := ctx.KVStore(k.storeKey)
	iterator, _ := sdk.KVStorePrefixIterator(store, types.AllProvidersKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		provider, err := types.UnmarshalProvider(k.Cdc, ctx, iterator.Value())
		if err != nil {
			k.Logger(ctx).Error("couldn't unmarshal provider in GetAllProvidersWithOpts call: " + string(iterator.Value()) + "\n" + err.Error())
			continue
		}
		if opts.IsValid(provider) {
			providers = append(providers, provider)
		}
	}
	return providers
}

// GetProviders - Retrieve a a given amount of all the providers
func (k Keeper) GetProviders(ctx sdk.Ctx, maxRetrieve uint16) (providers types.Providers) {
	store := ctx.KVStore(k.storeKey)
	providers = make([]types.Provider, maxRetrieve)

	iterator, _ := sdk.KVStorePrefixIterator(store, types.AllProvidersKey)
	defer iterator.Close()

	i := 0
	for ; iterator.Valid() && i < int(maxRetrieve); iterator.Next() {
		provider, err := types.UnmarshalProvider(k.Cdc, ctx, iterator.Value())
		if err != nil {
			k.Logger(ctx).Error("couldn't unmarshal provider in GetProviders call: " + string(iterator.Value()) + "\n" + err.Error())
			continue
		}
		providers[i] = provider
		i++
	}
	return providers[:i] // trim if the array length < maxRetrieve
}

// IterateAndExecuteOverProviders - Goes through the provider set and perform the provided function
func (k Keeper) IterateAndExecuteOverProviders(
	ctx sdk.Ctx, fn func(index int64, provider exported.ProviderI) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator, _ := sdk.KVStorePrefixIterator(store, types.AllProvidersKey)
	defer iterator.Close()
	i := int64(0)
	for ; iterator.Valid(); iterator.Next() {
		provider, err := types.UnmarshalProvider(k.Cdc, ctx, iterator.Value())
		if err != nil {
			k.Logger(ctx).Error("couldn't unmarshal provider in IterateAndExecuteOverProviders call: " + string(iterator.Value()) + "\n" + err.Error())
			continue
		}
		stop := fn(i, provider) // XXX is this safe will the provider unexposed fields be able to get written to?
		if stop {
			break
		}
		i++
	}
}

func (k Keeper) CalculateProviderRelays(ctx sdk.Ctx, provider types.Provider) sdk.BigInt {
	stakingAdjustment := sdk.NewDec(k.StakingAdjustment(ctx))
	participationRate := sdk.NewDec(1)
	baseRate := sdk.NewInt(k.BaselineThroughputStakeRate(ctx))
	if k.ParticipationRate(ctx) {
		providerStakedCoins := k.GetStakedTokens(ctx)
		servicerStakedCoins := k.POSKeeper.GetStakedTokens(ctx)
		totalTokens := k.TotalTokens(ctx)
		participationRate = providerStakedCoins.Add(servicerStakedCoins).ToDec().Quo(totalTokens.ToDec())
	}
	basePercentage := baseRate.ToDec().Quo(sdk.NewDec(100))
	baselineThroughput := basePercentage.Mul(provider.StakedTokens.ToDec().Quo(sdk.NewDec(1000000)))
	result := participationRate.Mul(baselineThroughput).Add(stakingAdjustment).TruncateInt()

	// Max Amount of relays Value
	maxRelays := sdk.NewIntFromBigInt(new(big.Int).SetUint64(math.MaxUint64))
	if result.GTE(maxRelays) {
		result = maxRelays
	}

	return result
}

// RelaysPerStakedVIPR = VIPR price(30 day avg.) / (USD relay target * Sessions/Day * Average days per month * ROI target)
