package keeper

import (
	"fmt"

	"github.com/vipernet-xyz/viper-network/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/governance/types"

	"github.com/tendermint/tendermint/libs/log"
)

// Keeper of the global paramstore
type Keeper struct {
	cdc              *codec.Codec
	key              sdk.StoreKey
	tkey             sdk.StoreKey
	discountStoreKey sdk.StoreKey
	codespace        sdk.CodespaceType
	paramstore       sdk.Subspace
	AuthKeeper       types.AuthKeeper
	spaces           map[string]sdk.Subspace
}

func NewKeeper(cdc *codec.Codec, key *sdk.KVStoreKey, tkey *sdk.TransientStoreKey, discountStoreKey *sdk.KVStoreKey, codespace sdk.CodespaceType, authKeeper types.AuthKeeper, subspaces ...sdk.Subspace) (k Keeper) {
	k = Keeper{
		cdc:              cdc,
		key:              key,
		tkey:             tkey,
		discountStoreKey: discountStoreKey,
		codespace:        codespace,
		AuthKeeper:       authKeeper,
		spaces:           make(map[string]sdk.Subspace),
	}
	k.paramstore = sdk.NewSubspace(types.ModuleName).WithKeyTable(types.ParamKeyTable())
	k.paramstore.SetCodec(k.cdc)
	subspaces = append(subspaces, k.paramstore)
	k.AddSubspaces(subspaces...)
	return k
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Ctx) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
