package ica

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/vipernet-xyz/viper-network/client"
	"github.com/vipernet-xyz/viper-network/codec"
	codectypes "github.com/vipernet-xyz/viper-network/codec/types"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/types/module"

	"github.com/vipernet-xyz/viper-network/modules/apps/27-interchain-accounts/client/cli"
	controllerkeeper "github.com/vipernet-xyz/viper-network/modules/apps/27-interchain-accounts/controller/keeper"
	controllertypes "github.com/vipernet-xyz/viper-network/modules/apps/27-interchain-accounts/controller/types"
	genesistypes "github.com/vipernet-xyz/viper-network/modules/apps/27-interchain-accounts/genesis/types"
	"github.com/vipernet-xyz/viper-network/modules/apps/27-interchain-accounts/host"
	hostkeeper "github.com/vipernet-xyz/viper-network/modules/apps/27-interchain-accounts/host/keeper"
	hosttypes "github.com/vipernet-xyz/viper-network/modules/apps/27-interchain-accounts/host/types"
	"github.com/vipernet-xyz/viper-network/modules/apps/27-interchain-accounts/types"
	porttypes "github.com/vipernet-xyz/viper-network/modules/core/05-port/types"
	ibchost "github.com/vipernet-xyz/viper-network/modules/core/24-host"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}

	_ porttypes.IBCModule = host.IBCModule{}
)

// AppModuleBasic is the IBC interchain accounts AppModuleBasic
type AppModuleBasic struct{}

// Name implements AppModuleBasic interface
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec implements AppModuleBasic.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {}

// RegisterInterfaces registers module concrete types into protobuf Any
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	controllertypes.RegisterInterfaces(registry)
	types.RegisterInterfaces(registry)
}

// DefaultGenesis returns default genesis state as raw bytes for the IBC
// interchain accounts module
func (AppModuleBasic) DefaultGenesis() json.RawMessage {
	return types.ModuleCdc.MustMarshalJSON(genesistypes.DefaultGenesis())
}

// ValidateGenesis performs genesis state validation for the 29-fee module.
func (AppModuleBasic) ValidateGenesis(bz json.RawMessage) error {
	var gs genesistypes.GenesisState
	if err := types.ModuleCdc.UnmarshalJSON(bz, &gs); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}

	return gs.Validate()
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the interchain accounts module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	err := controllertypes.RegisterQueryHandlerClient(context.Background(), mux, controllertypes.NewQueryClient(clientCtx))
	if err != nil {
		panic(err)
	}

	err = hosttypes.RegisterQueryHandlerClient(context.Background(), mux, hosttypes.NewQueryClient(clientCtx))
	if err != nil {
		panic(err)
	}
}

// GetTxCmd implements AppModuleBasic interface
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.NewTxCmd()
}

// GetQueryCmd implements AppModuleBasic interface
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

// AppModule is the application module for the IBC interchain accounts module
type AppModule struct {
	AppModuleBasic
	controllerKeeper *controllerkeeper.Keeper
	hostKeeper       *hostkeeper.Keeper
}

// NewAppModule creates a new IBC interchain accounts module
func NewAppModule(controllerKeeper *controllerkeeper.Keeper, hostKeeper *hostkeeper.Keeper) AppModule {
	return AppModule{
		controllerKeeper: controllerKeeper,
		hostKeeper:       hostKeeper,
	}
}

// InitModule will initialize the interchain accounts moudule. It should only be
// called once and as an alternative to InitGenesis.
func (am AppModule) InitModule(ctx sdk.Ctx, controllerParams controllertypes.Params, hostParams hosttypes.Params) {
	if am.controllerKeeper != nil {
		am.controllerKeeper.SetParams(ctx, controllerParams)
	}

	if am.hostKeeper != nil {
		am.hostKeeper.SetParams(ctx, hostParams)

		cap := am.hostKeeper.BindPort(ctx, types.HostPortID)
		if err := am.hostKeeper.ClaimCapability(ctx, cap, ibchost.PortPath(types.HostPortID)); err != nil {
			panic(fmt.Sprintf("could not claim port capability: %v", err))
		}
	}
}

// RegisterInvariants implements the AppModule interface
func (AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
}

// RegisterServices registers module services
func (am AppModule) RegisterServices(cfg module.Configurator) {
	if am.controllerKeeper != nil {
		controllertypes.RegisterMsgServer(cfg.MsgServer(), controllerkeeper.NewMsgServerImpl(am.controllerKeeper))
		controllertypes.RegisterQueryServer(cfg.QueryServer(), am.controllerKeeper)
	}

	if am.hostKeeper != nil {
		hosttypes.RegisterQueryServer(cfg.QueryServer(), am.hostKeeper)
	}

	m := controllerkeeper.NewMigrator(am.controllerKeeper)
	if err := cfg.RegisterMigration(types.ModuleName, 1, m.AssertChannelCapabilityMigrations); err != nil {
		panic(fmt.Sprintf("failed to migrate interchainaccounts app from version 1 to 2: %v", err))
	}
}

// InitGenesis performs genesis initialization for the interchain accounts module.
// It returns no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Ctx, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState genesistypes.GenesisState
	types.ModuleCdc.MustUnmarshalJSON(data, &genesisState)

	if am.controllerKeeper != nil {
		controllerkeeper.InitGenesis(ctx, *am.controllerKeeper, genesisState.ControllerGenesisState)
	}

	if am.hostKeeper != nil {
		hostkeeper.InitGenesis(ctx, *am.hostKeeper, genesisState.HostGenesisState)
	}

	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the interchain accounts module
func (am AppModule) ExportGenesis(ctx sdk.Ctx) json.RawMessage {
	var (
		controllerGenesisState = genesistypes.DefaultControllerGenesis()
		hostGenesisState       = genesistypes.DefaultHostGenesis()
	)

	if am.controllerKeeper != nil {
		controllerGenesisState = controllerkeeper.ExportGenesis(ctx, *am.controllerKeeper)
	}

	if am.hostKeeper != nil {
		hostGenesisState = hostkeeper.ExportGenesis(ctx, *am.hostKeeper)
	}

	gs := genesistypes.NewGenesisState(controllerGenesisState, hostGenesisState)

	return types.ModuleCdc.MustMarshalJSON(gs)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 2 }

// BeginBlock implements the AppModule interface
func (am AppModule) BeginBlock(ctx sdk.Ctx, req abci.RequestBeginBlock) {
}

// EndBlock implements the AppModule interface
func (am AppModule) EndBlock(ctx sdk.Ctx, req abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

func (am AppModule) ConsensusParamsUpdate(ctx sdk.Ctx) *abci.ConsensusParams {
	return &abci.ConsensusParams{}
}

// NewHandler returns an sdk.Handler for the staking module.
func (am AppModule) NewHandler() sdk.Handler {
	return NewHandler(*am.controllerKeeper, *am.hostKeeper)
}

// NewQuerierHandler returns the staking module sdk.Querier.
func (am AppModule) NewQuerierHandler() sdk.Querier {
	return nil
}

// QuerierRoute returns the staking module's querier route name.
func (AppModule) QuerierRoute() string {
	return types.QuerierRoute
}

// RegisterCodec registers the staking module's types for the given codec.
func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
	types.RegisterCodec(cdc)
}

// Route returns the message routing key for the staking module.
func (AppModule) Route() string {
	return types.RouterKey
}
