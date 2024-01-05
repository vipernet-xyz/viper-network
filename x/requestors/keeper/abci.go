package keeper

import (
	sdk "github.com/vipernet-xyz/viper-network/types"

	abci "github.com/tendermint/tendermint/abci/types"
)

func BeginBlocker(ctx sdk.Ctx, _ abci.RequestBeginBlock, k Keeper) {}

// EndBlocker - Called at the end of every block, update validator set
func EndBlocker(ctx sdk.Ctx, k Keeper) []abci.ValidatorUpdate {
	// Unstake all mature requestors from the unstakeing queue.
	k.unstakeAllMatureRequestors(ctx)
	return []abci.ValidatorUpdate{}
}
