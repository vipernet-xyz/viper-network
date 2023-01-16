package keeper

import (
	"math/rand"
	"testing"

	types2 "github.com/vipernet-xyz/viper-network/codec/types"

	"github.com/vipernet-xyz/viper-network/crypto"
	"github.com/vipernet-xyz/viper-network/types/module"
	"github.com/vipernet-xyz/viper-network/x/apps/exported"
	govTypes "github.com/vipernet-xyz/viper-network/x/governance/types"
	"github.com/vipernet-xyz/viper-network/x/providers"
	nodeskeeper "github.com/vipernet-xyz/viper-network/x/providers/keeper"
	nodestypes "github.com/vipernet-xyz/viper-network/x/providers/types"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/vipernet-xyz/viper-network/codec"
	"github.com/vipernet-xyz/viper-network/store"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/apps/types"
	"github.com/vipernet-xyz/viper-network/x/authentication"
	"github.com/vipernet-xyz/viper-network/x/governance"
)

// : deadcode unused
var (
	ModuleBasics = module.NewBasicManager(
		authentication.AppModuleBasic{},
		providers.AppModuleBasic{},
	)
)

// : deadcode unused
// create a codec used only for testing
func makeTestCodec() *codec.Codec {
	var cdc = codec.NewCodec(types2.NewInterfaceRegistry())
	authentication.RegisterCodec(cdc)
	governance.RegisterCodec(cdc)
	nodestypes.RegisterCodec(cdc)
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
func createTestInput(t *testing.T, isCheckTx bool) (sdk.Context, []authentication.Account, Keeper) {
	initPower := int64(100000000000)
	nAccs := int64(4)

	keyAcc := sdk.NewKVStoreKey(authentication.StoreKey)
	keyParams := sdk.ParamsKey
	tkeyParams := sdk.ParamsTKey
	nodesKey := sdk.NewKVStoreKey(nodestypes.StoreKey)
	appsKey := sdk.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db, false, 5000000)
	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(nodesKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(appsKey, sdk.StoreTypeIAVL, db)
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
		nodestypes.StakedPoolName:       {authentication.Burner, authentication.Staking},
		govTypes.DAOAccountName:         {authentication.Burner, authentication.Staking},
	}

	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authentication.NewModuleAddress(acc).String()] = true
	}
	valTokens := sdk.TokensFromConsensusPower(initPower)
	accSubspace := sdk.NewSubspace(authentication.DefaultParamspace)
	nodesSubspace := sdk.NewSubspace(nodestypes.DefaultParamspace)
	appSubspace := sdk.NewSubspace(DefaultParamspace)
	ak := authentication.NewKeeper(cdc, keyAcc, accSubspace, maccPerms)
	nk := nodeskeeper.NewKeeper(cdc, nodesKey, ak, nodesSubspace, "pos")
	moduleManager := module.NewManager(
		authentication.NewAppModule(ak),
		providers.NewAppModule(nk),
	)
	genesisState := ModuleBasics.DefaultGenesis()
	moduleManager.InitGenesis(ctx, genesisState)
	initialCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultStakeDenom, valTokens))
	accs := createTestAccs(ctx, int(nAccs), initialCoins, &ak)
	keeper := NewKeeper(cdc, appsKey, nk, ak, MockViperKeeper{}, appSubspace, "apps")
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

func getRandomApplicationAddress() sdk.Address {
	return sdk.Address(getRandomPubKey().Address())
}

func getApplication() types.Application {
	pub := getRandomPubKey()
	return types.Application{
		Address:      sdk.Address(pub.Address()),
		StakedTokens: sdk.NewInt(100000000000),
		PublicKey:    pub,
		Jailed:       false,
		Status:       sdk.Staked,
		MaxRelays:    sdk.NewInt(100000000000),
		Chains:       []string{"0001"},
	}
}

func getStakedApplication() types.Application {
	return getApplication()
}

func getUnstakedApplication() types.Application {
	v := getApplication()
	return v.UpdateStatus(sdk.Unstaked)
}

func getUnstakingApplication() types.Application {
	v := getApplication()
	return v.UpdateStatus(sdk.Unstaking)
}

func modifyFn(i *int) func(index int64, application exported.ApplicationI) (stop bool) {
	return func(index int64, application exported.ApplicationI) (stop bool) {
		app := application.(types.Application)
		app.StakedTokens = sdk.NewInt(100)
		if index == 1 {
			stop = true
		}
		*i++
		return
	}
}
