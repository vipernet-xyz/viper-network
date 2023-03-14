package v7

import (
	sdk "github.com/vipernet-xyz/viper-network/types"

	"github.com/vipernet-xyz/viper-network/modules/core/exported"
)

// ClientKeeper expected IBC client keeper
type ClientKeeper interface {
	GetClientState(ctx sdk.Context, clientID string) (exported.ClientState, bool)
	SetClientState(ctx sdk.Context, clientID string, clientState exported.ClientState)
	ClientStore(ctx sdk.Context, clientID string) sdk.KVStore
}
