package keeper

import (
	"context"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"os"
	"testing"
	"time"

	types2 "github.com/vipernet-xyz/viper-network/codec/types"

	"github.com/vipernet-xyz/viper-network/codec"
	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	"github.com/vipernet-xyz/viper-network/crypto/keys"
	"github.com/vipernet-xyz/viper-network/store"
	storeTypes "github.com/vipernet-xyz/viper-network/store/types"
	sdk "github.com/vipernet-xyz/viper-network/types"
	viperTypes "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/types/module"
	auth "github.com/vipernet-xyz/viper-network/x/authentication"
	gov "github.com/vipernet-xyz/viper-network/x/governance"
	govKeeper "github.com/vipernet-xyz/viper-network/x/governance/keeper"
	govTypes "github.com/vipernet-xyz/viper-network/x/governance/types"
	requestors "github.com/vipernet-xyz/viper-network/x/requestors"
	requestorsKeeper "github.com/vipernet-xyz/viper-network/x/requestors/keeper"
	requestorsTypes "github.com/vipernet-xyz/viper-network/x/requestors/types"
	servicers "github.com/vipernet-xyz/viper-network/x/servicers"
	servicersKeeper "github.com/vipernet-xyz/viper-network/x/servicers/keeper"
	servicersTypes "github.com/vipernet-xyz/viper-network/x/servicers/types"
	"github.com/vipernet-xyz/viper-network/x/viper-main/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/privval"
	tmStore "github.com/tendermint/tendermint/store"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"
)

var (
	ModuleBasics = module.NewBasicManager(
		auth.AppModuleBasic{},
		gov.AppModuleBasic{},
	)
)

func TestMain(m *testing.M) {
	m.Run()
	err := os.RemoveAll("data")
	if err != nil {
		panic(err)
	}
	os.Exit(0)
}

type simulateRelayKeys struct {
	private crypto.PrivateKey
	client  crypto.PrivateKey
}

type simulateTestResultKeys struct {
	private crypto.PrivateKey
	client  crypto.PrivateKey
}

func NewTestKeybase() keys.Keybase {
	return keys.NewInMemory()
}

// create a codec used only for testing
func makeTestCodec() *codec.Codec {
	var cdc = codec.NewCodec(types2.NewInterfaceRegistry())
	auth.RegisterCodec(cdc)
	gov.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	crypto.RegisterAmino(cdc.AminoCodec().Amino)
	return cdc
}

