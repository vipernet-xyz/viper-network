package keeper

import (
	"fmt"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/platforms/exported"
	"github.com/vipernet-xyz/viper-network/x/platforms/types"
)

// SetStakedPlatform - Store staked platform
func (k Keeper) SetStakedPlatform(ctx sdk.Ctx, platform types.Platform) {
	if platform.Jailed {
		return // jailed platforms are not kept in the staking set
	}
	store := ctx.KVStore(k.storeKey)
	_ = store.Set(types.KeyForPlatformInStakingSet(platform), platform.Address)
	ctx.Logger().Info("Setting Platform on Staking Set " + platform.Address.String())
}

// StakeDenom - Retrieve the denomination of coins.
func (k Keeper) StakeDenom(ctx sdk.Ctx) string {
	return k.POSKeeper.StakeDenom(ctx)
}

// deletePlatformFromStakingSet - Remove platform from staked set
func (k Keeper) deletePlatformFromStakingSet(ctx sdk.Ctx, platform types.Platform) {
	store := ctx.KVStore(k.storeKey)
	_ = store.Delete(types.KeyForPlatformInStakingSet(platform))
	ctx.Logger().Info("Removing Platform From Staking Set " + platform.Address.String())
}

// removePlatformTokens - Update the staked tokens of an existing platform, update the platforms power index key
func (k Keeper) removePlatformTokens(ctx sdk.Ctx, platform types.Platform, tokensToRemove sdk.BigInt) (types.Platform, error) {
	ctx.Logger().Info("Removing Platform Tokens, tokensToRemove: " + tokensToRemove.String() + " Platform Address: " + platform.Address.String())
	k.deletePlatformFromStakingSet(ctx, platform)
	platform, err := platform.RemoveStakedTokens(tokensToRemove)
	if err != nil {
		return types.Platform{}, err
	}
	k.SetPlatform(ctx, platform)
	return platform, nil
}

// getStakedPlatforms - Retrieve the current staked platforms sorted by power-rank
func (k Keeper) getStakedPlatforms(ctx sdk.Ctx) types.Platforms {
	var platforms = make(types.Platforms, 0)
	iterator, _ := k.stakedPlatformsIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		address := iterator.Value()
		platform, found := k.GetPlatform(ctx, address)
		if !found {
			k.Logger(ctx).Error(fmt.Errorf("platform %s in staking set but not found in all platforms store", address).Error())
			continue
		}
		if platform.IsStaked() {
			platforms = append(platforms, platform)
		}
	}
	return platforms
}

// getStakedPlatformsCount returns a count of the total staked platformlcations currently
func (k Keeper) getStakedPlatformsCount(ctx sdk.Ctx) (count int64) {
	iterator, _ := k.stakedPlatformsIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		count++
	}
	return
}

// stakedPlatformsIterator - Retrieve an iterator for the current staked platforms
func (k Keeper) stakedPlatformsIterator(ctx sdk.Ctx) (sdk.Iterator, error) {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStoreReversePrefixIterator(store, types.StakedPlatformsKey)
}

// IterateAndExecuteOverStakedPlatforms - Goes through the staked platform set and execute handler
func (k Keeper) IterateAndExecuteOverStakedPlatforms(
	ctx sdk.Ctx, fn func(index int64, platform exported.PlatformI) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator, _ := sdk.KVStoreReversePrefixIterator(store, types.StakedPlatformsKey)
	defer iterator.Close()
	i := int64(0)
	for ; iterator.Valid(); iterator.Next() {
		address := iterator.Value()
		platform, found := k.GetPlatform(ctx, address)
		if !found {
			k.Logger(ctx).Error(fmt.Errorf("platform %s in staking set but not found in all platforms store", address).Error())
			continue
		}
		if platform.IsStaked() {
			stop := fn(i, platform) // XXX is this safe will the platform unexposed fields be able to get written to?
			if stop {
				break
			}
			i++
		}
	}
}
