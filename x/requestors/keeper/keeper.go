package keeper

import (
	"fmt"
	log2 "log"

	"github.com/vipernet-xyz/viper-network/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/requestors/types"

	"github.com/tendermint/tendermint/libs/log"
)

// Implements RequestorSet interface
var _ types.RequestorSet = Keeper{}

// Keeper of the staking store
type Keeper struct {
	storeKey      sdk.StoreKey
	Cdc           *codec.Codec
	AccountKeeper types.AuthKeeper
	POSKeeper     types.PosKeeper
	ViperKeeper   types.ViperKeeper
	Paramstore    sdk.Subspace

	// codespace
	codespace sdk.CodespaceType
	// Cache
	RequestorCache *sdk.Cache
}

// NewKeeper creates a new staking Keeper instance
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, posKeeper types.PosKeeper, supplyKeeper types.AuthKeeper, viperKeeper types.ViperKeeper,
	paramstore sdk.Subspace, codespace sdk.CodespaceType) Keeper {

	// ensure staked module accounts are set
	if addr := supplyKeeper.GetModuleAddress(types.StakedPoolName); addr == nil {
		log2.Fatal(fmt.Errorf("%s module account has not been set", types.StakedPoolName))
	}
	cache := sdk.NewCache(int(types.RequestorCacheSize))

	return Keeper{
		storeKey:       key,
		ViperKeeper:    viperKeeper,
		AccountKeeper:  supplyKeeper,
		POSKeeper:      posKeeper,
		Paramstore:     paramstore.WithKeyTable(ParamKeyTable()),
		codespace:      codespace,
		RequestorCache: cache,
		Cdc:            cdc,
	}
}

// Logger - returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Ctx) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// Codespace - Retrieve the codespace
func (k Keeper) Codespace() sdk.CodespaceType {
	return k.codespace
}