// : deadcode unused
func createTestInput(t *testing.T, isCheckTx bool) (sdk.Ctx, []servicersTypes.Validator, []requestorsTypes.Requestor, []auth.BaseAccount, Keeper, map[string]*sdk.KVStoreKey, keys.Keybase) {
	sdk.VbCCache = sdk.NewCache(1)
	initPower := int64(100000000000)
	nAccs := int64(5)
	kb := NewTestKeybase()
	_, err := kb.Create("test")
	assert.Nil(t, err)

	keyAcc := sdk.NewKVStoreKey(auth.StoreKey)
	keyParams := sdk.ParamsKey
	tkeyParams := sdk.ParamsTKey
	servicersKey := sdk.NewKVStoreKey(servicersTypes.StoreKey)
	requestorsKey := sdk.NewKVStoreKey(requestorsTypes.StoreKey)
	viperKey := sdk.NewKVStoreKey(types.StoreKey)
	govKey := sdk.NewKVStoreKey(govTypes.StoreKey)
	dKey := sdk.NewKVStoreKey("DiscountKey")

	keys := make(map[string]*sdk.KVStoreKey)
	keys["params"] = keyParams
	keys["pos"] = servicersKey
	keys["requestor"] = requestorsKey

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db, false, 5000000)
	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(servicersKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(requestorsKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(viperKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	err = ms.LoadLatestVersion()
	require.Nil(t, err)

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "test-chain"}, isCheckTx, log.NewNopLogger())
	ctx = ctx.WithConsensusParams(
		&abci.ConsensusParams{
			Validator: &abci.ValidatorParams{
				PubKeyTypes: []string{tmtypes.ABCIPubKeyTypeEd25519},
			},
		},
	)
	ctx = ctx.WithBlockHeader(abci.Header{
		Height: 977,
		Time:   time.Time{},
		LastBlockId: abci.BlockID{
			Hash: types.Hash([]byte("fake")),
		},
	})
	cdc := makeTestCodec()

	maccPerms := map[string][]string{
		auth.FeeCollectorName:          nil,
		requestorsTypes.StakedPoolName: {auth.Burner, auth.Staking, auth.Minter},
		servicersTypes.StakedPoolName:  {auth.Burner, auth.Staking},
		govTypes.DAOAccountName:        {auth.Burner, auth.Staking},
	}

	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[auth.NewModuleAddress(acc).String()] = true
	}
	valTokens := sdk.TokensFromConsensusPower(initPower)

	ethereum := hex.EncodeToString([]byte{01})
	US := hex.EncodeToString([]byte{01})
	hb := types.HostedBlockchains{
		M: map[string]types.HostedBlockchain{ethereum: {
			ID:           ethereum,
			HTTPURL:      "https://www.google.com:443",
			WebSocketURL: "wss://www.google.com/ws",
		}},
	}

	hg := types.HostedGeoZones{
		M: map[string]types.GeoZone{US: {
			ID: US,
		}},
	}

	cb, err := kb.GetCoinbase()
	assert.Nil(t, err)
	addr := tmtypes.Address(cb.GetAddress())
	pk, err := kb.ExportPrivateKeyObject(cb.GetAddress(), "test")
	assert.Nil(t, err)
	types.CleanViperNodes()
	types.AddViperNodeByFilePVKey(privval.FilePVKey{
		Address: addr,
		PubKey:  cb.PublicKey,
		PrivKey: pk,
	}, ctx.Logger())
	types.InitConfig(&hb, &hg, log.NewTMLogger(os.Stdout), sdk.DefaultTestingViperConfig())

	authSubspace := sdk.NewSubspace(auth.DefaultParamspace)
	nodesSubspace := sdk.NewSubspace(servicersTypes.DefaultParamspace)
	appSubspace := sdk.NewSubspace(requestorsTypes.DefaultParamspace)
	viperSubspace := sdk.NewSubspace(types.DefaultParamspace)
	ak := auth.NewKeeper(cdc, keyAcc, authSubspace, maccPerms)
	govKeeper := govKeeper.NewKeeper(cdc, govKey, tkeyParams, dKey, govTypes.ModuleName, ak)
	nk := servicersKeeper.NewKeeper(cdc, servicersKey, ak, nil, govKeeper, nodesSubspace, servicersTypes.ModuleName)
	appk := requestorsKeeper.NewKeeper(cdc, requestorsKey, nk, ak, nil, appSubspace, requestorsTypes.ModuleName)
	appk.SetRequestor(ctx, getTestRequestor())
	keeper := NewKeeper(viperKey, cdc, ak, nk, appk, &hb, &hg, viperSubspace)
	appk.ViperKeeper = keeper
	nk.RequestorKeeper = appk
	assert.Nil(t, err)
	moduleManager := module.NewManager(
		auth.NewAppModule(ak),
		servicers.NewAppModule(nk),
		requestors.NewAppModule(appk),
	)
	genesisState := ModuleBasics.DefaultGenesis()
	moduleManager.InitGenesis(ctx, genesisState)
	initialCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultStakeDenom, valTokens))
	accs := createTestAccs(ctx, int(nAccs), initialCoins, &ak)
	ap := createTestApps(ctx, int(nAccs), sdk.NewIntFromBigInt(new(big.Int).SetUint64(math.MaxUint64)), appk, ak)
	vals := createTestValidators(ctx, int(nAccs), sdk.ZeroInt(), &nk, ak, kb)
	appk.SetParams(ctx, requestorsTypes.DefaultParams())
	nk.SetParams(ctx, servicersTypes.DefaultParams())
	defaultViperParams := types.DefaultParams()
	defaultViperParams.SupportedBlockchains = []string{getTestSupportedBlockchain()}
	keeper.SetParams(ctx, defaultViperParams)
	return ctx, vals, ap, accs, keeper, keys, kb
}

func createTestInputWithLean(t *testing.T, isCheckTx bool) (sdk.Ctx, []servicersTypes.Validator, []requestorsTypes.Requestor, []auth.BaseAccount, Keeper, map[string]*sdk.KVStoreKey, keys.Keybase) {
	ctx, vals, ap, accs, keeper, keys, kb := createTestInput(t, isCheckTx)
	return ctx, vals, ap, accs, keeper, keys, kb
}

var (
	testApp            requestorsTypes.Requestor
	testAppPrivateKey  crypto.PrivateKey
	testSupportedChain string
)

func getTestSupportedBlockchain() string {
	if testSupportedChain == "" {
		testSupportedChain = hex.EncodeToString([]byte{01})
	}
	return testSupportedChain
}

func getTestRequestorPrivateKey() crypto.PrivateKey {
	if testAppPrivateKey == nil {
		testAppPrivateKey = getRandomPrivateKey()
	}
	return testAppPrivateKey
}

func getTestRequestor() requestorsTypes.Requestor {
	if testApp.Address == nil {
		pk := getTestRequestorPrivateKey().PublicKey()
		testApp = requestorsTypes.Requestor{
			Address:                 sdk.Address(pk.Address()),
			PublicKey:               pk,
			Jailed:                  false,
			Status:                  2,
			Chains:                  []string{getTestSupportedBlockchain()},
			StakedTokens:            sdk.NewInt(10000000),
			MaxRelays:               sdk.NewInt(10000000),
			UnstakingCompletionTime: time.Time{},
		}
	}
	return testApp
}

