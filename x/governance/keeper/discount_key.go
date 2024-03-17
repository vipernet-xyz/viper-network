package keeper

import (
	"fmt"

	sdk "github.com/vipernet-xyz/viper-network/types"
)

// HasDiscountKey checks if a discount key already exists for the given address
func (k Keeper) HasDiscountKey(ctx sdk.Ctx, addr sdk.Address) bool {
	if k.discountStoreKey == nil {
		k.Logger(ctx).Info("discountStoreKey is nil")
		return false
	}
	store := ctx.KVStore(k.discountStoreKey) // use the discountStoreKey
	h, _ := store.Has(addr.Bytes())
	return h
}

// SetDiscountKey sets a discount key for the given address
func (k Keeper) SetDiscountKey(ctx sdk.Ctx, addr sdk.Address, discountKey string) error {
	store := ctx.KVStore(k.discountStoreKey) // use the discountStoreKey
	h, _ := store.Has(addr.Bytes())
	if h {
		return fmt.Errorf("Discount Key already exists for address %s", addr)
	}
	store.Set(addr.Bytes(), []byte(discountKey))
	return nil
}

func (k Keeper) GetDiscountKey(ctx sdk.Ctx, addr sdk.Address) string {
	store := ctx.KVStore(k.discountStoreKey)
	d, _ := store.Has(addr)
	if !d {
		return "" // Return an empty string if the key does not exist
	}
	dk, _ := store.Get(addr)
	return string(dk)
}
