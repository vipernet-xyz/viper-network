package keeper

import (
	"fmt"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/requestors/exported"
	"github.com/vipernet-xyz/viper-network/x/requestors/types"
)

// SetStakedRequestor - Store staked requestor
func (k Keeper) SetStakedRequestor(ctx sdk.Ctx, requestor types.Requestor) {
	if requestor.Jailed {
		return // jailed requestors are not kept in the staking set
	}
	store := ctx.KVStore(k.storeKey)
	_ = store.Set(types.KeyForRequestorInStakingSet(requestor), requestor.Address)
}

// StakeDenom - Retrieve the denomination of coins.
func (k Keeper) StakeDenom(ctx sdk.Ctx) string {
	return k.POSKeeper.StakeDenom(ctx)
}

// deleteRequestorFromStakingSet - Remove requestor from staked set
func (k Keeper) deleteRequestorFromStakingSet(ctx sdk.Ctx, requestor types.Requestor) {
	store := ctx.KVStore(k.storeKey)
	_ = store.Delete(types.KeyForRequestorInStakingSet(requestor))
}

// RemoveRequestorTokens - Update the staked tokens of an existing requestor, update the requestors power index key
func (k Keeper) removeRequestorTokens(ctx sdk.Ctx, requestor types.Requestor, tokensToRemove sdk.BigInt) (types.Requestor, error) {
	k.deleteRequestorFromStakingSet(ctx, requestor)
	requestor, err := requestor.RemoveStakedTokens(tokensToRemove)
	if err != nil {
		return types.Requestor{}, err
	}
	requestor.MaxRelays = k.CalculateRequestorRelays(ctx, requestor)
	k.SetRequestor(ctx, requestor)
	k.SetStakedRequestor(ctx, requestor)
	return requestor, nil
}

// getStakedRequestors - Retrieve the current staked requestors sorted by power-rank
func (k Keeper) getStakedRequestors(ctx sdk.Ctx) types.Requestors {
	var requestors = make(types.Requestors, 0)
	iterator, _ := k.stakedRequestorsIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		address := iterator.Value()
		requestor, found := k.GetRequestor(ctx, address)
		if !found {
			k.Logger(ctx).Error(fmt.Errorf("requestor %s in staking set but not found in all requestors store", address).Error())
			continue
		}
		if requestor.IsStaked() {
			requestors = append(requestors, requestor)
		}
	}
	return requestors
}

// getStakedRequestorsCount returns a count of the total staked requestorlcations currently
func (k Keeper) getStakedRequestorsCount(ctx sdk.Ctx) (count int64) {
	iterator, _ := k.stakedRequestorsIterator(ctx)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		count++
	}
	return
}

// stakedRequestorsIterator - Retrieve an iterator for the current staked requestors
func (k Keeper) stakedRequestorsIterator(ctx sdk.Ctx) (sdk.Iterator, error) {
	store := ctx.KVStore(k.storeKey)
	return sdk.KVStoreReversePrefixIterator(store, types.StakedRequestorsKey)
}

// IterateAndExecuteOverStakedRequestors - Goes through the staked requestor set and execute handler
func (k Keeper) IterateAndExecuteOverStakedRequestors(
	ctx sdk.Ctx, fn func(index int64, requestor exported.RequestorI) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator, _ := sdk.KVStoreReversePrefixIterator(store, types.StakedRequestorsKey)
	defer iterator.Close()
	i := int64(0)
	for ; iterator.Valid(); iterator.Next() {
		address := iterator.Value()
		requestor, found := k.GetRequestor(ctx, address)
		if !found {
			k.Logger(ctx).Error(fmt.Errorf("requestor %s in staking set but not found in all requestors store", address).Error())
			continue
		}
		if requestor.IsStaked() {
			stop := fn(i, requestor) // XXX is this safe will the requestor unexposed fields be able to get written to?
			if stop {
				break
			}
			i++
		}
	}
}
