// nolint
package app

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	rand2 "github.com/tendermint/tendermint/libs/rand"
	"github.com/tendermint/tendermint/node"
	tmTypes "github.com/tendermint/tendermint/types"
	"github.com/vipernet-xyz/viper-network/codec"
	"github.com/vipernet-xyz/viper-network/crypto"
	"github.com/vipernet-xyz/viper-network/crypto/keys"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/authentication/types"
	"github.com/vipernet-xyz/viper-network/x/governance"
	govTypes "github.com/vipernet-xyz/viper-network/x/governance/types"
	providers "github.com/vipernet-xyz/viper-network/x/providers"
	providersTypes "github.com/vipernet-xyz/viper-network/x/providers/types"
	"github.com/vipernet-xyz/viper-network/x/servicers"
	servicersTypes "github.com/vipernet-xyz/viper-network/x/servicers/types"
	viperTypes "github.com/vipernet-xyz/viper-network/x/vipernet/types"
)

func TestMain(m *testing.M) {
	viperTypes.CleanViperNodes()
	sdk.InitCtxCache(1)
	m.Run()
}

func TestUnstakeApp(t *testing.T) {
	tt := []struct {
		name         string
		memoryNodeFn func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		*upgrades
	}{
		{name: "unstake an amino app with amino codec", memoryNodeFn: NewInMemoryTendermintNodeAmino, upgrades: &upgrades{codecUpgrade: codecUpgrade{false, 7000}}},
		{name: "unstake a proto app with proto codec", memoryNodeFn: NewInMemoryTendermintNodeProto, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 2}}}, // todo: FULL PROTO SCENARIO
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if tc.upgrades != nil { // NOTE: Use to perform neccesary upgrades for test
				codec.UpgradeHeight = tc.upgrades.codecUpgrade.height
				_ = memCodecMod(tc.upgrades.codecUpgrade.upgradeMod)
			}
			_, kb, cleanup := tc.memoryNodeFn(t, oneAppTwoNodeGenesis())
			time.Sleep(1 * time.Second)
			kp, err := kb.GetCoinbase()
			assert.Nil(t, err)
			_, _, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
			var tx *sdk.TxResponse
			var chains = []string{"0001"}
			<-evtChan // Wait for block
			memCli, _, evtChan := subscribeTo(t, tmTypes.EventTx)
			tx, err = providers.StakeTx(memCodec(), memCli, kb, chains, sdk.NewInt(1000000), kp, "test", tc.codecUpgrade.upgradeMod)
			assert.Nil(t, err)
			assert.NotNil(t, tx)

			<-evtChan // Wait for tx
			got, err := VCA.QueryProviders(VCA.LastBlockHeight(), providersTypes.QueryProvidersWithOpts{
				Page:  1,
				Limit: 1})
			assert.Nil(t, err)
			res := got.Result.(providersTypes.Providers)
			assert.Equal(t, 1, len(res))
			memCli, stopCli, evtChan := subscribeTo(t, tmTypes.EventTx)
			_, _ = providers.UnstakeTx(memCodec(), memCli, kb, kp.GetAddress(), "test", tc.codecUpgrade.upgradeMod)

			<-evtChan // Wait for tx
			got, err = VCA.QueryProviders(VCA.LastBlockHeight(), providersTypes.QueryProvidersWithOpts{
				Page:          1,
				Limit:         1,
				StakingStatus: 1,
			})
			assert.Nil(t, err)
			res = got.Result.(providersTypes.Providers)
			assert.Equal(t, 1, len(res))
			got, err = VCA.QueryProviders(VCA.LastBlockHeight(), providersTypes.QueryProvidersWithOpts{
				Page:          1,
				Limit:         1,
				StakingStatus: 2,
			})
			assert.Nil(t, err)
			res = got.Result.(providersTypes.Providers)
			assert.Equal(t, 1, len(res)) // default genesis application

			cleanup()
			stopCli()
		})
	}
}
func TestStakeApp(t *testing.T) {
	tt := []struct {
		name         string
		memoryNodeFn func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		*upgrades
	}{
		{name: "stake app with amino codec", memoryNodeFn: NewInMemoryTendermintNodeAmino, upgrades: &upgrades{codecUpgrade: codecUpgrade{false, 7000}}},
		{name: "stake a proto app with proto codec", memoryNodeFn: NewInMemoryTendermintNodeProto, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 2}}}, // TODO FULL PROTO SCENARIO
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if tc.upgrades != nil { // NOTE: Use to perform neccesary upgrades for test
				codec.UpgradeHeight = tc.upgrades.codecUpgrade.height
				_ = memCodecMod(tc.upgrades.codecUpgrade.upgradeMod)
			}

			_, kb, cleanup := tc.memoryNodeFn(t, oneAppTwoNodeGenesis())
			time.Sleep(1 * time.Second)
			kp, err := kb.GetCoinbase()
			assert.Nil(t, err)
			_, _, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
			var tx *sdk.TxResponse
			var chains = []string{"0001"}

			<-evtChan // Wait for block
			memCli, stopCli, evtChan := subscribeTo(t, tmTypes.EventTx)
			tx, err = providers.StakeTx(memCodec(), memCli, kb, chains, sdk.NewInt(1000000), kp, "test", tc.codecUpgrade.upgradeMod)
			assert.Nil(t, err)
			assert.NotNil(t, tx)

			<-evtChan // Wait for tx
			got, err := VCA.QueryProviders(VCA.LastBlockHeight(), providersTypes.QueryProvidersWithOpts{
				Page:  1,
				Limit: 2,
			})
			assert.Nil(t, err)
			res := got.Result.(providersTypes.Providers)
			assert.Equal(t, 2, len(res))

			stopCli()
			cleanup()
		})
	}
}
func TestEditStakeApp(t *testing.T) {
	tt := []struct {
		name         string
		memoryNodeFn func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		*upgrades
	}{
		{name: "editStake a proto application with proto codec", memoryNodeFn: NewInMemoryTendermintNodeProto, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 2}}},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if tc.upgrades != nil { // NOTE: Use to perform neccesary upgrades for test
				codec.UpgradeHeight = tc.upgrades.codecUpgrade.height
				_ = memCodecMod(tc.upgrades.codecUpgrade.upgradeMod)
			}
			var newChains = []string{"2121"}
			_, kb, cleanup := tc.memoryNodeFn(t, oneAppTwoNodeGenesis())
			time.Sleep(1 * time.Second)
			kp, err := kb.GetCoinbase()
			assert.Nil(t, err)
			kps, err := kb.List()
			assert.Nil(t, err)
			for _, k := range kps {
				if !k.GetAddress().Equals(kp.GetAddress()) {
					kp = k
					break
				}
			}
			_, _, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
			var tx *sdk.TxResponse
			<-evtChan // Wait for block
			memCli, stopCli, evtChan := subscribeTo(t, tmTypes.EventTx)
			balance, err := VCA.QueryBalance(kp.GetAddress().String(), VCA.LastBlockHeight())
			assert.Nil(t, err)
			n, err := VCA.QueryProvider(kp.GetAddress().String(), VCA.LastBlockHeight())
			assert.Nil(t, err)
			var newBalance = balance.Sub(sdk.NewInt(100000)).Add(n.StakedTokens)
			tx, err = providers.StakeTx(memCodec(), memCli, kb, newChains, newBalance, kp, "test", tc.upgrades.codecUpgrade.upgradeMod)
			assert.Nil(t, err)
			assert.NotNil(t, tx)
			<-evtChan // Wait for tx
			appUpdated, err := VCA.QueryProvider(kp.GetAddress().String(), VCA.LastBlockHeight())
			assert.Nil(t, err)
			// assert not the same as the old node
			assert.NotEqual(t, appUpdated, n)
			// assert chains and stake updated
			assert.Equal(t, newChains, appUpdated.Chains)
			// assert chains and stake updated
			assert.Equal(t, newBalance, appUpdated.StakedTokens)
			cleanup()
			stopCli()
		})
	}
}

