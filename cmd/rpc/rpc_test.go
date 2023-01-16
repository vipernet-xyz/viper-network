package rpc

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/vipernet-xyz/viper-network/codec"

	"github.com/vipernet-xyz/viper-network/app"
	"github.com/vipernet-xyz/viper-network/crypto"

	rand2 "github.com/tendermint/tendermint/libs/rand"

	types3 "github.com/vipernet-xyz/viper-network/x/platforms/types"

	"github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/authentication"
	authTypes "github.com/vipernet-xyz/viper-network/x/authentication/types"
	"github.com/vipernet-xyz/viper-network/x/providers"
	types2 "github.com/vipernet-xyz/viper-network/x/providers/types"
	viperTypes "github.com/vipernet-xyz/viper-network/x/vipernet/types"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	core_types "github.com/tendermint/tendermint/rpc/core/types"
	tmTypes "github.com/tendermint/tendermint/types"
	"gopkg.in/h2non/gock.v1"
)

const (
	PlaceholderHash       = "0001"
	PlaceholderURL        = "https://foo.bar:8080"
	PlaceholderServiceURL = PlaceholderURL
)

func TestRPC_QueryBlock(t *testing.T) {
	codec.UpgradeHeight = 7000
	_, _, cleanup := NewInMemoryTendermintNode(t, oneValTwoNodeGenesisState())
	_, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)

	var params = HeightParams{
		Height: 1,
	}

	<-evtChan // Wait for block
	q := newQueryRequest("block", newBody(params))
	rec := httptest.NewRecorder()
	Block(rec, q, httprouter.Params{})
	resp := getJSONResponse(rec)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp)
	var blk core_types.ResultBlock
	err := memCodec().UnmarshalJSON([]byte(resp), &blk)
	assert.Nil(t, err)
	assert.NotEmpty(t, blk.Block.Height)

	<-evtChan // Wait for block
	q = newQueryRequest("block", newBody(params))
	rec = httptest.NewRecorder()
	Block(rec, q, httprouter.Params{})
	resp = getJSONResponse(rec)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp)
	var blk2 core_types.ResultBlock
	err = memCodec().UnmarshalJSON([]byte(resp), &blk2)
	assert.Nil(t, err)
	assert.NotEmpty(t, blk2.Block.Height)

	cleanup()
	stopCli()
}

func TestRPC_QueryTX(t *testing.T) {
	codec.UpgradeHeight = 7000
	var tx *types.TxResponse
	_, _, cleanup := NewInMemoryTendermintNode(t, oneValTwoNodeGenesisState())
	_, _, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
	<-evtChan // Wait for block
	_, stopCli, evtChan := subscribeTo(t, tmTypes.EventTx)
	kb := getInMemoryKeybase()
	cb, err := kb.GetCoinbase()
	assert.Nil(t, err)
	tx, err = providers.Send(memCodec(), memCLI, kb, cb.GetAddress(), cb.GetAddress(), "test", types.NewInt(100), true)
	assert.Nil(t, err)

	<-evtChan // Wait for tx
	var params = HashAndProveParams{
		Hash: tx.TxHash,
	}
	q := newQueryRequest("tx", newBody(params))
	rec := httptest.NewRecorder()
	Tx(rec, q, httprouter.Params{})
	resp := getJSONResponse(rec)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp)
	var resTX core_types.ResultTx
	err = json.Unmarshal([]byte(resp), &resTX)
	assert.Nil(t, err)
	assert.NotEmpty(t, resTX.Height)

	memCLI, _, evtChan = subscribeTo(t, tmTypes.EventNewBlock)
	<-evtChan // Wait for block
	q = newQueryRequest("tx", newBody(params))
	rec = httptest.NewRecorder()
	Tx(rec, q, httprouter.Params{})
	resp = getJSONResponse(rec)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp)
	var resTX2 core_types.ResultTx
	err = json.Unmarshal([]byte(resp), &resTX2)
	assert.Nil(t, err)
	assert.NotEmpty(t, resTX2.Height)

	cleanup()
	stopCli()
}

func TestRPC_QueryAccountTXs(t *testing.T) {
	codec.UpgradeHeight = 7000
	var tx *types.TxResponse
	_, _, cleanup := NewInMemoryTendermintNode(t, oneValTwoNodeGenesisState())
	memCLI, _, evtChan := subscribeTo(t, tmTypes.EventNewBlock)

	<-evtChan // Wait for block
	var err error
	_, stopCli, evtChan := subscribeTo(t, tmTypes.EventTx)
	kb := getInMemoryKeybase()
	cb, err := kb.GetCoinbase()
	assert.Nil(t, err)
	tx, err = providers.Send(memCodec(), memCLI, kb, cb.GetAddress(), cb.GetAddress(), "test", types.NewInt(100), true)
	assert.Nil(t, err)
	assert.NotNil(t, tx)

	<-evtChan // Wait for tx
	kb = getInMemoryKeybase()
	cb, err = kb.GetCoinbase()
	assert.Nil(t, err)
	var params = PaginateAddrParams{
		Address: cb.GetAddress().String(),
	}
	q := newQueryRequest("accounttxs", newBody(params))
	rec := httptest.NewRecorder()
	AccountTxs(rec, q, httprouter.Params{})
	resp := getJSONResponse(rec)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp)
	var resTXs core_types.ResultTxSearch
	unmarshalErr := json.Unmarshal([]byte(resp), &resTXs)
	assert.Nil(t, unmarshalErr)
	assert.NotEmpty(t, resTXs.Txs)

	_, _, evtChan = subscribeTo(t, tmTypes.EventNewBlock)
	<-evtChan // Wait for block
	q = newQueryRequest("accounttxs", newBody(params))
	rec = httptest.NewRecorder()
	AccountTxs(rec, q, httprouter.Params{})
	resp = getJSONResponse(rec)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp)
	var resTXs2 core_types.ResultTxSearch
	unmarshalErr = json.Unmarshal([]byte(resp), &resTXs2)
	assert.Nil(t, unmarshalErr)
	assert.NotEmpty(t, resTXs2.Txs)

	cleanup()
	stopCli()
}

