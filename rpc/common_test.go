package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	fp "path/filepath"
	"testing"
	"time"

	"github.com/tendermint/tendermint/privval"

	types2 "github.com/vipernet-xyz/viper-network/codec/types"

	"github.com/tendermint/tendermint/rpc/client/http"

	"github.com/vipernet-xyz/viper-network/app"
	bam "github.com/vipernet-xyz/viper-network/baseapp"
	"github.com/vipernet-xyz/viper-network/codec"
	"github.com/vipernet-xyz/viper-network/crypto"
	"github.com/vipernet-xyz/viper-network/crypto/keys"
	"github.com/vipernet-xyz/viper-network/store"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/types/module"
	"github.com/vipernet-xyz/viper-network/x/authentication"
	"github.com/vipernet-xyz/viper-network/x/governance"
	govTypes "github.com/vipernet-xyz/viper-network/x/governance/types"
	providers "github.com/vipernet-xyz/viper-network/x/providers"
	providersTypes "github.com/vipernet-xyz/viper-network/x/providers/types"
	"github.com/vipernet-xyz/viper-network/x/servicers"
	servicersTypes "github.com/vipernet-xyz/viper-network/x/servicers/types"
	"github.com/vipernet-xyz/viper-network/x/transfer"
	viper "github.com/vipernet-xyz/viper-network/x/vipernet"
	viperTypes "github.com/vipernet-xyz/viper-network/x/vipernet/types"

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

var FS = string(fp.Separator)

func NewInMemoryTendermintNode(t *testing.T, genesisState []byte) (tendermintNode *node.Node, keybase keys.Keybase, cleanup func()) {
	sdk.VbCCache = sdk.NewCache(1)
	app.MakeCodec() // needed for queries and tx
	// create the in memory tendermint node and keybase
	tendermintNode, keybase = inMemTendermintNode(genesisState)
	// test assertions
	if tendermintNode == nil {
		panic("tendermintNode should not be nil")
	}
	if keybase == nil {
		panic("should not be nil")
	}
	assert.NotNil(t, tendermintNode)
	assert.NotNil(t, keybase)
	// chains := &viperTypes.HostedBlockchains{M: make(map[string]viperTypes.HostedBlockchain)}
	// chains.M[dummyChainsHash] = viperTypes.HostedBlockchain{ID: dummyChainsHash, URL: dummyChainsURL }
	// init cache in memory
	viperTypes.InitConfig(&viperTypes.HostedBlockchains{
		M: make(map[string]viperTypes.HostedBlockchain),
	}, tendermintNode.Logger, sdk.DefaultTestingViperConfig())
	// start the in memory node
	err := tendermintNode.Start()
	// assert that it is not nil
	assert.Nil(t, err)
	// provide cleanup function
	cleanup = func() {
		err = tendermintNode.Stop()
		if err != nil {
			panic(err)
		}
		viperTypes.ClearEvidence()
		viperTypes.ClearSessionCache()
		inMemKB = nil
		//err = os.RemoveAll(tendermintNode.Config().DBPath)
		if err != nil {
			panic(err)
		}
		err = os.RemoveAll("data")
		if err != nil {
			panic(err)
		}
		time.Sleep(1 * time.Second)
	}
	return
}

func TestNewInMemory(t *testing.T) {
	_, _, cleanup := NewInMemoryTendermintNode(t, oneValTwoNodeGenesisState())
	defer cleanup()
}

var (
	memCDC  *codec.Codec
	inMemKB keys.Keybase
	memCLI  client.Client
)

