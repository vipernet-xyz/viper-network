package app

import (
	"encoding/json"
	"fmt"

	"github.com/tendermint/tendermint/libs/os"

	bam "github.com/vipernet-xyz/viper-network/baseapp"
	"github.com/vipernet-xyz/viper-network/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/types/module"
	appsKeeper "github.com/vipernet-xyz/viper-network/x/apps/keeper"
	appsTypes "github.com/vipernet-xyz/viper-network/x/apps/types"
	"github.com/vipernet-xyz/viper-network/x/auth"
	"github.com/vipernet-xyz/viper-network/x/gov"
	govKeeper "github.com/vipernet-xyz/viper-network/x/gov/keeper"
	govTypes "github.com/vipernet-xyz/viper-network/x/gov/types"
	nodesKeeper "github.com/vipernet-xyz/viper-network/x/nodes/keeper"
	nodesTypes "github.com/vipernet-xyz/viper-network/x/nodes/types"
	viperKeeper "github.com/vipernet-xyz/viper-network/x/vipercore/keeper"
	viperTypes "github.com/vipernet-xyz/viper-network/x/vipercore/types"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/rpc/client"
	"github.com/tendermint/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
	db "github.com/tendermint/tm-db"
)

// viper core is an extension of baseapp
type ViperCoreApp struct {
	// extends baseapp
	*bam.BaseApp
	// the codec (uses amino)
	cdc *codec.Codec
	// Keys to access the substores
	Keys  map[string]*sdk.KVStoreKey
	Tkeys map[string]*sdk.TransientStoreKey
	// Keepers for each module
	accountKeeper auth.Keeper
	appsKeeper    appsKeeper.Keeper
	nodesKeeper   nodesKeeper.Keeper
	govKeeper     govKeeper.Keeper
	viperKeeper   viperKeeper.Keeper
	// Module Manager
	mm *module.Manager
}

// new viper core base
func NewViperBaseApp(logger log.Logger, db db.DB, cache bool, iavlCacheSize int64, options ...func(*bam.BaseApp)) *ViperCoreApp {
	cdc = Codec()
	bam.SetABCILogging(GlobalConfig.ViperConfig.ABCILogging)
	// BaseApp handles interactions with Tendermint through the ABCI protocol
	bApp := bam.NewBaseApp(appName, logger, db, cache, iavlCacheSize, auth.DefaultTxDecoder(cdc), cdc, options...)
	// set version of the baseapp
	bApp.SetAppVersion(AppVersion)
	// setup the key value store Keys
	k := sdk.NewKVStoreKeys(bam.MainStoreKey, auth.StoreKey, nodesTypes.StoreKey, appsTypes.StoreKey, gov.StoreKey, viperTypes.StoreKey)
	// setup the transient store Keys
	tkeys := sdk.NewTransientStoreKeys(nodesTypes.TStoreKey, appsTypes.TStoreKey, viperTypes.TStoreKey, gov.TStoreKey)
	// add params Keys too
	// Create the application
	return &ViperCoreApp{
		BaseApp: bApp,
		cdc:     cdc,
		Keys:    k,
		Tkeys:   tkeys,
	}
}

// inits from genesis
func (app *ViperCoreApp) InitChainer(ctx sdk.Ctx, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState GenesisState
	switch GlobalGenesisType {
	case MainnetGenesisType:
		genesisState = GenesisStateFromJson(mainnetGenesis)
	case TestnetGenesisType:
		genesisState = GenesisStateFromJson(testnetGenesis)
	default:
		genesisState = GenesisStateFromFile(cdc, GlobalConfig.ViperConfig.DataDir+FS+sdk.ConfigDirName+FS+GlobalConfig.ViperConfig.GenesisName)
	}
	return app.mm.InitGenesis(ctx, genesisState)
}

var GenState GenesisState

// inits from genesis
func (app *ViperCoreApp) InitChainerWithGenesis(ctx sdk.Ctx, req abci.RequestInitChain) abci.ResponseInitChain {
	return app.mm.InitGenesis(ctx, GenState)
}