func TestRPC_QueryBlockTXs(t *testing.T) {
	codec.UpgradeHeight = 7000
	var tx *types.TxResponse
	_, _, cleanup := NewInMemoryTendermintNode(t, oneValTwoNodeGenesisState())
	memCLI, _, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
	<-evtChan //Wait for block
	var err error
	_, stopCli, evtChan := subscribeTo(t, tmTypes.EventTx)
	kb := getInMemoryKeybase()
	cb, err := kb.GetCoinbase()
	assert.Nil(t, err)
	tx, err = providers.Send(memCodec(), memCLI, kb, cb.GetAddress(), cb.GetAddress(), "test", types.NewInt(100), true)
	assert.Nil(t, err)

	<-evtChan // Wait for tx
	// Step 1: Get the transaction by it's hash
	var params = HashAndProveParams{
		Hash: tx.TxHash,
	}
	q := newQueryRequest("tx", newBody(params))
	rec := httptest.NewRecorder()
	Tx(rec, q, httprouter.Params{})
	resp := getJSONResponse(rec)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp)
	var resTX core_types.ResultTx
	err = json.Unmarshal([]byte(resp), &resTX)
	assert.Nil(t, err)
	assert.NotEmpty(t, resTX.Height)

	// Step 2: Get the transaction by it's height
	var heightParams = PaginatedHeightParams{
		Height: resTX.Height,
	}
	heightQ := newQueryRequest("blocktxs", newBody(heightParams))
	heightRec := httptest.NewRecorder()
	BlockTxs(heightRec, heightQ, httprouter.Params{})
	heightResp := getJSONResponse(heightRec)
	assert.NotNil(t, heightResp)
	assert.NotEmpty(t, heightResp)
	var resTXs core_types.ResultTxSearch
	unmarshalErr := json.Unmarshal([]byte(heightResp), &resTXs)
	assert.Nil(t, unmarshalErr)
	assert.NotEmpty(t, resTXs.Txs)

	_, _, evtChan = subscribeTo(t, tmTypes.EventNewBlock)
	<-evtChan // Wait for block
	heightQ = newQueryRequest("blocktxs", newBody(heightParams))
	heightRec = httptest.NewRecorder()
	BlockTxs(heightRec, heightQ, httprouter.Params{})
	heightResp = getJSONResponse(heightRec)
	assert.NotNil(t, heightResp)
	assert.NotEmpty(t, heightResp)
	var resTXs2 core_types.ResultTxSearch
	unmarshalErr = json.Unmarshal([]byte(heightResp), &resTXs2)
	assert.Nil(t, unmarshalErr)
	assert.NotEmpty(t, resTXs2.Txs)

	cleanup()
	stopCli()
}

func TestRPC_QueryBalance(t *testing.T) {
	codec.UpgradeHeight = 7000
	_, _, cleanup := NewInMemoryTendermintNode(t, oneValTwoNodeGenesisState())
	_, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)

	<-evtChan // Wait for block
	kb := getInMemoryKeybase()
	cb, err := kb.GetCoinbase()
	assert.Nil(t, err)
	var params = HeightAndAddrParams{
		Height:  0,
		Address: cb.GetAddress().String(),
	}
	q := newQueryRequest("balance", newBody(params))
	rec := httptest.NewRecorder()
	Balance(rec, q, httprouter.Params{})
	resp := getJSONResponse(rec)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp)

	var b queryBalanceResponse
	err = json.Unmarshal([]byte(resp), &b)
	assert.Nil(t, err)
	assert.NotZero(t, b.Balance)
	<-evtChan // Wait for blockk
	params = HeightAndAddrParams{
		Height:  2,
		Address: cb.GetAddress().String(),
	}
	q = newQueryRequest("balance", newBody(params))
	rec = httptest.NewRecorder()
	Balance(rec, q, httprouter.Params{})
	resp = getJSONResponse(rec)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp)

	var b2 queryBalanceResponse
	err = json.Unmarshal([]byte(resp), &b2)
	assert.Nil(t, err)
	assert.NotZero(t, b2.Balance)

	cleanup()
	stopCli()
}

func TestRPC_QueryAccount(t *testing.T) {
	codec.UpgradeHeight = 7000
	_, _, cleanup := NewInMemoryTendermintNode(t, oneValTwoNodeGenesisState())
	_, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
	<-evtChan
	kb := getInMemoryKeybase()
	cb, err := kb.GetCoinbase()
	assert.Nil(t, err)
	var params = HeightAndAddrParams{
		Height:  0,
		Address: cb.GetAddress().String(),
	}
	q := newQueryRequest("account", newBody(params))
	rec := httptest.NewRecorder()
	Account(rec, q, httprouter.Params{})
	resp := getJSONResponse(rec)
	assert.Regexp(t, "uvipr", string(resp))

	<-evtChan
	q = newQueryRequest("account", newBody(params))
	rec = httptest.NewRecorder()
	Account(rec, q, httprouter.Params{})
	resp = getJSONResponse(rec)
	assert.Regexp(t, "uvipr", string(resp))

	cleanup()
	stopCli()
}

func TestRPC_QueryAccounts(t *testing.T) {
	codec.UpgradeHeight = 7000
	_, _, cleanup := NewInMemoryTendermintNode(t, oneValTwoNodeGenesisState())
	_, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
	<-evtChan
	kb := getInMemoryKeybase()
	cb, err := kb.GetCoinbase()
	assert.Nil(t, err)
	var params = PaginatedHeightParams{
		Height: 0,
	}
	address := cb.GetAddress().String()
	q := newQueryRequest("accounts", newBody(params))
	rec := httptest.NewRecorder()
	Accounts(rec, q, httprouter.Params{})
	body := rec.Body.String()
	assert.True(t, strings.Contains(body, address))

	cleanup()
	stopCli()
}