// : unparam deadcode unused
func createTestAccs(ctx sdk.Ctx, numAccs int, initialCoins sdk.Coins, ak *auth.Keeper) (accs []auth.BaseAccount) {
	for i := 0; i < numAccs; i++ {
		privKey := crypto.Ed25519PrivateKey{}.GenPrivateKey()
		pubKey := privKey.PublicKey()
		addr := sdk.Address(pubKey.Address())
		acc := auth.NewBaseAccountWithAddress(addr)
		acc.Coins = initialCoins
		acc.PubKey = pubKey
		ak.SetAccount(ctx, &acc)
		accs = append(accs, acc)
	}
	return
}

func createTestValidators(ctx sdk.Ctx, numAccs int, valCoins sdk.BigInt, nk *servicersKeeper.Keeper, ak auth.Keeper, kb keys.Keybase) (accs servicersTypes.Validators) {
	ethereum := hex.EncodeToString([]byte{01})
	US := hex.EncodeToString([]byte{01})
	for i := 0; i < numAccs-1; i++ {
		privKey := crypto.Ed25519PrivateKey{}.GenPrivateKey()
		pubKey := privKey.PublicKey()
		addr := sdk.Address(pubKey.Address())
		privKey2 := crypto.Ed25519PrivateKey{}.GenPrivateKey()
		pubKey2 := privKey2.PublicKey()
		addr2 := sdk.Address(pubKey2.Address())
		val := servicersTypes.NewValidator(addr, pubKey, []string{ethereum}, "https://www.google.com:443", valCoins, []string{US}, addr2, servicersTypes.ReportCard{TotalSessions: 0, TotalLatencyScore: sdk.NewDec(0), TotalAvailabilityScore: sdk.NewDec(0), TotalReliabilityScore: sdk.NewDec(0)})
		// set the vals from the data
		nk.SetValidator(ctx, val)
		nk.SetStakedValidatorByChains(ctx, val)
		nk.SetStakedValidatorByGeoZone(ctx, val)
		// ensure there's a signing info entry for the val (used in slashing)
		_, found := nk.GetValidatorSigningInfo(ctx, val.GetAddress())
		if !found {
			signingInfo := servicersTypes.ValidatorSigningInfo{
				Address:     val.GetAddress(),
				StartHeight: ctx.BlockHeight(),
				JailedUntil: time.Unix(0, 0),
				PausedUntil: time.Unix(0, 0),
			}
			nk.SetValidatorSigningInfo(ctx, val.GetAddress(), signingInfo)
		}
		accs = append(accs, val)
	}
	// add self node to it
	kp, er := kb.GetCoinbase()
	if er != nil {
		panic(er)
	}
	val := servicersTypes.NewValidator(sdk.Address(kp.GetAddress()), kp.PublicKey, []string{ethereum}, "https://www.google.com:443", valCoins, []string{US}, kp.GetAddress(), servicersTypes.ReportCard{TotalSessions: 0, TotalLatencyScore: sdk.NewDec(0), TotalAvailabilityScore: sdk.NewDec(0), TotalReliabilityScore: sdk.NewDec(0)})
	// set the vals from the data
	nk.SetValidator(ctx, val)
	nk.SetStakedValidatorByChains(ctx, val)
	nk.SetStakedValidatorByGeoZone(ctx, val)
	// ensure there's a signing info entry for the val (used in slashing)
	_, found := nk.GetValidatorSigningInfo(ctx, val.GetAddress())
	if !found {
		signingInfo := servicersTypes.ValidatorSigningInfo{
			Address:     val.GetAddress(),
			StartHeight: ctx.BlockHeight(),
			JailedUntil: time.Unix(0, 0),
		}
		nk.SetValidatorSigningInfo(ctx, val.GetAddress(), signingInfo)
	}
	accs = append(accs, val)
	// end self node logic
	stakedTokens := sdk.NewInt(int64(numAccs)).Mul(valCoins)
	// take the staked amount and create the corresponding coins object
	stakedCoins := sdk.NewCoins(sdk.NewCoin(nk.StakeDenom(ctx), stakedTokens))
	// check if the staked pool accounts exists
	stakedPool := nk.GetStakedPool(ctx)
	// if the stakedPool is nil
	if stakedPool == nil {
		panic(fmt.Sprintf("%s module account has not been set", servicersTypes.StakedPoolName))
	}
	// add coins if not provided on genesis (there's an option to provide the coins in genesis)
	if stakedPool.GetCoins().IsZero() {
		if err := stakedPool.SetCoins(stakedCoins); err != nil {
			panic(err)
		}
		ak.SetModuleAccount(ctx, stakedPool)
	} else {
		// if it is provided in the genesis file then ensure the two are equal
		if !stakedPool.GetCoins().IsEqual(stakedCoins) {
			panic(fmt.Sprintf("%s module account total does not equal the amount in each validator account", servicersTypes.StakedPoolName))
		}
	}
	return
}

