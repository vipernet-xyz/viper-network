package keeper

import (
	"fmt"
	"testing"

	"github.com/vipernet-xyz/viper-network/store"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/types/module"
	"github.com/vipernet-xyz/viper-network/x/authentication"
	govTypes "github.com/vipernet-xyz/viper-network/x/governance/types"
	"github.com/vipernet-xyz/viper-network/x/platforms/types"
	"github.com/vipernet-xyz/viper-network/x/providers"
	nodeskeeper "github.com/vipernet-xyz/viper-network/x/providers/keeper"
	nodestypes "github.com/vipernet-xyz/viper-network/x/providers/types"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"
)

func TestKeeper_Codespace(t *testing.T) {
	_, _, keeper := createTestInput(t, true)
	if got := keeper.Codespace(); got != "platforms" {
		t.Errorf("Codespace() = %v, want %v", got, "platforms")
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
			nodesKey := sdk.NewKVStoreKey(nodestypes.StoreKey)
			platformsKey := sdk.NewKVStoreKey(types.StoreKey)

			db := dbm.NewMemDB()
			ms := store.NewCommitMultiStore(db, false, 5000000)
			ms.MountStoreWithDB(keyAcc, sdk.StoreTypeIAVL, db)
			ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
			ms.MountStoreWithDB(nodesKey, sdk.StoreTypeIAVL, db)
			ms.MountStoreWithDB(platformsKey, sdk.StoreTypeIAVL, db)
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
				nodestypes.StakedPoolName:       {authentication.Burner, authentication.Staking},
				govTypes.DAOAccountName:         {authentication.Burner, authentication.Staking},
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
			nodesSubspace := sdk.NewSubspace(nodestypes.DefaultParamspace)
			platformSubspace := sdk.NewSubspace(DefaultParamspace)
			ak := authentication.NewKeeper(cdc, keyAcc, accSubspace, maccPerms)
			nk := nodeskeeper.NewKeeper(cdc, nodesKey, ak, nodesSubspace, "pos")
			moduleManager := module.NewManager(
				authentication.NewPlatformModule(ak),
				providers.NewPlatformModule(nk),
			)
			genesisState := ModuleBasics.DefaultGenesis()
			moduleManager.InitGenesis(ctx, genesisState)
			initialCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultStakeDenom, valTokens))

			_ = createTestAccs(ctx, int(nAccs), initialCoins, &ak)

			if tt.hasError {
				return
			}
			_ = NewKeeper(cdc, platformsKey, nk, ak, MockViperKeeper{}, platformSubspace, "platforms")
		})
	}
}