func TestRPC_QueryProviders(t *testing.T) {
	codec.UpgradeHeight = 7000
	_, _, cleanup := NewInMemoryTendermintNode(t, oneValTwoNodeGenesisState())
	_, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)

	<-evtChan // Wait for block
	kb := getInMemoryKeybase()
	cb, err := kb.GetCoinbase()
	assert.Nil(t, err)
	var params = HeightAndValidatorOptsParams{
		Height: 0,
		Opts: types2.QueryValidatorsParams{
			StakingStatus: types.Staked,
			Page:          1,
			Limit:         1,
		},
	}
	q := newQueryRequest("providers", newBody(params))
	rec := httptest.NewRecorder()
	Providers(rec, q, httprouter.Params{})
	body := rec.Body.String()
	address := cb.GetAddress().String()
	assert.True(t, strings.Contains(body, address))

	<-evtChan // Wait for block
	q = newQueryRequest("providers", newBody(params))
	rec = httptest.NewRecorder()
	Providers(rec, q, httprouter.Params{})
	body = rec.Body.String()
	address = cb.GetAddress().String()
	assert.True(t, strings.Contains(body, address))

	cleanup()
	stopCli()
}

func TestRPC_QueryNode(t *testing.T) {
	codec.UpgradeHeight = 7000
	_, _, cleanup := NewInMemoryTendermintNode(t, oneValTwoNodeGenesisState())
	_, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)

	<-evtChan // Wait for block
	kb := getInMemoryKeybase()
	cb, err := kb.GetCoinbase()
	assert.Nil(t, err)
	var params = HeightAndAddrParams{
		Height:  0,
		Address: cb.GetAddress().String(),
	}
	q := newQueryRequest("provider", newBody(params))
	rec := httptest.NewRecorder()
	Node(rec, q, httprouter.Params{})
	resp := getJSONResponse(rec)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp)
	assert.True(t, strings.Contains(rec.Body.String(), cb.GetAddress().String()))

	<-evtChan // Wait for block
	params = HeightAndAddrParams{
		Height:  2,
		Address: cb.GetAddress().String(),
	}
	q = newQueryRequest("provider", newBody(params))
	rec = httptest.NewRecorder()
	Node(rec, q, httprouter.Params{})
	resp = getJSONResponse(rec)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp)
	assert.True(t, strings.Contains(rec.Body.String(), cb.GetAddress().String()))

	cleanup()
	stopCli()
}

func TestRPC_QueryApp(t *testing.T) {
	codec.UpgradeHeight = 7000
	gBZ, _, _, app := fiveValidatorsOneAppGenesis()
	_, _, cleanup := NewInMemoryTendermintNode(t, gBZ)
	_, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
	<-evtChan // Wait for block
	var params = HeightAndAddrParams{
		Height:  0,
		Address: app.GetAddress().String(),
	}
	q := newQueryRequest("app", newBody(params))
	rec := httptest.NewRecorder()
	App(rec, q, httprouter.Params{})
	resp := getJSONResponse(rec)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp)
	assert.True(t, strings.Contains(rec.Body.String(), app.GetAddress().String()))

	<-evtChan // Wait for block
	params = HeightAndAddrParams{
		Height:  2,
		Address: app.GetAddress().String(),
	}
	q = newQueryRequest("app", newBody(params))
	rec = httptest.NewRecorder()
	App(rec, q, httprouter.Params{})
	resp = getJSONResponse(rec)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp)
	assert.True(t, strings.Contains(rec.Body.String(), app.GetAddress().String()))

	cleanup()
	stopCli()
}

func TestRPC_QueryApps(t *testing.T) {
	codec.UpgradeHeight = 7000
	gBZ, _, _, app := fiveValidatorsOneAppGenesis()
	_, _, cleanup := NewInMemoryTendermintNode(t, gBZ)
	_, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
	<-evtChan // Wait for block
	var params = HeightAndApplicaitonOptsParams{
		Height: 0,
		Opts: types3.QueryPlatformsWithOpts{
			StakingStatus: types.Staked,
			Page:          1,
			Limit:         10000,
		},
	}
	q := newQueryRequest("platforms", newBody(params))
	rec := httptest.NewRecorder()
	Apps(rec, q, httprouter.Params{})
	body := rec.Body.String()
	address := app.GetAddress().String()
	assert.True(t, strings.Contains(body, address))

	<-evtChan // Wait for block
	params = HeightAndApplicaitonOptsParams{
		Height: 2,
		Opts: types3.QueryPlatformsWithOpts{
			StakingStatus: types.Staked,
			Page:          1,
			Limit:         10000,
		},
	}
	q = newQueryRequest("platforms", newBody(params))
	rec = httptest.NewRecorder()
	Apps(rec, q, httprouter.Params{})
	body = rec.Body.String()
	address = app.GetAddress().String()
	assert.True(t, strings.Contains(body, address))

	cleanup()
	stopCli()
}

func TestRPC_QueryNodeParams(t *testing.T) {
	codec.UpgradeHeight = 7000
	gBZ, _, _, _ := fiveValidatorsOneAppGenesis()
	_, _, cleanup := NewInMemoryTendermintNode(t, gBZ)
	_, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
	<-evtChan // Wait for block
	var params = HeightParams{
		Height: 0,
	}
	q := newQueryRequest("providerparams", newBody(params))
	rec := httptest.NewRecorder()
	NodeParams(rec, q, httprouter.Params{})
	resp := getJSONResponse(rec)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp)
	assert.True(t, strings.Contains(rec.Body.String(), "max_validators"))

	<-evtChan // Wait for block
	params = HeightParams{
		Height: 2,
	}
	q = newQueryRequest("providerparams", newBody(params))
	rec = httptest.NewRecorder()
	NodeParams(rec, q, httprouter.Params{})
	resp = getJSONResponse(rec)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp)
	assert.True(t, strings.Contains(rec.Body.String(), "max_validators"))

	cleanup()
	stopCli()
}

