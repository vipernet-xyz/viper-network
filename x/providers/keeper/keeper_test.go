package keeper

import (
	"fmt"
	"testing"

	"github.com/vipernet-xyz/viper-network/store"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/types/module"
	"github.com/vipernet-xyz/viper-network/x/authentication"
	governanceTypes "github.com/vipernet-xyz/viper-network/x/governance/types"
	"github.com/vipernet-xyz/viper-network/x/providers/types"
	"github.com/vipernet-xyz/viper-network/x/servicers"
	servicerskeeper "github.com/vipernet-xyz/viper-network/x/servicers/keeper"
	servicerstypes "github.com/vipernet-xyz/viper-network/x/servicers/types"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"
)

func TestKeeper_Codespace(t *testing.T) {
	_, _, keeper := createTestInput(t, true)
	if got := keeper.Codespace(); got != "apps" {
		t.Errorf("Codespace() = %v, want %v", got, "apps")
	}
}

func TestKeepers_NewKeeper(t *testing.T) {
	tests := []struct {
		name     string
		hasError bool
		msg      string
	}{
		{
			name:     "create a keeper",
			hasError: false,
		},
		{
			name:     "errors if no GetModuleAddress is nill",
			msg:      fmt.Sprintf("%s module account has not been set", types.StakedPoolName),
			hasError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initPower := int64(100000000000)
			nAccs := int64(4)

			keyAcc := sdk.NewKVStoreKey(authentication.StoreKey)
			keyParams := sdk.ParamsKey
			tkeyParams := sdk.ParamsTKey
			servicersKey := sdk.NewKVStoreKey(servicerstypes.StoreKey)
			providersKey := sdk.NewKVStoreKey(types.StoreKey)

			db := dbm.NewMemDB()
			ms := store.NewCommitMultiStore(db, false, 5000000)
			ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
			ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
			ms.MountStoreWithDB(servicersKey, sdk.StoreTypeIAVL, db)
			ms.MountStoreWithDB(providersKey, sdk.StoreTypeIAVL, db)
			ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
			err := ms.LoadLatestVersion()
			if err != nil {
				t.FailNow()
			}

			ctx := sdk.NewContext(ms, abci.Header{ChainID: "test-chain"}, true, log.NewNopLogger()).WithAppVersion("0.0.0")
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
				servicerstypes.StakedPoolName:   {authentication.Burner, authentication.Staking},
				governanceTypes.DAOAccountName:  {authentication.Burner, authentication.Staking},
			}
			if !tt.hasError {
				maccPerms[types.StakedPoolName] = []string{authentication.Burner, authentication.Staking, authentication.Minter}
			}

			modAccAddrs := make(map[string]bool)
			for acc := range maccPerms {
				modAccAddrs[authentication.NewModuleAddress(acc).String()] = true
			}
			valTokens := sdk.TokensFromConsensusPower(initPower)

			accSubspace := sdk.NewSubspace(authentication.DefaultParamspace)
			servicersSubspace := sdk.NewSubspace(servicerstypes.DefaultParamspace)
			providerSubspace := sdk.NewSubspace(DefaultParamspace)
			ak := authentication.NewKeeper(cdc, keyAcc, accSubspace, maccPerms)
			nk := servicerskeeper.NewKeeper(cdc, servicersKey, ak, servicersSubspace, "pos")
			moduleManager := module.NewManager(
				authentication.NewAppModule(ak),
				servicers.NewAppModule(nk),
			)
			genesisState := ModuleBasics.DefaultGenesis()
			moduleManager.InitGenesis(ctx, genesisState)
			initialCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultStakeDenom, valTokens))

			_ = createTestAccs(ctx, int(nAccs), initialCoins, &ak)

			if tt.hasError {
				return
			}
			_ = NewKeeper(cdc, providersKey, nk, ak, MockViperKeeper{}, providerSubspace, "providers")
		})
	}
}
