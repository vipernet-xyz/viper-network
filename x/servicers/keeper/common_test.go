package keeper

import (
	"encoding/hex"
	"math/rand"
	"testing"
	"time"

	types2 "github.com/vipernet-xyz/viper-network/codec/types"

	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	"github.com/vipernet-xyz/viper-network/types/module"
	"github.com/vipernet-xyz/viper-network/x/governance"
	requestors "github.com/vipernet-xyz/viper-network/x/requestors"
	"github.com/vipernet-xyz/viper-network/x/servicers/exported"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/vipernet-xyz/viper-network/codec"
	"github.com/vipernet-xyz/viper-network/store"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/authentication"
	govKeeper "github.com/vipernet-xyz/viper-network/x/governance/keeper"
	govTypes "github.com/vipernet-xyz/viper-network/x/governance/types"
	requestorsKeeper "github.com/vipernet-xyz/viper-network/x/requestors/keeper"
	requestorsTypes "github.com/vipernet-xyz/viper-network/x/requestors/types"
	"github.com/vipernet-xyz/viper-network/x/servicers/types"
)

var (
	ModuleBasics = module.NewBasicManager(
		authentication.AppModuleBasic{},
		requestors.AppModuleBasic{},
		governance.AppModuleBasic{},
	)
)

// : deadcode unused
// create a codec used only for testing
func makeTestCodec() *codec.Codec {
	var cdc = codec.NewCodec(types2.NewInterfaceRegistry())
	authentication.RegisterCodec(cdc)
	governance.RegisterCodec(cdc)
	requestorsTypes.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	crypto.RegisterAmino(cdc.AminoCodec().Amino)
	return cdc
}

type MockViperKeeper struct{}

func (m MockViperKeeper) ClearSessionCache() {
	return
}

var _ types.ViperKeeper = MockViperKeeper{}

// : deadcode unused
func createTestInput(t *testing.T, isCheckTx bool) (sdk.Ctx, []authentication.Account, Keeper) {
	initPower := int64(100000000000)
	nAccs := int64(4)
	keyAcc := sdk.NewKVStoreKey(authentication.StoreKey)
	keyParams := sdk.ParamsKey
	tkeyParams := sdk.ParamsTKey
	keyPOS := sdk.NewKVStoreKey(types.ModuleName)
	requestorsKey := sdk.NewKVStoreKey(requestorsTypes.StoreKey)
	govKey := sdk.NewKVStoreKey(govTypes.StoreKey)
	dKey := sdk.NewKVStoreKey("DiscountKey")
	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db, false, 5000000)
	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyPOS, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(requestorsKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(govKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	ms.MountStoreWithDB(dKey, sdk.StoreTypeDB, db)
	err := ms.LoadLatestVersion()
	require.Nil(t, err)

	ctx := sdk.NewContext(ms, abci.Header{ChainID: "test-chain"}, isCheckTx, log.NewNopLogger()).WithAppVersion("0.0.0")
	ctx = ctx.WithConsensusParams(
		&abci.ConsensusParams{
			Validator: &abci.ValidatorParams{
				PubKeyTypes: []string{tmtypes.ABCIPubKeyTypeEd25519},
			},
		},
	)
	cdc := makeTestCodec()

	maccPerms := map[string][]string{
		authentication.FeeCollectorName: nil,
		requestorsTypes.StakedPoolName:  {authentication.Burner, authentication.Staking, authentication.Minter},
		types.StakedPoolName:            {authentication.Burner, authentication.Staking, authentication.Minter},
		types.ModuleName:                {authentication.Burner, authentication.Staking, authentication.Minter},
		govTypes.DAOAccountName:         {authentication.Burner, authentication.Staking},
	}
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authentication.NewModuleAddress(acc).String()] = true
	}
	valTokens := sdk.TokensFromConsensusPower(initPower)
	accSubspace := sdk.NewSubspace(authentication.DefaultParamspace)
	posSubspace := sdk.NewSubspace(DefaultParamspace)
	ak := authentication.NewKeeper(cdc, keyAcc, accSubspace, maccPerms)
	requestorSubspace := sdk.NewSubspace(requestorsTypes.DefaultParamspace)
	requestorKeeper := requestorsKeeper.NewKeeper(cdc, requestorsKey, nil, ak, nil, requestorSubspace, requestorsTypes.ModuleName)
	govKeeper := govKeeper.NewKeeper(cdc, govKey, tkeyParams, dKey, govTypes.ModuleName, ak)
	keeper := NewKeeper(cdc, keyPOS, ak, nil, govKeeper, posSubspace, "pos")
	requestorKeeper.POSKeeper = keeper
	requestorKeeper.ViperKeeper = MockViperKeeper{}
	requestorKeeper.SetRequestor(ctx, getTestRequestor())
	keeper.ViperKeeper = MockViperKeeper{}
	keeper.RequestorKeeper = requestorKeeper
	moduleManager := module.NewManager(
		authentication.NewAppModule(ak),
	)
	genesisState := ModuleBasics.DefaultGenesis()
	moduleManager.InitGenesis(ctx, genesisState)
	initialCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultStakeDenom, valTokens))
	accs := createTestAccs(ctx, int(nAccs), initialCoins, &ak)
	requestorKeeper.SetParams(ctx, requestorsTypes.DefaultParams())
	govKeeper.SetParams(ctx, govTypes.DefaultParams())
	params := types.DefaultParams()
	keeper.SetParams(ctx, params)
	return ctx, accs, keeper
}