func TestRPC_QueryAppParams(t *testing.T) {
	codec.UpgradeHeight = 7000
	gBZ, _, _, _ := fiveValidatorsOneAppGenesis()
	_, _, cleanup := NewInMemoryTendermintNode(t, gBZ)
	_, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
	<-evtChan // Wait for block
	var params = HeightParams{
		Height: 0,
	}
	q := newQueryRequest("appparams", newBody(params))
	rec := httptest.NewRecorder()
	AppParams(rec, q, httprouter.Params{})
	resp := getJSONResponse(rec)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp)
	assert.True(t, strings.Contains(rec.Body.String(), "max_applications"))

	<-evtChan // Wait for block
	params = HeightParams{
		Height: 2,
	}
	q = newQueryRequest("appparams", newBody(params))
	rec = httptest.NewRecorder()
	AppParams(rec, q, httprouter.Params{})
	resp = getJSONResponse(rec)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp)
	assert.True(t, strings.Contains(rec.Body.String(), "max_applications"))

	cleanup()
	stopCli()
}

func TestRPC_QueryViperParams(t *testing.T) {
	codec.UpgradeHeight = 7000
	gBZ, _, _, _ := fiveValidatorsOneAppGenesis()
	_, _, cleanup := NewInMemoryTendermintNode(t, gBZ)
	_, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
	<-evtChan
	var params = HeightParams{
		Height: 0,
	}
	q := newQueryRequest("viperparams", newBody(params))
	rec := httptest.NewRecorder()
	ViperParams(rec, q, httprouter.Params{})
	resp := getJSONResponse(rec)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp)
	assert.True(t, strings.Contains(rec.Body.String(), "chains"))

	<-evtChan
	params = HeightParams{
		Height: 2,
	}
	q = newQueryRequest("viperparams", newBody(params))
	rec = httptest.NewRecorder()
	ViperParams(rec, q, httprouter.Params{})
	resp = getJSONResponse(rec)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp)
	assert.True(t, strings.Contains(rec.Body.String(), "chains"))

	cleanup()
	stopCli()
}

func TestRPC_QuerySupportedChains(t *testing.T) {
	codec.UpgradeHeight = 7000
	_, _, cleanup := NewInMemoryTendermintNode(t, oneValTwoNodeGenesisState())
	_, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
	<-evtChan // Wait for block
	var params = HeightParams{
		Height: 0,
	}
	q := newQueryRequest("supportedchains", newBody(params))
	rec := httptest.NewRecorder()
	SupportedChains(rec, q, httprouter.Params{})
	resp := getResponse(rec)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp)
	assert.True(t, strings.Contains(resp, dummyChainsHash))

	<-evtChan // Wait for block
	params = HeightParams{
		Height: 2,
	}
	q = newQueryRequest("supportedchains", newBody(params))
	rec = httptest.NewRecorder()
	SupportedChains(rec, q, httprouter.Params{})
	resp = getResponse(rec)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp)
	assert.True(t, strings.Contains(resp, dummyChainsHash))

	cleanup()
	stopCli()
}
func TestRPC_QuerySupply(t *testing.T) {
	codec.UpgradeHeight = 7000
	_, _, cleanup := NewInMemoryTendermintNode(t, oneValTwoNodeGenesisState())
	_, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
	<-evtChan // Wait for block
	var params = HeightParams{
		Height: 0,
	}
	q := newQueryRequest("supply", newBody(params))
	rec := httptest.NewRecorder()
	Supply(rec, q, httprouter.Params{})

	resp := getJSONResponse(rec)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp)

	var supply querySupplyResponse
	err := json.Unmarshal([]byte(resp), &supply)
	assert.Nil(t, err)
	assert.NotZero(t, supply.Total)

	<-evtChan // Wait for block
	params = HeightParams{
		Height: 2,
	}
	q = newQueryRequest("supply", newBody(params))
	rec = httptest.NewRecorder()
	Supply(rec, q, httprouter.Params{})

	resp = getJSONResponse(rec)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp)

	var supply2 querySupplyResponse
	err = json.Unmarshal([]byte(resp), &supply2)
	assert.Nil(t, err)
	assert.NotZero(t, supply2.Total)

	cleanup()
	stopCli()
}

func TestRPC_QueryDAOOwner(t *testing.T) {
	codec.UpgradeHeight = 7000
	_, kb, cleanup := NewInMemoryTendermintNode(t, oneValTwoNodeGenesisState())
	cb, err := kb.GetCoinbase()
	assert.Nil(t, err)
	_, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
	<-evtChan // Wait for block
	var params = HeightParams{
		Height: 0,
	}
	q := newQueryRequest("DAOOwner", newBody(params))
	rec := httptest.NewRecorder()
	DAOOwner(rec, q, httprouter.Params{})
	resp := getJSONResponse(rec)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp)
	assert.True(t, strings.Contains(string(resp), cb.GetAddress().String()))

	<-evtChan // Wait for block
	params = HeightParams{
		Height: 2,
	}
	q = newQueryRequest("DAOOwner", newBody(params))
	rec = httptest.NewRecorder()
	DAOOwner(rec, q, httprouter.Params{})
	resp = getJSONResponse(rec)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp)
	assert.True(t, strings.Contains(string(resp), cb.GetAddress().String()))

	cleanup()
	stopCli()
}

func TestRPC_QueryUpgrade(t *testing.T) {
	codec.UpgradeHeight = 7000
	_, _, cleanup := NewInMemoryTendermintNode(t, oneValTwoNodeGenesisState())
	_, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
	<-evtChan // Wait for block
	var params = HeightParams{
		Height: 0,
	}
	q := newQueryRequest("Upgrade", newBody(params))
	rec := httptest.NewRecorder()
	Upgrade(rec, q, httprouter.Params{})
	resp := getJSONResponse(rec)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp)
	assert.True(t, strings.Contains(string(resp), "2.0.0"))

	<-evtChan // Wait for block
	params = HeightParams{
		Height: 2,
	}
	q = newQueryRequest("Upgrade", newBody(params))
	rec = httptest.NewRecorder()
	Upgrade(rec, q, httprouter.Params{})
	resp = getJSONResponse(rec)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp)
	assert.True(t, strings.Contains(string(resp), "2.0.0"))

	cleanup()
	stopCli()
}

