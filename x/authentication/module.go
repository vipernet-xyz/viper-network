package authentication

import (
	"encoding/json"

	"github.com/vipernet-xyz/viper-network/x/authentication/keeper"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/vipernet-xyz/viper-network/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/types/module"
	"github.com/vipernet-xyz/viper-network/x/authentication/types"
)

var (
	_ module.ProviderModule      = ProviderModule{}
	_ module.ProviderModuleBasic = ProviderModuleBasic{}
)

// ProviderModuleBasic app module basics object
type ProviderModuleBasic struct{}

// Name module name
func (ProviderModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterCodec register module codec
func (ProviderModuleBasic) RegisterCodec(cdc *codec.Codec) {
	types.RegisterCodec(cdc)
}

// DefaultGenesis default genesis state
func (ProviderModuleBasic) DefaultGenesis() json.RawMessage {
	return types.ModuleCdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis module validate genesis
func (ProviderModuleBasic) ValidateGenesis(bz json.RawMessage) error {
	var data types.GenesisState
	err := types.ModuleCdc.UnmarshalJSON(bz, &data)
	if err != nil {
		return err
	}
	return types.ValidateGenesis(data)
}

// ProviderModule app module object
// ___________________________
type ProviderModule struct {
	ProviderModuleBasic
	accountKeeper keeper.Keeper
}

func (am ProviderModule) ConsensusParamsUpdate(ctx sdk.Ctx) *abci.ConsensusParams {
	return &abci.ConsensusParams{}
}

// NewProviderModule creates a new ProviderModule object
func NewProviderModule(accountKeeper keeper.Keeper) ProviderModule {
	return ProviderModule{
		ProviderModuleBasic: ProviderModuleBasic{},
		accountKeeper:       accountKeeper,
	}
}

// Name module name
func (ProviderModule) Name() string {
	return types.ModuleName
}

// RegisterInvariants register invariants
func (ProviderModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// Route module message route name
func (ProviderModule) Route() string { return "" }

func (am ProviderModule) UpgradeCodec(ctx sdk.Ctx) {
	am.accountKeeper.UpgradeCodec(ctx)
}

// NewHandler module handler
func (ProviderModule) NewHandler() sdk.Handler { return nil }

// QuerierRoute module querier route name
func (ProviderModule) QuerierRoute() string {
	return types.QuerierRoute
}

// NewQuerierHandler module querier
func (am ProviderModule) NewQuerierHandler() sdk.Querier {
	return NewQuerier(am.accountKeeper)
}

// InitGenesis module init-genesis
func (am ProviderModule) InitGenesis(ctx sdk.Ctx, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState
	if data == nil {
		genesisState = types.DefaultGenesisState()
	} else {
		ModuleCdc.MustUnmarshalJSON(data, &genesisState)
	}
	InitGenesis(ctx, am.accountKeeper, genesisState)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis module export genesis
func (am ProviderModule) ExportGenesis(ctx sdk.Ctx) json.RawMessage {
	gs := ExportGenesis(ctx, am.accountKeeper)
	return types.ModuleCdc.MustMarshalJSON(gs)
}

// BeginBlock module begin-block
func (am ProviderModule) BeginBlock(ctx sdk.Ctx, _ abci.RequestBeginBlock) {
}

// EndBlock module end-block
func (ProviderModule) EndBlock(_ sdk.Ctx, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}