func TestUnstakeNode(t *testing.T) {
	tt := []struct {
		name           string
		memoryNodeFn   func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		outputIsSigner bool
		*upgrades
	}{
		{name: "unstake a proto node with proto codec", memoryNodeFn: NewInMemoryTendermintNodeProto, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 2}}},
		{name: "unstake an amino node with amino codec", memoryNodeFn: NewInMemoryTendermintNodeAmino, upgrades: &upgrades{codecUpgrade: codecUpgrade{false, 7000}}},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if tc.upgrades != nil { // NOTE: Use to perform neccesary upgrades for test
				codec.UpgradeHeight = tc.upgrades.codecUpgrade.height
				_ = memCodecMod(tc.upgrades.codecUpgrade.upgradeMod)
			}

			var chains = []string{"0001"}
			gen, _ := twoValTwoNodeGenesisState()
			_, kb, cleanup := tc.memoryNodeFn(t, gen)
			time.Sleep(1 * time.Second)
			kp, err := kb.GetCoinbase()
			assert.Nil(t, err)
			_, _, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
			var tx *sdk.TxResponse
			<-evtChan // Wait for block
			memCli, _, evtChan := subscribeTo(t, tmTypes.EventTx)
			_, err = VCA.QueryBalance(kp.GetAddress().String(), VCA.LastBlockHeight())
			assert.Nil(t, err)
			signer := kp.GetAddress()
			if tc.outputIsSigner {
				list, err := kb.List()
				assert.Nil(t, err)
				signer = list[2].GetAddress()
			}
			tx, err = servicers.UnstakeTx(memCodec(), memCli, kb, kp.GetAddress(), signer, "test", tc.upgrades.codecUpgrade.upgradeMod)
			assert.Nil(t, err)
			assert.NotNil(t, tx)
			<-evtChan // Wait for tx
			_, _, evtChan = subscribeTo(t, tmTypes.EventNewBlockHeader)
			for {
				select {
				case res := <-evtChan:
					if len(res.Events["begin_unstake.module"]) == 1 {
						got, err := VCA.QueryServicers(VCA.LastBlockHeight(), servicersTypes.QueryValidatorsParams{StakingStatus: 1, JailedStatus: 0, Blockchain: "", Page: 1, Limit: 1}) // unstaking
						assert.Nil(t, err)
						res := got.Result.([]servicersTypes.Validator)
						assert.Equal(t, 1, len(res))
						got, err = VCA.QueryServicers(VCA.LastBlockHeight(), servicersTypes.QueryValidatorsParams{StakingStatus: 2, JailedStatus: 0, Blockchain: "", Page: 1, Limit: 1}) // staked
						assert.Nil(t, err)
						res = got.Result.([]servicersTypes.Validator)
						assert.Equal(t, 1, len(res))
						memCli, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlockHeader)
						header := <-evtChan // Wait for header
						if len(header.Events["unstake.module"]) == 1 {
							got, err := VCA.QueryServicers(VCA.LastBlockHeight(), servicersTypes.QueryValidatorsParams{StakingStatus: 0, JailedStatus: 0, Blockchain: "", Page: 1, Limit: 1})
							assert.Nil(t, err)
							res := got.Result.([]servicersTypes.Validator)
							assert.Equal(t, 1, len(res))
							vals := got.Result.([]servicersTypes.Validator)
							addr := vals[0].Address
							balance, err := VCA.QueryBalance(addr.String(), VCA.LastBlockHeight())
							assert.Nil(t, err)
							assert.NotEqual(t, balance, sdk.ZeroInt())
							tx, err = servicers.StakeTx(memCodec(), memCli, kb, chains, "https://myViperNode.com:8080", sdk.NewInt(10000000), kp, signer, "test", tc.upgrades.codecUpgrade.upgradeMod, signer)
							assert.Nil(t, err)
							assert.NotNil(t, tx)
							assert.Equal(t, tx.Code, uint32(0x0))
							cleanup()
							stopCli()

						}
						return
					}
				default:
					continue
				}
			}
		})
	}

}
func TestStakeNode(t *testing.T) {
	tt := []struct {
		name           string
		memoryNodeFn   func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		outputIsSigner bool
		*upgrades
	}{
		{name: "stake node with amino codec", memoryNodeFn: NewInMemoryTendermintNodeAmino, upgrades: &upgrades{codecUpgrade: codecUpgrade{false, 7000}}},
		{name: "stake a proto node with proto codec", memoryNodeFn: NewInMemoryTendermintNodeProto, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 2}}},
		{name: "stake a proto node with proto codec bad signer", memoryNodeFn: NewInMemoryTendermintNodeProto, outputIsSigner: true, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 2}}},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if tc.upgrades != nil { // NOTE: Use to perform neccesary upgrades for test
				codec.UpgradeHeight = tc.upgrades.codecUpgrade.height
				_ = memCodecMod(tc.upgrades.codecUpgrade.upgradeMod)
			}

			gen, vals := twoValTwoNodeGenesisState()
			_, kb, cleanup := tc.memoryNodeFn(t, gen)
			time.Sleep(1 * time.Second)
			kp, err := kb.GetCoinbase()
			signer := kp.GetAddress()
			if tc.outputIsSigner {
				for _, val := range vals {
					if val.Address.String() != signer.String() {
						signer = val.Address
						break
					}
				}
			}
			assert.Nil(t, err)
			_, _, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
			var tx *sdk.TxResponse
			var chains = []string{"0001"}
			<-evtChan // Wait for block
			memCli, stopCli, _ := subscribeTo(t, tmTypes.EventTx)
			tx, err = servicers.StakeTx(memCodec(), memCli, kb, chains, "https://myViperNode.com:8080", sdk.NewInt(10000000), kp, signer, "test", tc.upgrades.codecUpgrade.upgradeMod, signer)
			assert.Nil(t, err)
			assert.NotNil(t, tx)
			if tc.outputIsSigner {
				assert.Equal(t, 4, int(tx.Code))
			} else {
				assert.Equal(t, 0, int(tx.Code))
			}
			cleanup()
			stopCli()

		})
	}
}
func TestEditStakeNode(t *testing.T) {
	tt := []struct {
		name           string
		memoryNodeFn   func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		outputIsSigner bool
		*upgrades
	}{
		{name: "editStake after proto upgrade", memoryNodeFn: NewInMemoryTendermintNodeProto, outputIsSigner: false, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 2}}},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			codec.TestMode = 0
			if tc.upgrades != nil { // NOTE: Use to perform neccesary upgrades for test
				codec.UpgradeHeight = tc.upgrades.codecUpgrade.height
				_ = memCodecMod(tc.upgrades.codecUpgrade.upgradeMod)
			}

			{
				codec.TestMode = -2
			}
			var newChains = []string{"2121"}
			var newServiceURL = "https://newServiceUrl.com:8081"
			gen, vals := twoValTwoNodeGenesisState()
			_, kb, cleanup := tc.memoryNodeFn(t, gen)
			time.Sleep(1 * time.Second)
			kp, err := kb.GetCoinbase()
			assert.Nil(t, err)
			signer := kp.GetAddress()
			assert.Nil(t, err)
			if tc.outputIsSigner {
				for _, val := range vals {
					if val.Address.String() == signer.String() {
						signer = val.OutputAddress
					}
				}
			}
			_, _, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
			var tx *sdk.TxResponse
			<-evtChan // Wait for block
			memCli, stopCli, evtChan := subscribeTo(t, tmTypes.EventTx)
			balance, err := VCA.QueryBalance(kp.GetAddress().String(), VCA.LastBlockHeight())
			assert.Nil(t, err)
			n, err := VCA.QueryServicer(kp.GetAddress().String(), VCA.LastBlockHeight())
			assert.Nil(t, err)
			var newBalance = balance.Sub(sdk.NewInt(100000)).Add(n.StakedTokens)
			fmt.Println(signer.String())
			tx, err = servicers.StakeTx(memCodec(), memCli, kb, newChains, newServiceURL, newBalance, kp, signer, "test", tc.upgrades.codecUpgrade.upgradeMod, signer)
			assert.Nil(t, err)
			assert.NotNil(t, tx)
			<-evtChan // Wait for tx
			nodeUpdated, err := VCA.QueryServicer(kp.GetAddress().String(), VCA.LastBlockHeight())
			assert.Nil(t, err)
			// assert not the same as the old node
			assert.NotEqual(t, nodeUpdated, n)
			// assert chains, serviceurl, and stake updated
			assert.Equal(t, newChains, nodeUpdated.Chains)
			// assert chains, serviceurl, and stake updated
			assert.Equal(t, newServiceURL, nodeUpdated.ServiceURL)
			// assert chains, serviceurl, and stake updated
			assert.Equal(t, newBalance, nodeUpdated.StakedTokens)
			cleanup()
			stopCli()
		})
	}
}