func TestRPCQueryACL(t *testing.T) {
	codec.UpgradeHeight = 7000
	_, _, cleanup := NewInMemoryTendermintNode(t, oneValTwoNodeGenesisState())
	_, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
	<-evtChan // Wait for block
	var params = HeightParams{
		Height: 0,
	}
	q := newQueryRequest("ACL", newBody(params))
	rec := httptest.NewRecorder()
	ACL(rec, q, httprouter.Params{})
	resp := getJSONResponse(rec)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp)

	<-evtChan // Wait for block
	params = HeightParams{
		Height: 2,
	}
	q = newQueryRequest("ACL", newBody(params))
	rec = httptest.NewRecorder()
	ACL(rec, q, httprouter.Params{})
	resp = getJSONResponse(rec)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp)

	cleanup()
	stopCli()
}

func TestRPCQueryAllParams(t *testing.T) {
	codec.UpgradeHeight = 7000
	_, _, cleanup := NewInMemoryTendermintNode(t, oneValTwoNodeGenesisState())
	_, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
	<-evtChan // Wait for block
	var params = HeightParams{
		Height: 0,
	}
	q := newQueryRequest("allparams", newBody(params))
	rec := httptest.NewRecorder()
	AllParams(rec, q, httprouter.Params{})
	resp := getJSONResponse(rec)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp)

	<-evtChan // Wait for block
	params = HeightParams{
		Height: 2,
	}
	q = newQueryRequest("allparams", newBody(params))
	rec = httptest.NewRecorder()
	AllParams(rec, q, httprouter.Params{})
	resp = getJSONResponse(rec)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp)

	cleanup()
	stopCli()
}

func TestRPCQueryParam(t *testing.T) {
	_, _, cleanup := NewInMemoryTendermintNode(t, oneValTwoNodeGenesisState())
	_, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
	<-evtChan // Wait for block
	var params = HeightAndKeyParams{
		Height: 0,
		Key:    "governance/upgrade",
	}
	q := newQueryRequest("param", newBody(params))
	rec := httptest.NewRecorder()
	Param(rec, q, httprouter.Params{})
	resp := getJSONResponse(rec)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp)
	unm := app.SingleParamReturn{}
	_ = json.Unmarshal(resp, &unm)
	assert.NotEmpty(t, unm.Value)

	cleanup()
	stopCli()
}

const (
	acaoHeaderKey   = "Access-Control-Allow-Origin"
	acaoHeaderValue = "*"
	aclmHeaderKey   = "Access-Control-Allow-Methods"
	aclmHeaderValue = "POST"
	acahHeaderKey   = "Access-Control-Allow-Headers"
	acahHeaderValue = "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization"
)

func validateResponseCORSHeaders(t *testing.T, headerMap http.Header) {
	assert.Contains(t, headerMap, acaoHeaderKey)
	assert.Contains(t, headerMap, aclmHeaderKey)
	assert.Contains(t, headerMap, acahHeaderKey)
	assert.Equal(t, headerMap[acaoHeaderKey], []string{acaoHeaderValue})
	assert.Equal(t, headerMap[aclmHeaderKey], []string{aclmHeaderValue})
	assert.Equal(t, headerMap[acahHeaderKey], []string{acahHeaderValue})
}

func TestRPC_ChallengeCORS(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}
	//kb := getInMemoryKeybase()
	genBZ, _, _, _ := fiveValidatorsOneAppGenesis()
	_, _, cleanup := NewInMemoryTendermintNode(t, genBZ)
	// setup the query
	_, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
	<-evtChan // Wait for block
	q := newCORSRequest("challenge", newBody(""))
	rec := httptest.NewRecorder()
	Challenge(rec, q, httprouter.Params{})
	validateResponseCORSHeaders(t, rec.Result().Header)
	cleanup()
	stopCli()
}

func TestRPC_RelayCORS(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}
	codec.UpgradeHeight = 7000
	//kb := getInMemoryKeybase()
	genBZ, _, _, _ := fiveValidatorsOneAppGenesis()
	_, _, cleanup := NewInMemoryTendermintNode(t, genBZ)
	// setup the query
	_, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
	<-evtChan // Wait for block
	q := newCORSRequest("relay", newBody(""))
	rec := httptest.NewRecorder()
	Relay(rec, q, httprouter.Params{})
	validateResponseCORSHeaders(t, rec.Result().Header)

	<-evtChan // Wait for block
	q = newCORSRequest("relay", newBody(""))
	rec = httptest.NewRecorder()
	Relay(rec, q, httprouter.Params{})
	validateResponseCORSHeaders(t, rec.Result().Header)

	cleanup()
	stopCli()
}

func TestRPC_DispatchCORS(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}
	codec.UpgradeHeight = 7000
	//kb := getInMemoryKeybase()
	genBZ, _, _, _ := fiveValidatorsOneAppGenesis()
	_, _, cleanup := NewInMemoryTendermintNode(t, genBZ)
	// setup the query
	_, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
	<-evtChan // Wait for block
	q := newCORSRequest("dispatch", newBody(""))
	rec := httptest.NewRecorder()
	Dispatch(rec, q, httprouter.Params{})
	validateResponseCORSHeaders(t, rec.Result().Header)

	<-evtChan // Wait for block
	q = newCORSRequest("dispatch", newBody(""))
	rec = httptest.NewRecorder()
	Dispatch(rec, q, httprouter.Params{})
	validateResponseCORSHeaders(t, rec.Result().Header)
	cleanup()
	stopCli()
}

