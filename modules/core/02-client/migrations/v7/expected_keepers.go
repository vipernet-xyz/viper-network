package v7

import (
	sdk "github.com/vipernet-xyz/viper-network/types"

	"github.com/vipernet-xyz/viper-network/modules/core/exported"
)

// ClientKeeper expected IBC client keeper
type ClientKeeper interface {
	GetClientState(ctx sdk.Ctx, clientID string) (exported.ClientState, bool)
	SetClientState(ctx sdk.Ctx, clientID string, clientState exported.ClientState)
	ClientStore(ctx sdk.Ctx, clientID string) sdk.KVStore
}