const (
	dummyChainsHash = "0001"
	dummyChainsURL  = "http:127.0.0.1:8081"
	dummyServiceURL = "https://foo.bar:8081"
	defaultTMURI    = "tcp://localhost:26657"
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

func inMemTendermintNode(genesisState []byte) (*node.Node, keys.Keybase) {
	kb := getInMemoryKeybase()
	cb, err := kb.GetCoinbase()
	if err != nil {
		panic(err)
	}
	pk, err := kb.ExportPrivateKeyObject(cb.GetAddress(), "test")
	if err != nil {
		panic(err)
	}
	genDocServicer := func() (*types.GenesisDoc, error) {
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
		Logger:   log.NewTMLogger(log.NewSyncWriter(loggerFile)),
	}
	db := dbm.NewMemDB()
	nodeKey := p2p.NodeKey{PrivKey: pk}
	privVal := privval.GenFilePV(c.TmConfig.PrivValidatorKey, c.TmConfig.PrivValidatorState)
	privVal.Key.PrivKey = pk
	privVal.Key.PubKey = pk.PubKey()
	privVal.Key.Address = pk.PubKey().Address()
	viperTypes.InitPVKeyFile(privVal.Key)

	creator := func(logger log.Logger, db dbm.DB, _ io.Writer) *app.ViperCoreApp {
		m := map[string]viperTypes.HostedBlockchain{sdk.PlaceholderHash: {
			ID:  sdk.PlaceholderHash,
			URL: dummyChainsURL,
		}}
		p := app.NewViperCoreApp(app.GenState, getInMemoryKeybase(), getInMemoryTMClient(), &viperTypes.HostedBlockchains{M: m}, logger, db, false, 5000000, bam.SetPruning(store.PruneNothing))
		return p
	}
	//upgradePrivVal(c.TmConfig)
	dbServicer := func(*node.DBContext) (dbm.DB, error) {
		return db, nil
	}
	txDB := dbm.NewMemDB()
	baseprovider := creator(c.Logger, db, io.Writer(nil))
	tmNode, err := node.NewNode(baseprovider,
		c.TmConfig,
		0,
		privVal,
		&nodeKey,
		proxy.NewLocalClientCreator(baseprovider),
		sdk.NewTransactionIndexer(txDB),
		genDocServicer,
		dbServicer,
		node.DefaultMetricsProvider(c.TmConfig.Instrumentation),
		c.Logger.With("module", "node"),
	)
	if err != nil {
		panic(err)
	}
	baseprovider.SetTxIndexer(tmNode.TxIndexer())
	baseprovider.SetBlockstore(tmNode.BlockStore())
	baseprovider.SetEvidencePool(tmNode.EvidencePool())
	baseprovider.SetTendermintNode(tmNode)
	app.VCA = baseprovider
	return tmNode, kb
}

func memCodec() *codec.Codec {
	if memCDC == nil {
		memCDC = codec.NewCodec(types2.NewInterfaceRegistry())
		module.NewBasicManager(
			providers.AppModuleBasic{},
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
			providers.AppModuleBasic{},
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
		memCLI, _ = http.New(defaultTMURI, "/websocket")
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
	eventChan, err := cli.Subscribe(ctx, "helpers", types.QueryForEvent(eventType).String())
	if err != nil {
		panic(err)
	}
	return
}

func getBackgroundContext() (context.Context, func()) {
	return context.WithCancel(context.Background())
}

func getTestConfig() (tmConfg *tmCfg.Config) {
	tmConfg = tmCfg.TestConfig()
	tmConfg.RPC.ListenAddress = defaultTMURI
	tmConfg.Consensus.CreateEmptyBlocks = true // Set this to false to only produce blocks when there are txs or when the AppHash changes
	tmConfg.Consensus.SkipTimeoutCommit = false
	tmConfg.Consensus.CreateEmptyBlocksInterval = time.Duration(50) * time.Millisecond
	tmConfg.Consensus.TimeoutCommit = time.Duration(50) * time.Millisecond
	tmConfg.TxIndex.Indexer = "kv"
	tmConfg.TxIndex.IndexKeys = "tx.hash,tx.height,message.sender"
	return
}

func oneValTwoNodeGenesisState() []byte {
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
		providers.AppModuleBasic{},
		authentication.AppModuleBasic{},
		governance.AppModuleBasic{},
		servicers.AppModuleBasic{},
		viper.AppModuleBasic{},
		governance.AppModuleBasic{},
		transfer.AppModuleBasic{},
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
			ServiceURL:   dummyServiceURL,
			StakedTokens: sdk.NewInt(1000000000000000)})
	res := memCodec().MustMarshalJSON(posGenesisState)
	defaultGenesis[servicersTypes.ModuleName] = res
	// set coinbase as account holding coins
	rawAccounts := defaultGenesis[authentication.ModuleName]
	var authenticationGenState authentication.GenesisState
	memCodec().MustUnmarshalJSON(rawAccounts, &authenticationGenState)
	authenticationGenState.Accounts = append(authenticationGenState.Accounts, &authentication.BaseAccount{
		Address: sdk.Address(pubKey.Address()),
		Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultStakeDenom, sdk.NewInt(1000000000))),
		PubKey:  pubKey,
	})
	// add second account
	authenticationGenState.Accounts = append(authenticationGenState.Accounts, &authentication.BaseAccount{
		Address: sdk.Address(pubKey2.Address()),
		Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultStakeDenom, sdk.NewInt(1000000000))),
		PubKey:  pubKey,
	})
	res2 := memCodec().MustMarshalJSON(authenticationGenState)
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
	mACL := createTestACL(kp1)
	govGenesisState.Params.ACL = mACL
	govGenesisState.Params.DAOOwner = kp1.GetAddress()
	govGenesisState.Params.Upgrade = govTypes.NewUpgrade(10000, "2.0.0")
	res4 := memCodec().MustMarshalJSON(govGenesisState)
	defaultGenesis[govTypes.ModuleName] = res4
	viperGenesisState.Params.SupportedBlockchains = []string{dummyChainsHash}
	// end genesis setup
	app.GenState = defaultGenesis
	j, _ := memCodec().MarshalJSONIndent(defaultGenesis, "", "    ")
	return j
}