func createTestApps(ctx sdk.Ctx, numAccs int, valCoins sdk.BigInt, ak requestorsKeeper.Keeper, sk auth.Keeper) (accs requestorsTypes.Requestors) {
	ethereum := hex.EncodeToString([]byte{01})
	US := hex.EncodeToString([]byte{01})
	for i := 0; i < numAccs; i++ {
		privKey := crypto.Ed25519PrivateKey{}.GenPrivateKey()
		pubKey := privKey.PublicKey()
		addr := sdk.Address(pubKey.Address())
		app := requestorsTypes.NewRequestor(addr, pubKey, []string{ethereum}, valCoins, []string{US}, 5)
		// set the vals from the data
		// calculate relays
		app.MaxRelays = ak.CalculateRequestorRelays(ctx, app)
		ak.SetRequestor(ctx, app)
		ak.SetStakedRequestor(ctx, app)
		accs = append(accs, app)
	}
	stakedTokens := sdk.NewInt(int64(numAccs)).Mul(valCoins)
	// take the staked amount and create the corresponding coins object
	stakedCoins := sdk.NewCoins(sdk.NewCoin(ak.StakeDenom(ctx), stakedTokens))
	// check if the staked pool accounts exists
	stakedPool := ak.GetStakedPool(ctx)
	// if the stakedPool is nil
	if stakedPool == nil {
		panic(fmt.Sprintf("%s module account has not been set", requestorsTypes.StakedPoolName))
	}
	// add coins if not provided on genesis (there's an option to provide the coins in genesis)
	if stakedPool.GetCoins().IsZero() {
		if err := stakedPool.SetCoins(stakedCoins); err != nil {
			panic(err)
		}
		sk.SetModuleAccount(ctx, stakedPool)
	} else {
		// if it is provided in the genesis file then ensure the two are equal
		if !stakedPool.GetCoins().IsEqual(stakedCoins) {
			panic(fmt.Sprintf("%s module account total does not equal the amount in each app account", requestorsTypes.StakedPoolName))
		}
	}
	return
}

func getRandomPrivateKey() crypto.Ed25519PrivateKey {
	return crypto.Ed25519PrivateKey{}.GenPrivateKey().(crypto.Ed25519PrivateKey)
}

func getRandomPubKey() crypto.Ed25519PublicKey {
	pk := crypto.Ed25519PrivateKey{}.GenPrivateKey()
	return pk.PublicKey().(crypto.Ed25519PublicKey)
}

func getRandomValidatorAddress() sdk.Address {
	return sdk.Address(getRandomPubKey().Address())
}

func simulateRelays(t *testing.T, k Keeper, ctx *sdk.Ctx, maxRelays int) (npk crypto.PublicKey, validHeader types.SessionHeader, keys simulateRelayKeys) {
	npk = getRandomPubKey()
	ethereum := hex.EncodeToString([]byte{01})
	US := hex.EncodeToString([]byte{01})
	clientKey := getRandomPrivateKey()
	validHeader = types.SessionHeader{
		RequestorPubKey:    getTestRequestor().PublicKey.RawString(),
		Chain:              ethereum,
		GeoZone:            US,
		NumServicers:       5,
		SessionBlockHeight: 1,
	}
	logger := log.NewNopLogger()
	types.InitConfig(&types.HostedBlockchains{
		M: make(map[string]types.HostedBlockchain),
	}, &types.HostedGeoZones{
		M: make(map[string]types.GeoZone),
	}, logger, sdk.DefaultTestingViperConfig())

	// NOTE Add a minimum of 5 proofs to memInvoice to be able to create a merkle tree
	for j := 0; j < maxRelays; j++ {
		proof := createProof(getTestRequestorPrivateKey(), clientKey, npk, ethereum, US, j)
		types.SetProof(validHeader, types.RelayEvidence, proof, sdk.NewInt(100000), types.GlobalEvidenceCache)
	}
	mockCtx := new(Ctx)
	mockCtx.On("KVStore", k.storeKey).Return((*ctx).KVStore(k.storeKey))
	mockCtx.On("PrevCtx", validHeader.SessionBlockHeight).Return(*ctx, nil)
	mockCtx.On("Logger").Return((*ctx).Logger())
	keys = simulateRelayKeys{getTestRequestorPrivateKey(), clientKey}
	return
}

func createProof(private, client crypto.PrivateKey, npk crypto.PublicKey, chain string, geoZone string, entropy int) types.Proof {
	aat := types.AAT{
		Version:            "0.0.1",
		RequestorPublicKey: private.PublicKey().RawString(),
		ClientPublicKey:    client.PublicKey().RawString(),
		RequestorSignature: "",
	}
	sig, err := private.Sign(aat.Hash())
	if err != nil {
		panic(err)
	}
	aat.RequestorSignature = hex.EncodeToString(sig)
	proof := types.RelayProof{
		Entropy:            int64(entropy + 1),
		RequestHash:        aat.HashString(), // fake
		SessionBlockHeight: 1,
		ServicerPubKey:     npk.RawString(),
		Blockchain:         chain,
		Token:              aat,
		Signature:          "",
		GeoZone:            geoZone,
		NumServicers:       5,
	}
	clientSig, er := client.Sign(proof.Hash())
	if er != nil {
		panic(er)
	}
	proof.Signature = hex.EncodeToString(clientSig)
	return proof
}

