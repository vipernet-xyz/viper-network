package keeper

import (
	"math/rand"
	"testing"

	types2 "github.com/vipernet-xyz/viper-network/codec/types"
	"github.com/vipernet-xyz/viper-network/x/governance"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"
	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	"github.com/vipernet-xyz/viper-network/types/module"
	govKeeper "github.com/vipernet-xyz/viper-network/x/governance/keeper"
	govTypes "github.com/vipernet-xyz/viper-network/x/governance/types"
	"github.com/vipernet-xyz/viper-network/x/requestors/exported"
	"github.com/vipernet-xyz/viper-network/x/servicers"
	servicerskeeper "github.com/vipernet-xyz/viper-network/x/servicers/keeper"
	servicerstypes "github.com/vipernet-xyz/viper-network/x/servicers/types"

	"github.com/vipernet-xyz/viper-network/codec"
	"github.com/vipernet-xyz/viper-network/store"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/authentication"
	"github.com/vipernet-xyz/viper-network/x/requestors/types"
)

// : deadcode unused
var (
	ModuleBasics = module.NewBasicManager(
		authentication.AppModuleBasic{},
		servicers.AppModuleBasic{},
	)
)

// : deadcode unused
// create a codec used only for testing
func makeTestCodec() *codec.Codec {
	var cdc = codec.NewCodec(types2.NewInterfaceRegistry())
	authentication.RegisterCodec(cdc)
	governance.RegisterCodec(cdc)
	servicerstypes.RegisterCodec(cdc)
	types.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	crypto.RegisterAmino(cdc.AminoCodec().Amino)

	return cdc
}

type MockViperKeeper struct {
}

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
	servicersKey := sdk.NewKVStoreKey(servicerstypes.StoreKey)
	requestorsKey := sdk.NewKVStoreKey(types.StoreKey)
	govKey := sdk.NewKVStoreKey(govTypes.StoreKey)
	dKey := sdk.NewKVStoreKey("DiscountKey")

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db, false, 5000000)
	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(servicersKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(requestorsKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
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
		types.StakedPoolName:            {authentication.Burner, authentication.Staking, authentication.Minter},
		servicerstypes.StakedPoolName:   {authentication.Burner, authentication.Staking},
		govTypes.DAOAccountName:         {authentication.Burner, authentication.Staking},
	}

	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authentication.NewModuleAddress(acc).String()] = true
	}
	valTokens := sdk.TokensFromConsensusPower(initPower)
	accSubspace := sdk.NewSubspace(authentication.DefaultParamspace)
	servicersSubspace := sdk.NewSubspace(servicerstypes.DefaultParamspace)
	appSubspace := sdk.NewSubspace(DefaultParamspace)
	ak := authentication.NewKeeper(cdc, keyAcc, accSubspace, maccPerms)
	govKeeper := govKeeper.NewKeeper(cdc, govKey, tkeyParams, dKey, govTypes.ModuleName, ak)
	nk := servicerskeeper.NewKeeper(cdc, servicersKey, ak, nil, govKeeper, servicersSubspace, "pos")
	moduleManager := module.NewManager(
		authentication.NewAppModule(ak),
		servicers.NewAppModule(nk),
	)
	genesisState := ModuleBasics.DefaultGenesis()
	moduleManager.InitGenesis(ctx, genesisState)
	initialCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultStakeDenom, valTokens))
	accs := createTestAccs(ctx, int(nAccs), initialCoins, &ak)
	keeper := NewKeeper(cdc, requestorsKey, nk, ak, MockViperKeeper{}, appSubspace, "apps")
	p := types.DefaultParams()
	keeper.SetParams(ctx, p)
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

func getRandomRequestorAddress() sdk.Address {
	return sdk.Address(getRandomPubKey().Address())
}

func getRequestor() types.Requestor {
	pub := getRandomPubKey()
	return types.Requestor{
		Address:      sdk.Address(pub.Address()),
		StakedTokens: sdk.NewInt(100000000000),
		PublicKey:    pub,
		Jailed:       false,
		Status:       sdk.Staked,
		MaxRelays:    sdk.NewInt(100000000000),
		Chains:       []string{"0001"},
		GeoZones:     []string{"0001"},
		NumServicers: 10,
	}
}

func getStakedRequestor() types.Requestor {
	return getRequestor()
}

func getUnstakedRequestor() types.Requestor {
	v := getRequestor()
	return v.UpdateStatus(sdk.Unstaked)
}

func getUnstakingRequestor() types.Requestor {
	v := getRequestor()
	return v.UpdateStatus(sdk.Unstaking)
}

func modifyFn(i *int) func(index int64, requestor exported.RequestorI) (stop bool) {
	return func(index int64, requestor exported.RequestorI) (stop bool) {
		app := requestor.(types.Requestor)
		app.StakedTokens = sdk.NewInt(100)
		if index == 1 {
			stop = true
		}
		*i++
		return
	}
}
