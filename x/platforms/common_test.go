package pos

import (
	"math/rand"
	"testing"

	types3 "github.com/vipernet-xyz/viper-network/codec/types"

	"github.com/vipernet-xyz/viper-network/codec"
	"github.com/vipernet-xyz/viper-network/crypto"
	"github.com/vipernet-xyz/viper-network/store"
	"github.com/vipernet-xyz/viper-network/types/module"
	"github.com/vipernet-xyz/viper-network/x/authentication"
	"github.com/vipernet-xyz/viper-network/x/governance"
	governanceTypes "github.com/vipernet-xyz/viper-network/x/governance/types"
	"github.com/vipernet-xyz/viper-network/x/platforms/keeper"
	"github.com/vipernet-xyz/viper-network/x/platforms/types"
	"github.com/vipernet-xyz/viper-network/x/providers"
	providerskeeper "github.com/vipernet-xyz/viper-network/x/providers/keeper"
	providerstypes "github.com/vipernet-xyz/viper-network/x/providers/types"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	sdk "github.com/vipernet-xyz/viper-network/types"
)

var (
	ModuleBasics = module.NewBasicManager(
		authentication.PlatformModuleBasic{},
		governance.PlatformModuleBasic{},
		providers.PlatformModuleBasic{},
	)
)

// create a codec used only for testing
func makeTestCodec() *codec.Codec {
	var cdc = codec.NewCodec(types3.NewInterfaceRegistry())
	authentication.RegisterCodec(cdc)
	governance.RegisterCodec(cdc)
	providerstypes.RegisterCodec(cdc)
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

func createTestInput(t *testing.T, isCheckTx bool) (sdk.Ctx, keeper.Keeper, types.AuthKeeper, types.PosKeeper) {
	initPower := int64(100000000000)
	nAccs := int64(4)
	keyAcc := sdk.NewKVStoreKey(authentication.StoreKey)
	keyParams := sdk.ParamsKey
	tkeyParams := sdk.ParamsTKey
	providersKey := sdk.NewKVStoreKey(providerstypes.StoreKey)
	platformsKey := sdk.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db, false, 5000000)
	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(providersKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(platformsKey, sdk.StoreTypeIAVL, db)
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
		providerstypes.StakedPoolName:   {authentication.Burner, authentication.Staking},
		governanceTypes.DAOAccountName:  {authentication.Burner, authentication.Staking},
	}
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authentication.NewModuleAddress(acc).String()] = true
	}
	valTokens := sdk.TokensFromConsensusPower(initPower)
	accSubspace := sdk.NewSubspace(authentication.DefaultParamspace)
	providersSubspace := sdk.NewSubspace(providerstypes.DefaultParamspace)
	platformSubspace := sdk.NewSubspace(types.DefaultParamspace)
	ak := authentication.NewKeeper(cdc, keyAcc, accSubspace, maccPerms)
	nk := providerskeeper.NewKeeper(cdc, providersKey, ak, providersSubspace, "pos")
	moduleManager := module.NewManager(
		authentication.NewPlatformModule(ak),
		providers.NewPlatformModule(nk),
	)
	genesisState := ModuleBasics.DefaultGenesis()
	moduleManager.InitGenesis(ctx, genesisState)

	initialCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultStakeDenom, valTokens))
	_ = createTestAccs(ctx, int(nAccs), initialCoins, &ak)
	keeper := keeper.NewKeeper(cdc, platformsKey, nk, ak, MockViperKeeper{}, platformSubspace, "platforms")
	p := types.DefaultParams()
	keeper.SetParams(ctx, p)
	return ctx, keeper, ak, nk
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

func getRandomPubKey() crypto.Ed25519PublicKey {
	var pub crypto.Ed25519PublicKey
	_, err := rand.Read(pub[:])
	if err != nil {
		_ = err
	}
	return pub
}

func getPlatform() types.Platform {
	pub := getRandomPubKey()
	return types.Platform{
		Address:      sdk.Address(pub.Address()),
		StakedTokens: sdk.NewInt(100000000000),
		PublicKey:    pub,
		Jailed:       false,
		Status:       sdk.Staked,
		MaxRelays:    sdk.NewInt(100000000000),
		Chains:       []string{"0001"},
	}
}
