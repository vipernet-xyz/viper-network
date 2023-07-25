package migrations

import (
	"github.com/cometbft/cometbft/libs/log"
	sdk "github.com/vipernet-xyz/viper-network/types"

	"github.com/vipernet-xyz/viper-network/modules/core/exported"
)

// ClientKeeper expected account IBC client keeper
type ClientKeeper interface {
	GetClientState(ctx sdk.Ctx, clientID string) (exported.ClientState, bool)
	IterateClientStates(ctx sdk.Ctx, prefix []byte, cb func(string, exported.ClientState) bool)
	ClientStore(ctx sdk.Ctx, clientID string) sdk.KVStore
	Logger(ctx sdk.Ctx) log.Logger
}
