// nolint
package app

/*
import (
	"encoding/hex"
	"encoding/json"
	"math/big"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/vipernet-xyz/viper-network/codec"

	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	"github.com/vipernet-xyz/viper-network/crypto/keys"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/governance"
	requestors "github.com/vipernet-xyz/viper-network/x/requestors"
	types3 "github.com/vipernet-xyz/viper-network/x/requestors/types"
	"github.com/vipernet-xyz/viper-network/x/servicers"
	types2 "github.com/vipernet-xyz/viper-network/x/servicers/types"
	"github.com/vipernet-xyz/viper-network/x/viper-main/types"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/node"
	tmTypes "github.com/tendermint/tendermint/types"
	"gopkg.in/h2non/gock.v1"
)

func TestQueryBlock(t *testing.T) {

	tt := []struct {
		name         string
		memoryNodeFn func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		*upgrades
	}{
		{name: "query block amino", memoryNodeFn: NewInMemoryTendermintNodeAmino, upgrades: &upgrades{codecUpgrade: codecUpgrade{false, 7000}}},
		{name: "query block proto", memoryNodeFn: NewInMemoryTendermintNodeProto, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 2}}},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			_, _, cleanup := tc.memoryNodeFn(t, oneAppTwoNodeGenesis())
			_, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
			height := int64(1)
			<-evtChan // Wait for block
			got, err := VCA.QueryBlock(&height)
			assert.Nil(t, err)
			assert.NotNil(t, got)

			cleanup()
			stopCli()
		})
	}
}

func TestQueryChainHeight(t *testing.T) {

	tt := []struct {
		name         string
		memoryNodeFn func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		*upgrades
	}{
		{name: "query height amino", memoryNodeFn: NewInMemoryTendermintNodeAmino, upgrades: &upgrades{codecUpgrade: codecUpgrade{false, 7000}}},
		{name: "query height proto", memoryNodeFn: NewInMemoryTendermintNodeProto, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 2}}},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			_, _, cleanup := tc.memoryNodeFn(t, oneAppTwoNodeGenesis())
			_, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
			<-evtChan // Wait for block
			got, err := VCA.QueryHeight()
			assert.Nil(t, err)
			assert.Equal(t, int64(1), got) // should not be 0 due to empty blocks

			cleanup()
			stopCli()
		})
	}
}

func TestQueryTx(t *testing.T) {
	tt := []struct {
		name         string
		memoryNodeFn func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		*upgrades
	}{
		{name: "query tx from proto account with proto cdec", memoryNodeFn: NewInMemoryTendermintNodeProto, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 2}}},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if tc.upgrades != nil { // NOTE: Use to perform neccesary upgrades for test
				codec.UpgradeHeight = tc.upgrades.codecUpgrade.height
				_ = memCodecMod(tc.upgrades.codecUpgrade.upgradeMod)
			}
			_, kb, cleanup := tc.memoryNodeFn(t, oneAppTwoNodeGenesis())
			time.Sleep(time.Second * 2)
			cb, err := kb.GetCoinbase()
			assert.Nil(t, err)
			kp, err := kb.Create("test")
			assert.Nil(t, err)
			_, _, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
			var tx *sdk.TxResponse
			<-evtChan // Wait for block
			memCli, stopCli, evtChan := subscribeTo(t, tmTypes.EventTx)
			tx, err = servicers.Send(memCodec(), memCli, kb, cb.GetAddress(), kp.GetAddress(), "test", sdk.NewInt(1000), tc.upgrades.codecUpgrade.upgradeMod)
			assert.Nil(t, err)
			assert.NotNil(t, tx)

			<-evtChan // Wait for tx
			got, err := VCA.QueryTx(tx.TxHash, false)
			assert.Nil(t, err)
			balance, err := VCA.QueryBalance(kp.GetAddress().String(), VCA.BaseApp.LastBlockHeight())
			assert.Nil(t, err)
			assert.Equal(t, int64(1000), balance.Int64())
			assert.NotNil(t, got)

			cleanup()
			stopCli()
		})
	}
}

func TestQueryAminoTx(t *testing.T) {
	tt := []struct {
		name         string
		memoryNodeFn func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		*upgrades
	}{
		{name: "query tx amino", memoryNodeFn: NewInMemoryTendermintNodeAmino, upgrades: &upgrades{codecUpgrade: codecUpgrade{false, 7000}}},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if tc.upgrades != nil { // NOTE: Use to perform neccesary upgrades for test
				codec.UpgradeHeight = tc.upgrades.codecUpgrade.height
				_ = memCodecMod(tc.upgrades.codecUpgrade.upgradeMod)
			}
			_, kb, cleanup := tc.memoryNodeFn(t, oneAppTwoNodeGenesis())
			time.Sleep(time.Second * 2)
			cb, err := kb.GetCoinbase()
			assert.Nil(t, err)
			kp, err := kb.Create("test")
			assert.Nil(t, err)
			_, _, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
			var tx *sdk.TxResponse
			<-evtChan // Wait for block
			memCli, stopCli, evtChan := subscribeTo(t, tmTypes.EventTx)
			tx, err = servicers.Send(memCodec(), memCli, kb, cb.GetAddress(), kp.GetAddress(), "test", sdk.NewInt(1000), tc.upgrades.codecUpgrade.upgradeMod)
			assert.Nil(t, err)
			assert.NotNil(t, tx)

			<-evtChan // Wait for tx
			got, err := VCA.QueryTx(tx.TxHash, false)
			assert.Nil(t, err)
			validator, err := VCA.QueryBalance(kp.GetAddress().String(), VCA.LastBlockHeight())
			assert.Nil(t, err)
			assert.True(t, validator.Equal(sdk.NewInt(1000)))
			assert.NotNil(t, got)

			cleanup()
			stopCli()
		})
	}
}

func TestQueryValidators(t *testing.T) {
	tt := []struct {
		name         string
		memoryNodeFn func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		*upgrades
	}{
		{name: "query validators proto", memoryNodeFn: NewInMemoryTendermintNodeProto, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 1}}},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if tc.upgrades != nil { // NOTE: Use to perform neccesary upgrades for test
				codec.UpgradeHeight = tc.upgrades.codecUpgrade.height
				_ = memCodecMod(tc.upgrades.codecUpgrade.upgradeMod)
			}
			gen, _ := twoValTwoNodeGenesisState()
			_, _, cleanup := tc.memoryNodeFn(t, gen)
			time.Sleep(2 * time.Second)
			_, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
			<-evtChan // Wait for block
			got, err := VCA.QueryServicers(VCA.LastBlockHeight(), types2.QueryValidatorsParams{Page: 1, Limit: 1})
			assert.Nil(t, err)
			res := got.Result.([]types2.Validator)
			assert.Equal(t, 1, len(res))
			got, err = VCA.QueryServicers(0, types2.QueryValidatorsParams{Page: 2, Limit: 1})
			assert.Nil(t, err)
			res = got.Result.([]types2.Validator)
			assert.Equal(t, 1, len(res))
			got, err = VCA.QueryServicers(0, types2.QueryValidatorsParams{Page: 1, Limit: 1000})
			assert.Nil(t, err)
			res = got.Result.([]types2.Validator)
			assert.Equal(t, 2, len(res))
			cleanup()
			stopCli()
		})
	}
}
func TestQueryRequestors(t *testing.T) {
	tt := []struct {
		name         string
		memoryNodeFn func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		*upgrades
	}{
		{name: "query requestors from amino account with amino codec", memoryNodeFn: NewInMemoryTendermintNodeAmino, upgrades: &upgrades{codecUpgrade: codecUpgrade{false, 7000}}},
		{name: "query requestors from proto account with proto cdec", memoryNodeFn: NewInMemoryTendermintNodeProto, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 2}}},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if tc.upgrades != nil { // NOTE: Use to perform necessary upgrades for test
				codec.UpgradeHeight = tc.upgrades.codecUpgrade.height
				_ = memCodecMod(tc.upgrades.codecUpgrade.upgradeMod)
			}
			_, kb, cleanup := tc.memoryNodeFn(t, oneAppTwoNodeGenesis())
			time.Sleep(time.Second * 2)
			kp, err := kb.GetCoinbase()
			assert.Nil(t, err)
			_, _, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
			var tx *sdk.TxResponse
			var chains = []string{"0001"}
			var geozones = []string{"0001"}

			<-evtChan // Wait for block
			memCli, stopCli, evtChan := subscribeTo(t, tmTypes.EventTx)
			tx, err = requestors.StakeTx(memCodec(), memCli, kb, chains, geozones, 5, sdk.NewInt(1000000), kp, "test", tc.codecUpgrade.upgradeMod)
			assert.Nil(t, err)
			assert.NotNil(t, tx)

			<-evtChan // Wait for tx
			got, err := VCA.QueryRequestors(VCA.LastBlockHeight(), types3.QueryRequestorsWithOpts{
				Page:  1,
				Limit: 1,
			})
			assert.Nil(t, err)
			slice, ok := takeArg(got.Result, reflect.Slice)
			if !ok {
				t.Fatalf("couldn't convert arg to slice")
			}
			assert.Equal(t, 1, slice.Len())
			got, err = VCA.QueryRequestors(VCA.LastBlockHeight(), types3.QueryRequestorsWithOpts{
				Page:  2,
				Limit: 1,
			})
			assert.Nil(t, err)
			slice, ok = takeArg(got.Result, reflect.Slice)
			if !ok {
				t.Fatalf("couldn't convert arg to slice")
			}
			assert.Equal(t, 1, slice.Len())
			got, err = VCA.QueryRequestors(VCA.LastBlockHeight(), types3.QueryRequestorsWithOpts{
				Page:  1,
				Limit: 2,
			})
			assert.Nil(t, err)
			slice, ok = takeArg(got.Result, reflect.Slice)
			if !ok {
				t.Fatalf("couldn't convert arg to slice")
			}
			assert.Equal(t, 2, slice.Len())

			stopCli()
			cleanup()
		})
	}
}

func TestQueryValidator(t *testing.T) {
	tt := []struct {
		name         string
		memoryNodeFn func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		*upgrades
	}{
		{name: "query validator amino", memoryNodeFn: NewInMemoryTendermintNodeAmino},
		{name: "query validator proto", memoryNodeFn: NewInMemoryTendermintNodeProto, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 2}}},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			_, kb, cleanup := tc.memoryNodeFn(t, oneAppTwoNodeGenesis())
			cb, err := kb.GetCoinbase()
			if err != nil {
				t.Fatal(err)
			}
			_, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
			<-evtChan // Wait for block
			got, err := VCA.QueryServicer(cb.GetAddress().String(), VCA.LastBlockHeight())
			assert.Nil(t, err)
			assert.Equal(t, cb.GetAddress(), got.Address)
			assert.False(t, got.Jailed)
			assert.True(t, got.StakedTokens.Equal(sdk.NewInt(1000000000000000)))

			cleanup()
			stopCli()
		})
	}
}

func TestQueryDaoBalance(t *testing.T) {
	tt := []struct {
		name         string
		memoryNodeFn func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		*upgrades
	}{
		{name: "query dao balance from amino account with amino codec", memoryNodeFn: NewInMemoryTendermintNodeAmino, upgrades: &upgrades{codecUpgrade: codecUpgrade{false, 7000}}},
		{name: "query dao balance from proto account with proto cdec", memoryNodeFn: NewInMemoryTendermintNodeProto, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 2}}},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if tc.upgrades != nil { // NOTE: Use to perform neccesary upgrades for test
				codec.UpgradeHeight = tc.upgrades.codecUpgrade.height
				_ = memCodecMod(tc.upgrades.codecUpgrade.upgradeMod)
			}
			_, _, cleanup := tc.memoryNodeFn(t, oneAppTwoNodeGenesis())
			memCli, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
			<-evtChan // Wait for block
			got, err := governance.QueryDAO(memCodec(), memCli, VCA.LastBlockHeight())
			assert.Nil(t, err)
			assert.Equal(t, big.NewInt(1000), got.BigInt())

			cleanup()
			stopCli()
		})
	}
}

func TestQueryACL(t *testing.T) {
	tt := []struct {
		name         string
		memoryNodeFn func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		*upgrades
	}{
		{name: "query dao balance from amino account with amino codec", memoryNodeFn: NewInMemoryTendermintNodeAmino, upgrades: &upgrades{codecUpgrade: codecUpgrade{false, 7000}}},
		{name: "query dao balance from proto account with proto cdec", memoryNodeFn: NewInMemoryTendermintNodeProto, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 2}}},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if tc.upgrades != nil { // NOTE: Use to perform neccesary upgrades for test
				codec.UpgradeHeight = tc.upgrades.codecUpgrade.height
				_ = memCodecMod(tc.upgrades.codecUpgrade.upgradeMod)
			}
			_, _, cleanup := tc.memoryNodeFn(t, oneAppTwoNodeGenesis())
			memCli, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
			<-evtChan // Wait for block
			got, err := governance.QueryACL(memCodec(), memCli, VCA.LastBlockHeight())
			assert.Nil(t, err)
			assert.Equal(t, got, testACL)

			cleanup()
			stopCli()
		})
	}
}

func TestQueryDaoOwner(t *testing.T) {
	tt := []struct {
		name         string
		memoryNodeFn func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		*upgrades
	}{
		{name: "query dao owner from amino account with amino codec", memoryNodeFn: NewInMemoryTendermintNodeAmino, upgrades: &upgrades{codecUpgrade: codecUpgrade{false, 7000}}},
		{name: "query dao owner from proto account with proto cdec", memoryNodeFn: NewInMemoryTendermintNodeProto, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 2}}},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if tc.upgrades != nil { // NOTE: Use to perform neccesary upgrades for test
				codec.UpgradeHeight = tc.upgrades.codecUpgrade.height
				_ = memCodecMod(tc.upgrades.codecUpgrade.upgradeMod)
			}
			_, _, cleanup := tc.memoryNodeFn(t, oneAppTwoNodeGenesis())
			kb := getInMemoryKeybase()
			cb, err := kb.GetCoinbase()
			if err != nil {
				t.Fatal(err)
			}
			memCli, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
			<-evtChan // Wait for block
			got, err := governance.QueryDAOOwner(memCodec(), memCli, VCA.LastBlockHeight())
			assert.Nil(t, err)
			assert.Equal(t, got.String(), cb.GetAddress().String())

			cleanup()
			stopCli()
		})
	}
}

func TestQueryUpgrade(t *testing.T) {
	tt := []struct {
		name         string
		memoryNodeFn func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		*upgrades
	}{
		{name: "query upgrade with amino codec", memoryNodeFn: NewInMemoryTendermintNodeAmino, upgrades: &upgrades{codecUpgrade: codecUpgrade{false, 7000}}},
		{name: "query upgrade with proto codec", memoryNodeFn: NewInMemoryTendermintNodeAmino, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 2}}},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if tc.upgrades != nil { // NOTE: Use to perform neccesary upgrades for test
				codec.UpgradeHeight = tc.upgrades.codecUpgrade.height
				_ = memCodecMod(tc.upgrades.codecUpgrade.upgradeMod)
			}
			_, _, cleanup := tc.memoryNodeFn(t, oneAppTwoNodeGenesis())
			memCli, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
			<-evtChan // Wait for block
			var err error
			got, err := governance.QueryUpgrade(memCodec(), memCli, VCA.LastBlockHeight())
			assert.Nil(t, err)
			assert.Equal(t, got.UpgradeHeight(), int64(10000))

			cleanup()
			stopCli()
		})
	}
}

func TestQuerySupply(t *testing.T) {
	tt := []struct {
		name         string
		memoryNodeFn func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		*upgrades
	}{
		{name: "query supply amino", memoryNodeFn: NewInMemoryTendermintNodeAmino},
		{name: "query supply proto", memoryNodeFn: NewInMemoryTendermintNodeProto},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if tc.upgrades != nil { // NOTE: Use to perform neccesary upgrades for test
				codec.UpgradeHeight = tc.upgrades.codecUpgrade.height
				_ = memCodecMod(tc.upgrades.codecUpgrade.upgradeMod)
			}
			_, _, cleanup := tc.memoryNodeFn(t, oneAppTwoNodeGenesis())
			_, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
			<-evtChan // Wait for block
			gotStaked, total, err := VCA.QueryTotalServicerCoins(VCA.LastBlockHeight())
			//fmt.Println(err)
			assert.Nil(t, err)
			//fmt.Println(gotStaked, total)
			assert.True(t, gotStaked.Equal(sdk.NewInt(1000000000000000)))
			assert.True(t, total.Equal(sdk.NewInt(1000002010001000)))

			cleanup()
			stopCli()
		})
	}
}

func TestQueryPOSParams(t *testing.T) {
	tt := []struct {
		name         string
		memoryNodeFn func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		*upgrades
	}{
		{name: "query POS params amino", memoryNodeFn: NewInMemoryTendermintNodeAmino, upgrades: &upgrades{codecUpgrade: codecUpgrade{false, 7000}}},
		{name: "query POS params proto", memoryNodeFn: NewInMemoryTendermintNodeProto, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 2}}},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			_, _, cleanup := tc.memoryNodeFn(t, oneAppTwoNodeGenesis())
			_, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
			<-evtChan // Wait for block
			got, err := VCA.QueryServicerParams(VCA.LastBlockHeight())
			assert.Nil(t, err)
			assert.Equal(t, int64(5000), got.MaxValidators)
			assert.Equal(t, int64(1000000), got.StakeMinimum)
			assert.Equal(t, int64(10), got.DAOAllocation)
			assert.Equal(t, sdk.DefaultStakeDenom, got.StakeDenom)

			cleanup()
			stopCli()
		})
	}
}

func TestAccountBalance(t *testing.T) {
	tt := []struct {
		name         string
		memoryNodeFn func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		*upgrades
	}{
		{name: "query account balance from amino", memoryNodeFn: NewInMemoryTendermintNodeAmino, upgrades: &upgrades{codecUpgrade: codecUpgrade{false, 7000}}},
		{name: "query account balance from proto", memoryNodeFn: NewInMemoryTendermintNodeProto, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 2}}},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if tc.upgrades != nil { // NOTE: Use to perform neccesary upgrades for test
				codec.UpgradeHeight = tc.upgrades.codecUpgrade.height
				_ = memCodecMod(tc.upgrades.codecUpgrade.upgradeMod)
			}
			_, kb, cleanup := tc.memoryNodeFn(t, oneAppTwoNodeGenesis())
			cb, err := kb.GetCoinbase()
			assert.Nil(t, err)
			_, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
			<-evtChan // Wait for block
			got, err := VCA.QueryBalance(cb.GetAddress().String(), VCA.LastBlockHeight())
			assert.Nil(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, got, sdk.NewInt(1000000000))

			cleanup()
			stopCli()
		})
	}
}

func TestQuerySigningInfo(t *testing.T) {
	tt := []struct {
		name         string
		memoryNodeFn func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		*upgrades
	}{
		{name: "query signign info amino ", memoryNodeFn: NewInMemoryTendermintNodeAmino, upgrades: &upgrades{codecUpgrade: codecUpgrade{false, 7000}}},
		{name: "query signing info proto", memoryNodeFn: NewInMemoryTendermintNodeProto, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 2}}},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			_, kb, cleanup := tc.memoryNodeFn(t, oneAppTwoNodeGenesis())
			cb, err := kb.GetCoinbase()
			assert.Nil(t, err)
			cbAddr := cb.GetAddress()
			assert.Nil(t, err)
			_, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
			<-evtChan // Wait for block
			got, err := VCA.QuerySigningInfo(0, cbAddr.String())
			assert.Nil(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, got.Address.String(), cbAddr.String())

			cleanup()
			stopCli()
		})
	}
}

func TestQueryViperSupportedBlockchains(t *testing.T) {
	tt := []struct {
		name         string
		memoryNodeFn func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		*upgrades
	}{
		{name: "query supported blockchains amino with amino codec", memoryNodeFn: NewInMemoryTendermintNodeAmino, upgrades: &upgrades{codecUpgrade: codecUpgrade{false, 7000}}},
		{name: "query supported blockchains from proto with proto codec", memoryNodeFn: NewInMemoryTendermintNodeProto, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 2}}},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if tc.upgrades != nil { // NOTE: Use to perform neccesary upgrades for test
				codec.UpgradeHeight = tc.upgrades.codecUpgrade.height
				_ = memCodecMod(tc.upgrades.codecUpgrade.upgradeMod)
			}
			_, _, cleanup := tc.memoryNodeFn(t, oneAppTwoNodeGenesis())
			_, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
			<-evtChan // Wait for block
			var err error
			got, err := VCA.QueryViperSupportedBlockchains(VCA.LastBlockHeight())
			assert.Nil(t, err)
			assert.NotNil(t, got)
			assert.Contains(t, got, sdk.PlaceholderHash)

			cleanup()
			stopCli()
		})
	}
}

func TestQueryViperParams(t *testing.T) {
	tt := []struct {
		name         string
		memoryNodeFn func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		*upgrades
	}{
		{name: "query viper params amino ", memoryNodeFn: NewInMemoryTendermintNodeAmino, upgrades: &upgrades{codecUpgrade: codecUpgrade{false, 7000}}},
		{name: "query viper params proto", memoryNodeFn: NewInMemoryTendermintNodeProto, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 2}}},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			_, _, cleanup := tc.memoryNodeFn(t, oneAppTwoNodeGenesis())
			_, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
			<-evtChan // Wait for block
			got, err := VCA.QueryViperParams(VCA.LastBlockHeight())
			assert.Nil(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, int64(3), got.ClaimSubmissionWindow)
			assert.Equal(t, int64(100), got.ClaimExpiration)
			assert.Contains(t, got.SupportedBlockchains, sdk.PlaceholderHash)

			cleanup()
			stopCli()
		})
	}
}

func TestQueryAccount(t *testing.T) {
	tt := []struct {
		name         string
		memoryNodeFn func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		*upgrades
	}{
		{name: "query account amino", memoryNodeFn: NewInMemoryTendermintNodeAmino, upgrades: &upgrades{codecUpgrade: codecUpgrade{false, 7000}}},
		{name: "query account proto", memoryNodeFn: NewInMemoryTendermintNodeProto, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 2}}},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if tc.upgrades != nil { // NOTE: Use to perform neccesary upgrades for test
				codec.UpgradeHeight = tc.upgrades.codecUpgrade.height
				_ = memCodecMod(tc.upgrades.codecUpgrade.upgradeMod)
			}
			_, kb, cleanup := tc.memoryNodeFn(t, oneAppTwoNodeGenesis())
			_, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
			acc := getUnstakedAccount(kb)
			assert.NotNil(t, acc)
			<-evtChan // Wait for block
			got, err := VCA.QueryAccount(acc.GetAddress().String(), VCA.LastBlockHeight())
			assert.Nil(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, acc.GetAddress(), (*got).GetAddress())

			cleanup()
			stopCli()
		})
	}
}

func TestQueryStakedApp(t *testing.T) {
	tt := []struct {
		name         string
		memoryNodeFn func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		*upgrades
	}{
		{name: "query query staked app amino", memoryNodeFn: NewInMemoryTendermintNodeAmino, upgrades: &upgrades{codecUpgrade: codecUpgrade{false, 7000}}},
		{name: "query query staked app proto", memoryNodeFn: NewInMemoryTendermintNodeProto, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 2}}},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if tc.upgrades != nil { // NOTE: Use to perform neccesary upgrades for test
				codec.UpgradeHeight = tc.upgrades.codecUpgrade.height
				_ = memCodecMod(tc.upgrades.codecUpgrade.upgradeMod)
			}
			_, kb, cleanup := tc.memoryNodeFn(t, oneAppTwoNodeGenesis())
			time.Sleep(2 * time.Second)
			kp, err := kb.GetCoinbase()
			assert.Nil(t, err)
			_, _, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
			var tx *sdk.TxResponse
			var chains = []string{"0001"}
			var geozones = []string{"0001"}
			<-evtChan // Wait for block
			memCli, stopCli, evtChan := subscribeTo(t, tmTypes.EventTx)
			tx, err = requestors.StakeTx(memCodec(), memCli, kb, chains, geozones, 5, sdk.NewInt(1000000), kp, "test", tc.upgrades.codecUpgrade.upgradeMod)
			assert.Nil(t, err)
			assert.NotNil(t, tx)

			<-evtChan // Wait for  tx
			got, err := VCA.QueryRequestor(kp.GetAddress().String(), VCA.LastBlockHeight())
			assert.Nil(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, sdk.Staked, got.Status)
			assert.Equal(t, false, got.Jailed)

			cleanup()
			stopCli()
		})
	}
}

func TestRelayGenerator(t *testing.T) {
	const appPrivKey = "70906c8e250352e811a6ca994b674c4da1c6ba4be1e0b3edeadaf59979236c96a25e182d490e9722e72ba90eb21fe0124d03bcb75d2bf6f45b2a1d2b1dc92fac"
	const nodePublicKey = "a25e182d490e9722e72ba90eb21fe0124d03bcb75d2bf6f45b2a1d2b1dc92fac"
	const sessionBlockheight = 1
	const query = `{"jsonrpc":"2.0","method":"net_version","params":[],"id":67}`
	const supportedBlockchain = "0001"
	apkBz, err := hex.DecodeString(appPrivKey)
	if err != nil {
		panic(err)
	}
	var requestorPrivateKey crypto.Ed25519PrivateKey
	copy(requestorPrivateKey[:], apkBz)
	aat := types.AAT{
		Version:           "0.0.1",
		RequestorPublicKey: requestorPrivateKey.PublicKey().RawString(),
		ClientPublicKey:   requestorPrivateKey.PublicKey().RawString(),
		RequestorSignature: "",
	}
	sig, err := requestorPrivateKey.Sign(aat.Hash())
	if err != nil {
		panic(err)
	}
	aat.RequestorSignature = hex.EncodeToString(sig)
	payload := types.Payload{
		Data: query,
	}
	// setup relay
	relay := types.Relay{
		Payload: payload,
		Proof: types.RelayProof{
			Entropy:            int64(rand.Int()),
			SessionBlockHeight: sessionBlockheight,
			ServicerPubKey:     nodePublicKey,
			Blockchain:         supportedBlockchain,
			Token:              aat,
			Signature:          "",
		},
	}
	relay.Proof.RequestHash = relay.RequestHashString()
	sig, err = requestorPrivateKey.Sign(relay.Proof.Hash())
	if err != nil {
		panic(err)
	}
	relay.Proof.Signature = hex.EncodeToString(sig)
	_, err = json.MarshalIndent(relay, "", "  ")
	if err != nil {
		panic(err)
	}
}

func TestQueryRelay(t *testing.T) {
	const headerKey = "foo"
	const headerVal = "bar"

	expectedRequest := `"jsonrpc":"2.0","method":"web3_sha3","params":["0x68656c6c6f20776f726c64"],"id":64`
	expectedResponse := "0x47173285a8d7341e5e972fc677286384f802f8ef42a5ec5f03bbfa254cb01fad"
	tt := []struct {
		name         string
		memoryNodeFn func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		*upgrades
	}{
		{name: "query relay amino", memoryNodeFn: NewInMemoryTendermintNodeAmino, upgrades: &upgrades{codecUpgrade: codecUpgrade{false, 7000}}},
		{name: "query relay proto", memoryNodeFn: NewInMemoryTendermintNodeProto, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 2}}},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			sdk.VbCCache = sdk.NewCache(1)
			if tc.upgrades != nil { // NOTE: Use to perform neccesary upgrades for test
				codec.UpgradeHeight = tc.upgrades.codecUpgrade.height
				_ = memCodecMod(tc.upgrades.codecUpgrade.upgradeMod)
			}
			time.Sleep(time.Second * 2)
			genBz, _, validators, app := fiveValidatorsOneAppGenesis()
			// setup relay endpoint
			gock.New(sdk.PlaceholderURL).
				Post("").
				BodyString(expectedRequest).
				MatchHeader(headerKey, headerVal).
				Reply(200).
				BodyString(expectedResponse)
			_, kb, cleanup := tc.memoryNodeFn(t, genBz)
			appPrivateKey, err := kb.ExportPrivateKeyObject(app.Address, "test")
			assert.Nil(t, err)
			// setup AAT
			aat := types.AAT{
				Version:           "0.0.1",
				RequestorPublicKey: appPrivateKey.PublicKey().RawString(),
				ClientPublicKey:   appPrivateKey.PublicKey().RawString(),
				RequestorSignature: "",
			}
			sig, err := appPrivateKey.Sign(aat.Hash())
			if err != nil {
				panic(err)
			}
			aat.RequestorSignature = hex.EncodeToString(sig)
			payload := types.Payload{
				Data:    expectedRequest,
				Headers: map[string]string{headerKey: headerVal},
			}
			// setup relay
			relay := types.Relay{
				Payload: payload,
				Meta:    types.RelayMeta{BlockHeight: 5}, // todo race condition here
				Proof: types.RelayProof{
					Entropy:            32598345349034509,
					SessionBlockHeight: 1,
					ServicerPubKey:     validators[0].PublicKey.RawString(),
					Blockchain:         sdk.PlaceholderHash,
					Token:              aat,
					Signature:          "",
				},
			}
			relay.Proof.RequestHash = relay.RequestHashString()
			sig, err = appPrivateKey.Sign(relay.Proof.Hash())
			if err != nil {
				panic(err)
			}
			relay.Proof.Signature = hex.EncodeToString(sig)
			_, _, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
			<-evtChan // Wait for block
			res, _, err := VCA.HandleRelay(relay)
			assert.Nil(t, err, err)
			assert.Equal(t, expectedResponse, res.Response)
			gock.New(sdk.PlaceholderURL).
				Post("").
				BodyString(expectedRequest).
				Reply(200).
				BodyString(expectedResponse)
			_, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
			select {
			case <-evtChan:
				inv, err := types.GetEvidence(types.SessionHeader{
					RequestorPubKey:     aat.RequestorPublicKey,
					Chain:              relay.Proof.Blockchain,
					SessionBlockHeight: relay.Proof.SessionBlockHeight,
				}, types.RelayEvidence, sdk.NewInt(10000), types.GlobalEvidenceCache)
				assert.Nil(t, err)
				assert.NotNil(t, inv)
				assert.Equal(t, inv.NumOfProofs, int64(1))
				cleanup()
				stopCli()
				gock.Off()
				return
			}
		})
	}
}
func TestQueryDispatch(t *testing.T) {
	tt := []struct {
		name         string
		memoryNodeFn func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		*upgrades
	}{
		{name: "query dispatch amino", memoryNodeFn: NewInMemoryTendermintNodeAmino, upgrades: &upgrades{codecUpgrade: codecUpgrade{false, 7000}}},
		{name: "query dispatch proto", memoryNodeFn: NewInMemoryTendermintNodeProto, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 2}}},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			sdk.VbCCache = sdk.NewCache(1)
			if tc.upgrades != nil { // NOTE: Use to perform neccesary upgrades for test
				codec.UpgradeHeight = tc.upgrades.codecUpgrade.height
				_ = memCodecMod(tc.upgrades.codecUpgrade.upgradeMod)
			}
			genBz, _, validators, app := fiveValidatorsOneAppGenesis()
			_, kb, cleanup := tc.memoryNodeFn(t, genBz)
			requestorPrivateKey, err := kb.ExportPrivateKeyObject(app.Address, "test")
			assert.Nil(t, err)
			// Setup HandleDispatch Request
			key := types.SessionHeader{
				RequestorPubKey:     requestorPrivateKey.PublicKey().RawString(),
				Chain:              sdk.PlaceholderHash,
				SessionBlockHeight: 1,
			}
			// setup the query
			_, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
			<-evtChan // Wait for block
			res, err := VCA.HandleDispatch(key)
			assert.Nil(t, err)
			for _, val := range validators {
				assert.Contains(t, res.Session.SessionServicers, val)
			}
			cleanup()
			stopCli()
		})
	}
}

func TestQueryAllParams(t *testing.T) {

	tt := []struct {
		name         string
		memoryNodeFn func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		*upgrades
	}{
		{name: "query all params amino ", memoryNodeFn: NewInMemoryTendermintNodeAmino},
		{name: "query all params proto", memoryNodeFn: NewInMemoryTendermintNodeProto},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			resetTestACL()
			if tc.upgrades != nil { // NOTE: Use to perform neccesary upgrades for test
				codec.UpgradeHeight = tc.upgrades.codecUpgrade.height
				_ = memCodecMod(tc.upgrades.codecUpgrade.upgradeMod)
			}
			_, _, cleanup := tc.memoryNodeFn(t, oneAppTwoNodeGenesis())
			res, err := VCA.QueryAllParams(VCA.LastBlockHeight())
			assert.Nil(t, err)
			assert.NotNil(t, res)

			assert.NotZero(t, len(res.AppParams))
			cleanup()
		})
	}
}
func TestQueryParam(t *testing.T) {

	tt := []struct {
		name         string
		memoryNodeFn func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		*upgrades
	}{
		{name: "query param amino ", memoryNodeFn: NewInMemoryTendermintNodeAmino},
		{name: "query param proto ", memoryNodeFn: NewInMemoryTendermintNodeProto},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			resetTestACL()
			_, _, cleanup := tc.memoryNodeFn(t, oneAppTwoNodeGenesis())
			res, err := VCA.QueryParam(0, "vipernet/SupportedBlockchains")
			assert.Nil(t, err)
			assert.NotNil(t, res)

			assert.NotNil(t, res.Value)
			cleanup()
		})
	}
}

func TestQueryAccountBalance(t *testing.T) {

	tt := []struct {
		name         string
		memoryNodeFn func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		*upgrades
	}{
		{name: "query staked app amino with amino codec", memoryNodeFn: NewInMemoryTendermintNodeAmino, upgrades: &upgrades{codecUpgrade: codecUpgrade{false, 7000}}},
		// {name: "query staked app from amino with proto codec", memoryNodeFn: NewInMemoryTendermintNodeAmino, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 2}}},
		{name: "query staked app params from proto with proto codec", memoryNodeFn: NewInMemoryTendermintNodeProto, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 2}}},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if tc.upgrades != nil { // NOTE: Use to perform neccesary upgrades for test
				codec.UpgradeHeight = tc.upgrades.codecUpgrade.height
				_ = memCodecMod(tc.upgrades.codecUpgrade.upgradeMod)
			}
			_, kb, cleanup := tc.memoryNodeFn(t, oneAppTwoNodeGenesis())
			_, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
			acc := getUnstakedAccount(kb)
			assert.NotNil(t, acc)
			<-evtChan // Wait for block
			got, err := VCA.QueryBalance(acc.GetAddress().String(), VCA.LastBlockHeight())
			assert.Nil(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, sdk.NewInt(1000000000), got)
			cleanup()
			stopCli()
		})
	}
}

func TestQueryNonExistingAccountBalance(t *testing.T) {
	tt := []struct {
		name         string
		memoryNodeFn func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		*upgrades
	}{
		{name: "query non existing account balance amino with amino codec", memoryNodeFn: NewInMemoryTendermintNodeAmino, upgrades: &upgrades{codecUpgrade: codecUpgrade{false, 7000}}},
		{name: "query staked app from amino with proto codec", memoryNodeFn: NewInMemoryTendermintNodeProto, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 2}}},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			_, _, cleanup := tc.memoryNodeFn(t, oneAppTwoNodeGenesis())
			_, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
			<-evtChan // Wait for block
			got, err := VCA.QueryBalance("802fddec29f99cae7a601cf648eafced1c062d39", VCA.LastBlockHeight())
			assert.Nil(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, sdk.NewInt(0), got)
			cleanup()
			stopCli()
		})
	}
}

func TestQueryAccounts(t *testing.T) {
	tt := []struct {
		name         string
		memoryNodeFn func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		*upgrades
	}{
		{name: "query accounts amino", memoryNodeFn: NewInMemoryTendermintNodeAmino, upgrades: &upgrades{codecUpgrade: codecUpgrade{false, 7000}}},
		{name: "query accounts proto", memoryNodeFn: NewInMemoryTendermintNodeProto, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 2}}},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if tc.upgrades != nil { // NOTE: Use to perform neccesary upgrades for test
				codec.UpgradeHeight = tc.upgrades.codecUpgrade.height
				_ = memCodecMod(tc.upgrades.codecUpgrade.upgradeMod)
			}
			_, _, cleanup := tc.memoryNodeFn(t, oneAppTwoNodeGenesis())
			got, err := VCA.QueryAccounts(VCA.LastBlockHeight(), 1, 1)
			assert.Nil(t, err)
			assert.NotNil(t, got)
			assert.NotEqual(t, 1, got.Total)

			cleanup()
		})
	}
}
*/
