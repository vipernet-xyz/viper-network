package app

/*
import (
	"context"
	"io"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/tendermint/tendermint/privval"

	types2 "github.com/vipernet-xyz/viper-network/codec/types"
	viperTypes "github.com/vipernet-xyz/viper-network/x/viper-main/types"

	"github.com/tendermint/tendermint/rpc/client/http"
	"github.com/tendermint/tendermint/rpc/client/local"

	bam "github.com/vipernet-xyz/viper-network/baseapp"
	"github.com/vipernet-xyz/viper-network/codec"
	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	"github.com/vipernet-xyz/viper-network/crypto/keys"
	"github.com/vipernet-xyz/viper-network/store"

	// sdk "github.com/vipernet-xyz/viper-network/types"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/types/module"
	"github.com/vipernet-xyz/viper-network/x/authentication"
	"github.com/vipernet-xyz/viper-network/x/governance"
	govTypes "github.com/vipernet-xyz/viper-network/x/governance/types"
	requestors "github.com/vipernet-xyz/viper-network/x/requestors"
	requestorsTypes "github.com/vipernet-xyz/viper-network/x/requestors/types"
	"github.com/vipernet-xyz/viper-network/x/servicers"
	servicersTypes "github.com/vipernet-xyz/viper-network/x/servicers/types"
	viper "github.com/vipernet-xyz/viper-network/x/viper-main"

	"github.com/stretchr/testify/assert"
	tmCfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/proxy"
	"github.com/tendermint/tendermint/rpc/client"
	cTypes "github.com/tendermint/tendermint/rpc/core/types"
	"github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"
)

const (
	dummyChainsHash = "0001"
)

type upgrades struct {
	codecUpgrade codecUpgrade
	Upgrade      upgrade
}
type upgrade struct {
	height int64
}
type codecUpgrade struct {
	upgradeMod bool
	height     int64
}

// NewInMemoryTendermintNodeAmino will create a TM node with only one validator. LeanViper is disabled.
func NewInMemoryTendermintNodeAmino(t *testing.T, genesisState []byte) (tendermintNode *node.Node, keybase keys.Keybase, cleanup func()) {
	return NewInMemoryTendermintNodeAminoWithValidators(t, genesisState, nil)
}

// NewInMemoryTendermintNodeAminoWithValidators will create a TM node with 'n' "validators".
// If "validators" is nil, LeanVIPR is disabled
func NewInMemoryTendermintNodeAminoWithValidators(t *testing.T, genesisState []byte, validators []crypto.PrivateKey) (tendermintNode *node.Node, keybase keys.Keybase, cleanup func()) {
	// create the in memory tendermint node and keybase
	tendermintNode, keybase = inMemTendermintNodeWithValidators(genesisState, validators)
	// test assertions
	if tendermintNode == nil {
		panic("tendermintNode should not be nil")
	}
	if keybase == nil {
		panic("should not be nil")
	}
	assert.NotNil(t, tendermintNode)
	assert.NotNil(t, keybase)

	// init cache in memory
	defaultConfig := sdk.DefaultTestingViperConfig()
	if validators != nil {
		defaultConfig.ViperConfig.LeanViper = true
	}

	viperTypes.InitConfig(&viperTypes.HostedBlockchains{
		M: make(map[string]viperTypes.HostedBlockchain),
	}, nil, tendermintNode.Logger, defaultConfig)
	// start the in memory node
	err := tendermintNode.Start()
	if err != nil {
		panic(err)
	}
	// assert that it is not nil
	assert.Nil(t, err)
	// provide cleanup function
	cleanup = func() {
		err = tendermintNode.Stop()
		if err != nil {
			panic(err)
		}
		viperTypes.CleanViperNodes()
		VCA = nil
		inMemKB = nil
		err := inMemDB.Close()
		if err != nil {
			panic(err)
		}
		cdc = nil
		memCDC = nil
		inMemDB = nil
		sdk.GlobalCtxCache = nil
		err = os.RemoveAll("data")
		if err != nil {
			panic(err)
		}
		time.Sleep(2 * time.Second)
		codec.TestMode = 0
	}
	return
}

// NewInMemoryTendermintNodeProto will create a TM node with only one validator. LeanViper is disabled.
func NewInMemoryTendermintNodeProto(t *testing.T, genesisState []byte) (tendermintNode *node.Node, keybase keys.Keybase, cleanup func()) {
	return NewInMemoryTendermintNodeProtoWithValidators(t, genesisState, nil)
}

// NewInMemoryTendermintNodeWithValidators will create a TM node with 'n' "validators".
// If "validators" is nil, this creates a pre-leanvipr TM node, else it will enable lean viper
func NewInMemoryTendermintNodeProtoWithValidators(t *testing.T, genesisState []byte, validators []crypto.PrivateKey) (tendermintNode *node.Node, keybase keys.Keybase, cleanup func()) {
	// create the in memory tendermint node and keybase
	tendermintNode, keybase = inMemTendermintNodeWithValidators(genesisState, validators)
	// test assertions
	if tendermintNode == nil {
		panic("tendermintNode should not be nil")
	}
	if keybase == nil {
		panic("should not be nil")
	}
	assert.NotNil(t, tendermintNode)
	assert.NotNil(t, keybase)

	// init cache in memory
	defaultConfig := sdk.DefaultTestingViperConfig()
	if validators != nil {
		defaultConfig.ViperConfig.LeanViper = true
	}
	viperTypes.InitConfig(&viperTypes.HostedBlockchains{
		M: make(map[string]viperTypes.HostedBlockchain),
	}, nil, tendermintNode.Logger, defaultConfig)
	// start the in memory node
	err := tendermintNode.Start()
	if err != nil {
		panic(err)
	}
	// assert that it is not nil
	assert.Nil(t, err)
	// provide cleanup function
	cleanup = func() {
		codec.TestMode = 0

		err = tendermintNode.Stop()
		if err != nil {
			panic(err)
		}

		viperTypes.CleanViperNodes()

		VCA = nil
		inMemKB = nil
		err := inMemDB.Close()
		if err != nil {
			panic(err)
		}
		cdc = nil
		memCDC = nil
		inMemDB = nil
		sdk.GlobalCtxCache = nil
		err = os.RemoveAll("data")
		if err != nil {
			panic(err)
		}
		time.Sleep(3 * time.Second)
	}
	return
}

// inMemTendermintNodeWithValidators will create a TM node with 'n' "validators".
// If "validators" is nil, LeanVipr is disabled and uses in memory CB as the sole validator for consensus
func inMemTendermintNodeWithValidators(genesisState []byte, validatorsPk []crypto.PrivateKey) (*node.Node, keys.Keybase) {
	kb := getInMemoryKeybase()
	cb, err := kb.GetCoinbase()
	if err != nil {
		panic(err)
	}
	pk, err := kb.ExportPrivateKeyObject(cb.GetAddress(), "test")
	if err != nil {
		panic(err)
	}
	genDocRequestor := func() (*types.GenesisDoc, error) {
		return &types.GenesisDoc{
			GenesisTime: time.Time{},
			ChainID:     "viper-test",
			ConsensusParams: &types.ConsensusParams{
				Block: types.BlockParams{
					MaxBytes:   15000,
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
			AppState:   genesisState,
		}, nil
	}
	loggerFile, _ := os.Open(os.DevNull)
	c := config{
		TmConfig: getTestConfig(),
		Logger:   log.NewTMLogger(loggerFile),
	}
	db := getInMemoryDB()
	traceWriter, err := openTraceWriter(c.TraceWriter)
	if err != nil {
		panic(err)
	}
	nodeKey := p2p.NodeKey{PrivKey: pk}
	var privVal *privval.FilePVLean
	if validatorsPk == nil {
		// only set cb as validator
		privVal = privval.GenFilePVLean(c.TmConfig.PrivValidatorKey, c.TmConfig.PrivValidatorState)
		privVal.Keys[0].PrivKey = pk
		privVal.Keys[0].PubKey = pk.PubKey()
		privVal.Keys[0].Address = pk.PubKey().Address()
		viperTypes.CleanViperNodes()
		viperTypes.AddViperNodeByFilePVKey(privVal.Keys[0], c.Logger)
	} else {
		// (LeanVIPR) Set multiple nodes as validators
		viperTypes.CleanViperNodes()
		// generating a stub of n validators
		privVal = privval.GenFilePVsLean(c.TmConfig.PrivValidatorKey, c.TmConfig.PrivValidatorState, uint(len(validatorsPk)))
		// replace the stub validators with the correct validators
		for i, pk := range validatorsPk {
			privVal.Keys[i].PrivKey = pk.PrivKey()
			privVal.Keys[i].PubKey = pk.PubKey()
			privVal.Keys[i].Address = pk.PubKey().Address()
			viperTypes.AddViperNode(pk, c.Logger)
		}
	}

	dbRequestor := func(*node.DBContext) (dbm.DB, error) {
		return db, nil
	}
	app := GetApp(c.Logger, db, traceWriter)
	txDB := dbm.NewMemDB()
	tmNode, err := node.NewNode(app.BaseApp,
		c.TmConfig,
		0,
		privVal,
		&nodeKey,
		proxy.NewLocalClientCreator(app),
		sdk.NewTransactionIndexer(txDB),
		genDocRequestor,
		dbRequestor,
		node.DefaultMetricsRequestor(c.TmConfig.Instrumentation),
		c.Logger.With("module", "node"),
	)
	if err != nil {
		panic(err)
	}
	VCA = app
	app.SetTxIndexer(tmNode.TxIndexer())
	app.SetBlockstore(tmNode.BlockStore())
	app.SetEvidencePool(tmNode.EvidencePool())
	app.viperKeeper.TmNode = local.New(tmNode)
	app.SetTendermintNode(tmNode)
	return tmNode, kb
}

func TestNewInMemoryAmino(t *testing.T) {
	_, _, cleanup := NewInMemoryTendermintNodeAmino(t, oneAppTwoNodeGenesis())
	defer cleanup()
}
func TestNewInMemoryProto(t *testing.T) {
	_, _, cleanup := NewInMemoryTendermintNodeProto(t, oneAppTwoNodeGenesis())
	defer cleanup()
}

var (
	memCDC  *codec.Codec
	inMemKB keys.Keybase
	memCLI  client.Client
	inMemDB dbm.DB
)

func getInMemoryKeybase() keys.Keybase {
	if inMemKB == nil {
		inMemKB = keys.NewInMemory()
		_, err := inMemKB.Create("test")
		if err != nil {
			panic(err)
		}
		_, err = inMemKB.GetCoinbase()
		if err != nil {
			panic(err)
		}
	}
	return inMemKB
}

func getInMemoryDB() dbm.DB {
	if inMemDB == nil {
		inMemDB = dbm.NewMemDB()
	}
	return inMemDB
}

// GenFilePV generates a new validator with randomly generated private key
// and sets the filePaths, but does not call Save().
func GenFilePV(keyFilePath, stateFilePath string) *privval.FilePV {
	return privval.GenFilePV(keyFilePath, stateFilePath)
}

func GetApp(logger log.Logger, db dbm.DB, traceWriter io.Writer) *ViperCoreApp {
	creator := func(logger log.Logger, db dbm.DB, _ io.Writer) *ViperCoreApp {
		m := map[string]viperTypes.HostedBlockchain{"0001": {
			ID:  sdk.PlaceholderHash,
			URL: sdk.PlaceholderURL,
		}}
		m1 := map[string]viperTypes.GeoZone{"0001": {
			ID: sdk.PlaceholderHash,
		}}
		p := NewViperCoreApp(GenState, getInMemoryKeybase(), getInMemoryTMClient(), &viperTypes.HostedBlockchains{M: m, L: sync.Mutex{}}, &viperTypes.HostedGeoZones{M: m1, L: sync.Mutex{}}, logger, db, false, 5000000, bam.SetPruning(store.PruneNothing))
		return p
	}
	return creator(logger, db, traceWriter)
}

func memCodec() *codec.Codec {
	if memCDC == nil {
		memCDC = codec.NewCodec(types2.NewInterfaceRegistry())
		module.NewBasicManager(
			requestors.AppModuleBasic{},
			authentication.AppModuleBasic{},
			governance.AppModuleBasic{},
			servicers.AppModuleBasic{},
			viper.AppModuleBasic{},
		).RegisterCodec(memCDC)
		sdk.RegisterCodec(memCDC)
		crypto.RegisterAmino(memCDC.AminoCodec().Amino)
	}
	return memCDC
}

func memCodecMod(upgrade bool) *codec.Codec {
	if memCDC == nil {
		memCDC = codec.NewCodec(types2.NewInterfaceRegistry())
		module.NewBasicManager(
			requestors.AppModuleBasic{},
			authentication.AppModuleBasic{},
			governance.AppModuleBasic{},
			servicers.AppModuleBasic{},
			viper.AppModuleBasic{},
		).RegisterCodec(memCDC)
		sdk.RegisterCodec(memCDC)
		crypto.RegisterAmino(memCDC.AminoCodec().Amino)
	}
	memCDC.SetUpgradeOverride(upgrade)
	return memCDC
}

func getInMemoryTMClient() client.Client {
	if memCLI == nil || !memCLI.IsRunning() {
		memCLI, _ = http.New(tmCfg.TestConfig().RPC.ListenAddress, "/websocket")
	}
	return memCLI
}

func subscribeTo(t *testing.T, eventType string) (cli client.Client, stopClient func(), eventChan <-chan cTypes.ResultEvent) {
	ctx, cancel := getBackgroundContext()
	cli = getInMemoryTMClient()
	if !cli.IsRunning() {
		_ = cli.Start()
	}
	stopClient = func() {
		err := cli.UnsubscribeAll(ctx, "helpers")
		if err != nil {
			t.Fatal(err)
		}
		err = cli.Stop()
		if err != nil {
			t.Fatal(err)
		}
		memCLI = nil
		cancel()
	}
	eventChan, err := cli.Subscribe(ctx, "helpers", types.QueryForEvent(eventType).String(), 5)
	if err != nil {
		panic(err)
	}
	return
}

func getBackgroundContext() (context.Context, func()) {
	return context.WithCancel(context.Background())
}

func getTestConfig() (newTMConfig *tmCfg.Config) {
	newTMConfig = tmCfg.DefaultConfig()
	// setup tendermint node config
	newTMConfig.SetRoot("data")
	newTMConfig.FastSyncMode = false
	newTMConfig.NodeKey = "data" + FS + sdk.DefaultNKName
	newTMConfig.PrivValidatorKey = "data" + FS + sdk.DefaultPVKName
	newTMConfig.PrivValidatorState = "data" + FS + sdk.DefaultPVSName
	newTMConfig.RPC.ListenAddress = sdk.DefaultListenAddr + "36657"
	newTMConfig.P2P.ListenAddress = sdk.DefaultListenAddr + "36656" // Node listen address. (0.0.0.0:0 means any interface, any port)
	newTMConfig.Consensus = tmCfg.TestConsensusConfig()
	newTMConfig.Consensus.CreateEmptyBlocks = true // Set this to false to only produce blocks when there are txs or when the AppHash changes
	newTMConfig.Consensus.SkipTimeoutCommit = false
	newTMConfig.Consensus.CreateEmptyBlocksInterval = time.Duration(500) * time.Millisecond
	newTMConfig.Consensus.TimeoutCommit = time.Duration(500) * time.Millisecond
	newTMConfig.P2P.MaxNumInboundPeers = 4
	newTMConfig.P2P.MaxNumOutboundPeers = 4
	viperTypes.InitClientBlockAllowance(10000)
	return
}

func getUnstakedAccount(kb keys.Keybase) *keys.KeyPair {
	cb, err := kb.GetCoinbase()
	if err != nil {
		panic(err)
	}
	kps, err := kb.List()
	if err != nil {
		panic(err)
	}
	if len(kps) > 2 {
		panic("get unstaked account only works with the default 2 keypairs")
	}
	for _, kp := range kps {
		if kp.PublicKey != cb.PublicKey {
			return &kp
		}
	}
	return nil
}

func oneAppTwoNodeGenesis() []byte {
	kb := getInMemoryKeybase()
	kp1, err := kb.GetCoinbase()
	if err != nil {
		panic(err)
	}
	kp2, err := kb.Create("test")
	if err != nil {
		panic(err)
	}
	pubKey := kp1.PublicKey
	pubKey2 := kp2.PublicKey
	defaultGenesis := module.NewBasicManager(
		requestors.AppModuleBasic{},
		authentication.AppModuleBasic{},
		governance.AppModuleBasic{},
		servicers.AppModuleBasic{},
		viper.AppModuleBasic{},
	).DefaultGenesis()
	// set coinbase as a validator
	rawPOS := defaultGenesis[servicersTypes.ModuleName]
	var posGenesisState servicersTypes.GenesisState
	memCodec().MustUnmarshalJSON(rawPOS, &posGenesisState)
	posGenesisState.Validators = append(posGenesisState.Validators,
		servicersTypes.Validator{Address: sdk.Address(pubKey.Address()),
			PublicKey:    pubKey,
			Status:       sdk.Staked,
			Chains:       []string{dummyChainsHash},
			ServiceURL:   sdk.PlaceholderServiceURL,
			StakedTokens: sdk.NewInt(1000000000000000)})
	res := memCodec().MustMarshalJSON(posGenesisState)
	defaultGenesis[servicersTypes.ModuleName] = res

	// setup application
	rawApps := defaultGenesis[requestorsTypes.ModuleName]
	var requestorsGenesisState requestorsTypes.GenesisState
	memCodec().MustUnmarshalJSON(rawApps, &requestorsGenesisState)
	// app 1
	requestorsGenesisState.Requestors = append(requestorsGenesisState.Requestors, requestorsTypes.Requestor{
		Address:                 kp2.GetAddress(),
		PublicKey:               kp2.PublicKey,
		Jailed:                  false,
		Status:                  sdk.Staked,
		Chains:                  []string{dummyChainsHash},
		StakedTokens:            sdk.NewInt(10000000),
		MaxRelays:               sdk.NewInt(100000),
		UnstakingCompletionTime: time.Time{},
	})
	res2 := memCodec().MustMarshalJSON(requestorsGenesisState)
	defaultGenesis[requestorsTypes.ModuleName] = res2
	// set coinbase as account holding coins
	rawAccounts := defaultGenesis[authentication.ModuleName]
	var authGenState authentication.GenesisState
	memCodec().MustUnmarshalJSON(rawAccounts, &authGenState)
	authGenState.Accounts = append(authGenState.Accounts, &authentication.BaseAccount{
		Address: sdk.Address(pubKey.Address()),
		Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultStakeDenom, sdk.NewInt(1000000000))),
		PubKey:  pubKey,
	})
	// add second account
	authGenState.Accounts = append(authGenState.Accounts, &authentication.BaseAccount{
		Address: sdk.Address(pubKey2.Address()),
		Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultStakeDenom, sdk.NewInt(1000000000))),
		PubKey:  pubKey,
	})
	res3 := memCodec().MustMarshalJSON(authGenState)
	defaultGenesis[authentication.ModuleName] = res3
	// set default chain for module
	rawViper := defaultGenesis[viperTypes.ModuleName]
	var viperGenesisState viperTypes.GenesisState
	memCodec().MustUnmarshalJSON(rawViper, &viperGenesisState)
	viperGenesisState.Params.SupportedBlockchains = []string{"0001"}
	res4 := memCodec().MustMarshalJSON(viperGenesisState)
	defaultGenesis[viperTypes.ModuleName] = res4
	// set default governance in genesis
	var govGenesisState govTypes.GenesisState
	rawGov := defaultGenesis[govTypes.ModuleName]
	memCodec().MustUnmarshalJSON(rawGov, &govGenesisState)
	nMACL := createTestACL(kp1)
	govGenesisState.Params.Upgrade = govTypes.NewUpgrade(10000, "2.0.0")
	govGenesisState.Params.ACL = nMACL
	govGenesisState.Params.DAOOwner = kp1.GetAddress()
	govGenesisState.DAOTokens = sdk.NewInt(1000)
	res5 := memCodec().MustMarshalJSON(govGenesisState)
	defaultGenesis[govTypes.ModuleName] = res5
	// end genesis setup
	GenState = defaultGenesis
	j, _ := memCodec().MarshalJSONIndent(defaultGenesis, "", "    ")
	return j
}

var testACL govTypes.ACL

func resetTestACL() {
	testACL = nil
}

func createTestACL(kp keys.KeyPair) govTypes.ACL {
	if testACL == nil {
		acl := govTypes.ACL{}
		acl = make([]govTypes.ACLPair, 0)
		acl.SetOwner("requestor/MinimumRequestorStake", kp.GetAddress())
		acl.SetOwner("requestor/RequestorUnstakingTime", kp.GetAddress())
		acl.SetOwner("requestor/BaseRelaysPerVIPR", kp.GetAddress())
		acl.SetOwner("requestor/MaxRequestors", kp.GetAddress())
		acl.SetOwner("requestor/MaximumChains", kp.GetAddress())
		acl.SetOwner("requestor/ParticipationRate", kp.GetAddress())
		acl.SetOwner("requestor/StabilityModulation", kp.GetAddress())
		acl.SetOwner("requestor/MinNumServicers", kp.GetAddress())
		acl.SetOwner("requestor/MaxNumServicers", kp.GetAddress())
		acl.SetOwner("authentication/MaxMemoCharacters", kp.GetAddress())
		acl.SetOwner("authentication/TxSigLimit", kp.GetAddress())
		acl.SetOwner("authentication/FeeMultipliers", kp.GetAddress())
		acl.SetOwner("governance/acl", kp.GetAddress())
		acl.SetOwner("governance/daoOwner", kp.GetAddress())
		acl.SetOwner("governance/upgrade", kp.GetAddress())
		acl.SetOwner("vipernet/ClaimExpiration", kp.GetAddress())
		acl.SetOwner("vipernet/ClaimSubmissionWindow", kp.GetAddress())
		acl.SetOwner("vipernet/MinimumNumberOfProofs", kp.GetAddress())
		acl.SetOwner("vipernet/ReplayAttackBurnMultiplier", kp.GetAddress())
		acl.SetOwner("vipernet/SupportedBlockchains", kp.GetAddress())
		acl.SetOwner("pos/BlocksPerSession", kp.GetAddress())
		acl.SetOwner("pos/DAOAllocation", kp.GetAddress())
		acl.SetOwner("pos/DowntimeJailDuration", kp.GetAddress())
		acl.SetOwner("pos/MaxEvidenceAge", kp.GetAddress())
		acl.SetOwner("pos/MaximumChains", kp.GetAddress())
		acl.SetOwner("pos/MaxJailedBlocks", kp.GetAddress())
		acl.SetOwner("pos/MaxValidators", kp.GetAddress())
		acl.SetOwner("pos/MinSignedPerWindow", kp.GetAddress())
		acl.SetOwner("pos/ProposerPercentage", kp.GetAddress())
		acl.SetOwner("pos/RequestorAllocation", kp.GetAddress())
		acl.SetOwner("pos/TokenRewardFactor", kp.GetAddress())
		acl.SetOwner("pos/SignedBlocksWindow", kp.GetAddress())
		acl.SetOwner("pos/SlashFractionDoubleSign", kp.GetAddress())
		acl.SetOwner("pos/SlashFractionDowntime", kp.GetAddress())
		acl.SetOwner("pos/StakeDenom", kp.GetAddress())
		acl.SetOwner("pos/StakeMinimum", kp.GetAddress())
		acl.SetOwner("pos/UnstakingTime", kp.GetAddress())
		acl.SetOwner("pos/UnstakingTime", kp.GetAddress())
		testACL = acl
	}
	return testACL
}

func twoValTwoNodeGenesisState8() (genbz []byte, vals []servicersTypes.Validator) {
	kb := getInMemoryKeybase()
	kp1, err := kb.GetCoinbase()
	if err != nil {
		panic(err)
	}
	kp2, err := kb.Create("test")
	if err != nil {
		panic(err)
	}
	kp3, err := kb.Create("test")
	if err != nil {
		panic(err)
	}
	kp4, err := kb.Create("test")
	if err != nil {
		panic(err)
	}
	pubKey := kp1.PublicKey
	pubKey2 := kp2.PublicKey
	pubKey3 := kp3.PublicKey
	pubkey4 := kp4.PublicKey
	defaultGenesis := module.NewBasicManager(
		requestors.AppModuleBasic{},
		authentication.AppModuleBasic{},
		governance.AppModuleBasic{},
		servicers.AppModuleBasic{},
		viper.AppModuleBasic{},
	).DefaultGenesis()
	// set coinbase as a validator
	rawPOS := defaultGenesis[servicersTypes.ModuleName]
	var posGenesisState servicersTypes.GenesisState
	memCodec().MustUnmarshalJSON(rawPOS, &posGenesisState)
	posGenesisState.Validators = append(posGenesisState.Validators,
		servicersTypes.Validator{
			Address:                 sdk.Address(pubKey.Address()),
			PublicKey:               pubKey,
			Jailed:                  false,
			Status:                  sdk.Staked,
			Chains:                  []string{dummyChainsHash},
			ServiceURL:              sdk.PlaceholderServiceURL,
			StakedTokens:            sdk.NewInt(1000000000000000),
			UnstakingCompletionTime: time.Time{},
			OutputAddress:           kp3.GetAddress(),
		})
	posGenesisState.Validators = append(posGenesisState.Validators,
		servicersTypes.Validator{
			Address:                 sdk.Address(pubKey2.Address()),
			PublicKey:               pubKey2,
			Jailed:                  false,
			Status:                  sdk.Staked,
			Chains:                  []string{dummyChainsHash},
			ServiceURL:              sdk.PlaceholderServiceURL,
			StakedTokens:            sdk.NewInt(1000000000),
			UnstakingCompletionTime: time.Time{},
			OutputAddress:           kp4.GetAddress(),
		})
	posGenesisState.Params.UnstakingTime = time.Nanosecond
	posGenesisState.Params.SessionBlockFrequency = 5
	res := memCodec().MustMarshalJSON(posGenesisState)
	defaultGenesis[servicersTypes.ModuleName] = res
	// set coinbase as account holding coins
	rawAccounts := defaultGenesis[authentication.ModuleName]
	var authGenState authentication.GenesisState
	memCodec().MustUnmarshalJSON(rawAccounts, &authGenState)
	authGenState.Accounts = append(authGenState.Accounts, &authentication.BaseAccount{
		Address: sdk.Address(pubKey.Address()),
		Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultStakeDenom, sdk.NewInt(1000000000))),
		PubKey:  pubKey,
	})
	// add second account
	authGenState.Accounts = append(authGenState.Accounts, &authentication.BaseAccount{
		Address: sdk.Address(pubKey2.Address()),
		Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultStakeDenom, sdk.NewInt(1000000000))),
		PubKey:  pubKey,
	})
	authGenState.Accounts = append(authGenState.Accounts, &authentication.BaseAccount{
		Address: sdk.Address(pubKey3.Address()),
		Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultStakeDenom, sdk.NewInt(1000000000))),
		PubKey:  pubKey3,
	})
	// add second account
	authGenState.Accounts = append(authGenState.Accounts, &authentication.BaseAccount{
		Address: sdk.Address(pubkey4.Address()),
		Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultStakeDenom, sdk.NewInt(1000000000))),
		PubKey:  pubkey4,
	})
	res2 := memCodec().MustMarshalJSON(authGenState)
	defaultGenesis[authentication.ModuleName] = res2
	// set default chain for module
	rawViper := defaultGenesis[viperTypes.ModuleName]
	var viperGenesisState viperTypes.GenesisState
	memCodec().MustUnmarshalJSON(rawViper, &viperGenesisState)
	viperGenesisState.Params.SupportedBlockchains = []string{dummyChainsHash}
	res3 := memCodec().MustMarshalJSON(viperGenesisState)
	defaultGenesis[viperTypes.ModuleName] = res3
	// set default governance in genesis
	var govGenesisState govTypes.GenesisState
	rawGov := defaultGenesis[govTypes.ModuleName]
	memCodec().MustUnmarshalJSON(rawGov, &govGenesisState)
	nMACL := createTestACL(kp1)
	govGenesisState.Params.Upgrade = govTypes.NewUpgrade(10000, "2.0.0")
	govGenesisState.Params.ACL = nMACL
	govGenesisState.Params.DAOOwner = kp1.GetAddress()
	govGenesisState.DAOTokens = sdk.NewInt(1000)
	res4 := memCodec().MustMarshalJSON(govGenesisState)
	defaultGenesis[govTypes.ModuleName] = res4
	// end genesis setup
	GenState = defaultGenesis
	j, _ := memCodec().MarshalJSONIndent(defaultGenesis, "", "    ")
	return j, posGenesisState.Validators
}

func twoValTwoNodeGenesisState() (genbz []byte, vals []servicersTypes.Validator) {
	kb := getInMemoryKeybase()
	kp1, err := kb.GetCoinbase()
	if err != nil {
		panic(err)
	}
	kp2, err := kb.Create("test")
	if err != nil {
		panic(err)
	}
	kp3, err := kb.Create("test")
	if err != nil {
		panic(err)
	}
	kp4, err := kb.Create("test")
	if err != nil {
		panic(err)
	}
	pubKey := kp1.PublicKey
	pubKey2 := kp2.PublicKey
	pubKey3 := kp3.PublicKey
	pubkey4 := kp4.PublicKey
	defaultGenesis := module.NewBasicManager(
		requestors.AppModuleBasic{},
		authentication.AppModuleBasic{},
		governance.AppModuleBasic{},
		servicers.AppModuleBasic{},
		viper.AppModuleBasic{},
	).DefaultGenesis()
	// set coinbase as a validator
	rawPOS := defaultGenesis[servicersTypes.ModuleName]
	var posGenesisState servicersTypes.GenesisState
	memCodec().MustUnmarshalJSON(rawPOS, &posGenesisState)
	posGenesisState.Validators = append(posGenesisState.Validators,
		servicersTypes.Validator{
			Address:                 sdk.Address(pubKey.Address()),
			PublicKey:               pubKey,
			Jailed:                  false,
			Status:                  sdk.Staked,
			Chains:                  []string{dummyChainsHash},
			ServiceURL:              sdk.PlaceholderServiceURL,
			StakedTokens:            sdk.NewInt(1000000000000000),
			UnstakingCompletionTime: time.Time{},
			OutputAddress:           nil,
		})
	posGenesisState.Validators = append(posGenesisState.Validators,
		servicersTypes.Validator{
			Address:                 sdk.Address(pubKey2.Address()),
			PublicKey:               pubKey2,
			Jailed:                  false,
			Status:                  sdk.Staked,
			Chains:                  []string{dummyChainsHash},
			ServiceURL:              sdk.PlaceholderServiceURL,
			StakedTokens:            sdk.NewInt(1000000000),
			UnstakingCompletionTime: time.Time{},
			OutputAddress:           nil,
		})
	posGenesisState.Params.UnstakingTime = time.Nanosecond
	posGenesisState.Params.SessionBlockFrequency = 5
	res := memCodec().MustMarshalJSON(posGenesisState)
	defaultGenesis[servicersTypes.ModuleName] = res
	// set coinbase as account holding coins
	rawAccounts := defaultGenesis[authentication.ModuleName]
	var authGenState authentication.GenesisState
	memCodec().MustUnmarshalJSON(rawAccounts, &authGenState)
	authGenState.Accounts = append(authGenState.Accounts, &authentication.BaseAccount{
		Address: sdk.Address(pubKey.Address()),
		Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultStakeDenom, sdk.NewInt(1000000000))),
		PubKey:  pubKey,
	})
	// add second account
	authGenState.Accounts = append(authGenState.Accounts, &authentication.BaseAccount{
		Address: sdk.Address(pubKey2.Address()),
		Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultStakeDenom, sdk.NewInt(1000000000))),
		PubKey:  pubKey,
	})
	authGenState.Accounts = append(authGenState.Accounts, &authentication.BaseAccount{
		Address: sdk.Address(pubKey3.Address()),
		Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultStakeDenom, sdk.NewInt(1000000000))),
		PubKey:  pubKey3,
	})
	// add second account
	authGenState.Accounts = append(authGenState.Accounts, &authentication.BaseAccount{
		Address: sdk.Address(pubkey4.Address()),
		Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultStakeDenom, sdk.NewInt(1000000000))),
		PubKey:  pubkey4,
	})
	res2 := memCodec().MustMarshalJSON(authGenState)
	defaultGenesis[authentication.ModuleName] = res2
	// set default chain for module
	rawViper := defaultGenesis[viperTypes.ModuleName]
	var viperGenesisState viperTypes.GenesisState
	memCodec().MustUnmarshalJSON(rawViper, &viperGenesisState)
	viperGenesisState.Params.SupportedBlockchains = []string{dummyChainsHash}
	res3 := memCodec().MustMarshalJSON(viperGenesisState)
	defaultGenesis[viperTypes.ModuleName] = res3
	// set default governance in genesis
	var govGenesisState govTypes.GenesisState
	rawGov := defaultGenesis[govTypes.ModuleName]
	memCodec().MustUnmarshalJSON(rawGov, &govGenesisState)
	nMACL := createTestACL(kp1)
	govGenesisState.Params.Upgrade = govTypes.NewUpgrade(10000, "2.0.0")
	govGenesisState.Params.ACL = nMACL
	govGenesisState.Params.DAOOwner = kp1.GetAddress()
	govGenesisState.DAOTokens = sdk.NewInt(1000)
	res4 := memCodec().MustMarshalJSON(govGenesisState)
	defaultGenesis[govTypes.ModuleName] = res4
	// end genesis setup
	GenState = defaultGenesis
	j, _ := memCodec().MarshalJSONIndent(defaultGenesis, "", "    ")
	return j, posGenesisState.Validators
}

func fiveValidatorsOneAppGenesis() (genBz []byte, keys []crypto.PrivateKey, validators servicersTypes.Validators, app requestorsTypes.Requestor) {
	kb := getInMemoryKeybase()
	// create keypairs
	kp1, err := kb.GetCoinbase()
	if err != nil {
		panic(err)
	}
	kp2, err := kb.Create("test")
	if err != nil {
		panic(err)
	}
	pk1, err := kb.ExportPrivateKeyObject(kp1.GetAddress(), "test")
	if err != nil {
		panic(err)
	}
	pk2, err := kb.ExportPrivateKeyObject(kp2.GetAddress(), "test")
	if err != nil {
		panic(err)
	}
	var kys []crypto.PrivateKey
	kys = append(kys, pk1, pk2, crypto.GenerateEd25519PrivKey(), crypto.GenerateEd25519PrivKey(), crypto.GenerateEd25519PrivKey())
	// get public kys
	pubKey := kp1.PublicKey
	pubKey2 := kp2.PublicKey
	pubKey3 := kys[2].PublicKey()
	pubKey4 := kys[3].PublicKey()
	pubKey5 := kys[4].PublicKey()
	defaultGenesis := module.NewBasicManager(
		requestors.AppModuleBasic{},
		authentication.AppModuleBasic{},
		governance.AppModuleBasic{},
		servicers.AppModuleBasic{},
		viper.AppModuleBasic{},
	).DefaultGenesis()
	// setup validators
	rawPOS := defaultGenesis[servicersTypes.ModuleName]
	var posGenesisState servicersTypes.GenesisState
	memCodec().MustUnmarshalJSON(rawPOS, &posGenesisState)
	// validator 1
	posGenesisState.Validators = append(posGenesisState.Validators,
		servicersTypes.Validator{Address: sdk.Address(pubKey.Address()),
			PublicKey:    pubKey,
			Status:       sdk.Staked,
			Chains:       []string{dummyChainsHash},
			ServiceURL:   sdk.PlaceholderServiceURL,
			StakedTokens: sdk.NewInt(1000000000000000000)})
	// validator 2
	posGenesisState.Validators = append(posGenesisState.Validators,
		servicersTypes.Validator{Address: sdk.Address(pubKey2.Address()),
			PublicKey:    pubKey2,
			Status:       sdk.Staked,
			Chains:       []string{dummyChainsHash},
			ServiceURL:   sdk.PlaceholderServiceURL,
			StakedTokens: sdk.NewInt(10000000)})
	// validator 3
	posGenesisState.Validators = append(posGenesisState.Validators,
		servicersTypes.Validator{Address: sdk.Address(pubKey3.Address()),
			PublicKey:    pubKey3,
			Status:       sdk.Staked,
			Chains:       []string{dummyChainsHash},
			ServiceURL:   sdk.PlaceholderServiceURL,
			StakedTokens: sdk.NewInt(10000000)})
	// validator 4
	posGenesisState.Validators = append(posGenesisState.Validators,
		servicersTypes.Validator{Address: sdk.Address(pubKey4.Address()),
			PublicKey:    pubKey4,
			Status:       sdk.Staked,
			Chains:       []string{dummyChainsHash},
			ServiceURL:   sdk.PlaceholderServiceURL,
			StakedTokens: sdk.NewInt(10000000)})
	// validator 5
	posGenesisState.Validators = append(posGenesisState.Validators,
		servicersTypes.Validator{Address: sdk.Address(pubKey5.Address()),
			PublicKey:    pubKey5,
			Status:       sdk.Staked,
			Chains:       []string{dummyChainsHash},
			ServiceURL:   sdk.PlaceholderServiceURL,
			StakedTokens: sdk.NewInt(10000000)})
	// marshal into json
	res := memCodec().MustMarshalJSON(posGenesisState)
	defaultGenesis[servicersTypes.ModuleName] = res
	// setup applications
	rawApps := defaultGenesis[requestorsTypes.ModuleName]
	var requestorsGenesisState requestorsTypes.GenesisState
	memCodec().MustUnmarshalJSON(rawApps, &requestorsGenesisState)
	// app 1
	requestorsGenesisState.Requestors = append(requestorsGenesisState.Requestors, requestorsTypes.Requestor{
		Address:                 kp2.GetAddress(),
		PublicKey:               kp2.PublicKey,
		Jailed:                  false,
		Status:                  sdk.Staked,
		Chains:                  []string{dummyChainsHash},
		StakedTokens:            sdk.NewInt(10000000),
		MaxRelays:               sdk.NewInt(100000),
		UnstakingCompletionTime: time.Time{},
	})
	res2 := memCodec().MustMarshalJSON(requestorsGenesisState)
	defaultGenesis[requestorsTypes.ModuleName] = res2
	// accounts
	rawAccounts := defaultGenesis[authentication.ModuleName]
	var authGenState authentication.GenesisState
	memCodec().MustUnmarshalJSON(rawAccounts, &authGenState)
	authGenState.Accounts = append(authGenState.Accounts, &authentication.BaseAccount{
		Address: sdk.Address(pubKey.Address()),
		Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultStakeDenom, sdk.NewInt(1000000000))),
		PubKey:  pubKey,
	})
	res = memCodec().MustMarshalJSON(authGenState)
	defaultGenesis[authentication.ModuleName] = res
	// setup supported blockchains
	rawViper := defaultGenesis[viperTypes.ModuleName]
	var viperGenesisState viperTypes.GenesisState
	memCodec().MustUnmarshalJSON(rawViper, &viperGenesisState)
	viperGenesisState.Params.SupportedBlockchains = []string{dummyChainsHash}
	viperGenesisState.Params.ClaimSubmissionWindow = 10
	res3 := memCodec().MustMarshalJSON(viperGenesisState)
	defaultGenesis[viperTypes.ModuleName] = res3
	// set default governance in genesis
	var govGenesisState govTypes.GenesisState
	rawGov := defaultGenesis[govTypes.ModuleName]
	memCodec().MustUnmarshalJSON(rawGov, &govGenesisState)
	nMACL := createTestACL(kp1)
	govGenesisState.Params.Upgrade = govTypes.NewUpgrade(10000, "2.0.0")
	govGenesisState.Params.ACL = nMACL
	govGenesisState.Params.DAOOwner = kp1.GetAddress()
	govGenesisState.DAOTokens = sdk.NewInt(1000)
	res4 := memCodec().MustMarshalJSON(govGenesisState)
	defaultGenesis[govTypes.ModuleName] = res4
	// end genesis setup
	GenState = defaultGenesis
	j, _ := memCodec().MarshalJSONIndent(defaultGenesis, "", "    ")
	return j, kys, posGenesisState.Validators, requestorsGenesisState.Requestors[0]
}

//
//func TestGatewayChecker(t *testing.T) {
//	startheight := 14681
//	iterations := 30
//	blocks := 96 // ~ 24 hours
//	oldTotalSupply := 0
//	// Code below
//	type Supply struct {
//		Total string `json:"total"`
//	}
//	type Result struct {
//		Inflation    int `json:"inflation"`
//		Day          int `json:"days_ago"`
//		Height       int `json:"height"`
//		DeviationPer int `json:"dev_perc"`
//	}
//	var results []Result
//	var supply Supply
//	var sum int
//	for i := 0; i <= iterations; i++ {
//		jsonStr := `{"height":` + strconv.Itoa(startheight) + `}`
//		req, _ := http2.NewRequest("POST", "http://localhost:8081/v1/query/supply", bytes.NewBuffer([]byte(jsonStr)))
//		client := http2.Client{}
//		resp, err := client.Do(req)
//		if err != nil {
//			panic(err)
//		}
//		bd, err := ioutil.ReadAll(resp.Body)
//		if err != nil {
//			panic(err)
//		}
//		err = json.Unmarshal(bd, &supply)
//		if err != nil {
//			panic(err)
//		}
//		total, err := strconv.Atoi(supply.Total)
//		if err != nil {
//			panic(err)
//		}
//		if oldTotalSupply != 0 {
//			results = append(results, Result{
//				Inflation: oldTotalSupply - total,
//				Day:       i,
//				Height:    startheight,
//			})
//			sum += oldTotalSupply - total
//		}
//		oldTotalSupply = total
//		startheight = startheight - blocks
//	}
//	avg := sum / iterations
//	for _, result := range results {
//		result.DeviationPer = (100 * (result.Inflation - avg)) / avg
//		bz, err := json.MarshalIndent(result, "", "  ")
//		if err != nil {
//			panic(err)
//		}
//		fmt.Println(string(bz))
//	}
//}
*/
