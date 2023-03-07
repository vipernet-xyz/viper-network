package governance

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/vipernet-xyz/viper-network/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/types/module"
	"github.com/vipernet-xyz/viper-network/x/governance/keeper"
	"github.com/vipernet-xyz/viper-network/x/governance/types"

	abci "github.com/tendermint/tendermint/abci/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

const moduleName = "governance"

// AppModuleBasic app module basics object
type AppModuleBasic struct{}

// Name module name
func (AppModuleBasic) Name() string {
	return moduleName
}

// RegisterCodec register module codec
func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
	types.RegisterCodec(cdc)
}

// DefaultGenesis default genesis state
func (AppModuleBasic) DefaultGenesis() json.RawMessage {
	return types.ModuleCdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis module validate genesis
func (AppModuleBasic) ValidateGenesis(_ json.RawMessage) error { return nil }

// AppModule implements an provider module for the staking module.
type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
}

func (pm AppModule) ConsensusParamsUpdate(ctx sdk.Ctx) *abci.ConsensusParams {
	return &abci.ConsensusParams{}
}

// NewAppModule creates a new AppModule object
func NewAppModule(keeper keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         keeper,
	}
}

// Name returns the staking module's name.
func (AppModule) Name() string {
	return types.ModuleName
}

// RegisterInvariants registers the staking module invariants.
func (pm AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {
}

func (pm AppModule) UpgradeCodec(ctx sdk.Ctx) {
	pm.keeper.UpgradeCodec(ctx)
}

// Route returns the message routing key for the staking module.
func (AppModule) Route() string {
	return types.RouterKey
}

// NewHandler returns an sdk.Handler for the staking module.
func (pm AppModule) NewHandler() sdk.Handler {
	return NewHandler(pm.keeper)
}

// QuerierRoute returns the staking module's querier route name.
func (AppModule) QuerierRoute() string {
	return types.QuerierRoute
}

// NewQuerierHandler returns the staking module sdk.Querier.
func (pm AppModule) NewQuerierHandler() sdk.Querier {
	return keeper.NewQuerier(pm.keeper)
}

// InitGenesis performs genesis initialization for the pos module. It returns
// no validator updates.
func (pm AppModule) InitGenesis(ctx sdk.Ctx, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState
	if ctx.AppVersion() == "" {
		fmt.Println(fmt.Errorf("must set app version in context, set with ctx.WithAppVersion(<version>)").Error())
		os.Exit(1)
	}
	if data == nil {
		genesisState = types.DefaultGenesisState()
	} else {
		types.ModuleCdc.MustUnmarshalJSON(data, &genesisState)
	}
	return pm.keeper.InitGenesis(ctx, genesisState)
}

// ExportGenesis returns the exported genesis state as raw bytes for the staking
// module.
func (pm AppModule) ExportGenesis(ctx sdk.Ctx) json.RawMessage {
	gs := pm.keeper.ExportGenesis(ctx)
	return types.ModuleCdc.MustMarshalJSON(gs)
}

// BeginBlock module begin-block
func (pm AppModule) BeginBlock(ctx sdk.Ctx, req abci.RequestBeginBlock) {

	ActivateAdditionalParametersACL(ctx, pm)

	u := pm.keeper.GetUpgrade(ctx)
	if ctx.AppVersion() < u.Version && ctx.BlockHeight() >= u.UpgradeHeight() && ctx.BlockHeight() != 0 {
		ctx.Logger().Error("MUST UPGRADE TO NEXT VERSION: ", u.Version)
		ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventMustUpgrade,
			sdk.NewAttribute("VERSION:", u.UpgradeVersion())))
		ctx.Logger().Error(fmt.Sprintf("GRACEFULLY EXITING FOR UPGRADE, AT HEIGHT: %d", ctx.BlockHeight()))
		p, err := os.FindProcess(os.Getpid())
		if err != nil {
			ctx.Logger().Error(err.Error())
			os.Exit(1)
		}
		err = p.Signal(os.Interrupt)
		if err != nil {
			ctx.Logger().Error(err.Error())
			os.Exit(1)
		}
		os.Exit(2)
		select {}
	}
}

// ActivateAdditionalParametersACL ActivateAdditionalParameters activate additional parameters on their respective upgrade heights
func ActivateAdditionalParametersACL(ctx sdk.Ctx, pm AppModule) {

	// activate BlockSizeModify params
	if pm.keeper.GetCodec().IsOnNamedFeatureActivationHeight(ctx.BlockHeight(), codec.BlockSizeModifyKey) {
		gParams := pm.keeper.GetParams(ctx)
		//on the height we get the ACL and insert the key
		gParams.ACL.SetOwner(types.NewACLKey(types.VipercoreSubspace, "BlockByteSize"), pm.keeper.GetDAOOwner(ctx))
		//update params
		pm.keeper.SetParams(ctx, gParams)
	}
	//activate RSCALKey params
	if pm.keeper.GetCodec().IsOnNamedFeatureActivationHeight(ctx.BlockHeight(), codec.RSCALKey) {
		params := pm.keeper.GetParams(ctx)
		params.ACL.SetOwner(types.NewACLKey(types.ServicersSubspace, "MinServicerStakeBinWidth"), pm.keeper.GetDAOOwner(ctx))
		params.ACL.SetOwner(types.NewACLKey(types.ServicersSubspace, "ServicerStakeWeight"), pm.keeper.GetDAOOwner(ctx))
		params.ACL.SetOwner(types.NewACLKey(types.ServicersSubspace, "MaxServicerStakeBin"), pm.keeper.GetDAOOwner(ctx))
		params.ACL.SetOwner(types.NewACLKey(types.ServicersSubspace, "ServicerStakeBinExponent"), pm.keeper.GetDAOOwner(ctx))
		pm.keeper.SetParams(ctx, params)
	}
}

// EndBlock returns the end blocker for the staking module. It returns no validator
// updates.
func (pm AppModule) EndBlock(ctx sdk.Ctx, req abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}
