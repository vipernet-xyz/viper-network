package app

import (
	bam "github.com/vipernet-xyz/viper-network/baseapp"
	"github.com/vipernet-xyz/viper-network/codec"
	"github.com/vipernet-xyz/viper-network/crypto/keys"
	ibc "github.com/vipernet-xyz/viper-network/modules/core"
	port "github.com/vipernet-xyz/viper-network/modules/core/05-port/types"
	ibcexported "github.com/vipernet-xyz/viper-network/modules/core/exported"
	ibckeeper "github.com/vipernet-xyz/viper-network/modules/core/keeper"
	ibctypes "github.com/vipernet-xyz/viper-network/modules/core/types"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/types/module"
	"github.com/vipernet-xyz/viper-network/x/authentication"
	"github.com/vipernet-xyz/viper-network/x/authentication/ante"
	"github.com/vipernet-xyz/viper-network/x/capability"
	capabilityKeeper "github.com/vipernet-xyz/viper-network/x/capability/keeper"
	capabilityTypes "github.com/vipernet-xyz/viper-network/x/capability/types"
	"github.com/vipernet-xyz/viper-network/x/governance"
	governanceKeeper "github.com/vipernet-xyz/viper-network/x/governance/keeper"
	governanceTypes "github.com/vipernet-xyz/viper-network/x/governance/types"
	providers "github.com/vipernet-xyz/viper-network/x/providers"
	providersKeeper "github.com/vipernet-xyz/viper-network/x/providers/keeper"
	providersTypes "github.com/vipernet-xyz/viper-network/x/providers/types"
	"github.com/vipernet-xyz/viper-network/x/servicers"
	servicersKeeper "github.com/vipernet-xyz/viper-network/x/servicers/keeper"
	servicersTypes "github.com/vipernet-xyz/viper-network/x/servicers/types"
	"github.com/vipernet-xyz/viper-network/x/transfer"
	transferKeeper "github.com/vipernet-xyz/viper-network/x/transfer/keeper"
	transferTypes "github.com/vipernet-xyz/viper-network/x/transfer/types"
	viper "github.com/vipernet-xyz/viper-network/x/vipernet"
	viperKeeper "github.com/vipernet-xyz/viper-network/x/vipernet/keeper"
	viperTypes "github.com/vipernet-xyz/viper-network/x/vipernet/types"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	cmn "github.com/tendermint/tendermint/libs/os"
	"github.com/tendermint/tendermint/rpc/client"
	dbm "github.com/tendermint/tm-db"
)

const (
	AppVersion = "RC-0.1.0"
)

