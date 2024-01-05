package keeper

import (
	"encoding/hex"
	"fmt"
	"time"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/servicers/exported"
	"github.com/vipernet-xyz/viper-network/x/servicers/types"
)

// SetStakedValidator - Store staked validator
func (k Keeper) SetStakedValidator(ctx sdk.Ctx, validator types.Validator) {
	store := ctx.KVStore(k.storeKey)
	_ = store.Set(types.KeyForValidatorInStakingSet(validator), validator.Address)
	// save in the network id stores for quick session generations
	//k.SetStakedValidatorByChains(ctx, validator)
}

// SetStakedValidatorByChains - Store staked validator using networkId
func (k Keeper) SetStakedValidatorByChains(ctx sdk.Ctx, validator types.Validator) {
	store := ctx.KVStore(k.storeKey)
	for _, c := range validator.Chains {
		cBz, err := hex.DecodeString(c)
		if err != nil {
			ctx.Logger().Error(fmt.Errorf("could not hex decode chains for validator: %s with network ID: %s", validator.Address, c).Error())
			continue
		}
		_ = store.Set(types.KeyForValidatorByNetworkID(validator.Address, cBz), []byte{}) // use empty byte slice to save space
	}
}

// SetStakedValidatorByGeoZone - Store staked validator using geoZone
func (k Keeper) SetStakedValidatorByGeoZone(ctx sdk.Ctx, validator types.Validator) {
	store := ctx.KVStore(k.storeKey)
	for _, g := range validator.GeoZone {
		gBz, err := hex.DecodeString(g)
		if err != nil {
			ctx.Logger().Error(fmt.Errorf("could not hex decode chains for validator: %s with geoZone: %s", validator.Address, g).Error())
			continue
		}
		_ = store.Set(types.KeyForValidatorByGeoZone(validator.Address, gBz), []byte{}) // use empty byte slice to save space
	}
}

// GetValidatorByChains - Returns the validator staked by network identifier
func (k Keeper) GetValidatorsByChain(ctx sdk.Ctx, networkID string) (validators []sdk.Address, count int) {
	defer sdk.TimeTrack(time.Now())
	l, exist := sdk.VbCCache.Get(sdk.GetCacheKey(int(ctx.BlockHeight()), networkID))

	if !exist {
		cBz, err := hex.DecodeString(networkID)
		if err != nil {
			ctx.Logger().Error(fmt.Errorf("could not hex decode chains when GetValidatorByChain: with network ID: %s, at height: %d", networkID, ctx.BlockHeight()).Error())
			return
		}
		iterator, _ := k.validatorByChainsIterator(ctx, cBz)
		defer iterator.Close()
		for ; iterator.Valid(); iterator.Next() {
			address := types.AddressForValidatorByNetworkIDKey(iterator.Key(), cBz)
			count++
			validators = append(validators, address)
		}
		if sdk.VbCCache.Cap() > 1 {
			_ = sdk.VbCCache.Add(sdk.GetCacheKey(int(ctx.BlockHeight()), networkID), validators)
		}

		return validators, count
	}

	validators = l.([]sdk.Address)
	return validators, len(validators)
}

func (k Keeper) deleteValidatorForChains(ctx sdk.Ctx, validator types.Validator) {
	store := ctx.KVStore(k.storeKey)
	for _, c := range validator.Chains {
		cBz, err := hex.DecodeString(c)
		if err != nil {
			ctx.Logger().Error(fmt.Errorf("could not hex decode chains for validator: %s with network ID: %s, at height %d", validator.Address, c, ctx.BlockHeight()).Error())
			continue
		}
		_ = store.Delete(types.KeyForValidatorByNetworkID(validator.Address, cBz))
	}
}

func (k Keeper) deleteValidatorForGeoZone(ctx sdk.Ctx, validator types.Validator) {
	store := ctx.KVStore(k.storeKey)
	for _, g := range validator.GeoZone {
		gBz, err := hex.DecodeString(string(g))
		if err != nil {
			ctx.Logger().Error(fmt.Errorf("could not hex decode geozone for validator: %s with geozone: %s, at height %d", validator.Address, g, ctx.BlockHeight()).Error())
			continue
		}
		_ = store.Delete(types.KeyForValidatorByGeoZone(validator.Address, gBz))
	}
}

// validatorByChainsIterator - returns an iterator for the current staked validators
func (k Keeper) validatorByChainsIterator(ctx sdk.Ctx, networkIDBz []byte) (sdk.Iterator, error) {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, types.KeyForValidatorsByNetworkID(networkIDBz))
}

// deleteValidatorFromStakingSet - delete validator from staked set
func (k Keeper) deleteValidatorFromStakingSet(ctx sdk.Ctx, validator types.Validator) {
	store := ctx.KVStore(k.storeKey)
	_ = store.Delete(types.KeyForValidatorInStakingSet(validator))
}