func TestEditStakeNodeOutput(t *testing.T) {
	tt := []struct {
		name           string
		memoryNodeFn   func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		outputIsSigner bool
		*upgrades
	}{
		{name: "editStake output update flow", memoryNodeFn: NewInMemoryTendermintNodeProto, outputIsSigner: true, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 2}}},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			codec.TestMode = 0
			if tc.upgrades != nil { // NOTE: Use to perform necessary upgrades for test
				codec.UpgradeHeight = tc.upgrades.codecUpgrade.height
				_ = memCodecMod(tc.upgrades.codecUpgrade.upgradeMod)
			}
			codec.TestMode = -2
			var newChains = []string{"2121"}
			var newServiceURL = "https://newServiceUrl.com:8081"
			gen, _ := twoValTwoNodeGenesisState()
			_, kb, cleanup := tc.memoryNodeFn(t, gen)
			time.Sleep(1 * time.Second)
			kp, err := kb.GetCoinbase()
			assert.Nil(t, err)
			signer := kp.GetAddress()
			assert.Nil(t, err)
			_, _, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
			var tx *sdk.TxResponse
			<-evtChan // Wait for block
			memCli, stopCli, evtChan := subscribeTo(t, tmTypes.EventTx)
			balance, err := VCA.QueryBalance(kp.GetAddress().String(), VCA.LastBlockHeight())
			assert.Nil(t, err)
			n, err := VCA.QueryServicer(kp.GetAddress().String(), VCA.LastBlockHeight())
			assert.Nil(t, err)
			var newBalance = balance.Sub(sdk.NewInt(100000)).Add(n.StakedTokens)
			tx, err = servicers.StakeTx(memCodec(), memCli, kb, newChains, newServiceURL, newBalance, kp, signer, "test", tc.upgrades.codecUpgrade.upgradeMod, signer)
			assert.Nil(t, err)
			assert.NotNil(t, tx)
			<-evtChan // Wait for tx
			nodeUpdated, err := VCA.QueryServicer(kp.GetAddress().String(), VCA.LastBlockHeight())
			assert.Nil(t, err)
			// assert not the same as the old node
			assert.NotEqual(t, nodeUpdated, n)
			// assert chains, serviceurl, and stake updated
			assert.Equal(t, newChains, nodeUpdated.Chains)
			// assert chains, serviceurl, and stake updated
			assert.Equal(t, newServiceURL, nodeUpdated.ServiceURL)
			// assert chains, serviceurl, and stake updated
			assert.Equal(t, newBalance, nodeUpdated.StakedTokens)
			codec.TestMode = -3
			newBalance = nodeUpdated.StakedTokens.Add(sdk.NewInt(1000))
			tx, err = servicers.StakeTx(memCodec(), memCli, kb, newChains, newServiceURL, newBalance, kp, signer, "test", tc.upgrades.codecUpgrade.upgradeMod, signer)
			assert.Nil(t, err)
			assert.NotNil(t, tx)
			<-evtChan // Wait for tx
			nodeUpdated, err = VCA.QueryServicer(kp.GetAddress().String(), VCA.LastBlockHeight())
			assert.Nil(t, err)
			assert.Equal(t, newBalance, nodeUpdated.StakedTokens)
			assert.Equal(t, signer, nodeUpdated.OutputAddress)

			cleanup()
			stopCli()
		})
	}
}

