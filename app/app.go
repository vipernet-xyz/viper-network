package app

import (
	bam "github.com/vipernet-xyz/viper-network/baseapp"
	"github.com/vipernet-xyz/viper-network/codec"
	"github.com/vipernet-xyz/viper-network/crypto/keys"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/types/module"
	"github.com/vipernet-xyz/viper-network/x/authentication"
	"github.com/vipernet-xyz/viper-network/x/authentication/ante"
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

	/*"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/capability"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"*/
	transferKeeper "github.com/vipernet-xyz/viper-network/x/transfer/keeper"
	transferTypes "github.com/vipernet-xyz/viper-network/x/transfer/types"

	/*ibc "github.com/cosmos/ibc-go/v3/modules/core"
	ibcclient "github.com/cosmos/ibc-go/v3/modules/core/02-client"
	ibcclientclient "github.com/cosmos/ibc-go/v3/modules/core/02-client/client"
	ibcclienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	ibcporttypes "github.com/cosmos/ibc-go/v3/modules/core/05-port/types"
	ibchost "github.com/cosmos/ibc-go/v3/modules/core/24-host"
	ibckeeper "github.com/cosmos/ibc-go/v3/modules/core/keeper"*/
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

	app.transferKeeper = transferKeeper.NewKeeper(
		app.cdc,
		app.Keys[transferTypes.StoreKey],
		transferSubspace,
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		app.scopedTransferKeeper,
	)

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
		transfer.NewAppModule(app.transferKeeper),
		governance.NewAppModule(app.governanceKeeper),
	)
	// setup the order of begin and end blockers
	app.mm.SetOrderBeginBlockers(servicersTypes.ModuleName, providersTypes.ModuleName, viperTypes.ModuleName, governanceTypes.ModuleName)
	app.mm.SetOrderEndBlockers(servicersTypes.ModuleName, providersTypes.ModuleName, viperTypes.ModuleName, governanceTypes.ModuleName)
	// setup the order of Genesis
	app.mm.SetOrderInitGenesis(
		authentication.ModuleName,
		servicersTypes.ModuleName,
		providersTypes.ModuleName,
		transferTypes.ModuleName,
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