var testACL govTypes.ACL

func createTestACL(kp keys.KeyPair) govTypes.ACL {
	if testACL == nil {
		acl := govTypes.ACL{}
		acl = make([]govTypes.ACLPair, 0)
		acl.SetOwner("authentication/MaxMemoCharacters", kp.GetAddress())
		acl.SetOwner("authentication/TxSigLimit", kp.GetAddress())
		acl.SetOwner("governance/daoOwner", kp.GetAddress())
		acl.SetOwner("governance/acl", kp.GetAddress())
		acl.SetOwner("pos/StakeDenom", kp.GetAddress())
		acl.SetOwner("vipercore/SupportedBlockchains", kp.GetAddress())
		acl.SetOwner("pos/DowntimeJailDuration", kp.GetAddress())
		acl.SetOwner("pos/SlashFractionDoubleSign", kp.GetAddress())
		acl.SetOwner("pos/SlashFractionDowntime", kp.GetAddress())
		acl.SetOwner("authentication/FeeMultipliers", kp.GetAddress())
		acl.SetOwner("provider/MinProviderStake", kp.GetAddress())
		acl.SetOwner("vipercore/ClaimExpiration", kp.GetAddress())
		acl.SetOwner("vipercore/SessionNodeCount", kp.GetAddress())
		acl.SetOwner("vipercore/MinimumNumberOfProofs", kp.GetAddress())
		acl.SetOwner("vipercore/ReplayAttackBurnMultiplier", kp.GetAddress())
		acl.SetOwner("pos/MaxValidators", kp.GetAddress())
		acl.SetOwner("pos/ProposerPercentage", kp.GetAddress())
		acl.SetOwner("provider/StabilityAdjustment", kp.GetAddress())
		acl.SetOwner("provider/ProviderUnstakingTime", kp.GetAddress())
		acl.SetOwner("provider/ParticipationRateOn", kp.GetAddress())
		acl.SetOwner("pos/MaxEvidenceAge", kp.GetAddress())
		acl.SetOwner("pos/MinSignedPerWindow", kp.GetAddress())
		acl.SetOwner("pos/StakeMinimum", kp.GetAddress())
		acl.SetOwner("pos/UnstakingTime", kp.GetAddress())
		acl.SetOwner("pos/TokenRewardFactor", kp.GetAddress())
		acl.SetOwner("provider/BaseRelaysPerVIPR", kp.GetAddress())
		acl.SetOwner("vipercore/ClaimSubmissionWindow", kp.GetAddress())
		acl.SetOwner("pos/DAOAllocation", kp.GetAddress())
		acl.SetOwner("pos/SignedBlocksWindow", kp.GetAddress())
		acl.SetOwner("pos/BlocksPerSession", kp.GetAddress())
		acl.SetOwner("provider/MaxProviders", kp.GetAddress())
		acl.SetOwner("governance/daoOwner", kp.GetAddress())
		acl.SetOwner("governance/upgrade", kp.GetAddress())
		acl.SetOwner("provider/MaximumChains", kp.GetAddress())
		acl.SetOwner("pos/MaximumChains", kp.GetAddress())
		acl.SetOwner("pos/MaxJailedBlocks", kp.GetAddress())

		testACL = acl
	}
	return testACL
}

