package app

import (
	bam "github.com/vipernet-xyz/viper-network/baseapp"
	"github.com/vipernet-xyz/viper-network/codec"
	"github.com/vipernet-xyz/viper-network/crypto/keys"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/types/module"
	"github.com/vipernet-xyz/viper-network/x/authentication"
	"github.com/vipernet-xyz/viper-network/x/governance"
	govKeeper "github.com/vipernet-xyz/viper-network/x/governance/keeper"
	govTypes "github.com/vipernet-xyz/viper-network/x/governance/types"
	platforms "github.com/vipernet-xyz/viper-network/x/platforms"
	platformsKeeper "github.com/vipernet-xyz/viper-network/x/platforms/keeper"
	platformsTypes "github.com/vipernet-xyz/viper-network/x/platforms/types"
	"github.com/vipernet-xyz/viper-network/x/providers"
	providersKeeper "github.com/vipernet-xyz/viper-network/x/providers/keeper"
	providersTypes "github.com/vipernet-xyz/viper-network/x/providers/types"
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
	providersSubspace := sdk.NewSubspace(providersTypes.DefaultParamspace)
	platformsSubspace := sdk.NewSubspace(platformsTypes.DefaultParamspace)
	viperSubspace := sdk.NewSubspace(viperTypes.DefaultParamspace)
	// The AuthKeeper handles address -> account lookups
	app.accountKeeper = authentication.NewKeeper(
		app.cdc,
		app.Keys[authentication.StoreKey],
		authSubspace,
		moduleAccountPermissions,
	)
	// The providersKeeper keeper handles viper core providers
	app.providersKeeper = providersKeeper.NewKeeper(
		app.cdc,
		app.Keys[providersTypes.StoreKey],
		app.accountKeeper,
		providersSubspace,
		providersTypes.DefaultCodespace,
	)
	// The platforms keeper handles viper core applications
	app.platformsKeeper = platformsKeeper.NewKeeper(
		app.cdc,
		app.Keys[platformsTypes.StoreKey],
		app.providersKeeper,
		app.accountKeeper,
		app.viperKeeper,
		platformsSubspace,
		platformsTypes.DefaultCodespace,
	)
	// The main viper core
	app.viperKeeper = viperKeeper.NewKeeper(
		app.Keys[viperTypes.StoreKey],
		app.cdc,
		app.accountKeeper,
		app.providersKeeper,
		app.platformsKeeper,
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
		authSubspace, providersSubspace, platformsSubspace, viperSubspace,
	)
	// add the keybase to the viper core keeper
	app.viperKeeper.TmNode = tmClient
	// give viper keeper to providers module for easy cache clearing
	app.providersKeeper.ViperKeeper = app.viperKeeper
	app.platformsKeeper.ViperKeeper = app.viperKeeper
	// setup module manager
	app.mm = module.NewManager(
		authentication.NewPlatformModule(app.accountKeeper),
		providers.NewPlatformModule(app.providersKeeper),
		platforms.NewPlatformModule(app.platformsKeeper),
		viper.NewPlatformModule(app.viperKeeper),
		governance.NewPlatformModule(app.govKeeper),
	)
	// setup the order of begin and end blockers
	app.mm.SetOrderBeginBlockers(providersTypes.ModuleName, platformsTypes.ModuleName, viperTypes.ModuleName, govTypes.ModuleName)
	app.mm.SetOrderEndBlockers(providersTypes.ModuleName, platformsTypes.ModuleName, viperTypes.ModuleName, govTypes.ModuleName)
	// setup the order of Genesis
	app.mm.SetOrderInitGenesis(
		authentication.ModuleName,
		providersTypes.ModuleName,
		platformsTypes.ModuleName,
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