func simulateTestRelays(t *testing.T, k Keeper, ctx *sdk.Ctx, maxTestResults int) (servicerPk, fishermanPk crypto.PublicKey, header types.SessionHeader, keys simulateTestResultKeys) {
	// Generate random public keys for servicer and fisherman
	servicerPk = getRandomPubKey()
	fishermanPk = getRandomPubKey()
	clientKey := getRandomPrivateKey()
	ethereum := hex.EncodeToString([]byte{01})
	US := hex.EncodeToString([]byte{01})

	// Create a sample SessionHeader
	header = types.SessionHeader{
		RequestorPubKey:    getTestRequestor().PublicKey.RawString(),
		Chain:              ethereum,
		GeoZone:            US,
		NumServicers:       5,
		SessionBlockHeight: 1,
	}

	// Create a logger
	logger := log.NewNopLogger()

	// Initialize config
	types.InitConfig(&types.HostedBlockchains{
		M: make(map[string]types.HostedBlockchain),
	}, &types.HostedGeoZones{
		M: make(map[string]types.GeoZone),
	}, logger, sdk.DefaultTestingViperConfig())

	// Add test results to the cache
	for j := 0; j < maxTestResults; j++ {
		testResult := createTestResult(getTestRequestorPrivateKey(), clientKey, servicerPk)
		types.SetTestResult(header, types.FishermanTestEvidence, testResult, servicerPk.Address().Bytes(), types.GlobalTestCache)
	}

	// Mock the context for testing
	mockCtx := new(Ctx)
	mockCtx.On("KVStore", k.storeKey).Return((*ctx).KVStore(k.storeKey))
	mockCtx.On("PrevCtx", header.SessionBlockHeight).Return(*ctx, nil)
	mockCtx.On("Logger").Return((*ctx).Logger())

	// Store keys for later verification
	keys = simulateTestResultKeys{getTestRequestorPrivateKey(), clientKey}
	return
}

func createTestResult(private, client crypto.PrivateKey, npk crypto.PublicKey) types.TestResult {
	aat := types.AAT{
		Version:            "0.0.1",
		RequestorPublicKey: private.PublicKey().RawString(),
		ClientPublicKey:    client.PublicKey().RawString(),
		RequestorSignature: "",
	}
	sig, err := private.Sign(aat.Hash())
	if err != nil {
		panic(err)
	}
	aat.RequestorSignature = hex.EncodeToString(sig)
	test := types.TestResult{
		ServicerAddress: npk.Address().Bytes(),
		Timestamp:       time.Now(),
		Latency:         time.Duration(100 * time.Millisecond),
		IsAvailable:     true,
		IsReliable:      true,
	}
	return test
}

// Ctx is an autogenerated mock type for the Ctx type
type Ctx struct {
	mock.Mock
}

// GetPrevBlockHash provides a mock function with given fields: height
func (_m *Ctx) GetPrevBlockHash(height int64) ([]byte, error) {
	ret := _m.Called(height)
	var r0 []byte
	if rf, ok := ret.Get(0).(func(int64) []byte); ok {
		r0 = rf(height)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}
	var r1 error
	if rf, ok := ret.Get(1).(func(int64) error); ok {
		r1 = rf(height)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

func (_m *Ctx) IsPrevCtx() bool {
	return true
}

// BlockGasMeter provides a mock function with given fields:
func (_m *Ctx) BlockGasMeter() storeTypes.GasMeter {
	ret := _m.Called()

	var r0 storeTypes.GasMeter
	if rf, ok := ret.Get(0).(func() storeTypes.GasMeter); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(storeTypes.GasMeter)
		}
	}

	return r0
}

// BlockHeader provides a mock function with given fields:
func (_m *Ctx) BlockHeader() abcitypes.Header {
	ret := _m.Called()

	var r0 abcitypes.Header
	if rf, ok := ret.Get(0).(func() abcitypes.Header); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(abcitypes.Header)
	}

	return r0
}

// BlockHeight provides a mock function with given fields:
func (_m *Ctx) BlockHeight() int64 {
	ret := _m.Called()

	var r0 int64
	if rf, ok := ret.Get(0).(func() int64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int64)
	}

	return r0
}

// BlockStore provides a mock function with given fields:
func (_m *Ctx) BlockStore() *tmStore.BlockStore {
	ret := _m.Called()

	var r0 *tmStore.BlockStore
	if rf, ok := ret.Get(0).(func() *tmStore.BlockStore); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*tmStore.BlockStore)
		}
	}

	return r0
}

