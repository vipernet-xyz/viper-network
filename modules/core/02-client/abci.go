package client

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/vipernet-xyz/ibc-go/v7/modules/core/02-client/keeper"
	ibctm "github.com/vipernet-xyz/ibc-go/v7/modules/light-clients/07-tendermint"
)

// BeginBlocker is used to perform IBC client upgrades
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	plan, found := k.GetUpgradePlan(ctx)
	if found {
		// Once we are at the last block this chain will commit, set the upgraded consensus state
		// so that IBC clients can use the last NextValidatorsHash as a trusted kernel for verifying
		// headers on the next version of the chain.
		// Set the time to the last block time of the current chain.
		// In order for a client to upgrade successfully, the first block of the new chain must be committed
		// within the trusting period of the last block time on this chain.
		_, exists := k.GetUpgradedClient(ctx, plan.Height)
		if exists && ctx.BlockHeight() == plan.Height-1 {
			upgradedConsState := &ibctm.ConsensusState{
				Timestamp:          ctx.BlockTime(),
				NextValidatorsHash: ctx.BlockHeader().NextValidatorsHash,
			}
			bz := k.MustMarshalConsensusState(upgradedConsState)

			// SetUpgradedConsensusState always returns nil, hence the blank here.
			_ = k.SetUpgradedConsensusState(ctx, plan.Height, bz)

			keeper.EmitUpgradeChainEvent(ctx, plan.Height)
		}
	}
}