func TestSendTransaction(t *testing.T) {
	tt := []struct {
		name         string
		memoryNodeFn func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		*upgrades
	}{
		{name: "send tx from an amino account with amino codec", memoryNodeFn: NewInMemoryTendermintNodeAmino, upgrades: &upgrades{codecUpgrade: codecUpgrade{false, 7000}}},
		{name: "send tx from a proto account with proto codec", memoryNodeFn: NewInMemoryTendermintNodeProto, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 2}}},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if tc.upgrades != nil { // NOTE: Use to perform neccesary upgrades for test
				codec.UpgradeHeight = tc.upgrades.codecUpgrade.height
				_ = memCodecMod(tc.upgrades.codecUpgrade.upgradeMod)
			}
			_, kb, cleanup := tc.memoryNodeFn(t, oneAppTwoNodeGenesis())
			time.Sleep(1 * time.Second)
			cb, err := kb.GetCoinbase()
			assert.Nil(t, err)
			kp, err := kb.Create("test")
			assert.Nil(t, err)
			_, _, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
			var transferAmount = sdk.NewInt(1000)
			var tx *sdk.TxResponse

			<-evtChan // Wait for block
			memCli, stopCli, evtChan := subscribeTo(t, tmTypes.EventTx)
			tx, err = servicers.Send(memCodec(), memCli, kb, cb.GetAddress(), kp.GetAddress(), "test", transferAmount, tc.upgrades.codecUpgrade.upgradeMod)
			assert.Nil(t, err)
			assert.NotNil(t, tx)
			assert.Equal(t, int(tx.Code), 0)

			<-evtChan // Wait for tx
			balance, err := VCA.QueryBalance(kp.GetAddress().String(), VCA.LastBlockHeight())
			assert.Nil(t, err)
			assert.True(t, balance.Equal(transferAmount))
			balance, err = VCA.QueryBalance(cb.GetAddress().String(), VCA.LastBlockHeight())
			assert.Nil(t, err)

			cleanup()
			stopCli()
		})
	}
}

func TestDuplicateTxWithRawTx(t *testing.T) {
	tt := []struct {
		name         string
		memoryNodeFn func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		*upgrades
	}{
		{name: "send duplicate tx from an amino account with amino codec", memoryNodeFn: NewInMemoryTendermintNodeAmino, upgrades: &upgrades{codecUpgrade: codecUpgrade{false, 7000}}},
		{name: "send duplicate tx from a proto account with proto codec", memoryNodeFn: NewInMemoryTendermintNodeProto, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 2}}}, // TODO:  FULL PROTO SCENARIO
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if tc.upgrades != nil { // NOTE: Use to perform neccesary upgrades for test
				codec.UpgradeHeight = tc.upgrades.codecUpgrade.height
				_ = memCodecMod(tc.upgrades.codecUpgrade.upgradeMod)
			}
			_, kb, cleanup := tc.memoryNodeFn(t, oneAppTwoNodeGenesis())
			time.Sleep(1 * time.Second)
			cb, err := kb.GetCoinbase()
			assert.Nil(t, err)
			kp, err := kb.Create("test")
			assert.Nil(t, err)
			pk, err := kb.ExportPrivateKeyObject(cb.GetAddress(), "test")
			assert.Nil(t, err)
			_, _, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
			// create the transaction
			txBz, err := types.DefaultTxEncoder(memCodec())(types.NewTestTx(sdk.Context{}.WithChainID("viper-test"),
				&servicersTypes.MsgSend{
					FromAddress: cb.GetAddress(),
					ToAddress:   kp.GetAddress(),
					Amount:      sdk.NewInt(1),
				},
				pk,
				rand2.Int64(),
				sdk.NewCoins(sdk.NewCoin(sdk.DefaultStakeDenom, sdk.NewInt(100000)))), -1)
			assert.Nil(t, err)
			// create the transaction
			_, err = types.DefaultTxEncoder(memCodec())(types.NewTestTx(sdk.Context{}.WithChainID("viper-test"),
				&servicersTypes.MsgSend{
					FromAddress: cb.GetAddress(),
					ToAddress:   kp.GetAddress(),
					Amount:      sdk.NewInt(1),
				},
				pk,
				rand2.Int64(),
				sdk.NewCoins(sdk.NewCoin(sdk.DefaultStakeDenom, sdk.NewInt(100000)))), -1)
			assert.Nil(t, err)

			<-evtChan // Wait for block
			memCli, _, evtChan := subscribeTo(t, tmTypes.EventTx)
			_, err = servicers.RawTx(memCodec(), memCli, cb.GetAddress(), txBz)
			assert.Nil(t, err)
			// next tx
			<-evtChan // Wait for tx
			_, _, evtChan = subscribeTo(t, tmTypes.EventNewBlock)
			<-evtChan // Wait for  block
			memCli, stopCli, _ := subscribeTo(t, tmTypes.EventTx)
			txResp, err := servicers.RawTx(memCodec(), memCli, cb.GetAddress(), txBz)
			if err == nil && txResp.Code == 0 {
				t.Fatal("should fail on replay attack")
			}
			cleanup()
			stopCli()
		})
	}

}