func TestRPC_Relay(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}
	codec.UpgradeHeight = 7000

	kb := getInMemoryKeybase()
	genBZ, _, validators, app := fiveValidatorsOneAppGenesis()
	_, _, cleanup := NewInMemoryTendermintNode(t, genBZ)
	// setup relay endpoint
	expectedRequest := `"jsonrpc":"2.0","method":"web3_sha3","params":["0x68656c6c6f20776f726c64"],"id":64`
	expectedResponse := "0x47173285a8d7341e5e972fc677286384f802f8ef42a5ec5f03bbfa254cb01fad"
	gock.New(dummyChainsURL).
		Post("").
		BodyString(expectedRequest).
		Reply(200).
		BodyString(expectedResponse)
	appPrivateKey, err := kb.ExportPrivateKeyObject(app.Address, "test")
	assert.Nil(t, err)
	// setup AAT
	aat := viperTypes.AAT{
		Version:           "0.0.1",
		PlatformPublicKey: appPrivateKey.PublicKey().RawString(),
		ClientPublicKey:   appPrivateKey.PublicKey().RawString(),
		PlatformSignature: "",
	}
	sig, err := appPrivateKey.Sign(aat.Hash())
	if err != nil {
		panic(err)
	}
	aat.PlatformSignature = hex.EncodeToString(sig)
	payload := viperTypes.Payload{
		Data:   expectedRequest,
		Method: "POST",
	}
	// setup relay
	relay := viperTypes.Relay{
		Payload: payload,
		Meta:    viperTypes.RelayMeta{BlockHeight: 5}, // todo race condition here
		Proof: viperTypes.RelayProof{
			Entropy:            32598345349034509,
			SessionBlockHeight: 1,
			ServicerPubKey:     validators[0].PublicKey.RawString(),
			Blockchain:         dummyChainsHash,
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
	relay2 := viperTypes.Relay{
		Payload: payload,
		Meta:    viperTypes.RelayMeta{BlockHeight: 5}, // todo race condition here
		Proof: viperTypes.RelayProof{
			Entropy:            32598345349034519,
			SessionBlockHeight: 1,
			ServicerPubKey:     validators[0].PublicKey.RawString(),
			Blockchain:         dummyChainsHash,
			Token:              aat,
			Signature:          "",
		},
	}
	relay2.Proof.RequestHash = relay2.RequestHashString()
	sig2, err := appPrivateKey.Sign(relay2.Proof.Hash())
	if err != nil {
		panic(err)
	}
	relay2.Proof.Signature = hex.EncodeToString(sig2)
	// setup the query
	_, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
	<-evtChan // Wait for block
	q := newClientRequest("relay", newBody(relay))
	rec := httptest.NewRecorder()
	Relay(rec, q, httprouter.Params{})
	resp := getJSONResponse(rec)
	var response RPCRelayResponse
	err = json.Unmarshal(resp, &response)
	assert.Nil(t, err)
	assert.Equal(t, expectedResponse, response.Response)
	gock.Off()

	<-evtChan // Wait for block
	gock.New(dummyChainsURL).
		Post("").
		BodyString(expectedRequest).
		Reply(200).
		BodyString(expectedResponse)

	q2 := newClientRequest("relay", newBody(relay2))
	rec2 := httptest.NewRecorder()
	Relay(rec2, q2, httprouter.Params{})
	resp = getJSONResponse(rec2)
	var response2 RPCRelayResponse
	err = json.Unmarshal(resp, &response2)
	assert.Nil(t, err)
	assert.Equal(t, expectedResponse, response2.Response)
	gock.Off()

	cleanup()
	stopCli()
}

func TestRPC_Dispatch(t *testing.T) {
	codec.UpgradeHeight = 7000
	kb := getInMemoryKeybase()
	genBZ, _, validators, app := fiveValidatorsOneAppGenesis()
	_, _, cleanup := NewInMemoryTendermintNode(t, genBZ)
	appPrivateKey, err := kb.ExportPrivateKeyObject(app.Address, "test")
	assert.Nil(t, err)
	// Setup HandleDispatch Request
	key := viperTypes.SessionHeader{
		PlatformPubKey:     appPrivateKey.PublicKey().RawString(),
		Chain:              dummyChainsHash,
		SessionBlockHeight: 1,
	}
	// setup the query
	_, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
	<-evtChan // Wait for block
	q := newClientRequest("dispatch", newBody(key))
	rec := httptest.NewRecorder()
	Dispatch(rec, q, httprouter.Params{})
	resp := getJSONResponse(rec)
	rawResp := string(resp)
	assert.Regexp(t, key.PlatformPubKey, rawResp)
	assert.Regexp(t, key.Chain, rawResp)

	for _, validator := range validators {
		assert.Regexp(t, validator.Address.String(), rawResp)
	}

	<-evtChan // Wait for block
	q = newClientRequest("dispatch", newBody(key))
	rec = httptest.NewRecorder()
	Dispatch(rec, q, httprouter.Params{})
	resp = getJSONResponse(rec)
	rawResp = string(resp)
	assert.Regexp(t, key.PlatformPubKey, rawResp)
	assert.Regexp(t, key.Chain, rawResp)

	for _, validator := range validators {
		assert.Regexp(t, validator.Address.String(), rawResp)
	}
	cleanup()
	stopCli()

}

func TestRPC_RawTX(t *testing.T) {
	codec.UpgradeHeight = 7000
	_, kb, cleanup := NewInMemoryTendermintNode(t, oneValTwoNodeGenesisState())
	cb, err := kb.GetCoinbase()
	assert.Nil(t, err)
	kp, err := kb.Create("test")
	assert.Nil(t, err)
	pk, err := kb.ExportPrivateKeyObject(cb.GetAddress(), "test")
	assert.Nil(t, err)
	_, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
	// create the transaction
	txBz, err := authentication.DefaultTxEncoder(memCodec())(authTypes.NewTestTx(types.Context{}.WithChainID("viper-test"),
		&types2.MsgSend{
			FromAddress: cb.GetAddress(),
			ToAddress:   kp.GetAddress(),
			Amount:      types.NewInt(1),
		},
		pk,
		rand2.Int64(),
		types.NewCoins(types.NewCoin(types.DefaultStakeDenom, types.NewInt(100000)))), 0)
	assert.Nil(t, err)

	_ = memCodecMod(true)
	txBz2, err := authentication.DefaultTxEncoder(memCodec())(authTypes.NewTestTx(types.Context{}.WithChainID("viper-test"),
		&types2.MsgSend{
			FromAddress: cb.GetAddress(),
			ToAddress:   kp.GetAddress(),
			Amount:      types.NewInt(2),
		},
		pk,
		rand2.Int64(),
		types.NewCoins(types.NewCoin(types.DefaultStakeDenom, types.NewInt(100000)))), 0)
	assert.Nil(t, err)
	<-evtChan // Wait for block
	params := SendRawTxParams{
		Addr:        cb.GetAddress().String(),
		RawHexBytes: hex.EncodeToString(txBz),
	}
	q := newClientRequest("rawtx", newBody(params))
	rec := httptest.NewRecorder()
	SendRawTx(rec, q, httprouter.Params{})
	resp := getResponse(rec)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	var response types.TxResponse
	err = memCodec().UnmarshalJSON([]byte(resp), &response)
	assert.Nil(t, err)
	assert.Equal(t, uint32(0), response.Code)

	<-evtChan // Wait for block
	params = SendRawTxParams{
		Addr:        cb.GetAddress().String(),
		RawHexBytes: hex.EncodeToString(txBz2),
	}
	q2 := newClientRequest("rawtx", newBody(params))
	rec2 := httptest.NewRecorder()
	SendRawTx(rec2, q2, httprouter.Params{})
	resp2 := getResponse(rec2)
	assert.Nil(t, err)
	assert.NotNil(t, resp2)
	var response2 types.TxResponse
	err = memCodec().UnmarshalJSON([]byte(resp2), &response2)
	assert.Nil(t, err)
	assert.Nil(t, response2.Logs)

	cleanup()
	stopCli()
}

func TestRPC_QueryNodeClaims(t *testing.T) {
	codec.UpgradeHeight = 7000
	_, _, cleanup := NewInMemoryTendermintNode(t, oneValTwoNodeGenesisState())
	_, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
	<-evtChan
	kb := getInMemoryKeybase()
	cb, err := kb.GetCoinbase()
	assert.Nil(t, err)
	var params = PaginatedHeightAndAddrParams{
		Height: 0,
		Addr:   cb.GetAddress().String(),
	}
	q := newQueryRequest("providerclaims", newBody(params))
	rec := httptest.NewRecorder()
	NodeClaims(rec, q, httprouter.Params{})
	getJSONResponse(rec)

	<-evtChan
	params = PaginatedHeightAndAddrParams{
		Height: 2,
		Addr:   cb.GetAddress().String(),
	}
	q = newQueryRequest("providerclaims", newBody(params))
	rec = httptest.NewRecorder()
	NodeClaims(rec, q, httprouter.Params{})
	getJSONResponse(rec)

	cleanup()
	stopCli()
}

func TestRPC_QueryNodeClaim(t *testing.T) {
	codec.UpgradeHeight = 7000
	_, _, cleanup := NewInMemoryTendermintNode(t, oneValTwoNodeGenesisState())
	_, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
	<-evtChan
	kb := getInMemoryKeybase()
	cb, err := kb.GetCoinbase()
	assert.Nil(t, err)
	var params = QueryNodeReceiptParam{
		Address:      cb.GetAddress().String(),
		Blockchain:   "0001",
		AppPubKey:    cb.PublicKey.RawString(),
		SBlockHeight: 1,
		Height:       0,
		ReceiptType:  "relay",
	}
	q := newQueryRequest("providerclaim", newBody(params))
	rec := httptest.NewRecorder()
	NodeClaim(rec, q, httprouter.Params{})
	getJSONResponse(rec)

	<-evtChan
	params = QueryNodeReceiptParam{
		Address:      cb.GetAddress().String(),
		Blockchain:   "0001",
		AppPubKey:    cb.PublicKey.RawString(),
		SBlockHeight: 1,
		Height:       0,
		ReceiptType:  "relay",
	}
	q = newQueryRequest("providerclaim", newBody(params))
	rec = httptest.NewRecorder()
	NodeClaim(rec, q, httprouter.Params{})
	getJSONResponse(rec)

	cleanup()
	stopCli()
}

func TestRPC_Challenge(t *testing.T) {
	types.VbCCache = types.NewCache(1)
	codec.UpgradeHeight = 7000
	kb := getInMemoryKeybase()
	genBZ, keys, _, app := fiveValidatorsOneAppGenesis()
	_, _, cleanup := NewInMemoryTendermintNode(t, genBZ)
	_, err := kb.ExportPrivateKeyObject(app.Address, "test")
	assert.Nil(t, err)
	// Setup HandleDispatch Request
	key := NewValidChallengeProof(t, keys)
	// setup the query
	_, stopCli, evtChan := subscribeTo(t, tmTypes.EventNewBlock)
	<-evtChan // Wait for block
	q := newClientRequest("challenge", newBody(key))
	rec := httptest.NewRecorder()
	Challenge(rec, q, httprouter.Params{})
	resp := getJSONResponse(rec)
	rawResp := string(resp)
	assert.Equal(t, rec.Code, 200)
	assert.Contains(t, rawResp, "success")

	<-evtChan // Wait for block
	q = newClientRequest("challenge", newBody(key))
	rec = httptest.NewRecorder()
	Challenge(rec, q, httprouter.Params{})
	resp = getJSONResponse(rec)
	rawResp = string(resp)
	assert.Equal(t, rec.Code, 200)
	assert.Contains(t, rawResp, "success")

	cleanup()
	stopCli()
}

func TestRPC_SimRelay(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode") // TODO: Cannot create a config dir on pipeline
	}
	home := os.TempDir()
	datadir := home + types.DefaultDDName
	configPath := datadir + FS + types.ConfigDirName
	fmt.Println(configPath)
	app.GlobalConfig.ViperConfig = types.ViperConfig{
		ChainsName: types.DefaultChainsName,
		DataDir:    datadir,
	}
	generateChainsJson(configPath, []viperTypes.HostedBlockchain{{ID: dummyChainsHash, URL: dummyChainsURL}})
	expectedRequest := `"jsonrpc":"2.0","method":"web3_sha3","params":["0x68656c6c6f20776f726c64"],"id":64`
	expectedResponse := "0x47173285a8d7341e5e972fc677286384f802f8ef42a5ec5f03bbfa254cb01fad"
	defer gock.Off()
	gock.New(dummyChainsURL).
		Post("").
		BodyString(expectedRequest).
		Reply(200).
		BodyString(expectedResponse)
	payload := viperTypes.Payload{
		Path:   "/",
		Data:   expectedRequest,
		Method: "POST",
	}
	simParams := simRelayParams{
		RelayNetworkID: dummyChainsHash,
		Payload:        payload,
	}
	req := newClientRequest("sim", newBody(simParams))
	rec := httptest.NewRecorder()
	SimRequest(rec, req, httprouter.Params{})
	resp := getResponse(rec)
	assert.Equal(t, resp, expectedResponse)
}