// BlockTime provides a mock function with given fields:
func (_m *Ctx) BlockTime() time.Time {
	ret := _m.Called()

	var r0 time.Time
	if rf, ok := ret.Get(0).(func() time.Time); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(time.Time)
	}

	return r0
}

// CacheContext provides a mock function with given fields:
func (_m *Ctx) CacheContext() (viperTypes.Context, func()) {
	ret := _m.Called()

	var r0 viperTypes.Context
	if rf, ok := ret.Get(0).(func() viperTypes.Context); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(viperTypes.Context)
	}

	var r1 func()
	if rf, ok := ret.Get(1).(func() func()); ok {
		r1 = rf()
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(func())
		}
	}

	return r0, r1
}

// ChainID provides a mock function with given fields:
func (_m *Ctx) ChainID() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// ConsensusParams provides a mock function with given fields:
func (_m *Ctx) ConsensusParams() *abcitypes.ConsensusParams {
	ret := _m.Called()

	var r0 *abcitypes.ConsensusParams
	if rf, ok := ret.Get(0).(func() *abcitypes.ConsensusParams); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*abcitypes.ConsensusParams)
		}
	}

	return r0
}

// Context provides a mock function with given fields:
func (_m *Ctx) Context() context.Context {
	ret := _m.Called()

	var r0 context.Context
	if rf, ok := ret.Get(0).(func() context.Context); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(context.Context)
		}
	}

	return r0
}

// EventManager provides a mock function with given fields:
func (_m *Ctx) EventManager() *viperTypes.EventManager {
	ret := _m.Called()

	var r0 *viperTypes.EventManager
	if rf, ok := ret.Get(0).(func() *viperTypes.EventManager); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*viperTypes.EventManager)
		}
	}

	return r0
}

// GasMeter provides a mock function with given fields:
func (_m *Ctx) GasMeter() storeTypes.GasMeter {
	ret := _m.Called()

	var r0 storeTypes.GasMeter
	if rf, ok := ret.Get(0).(func() storeTypes.GasMeter); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(storeTypes.GasMeter)
		}
	}

	return r0
}

// IsCheckTx provides a mock function with given fields:
func (_m *Ctx) IsCheckTx() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// IsZero provides a mock function with given fields:
func (_m *Ctx) IsZero() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// IsZero provides a mock function with given fields:
func (_m *Ctx) IsAfterUpgradeHeight() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// IsZero provides a mock function with given fields:
func (_m *Ctx) IsOnUpgradeHeight() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// KVStore provides a mock function with given fields: key
func (_m *Ctx) KVStore(key storeTypes.StoreKey) storeTypes.KVStore {
	ret := _m.Called(key)

	var r0 storeTypes.KVStore
	if rf, ok := ret.Get(0).(func(storeTypes.StoreKey) storeTypes.KVStore); ok {
		r0 = rf(key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(storeTypes.KVStore)
		}
	}

	return r0
}

// Logger provides a mock function with given fields:
func (_m *Ctx) Logger() log.Logger {
	ret := _m.Called()

	var r0 log.Logger
	if rf, ok := ret.Get(0).(func() log.Logger); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(log.Logger)
		}
	}

	return r0
}

// MinGasPrices provides a mock function with given fields:
func (_m *Ctx) MinGasPrices() viperTypes.DecCoins {
	ret := _m.Called()

	var r0 viperTypes.DecCoins
	if rf, ok := ret.Get(0).(func() viperTypes.DecCoins); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(viperTypes.DecCoins)
		}
	}

	return r0
}

// MultiStore provides a mock function with given fields:
func (_m *Ctx) MultiStore() storeTypes.MultiStore {
	ret := _m.Called()

	var r0 storeTypes.MultiStore
	if rf, ok := ret.Get(0).(func() storeTypes.MultiStore); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(storeTypes.MultiStore)
		}
	}

	return r0
}

// MustGetPrevCtx provides a mock function with given fields: height
func (_m *Ctx) MustGetPrevCtx(height int64) viperTypes.Context {
	ret := _m.Called(height)

	var r0 viperTypes.Context
	if rf, ok := ret.Get(0).(func(int64) viperTypes.Context); ok {
		r0 = rf(height)
	} else {
		r0 = ret.Get(0).(viperTypes.Context)
	}

	return r0
}

