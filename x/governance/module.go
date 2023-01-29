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
	_ module.ProviderModule      = ProviderModule{}
	_ module.ProviderModuleBasic = ProviderModuleBasic{}
)

const moduleName = "governance"

// ProviderModuleBasic app module basics object
type ProviderModuleBasic struct{}

// Name module name
func (ProviderModuleBasic) Name() string {
	return moduleName
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
func (ProviderModuleBasic) ValidateGenesis(_ json.RawMessage) error { return nil }

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
func (pm ProviderModule) RegisterInvariants(_ sdk.InvariantRegistry) {
}

func (pm ProviderModule) UpgradeCodec(ctx sdk.Ctx) {
	pm.keeper.UpgradeCodec(ctx)
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
func (pm ProviderModule) ExportGenesis(ctx sdk.Ctx) json.RawMessage {
	gs := pm.keeper.ExportGenesis(ctx)
	return types.ModuleCdc.MustMarshalJSON(gs)
}

// BeginBlock module begin-block
func (pm ProviderModule) BeginBlock(ctx sdk.Ctx, req abci.RequestBeginBlock) {

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
func ActivateAdditionalParametersACL(ctx sdk.Ctx, pm ProviderModule) {

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
func (pm ProviderModule) EndBlock(ctx sdk.Ctx, req abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}
