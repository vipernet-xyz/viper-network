package keeper

import (
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/platforms/exported"
	"github.com/vipernet-xyz/viper-network/x/platforms/types"
)

// Platform - wrplatformer for GetPlatform call
func (k Keeper) Platform(ctx sdk.Ctx, address sdk.Address) exported.PlatformI {
	platform, found := k.GetPlatform(ctx, address)
	if !found {
		return nil
	}
	return platform
}

// AllPlatforms - Retrieve a list of all platforms
func (k Keeper) AllPlatforms(ctx sdk.Ctx) (platforms []exported.PlatformI) {
	store := ctx.KVStore(k.storeKey)
	iterator, _ := sdk.KVStorePrefixIterator(store, types.AllPlatformsKey)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		platform, err := types.UnmarshalPlatform(k.Cdc, ctx, iterator.Value())
		if err != nil {
			k.Logger(ctx).Error("couldn't unmarshal platform in AllPlatforms call: " + string(iterator.Value()) + "\n" + err.Error())
			continue
		}
		platforms = append(platforms, platform)
	}
	return platforms
}
