package keeper

import (
	"bytes"
	"fmt"
	"time"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/platforms/types"
)

// SetUnstakingPlatform - Store an platform address to the platformropriate position in the unstaking queue
func (k Keeper) SetUnstakingPlatform(ctx sdk.Ctx, val types.Platform) {
	platforms := k.getUnstakingPlatforms(ctx, val.UnstakingCompletionTime)
	platforms = append(platforms, val.Address)
	k.setUnstakingPlatforms(ctx, val.UnstakingCompletionTime, platforms)
}

// deleteUnstakingPlatformlicaiton - DeleteEvidence an platform address from the unstaking queue
func (k Keeper) deleteUnstakingPlatform(ctx sdk.Ctx, val types.Platform) {
	platforms := k.getUnstakingPlatforms(ctx, val.UnstakingCompletionTime)
	var newPlatforms []sdk.Address
	for _, addr := range platforms {
		if !bytes.Equal(addr, val.Address) {
			newPlatforms = append(newPlatforms, addr)
		}
	}
	if len(newPlatforms) == 0 {
		k.deleteUnstakingPlatforms(ctx, val.UnstakingCompletionTime)
	} else {
		k.setUnstakingPlatforms(ctx, val.UnstakingCompletionTime, newPlatforms)
	}
}

// getAllUnstakingPlatforms - Retrieve the set of all unstaking platforms with no limits
func (k Keeper) getAllUnstakingPlatforms(ctx sdk.Ctx) (platforms []types.Platform) {
	platforms = make(types.Platforms, 0)
	store := ctx.KVStore(k.storeKey)
	iterator, _ := sdk.KVStorePrefixIterator(store, types.UnstakingPlatformsKey)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var addrs sdk.Addresses
		err := k.Cdc.UnmarshalBinaryLengthPrefixed(iterator.Value(), &addrs, ctx.BlockHeight())
		if err != nil {
			k.Logger(ctx).Error(fmt.Errorf("could not unmarshal unstakingPlatforms in getAllUnstakingPlatforms call: %s", string(iterator.Value())).Error())
			return
		}
		for _, addr := range addrs {
			platform, found := k.GetPlatform(ctx, addr)
			if !found {
				k.Logger(ctx).Error(fmt.Errorf("platform %s in unstakingSet but not found in all platforms store", platform.Address).Error())
				continue
			}
			platforms = append(platforms, platform)
		}

	}
	return platforms
}

// getUnstakingPlatforms - Retrieve all of the platforms who will be unstaked at exactly this time
func (k Keeper) getUnstakingPlatforms(ctx sdk.Ctx, unstakingTime time.Time) (valAddrs sdk.Addresses) {
	store := ctx.KVStore(k.storeKey)
	bz, _ := store.Get(types.KeyForUnstakingPlatforms(unstakingTime))
	if bz == nil {
		return []sdk.Address{}
	}
	err := k.Cdc.UnmarshalBinaryLengthPrefixed(bz, &valAddrs, ctx.BlockHeight())
	if err != nil {
		panic(err)
	}
	return valAddrs

}

// setUnstakingPlatforms - Store platforms in unstaking queue at a certain unstaking time
func (k Keeper) setUnstakingPlatforms(ctx sdk.Ctx, unstakingTime time.Time, keys sdk.Addresses) {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.Cdc.MarshalBinaryLengthPrefixed(&keys, ctx.BlockHeight())
	if err != nil {
		panic(err)
	}
	_ = store.Set(types.KeyForUnstakingPlatforms(unstakingTime), bz)
}

// delteUnstakingPlatforms - Remove all the platforms for a specific unstaking time
func (k Keeper) deleteUnstakingPlatforms(ctx sdk.Ctx, unstakingTime time.Time) {
	store := ctx.KVStore(k.storeKey)
	_ = store.Delete(types.KeyForUnstakingPlatforms(unstakingTime))
}

// unstakingPlatformsIterator - Retrieve an iterator for all unstaking platforms up to a certain time
func (k Keeper) unstakingPlatformsIterator(ctx sdk.Ctx, endTime time.Time) (sdk.Iterator, error) {
	store := ctx.KVStore(k.storeKey)
	return store.Iterator(types.UnstakingPlatformsKey, sdk.InclusiveEndBytes(types.KeyForUnstakingPlatforms(endTime)))
}

// getMaturePlatforms - Retrieve a list of all the mature validators
func (k Keeper) getMaturePlatforms(ctx sdk.Ctx) (matureValsAddrs sdk.Addresses) {
	matureValsAddrs = make([]sdk.Address, 0)
	unstakingValsIterator, _ := k.unstakingPlatformsIterator(ctx, ctx.BlockHeader().Time)
	defer unstakingValsIterator.Close()
	for ; unstakingValsIterator.Valid(); unstakingValsIterator.Next() {
		var platforms sdk.Addresses
		err := k.Cdc.UnmarshalBinaryLengthPrefixed(unstakingValsIterator.Value(), &platforms, ctx.BlockHeight())
		if err != nil {
			panic(err)
		}
		matureValsAddrs = append(matureValsAddrs, platforms...)

	}
	return matureValsAddrs
}

// unstakeAllMatureValidators - Unstake all the unstaking platforms that have finished their unstaking period
func (k Keeper) unstakeAllMaturePlatforms(ctx sdk.Ctx) {
	store := ctx.KVStore(k.storeKey)
	unstakingPlatformsIterator, _ := k.unstakingPlatformsIterator(ctx, ctx.BlockHeader().Time)
	defer unstakingPlatformsIterator.Close()
	for ; unstakingPlatformsIterator.Valid(); unstakingPlatformsIterator.Next() {
		var unstakingVals sdk.Addresses
		err := k.Cdc.UnmarshalBinaryLengthPrefixed(unstakingPlatformsIterator.Value(), &unstakingVals, ctx.BlockHeight())
		if err != nil {
			panic(err)
		}
		for _, valAddr := range unstakingVals {
			val, found := k.GetPlatform(ctx, valAddr)
			if !found {
				k.Logger(ctx).Error(fmt.Errorf("platform %s, in the unstaking queue was not found", valAddr).Error())
				continue
			}
			err := k.ValidatePlatformFinishUnstaking(ctx, val)
			if err != nil {
				ctx.Logger().Error(fmt.Sprintf("Could not finish unstaking mature platform at height %d: ", ctx.BlockHeight()) + err.Error())
				continue
			}
			k.FinishUnstakingPlatform(ctx, val)
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeCompleteUnstaking,
					sdk.NewAttribute(types.AttributeKeyPlatform, valAddr.String()),
				),
			)
			if ctx.IsAfterUpgradeHeight() {
				k.DeletePlatform(ctx, valAddr)
			}
		}
		_ = store.Delete(unstakingPlatformsIterator.Key())
	}
}
