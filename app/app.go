package app

import (
	bam "github.com/vipernet-xyz/viper-network/baseapp"
	"github.com/vipernet-xyz/viper-network/codec"
	"github.com/vipernet-xyz/viper-network/crypto/keys"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/types/module"
	apps "github.com/vipernet-xyz/viper-network/x/apps"
	appsKeeper "github.com/vipernet-xyz/viper-network/x/apps/keeper"
	appsTypes "github.com/vipernet-xyz/viper-network/x/apps/types"
	"github.com/vipernet-xyz/viper-network/x/authentication"
	"github.com/vipernet-xyz/viper-network/x/governance"
	govKeeper "github.com/vipernet-xyz/viper-network/x/governance/keeper"
	govTypes "github.com/vipernet-xyz/viper-network/x/governance/types"
	"github.com/vipernet-xyz/viper-network/x/providers"
	nodesKeeper "github.com/vipernet-xyz/viper-network/x/providers/keeper"
	nodesTypes "github.com/vipernet-xyz/viper-network/x/providers/types"
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
	nodesSubspace := sdk.NewSubspace(nodesTypes.DefaultParamspace)
	appsSubspace := sdk.NewSubspace(appsTypes.DefaultParamspace)
	viperSubspace := sdk.NewSubspace(viperTypes.DefaultParamspace)
	// The AuthKeeper handles address -> account lookups
	app.accountKeeper = authentication.NewKeeper(
		app.cdc,
		app.Keys[authentication.StoreKey],
		authSubspace,
		moduleAccountPermissions,
	)
	// The nodesKeeper keeper handles viper core providers
	app.nodesKeeper = nodesKeeper.NewKeeper(
		app.cdc,
		app.Keys[nodesTypes.StoreKey],
		app.accountKeeper,
		nodesSubspace,
		nodesTypes.DefaultCodespace,
	)
	// The apps keeper handles viper core applications
	app.appsKeeper = appsKeeper.NewKeeper(
		app.cdc,
		app.Keys[appsTypes.StoreKey],
		app.nodesKeeper,
		app.accountKeeper,
		app.viperKeeper,
		appsSubspace,
		appsTypes.DefaultCodespace,
	)
	// The main viper core
	app.viperKeeper = viperKeeper.NewKeeper(
		app.Keys[viperTypes.StoreKey],
		app.cdc,
		app.accountKeeper,
		app.nodesKeeper,
		app.appsKeeper,
		hostedChains,
		viperSubspace,
	)
	// The governance keeper
	app.govKeeper = govKeeper.NewKeeper(
		app.cdc,
		app.Keys[viperTypes.StoreKey],
		app.Tkeys[viperTypes.StoreKey],
		govTypes.DefaultCodespace,
		app.accountKeeper,
		authSubspace, nodesSubspace, appsSubspace, viperSubspace,
	)
	// add the keybase to the viper core keeper
	app.viperKeeper.TmNode = tmClient
	// give viper keeper to providers module for easy cache clearing
	app.nodesKeeper.ViperKeeper = app.viperKeeper
	app.appsKeeper.ViperKeeper = app.viperKeeper
	// setup module manager
	app.mm = module.NewManager(
		authentication.NewAppModule(app.accountKeeper),
		providers.NewAppModule(app.nodesKeeper),
		apps.NewAppModule(app.appsKeeper),
		viper.NewAppModule(app.viperKeeper),
		governance.NewAppModule(app.govKeeper),
	)
	// setup the order of begin and end blockers
	app.mm.SetOrderBeginBlockers(nodesTypes.ModuleName, appsTypes.ModuleName, viperTypes.ModuleName, govTypes.ModuleName)
	app.mm.SetOrderEndBlockers(nodesTypes.ModuleName, appsTypes.ModuleName, viperTypes.ModuleName, govTypes.ModuleName)
	// setup the order of Genesis
	app.mm.SetOrderInitGenesis(
		authentication.ModuleName,
		nodesTypes.ModuleName,
		appsTypes.ModuleName,
		viperTypes.ModuleName,
		governance.ModuleName,
	)
	// register all module routes and module queriers
	app.mm.RegisterRoutes(app.Router(), app.QueryRouter())
	// The initChainer handles translating the genesis.json file into initial state for the network
	if genState == nil {
		app.SetInitChainer(app.InitChainer)
	} else {
		app.SetInitChainer(app.InitChainerWithGenesis)
	}
	app.SetAnteHandler(authentication.NewAnteHandler(app.accountKeeper))
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
	if upgrade := app.govKeeper.GetUpgrade(ctx); upgrade.Height != 0 {
		codec.UpgradeHeight = upgrade.Height
		codec.OldUpgradeHeight = upgrade.OldUpgradeHeight
		codec.UpgradeFeatureMap = codec.SliceToExistingMap(upgrade.GetFeatures(), codec.UpgradeFeatureMap)
	}
	return app
}
