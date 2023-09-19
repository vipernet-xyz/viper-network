package keeper

import (
	"fmt"

	"github.com/vipernet-xyz/viper-network/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/authentication/types"

	"github.com/tendermint/tendermint/libs/log"
)

// Keeper of the supply store
type Keeper struct {
	Cdc       *codec.Codec
	Bcdc      codec.BinaryCodec
	storeKey  sdk.StoreKey
	subspace  sdk.Subspace
	permAddrs map[string]types.PermissionsForAddress
}

// NewKeeper creates a new Keeper instance
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, subspace sdk.Subspace, maccPerms map[string][]string) Keeper {
	// set the addresses
	permAddrs := make(map[string]types.PermissionsForAddress)
	for name, perms := range maccPerms {
		permAddrs[name] = types.NewPermissionsForAddress(name, perms)
	}

	return Keeper{
		Cdc:       cdc,
		storeKey:  key,
		subspace:  subspace.WithKeyTable(types.ParamKeyTable()),
		permAddrs: permAddrs,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Ctx) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// Codespace returns the keeper's codespace.
func (k Keeper) Codespace() sdk.CodespaceType {
	return types.DefaultCodespace
}