// removeValidatorTokens - Update the staked tokens of an existing validator, update the validators power index key
func (k Keeper) removeValidatorTokens(ctx sdk.Ctx, v types.Validator, tokensToRemove sdk.BigInt) (types.Validator, error) {
	k.deleteValidatorFromStakingSet(ctx, v)
	v, err := v.RemoveStakedTokens(tokensToRemove)
	if err != nil {
		return v, err
	}
	k.SetValidator(ctx, v)
	return v, nil
}

// GetStakedValidators - Retrieve StakedValidators
func (k Keeper) GetStakedValidators(ctx sdk.Ctx) (validators []exported.ValidatorI) {
	store := ctx.KVStore(k.storeKey)
	iterator, _ := sdk.KVStorePrefixIterator(store, types.StakedValidatorsKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		validator, found := k.GetValidator(ctx, iterator.Value())
		if !found {
			ctx.Logger().Error(fmt.Errorf("cannot find validator from staking set: %v, at height %d\n", iterator.Value(), ctx.BlockHeight()).Error())
			continue
		}
		if validator.IsStaked() {
			validators = append(validators, validator)
		}
	}
	return validators
}

// stakedValsIterator - Retrieve an iterator for the current staked validators
func (k Keeper) stakedValsIterator(ctx sdk.Ctx) (sdk.Iterator, error) {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStoreReversePrefixIterator(store, types.StakedValidatorsKey)
}

// IterateAndExecuteOverStakedVals - Goes through the staked validator set and execute handler
func (k Keeper) IterateAndExecuteOverStakedVals(
	ctx sdk.Ctx, fn func(index int64, validator exported.ValidatorI) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator, _ := sdk.KVStoreReversePrefixIterator(store, types.StakedValidatorsKey)
	defer iterator.Close()
	i := int64(0)
	for ; iterator.Valid(); iterator.Next() {
		address := iterator.Value()
		validator, found := k.GetValidator(ctx, address)
		if !found {
			k.Logger(ctx).Error(fmt.Errorf("%s is not found int the main validator state", validator.Address).Error())
			continue
		}
		if validator.IsStaked() {
			stop := fn(i, validator)
			if stop {
				break
			}
			i++
		}
	}
}

// GetValidatorsByGeozone returns the validators staked by geozone identifier
func (k Keeper) GetValidatorsByGeoZone(ctx sdk.Ctx, geoZone string) (validators []sdk.Address, count int) {
	defer sdk.TimeTrack(time.Now())
	l, exist := sdk.VbGZCache.Get(sdk.GetCacheKey(int(ctx.BlockHeight()), geoZone))

	if !exist {
		gBz, err := hex.DecodeString(geoZone)
		if err != nil {
			ctx.Logger().Error(fmt.Errorf("could not hex decode geozone when GetValidatorsByGeozone: with geoZone: %s, at height: %d", geoZone, ctx.BlockHeight()).Error())
			return
		}

		iterator, _ := k.validatorByGeozoneIterator(ctx, gBz)
		defer iterator.Close()
		for ; iterator.Valid(); iterator.Next() {
			address := types.AddressForValidatorByGeozoneKey(iterator.Key(), gBz)
			count++
			validators = append(validators, address)
		}

		if sdk.VbCCache.Cap() > 1 {
			_ = sdk.VbGZCache.Add(sdk.GetCacheKey(int(ctx.BlockHeight()), geoZone), validators)
		}

		return validators, count
	}

	validators = l.([]sdk.Address)
	return validators, len(validators)
}

// validatorByGeozoneIterator returns an iterator for the current staked validators by geozone
func (k Keeper) validatorByGeozoneIterator(ctx sdk.Ctx, geoZoneBz []byte) (sdk.Iterator, error) {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStorePrefixIterator(store, types.KeyForValidatorsByGeoZone(geoZoneBz))
}

func (k Keeper) GetStakedValidatorsLimit(ctx sdk.Ctx, maxRetrieve int64) (validators []exported.ValidatorI) {
	store := ctx.KVStore(k.storeKey)
	validators = make([]exported.ValidatorI, 0, maxRetrieve)
	iterator, _ := sdk.KVStorePrefixIterator(store, types.StakedValidatorsKey)
	defer iterator.Close()

	i := 0
	for ; iterator.Valid() && i < int(maxRetrieve); iterator.Next() {
		validator, found := k.GetValidator(ctx, iterator.Value())
		if !found {
			ctx.Logger().Error(fmt.Errorf("cannot find validator from staking set: %v, at height %d\n", iterator.Value(), ctx.BlockHeight()).Error())
			continue
		}
		if validator.IsStaked() {
			validators = append(validators, validator)
			i++
		}
	}
	return validators
}
