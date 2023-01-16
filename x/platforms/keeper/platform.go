package keeper

import (
	"math"
	"math/big"
	"os"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/platforms/exported"
	"github.com/vipernet-xyz/viper-network/x/platforms/types"
)

// GetPlatform - Retrieve a single platform from the main store
func (k Keeper) GetPlatform(ctx sdk.Ctx, addr sdk.Address) (platform types.Platform, found bool) {
	user, found := k.PlatformCache.GetWithCtx(ctx, addr.String())
	if found && user != nil {
		return user.(types.Platform), found
	}
	store := ctx.KVStore(k.storeKey)
	value, _ := store.Get(types.KeyForPlatformByAllPlatforms(addr))
	if value == nil {
		return platform, false
	}
	platform, err := types.UnmarshalPlatform(k.Cdc, ctx, value)
	if err != nil {
		k.Logger(ctx).Error("could not unmarshal platform from store")
		return platform, false
	}
	_ = k.PlatformCache.AddWithCtx(ctx, addr.String(), platform)
	return platform, true
}

// SetPlatform - Add a single platform the main store
func (k Keeper) SetPlatform(ctx sdk.Ctx, platform types.Platform) {
	store := ctx.KVStore(k.storeKey)
	bz, err := types.MarshalPlatform(k.Cdc, ctx, platform)
	if err != nil {
		k.Logger(ctx).Error("could not marshal platform object", err.Error())
		os.Exit(1)
	}
	_ = store.Set(types.KeyForPlatformByAllPlatforms(platform.Address), bz)
	ctx.Logger().Info("Setting Platform on Main Store " + platform.Address.String())
	if platform.IsUnstaking() {
		k.SetUnstakingPlatform(ctx, platform)
	}
	if platform.IsStaked() && !platform.IsJailed() {
		k.SetStakedPlatform(ctx, platform)
	}
	_ = k.PlatformCache.AddWithCtx(ctx, platform.Address.String(), platform)
}

func (k Keeper) SetPlatforms(ctx sdk.Ctx, platforms types.Platforms) {
	for _, platform := range platforms {
		k.SetPlatform(ctx, platform)
	}
}

// SetValidator - Store validator in the main store
func (k Keeper) DeletePlatform(ctx sdk.Ctx, addr sdk.Address) {
	store := ctx.KVStore(k.storeKey)
	_ = store.Delete(types.KeyForPlatformByAllPlatforms(addr))
	k.PlatformCache.RemoveWithCtx(ctx, addr.String())
}

// GetAllPlatforms - Retrieve the set of all platforms with no limits from the main store
func (k Keeper) GetAllPlatforms(ctx sdk.Ctx) (platforms types.Platforms) {
	platforms = make([]types.Platform, 0)
	store := ctx.KVStore(k.storeKey)
	iterator, _ := sdk.KVStorePrefixIterator(store, types.AllPlatformsKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		platform, err := types.UnmarshalPlatform(k.Cdc, ctx, iterator.Value())
		if err != nil {
			k.Logger(ctx).Error("couldn't unmarshal platform in GetAllPlatforms call: " + string(iterator.Value()) + "\n" + err.Error())
			continue
		}
		platforms = append(platforms, platform)
	}
	return platforms
}

// GetAllPlatformsWithOpts - Retrieve the set of all platforms with no limits from the main store
func (k Keeper) GetAllPlatformsWithOpts(ctx sdk.Ctx, opts types.QueryPlatformsWithOpts) (platforms types.Platforms) {
	platforms = make([]types.Platform, 0)
	store := ctx.KVStore(k.storeKey)
	iterator, _ := sdk.KVStorePrefixIterator(store, types.AllPlatformsKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		platform, err := types.UnmarshalPlatform(k.Cdc, ctx, iterator.Value())
		if err != nil {
			k.Logger(ctx).Error("couldn't unmarshal platform in GetAllPlatformsWithOpts call: " + string(iterator.Value()) + "\n" + err.Error())
			continue
		}
		if opts.IsValid(platform) {
			platforms = append(platforms, platform)
		}
	}
	return platforms
}

// GetPlatforms - Retrieve a a given amount of all the platforms
func (k Keeper) GetPlatforms(ctx sdk.Ctx, maxRetrieve uint16) (platforms types.Platforms) {
	store := ctx.KVStore(k.storeKey)
	platforms = make([]types.Platform, maxRetrieve)

	iterator, _ := sdk.KVStorePrefixIterator(store, types.AllPlatformsKey)
	defer iterator.Close()

	i := 0
	for ; iterator.Valid() && i < int(maxRetrieve); iterator.Next() {
		platform, err := types.UnmarshalPlatform(k.Cdc, ctx, iterator.Value())
		if err != nil {
			k.Logger(ctx).Error("couldn't unmarshal platform in GetPlatforms call: " + string(iterator.Value()) + "\n" + err.Error())
			continue
		}
		platforms[i] = platform
		i++
	}
	return platforms[:i] // trim if the array length < maxRetrieve
}

// IterateAndExecuteOverPlatforms - Goes through the platform set and perform the provided function
func (k Keeper) IterateAndExecuteOverPlatforms(
	ctx sdk.Ctx, fn func(index int64, platform exported.PlatformI) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator, _ := sdk.KVStorePrefixIterator(store, types.AllPlatformsKey)
	defer iterator.Close()
	i := int64(0)
	for ; iterator.Valid(); iterator.Next() {
		platform, err := types.UnmarshalPlatform(k.Cdc, ctx, iterator.Value())
		if err != nil {
			k.Logger(ctx).Error("couldn't unmarshal platform in IterateAndExecuteOverPlatforms call: " + string(iterator.Value()) + "\n" + err.Error())
			continue
		}
		stop := fn(i, platform) // XXX is this safe will the platform unexposed fields be able to get written to?
		if stop {
			break
		}
		i++
	}
}

func (k Keeper) CalculatePlatformRelays(ctx sdk.Ctx, platform types.Platform) sdk.BigInt {
	stakingAdjustment := sdk.NewDec(k.StakingAdjustment(ctx))
	participationRate := sdk.NewDec(1)
	baseRate := sdk.NewInt(k.BaselineThroughputStakeRate(ctx))
	if k.ParticipationRate(ctx) {
		platformStakedCoins := k.GetStakedTokens(ctx)
		providerStakedCoins := k.POSKeeper.GetStakedTokens(ctx)
		totalTokens := k.TotalTokens(ctx)
		participationRate = platformStakedCoins.Add(providerStakedCoins).ToDec().Quo(totalTokens.ToDec())
	}
	basePercentage := baseRate.ToDec().Quo(sdk.NewDec(100))
	baselineThroughput := basePercentage.Mul(platform.StakedTokens.ToDec().Quo(sdk.NewDec(1000000)))
	result := participationRate.Mul(baselineThroughput).Add(stakingAdjustment).TruncateInt()

	// Max Amount of relays Value
	maxRelays := sdk.NewIntFromBigInt(new(big.Int).SetUint64(math.MaxUint64))
	if result.GTE(maxRelays) {
		result = maxRelays
	}

	return result
}

// RelaysPerStakedVIPR = VIPR price(30 day avg.) / (USD relay target * Sessions/Day * Average days per month * ROI target)
