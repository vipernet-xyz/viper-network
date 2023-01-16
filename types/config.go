package types

import (
	"sync"
	"time"

	"github.com/tendermint/tendermint/config"
	db "github.com/tendermint/tm-db"
)

// TmConfig is the structure that holds the SDK configuration parameters.
// This could be used to initialize certain configuration parameters for the SDK.
type SDKConfig struct {
	mtx             sync.RWMutex
	sealed          bool
	txEncoder       TxEncoder
	addressVerifier func([]byte) error
}

type ViperConfig struct {
	DataDir                  string `json:"data_dir"`
	GenesisName              string `json:"genesis_file"`
	ChainsName               string `json:"chains_name"`
	EvidenceDBName           string `json:"evidence_db_name"`
	TendermintURI            string `json:"tendermint_uri"`
	KeybaseName              string `json:"keybase_name"`
	RPCPort                  string `json:"rpc_port"`
	ClientBlockSyncAllowance int    `json:"client_block_sync_allowance"`
	MaxEvidenceCacheEntires  int    `json:"max_evidence_cache_entries"`
	MaxSessionCacheEntries   int    `json:"max_session_cache_entries"`
	JSONSortRelayResponses   bool   `json:"json_sort_relay_responses"`
	RemoteCLIURL             string `json:"remote_cli_url"`
	UserAgent                string `json:"user_agent"`
	ValidatorCacheSize       int64  `json:"validator_cache_size"`
	PlatformCacheSize        int64  `json:"application_cache_size"`
	RPCTimeout               int64  `json:"rpc_timeout"`
	PrometheusAddr           string `json:"viper_prometheus_port"`
	PrometheusMaxOpenfiles   int    `json:"prometheus_max_open_files"`
	MaxClaimAgeForProofRetry int    `json:"max_claim_age_for_proof_retry"`
	ProofPrevalidation       bool   `json:"proof_prevalidation"`
	CtxCacheSize             int    `json:"ctx_cache_size"`
	ABCILogging              bool   `json:"abci_logging"`
	RelayErrors              bool   `json:"show_relay_errors"`
	DisableTxEvents          bool   `json:"disable_tx_events"`
	Cache                    bool   `json:"-"`
	IavlCacheSize            int64  `json:"iavl_cache_size"`
	ChainsHotReload          bool   `json:"chains_hot_reload"`
	GenerateTokenOnStart     bool   `json:"generate_token_on_start"`
}

type Config struct {
	TendermintConfig config.Config `json:"tendermint_config"`
	ViperConfig      ViperConfig   `json:"viper_config"`
}

type AuthToken struct {
	Value  string
	Issued time.Time
}

const (
	DefaultDDName                      = ".viper"
	DefaultKeybaseName                 = "viper-keybase"
	DefaultPVKName                     = "priv_val_key.json"
	DefaultPVSName                     = "priv_val_state.json"
	DefaultNKName                      = "node_key.json"
	DefaultChainsName                  = "chains.json"
	DefaultGenesisName                 = "genesis.json"
	DefaultRPCPort                     = "8081"
	DefaultEvidenceDBName              = "viper_evidence"
	DefaultTMURI                       = "tcp://localhost:26657"
	DefaultMaxSessionCacheEntries      = 500
	DefaultMaxEvidenceCacheEntries     = 500
	DefaultListenAddr                  = "tcp://0.0.0.0:"
	DefaultClientBlockSyncAllowance    = 10
	DefaultJSONSortRelayResponses      = true
	DefaultTxIndexer                   = "kv"
	DefaultRPCDisableTransactionEvents = true
	DefaultTxIndexTags                 = "tx.hash,tx.height,message.sender,transfer.recipient"
	ConfigDirName                      = "config"
	ConfigFileName                     = "config.json"
	PlatformDBName                     = "application"
	TransactionIndexerDBName           = "txindexer"
	PlaceholderHash                    = "0001"
	PlaceholderURL                     = "http://127.0.0.1:8081"
	PlaceholderServiceURL              = PlaceholderURL
	DefaultRemoteCLIURL                = "http://localhost:8081"
	DefaultUserAgent                   = ""
	DefaultValidatorCacheSize          = 40000
	DefaultPlatformCacheSize           = DefaultValidatorCacheSize / 4
	DefaultViperPrometheusListenAddr   = "8083"
	DefaultPrometheusMaxOpenFile       = 3
	DefaultRPCTimeout                  = 30000
	DefaultMaxClaimProofRetryAge       = 32
	DefaultProofPrevalidation          = false
	DefaultCtxCacheSize                = 20
	DefaultABCILogging                 = false
	DefaultRelayErrors                 = true
	AuthFileName                       = "authentication.json"
	DefaultIavlCacheSize               = 5000000
	DefaultChainHotReload              = false
	DefaultGenerateTokenOnStart        = true
)

