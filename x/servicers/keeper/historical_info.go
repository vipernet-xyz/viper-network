package keeper

import (
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/servicers/types"
)

// GetHistoricalInfo gets the historical info at a given height
func (k Keeper) GetHistoricalInfo(ctx sdk.Ctx, height int64) (types.HistoricalInfo, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetHistoricalInfoKey(height)

	value, _ := store.Get(key)
	if value == nil {
		return types.HistoricalInfo{}, false
	}

	return types.MustUnmarshalHistoricalInfo(k.Cdc, value), true
}
