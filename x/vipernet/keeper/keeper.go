package keeper

import (
	"github.com/vipernet-xyz/viper-network/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/vipernet/types"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/rpc/client"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
)

// Keeper maintains the link to storage and exposes getter/setter methods for the various parts of the state machine
type Keeper struct {
	authKeeper        types.AuthKeeper
	posKeeper         types.PosKeeper
	providerKeeper    types.ProvidersKeeper
	TmNode            client.Client
	hostedBlockchains *types.HostedBlockchains
	hostedGeoZone     *types.HostedGeoZones
	Paramstore        sdk.Subspace
	storeKey          sdk.StoreKey // Unexposed key to access store from sdk.Ctx
	Cdc               *codec.Codec // The wire codec for binary encoding/decoding.
}

// NewKeeper creates new instances of the vipernet module Keeper
func NewKeeper(storeKey sdk.StoreKey, cdc *codec.Codec, authKeeper types.AuthKeeper, posKeeper types.PosKeeper, providerKeeper types.ProvidersKeeper, hostedChains *types.HostedBlockchains, hostedGeoZone *types.HostedGeoZones, paramstore sdk.Subspace) Keeper {
	return Keeper{
		authKeeper:        authKeeper,
		posKeeper:         posKeeper,
		providerKeeper:    providerKeeper,
		hostedBlockchains: hostedChains,
		hostedGeoZone:     hostedGeoZone,
		Paramstore:        paramstore.WithKeyTable(ParamKeyTable()),
		storeKey:          storeKey,
		Cdc:               cdc,
	}
}

func (k Keeper) Codec() *codec.Codec {
	return k.Cdc
}

// "GetBlock" returns the block from the tendermint servicer at a certain height
func (k Keeper) GetBlock(height int) (*coretypes.ResultBlock, error) {
	h := int64(height)
	return k.TmNode.Block(&h)
}

func (k Keeper) ConsensusParamUpdate(ctx sdk.Ctx) *abci.ConsensusParams {
	currentHeightBlockSize := k.BlockByteSize(ctx)
	//If not 0 and different update
	if currentHeightBlockSize > 0 {
		previousBlockCtx, _ := ctx.PrevCtx(ctx.BlockHeight() - 1)
		lastBlockSize := k.BlockByteSize(previousBlockCtx)
		if lastBlockSize != currentHeightBlockSize {
			//not go under default value
			if currentHeightBlockSize < types.DefaultBlockByteSize {
				return &abci.ConsensusParams{}
			}
			return &abci.ConsensusParams{
				Block: &abci.BlockParams{
					MaxBytes: currentHeightBlockSize,
					MaxGas:   -1,
				},
				Evidence: &abci.EvidenceParams{
					MaxAge: 50,
				},
			}
		}
	}

	return &abci.ConsensusParams{}
}
