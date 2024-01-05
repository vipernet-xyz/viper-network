package keeper

import (
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/requestors/exported"
	"github.com/vipernet-xyz/viper-network/x/requestors/types"
)

// Requestor - wrrequestorer for GetRequestor call
func (k Keeper) Requestor(ctx sdk.Ctx, address sdk.Address) exported.RequestorI {
	requestor, found := k.GetRequestor(ctx, address)
	if !found {
		return nil
	}
	return requestor
}

// AllRequestors - Retrieve a list of all requestors
func (k Keeper) AllRequestors(ctx sdk.Ctx) (requestors []exported.RequestorI) {
	store := ctx.KVStore(k.storeKey)
	iterator, _ := sdk.KVStorePrefixIterator(store, types.AllRequestorsKey)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		requestor, err := types.UnmarshalRequestor(k.Cdc, ctx, iterator.Value())
		if err != nil {
			k.Logger(ctx).Error("couldn't unmarshal requestor in AllRequestors call: " + string(iterator.Value()) + "\n" + err.Error())
			continue
		}
		requestors = append(requestors, requestor)
	}
	return requestors
}
