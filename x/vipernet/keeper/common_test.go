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
	"github.com/vipernet-xyz/viper-network/crypto"
	"github.com/vipernet-xyz/viper-network/crypto/keys"
	"github.com/vipernet-xyz/viper-network/store"
	storeTypes "github.com/vipernet-xyz/viper-network/store/types"
	sdk "github.com/vipernet-xyz/viper-network/types"
	viperTypes "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/types/module"
	"github.com/vipernet-xyz/viper-network/x/authentication"
	"github.com/vipernet-xyz/viper-network/x/governance"
	governanceTypes "github.com/vipernet-xyz/viper-network/x/governance/types"
	providers "github.com/vipernet-xyz/viper-network/x/providers"
	providersKeeper "github.com/vipernet-xyz/viper-network/x/providers/keeper"
	providersTypes "github.com/vipernet-xyz/viper-network/x/providers/types"
	"github.com/vipernet-xyz/viper-network/x/servicers"
	servicersKeeper "github.com/vipernet-xyz/viper-network/x/servicers/keeper"
	servicersTypes "github.com/vipernet-xyz/viper-network/x/servicers/types"
	"github.com/vipernet-xyz/viper-network/x/vipernet/types"

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
		authentication.AppModuleBasic{},
		governance.AppModuleBasic{},
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

func NewTestKeybase() keys.Keybase {
	return keys.NewInMemory()
}

// create a codec used only for testing
func makeTestCodec() *codec.Codec {
	var cdc = codec.NewCodec(types2.NewInterfaceRegistry())
	authentication.RegisterCodec(cdc)
	governance.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	crypto.RegisterAmino(cdc.AminoCodec().Amino)
	return cdc
}

// : deadcode unused
func createTestInput(t *testing.T, isCheckTx bool) (sdk.Ctx, []servicersTypes.Validator, []providersTypes.Provider, []authentication.BaseAccount, Keeper, map[string]*sdk.KVStoreKey, keys.Keybase) {
	sdk.VbCCache = sdk.NewCache(1)
	initPower := int64(100000000000)
	nAccs := int64(5)
	kb := NewTestKeybase()
	_, err := kb.Create("test")
	assert.Nil(t, err)
	cb, err := kb.GetCoinbase()
	assert.Nil(t, err)
	addr := tmtypes.Address(cb.GetAddress())
	pk, err := kb.ExportPrivateKeyObject(cb.GetAddress(), "test")
	assert.Nil(t, err)
	types.InitPVKeyFile(privval.FilePVKey{
		Address: addr,
		PubKey:  cb.PublicKey,
		PrivKey: pk,
	})
	keyAcc := sdk.NewKVStoreKey(authentication.StoreKey)
	keyParams := sdk.ParamsKey
	tkeyParams := sdk.ParamsTKey
	servicersKey := sdk.NewKVStoreKey(servicersTypes.StoreKey)
	providersKey := sdk.NewKVStoreKey(providersTypes.StoreKey)
	viperKey := sdk.NewKVStoreKey(types.StoreKey)

	keys := make(map[string]*sdk.KVStoreKey)
	keys["params"] = keyParams
	keys["pos"] = servicersKey
	keys["provider"] = providersKey

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db, false, 5000000)
	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(servicersKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(providersKey, sdk.StoreTypeIAVL, db)
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
		Height: 976,
		Time:   time.Time{},
		LastBlockId: abci.BlockID{
			Hash: types.Hash([]byte("fake")),
		},
	})
	cdc := makeTestCodec()

	maccPerms := map[string][]string{
		authentication.FeeCollectorName: nil,
		providersTypes.StakedPoolName:   {authentication.Burner, authentication.Staking, authentication.Minter},
		servicersTypes.StakedPoolName:   {authentication.Burner, authentication.Staking},
		governanceTypes.DAOAccountName:  {authentication.Burner, authentication.Staking},
	}

	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authentication.NewModuleAddress(acc).String()] = true
	}
	valTokens := sdk.TokensFromConsensusPower(initPower)

	ethereum := hex.EncodeToString([]byte{01})

	hb := types.HostedBlockchains{
		M: map[string]types.HostedBlockchain{ethereum: {
			ID:  ethereum,
			URL: "https://www.google.com:443",
		}},
	}
	types.InitConfig(&hb, log.NewTMLogger(os.Stdout), sdk.DefaultTestingViperConfig())
	authSubspace := sdk.NewSubspace(authentication.DefaultParamspace)
	servicersSubspace := sdk.NewSubspace(servicersTypes.DefaultParamspace)
	providerSubspace := sdk.NewSubspace(providersTypes.DefaultParamspace)
	viperSubspace := sdk.NewSubspace(types.DefaultParamspace)
	ak := authentication.NewKeeper(cdc, keyAcc, authSubspace, maccPerms)
	nk := servicersKeeper.NewKeeper(cdc, servicersKey, ak, servicersSubspace, servicersTypes.ModuleName)
	providerk := providersKeeper.NewKeeper(cdc, providersKey, nk, ak, nil, providerSubspace, providersTypes.ModuleName)
	providerk.SetProvider(ctx, getTestProvider())
	keeper := NewKeeper(viperKey, cdc, ak, nk, providerk, &hb, viperSubspace)
	providerk.ViperKeeper = keeper
	assert.Nil(t, err)
	moduleManager := module.NewManager(
		authentication.NewAppModule(ak),
		servicers.NewAppModule(nk),
		providers.NewAppModule(providerk),
	)
	genesisState := ModuleBasics.DefaultGenesis()
	moduleManager.InitGenesis(ctx, genesisState)
	initialCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultStakeDenom, valTokens))
	accs := createTestAccs(ctx, int(nAccs), initialCoins, &ak)
	ap := createTestProviders(ctx, int(nAccs), sdk.NewIntFromBigInt(new(big.Int).SetUint64(math.MaxUint64)), providerk, ak)
	vals := createTestValidators(ctx, int(nAccs), sdk.ZeroInt(), &nk, ak, kb)
	providerk.SetParams(ctx, providersTypes.DefaultParams())
	nk.SetParams(ctx, servicersTypes.DefaultParams())
	defaultViperParams := types.DefaultParams()
	defaultViperParams.SupportedBlockchains = []string{getTestSupportedBlockchain()}
	keeper.SetParams(ctx, defaultViperParams)
	return ctx, vals, ap, accs, keeper, keys, kb
}

