package app

import (
	"io"
	"os"
	"path/filepath"

	"github.com/vipernet-xyz/viper-network/codec"

	sdk "github.com/vipernet-xyz/viper-network/types"

	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/p2p"
	pvm "github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/proxy"
	dbm "github.com/tendermint/tm-db"
)

type AppCreator func(log.Logger, dbm.DB, io.Writer) *ViperCoreApp

func NewClient(c config, creator AppCreator) (*node.Node, *ViperCoreApp, error) {
	// setup the database
	appDB, err := OpenPlatformDB(GlobalConfig)
	if err != nil {
		return nil, nil, err
	}
	// setup the transaction indexer
	txDB, err := OpenTxIndexerDB(GlobalConfig)
	if err != nil {
		return nil, nil, err
	}
	transactionIndexer := sdk.NewTransactionIndexer(txDB)
	// open the tracewriter
	traceWriter, err := openTraceWriter(c.TraceWriter)
	if err != nil {
		return nil, nil, err
	}
	// load the node key
	nodeKey, err := p2p.LoadOrGenNodeKey(c.TmConfig.NodeKeyFile())
	if err != nil {
		return nil, nil, err
	}
	// upgrade the privVal file
	app := creator(c.Logger, appDB, traceWriter)
	VCA = app
	// create & start tendermint node
	tmNode, err := node.NewNode(app,
		c.TmConfig,
		codec.GetCodecUpgradeHeight(),
		pvm.LoadOrGenFilePV(c.TmConfig.PrivValidatorKeyFile(), c.TmConfig.PrivValidatorStateFile()),
		nodeKey,
		proxy.NewLocalClientCreator(app),
		transactionIndexer,
		node.DefaultGenesisDocProviderFunc(c.TmConfig),
		node.DefaultDBProvider,
		node.DefaultMetricsProvider(c.TmConfig.Instrumentation),
		c.Logger.With("module", "node"),
	)
	if err != nil {
		return nil, nil, err
	}

	return tmNode, app, nil
}

func OpenPlatformDB(config sdk.Config) (dbm.DB, error) {
	dataDir := filepath.Join(config.TendermintConfig.RootDir, GlobalConfig.TendermintConfig.DBPath)
	return sdk.NewLevelDB(sdk.PlatformDBName, dataDir, config.TendermintConfig.LevelDBOptions.ToGoLevelDBOpts())
}

func OpenTxIndexerDB(config sdk.Config) (dbm.DB, error) {
	dataDir := filepath.Join(config.TendermintConfig.RootDir, GlobalConfig.TendermintConfig.DBPath)
	return sdk.NewLevelDB(sdk.TransactionIndexerDBName, dataDir, config.TendermintConfig.LevelDBOptions.ToGoLevelDBOpts())
}

func openTraceWriter(traceWriterFile string) (w io.Writer, err error) {
	if traceWriterFile != "" {
		w, err = os.OpenFile(
			traceWriterFile,
			os.O_WRONLY|os.O_APPEND|os.O_CREATE,
			0666,
		)
		return
	}
	return
}

//// upgradePrivVal converts old priv_validator.json file (prior to Tendermint 0.28)
//// to the new priv_validator_key.json and priv_validator_state.json files.
//func upgradePrivVal(config *cfg.Config) {
//	if _, err := os.Stat(config.OldPrivValidatorFile()); !os.IsNotExist(err) {
//		if oldFilePV, err := pvm.LoadOldFilePV(config.OldPrivValidatorFile()); err == nil {
//			oldFilePV.Upgrade(config.PrivValidatorKeyFile(), config.PrivValidatorStateFile())
//		}
//	}
//}

type config struct {
	TmConfig    *cfg.Config
	Logger      log.Logger
	TraceWriter string
}

func modifyPrivValidatorsFile(config *cfg.Config, rollbackHeight int64) error {
	var sig []byte
	filePv := pvm.LoadOrGenFilePV(config.PrivValidatorKeyFile(), config.PrivValidatorStateFile())
	filePv.LastSignState.Height = rollbackHeight
	filePv.LastSignState.Round = 0
	filePv.LastSignState.Step = 0
	filePv.LastSignState.Signature = sig
	filePv.LastSignState.SignBytes = nil
	filePv.Save()
	return nil
}
