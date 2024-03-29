package servicers

import (
	"math/rand"
	"testing"

	types2 "github.com/vipernet-xyz/viper-network/codec/types"

	"github.com/tendermint/tendermint/rpc/client/http"

	"github.com/vipernet-xyz/viper-network/codec"
	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	"github.com/vipernet-xyz/viper-network/store"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/types/module"
	"github.com/vipernet-xyz/viper-network/x/authentication"
	"github.com/vipernet-xyz/viper-network/x/governance"
	govKeeper "github.com/vipernet-xyz/viper-network/x/governance/keeper"
	govTypes "github.com/vipernet-xyz/viper-network/x/governance/types"
	"github.com/vipernet-xyz/viper-network/x/servicers/keeper"
	"github.com/vipernet-xyz/viper-network/x/servicers/types"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/rpc/client"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"
)

// : deadcode unused
var (
	ModuleBasics = module.NewBasicManager(
		authentication.AppModuleBasic{},
		governance.AppModuleBasic{},
	)
)

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

func GetTestTendermintClient() client.Client {
	var tmNodeURI string
	var defaultTMURI = "tcp://localhost:26657"

	if tmNodeURI == "" {
		c, _ := http.New(defaultTMURI, "/websocket")
		return c
	}
	c, _ := http.New(tmNodeURI, "/websocket")
	return c
}

// : deadcode unused
func createTestInput(t *testing.T, isCheckTx bool) (sdk.Ctx, []authentication.Account, keeper.Keeper) {
	initPower := int64(100000000000)
	nAccs := int64(4)

	keyAcc := sdk.NewKVStoreKey(authentication.StoreKey)
	keyPOS := sdk.NewKVStoreKey(types.ModuleName)
	keyParams := sdk.ParamsKey
	tkeyParams := sdk.ParamsTKey
	govKey := sdk.NewKVStoreKey(govTypes.StoreKey)
	dKey := sdk.NewKVStoreKey("DiscountKey")

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db, false, 5000000)
	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	ms.MountStoreWithDB(keyPOS, sdk.StoreTypeIAVL, db)
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
		govTypes.DAOAccountName:         {authentication.Burner, authentication.Staking, authentication.Minter},
	}
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authentication.NewModuleAddress(acc).String()] = true
	}
	valTokens := sdk.TokensFromConsensusPower(initPower)

	accSubspace := sdk.NewSubspace(authentication.DefaultParamspace)
	posSubspace := sdk.NewSubspace(types.DefaultParamspace)

	ak := authentication.NewKeeper(cdc, keyAcc, accSubspace, maccPerms)
	moduleManager := module.NewManager(
		authentication.NewAppModule(ak),
	)

	genesisState := ModuleBasics.DefaultGenesis()
	moduleManager.InitGenesis(ctx, genesisState)

	initialCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultStakeDenom, valTokens))
	accs := createTestAccs(ctx, int(nAccs), initialCoins, &ak)
	govKeeper := govKeeper.NewKeeper(cdc, govKey, tkeyParams, dKey, govTypes.ModuleName, ak)
	keeper := keeper.NewKeeper(cdc, keyPOS, ak, nil, govKeeper, posSubspace, sdk.CodespaceType("pos"))

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

//func addMintedCoinsToModule(t *testing.T, ctx sdk.Ctx, k *keeper.Keeper, module string) {
//	coins := sdk.NewCoins(sdk.NewCoin(k.StakeDenom(ctx), sdk.NewInt(100000000000)))
//	mintErr := k.supplyKeeper.MintCoins(ctx, module, coins.Add(coins))
//	if mintErr != nil {
//		t.Fail()
//	}
//}
//
//func sendFromModuleToAccount(t *testing.T, ctx sdk.Ctx, k *keeper.Keeper, module string, address sdk.Address, amount sdk.BigInt) {
//	coins := sdk.NewCoins(sdk.NewCoin(k.StakeDenom(ctx), amount))
//	err := k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, module, sdk.Address(address), coins)
//	if err != nil {
//		t.Fail()
//	}
//}

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
	pub2 := getRandomPubKey()
	return types.Validator{
		Address:       sdk.Address(pub.Address()),
		StakedTokens:  sdk.NewInt(100000000000),
		PublicKey:     pub,
		Jailed:        false,
		Status:        sdk.Staked,
		ServiceURL:    "https://www.google.com:443",
		Chains:        []string{"0001"},
		OutputAddress: sdk.Address(pub2.Address()),
	}
}

func getStakedValidator() types.Validator {
	return getValidator()
}

func getGenesisStateForTest(ctx sdk.Ctx, keeper keeper.Keeper, defaultparams bool) types.GenesisState {
	keeper.SetPreviousProposer(ctx, sdk.GetAddress(getRandomPubKey()))
	var prm = types.DefaultParams()

	if !defaultparams {
		prm = keeper.GetParams(ctx)
	}
	prevStateTotalPower := keeper.PrevStateValidatorsPower(ctx)
	validators := keeper.GetAllValidators(ctx)
	var prevStateValidatorPowers []types.PrevStatePowerMrequestoring
	keeper.IterateAndExecuteOverPrevStateValsByPower(ctx, func(addr sdk.Address, power int64) (stop bool) {
		prevStateValidatorPowers = append(prevStateValidatorPowers, types.PrevStatePowerMrequestoring{Address: addr, Power: power})
		return false
	})
	signingInfos := make(map[string]types.ValidatorSigningInfo)
	missedBlocks := make(map[string][]types.MissedBlock)
	keeper.IterateAndExecuteOverValSigningInfo(ctx, func(address sdk.Address, info types.ValidatorSigningInfo) (stop bool) {
		addrstring := address.String()
		signingInfos[addrstring] = info
		localMissedBlocks := []types.MissedBlock{}

		keeper.IterateAndExecuteOverMissedArray(ctx, address, func(index int64, missed bool) (stop bool) {
			localMissedBlocks = append(localMissedBlocks, types.MissedBlock{Index: index, Missed: missed})
			return false
		})
		missedBlocks[addrstring] = localMissedBlocks

		return false
	})
	prevProposer := keeper.GetPreviousProposer(ctx)

	return types.GenesisState{
		Params:                   prm,
		PrevStateTotalPower:      prevStateTotalPower,
		PrevStateValidatorPowers: prevStateValidatorPowers,
		Validators:               validators,
		Exported:                 true,
		SigningInfos:             signingInfos,
		MissedBlocks:             missedBlocks,
		PreviousProposer:         prevProposer,
	}

}