func TestChangeParamsComplexTypeTx(t *testing.T) {
	tt := []struct {
		name         string
		memoryNodeFn func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		*upgrades
	}{
		{name: "change complex type params from an amino account with amino codec", memoryNodeFn: NewInMemoryTendermintNodeAmino, upgrades: &upgrades{codecUpgrade: codecUpgrade{false, 7000}}},
		{name: "change complex type params from a proto account with proto codec", memoryNodeFn: NewInMemoryTendermintNodeProto, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 2}}}, // TODO: FIX !!
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if tc.upgrades != nil { // NOTE: Use to perform neccesary upgrades for test
				codec.UpgradeHeight = tc.upgrades.codecUpgrade.height
				_ = memCodecMod(tc.upgrades.codecUpgrade.upgradeMod)
			}
			resetTestACL()
			_, kb, cleanup := tc.memoryNodeFn(t, oneAppTwoNodeGenesis())
			time.Sleep(1 * time.Second)
			cb, err := kb.GetCoinbase()
			assert.Nil(t, err)
			kps, err := kb.List()
			assert.Nil(t, err)
			kp2 := kps[1]
			_, _, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
			<-evtChan // Wait for block
			a := testACL
			a.SetOwner("governance/acl", kp2.GetAddress())
			memCli, stopCli, evtChan := subscribeTo(t, tmTypes.EventTx)
			tx, err := governance.ChangeParamsTx(memCodec(), memCli, kb, cb.GetAddress(), "governance/acl", a, "test", 1000000, false)
			assert.Nil(t, err)
			assert.NotNil(t, tx)
			select {
			case _ = <-evtChan:
				//fmt.Println(res)
				acl, err := VCA.QueryACL(VCA.LastBlockHeight())
				assert.Nil(t, err)
				o := acl.GetOwner("governance/acl")
				assert.Equal(t, kp2.GetAddress().String(), o.String())
				cleanup()
				stopCli()
			}
		})
	}
}

func TestChangeParamsSimpleTx(t *testing.T) {

	tt := []struct {
		name         string
		memoryNodeFn func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		*upgrades
	}{
		{name: "change complex type params from an amino account with amino codec", memoryNodeFn: NewInMemoryTendermintNodeAmino, upgrades: &upgrades{codecUpgrade: codecUpgrade{false, 7000}}},
		{name: "change complex type params from a proto account with proto codec", memoryNodeFn: NewInMemoryTendermintNodeProto, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 2}}}, // TODO: FULL PROTO SCENARIO
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if tc.upgrades != nil { // NOTE: Use to perform neccesary upgrades for test
				codec.UpgradeHeight = tc.upgrades.codecUpgrade.height
				_ = memCodecMod(tc.upgrades.codecUpgrade.upgradeMod)
			}
			resetTestACL()
			_, kb, cleanup := tc.memoryNodeFn(t, oneAppTwoNodeGenesis())
			time.Sleep(1 * time.Second)
			cb, err := kb.GetCoinbase()
			assert.Nil(t, err)
			_, err = kb.List()
			assert.Nil(t, err)
			_, _, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
			<-evtChan // Wait for block
			memCli, stopCli, evtChan := subscribeTo(t, tmTypes.EventTx)
			tx, err := governance.ChangeParamsTx(memCodec(), memCli, kb, cb.GetAddress(), "provider/StabilityAdjustment", 100, "test", 1000000, false)
			assert.Nil(t, err)
			assert.NotNil(t, tx)
			select {
			case _ = <-evtChan:
				//fmt.Println(res)
				assert.Nil(t, err)
				o, _ := VCA.QueryParam(VCA.LastBlockHeight(), "provider/StabilityAdjustment")
				assert.Equal(t, "100", o.Value)
				cleanup()
				stopCli()
			}
		})
	}
}

func TestChangeParamsMaxBlocksizeBeforeActivationHeight(t *testing.T) {

	tt := []struct {
		name         string
		memoryNodeFn func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		*upgrades
	}{
		{name: "change MaxBlocksize parameter before activation height", memoryNodeFn: NewInMemoryTendermintNodeProto, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 2}}}, // TODO: FULL PROTO SCENARIO
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			codec.TestMode = -2
			if tc.upgrades != nil { // NOTE: Use to perform neccesary upgrades for test
				codec.UpgradeHeight = tc.upgrades.codecUpgrade.height
				_ = memCodecMod(tc.upgrades.codecUpgrade.upgradeMod)
			}
			resetTestACL()
			_, kb, cleanup := tc.memoryNodeFn(t, oneAppTwoNodeGenesis())
			time.Sleep(1 * time.Second)
			cb, err := kb.GetCoinbase()
			assert.Nil(t, err)
			_, err = kb.List()
			assert.Nil(t, err)
			_, _, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
			<-evtChan // Wait for block
			//Before Activation of the parameter ACL do not exist and the value and parameter should be 0 or nil
			firstquery, _ := VCA.QueryParam(VCA.LastBlockHeight(), "vipercore/BlockByteSize")
			assert.Equal(t, "", firstquery.Value)
			memCli, stopCli, evtChan := subscribeTo(t, tmTypes.EventTx)
			//Tx wont modify anything as ACL is not configured (Txresult should be governance code 5)
			tx, err := governance.ChangeParamsTx(memCodec(), memCli, kb, cb.GetAddress(), "vipercore/BlockByteSize", 9000000, "test", 10000, false)
			assert.Nil(t, err)
			assert.NotNil(t, tx)
			select {
			case _ = <-evtChan:
				//fmt.Println(res)
				assert.Nil(t, err)
				o, _ := VCA.QueryParam(VCA.LastBlockHeight(), "vipercore/BlockByteSize")
				//value should be equal to the first query of the param
				assert.Equal(t, firstquery.Value, o.Value)
				cleanup()
				stopCli()
			}
		})
	}
}

func TestChangeParamsMaxBlocksizeAfterActivationHeight(t *testing.T) {

	tt := []struct {
		name         string
		memoryNodeFn func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		*upgrades
	}{
		{name: "change MaxBlocksize parameter past activation height", memoryNodeFn: NewInMemoryTendermintNodeProto, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 2}}}, // TODO: FULL PROTO SCENARIO
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			codec.TestMode = -2
			codec.UpgradeFeatureMap[codec.BlockSizeModifyKey] = tc.upgrades.codecUpgrade.height + 1
			if tc.upgrades != nil { // NOTE: Use to perform neccesary upgrades for test
				codec.UpgradeHeight = tc.upgrades.codecUpgrade.height
				_ = memCodecMod(tc.upgrades.codecUpgrade.upgradeMod)
			}
			resetTestACL()
			_, kb, cleanup := tc.memoryNodeFn(t, oneAppTwoNodeGenesis())
			time.Sleep(1 * time.Second)
			cb, err := kb.GetCoinbase()
			assert.Nil(t, err)
			_, err = kb.List()
			assert.Nil(t, err)
			_, _, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
			<-evtChan // Wait for block
			<-evtChan // Wait for another block
			//After Activation of the parameter ACL should be created(allowing modifying the value) and parameter should have default value of 4000000
			o, _ := VCA.QueryParam(VCA.LastBlockHeight(), "vipercore/BlockByteSize")
			assert.Equal(t, "4000000", o.Value)
			memCli, stopCli, evtChan := subscribeTo(t, tmTypes.EventTx)
			tx, err := governance.ChangeParamsTx(memCodec(), memCli, kb, cb.GetAddress(), "vipercore/BlockByteSize", 9000000, "test", 10000, false)
			assert.Nil(t, err)
			assert.NotNil(t, tx)
			select {
			case _ = <-evtChan:
				//fmt.Println(res)
				assert.Nil(t, err)
				o, _ := VCA.QueryParam(VCA.LastBlockHeight(), "vipercore/BlockByteSize")
				assert.Equal(t, "9000000", o.Value)
				cleanup()
				stopCli()
			}
		})
	}
}

