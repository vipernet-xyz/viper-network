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
	_ module.PlatformModule      = PlatformModule{}
	_ module.PlatformModuleBasic = PlatformModuleBasic{}
)

// PlatformModuleBasic app module basics object
type PlatformModuleBasic struct{}

// Name module name
func (PlatformModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterCodec register module codec
func (PlatformModuleBasic) RegisterCodec(cdc *codec.Codec) {
	types.RegisterCodec(cdc)
}

// DefaultGenesis default genesis state
func (PlatformModuleBasic) DefaultGenesis() json.RawMessage {
	return types.ModuleCdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis module validate genesis
func (PlatformModuleBasic) ValidateGenesis(bz json.RawMessage) error {
	var data types.GenesisState
	err := types.ModuleCdc.UnmarshalJSON(bz, &data)
	if err != nil {
		return err
	}
	return types.ValidateGenesis(data)
}

// PlatformModule app module object
// ___________________________
type PlatformModule struct {
	PlatformModuleBasic
	accountKeeper keeper.Keeper
}

func (am PlatformModule) ConsensusParamsUpdate(ctx sdk.Ctx) *abci.ConsensusParams {
	return &abci.ConsensusParams{}
}

// NewPlatformModule creates a new PlatformModule object
func NewPlatformModule(accountKeeper keeper.Keeper) PlatformModule {
	return PlatformModule{
		PlatformModuleBasic: PlatformModuleBasic{},
		accountKeeper:       accountKeeper,
	}
}

// Name module name
func (PlatformModule) Name() string {
	return types.ModuleName
}

// RegisterInvariants register invariants
func (PlatformModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// Route module message route name
func (PlatformModule) Route() string { return "" }

func (am PlatformModule) UpgradeCodec(ctx sdk.Ctx) {
	am.accountKeeper.UpgradeCodec(ctx)
}

// NewHandler module handler
func (PlatformModule) NewHandler() sdk.Handler { return nil }

// QuerierRoute module querier route name
func (PlatformModule) QuerierRoute() string {
	return types.QuerierRoute
}

// NewQuerierHandler module querier
func (am PlatformModule) NewQuerierHandler() sdk.Querier {
	return NewQuerier(am.accountKeeper)
}

// InitGenesis module init-genesis
func (am PlatformModule) InitGenesis(ctx sdk.Ctx, data json.RawMessage) []abci.ValidatorUpdate {
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
func (am PlatformModule) ExportGenesis(ctx sdk.Ctx) json.RawMessage {
	gs := ExportGenesis(ctx, am.accountKeeper)
	return types.ModuleCdc.MustMarshalJSON(gs)
}

// BeginBlock module begin-block
func (am PlatformModule) BeginBlock(ctx sdk.Ctx, _ abci.RequestBeginBlock) {
}

// EndBlock module end-block
func (PlatformModule) EndBlock(_ sdk.Ctx, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}
