package app

import (
	"encoding/json"
	"fmt"

	"github.com/tendermint/tendermint/libs/os"

	bam "github.com/vipernet-xyz/viper-network/baseapp"
	"github.com/vipernet-xyz/viper-network/codec"
	ibcExported "github.com/vipernet-xyz/viper-network/modules/core/exported"
	ibckeeper "github.com/vipernet-xyz/viper-network/modules/core/keeper"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/types/module"
	"github.com/vipernet-xyz/viper-network/x/authentication"
	capabilityKeeper "github.com/vipernet-xyz/viper-network/x/capability/keeper"
	capabilityTypes "github.com/vipernet-xyz/viper-network/x/capability/types"
	"github.com/vipernet-xyz/viper-network/x/governance"
	governanceKeeper "github.com/vipernet-xyz/viper-network/x/governance/keeper"
	governanceTypes "github.com/vipernet-xyz/viper-network/x/governance/types"
	providersKeeper "github.com/vipernet-xyz/viper-network/x/providers/keeper"
	providersTypes "github.com/vipernet-xyz/viper-network/x/providers/types"
	servicersKeeper "github.com/vipernet-xyz/viper-network/x/servicers/keeper"

	servicersTypes "github.com/vipernet-xyz/viper-network/x/servicers/types"
	transferKeeper "github.com/vipernet-xyz/viper-network/x/transfer/keeper"
	transferTypes "github.com/vipernet-xyz/viper-network/x/transfer/types"
	viperKeeper "github.com/vipernet-xyz/viper-network/x/vipernet/keeper"
	viperTypes "github.com/vipernet-xyz/viper-network/x/vipernet/types"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/rpc/client"
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
	Keys    map[string]*sdk.KVStoreKey
	Tkeys   map[string]*sdk.TransientStoreKey
	memKeys map[string]*sdk.MemoryStoreKey
	// Keepers for each module
	capabilityKeeper     *capabilityKeeper.Keeper
	accountKeeper        authentication.Keeper
	providersKeeper      providersKeeper.Keeper
	servicersKeeper      servicersKeeper.Keeper
	transferKeeper       transferKeeper.Keeper
	IBCKeeper            *ibckeeper.Keeper
	ScopedIBCKeeper      capabilityKeeper.ScopedKeeper
	ScopedTransferKeeper capabilityKeeper.ScopedKeeper
	viperKeeper          viperKeeper.Keeper
	governanceKeeper     governanceKeeper.Keeper
	// Module Manager
	mm *module.Manager
}

// new viper core base
func NewViperBaseApp(logger log.Logger, db db.DB, cache bool, iavlCacheSize int64, options ...func(*bam.BaseApp)) *ViperCoreApp {
	cdc = Codec()
	bam.SetABCILogging(GlobalConfig.ViperConfig.ABCILogging)
	// BaseApp handles interactions with Tendermint through the ABCI protocol
	bApp := bam.NewBaseApp(appName, logger, db, cache, iavlCacheSize, authentication.DefaultTxDecoder(cdc), cdc, options...)
	// set version of the baseapp
	bApp.SetAppVersion(AppVersion)
	// setup the key value store Keys
	k := sdk.NewKVStoreKeys(bam.MainStoreKey, capabilityTypes.StoreKey, authentication.StoreKey, servicersTypes.StoreKey, providersTypes.StoreKey, transferTypes.StoreKey, ibcExported.StoreKey, viperTypes.StoreKey, governance.StoreKey)
	// setup the transient store KeysibcExported.StoreKey, transferTypes.StoreKey)
	tkeys := sdk.NewTransientStoreKeys(capabilityTypes.TStoreKey, servicersTypes.TStoreKey, providersTypes.TStoreKey, transferTypes.TStoreKey, ibcExported.TStoreKey, viperTypes.TStoreKey, governance.TStoreKey)

	memkeys := sdk.NewMemoryStoreKeys(capabilityTypes.MemStoreKey)
	// add params Keys too
	// Create the application
	return &ViperCoreApp{
		BaseApp: bApp,
		cdc:     cdc,
		Keys:    k,
		Tkeys:   tkeys,
		memKeys: memkeys,
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
		modAccAddrs[authentication.NewModuleAddress(acc).String()] = true
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
	j, _ = Codec().MarshalJSONIndent(tmtypes.GenesisDoc{
		ChainID: chainID,
		ConsensusParams: &tmtypes.ConsensusParams{
			Block: tmtypes.BlockParams{
				MaxBytes:   8000000,
				MaxGas:     -1,
				TimeIotaMs: 1,
			},
			Evidence: tmtypes.EvidenceParams{
				MaxAge: 1000000,
			},
			Validator: tmtypes.ValidatorParams{
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
		capabilityTypes.ModuleName:      nil,
		authentication.FeeCollectorName: {authentication.Burner, authentication.Minter, authentication.Staking},
		servicersTypes.StakedPoolName:   {authentication.Burner, authentication.Minter, authentication.Staking},
		providersTypes.StakedPoolName:   {authentication.Burner, authentication.Minter, authentication.Staking},
		governanceTypes.DAOAccountName:  {authentication.Burner, authentication.Minter, authentication.Staking},
		servicersTypes.ModuleName:       {authentication.Burner, authentication.Minter, authentication.Staking},
		providersTypes.ModuleName:       nil,
		transferTypes.ModuleName:        {authentication.Burner, authentication.Minter},
	}
)

const (
	appName = "viper-network"
)
