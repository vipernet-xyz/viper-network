package keeper

import (
	"math/rand"
	"testing"

	"github.com/vipernet-xyz/viper-network/codec/types"

	"github.com/vipernet-xyz/viper-network/codec"
	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	"github.com/vipernet-xyz/viper-network/store"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/types/module"
	"github.com/vipernet-xyz/viper-network/x/authentication"
	"github.com/vipernet-xyz/viper-network/x/authentication/keeper"
	governanceTypes "github.com/vipernet-xyz/viper-network/x/governance/types"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"
)

var (
	ModuleBasics = module.NewBasicManager(
		authentication.AppModuleBasic{},
	)
)

// nolint: deadcode unused
// create a codec used only for testing
func makeTestCodec() *codec.Codec {
	var cdc = codec.NewCodec(types.NewInterfaceRegistry())
	authentication.RegisterCodec(cdc)
	governanceTypes.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	crypto.RegisterAmino(cdc.AminoCodec().Amino)
	return cdc
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

// nolint: deadcode unused
func createTestKeeperAndContext(t *testing.T, isCheckTx bool) (sdk.Context, Keeper) {
	keyAcc := sdk.NewKVStoreKey(authentication.StoreKey)
	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db, false, 5000000)
	ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(sdk.ParamsKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(sdk.ParamsTKey, sdk.StoreTypeTransient, db)
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
	cdc := makeTestCodec()
	maccPerms := map[string][]string{
		authentication.FeeCollectorName: nil,
		governanceTypes.DAOAccountName:  {"burner", "staking", "minter"},
		"FAKE":                          {"burner", "staking", "minter"},
	}
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authentication.NewModuleAddress(acc).String()] = true
	}
	akSubspace := sdk.NewSubspace(authentication.DefaultParamspace)
	ak := keeper.NewKeeper(cdc, keyAcc, akSubspace, maccPerms)
	ak.GetModuleAccount(ctx, "FAKE")
	pk := NewKeeper(cdc, sdk.ParamsKey, sdk.ParamsTKey, governanceTypes.DefaultParamspace, ak, akSubspace)
	moduleManager := module.NewManager(
		authentication.NewAppModule(ak),
	)
	genesisState := ModuleBasics.DefaultGenesis()
	moduleManager.InitGenesis(ctx, genesisState)
	params := governanceTypes.DefaultParams()
	pk.SetParams(ctx, params)
	gs := governanceTypes.DefaultGenesisState()
	acl := createTestACL()
	gs.Params.ACL = acl
	pk.InitGenesis(ctx, gs)
	return ctx, pk
}

var testACL governanceTypes.ACL

func createTestACL() governanceTypes.ACL {
	if testACL == nil {
		acl := governanceTypes.ACL(make([]governanceTypes.ACLPair, 0))
		acl.SetOwner("authentication/MaxMemoCharacters", getRandomValidatorAddress())
		acl.SetOwner("authentication/TxSigLimit", getRandomValidatorAddress())
		acl.SetOwner("authentication/FeeMultipliers", getRandomValidatorAddress())
		acl.SetOwner("governance/daoOwner", getRandomValidatorAddress())
		acl.SetOwner("governance/acl", getRandomValidatorAddress())
		acl.SetOwner("governance/upgrade", getRandomValidatorAddress())
		testACL = acl
	}
	return testACL
}

// Checks wether or not a Events slice contains an event that equals the values of event
func ContainsEvent(events sdk.Events, event abci.Event) bool {
	stringEvents := sdk.StringifyEvents(events.ToABCIEvents())
	stringEventStr := sdk.StringEvents{sdk.StringifyEvent(event)}.String()
	for _, item := range stringEvents {
		itemStr := sdk.StringEvents{item}.String()
		if itemStr == stringEventStr {
			return true
		}
	}
	return false
}