func TestUpgrade(t *testing.T) {
	tt := []struct {
		name         string
		memoryNodeFn func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		*upgrades
	}{
		{name: "change complex type params from an amino account with amino codec", memoryNodeFn: NewInMemoryTendermintNodeAmino, upgrades: &upgrades{codecUpgrade: codecUpgrade{false, 7000}}},
		{name: "change complex type params from a proto account with proto codec", memoryNodeFn: NewInMemoryTendermintNodeProto, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 2}}},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if tc.upgrades != nil { // NOTE: Use to perform neccesary upgrades for test
				codec.UpgradeHeight = tc.upgrades.codecUpgrade.height
				_ = memCodecMod(tc.upgrades.codecUpgrade.upgradeMod)
			}
			_, kb, cleanup := tc.memoryNodeFn(t, oneAppTwoNodeGenesis())
			time.Sleep(1 * time.Second)
			cb, err := kb.GetCoinbase()
			assert.Nil(t, err)
			_, _, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
			var tx *sdk.TxResponse
			<-evtChan // Wait for block
			memCli, stopCli, evtChan := subscribeTo(t, tmTypes.EventTx)
			tx, err = governance.UpgradeTx(memCodec(), memCli, kb, cb.GetAddress(), govTypes.Upgrade{
				Height:  1000,
				Version: "2.0.0",
			}, "test", 1000000, tc.codecUpgrade.upgradeMod)
			assert.Nil(t, err)
			assert.NotNil(t, tx)

			<-evtChan // Wait for tx
			u, err := VCA.QueryUpgrade(VCA.LastBlockHeight())
			assert.Nil(t, err)
			assert.True(t, u.UpgradeVersion() == "2.0.0")

			cleanup()
			stopCli()
		})
	}
}

func TestDAOTransfer(t *testing.T) {
	tt := []struct {
		name         string
		memoryNodeFn func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		*upgrades
	}{
		{name: "change complex type params from an amino account with amino codec", memoryNodeFn: NewInMemoryTendermintNodeAmino, upgrades: &upgrades{codecUpgrade: codecUpgrade{false, 7000}}},
		{name: "change complex type params from a proto account with proto codec", memoryNodeFn: NewInMemoryTendermintNodeProto, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 2}}},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if tc.upgrades != nil { // NOTE: Use to perform neccesary upgrades for test
				codec.UpgradeHeight = tc.upgrades.codecUpgrade.height
				_ = memCodecMod(tc.upgrades.codecUpgrade.upgradeMod)
			}
			_, kb, cleanup := tc.memoryNodeFn(t, oneAppTwoNodeGenesis())
			time.Sleep(1 * time.Second)
			cb, err := kb.GetCoinbase()
			assert.Nil(t, err)
			_, _, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
			var tx *sdk.TxResponse
			<-evtChan // Wait for block
			memCli, stopCli, evtChan := subscribeTo(t, tmTypes.EventTx)
			tx, err = governance.DAOTransferTx(memCodec(), memCli, kb, cb.GetAddress(), nil, sdk.OneInt(), govTypes.DAOBurn.String(), "test", 1000000, tc.codecUpgrade.upgradeMod)
			assert.Nil(t, err)
			assert.NotNil(t, tx)

			<-evtChan // Wait for tx
			balance, err := VCA.QueryDaoBalance(VCA.LastBlockHeight())
			assert.Nil(t, err)
			assert.True(t, balance.Equal(sdk.NewInt(999)))

			cleanup()
			stopCli()
		})
	}
}

func TestClaimAminoTx(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}
	tt := []struct {
		name         string
		memoryNodeFn func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		*upgrades
	}{
		{name: "claim tx from amino with amino codec ", memoryNodeFn: NewInMemoryTendermintNodeAmino, upgrades: &upgrades{codecUpgrade: codecUpgrade{false, 7000}}},
		//{name: "claim tx from a proto with proto codec", memoryNodeFn: NewInMemoryTendermintNodeProto, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 4}}},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if tc.upgrades != nil { // NOTE: Use to perform neccesary upgrades for test
				codec.UpgradeHeight = tc.upgrades.codecUpgrade.height
				_ = memCodecMod(tc.upgrades.codecUpgrade.upgradeMod)
			}
			genBz, _, validators, app := fiveValidatorsOneAppGenesis()
			kb := getInMemoryKeybase()
			_, _, cleanup := tc.memoryNodeFn(t, genBz)
			time.Sleep(1 * time.Second)
			for i := 0; i < 5; i++ {
				appPrivateKey, err := kb.ExportPrivateKeyObject(app.Address, "test")
				assert.Nil(t, err)
				// setup AAT
				aat := viperTypes.AAT{
					Version:           "0.0.1",
					ProviderPublicKey: appPrivateKey.PublicKey().RawString(),
					ClientPublicKey:   appPrivateKey.PublicKey().RawString(),
					ProviderSignature: "",
				}
				sig, err := appPrivateKey.Sign(aat.Hash())
				if err != nil {
					panic(err)
				}
				aat.ProviderSignature = hex.EncodeToString(sig)
				proof := viperTypes.RelayProof{
					Entropy:            int64(rand.Int()),
					RequestHash:        hex.EncodeToString(viperTypes.Hash([]byte("fake"))),
					SessionBlockHeight: 1,
					ServicerPubKey:     validators[0].PublicKey.RawString(),
					Blockchain:         sdk.PlaceholderHash,
					Token:              aat,
					Signature:          "",
				}
				sig, err = appPrivateKey.Sign(proof.Hash())
				if err != nil {
					t.Fatal(err)
				}
				proof.Signature = hex.EncodeToString(sig)
				viperTypes.SetProof(viperTypes.SessionHeader{
					ProviderPubKey:     appPrivateKey.PublicKey().RawString(),
					Chain:              sdk.PlaceholderHash,
					SessionBlockHeight: 1,
				}, viperTypes.RelayEvidence, proof, sdk.NewInt(1000000), viperTypes.GlobalEvidenceCache)
				assert.Nil(t, err)
			}
			_, _, evtChan := subscribeTo(t, tmTypes.EventTx)
			res := <-evtChan
			if res.Events["message.action"][0] != viperTypes.EventTypeClaim {
				t.Fatal("claim message was not received first")
			}
			_, stopCli, evtChan := subscribeTo(t, tmTypes.EventTx)
			res = <-evtChan
			if res.Events["message.action"][0] != viperTypes.EventTypeProof {
				t.Fatal("proof message was not received afterward")
			}
			cleanup()
			stopCli()
		})
	}
}