func fiveValidatorsOneAppGenesis() (genBz []byte, keys []crypto.PrivateKey, validators servicersTypes.Validators, provider providersTypes.Provider) {
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
		providers.AppModuleBasic{},
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
			ServiceURL:   PlaceholderServiceURL,
			StakedTokens: sdk.NewInt(1000000000000000000)})
	// validator 2
	posGenesisState.Validators = append(posGenesisState.Validators,
		servicersTypes.Validator{Address: sdk.Address(pubKey2.Address()),
			PublicKey:    pubKey2,
			Status:       sdk.Staked,
			Chains:       []string{dummyChainsHash},
			ServiceURL:   PlaceholderServiceURL,
			StakedTokens: sdk.NewInt(10000000)})
	// validator 3
	posGenesisState.Validators = append(posGenesisState.Validators,
		servicersTypes.Validator{Address: sdk.Address(pubKey3.Address()),
			PublicKey:    pubKey3,
			Status:       sdk.Staked,
			Chains:       []string{dummyChainsHash},
			ServiceURL:   PlaceholderServiceURL,
			StakedTokens: sdk.NewInt(10000000)})
	// validator 4
	posGenesisState.Validators = append(posGenesisState.Validators,
		servicersTypes.Validator{Address: sdk.Address(pubKey4.Address()),
			PublicKey:    pubKey4,
			Status:       sdk.Staked,
			Chains:       []string{dummyChainsHash},
			ServiceURL:   PlaceholderServiceURL,
			StakedTokens: sdk.NewInt(10000000)})
	// validator 5
	posGenesisState.Validators = append(posGenesisState.Validators,
		servicersTypes.Validator{Address: sdk.Address(pubKey5.Address()),
			PublicKey:    pubKey5,
			Status:       sdk.Staked,
			Chains:       []string{dummyChainsHash},
			ServiceURL:   PlaceholderServiceURL,
			StakedTokens: sdk.NewInt(10000000)})
	// marshal into json
	res := memCodec().MustMarshalJSON(posGenesisState)
	defaultGenesis[servicersTypes.ModuleName] = res
	// setup providers
	rawApps := defaultGenesis[providersTypes.ModuleName]
	var providersGenesisState providersTypes.GenesisState
	memCodec().MustUnmarshalJSON(rawApps, &providersGenesisState)
	// provider 1
	providersGenesisState.Providers = append(providersGenesisState.Providers, providersTypes.Provider{
		Address:                 kp2.GetAddress(),
		PublicKey:               kp2.PublicKey,
		Jailed:                  false,
		Status:                  sdk.Staked,
		Chains:                  []string{dummyChainsHash},
		StakedTokens:            sdk.NewInt(10000000),
		MaxRelays:               sdk.NewInt(100000),
		UnstakingCompletionTime: time.Time{},
	})
	res2 := memCodec().MustMarshalJSON(providersGenesisState)
	defaultGenesis[providersTypes.ModuleName] = res2
	// accounts
	rawAccounts := defaultGenesis[authentication.ModuleName]
	var authenticationGenState authentication.GenesisState
	memCodec().MustUnmarshalJSON(rawAccounts, &authenticationGenState)
	authenticationGenState.Accounts = append(authenticationGenState.Accounts, &authentication.BaseAccount{
		Address: sdk.Address(pubKey.Address()),
		Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultStakeDenom, sdk.NewInt(1000000000))),
		PubKey:  pubKey,
	})
	res = memCodec().MustMarshalJSON(authenticationGenState)
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
	app.GenState = defaultGenesis
	j, _ := memCodec().MarshalJSONIndent(defaultGenesis, "", "    ")
	return j, kys, posGenesisState.Validators, providersGenesisState.Providers[0]
}

type config struct {
	TmConfig    *tmCfg.Config
	Logger      log.Logger
	TraceWriter string
}

func generateChainsJson(configFilePath string, chains []viperTypes.HostedBlockchain) *viperTypes.HostedBlockchains {
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		// ensure directory path made
		err = os.MkdirAll(configFilePath, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
	chainsPath := configFilePath + FS + sdk.DefaultChainsName
	var jsonFile *os.File
	// if does not exist create one
	jsonFile, err := os.OpenFile(chainsPath, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	// generate hosted chains from user input
	// create dummy input for the file
	res, err := json.MarshalIndent(chains, "", "  ")
	if err != nil {
		panic(err)
	}
	// write to the file
	_, err = jsonFile.Write(res)
	if err != nil {
		panic(err)
	}
	// close the file
	err = jsonFile.Close()
	if err != nil {
		panic(err)
	}
	m := make(map[string]viperTypes.HostedBlockchain)
	for _, chain := range chains {
		if err := servicersTypes.ValidateNetworkIdentifier(chain.ID); err != nil {
			panic(errors.New(fmt.Sprintf("invalid ID: %s in network identifier in %s file", chain.ID, app.GlobalConfig.ViperConfig.ChainsName)))
		}
		m[chain.ID] = chain
	}
	// return the map
	return &viperTypes.HostedBlockchains{M: m}
}