func DefaultConfig(dataDir string) Config {
	c := Config{
		TendermintConfig: *config.DefaultConfig(),
		ViperConfig: ViperConfig{
			DataDir:                  dataDir,
			GenesisName:              DefaultGenesisName,
			ChainsName:               DefaultChainsName,
			EvidenceDBName:           DefaultEvidenceDBName,
			TendermintURI:            DefaultTMURI,
			KeybaseName:              DefaultKeybaseName,
			RPCPort:                  DefaultRPCPort,
			ClientBlockSyncAllowance: DefaultClientBlockSyncAllowance,
			MaxEvidenceCacheEntires:  DefaultMaxEvidenceCacheEntries,
			MaxSessionCacheEntries:   DefaultMaxSessionCacheEntries,
			JSONSortRelayResponses:   DefaultJSONSortRelayResponses,
			RemoteCLIURL:             DefaultRemoteCLIURL,
			UserAgent:                DefaultUserAgent,
			ValidatorCacheSize:       DefaultValidatorCacheSize,
			PlatformCacheSize:        DefaultPlatformCacheSize,
			RPCTimeout:               DefaultRPCTimeout,
			PrometheusAddr:           DefaultViperPrometheusListenAddr,
			PrometheusMaxOpenfiles:   DefaultPrometheusMaxOpenFile,
			MaxClaimAgeForProofRetry: DefaultMaxClaimProofRetryAge,
			ProofPrevalidation:       DefaultProofPrevalidation,
			CtxCacheSize:             DefaultCtxCacheSize,
			ABCILogging:              DefaultABCILogging,
			RelayErrors:              DefaultRelayErrors,
			DisableTxEvents:          DefaultRPCDisableTransactionEvents,
			IavlCacheSize:            DefaultIavlCacheSize,
			ChainsHotReload:          DefaultChainHotReload,
			GenerateTokenOnStart:     DefaultGenerateTokenOnStart,
		},
	}
	c.TendermintConfig.LevelDBOptions = config.DefaultLevelDBOpts()
	c.TendermintConfig.SetRoot(dataDir)
	c.TendermintConfig.NodeKey = DefaultNKName
	c.TendermintConfig.PrivValidatorKey = DefaultPVKName
	c.TendermintConfig.PrivValidatorState = DefaultPVSName
	c.TendermintConfig.P2P.AddrBookStrict = false
	c.TendermintConfig.P2P.MaxNumInboundPeers = 14
	c.TendermintConfig.P2P.MaxNumOutboundPeers = 7
	c.TendermintConfig.LogLevel = "*:info, *:error"
	c.TendermintConfig.TxIndex.Indexer = DefaultTxIndexer
	c.TendermintConfig.TxIndex.IndexKeys = DefaultTxIndexTags
	c.TendermintConfig.DBBackend = string(db.GoLevelDBBackend)
	c.TendermintConfig.RPC.GRPCMaxOpenConnections = 2500
	c.TendermintConfig.RPC.MaxOpenConnections = 2500
	c.TendermintConfig.Mempool.Size = 9000
	c.TendermintConfig.Mempool.CacheSize = 9000
	c.TendermintConfig.FastSync = &config.FastSyncConfig{
		Version: "v1",
	}
	DefaultViperConsensusConfig(c.TendermintConfig.Consensus)
	c.TendermintConfig.P2P.AllowDuplicateIP = true
	return c
}

func DefaultViperConsensusConfig(cconfig *config.ConsensusConfig) {
	cconfig.TimeoutPropose = 120000000000
	cconfig.TimeoutProposeDelta = 10000000000
	cconfig.TimeoutPrevote = 60000000000
	cconfig.TimeoutPrevoteDelta = 10000000000
	cconfig.TimeoutPrecommit = 60000000000
	cconfig.TimeoutPrecommitDelta = 10000000000
	cconfig.TimeoutCommit = 780000000000
	cconfig.SkipTimeoutCommit = false
	cconfig.CreateEmptyBlocks = true
	cconfig.CreateEmptyBlocksInterval = 900000000000
	cconfig.PeerGossipSleepDuration = 30000000000
	cconfig.PeerQueryMaj23SleepDuration = 20000000000
}

func DefaultTestingViperConfig() Config {
	c := DefaultConfig("data")
	c.ViperConfig.MaxClaimAgeForProofRetry = 1000
	t := config.TestConfig()
	t.LevelDBOptions = config.DefaultLevelDBOpts()
	return Config{
		TendermintConfig: *t,
		ViperConfig:      c.ViperConfig,
	}
}

var (
	// Initializing an instance of TmConfig
	sdkConfig = &SDKConfig{
		sealed:    false,
		txEncoder: nil,
	}
)

// GetConfig returns the config instance for the SDK.
func GetConfig() *SDKConfig {
	return sdkConfig
}

func (config *SDKConfig) assertNotSealed() {
	config.mtx.Lock()
	defer config.mtx.Unlock()

	if config.sealed {
		panic("TmConfig is sealed")
	}
}

// SetTxEncoder builds the TmConfig with TxEncoder used to marshal StdTx to bytes
func (config *SDKConfig) SetTxEncoder(encoder TxEncoder) {
	config.assertNotSealed()
	config.txEncoder = encoder
}

// SetAddressVerifier builds the TmConfig with the provided function for verifying that Addresses
// have the correct format
func (config *SDKConfig) SetAddressVerifier(addressVerifier func([]byte) error) {
	config.assertNotSealed()
	config.addressVerifier = addressVerifier
}

// Set the BIP-0044 CoinType code on the config
func (config *SDKConfig) SetCoinType(coinType uint32) {
	config.assertNotSealed()
}

// Seal seals the config such that the config state could not be modified further
func (config *SDKConfig) Seal() *SDKConfig {
	config.mtx.Lock()
	defer config.mtx.Unlock()

	config.sealed = true
	return config
}

// GetTxEncoder return function to encode transactions
func (config *SDKConfig) GetTxEncoder() TxEncoder {
	return config.txEncoder
}

// GetAddressVerifier returns the function to verify that Addresses have the correct format
func (config *SDKConfig) GetAddressVerifier() func([]byte) error {
	return config.addressVerifier
}