var (
	testProvider           providersTypes.Provider
	testProviderPrivateKey crypto.PrivateKey
	testSupportedChain     string
)

func getTestSupportedBlockchain() string {
	if testSupportedChain == "" {
		testSupportedChain = hex.EncodeToString([]byte{01})
	}
	return testSupportedChain
}

func getTestProviderPrivateKey() crypto.PrivateKey {
	if testProviderPrivateKey == nil {
		testProviderPrivateKey = getRandomPrivateKey()
	}
	return testProviderPrivateKey
}

func getTestProvider() providersTypes.Provider {
	if testProvider.Address == nil {
		pk := getTestProviderPrivateKey().PublicKey()
		testProvider = providersTypes.Provider{
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
	return testProvider
}

// : unparam deadcode unused
func createTestAccs(ctx sdk.Ctx, numAccs int, initialCoins sdk.Coins, ak *authentication.Keeper) (accs []authentication.BaseAccount) {
	for i := 0; i < numAccs; i++ {
		privKey := crypto.Ed25519PrivateKey{}.GenPrivateKey()
		pubKey := privKey.PublicKey()
		addr := sdk.Address(pubKey.Address())
		acc := authentication.NewBaseAccountWithAddress(addr)
		acc.Coins = initialCoins
		acc.PubKey = pubKey
		ak.SetAccount(ctx, &acc)
		accs = append(accs, acc)
	}
	return
}

func createTestValidators(ctx sdk.Ctx, numAccs int, valCoins sdk.BigInt, nk *servicersKeeper.Keeper, ak authentication.Keeper, kb keys.Keybase) (accs servicersTypes.Validators) {
	ethereum := hex.EncodeToString([]byte{01})
	for i := 0; i < numAccs-1; i++ {
		privKey := crypto.Ed25519PrivateKey{}.GenPrivateKey()
		pubKey := privKey.PublicKey()
		addr := sdk.Address(pubKey.Address())
		privKey2 := crypto.Ed25519PrivateKey{}.GenPrivateKey()
		pubKey2 := privKey2.PublicKey()
		addr2 := sdk.Address(pubKey2.Address())
		val := servicersTypes.NewValidator(addr, pubKey, []string{ethereum}, "https://www.google.com:443", valCoins, addr2)
		// set the vals from the data
		nk.SetValidator(ctx, val)
		nk.SetStakedValidatorByChains(ctx, val)
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
	}
	// add self servicer to it
	kp, er := kb.GetCoinbase()
	if er != nil {
		panic(er)
	}
	val := servicersTypes.NewValidator(sdk.Address(kp.GetAddress()), kp.PublicKey, []string{ethereum}, "https://www.google.com:443", valCoins, kp.GetAddress())
	// set the vals from the data
	nk.SetValidator(ctx, val)
	nk.SetStakedValidatorByChains(ctx, val)
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
	// end self servicer logic
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

func createTestProviders(ctx sdk.Ctx, numAccs int, valCoins sdk.BigInt, ak providersKeeper.Keeper, sk authentication.Keeper) (accs providersTypes.Providers) {
	ethereum := hex.EncodeToString([]byte{01})
	for i := 0; i < numAccs; i++ {
		privKey := crypto.Ed25519PrivateKey{}.GenPrivateKey()
		pubKey := privKey.PublicKey()
		addr := sdk.Address(pubKey.Address())
		provider := providersTypes.NewProvider(addr, pubKey, []string{ethereum}, valCoins)
		// set the vals from the data
		// calculate relays
		provider.MaxRelays = ak.CalculateProviderRelays(ctx, provider)
		ak.SetProvider(ctx, provider)
		ak.SetStakedProvider(ctx, provider)
		accs = append(accs, provider)
	}
	stakedTokens := sdk.NewInt(int64(numAccs)).Mul(valCoins)
	// take the staked amount and create the corresponding coins object
	stakedCoins := sdk.NewCoins(sdk.NewCoin(ak.StakeDenom(ctx), stakedTokens))
	// check if the staked pool accounts exists
	stakedPool := ak.GetStakedPool(ctx)
	// if the stakedPool is nil
	if stakedPool == nil {
		panic(fmt.Sprintf("%s module account has not been set", providersTypes.StakedPoolName))
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
			panic(fmt.Sprintf("%s module account total does not equal the amount in each provider account", providersTypes.StakedPoolName))
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
	clientKey := getRandomPrivateKey()
	validHeader = types.SessionHeader{
		ProviderPubKey:     getTestProvider().PublicKey.RawString(),
		Chain:              ethereum,
		SessionBlockHeight: 1,
	}
	logger := log.NewNopLogger()
	types.InitConfig(&types.HostedBlockchains{
		M: make(map[string]types.HostedBlockchain),
	}, logger, sdk.DefaultTestingViperConfig())

	// NOTE Add a minimum of 5 proofs to memInvoice to be able to create a merkle tree
	for j := 0; j < maxRelays; j++ {
		proof := createProof(getTestProviderPrivateKey(), clientKey, npk, ethereum, j)
		types.SetProof(validHeader, types.RelayEvidence, proof, sdk.NewInt(100000))
	}
	mockCtx := new(Ctx)
	mockCtx.On("KVStore", k.storeKey).Return((*ctx).KVStore(k.storeKey))
	mockCtx.On("PrevCtx", validHeader.SessionBlockHeight).Return(*ctx, nil)
	mockCtx.On("Logger").Return((*ctx).Logger())
	keys = simulateRelayKeys{getTestProviderPrivateKey(), clientKey}
	return
}
func createProof(private, client crypto.PrivateKey, npk crypto.PublicKey, chain string, entropy int) types.Proof {
	aat := types.AAT{
		Version:           "0.0.1",
		ProviderPublicKey: private.PublicKey().RawString(),
		ClientPublicKey:   client.PublicKey().RawString(),
		ProviderSignature: "",
	}
	sig, err := private.Sign(aat.Hash())
	if err != nil {
		panic(err)
	}
	aat.ProviderSignature = hex.EncodeToString(sig)
	proof := types.RelayProof{
		Entropy:            int64(entropy + 1),
		RequestHash:        aat.HashString(), // fake
		SessionBlockHeight: 1,
		ServicerPubKey:     npk.RawString(),
		Blockchain:         chain,
		Token:              aat,
		Signature:          "",
	}
	clientSig, er := client.Sign(proof.Hash())
	if er != nil {
		panic(er)
	}
	proof.Signature = hex.EncodeToString(clientSig)
	return proof
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