// PrevCtx provides a mock function with given fields: height
func (_m *Ctx) PrevCtx(height int64) (viperTypes.Context, error) {
	ret := _m.Called(height)

	var r0 viperTypes.Context
	if rf, ok := ret.Get(0).(func(int64) viperTypes.Context); ok {
		r0 = rf(height)
	} else {
		r0 = ret.Get(0).(viperTypes.Context)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int64) error); ok {
		r1 = rf(height)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TransientStore provides a mock function with given fields: key
func (_m *Ctx) TransientStore(key storeTypes.StoreKey) storeTypes.KVStore {
	ret := _m.Called(key)

	var r0 storeTypes.KVStore
	if rf, ok := ret.Get(0).(func(storeTypes.StoreKey) storeTypes.KVStore); ok {
		r0 = rf(key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(storeTypes.KVStore)
		}
	}

	return r0
}

// TxBytes provides a mock function with given fields:
func (_m *Ctx) TxBytes() []byte {
	ret := _m.Called()

	var r0 []byte
	if rf, ok := ret.Get(0).(func() []byte); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	return r0
}

// Value provides a mock function with given fields: key
func (_m *Ctx) Value(key interface{}) interface{} {
	ret := _m.Called(key)

	var r0 interface{}
	if rf, ok := ret.Get(0).(func(interface{}) interface{}); ok {
		r0 = rf(key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	return r0
}

// VoteInfos provides a mock function with given fields:
func (_m *Ctx) VoteInfos() []abcitypes.VoteInfo {
	ret := _m.Called()

	var r0 []abcitypes.VoteInfo
	if rf, ok := ret.Get(0).(func() []abcitypes.VoteInfo); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]abcitypes.VoteInfo)
		}
	}

	return r0
}

// WithBlockGasMeter provides a mock function with given fields: meter
func (_m *Ctx) WithBlockGasMeter(meter storeTypes.GasMeter) viperTypes.Context {
	ret := _m.Called(meter)

	var r0 viperTypes.Context
	if rf, ok := ret.Get(0).(func(storeTypes.GasMeter) viperTypes.Context); ok {
		r0 = rf(meter)
	} else {
		r0 = ret.Get(0).(viperTypes.Context)
	}

	return r0
}

// WithBlockHeader provides a mock function with given fields: header
func (_m *Ctx) WithBlockHeader(header abcitypes.Header) viperTypes.Context {
	ret := _m.Called(header)

	var r0 viperTypes.Context
	if rf, ok := ret.Get(0).(func(abcitypes.Header) viperTypes.Context); ok {
		r0 = rf(header)
	} else {
		r0 = ret.Get(0).(viperTypes.Context)
	}

	return r0
}

// WithBlockHeight provides a mock function with given fields: height
func (_m *Ctx) WithBlockHeight(height int64) viperTypes.Context {
	ret := _m.Called(height)

	var r0 viperTypes.Context
	if rf, ok := ret.Get(0).(func(int64) viperTypes.Context); ok {
		r0 = rf(height)
	} else {
		r0 = ret.Get(0).(viperTypes.Context)
	}

	return r0
}

// WithBlockStore provides a mock function with given fields: bs
func (_m *Ctx) WithBlockStore(bs *tmStore.BlockStore) viperTypes.Context {
	ret := _m.Called(bs)

	var r0 viperTypes.Context
	if rf, ok := ret.Get(0).(func(*tmStore.BlockStore) viperTypes.Context); ok {
		r0 = rf(bs)
	} else {
		r0 = ret.Get(0).(viperTypes.Context)
	}

	return r0
}

// WithBlockTime provides a mock function with given fields: newTime
func (_m *Ctx) WithBlockTime(newTime time.Time) viperTypes.Context {
	ret := _m.Called(newTime)

	var r0 viperTypes.Context
	if rf, ok := ret.Get(0).(func(time.Time) viperTypes.Context); ok {
		r0 = rf(newTime)
	} else {
		r0 = ret.Get(0).(viperTypes.Context)
	}

	return r0
}

// WithChainID provides a mock function with given fields: chainID
func (_m *Ctx) WithChainID(chainID string) viperTypes.Context {
	ret := _m.Called(chainID)

	var r0 viperTypes.Context
	if rf, ok := ret.Get(0).(func(string) viperTypes.Context); ok {
		r0 = rf(chainID)
	} else {
		r0 = ret.Get(0).(viperTypes.Context)
	}

	return r0
}

// WithConsensusParams provides a mock function with given fields: params
func (_m *Ctx) WithConsensusParams(params *abcitypes.ConsensusParams) viperTypes.Context {
	ret := _m.Called(params)

	var r0 viperTypes.Context
	if rf, ok := ret.Get(0).(func(*abcitypes.ConsensusParams) viperTypes.Context); ok {
		r0 = rf(params)
	} else {
		r0 = ret.Get(0).(viperTypes.Context)
	}

	return r0
}

// WithContext provides a mock function with given fields: ctx
func (_m *Ctx) WithContext(ctx context.Context) viperTypes.Context {
	ret := _m.Called(ctx)

	var r0 viperTypes.Context
	if rf, ok := ret.Get(0).(func(context.Context) viperTypes.Context); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(viperTypes.Context)
	}

	return r0
}

// WithEventManager provides a mock function with given fields: em
func (_m *Ctx) WithEventManager(em *viperTypes.EventManager) viperTypes.Context {
	ret := _m.Called(em)

	var r0 viperTypes.Context
	if rf, ok := ret.Get(0).(func(*viperTypes.EventManager) viperTypes.Context); ok {
		r0 = rf(em)
	} else {
		r0 = ret.Get(0).(viperTypes.Context)
	}

	return r0
}

