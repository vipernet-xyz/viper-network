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

func (k Keeper) UpgradeCodec(ctx sdk.Ctx) {
	if ctx.IsOnUpgradeHeight() {
		k.ConvertState(ctx)
	}
}

func (k Keeper) ConvertState(ctx sdk.Ctx) {
	k.cdc.SetUpgradeOverride(false)
	params := k.GetParams(ctx)
	k.cdc.SetUpgradeOverride(true)
	k.SetParams(ctx, params)
	k.cdc.DisableUpgradeOverride()
}

// HasDiscountKey checks if a discount key already exists for the given address
func (k Keeper) HasDiscountKey(ctx sdk.Context, addr sdk.Address) bool {
	store := ctx.KVStore(k.discountStoreKey) // use the discountStoreKey
	h, _ := store.Has(addr.Bytes())
	return h
}

// SetDiscountKey sets a discount key for the given address
func (k Keeper) SetDiscountKey(ctx sdk.Context, addr sdk.Address, discountKey string) error {
	store := ctx.KVStore(k.discountStoreKey) // use the discountStoreKey
	h, _ := store.Has(addr.Bytes())
	if h {
		return fmt.Errorf("Discount Key already exists for address %s", addr)
	}
	store.Set(addr.Bytes(), []byte(discountKey))
	return nil
}