// setups all of the begin blockers for each module
func (app *ViperCoreApp) BeginBlocker(ctx sdk.Ctx, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return app.mm.BeginBlock(ctx, req)
}

// setups all of the end blockers for each module
func (app *ViperCoreApp) EndBlocker(ctx sdk.Ctx, req abci.RequestEndBlock) abci.ResponseEndBlock {
	return app.mm.EndBlock(ctx, req)
}

// ModuleAccountAddrs returns all the pcInstance's module account addresses.
func (app *ViperCoreApp) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range moduleAccountPermissions {
		modAccAddrs[auth.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

type GenesisState map[string]json.RawMessage

func GenesisStateFromFile(cdc *codec.Codec, genFile string) GenesisState {
	if !os.FileExists(genFile) {
		panic(fmt.Errorf("%s does not exist, run `init` first", genFile))
	}
	genDoc := GenesisFileToGenDoc(genFile)
	return GenesisStateFromGenDoc(cdc, *genDoc)
}

func GenesisFileToGenDoc(genFile string) *tmtypes.GenesisDoc {
	if !os.FileExists(genFile) {
		panic(fmt.Errorf("%s does not exist, run `init` first", genFile))
	}
	genDoc, err := tmtypes.GenesisDocFromFile(genFile)
	if err != nil {
		panic(err)
	}
	return genDoc
}

func GenesisStateFromGenDoc(cdc *codec.Codec, genDoc tmtypes.GenesisDoc) (genesisState map[string]json.RawMessage) {
	if err := cdc.UnmarshalJSON(genDoc.AppState, &genesisState); err != nil {
		panic(err)
	}
	return genesisState
}

// exports the app state to json
func (app *ViperCoreApp) ExportAppState(height int64, forZeroHeight bool, jailWhiteList []string) (appState json.RawMessage, err error) {
	// as if they could withdraw from the start of the next block
	ctx, err := app.NewContext(height)
	if err != nil {
		return nil, err
	}
	genState := app.mm.ExportGenesis(ctx)
	appState, err = Codec().MarshalJSONIndent(genState, "", "    ")
	if err != nil {
		return nil, err
	}
	return appState, nil
}

func (app *ViperCoreApp) ExportState(height int64, chainID string) (string, error) {
	j, err := app.ExportAppState(height, false, nil)
	if err != nil {
		return "", err
	}
	if chainID == "" {
		chainID = "<Input New ChainID>"
	}
	j, _ = Codec().MarshalJSONIndent(types.GenesisDoc{
		ChainID: chainID,
		ConsensusParams: &types.ConsensusParams{
			Block: types.BlockParams{
				MaxBytes:   4000000,
				MaxGas:     -1,
				TimeIotaMs: 1,
			},
			Evidence: types.EvidenceParams{
				MaxAge: 1000000,
			},
			Validator: types.ValidatorParams{
				PubKeyTypes: []string{"ed25519"},
			},
		},
		Validators: nil,
		AppHash:    nil,
		AppState:   j,
	}, "", "    ")
	return SortJSON(j), err
}

func (app *ViperCoreApp) NewContext(height int64) (sdk.Ctx, error) {
	store := app.Store()
	blockStore := app.BlockStore()
	ctx := sdk.NewContext(store, abci.Header{}, false, app.Logger()).WithBlockStore(blockStore)
	return ctx.PrevCtx(height)
}

func (app *ViperCoreApp) GetClient() client.Client {
	return app.viperKeeper.TmNode
}

var (
	// module account permissions
	moduleAccountPermissions = map[string][]string{
		auth.FeeCollectorName:     {auth.Burner, auth.Minter, auth.Staking},
		nodesTypes.StakedPoolName: {auth.Burner, auth.Minter, auth.Staking},
		appsTypes.StakedPoolName:  {auth.Burner, auth.Minter, auth.Staking},
		govTypes.DAOAccountName:   {auth.Burner, auth.Minter, auth.Staking},
		nodesTypes.ModuleName:     {auth.Burner, auth.Minter, auth.Staking},
		appsTypes.ModuleName:      nil,
	}
)

const (
	appName = "viper-core"
)