func newBody(params interface{}) io.Reader {
	bz, err := json.Marshal(params)
	if err != nil {
		panic(err)
	}
	reader := bytes.NewReader(bz)
	return reader
}

func newCORSRequest(query string, body io.Reader) *http.Request {
	req, err := http.NewRequest("OPTIONS", "localhost:8081/v1/client/"+query, body)
	if err != nil {
		panic("could not create request: %v")
	}
	return req
}

func newClientRequest(query string, body io.Reader) *http.Request {
	req, err := http.NewRequest("POST", "localhost:8081/v1/client/"+query, body)
	if err != nil {
		panic("could not create request: %v")
	}
	return req
}

func newQueryRequest(query string, body io.Reader) *http.Request {
	req, err := http.NewRequest("POST", "localhost:8081/v1/query/"+query, body)
	if err != nil {
		panic("could not create request: %v")
	}
	return req
}

func getResponse(rec *httptest.ResponseRecorder) string {
	res := rec.Result()
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("could not read response: " + err.Error())
		return ""
	}
	if strings.Contains(string(b), "error") {
		return string(b)
	}

	resp, err := strconv.Unquote(string(b))
	if err != nil {
		fmt.Println("could not unquote resp: " + err.Error())
		return string(b)
	}
	return resp
}

func getJSONResponse(rec *httptest.ResponseRecorder) []byte {
	res := rec.Result()
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic("could not read response: " + err.Error())
	}
	return b
}

