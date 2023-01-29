package servicers

import (
	"encoding/json"

	"github.com/vipernet-xyz/viper-network/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/types/module"
	"github.com/vipernet-xyz/viper-network/x/servicers/keeper"
	"github.com/vipernet-xyz/viper-network/x/servicers/types"

	abci "github.com/tendermint/tendermint/abci/types"
)

var (
	_ module.ProviderModule      = ProviderModule{}
	_ module.ProviderModuleBasic = ProviderModuleBasic{}
)

// ProviderModuleBasic defines the basic provider module used by the staking module.
type ProviderModuleBasic struct{}

var _ module.ProviderModuleBasic = ProviderModuleBasic{}

// Name returns the staking module's name.
func (ProviderModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterCodec registers the staking module's types for the given codec.
func (ProviderModuleBasic) RegisterCodec(cdc *codec.Codec) {
	types.RegisterCodec(cdc)
}

func (pm ProviderModule) UpgradeCodec(ctx sdk.Ctx) {
	pm.keeper.UpgradeCodec(ctx)
}

// DefaultGenesis returns default genesis state as raw bytes for the staking
// module.
func (ProviderModuleBasic) DefaultGenesis() json.RawMessage {
	return types.ModuleCdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the staking module.
func (ProviderModuleBasic) ValidateGenesis(bz json.RawMessage) error {
	var data types.GenesisState
	err := types.ModuleCdc.UnmarshalJSON(bz, &data)
	if err != nil {
		return err
	}
	return ValidateGenesis(data)
}

// ProviderModule implements an provider module for the staking module.
type ProviderModule struct {
	ProviderModuleBasic
	keeper keeper.Keeper
}

func (pm ProviderModule) ConsensusParamsUpdate(ctx sdk.Ctx) *abci.ConsensusParams {
	return &abci.ConsensusParams{}
}

// NewProviderModule creates a new ProviderModule object
func NewProviderModule(keeper keeper.Keeper) ProviderModule {
	return ProviderModule{
		ProviderModuleBasic: ProviderModuleBasic{},
		keeper:              keeper,
	}
}

// Name returns the staking module's name.
func (ProviderModule) Name() string {
	return types.ModuleName
}

// RegisterInvariants registers the staking module invariants.
func (pm ProviderModule) RegisterInvariants(ir sdk.InvariantRegistry) {
}

// Route returns the message routing key for the staking module.
func (ProviderModule) Route() string {
	return types.RouterKey
}

// NewHandler returns an sdk.Handler for the staking module.
func (pm ProviderModule) NewHandler() sdk.Handler {
	return NewHandler(pm.keeper)
}

// QuerierRoute returns the staking module's querier route name.
func (ProviderModule) QuerierRoute() string {
	return types.QuerierRoute
}

// NewQuerierHandler returns the staking module sdk.Querier.
func (pm ProviderModule) NewQuerierHandler() sdk.Querier {
	return keeper.NewQuerier(pm.keeper)
}

// InitGenesis performs genesis initialization for the pos module. It returns
// no validator updates.
func (pm ProviderModule) InitGenesis(ctx sdk.Ctx, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState
	if data == nil {
		genesisState = types.DefaultGenesisState()
	} else {
		types.ModuleCdc.MustUnmarshalJSON(data, &genesisState)
	}
	return InitGenesis(ctx, pm.keeper, pm.keeper.AccountKeeper, genesisState)
}

// ExportGenesis returns the exported genesis state as raw bytes for the staking
// module.
func (pm ProviderModule) ExportGenesis(ctx sdk.Ctx) json.RawMessage {
	gs := ExportGenesis(ctx, pm.keeper)
	return types.ModuleCdc.MustMarshalJSON(gs)
}

// BeginBlock module begin-block
func (pm ProviderModule) BeginBlock(ctx sdk.Ctx, req abci.RequestBeginBlock) {
	ActivateAdditionalParameters(ctx, pm)
	keeper.BeginBlocker(ctx, req, pm.keeper)
}

// ActivateAdditionalParameters activate additional parameters on their respective upgrade heights
func ActivateAdditionalParameters(ctx sdk.Ctx, pm ProviderModule) {
	if pm.keeper.Cdc.IsOnNamedFeatureActivationHeight(ctx.BlockHeight(), codec.RSCALKey) {
		//on the height we set the default value
		params := pm.keeper.GetParams(ctx)
		params.MinServicerStakeBinWidth = types.DefaultMinServicerStakeBinWidth
		params.ServicerStakeWeight = types.DefaultServicerStakeWeight
		params.MaxServicerStakeBin = types.DefaultMaxServicerStakeBin
		params.ServicerStakeBinExponent = types.DefaultServicerStakeBinExponent
		// custom logic for minSignedPerWindow
		params.MinSignedPerWindow = params.MinSignedPerWindow.QuoInt64(params.SignedBlocksWindow)
		pm.keeper.SetParams(ctx, params)
	}
}

// EndBlock returns the end blocker for the staking module. It returns no validator
// updates.
func (pm ProviderModule) EndBlock(ctx sdk.Ctx, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return keeper.EndBlocker(ctx, pm.keeper)
}