// NewViperCoreApp is a constructor function for ViperCoreApp
func NewViperCoreApp(genState GenesisState, keybase keys.Keybase, tmClient client.Client, hostedChains *viperTypes.HostedBlockchains, logger log.Logger, db dbm.DB, cache bool, iavlCacheSize int64, baseAppOptions ...func(*bam.BaseApp)) *ViperCoreApp {
	app := NewViperBaseApp(logger, db, cache, iavlCacheSize, baseAppOptions...)
	// setup subspaces
	authSubspace := sdk.NewSubspace(authentication.DefaultParamspace)
	servicersSubspace := sdk.NewSubspace(servicersTypes.DefaultParamspace)
	providersSubspace := sdk.NewSubspace(providersTypes.DefaultParamspace)
	transferSubspace := sdk.NewSubspace(transferTypes.DefaultParamspace)
	viperSubspace := sdk.NewSubspace(viperTypes.DefaultParamspace)
	ibcSubspace := sdk.NewSubspace(ibctypes.DefaultParamspace)

	app.CapabilityKeeper = capabilityKeeper.NewKeeper(
		app.Bcdc,
		app.Keys[capabilityTypes.StoreKey],
		app.memKeys[capabilityTypes.MemStoreKey],
	)

	scopedIBCKeeper := app.CapabilityKeeper.ScopeToModule(ibcexported.ModuleName)
	scopedTransferKeeper := app.CapabilityKeeper.ScopeToModule(transferTypes.ModuleName)

	// The AuthKeeper handles address -> account lookups
	app.accountKeeper = authentication.NewKeeper(
		app.cdc,
		app.Keys[authentication.StoreKey],
		authSubspace,
		moduleAccountPermissions,
	)
	// The servicersKeeper keeper handles viper core servicers
	app.servicersKeeper = servicersKeeper.NewKeeper(
		app.cdc,
		app.Keys[servicersTypes.StoreKey],
		app.accountKeeper,
		servicersSubspace,
		servicersTypes.DefaultCodespace,
	)
	// The providers keeper handles viper core applications
	app.providersKeeper = providersKeeper.NewKeeper(
		app.cdc,
		app.Keys[providersTypes.StoreKey],
		app.servicersKeeper,
		app.accountKeeper,
		app.viperKeeper,
		providersSubspace,
		providersTypes.DefaultCodespace,
	)
	// The main viper core
	app.viperKeeper = viperKeeper.NewKeeper(
		app.Keys[viperTypes.StoreKey],
		app.cdc,
		app.accountKeeper,
		app.servicersKeeper,
		app.providersKeeper,
		hostedChains,
		viperSubspace,
	)
	// The governance keeper
	app.governanceKeeper = governanceKeeper.NewKeeper(
		app.cdc,
		app.Keys[viperTypes.StoreKey],
		app.Tkeys[viperTypes.StoreKey],
		governanceTypes.DefaultCodespace,
		app.accountKeeper,
		authSubspace, servicersSubspace, providersSubspace, viperSubspace,
	)

	app.IBCKeeper = ibckeeper.NewKeeper(
		app.Bcdc,
		app.Keys[ibcexported.StoreKey],
		ibcSubspace,
		app.StakingKeeper,
		app.UpgradeKeeper,
		scopedIBCKeeper,
	)

	app.transferKeeper = transferKeeper.NewKeeper(
		app.cdc,
		app.Keys[transferTypes.StoreKey],
		transferSubspace,
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		scopedTransferKeeper,
	)

	transferModule := transfer.NewAppModule(app.transferKeeper)
	transferIBCModule := transfer.NewIBCModule(app.transferKeeper)
	ibcRouter := port.NewRouter()
	ibcRouter.AddRoute(transferTypes.ModuleName, transferIBCModule)
	// Setting Router will finalize all routes by sealing router
	app.IBCKeeper.SetRouter(ibcRouter)
	// add the keybase to the viper core keeper
	app.viperKeeper.TmNode = tmClient
	// give viper keeper to servicers module for easy cache clearing
	app.servicersKeeper.ViperKeeper = app.viperKeeper
	app.providersKeeper.ViperKeeper = app.viperKeeper
	// setup module manager
	app.mm = module.NewManager(
		authentication.NewAppModule(app.accountKeeper),
		servicers.NewAppModule(app.servicersKeeper),
		providers.NewAppModule(app.providersKeeper),
		viper.NewAppModule(app.viperKeeper),
		capability.NewAppModule(*app.CapabilityKeeper),
		ibc.NewAppModule(app.IBCKeeper),
		governance.NewAppModule(app.governanceKeeper),
		transferModule,
	)
	// setup the order of begin and end blockers
	app.mm.SetOrderBeginBlockers(servicersTypes.ModuleName, providersTypes.ModuleName, viperTypes.ModuleName, governanceTypes.ModuleName, ibcexported.ModuleName)
	app.mm.SetOrderEndBlockers(servicersTypes.ModuleName, providersTypes.ModuleName, viperTypes.ModuleName, governanceTypes.ModuleName, ibcexported.ModuleName)
	// setup the order of Genesis
	app.mm.SetOrderInitGenesis(
		capabilityTypes.ModuleName,
		authentication.ModuleName,
		servicersTypes.ModuleName,
		providersTypes.ModuleName,
		viperTypes.ModuleName,
		governance.ModuleName,
		ibcexported.ModuleName,
		transferTypes.ModuleName,
	)
	// register all module routes and module queriers
	app.mm.RegisterRoutes(app.Router(), app.QueryRouter())
	// The initChainer handles translating the genesis.json file into initial state for the network
	if genState == nil {
		app.SetInitChainer(app.InitChainer)
	} else {
		app.SetInitChainer(app.InitChainerWithGenesis)
	}
	app.SetAnteHandler(ante.NewAnteHandler(app.accountKeeper))
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)
	// initialize stores
	app.MountKVStores(app.Keys)
	app.MountTransientStores(app.Tkeys)
	app.SetAppVersion(AppVersion)
	// load the latest persistent version of the store
	err := app.LoadLatestVersion(app.Keys[bam.MainStoreKey])
	if err != nil {
		cmn.Exit(err.Error())
	}
	ctx := sdk.NewContext(app.Store(), abci.Header{}, false, app.Logger()).WithBlockStore(app.BlockStore())
	if upgrade := app.governanceKeeper.GetUpgrade(ctx); upgrade.Height != 0 {
		codec.UpgradeHeight = upgrade.Height
		codec.OldUpgradeHeight = upgrade.OldUpgradeHeight
		codec.UpgradeFeatureMap = codec.SliceToExistingMap(upgrade.GetFeatures(), codec.UpgradeFeatureMap)
	}
	return app
}