func TestClaimProtoTx(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}
	tt := []struct {
		name         string
		memoryNodeFn func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		*upgrades
	}{
		//{name: "claim tx from amino with amino codec ", memoryNodeFn: NewInMemoryTendermintNodeAmino, upgrades: &upgrades{codecUpgrade: codecUpgrade{false, 7000}}},
		{name: "claim tx from a proto with proto codec", memoryNodeFn: NewInMemoryTendermintNodeProto, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 5}}},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if tc.upgrades != nil { // NOTE: Use to perform neccesary upgrades for test
				codec.UpgradeHeight = tc.upgrades.codecUpgrade.height
				_ = memCodecMod(tc.upgrades.codecUpgrade.upgradeMod)
			}
			genBz, _, validators, app := fiveValidatorsOneAppGenesis()
			kb := getInMemoryKeybase()
			_, _, cleanup := tc.memoryNodeFn(t, genBz)
			time.Sleep(1 * time.Second)
			for i := 0; i < 5; i++ {
				appPrivateKey, err := kb.ExportPrivateKeyObject(app.Address, "test")
				assert.Nil(t, err)
				// setup AAT
				aat := viperTypes.AAT{
					Version:           "0.0.1",
					ProviderPublicKey: appPrivateKey.PublicKey().RawString(),
					ClientPublicKey:   appPrivateKey.PublicKey().RawString(),
					ProviderSignature: "",
				}
				sig, err := appPrivateKey.Sign(aat.Hash())
				if err != nil {
					panic(err)
				}
				aat.ProviderSignature = hex.EncodeToString(sig)
				proof := viperTypes.RelayProof{
					Entropy:            int64(rand.Int()),
					RequestHash:        hex.EncodeToString(viperTypes.Hash([]byte("fake"))),
					SessionBlockHeight: 1,
					ServicerPubKey:     validators[0].PublicKey.RawString(),
					Blockchain:         sdk.PlaceholderHash,
					Token:              aat,
					Signature:          "",
				}
				sig, err = appPrivateKey.Sign(proof.Hash())
				if err != nil {
					t.Fatal(err)
				}
				proof.Signature = hex.EncodeToString(sig)
				viperTypes.SetProof(viperTypes.SessionHeader{
					ProviderPubKey:     appPrivateKey.PublicKey().RawString(),
					Chain:              sdk.PlaceholderHash,
					SessionBlockHeight: 1,
				}, viperTypes.RelayEvidence, proof, sdk.NewInt(1000000), viperTypes.GlobalEvidenceCache)
				assert.Nil(t, err)
			}
			_, _, evtChan := subscribeTo(t, tmTypes.EventTx)
			res := <-evtChan
			fmt.Println(res)
			if res.Events["message.action"][0] != viperTypes.EventTypeClaim {
				t.Fatal("claim message was not received first")
			}
			_, stopCli, evtChan := subscribeTo(t, tmTypes.EventTx)
			res = <-evtChan
			if res.Events["message.action"][0] != viperTypes.EventTypeProof {
				t.Fatal("proof message was not received afterward")
			}
			cleanup()
			stopCli()
		})
	}
}

func TestAminoClaimTxChallenge(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}
	tt := []struct {
		name         string
		memoryNodeFn func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		*upgrades
	}{
		{name: "challenge a claim tx from amino with amino codec ", memoryNodeFn: NewInMemoryTendermintNodeAmino, upgrades: &upgrades{codecUpgrade: codecUpgrade{false, 7000}}},
		//{name: "challenge a claim tx from a proto with proto codec", memoryNodeFn: NewInMemoryTendermintNodeProto, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 2}}}, // TODO: FULL PROT SCENARIO
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			genBz, keys, _, _ := fiveValidatorsOneAppGenesis()
			challenges := NewValidChallengeProof(t, keys, 5)
			_, _, cleanup := tc.memoryNodeFn(t, genBz)
			for _, c := range challenges {
				c.Store(sdk.NewInt(1000000), viperTypes.GlobalEvidenceCache)
			}
			_, _, evtChan := subscribeTo(t, tmTypes.EventTx)
			res := <-evtChan // Wait for tx
			if res.Events["message.action"][0] != viperTypes.EventTypeClaim {
				t.Fatal("claim message was not received first")
			}

			_, stopCli, evtChan := subscribeTo(t, tmTypes.EventTx)
			res = <-evtChan // Wait for tx
			if res.Events["message.action"][0] != viperTypes.EventTypeProof {
				t.Fatal("proof message was not received afterward")
			}
			cleanup()
			stopCli()
		})
	}
}

func TestProtoClaimTxChallenge(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}
	tt := []struct {
		name         string
		memoryNodeFn func(t *testing.T, genesisState []byte) (tendermint *node.Node, keybase keys.Keybase, cleanup func())
		*upgrades
	}{
		//{name: "challenge a claim tx from amino with amino codec ", memoryNodeFn: NewInMemoryTendermintNodeAmino, upgrades: &upgrades{codecUpgrade: codecUpgrade{false, 7000}}},
		{name: "challenge a claim tx from a proto with proto codec", memoryNodeFn: NewInMemoryTendermintNodeProto, upgrades: &upgrades{codecUpgrade: codecUpgrade{true, 2}}}, // TODO: FULL PROT SCENARIO
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			genBz, keys, _, _ := fiveValidatorsOneAppGenesis()
			challenges := NewValidChallengeProof(t, keys, 5)
			_, _, cleanup := tc.memoryNodeFn(t, genBz)
			for _, c := range challenges {
				c.Store(sdk.NewInt(1000000), viperTypes.GlobalEvidenceCache)
			}
			_, _, evtChan := subscribeTo(t, tmTypes.EventTx)
			res := <-evtChan // Wait for tx
			if res.Events["message.action"][0] != viperTypes.EventTypeClaim {
				t.Fatal("claim message was not received first")
			}

			_, stopCli, evtChan := subscribeTo(t, tmTypes.EventTx)
			res = <-evtChan // Wait for tx
			if res.Events["message.action"][0] != viperTypes.EventTypeProof {
				t.Fatal("proof message was not received afterward")
			}
			cleanup()
			stopCli()
		})
	}
}

