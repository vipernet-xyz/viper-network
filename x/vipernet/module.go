package vipernet

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/vipernet-xyz/viper-network/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/types/module"
	"github.com/vipernet-xyz/viper-network/x/vipernet/keeper"
	"github.com/vipernet-xyz/viper-network/x/vipernet/types"

	abci "github.com/tendermint/tendermint/abci/types"
)

// type check to ensure the interface is properly implemented
var (
	_ module.PlatformModule      = PlatformModule{}
	_ module.PlatformModuleBasic = PlatformModuleBasic{}
)

// PlatformModuleBasic "PlatformModuleBasic" - The fundamental building block of a sdk module
type PlatformModuleBasic struct{}

// Name "Name" - Returns the name of the module
func (PlatformModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterCodec "RegisterCodec" - Registers the codec for the module
func (PlatformModuleBasic) RegisterCodec(cdc *codec.Codec) {
	types.RegisterCodec(cdc)
}

// DefaultGenesis "DefaultGenesis" - Returns the default genesis for the module
func (PlatformModuleBasic) DefaultGenesis() json.RawMessage {
	return types.ModuleCdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis "ValidateGenesis" - Validation check for genesis state bytes
func (PlatformModuleBasic) ValidateGenesis(bytes json.RawMessage) error {
	var data types.GenesisState
	err := types.ModuleCdc.UnmarshalJSON(bytes, &data)
	if err != nil {
		return err
	}
	// Once json successfully marshalled, passes along to genesis.go
	return types.ValidateGenesis(data)
}

// PlatformModule "PlatformModule" - The higher level building block for a module
type PlatformModule struct {
	PlatformModuleBasic               // a fundamental structure for all mods
	keeper              keeper.Keeper // responsible for store operations
}

func (pm PlatformModule) ConsensusParamsUpdate(ctx sdk.Ctx) *abci.ConsensusParams {

	return pm.keeper.ConsensusParamUpdate(ctx)
}

func (pm PlatformModule) UpgradeCodec(ctx sdk.Ctx) {
	pm.keeper.UpgradeCodec(ctx)
}

// NewPlatformModule "NewPlatformModule" - Creates a new PlatformModule Object
func NewPlatformModule(keeper keeper.Keeper) PlatformModule {
	return PlatformModule{
		PlatformModuleBasic: PlatformModuleBasic{},
		keeper:              keeper,
	}
}

// RegisterInvariants "RegisterInvariants" - Unused crisis checking
func (pm PlatformModule) RegisterInvariants(ir sdk.InvariantRegistry) {}

// Route "Route" - returns the route of the module
func (pm PlatformModule) Route() string {
	return types.RouterKey
}

// NewHandler "NewHandler" - returns the handler for the module
func (pm PlatformModule) NewHandler() sdk.Handler {
	return NewHandler(pm.keeper)
}

// QuerierRoute "QuerierRoute" - returns the route of the module for queries
func (pm PlatformModule) QuerierRoute() string {
	return types.ModuleName
}

// NewQuerierHandler "NewQuerierHandler" - returns the query handler for the module
func (pm PlatformModule) NewQuerierHandler() sdk.Querier {
	return keeper.NewQuerier(pm.keeper)
}

// BeginBlock "BeginBlock" - Functionality that is called at the beginning of (every) block
func (pm PlatformModule) BeginBlock(ctx sdk.Ctx, req abci.RequestBeginBlock) {
	ActivateAdditionalParameters(ctx, pm)
	// delete the expired claims
	pm.keeper.DeleteExpiredClaims(ctx)
}

// ActivateAdditionalParameters activate additional parameters on their respective upgrade heights
func ActivateAdditionalParameters(ctx sdk.Ctx, pm PlatformModule) {
	if pm.keeper.Cdc.IsOnNamedFeatureActivationHeight(ctx.BlockHeight(), codec.BlockSizeModifyKey) {
		//on the height we set the default value
		params := pm.keeper.GetParams(ctx)
		params.BlockByteSize = types.DefaultBlockByteSize
		pm.keeper.SetParams(ctx, params)
	}
}

// EndBlock "EndBlock" - Functionality that is called at the end of (every) block
func (pm PlatformModule) EndBlock(ctx sdk.Ctx, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	// get blocks per session
	blocksPerSession := pm.keeper.BlocksPerSession(ctx)
	// get self address
	addr := pm.keeper.GetSelfAddress(ctx)
	if addr != nil {
		// use the offset as a trigger to see if it's time to attempt to submit proofs
		if (ctx.BlockHeight()+int64(addr[0]))%blocksPerSession == 1 && ctx.BlockHeight() != 1 {
			// run go routine because cannot access TmNode during end-block period
			go func() {
				// use this sleep timer to bypass the beginBlock lock over transactions
				time.Sleep(time.Duration(rand.Intn(5000)) * time.Millisecond)
				s, err := pm.keeper.TmNode.Status()
				if err != nil {
					ctx.Logger().Error(fmt.Sprintf("could not get status for tendermint provider (cannot submit claims/proofs in this state): %s", err.Error()))
				} else {
					if !s.SyncInfo.CatchingUp {
						// auto send the proofs
						pm.keeper.SendClaimTx(ctx, pm.keeper, pm.keeper.TmNode, ClaimTx)
						// auto claim the proofs
						pm.keeper.SendProofTx(ctx, pm.keeper.TmNode, ProofTx)
						// clear session cache and db
						types.ClearSessionCache()
					}
				}
			}()
		}
	} else {
		ctx.Logger().Error("could not get self address in end block")
	}
	return []abci.ValidatorUpdate{}
}

// InitGenesis "InitGenesis" - Inits the module genesis from raw json
func (pm PlatformModule) InitGenesis(ctx sdk.Ctx, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState
	if data == nil {
		genesisState = types.DefaultGenesisState()
	} else {
		types.ModuleCdc.MustUnmarshalJSON(data, &genesisState)
	}
	return InitGenesis(ctx, pm.keeper, genesisState)
}

// ExportGenesis "ExportGenesis" - Exports the genesis from raw json
func (pm PlatformModule) ExportGenesis(ctx sdk.Ctx) json.RawMessage {
	gs := ExportGenesis(ctx, pm.keeper)
	return types.ModuleCdc.MustMarshalJSON(gs)
}
