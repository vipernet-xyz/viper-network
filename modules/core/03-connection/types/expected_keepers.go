package types

import (
	sdk "github.com/vipernet-xyz/viper-network/types"

	"github.com/vipernet-xyz/viper-network/modules/core/exported"
)

// ClientKeeper expected account IBC client keeper
type ClientKeeper interface {
	GetClientState(ctx sdk.Ctx, clientID string) (exported.ClientState, bool)
	GetClientConsensusState(ctx sdk.Ctx, clientID string, height exported.Height) (exported.ConsensusState, bool)
	GetSelfConsensusState(ctx sdk.Ctx, height exported.Height) (exported.ConsensusState, error)
	ValidateSelfClient(ctx sdk.Ctx, clientState exported.ClientState) error
	IterateClientStates(ctx sdk.Ctx, prefix []byte, cb func(string, exported.ClientState) bool)
	ClientStore(ctx sdk.Ctx, clientID string) sdk.KVStore
}
