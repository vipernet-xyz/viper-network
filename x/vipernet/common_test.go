package vipernet

import (
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	types2 "github.com/vipernet-xyz/viper-network/codec/types"

	"github.com/vipernet-xyz/viper-network/codec"
	"github.com/vipernet-xyz/viper-network/crypto"
	"github.com/vipernet-xyz/viper-network/crypto/keys"
	"github.com/vipernet-xyz/viper-network/store"
	sdk "github.com/vipernet-xyz/viper-network/types"
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
	keep "github.com/vipernet-xyz/viper-network/x/vipernet/keeper"
	"github.com/vipernet-xyz/viper-network/x/vipernet/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"
)

var (
	ModuleBasics = module.NewBasicManager(
		authentication.AppModuleBasic{},
		governance.AppModuleBasic{},
	)
)

func NewTestKeybase() keys.Keybase {
	return keys.NewInMemory()
}

// : deadcode unused
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
func createTestInput(t *testing.T, isCheckTx bool) (sdk.Ctx, servicersKeeper.Keeper, providersKeeper.Keeper, keep.Keeper, keys.Keybase) {
	initPower := int64(100000000000)
	nAccs := int64(5)

	keyAcc := sdk.NewKVStoreKey(authentication.StoreKey)
	keyParams := sdk.ParamsKey
	tkeyParams := sdk.ParamsTKey
	servicersKey := sdk.NewKVStoreKey(servicersTypes.StoreKey)
	providersKey := sdk.NewKVStoreKey(providersTypes.StoreKey)
	viperKey := sdk.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db, false, 5000000)
	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(servicersKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(providersKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(viperKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	err := ms.LoadLatestVersion()
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
		Height: 1,
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

	accSubspace := sdk.NewSubspace(authentication.DefaultParamspace)
	servicersSubspace := sdk.NewSubspace(servicersTypes.DefaultParamspace)
	providerSubspace := sdk.NewSubspace(types.DefaultParamspace)
	viperSubspace := sdk.NewSubspace(types.DefaultParamspace)
	ak := authentication.NewKeeper(cdc, keyAcc, accSubspace, maccPerms)
	nk := servicersKeeper.NewKeeper(cdc, servicersKey, ak, servicersSubspace, "pos")
	providerk := providersKeeper.NewKeeper(cdc, providersKey, nk, ak, nil, providerSubspace, providersTypes.ModuleName)
	keeper := keep.NewKeeper(viperKey, cdc, ak, nk, providerk, &hb, viperSubspace)
	kb := NewTestKeybase()
	providerk.ViperKeeper = keeper
	_, err = kb.Create("test")
	assert.Nil(t, err)
	_, err = kb.GetCoinbase()
	assert.Nil(t, err)
	moduleManager := module.NewManager(
		authentication.NewAppModule(ak),
		servicers.NewAppModule(nk),
		providers.NewAppModule(providerk),
	)
	genesisState := ModuleBasics.DefaultGenesis()
	moduleManager.InitGenesis(ctx, genesisState)
	initialCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultStakeDenom, valTokens))
	_ = createTestAccs(ctx, int(nAccs), initialCoins, &ak)
	_ = createTestProviders(ctx, int(nAccs), sdk.NewInt(10000000), providerk, ak)
	_ = createTestValidators(ctx, int(nAccs), sdk.NewInt(10000000), sdk.ZeroInt(), &nk, ak, kb)
	providerk.SetParams(ctx, providersTypes.DefaultParams())
	nk.SetParams(ctx, servicersTypes.DefaultParams())
	keeper.SetParams(ctx, types.DefaultParams())
	return ctx, nk, providerk, keeper, kb
}

// : unparam deadcode unused
func createTestAccs(ctx sdk.Ctx, numAccs int, initialCoins sdk.Coins, ak *authentication.Keeper) (accs []authentication.BaseAccount) {
	for i := 0; i < numAccs; i++ {
		privKey := crypto.GenerateEd25519PrivKey()
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

func createTestValidators(ctx sdk.Ctx, numAccs int, valCoins sdk.BigInt, daoCoins sdk.BigInt, nk *servicersKeeper.Keeper, ak authentication.Keeper, kb keys.Keybase) (accs servicersTypes.Validators) {
	ethereum := hex.EncodeToString([]byte{01})
	for i := 0; i < numAccs-1; i++ {
		privKey := crypto.GenerateEd25519PrivKey()
		pubKey := privKey.PublicKey()
		addr := sdk.Address(pubKey.Address())
		privKey2 := crypto.GenerateEd25519PrivKey()
		pubKey2 := privKey2.PublicKey()
		addr2 := sdk.Address(pubKey2.Address())
		val := servicersTypes.NewValidator(addr, pubKey, []string{ethereum}, "https://www.google.com:443", valCoins, addr2)
		// set the vals from the data
		nk.SetValidator(ctx, val)
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
		privKey := crypto.GenerateEd25519PrivKey()
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