// : unparam deadcode unused
func createTestAccs(ctx sdk.Ctx, numAccs int, initialCoins sdk.Coins, ak *authentication.Keeper) (accs []authentication.Account) {
	for i := 0; i < numAccs; i++ {
		privKey := crypto.GenerateEd25519PrivKey()
		pubKey := privKey.PublicKey()
		addr := sdk.Address(pubKey.Address())
		acc := authentication.NewBaseAccountWithAddress(addr)
		acc.Coins = initialCoins
		acc.PubKey = pubKey
		ak.SetAccount(ctx, &acc)
		accs = append(accs, &acc)
	}
	return
}

func addMintedCoinsToModule(t *testing.T, ctx sdk.Ctx, k *Keeper, module string) {
	coins := sdk.NewCoins(sdk.NewCoin(k.StakeDenom(ctx), sdk.NewInt(100000000000)))
	mintErr := k.AccountKeeper.MintCoins(ctx, module, coins.Add(coins))
	if mintErr != nil {
		t.Fail()
	}
}

func sendFromModuleToAccount(t *testing.T, ctx sdk.Ctx, k *Keeper, module string, address sdk.Address, amount sdk.BigInt) {
	coins := sdk.NewCoins(sdk.NewCoin(k.StakeDenom(ctx), amount))
	err := k.AccountKeeper.SendCoinsFromModuleToAccount(ctx, module, sdk.Address(address), coins)
	if err != nil {
		t.Fail()
	}
}

func getRandomPubKey() crypto.Ed25519PublicKey {
	var pub crypto.Ed25519PublicKey
	_, err := rand.Read(pub[:])
	if err != nil {
		_ = err
	}
	return pub
}

func getRandomValidatorAddress() sdk.Address {
	return sdk.Address(getRandomPubKey().Address())
}

func getValidator() types.Validator {
	pub := getRandomPubKey()
	zeroDec := sdk.NewDec(0)
	return types.Validator{
		Address:       sdk.Address(pub.Address()),
		StakedTokens:  sdk.NewInt(100000000000),
		PublicKey:     pub,
		Jailed:        false,
		Paused:        false,
		Status:        sdk.Staked,
		ServiceURL:    "https://www.google.com:443",
		Chains:        []string{"0001", "0002", "FFFF"},
		GeoZone:       []string{"0001"},
		OutputAddress: sdk.Address{},
		ReportCard: types.ReportCard{
			TotalSessions:          0,
			TotalLatencyScore:      zeroDec,
			TotalAvailabilityScore: zeroDec,
			TotalReliabilityScore:  zeroDec,
		},
		UnstakingCompletionTime: time.Time{},
	}
}

func getStakedValidator() types.Validator {
	return getValidator()
}

func getUnstakedValidator() types.Validator {
	v := getValidator()
	return v.UpdateStatus(sdk.Unstaked)
}

func getUnstakingValidator() types.Validator {
	v := getValidator()
	return v.UpdateStatus(sdk.Unstaking)
}

func modifyFn(i *int) func(index int64, Validator exported.ValidatorI) (stop bool) {
	return func(index int64, validator exported.ValidatorI) (stop bool) {
		val := validator.(types.Validator)
		val.StakedTokens = sdk.NewInt(100)
		if index == 1 {
			stop = true
		}
		*i++
		return
	}
}

var (
	testRequestor           requestorsTypes.Requestor
	testRequestorPrivateKey crypto.PrivateKey
	testSupportedChain      string
	testSupportedGeoZone    string
)

func getTestRequestorPrivateKey() crypto.PrivateKey {
	if testRequestorPrivateKey == nil {
		testRequestorPrivateKey = getRandomPrivateKey()
	}
	return testRequestorPrivateKey
}
func getRandomPrivateKey() crypto.Ed25519PrivateKey {
	return crypto.Ed25519PrivateKey{}.GenPrivateKey().(crypto.Ed25519PrivateKey)
}
func getTestSupportedBlockchain() string {
	if testSupportedChain == "" {
		testSupportedChain = hex.EncodeToString([]byte{01})
	}
	return testSupportedChain
}

func getTestSupportedGeoZones() string {
	if testSupportedGeoZone == "" {
		testSupportedGeoZone = hex.EncodeToString([]byte{01})
	}
	return testSupportedGeoZone
}

func getTestRequestor() requestorsTypes.Requestor {
	if testRequestor.Address == nil {
		pk := getTestRequestorPrivateKey().PublicKey()
		testRequestor = requestorsTypes.Requestor{
			Address:                 sdk.Address(pk.Address()),
			PublicKey:               pk,
			Jailed:                  false,
			Status:                  2,
			Chains:                  []string{getTestSupportedBlockchain()},
			GeoZones:                []string{getTestSupportedGeoZones()},
			StakedTokens:            sdk.NewInt(100000000000),
			MaxRelays:               sdk.NewInt(100000000000),
			NumServicers:            5,
			UnstakingCompletionTime: time.Time{},
		}
	}
	return testRequestor
}
