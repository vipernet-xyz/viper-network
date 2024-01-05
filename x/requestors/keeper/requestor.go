package keeper

import (
	"math"
	"math/big"
	"os"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/requestors/exported"
	"github.com/vipernet-xyz/viper-network/x/requestors/types"
)

// GetRequestor - Retrieve a single requestor from the main store
func (k Keeper) GetRequestor(ctx sdk.Ctx, addr sdk.Address) (requestor types.Requestor, found bool) {
	user, found := k.RequestorCache.GetWithCtx(ctx, addr.String())
	if found && user != nil {
		return user.(types.Requestor), found
	}
	store := ctx.KVStore(k.storeKey)
	value, _ := store.Get(types.KeyForRequestorByAllRequestors(addr))
	if value == nil {
		return requestor, false
	}
	requestor, err := types.UnmarshalRequestor(k.Cdc, ctx, value)
	if err != nil {
		k.Logger(ctx).Error("could not unmarshal requestor from store")
		return requestor, false
	}
	_ = k.RequestorCache.AddWithCtx(ctx, addr.String(), requestor)
	return requestor, true
}

// SetRequestor - Add a single requestor the main store
func (k Keeper) SetRequestor(ctx sdk.Ctx, requestor types.Requestor) {
	store := ctx.KVStore(k.storeKey)
	bz, err := types.MarshalRequestor(k.Cdc, ctx, requestor)
	if err != nil {
		k.Logger(ctx).Error("could not marshal requestor object", err.Error())
		os.Exit(1)
	}
	_ = store.Set(types.KeyForRequestorByAllRequestors(requestor.Address), bz)
	ctx.Logger().Info("Setting Requestor on Main Store " + requestor.Address.String())
	if requestor.IsUnstaking() {
		k.SetUnstakingRequestor(ctx, requestor)
	}
	if requestor.IsStaked() && !requestor.IsJailed() {
		k.SetStakedRequestor(ctx, requestor)
	}
	_ = k.RequestorCache.AddWithCtx(ctx, requestor.Address.String(), requestor)
}

func (k Keeper) SetRequestors(ctx sdk.Ctx, requestors types.Requestors) {
	for _, requestor := range requestors {
		k.SetRequestor(ctx, requestor)
	}
}

// SetValidator - Store validator in the main store
func (k Keeper) DeleteRequestor(ctx sdk.Ctx, addr sdk.Address) {
	store := ctx.KVStore(k.storeKey)
	_ = store.Delete(types.KeyForRequestorByAllRequestors(addr))
	k.RequestorCache.RemoveWithCtx(ctx, addr.String())
}

// GetAllRequestors - Retrieve the set of all requestors with no limits from the main store
func (k Keeper) GetAllRequestors(ctx sdk.Ctx) (requestors types.Requestors) {
	requestors = make([]types.Requestor, 0)
	store := ctx.KVStore(k.storeKey)
	iterator, _ := sdk.KVStorePrefixIterator(store, types.AllRequestorsKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		requestor, err := types.UnmarshalRequestor(k.Cdc, ctx, iterator.Value())
		if err != nil {
			k.Logger(ctx).Error("couldn't unmarshal requestor in GetAllRequestors call: " + string(iterator.Value()) + "\n" + err.Error())
			continue
		}
		requestors = append(requestors, requestor)
	}
	return requestors
}

// GetAllRequestorsWithOpts - Retrieve the set of all requestors with no limits from the main store
func (k Keeper) GetAllRequestorsWithOpts(ctx sdk.Ctx, opts types.QueryRequestorsWithOpts) (requestors types.Requestors) {
	requestors = make([]types.Requestor, 0)
	store := ctx.KVStore(k.storeKey)
	iterator, _ := sdk.KVStorePrefixIterator(store, types.AllRequestorsKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		requestor, err := types.UnmarshalRequestor(k.Cdc, ctx, iterator.Value())
		if err != nil {
			k.Logger(ctx).Error("couldn't unmarshal requestor in GetAllRequestorsWithOpts call: " + string(iterator.Value()) + "\n" + err.Error())
			continue
		}
		if opts.IsValid(requestor) {
			requestors = append(requestors, requestor)
		}
	}
	return requestors
}

// GetRequestors - Retrieve a a given amount of all the requestors
func (k Keeper) GetRequestors(ctx sdk.Ctx, maxRetrieve uint16) (requestors types.Requestors) {
	store := ctx.KVStore(k.storeKey)
	requestors = make([]types.Requestor, maxRetrieve)

	iterator, _ := sdk.KVStorePrefixIterator(store, types.AllRequestorsKey)
	defer iterator.Close()

	i := 0
	for ; iterator.Valid() && i < int(maxRetrieve); iterator.Next() {
		requestor, err := types.UnmarshalRequestor(k.Cdc, ctx, iterator.Value())
		if err != nil {
			k.Logger(ctx).Error("couldn't unmarshal requestor in GetRequestors call: " + string(iterator.Value()) + "\n" + err.Error())
			continue
		}
		requestors[i] = requestor
		i++
	}
	return requestors[:i] // trim if the array length < maxRetrieve
}

// IterateAndExecuteOverRequestors - Goes through the requestor set and perform the provided function
func (k Keeper) IterateAndExecuteOverRequestors(
	ctx sdk.Ctx, fn func(index int64, requestor exported.RequestorI) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator, _ := sdk.KVStorePrefixIterator(store, types.AllRequestorsKey)
	defer iterator.Close()
	i := int64(0)
	for ; iterator.Valid(); iterator.Next() {
		requestor, err := types.UnmarshalRequestor(k.Cdc, ctx, iterator.Value())
		if err != nil {
			k.Logger(ctx).Error("couldn't unmarshal requestor in IterateAndExecuteOverRequestors call: " + string(iterator.Value()) + "\n" + err.Error())
			continue
		}
		stop := fn(i, requestor) // XXX is this safe will the requestor unexposed fields be able to get written to?
		if stop {
			break
		}
		i++
	}
}

func (k Keeper) CalculateRequestorRelays(ctx sdk.Ctx, requestor types.Requestor) sdk.BigInt {
	// If the stake is 0, return max relays from MaxFreeTierRelaysPerSession()
	if requestor.StakedTokens.IsZero() {
		return sdk.NewInt(k.MaxFreeTierRelaysPerSession(ctx))
	}
	stakingAdjustment := sdk.NewDec(k.StakingAdjustment(ctx))
	participationRate := sdk.NewDec(1)
	baseRate := sdk.NewInt(k.BaselineThroughputStakeRate(ctx))
	if k.ParticipationRate(ctx) {
		requestorStakedCoins := k.GetStakedTokens(ctx)
		servicerStakedCoins := k.POSKeeper.GetStakedTokens(ctx)
		totalTokens := k.TotalTokens(ctx)
		participationRate = requestorStakedCoins.Add(servicerStakedCoins).ToDec().Quo(totalTokens.ToDec())
	}
	basePercentage := baseRate.ToDec().Quo(sdk.NewDec(100))
	baselineThroughput := basePercentage.Mul(requestor.StakedTokens.ToDec().Quo(sdk.NewDec(1000000)))
	result := participationRate.Mul(baselineThroughput).Add(stakingAdjustment).TruncateInt()

	// Max Amount of relays Value
	maxRelays := sdk.NewIntFromBigInt(new(big.Int).SetUint64(math.MaxUint64))
	if result.GTE(maxRelays) {
		result = maxRelays
	}

	return result
}

// RelaysPerStakedVIPR = VIPR price(30 day avg.) / (USD relay target * Sessions/Day * Average days per month * ROI target)