func NewValidChallengeProof(t *testing.T, privateKeys []crypto.PrivateKey) (challenge viperTypes.ChallengeProofInvalidData) {
	appPrivateKey := privateKeys[1]
	servicerPrivKey1 := privateKeys[4]
	servicerPrivKey2 := privateKeys[2]
	servicerPrivKey3 := privateKeys[3]
	clientPrivateKey := servicerPrivKey3
	appPubKey := appPrivateKey.PublicKey().RawString()
	servicerPubKey := servicerPrivKey1.PublicKey().RawString()
	servicerPubKey2 := servicerPrivKey2.PublicKey().RawString()
	servicerPubKey3 := servicerPrivKey3.PublicKey().RawString()
	reporterPrivKey := privateKeys[0]
	reporterPubKey := reporterPrivKey.PublicKey()
	reporterAddr := reporterPubKey.Address()
	clientPubKey := clientPrivateKey.PublicKey().RawString()
	var proof viperTypes.ChallengeProofInvalidData
	validProof := viperTypes.RelayProof{
		Entropy:            int64(rand.Intn(500000)),
		SessionBlockHeight: 1,
		ServicerPubKey:     servicerPubKey,
		RequestHash:        clientPubKey, // fake
		Blockchain:         PlaceholderHash,
		Token: viperTypes.AAT{
			Version:           "0.0.1",
			PlatformPublicKey: appPubKey,
			ClientPublicKey:   clientPubKey,
			PlatformSignature: "",
		},
		Signature: "",
	}
	appSignature, er := appPrivateKey.Sign(validProof.Token.Hash())
	if er != nil {
		t.Fatalf(er.Error())
	}
	validProof.Token.PlatformSignature = hex.EncodeToString(appSignature)
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
		Blockchain:         PlaceholderHash,
		Token: viperTypes.AAT{
			Version:           "0.0.1",
			PlatformPublicKey: appPubKey,
			ClientPublicKey:   clientPubKey,
			PlatformSignature: "",
		},
		Signature: "",
	}
	appSignature, er = appPrivateKey.Sign(validProof2.Token.Hash())
	if er != nil {
		t.Fatalf(er.Error())
	}
	validProof2.Token.PlatformSignature = hex.EncodeToString(appSignature)
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
		Blockchain:         PlaceholderHash,
		Token: viperTypes.AAT{
			Version:           "0.0.1",
			PlatformPublicKey: appPubKey,
			ClientPublicKey:   clientPubKey,
			PlatformSignature: "",
		},
		Signature: "",
	}
	appSignature, er = appPrivateKey.Sign(validProof3.Token.Hash())
	if er != nil {
		t.Fatalf(er.Error())
	}
	validProof3.Token.PlatformSignature = hex.EncodeToString(appSignature)
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
	proof = viperTypes.ChallengeProofInvalidData{
		MajorityResponses: []viperTypes.RelayResponse{
			majResp1,
			majResp2,
		},
		MinorityResponse: minResp,
		ReporterAddress:  types.Address(reporterAddr),
	}
	return proof
}
