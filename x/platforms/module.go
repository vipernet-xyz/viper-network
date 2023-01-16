package pos

import (
	"encoding/json"

	"github.com/vipernet-xyz/viper-network/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/types/module"
	"github.com/vipernet-xyz/viper-network/x/platforms/keeper"
	"github.com/vipernet-xyz/viper-network/x/platforms/types"

	abci "github.com/tendermint/tendermint/abci/types"
)

var (
	_ module.PlatformModule      = PlatformModule{}
	_ module.PlatformModuleBasic = PlatformModuleBasic{}
)

// PlatformModuleBasic defines the basic platform module used by the staking module.
type PlatformModuleBasic struct{}

var _ module.PlatformModuleBasic = PlatformModuleBasic{}

// Name returns the staking module's name.
func (PlatformModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterCodec registers the staking module's types for the given codec.
func (PlatformModuleBasic) RegisterCodec(cdc *codec.Codec) {
	types.RegisterCodec(cdc)
}

// DefaultGenesis returns default genesis state as raw bytes for the staking
// module.
func (PlatformModuleBasic) DefaultGenesis() json.RawMessage {
	return types.ModuleCdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the staking module.
func (PlatformModuleBasic) ValidateGenesis(bz json.RawMessage) error {
	var data types.GenesisState
	err := types.ModuleCdc.UnmarshalJSON(bz, &data)
	if err != nil {
		return err
	}
	return ValidateGenesis(data)
}

// PlatformModule implements an platform module for the staking module.
type PlatformModule struct {
	PlatformModuleBasic
	keeper keeper.Keeper
}

func (pm PlatformModule) ConsensusParamsUpdate(ctx sdk.Ctx) *abci.ConsensusParams {
	return &abci.ConsensusParams{}
}

// NewPlatformModule creates a new PlatformModule object
func NewPlatformModule(keeper keeper.Keeper) PlatformModule {
	return PlatformModule{
		PlatformModuleBasic: PlatformModuleBasic{},
		keeper:              keeper,
	}
}

// Name returns the staking module's name.
func (PlatformModule) Name() string {
	return types.ModuleName
}

// RegisterInvariants registers the staking module invariants.
func (pm PlatformModule) RegisterInvariants(ir sdk.InvariantRegistry) {
}

// Route returns the message routing key for the staking module.
func (PlatformModule) Route() string {
	return types.RouterKey
}

// NewHandler returns an sdk.Handler for the staking module.
func (pm PlatformModule) NewHandler() sdk.Handler {
	return NewHandler(pm.keeper)
}

// QuerierRoute returns the staking module's querier route name.
func (PlatformModule) QuerierRoute() string {
	return types.QuerierRoute
}

// NewQuerierHandler returns the staking module sdk.Querier.
func (pm PlatformModule) NewQuerierHandler() sdk.Querier {
	return keeper.NewQuerier(pm.keeper)
}

// InitGenesis performs genesis initialization for the pos module. It returns
// no platform updates.
func (pm PlatformModule) InitGenesis(ctx sdk.Ctx, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState
	if data == nil {
		genesisState = types.DefaultGenesisState()
	} else {
		types.ModuleCdc.MustUnmarshalJSON(data, &genesisState)
	}
	InitGenesis(ctx, pm.keeper, pm.keeper.AccountKeeper, pm.keeper.POSKeeper, genesisState)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the staking
// module.
func (pm PlatformModule) ExportGenesis(ctx sdk.Ctx) json.RawMessage {
	gs := ExportGenesis(ctx, pm.keeper)
	return types.ModuleCdc.MustMarshalJSON(gs)
}

// module begin-block
func (pm PlatformModule) BeginBlock(ctx sdk.Ctx, req abci.RequestBeginBlock) {
	keeper.BeginBlocker(ctx, req, pm.keeper)
}

func (pm PlatformModule) UpgradeCodec(ctx sdk.Ctx) {
	pm.keeper.UpgradeCodec(ctx)
}

// EndBlock returns the end blocker for the staking module. It returns no platform
// updates.
func (pm PlatformModule) EndBlock(ctx sdk.Ctx, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return keeper.EndBlocker(ctx, pm.keeper)
}