func NewValidChallengeProof(t *testing.T, privateKeys []crypto.PrivateKey, numOfChallenges int) (challenge []viperTypes.ChallengeProofInvalidData) {

	providerPrivateKey := privateKeys[1]
	servicerPrivKey1 := privateKeys[4]
	servicerPrivKey2 := privateKeys[2]
	servicerPrivKey3 := privateKeys[3]
	clientPrivateKey := servicerPrivKey3
	providerPubKey := providerPrivateKey.PublicKey().RawString()
	servicerPubKey := servicerPrivKey1.PublicKey().RawString()
	servicerPubKey2 := servicerPrivKey2.PublicKey().RawString()
	servicerPubKey3 := servicerPrivKey3.PublicKey().RawString()
	reporterPrivKey := privateKeys[0]
	reporterPubKey := reporterPrivKey.PublicKey()
	reporterAddr := reporterPubKey.Address()
	clientPubKey := clientPrivateKey.PublicKey().RawString()
	var proofs []viperTypes.ChallengeProofInvalidData
	for i := 0; i < numOfChallenges; i++ {
		validProof := viperTypes.RelayProof{
			Entropy:            int64(rand.Intn(500000)),
			SessionBlockHeight: 1,
			ServicerPubKey:     servicerPubKey,
			RequestHash:        clientPubKey, // fake
			Blockchain:         sdk.PlaceholderHash,
			Token: viperTypes.AAT{
				Version:           "0.0.1",
				ProviderPublicKey: providerPubKey,
				ClientPublicKey:   clientPubKey,
				ProviderSignature: "",
			},
			Signature: "",
		}
		appSignature, er := providerPrivateKey.Sign(validProof.Token.Hash())
		if er != nil {
			t.Fatalf(er.Error())
		}
		validProof.Token.ProviderSignature = hex.EncodeToString(appSignature)
		clientSignature, er := clientPrivateKey.Sign(validProof.Hash())
		if er != nil {
			t.Fatalf(er.Error())
		}
		validProof.Signature = hex.EncodeToString(clientSignature)
		// valid proof 2
		validProof2 := viperTypes.RelayProof{
			Entropy:            0,
			SessionBlockHeight: 1,
			ServicerPubKey:     servicerPubKey2,
			RequestHash:        clientPubKey, // fake
			Blockchain:         sdk.PlaceholderHash,
			Token: viperTypes.AAT{
				Version:           "0.0.1",
				ProviderPublicKey: providerPubKey,
				ClientPublicKey:   clientPubKey,
				ProviderSignature: "",
			},
			Signature: "",
		}
		appSignature, er = providerPrivateKey.Sign(validProof2.Token.Hash())
		if er != nil {
			t.Fatalf(er.Error())
		}
		validProof2.Token.ProviderSignature = hex.EncodeToString(appSignature)
		clientSignature, er = clientPrivateKey.Sign(validProof2.Hash())
		if er != nil {
			t.Fatalf(er.Error())
		}
		validProof2.Signature = hex.EncodeToString(clientSignature)
		// valid proof 3
		validProof3 := viperTypes.RelayProof{
			Entropy:            0,
			SessionBlockHeight: 1,
			ServicerPubKey:     servicerPubKey3,
			RequestHash:        clientPubKey, // fake
			Blockchain:         sdk.PlaceholderHash,
			Token: viperTypes.AAT{
				Version:           "0.0.1",
				ProviderPublicKey: providerPubKey,
				ClientPublicKey:   clientPubKey,
				ProviderSignature: "",
			},
			Signature: "",
		}
		appSignature, er = providerPrivateKey.Sign(validProof3.Token.Hash())
		if er != nil {
			t.Fatalf(er.Error())
		}
		validProof3.Token.ProviderSignature = hex.EncodeToString(appSignature)
		clientSignature, er = clientPrivateKey.Sign(validProof3.Hash())
		if er != nil {
			t.Fatalf(er.Error())
		}
		validProof3.Signature = hex.EncodeToString(clientSignature)
		// create responses
		majorityResponsePayload := `{"id":67,"jsonrpc":"2.0","result":"Mist/v0.9.3/darwin/go1.4.1"}`
		minorityResponsePayload := `{"id":67,"jsonrpc":"2.0","result":"Mist/v0.9.3/darwin/go1.4.2"}`
		// majority response 1
		majResp1 := viperTypes.RelayResponse{
			Signature: "",
			Response:  majorityResponsePayload,
			Proof:     validProof,
		}
		sig, er := servicerPrivKey1.Sign(majResp1.Hash())
		if er != nil {
			t.Fatalf(er.Error())
		}
		majResp1.Signature = hex.EncodeToString(sig)
		// majority response 2
		majResp2 := viperTypes.RelayResponse{
			Signature: "",
			Response:  majorityResponsePayload,
			Proof:     validProof2,
		}
		sig, er = servicerPrivKey2.Sign(majResp2.Hash())
		if er != nil {
			t.Fatalf(er.Error())
		}
		majResp2.Signature = hex.EncodeToString(sig)
		// minority response
		minResp := viperTypes.RelayResponse{
			Signature: "",
			Response:  minorityResponsePayload,
			Proof:     validProof3,
		}
		sig, er = servicerPrivKey3.Sign(minResp.Hash())
		if er != nil {
			t.Fatalf(er.Error())
		}
		minResp.Signature = hex.EncodeToString(sig)
		// create valid challenge proof
		proofs = append(proofs, viperTypes.ChallengeProofInvalidData{
			MajorityResponses: []viperTypes.RelayResponse{
				majResp1,
				majResp2,
			},
			MinorityResponse: minResp,
			ReporterAddress:  sdk.Address(reporterAddr),
		})
	}
	return proofs
}
