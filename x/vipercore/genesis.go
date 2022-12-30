package vipercore

import (
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/vipercore/keeper"
	"github.com/vipernet-xyz/viper-network/x/vipercore/types"

	abci "github.com/tendermint/tendermint/abci/types"
)

// "InitGenesis" - Initializes the state with a genesis state object
func InitGenesis(ctx sdk.Ctx, keeper keeper.Keeper, data types.GenesisState) []abci.ValidatorUpdate {
	// set the params in store
	keeper.SetParams(ctx, data.Params)
	// set the claim objects in store
	keeper.SetClaims(ctx, data.Claims)
	return []abci.ValidatorUpdate{}
}

// "ExportGenesis" - Exports the state in a genesis state object
func ExportGenesis(ctx sdk.Ctx, k keeper.Keeper) types.GenesisState {
	return types.GenesisState{
		Params: k.GetParams(ctx),
		Claims: k.GetAllClaims(ctx),
	}
}
