package keeper

import (
	"bytes"
	"fmt"
	"time"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/requestors/types"
)

// SetUnstakingRequestor - Store an requestor address to the appropriate position in the unstaking queue
func (k Keeper) SetUnstakingRequestor(ctx sdk.Ctx, val types.Requestor) {
	requestors := k.getUnstakingRequestors(ctx, val.UnstakingCompletionTime)
	requestors = append(requestors, val.Address)
	k.setUnstakingRequestors(ctx, val.UnstakingCompletionTime, requestors)
}

// deleteUnstakingRequestor - DeleteEvidence an requestor address from the unstaking queue
func (k Keeper) deleteUnstakingRequestor(ctx sdk.Ctx, val types.Requestor) {
	requestors := k.getUnstakingRequestors(ctx, val.UnstakingCompletionTime)
	var newRequestors []sdk.Address
	for _, addr := range requestors {
		if !bytes.Equal(addr, val.Address) {
			newRequestors = append(newRequestors, addr)
		}
	}
	if len(newRequestors) == 0 {
		k.deleteUnstakingRequestors(ctx, val.UnstakingCompletionTime)
	} else {
		k.setUnstakingRequestors(ctx, val.UnstakingCompletionTime, newRequestors)
	}
}

// getAllUnstakingRequestors - Retrieve the set of all unstaking requestors with no limits
func (k Keeper) getAllUnstakingRequestors(ctx sdk.Ctx) (requestors []types.Requestor) {
	requestors = make(types.Requestors, 0)
	store := ctx.KVStore(k.storeKey)
	iterator, _ := sdk.KVStorePrefixIterator(store, types.UnstakingRequestorsKey)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var addrs sdk.Addresses
		err := k.Cdc.UnmarshalBinaryLengthPrefixed(iterator.Value(), &addrs)
		if err != nil {
			k.Logger(ctx).Error(fmt.Errorf("could not unmarshal unstakingRequestors in getAllUnstakingRequestors call: %s", string(iterator.Value())).Error())
			return
		}
		for _, addr := range addrs {
			requestor, found := k.GetRequestor(ctx, addr)
			if !found {
				k.Logger(ctx).Error(fmt.Errorf("requestor %s in unstakingSet but not found in all requestors store", requestor.Address).Error())
				continue
			}
			requestors = append(requestors, requestor)
		}

	}
	return requestors
}

// getUnstakingRequestors - Retrieve all of the requestors who will be unstaked at exactly this time
func (k Keeper) getUnstakingRequestors(ctx sdk.Ctx, unstakingTime time.Time) (valAddrs sdk.Addresses) {
	store := ctx.KVStore(k.storeKey)
	bz, _ := store.Get(types.KeyForUnstakingRequestors(unstakingTime))
	if bz == nil {
		return []sdk.Address{}
	}
	err := k.Cdc.UnmarshalBinaryLengthPrefixed(bz, &valAddrs)
	if err != nil {
		panic(err)
	}
	return valAddrs

}

// setUnstakingRequestors - Store requestors in unstaking queue at a certain unstaking time
func (k Keeper) setUnstakingRequestors(ctx sdk.Ctx, unstakingTime time.Time, keys sdk.Addresses) {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.Cdc.MarshalBinaryLengthPrefixed(&keys)
	if err != nil {
		panic(err)
	}
	_ = store.Set(types.KeyForUnstakingRequestors(unstakingTime), bz)
}

// delteUnstakingRequestors - Remove all the requestors for a specific unstaking time
func (k Keeper) deleteUnstakingRequestors(ctx sdk.Ctx, unstakingTime time.Time) {
	store := ctx.KVStore(k.storeKey)
	_ = store.Delete(types.KeyForUnstakingRequestors(unstakingTime))
}

// unstakingRequestorsIterator - Retrieve an iterator for all unstaking requestors up to a certain time
func (k Keeper) unstakingRequestorsIterator(ctx sdk.Ctx, endTime time.Time) (sdk.Iterator, error) {
	store := ctx.KVStore(k.storeKey)
	return store.Iterator(types.UnstakingRequestorsKey, sdk.InclusiveEndBytes(types.KeyForUnstakingRequestors(endTime)))
}

// getMatureRequestors - Retrieve a list of all the mature validators
func (k Keeper) getMatureRequestors(ctx sdk.Ctx) (matureValsAddrs sdk.Addresses) {
	matureValsAddrs = make([]sdk.Address, 0)
	unstakingValsIterator, _ := k.unstakingRequestorsIterator(ctx, ctx.BlockHeader().Time)
	defer unstakingValsIterator.Close()
	for ; unstakingValsIterator.Valid(); unstakingValsIterator.Next() {
		var requestors sdk.Addresses
		err := k.Cdc.UnmarshalBinaryLengthPrefixed(unstakingValsIterator.Value(), &requestors)
		if err != nil {
			panic(err)
		}
		matureValsAddrs = append(matureValsAddrs, requestors...)

	}
	return matureValsAddrs
}

// unstakeAllMatureValidators - Unstake all the unstaking requestors that have finished their unstaking period
func (k Keeper) unstakeAllMatureRequestors(ctx sdk.Ctx) {
	store := ctx.KVStore(k.storeKey)
	unstakingRequestorsIterator, _ := k.unstakingRequestorsIterator(ctx, ctx.BlockHeader().Time)
	defer unstakingRequestorsIterator.Close()
	for ; unstakingRequestorsIterator.Valid(); unstakingRequestorsIterator.Next() {
		var unstakingVals sdk.Addresses
		err := k.Cdc.UnmarshalBinaryLengthPrefixed(unstakingRequestorsIterator.Value(), &unstakingVals)
		if err != nil {
			panic(err)
		}
		for _, valAddr := range unstakingVals {
			val, found := k.GetRequestor(ctx, valAddr)
			if !found {
				k.Logger(ctx).Error(fmt.Errorf("requestor %s, in the unstaking queue was not found", valAddr).Error())
				continue
			}
			err := k.ValidateRequestorFinishUnstaking(ctx, val)
			if err != nil {
				ctx.Logger().Error(fmt.Sprintf("Could not finish unstaking mature requestor at height %d: ", ctx.BlockHeight()) + err.Error())
				continue
			}
			k.FinishUnstakingRequestor(ctx, val)
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeCompleteUnstaking,
					sdk.NewAttribute(types.AttributeKeyRequestor, valAddr.String()),
				),
			)
			k.DeleteRequestor(ctx, valAddr)

		}
		_ = store.Delete(unstakingRequestorsIterator.Key())
	}
}