// WithGasMeter provides a mock function with given fields: meter
func (_m *Ctx) WithGasMeter(meter storeTypes.GasMeter) viperTypes.Context {
	ret := _m.Called(meter)

	var r0 viperTypes.Context
	if rf, ok := ret.Get(0).(func(storeTypes.GasMeter) viperTypes.Context); ok {
		r0 = rf(meter)
	} else {
		r0 = ret.Get(0).(viperTypes.Context)
	}

	return r0
}

// WithIsCheckTx provides a mock function with given fields: isCheckTx
func (_m *Ctx) WithIsCheckTx(isCheckTx bool) viperTypes.Context {
	ret := _m.Called(isCheckTx)

	var r0 viperTypes.Context
	if rf, ok := ret.Get(0).(func(bool) viperTypes.Context); ok {
		r0 = rf(isCheckTx)
	} else {
		r0 = ret.Get(0).(viperTypes.Context)
	}

	return r0
}

// WithLogger provides a mock function with given fields: logger
func (_m *Ctx) WithLogger(logger log.Logger) viperTypes.Context {
	ret := _m.Called(logger)

	var r0 viperTypes.Context
	if rf, ok := ret.Get(0).(func(log.Logger) viperTypes.Context); ok {
		r0 = rf(logger)
	} else {
		r0 = ret.Get(0).(viperTypes.Context)
	}

	return r0
}

// WithMinGasPrices provides a mock function with given fields: gasPrices
func (_m *Ctx) WithMinGasPrices(gasPrices viperTypes.DecCoins) viperTypes.Context {
	ret := _m.Called(gasPrices)

	var r0 viperTypes.Context
	if rf, ok := ret.Get(0).(func(viperTypes.DecCoins) viperTypes.Context); ok {
		r0 = rf(gasPrices)
	} else {
		r0 = ret.Get(0).(viperTypes.Context)
	}

	return r0
}

// WithMultiStore provides a mock function with given fields: ms
func (_m *Ctx) WithMultiStore(ms storeTypes.MultiStore) viperTypes.Context {
	ret := _m.Called(ms)

	var r0 viperTypes.Context
	if rf, ok := ret.Get(0).(func(storeTypes.MultiStore) viperTypes.Context); ok {
		r0 = rf(ms)
	} else {
		r0 = ret.Get(0).(viperTypes.Context)
	}

	return r0
}

// WithProposer provides a mock function with given fields: addr
func (_m *Ctx) WithProposer(addr viperTypes.Address) viperTypes.Context {
	ret := _m.Called(addr)

	var r0 viperTypes.Context
	if rf, ok := ret.Get(0).(func(viperTypes.Address) viperTypes.Context); ok {
		r0 = rf(addr)
	} else {
		r0 = ret.Get(0).(viperTypes.Context)
	}

	return r0
}

// WithTxBytes provides a mock function with given fields: txBytes
func (_m *Ctx) WithTxBytes(txBytes []byte) viperTypes.Context {
	ret := _m.Called(txBytes)

	var r0 viperTypes.Context
	if rf, ok := ret.Get(0).(func([]byte) viperTypes.Context); ok {
		r0 = rf(txBytes)
	} else {
		r0 = ret.Get(0).(viperTypes.Context)
	}

	return r0
}

// WithValue provides a mock function with given fields: key, value
func (_m *Ctx) WithValue(key interface{}, value interface{}) viperTypes.Context {
	ret := _m.Called(key, value)

	var r0 viperTypes.Context
	if rf, ok := ret.Get(0).(func(interface{}, interface{}) viperTypes.Context); ok {
		r0 = rf(key, value)
	} else {
		r0 = ret.Get(0).(viperTypes.Context)
	}

	return r0
}

// WithValue provides a mock function with given fields: key, value
func (_m *Ctx) AppVersion() string {
	return ""
}

// WithVoteInfos provides a mock function with given fields: voteInfo
func (_m *Ctx) WithVoteInfos(voteInfo []abcitypes.VoteInfo) viperTypes.Context {
	ret := _m.Called(voteInfo)

	var r0 viperTypes.Context
	if rf, ok := ret.Get(0).(func([]abcitypes.VoteInfo) viperTypes.Context); ok {
		r0 = rf(voteInfo)
	} else {
		r0 = ret.Get(0).(viperTypes.Context)
	}

	return r0
}

func (_m *Ctx) ClearGlobalCache() {
	_m.Called()
}

func (_m *Ctx) BlockHash(cdc *codec.Codec, _ int64) ([]byte, error) {
	ret := _m.Called()

	var r0 []byte
	if rf, ok := ret.Get(0).(func() []byte); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).([]byte)
	}

	var r1 error
	if rf1, ok := ret.Get(1).(func() error); ok {
		r1 = rf1()
	} else {
		r1 = ret.Get(1).(error)
	}

	return r0, r1
}
