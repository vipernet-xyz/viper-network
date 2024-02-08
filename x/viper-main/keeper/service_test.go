package keeper

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"testing"

	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	"github.com/vipernet-xyz/viper-network/crypto/keys"
	sdk "github.com/vipernet-xyz/viper-network/types"
	requestorsKeeper "github.com/vipernet-xyz/viper-network/x/requestors/keeper"
	requestorsTypes "github.com/vipernet-xyz/viper-network/x/requestors/types"
	"github.com/vipernet-xyz/viper-network/x/viper-main/types"

	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func setupHandleRelayTest(t *testing.T) (
	ctx sdk.Ctx,
	keeper Keeper,
	kvkeys map[string]*sdk.KVStoreKey,
	clientPrivateKey, appPrivateKey crypto.Ed25519PrivateKey,
	nodePubKey crypto.PublicKey,
	chain string,
	geoZone string, numServicers int64,
) {
	var kb keys.Keybase
	ctx, _, _, _, keeper, kvkeys, kb = createTestInput(t, false)

	chain = hex.EncodeToString([]byte{01})
	geoZone = hex.EncodeToString([]byte{01})
	clientPrivateKey = getRandomPrivateKey()

	kp, _ := kb.GetCoinbase()
	nodePubKey = kp.PublicKey
	numServicers = 5
	appPrivateKey = getRandomPrivateKey()

	appPubKey := appPrivateKey.PublicKey()
	app := requestorsTypes.NewRequestor(
		sdk.Address(appPubKey.Address()),
		appPubKey,
		[]string{chain},
		sdk.NewInt(1000000000),
		[]string{geoZone},
		numServicers,
	)

	// Stake app
	ak := keeper.requestorKeeper.(requestorsKeeper.Keeper)
	app.MaxRelays = ak.CalculateRequestorRelays(ctx, app)
	ak.SetRequestor(ctx, app)

	return
}

func testRelayAt(
	t *testing.T,
	ctx sdk.Ctx,
	keeper Keeper,
	clientBlockHeight int64,
	clientPrivateKey, appPrivateKey crypto.Ed25519PrivateKey,
	nodePubKey crypto.PublicKey,
	chain string,
	geozone string,
	numServicers int64,
) (*types.RelayResponse, sdk.Error) {
	clientPubKey := clientPrivateKey.PublicKey()
	appPubKey := appPrivateKey.PublicKey()
	blocksPerSesssion := keeper.BlocksPerSession(ctx)
	clientSessionHeight :=
		((clientBlockHeight-1)/blocksPerSesssion)*blocksPerSesssion + 1

	validRelay := types.Relay{
		Payload: types.Payload{
			Data: `{
			"jsonrpc":"2.0",
			"method":"web3_clientVersion",
			"params":[],
			"id":67
		}`,
			Method:  "",
			Path:    "",
			Headers: nil,
		},
		Meta: types.RelayMeta{BlockHeight: clientSessionHeight},
		Proof: types.RelayProof{
			Entropy:            rand.Int63(),
			SessionBlockHeight: clientSessionHeight,
			ServicerPubKey:     nodePubKey.RawString(),
			Blockchain:         chain,
			Token: types.AAT{
				Version:            "0.0.1",
				RequestorPublicKey: appPubKey.RawString(),
				ClientPublicKey:    clientPubKey.RawString(),
				RequestorSignature: "",
			},
			Signature:    "",
			GeoZone:      geozone,
			NumServicers: numServicers,
		},
	}

	validRelay.Proof.RequestHash = validRelay.RequestHashString()
	appSig, er := appPrivateKey.Sign(validRelay.Proof.Token.Hash())
	if er != nil {
		t.Fatalf(er.Error())
	}
	validRelay.Proof.Token.RequestorSignature = hex.EncodeToString(appSig)

	clientSig, er := clientPrivateKey.Sign(validRelay.Proof.Hash())
	if er != nil {
		t.Fatalf(er.Error())
	}
	assert.Nil(t, er)
	validRelay.Proof.Signature = hex.EncodeToString(clientSig)

	gock.New("https://www.google.com:443").
		Post("/").
		Reply(200).
		BodyString("bar")
	return keeper.HandleRelay(ctx, validRelay)
}

func TestKeeper_HandleRelay(t *testing.T) {
	ctx, keeper, kvkeys, clientPrivateKey, appPrivateKey, nodePubKey, chain, geoZone, numServicers :=
		setupHandleRelayTest(t)

	// Store the original allowances to clean up at the end of this test
	originalClientBlockSyncAllowance := types.GlobalViperConfig.ClientBlockSyncAllowance
	originalClientSessionSyncAllowance := types.GlobalViperConfig.ClientSessionSyncAllowance

	// Eliminate the impact of ClientBlockSyncAllowance to avoid any relay meta validation errors (OutOfSyncError)
	types.GlobalViperConfig.ClientBlockSyncAllowance = 10000

	nodeBlockHeight := ctx.BlockHeight()
	blocksPerSesssion := keeper.BlocksPerSession(ctx)
	latestSessionHeight := keeper.GetLatestSessionBlockHeight(ctx)

	t.Cleanup(func() {
		types.GlobalViperConfig.ClientBlockSyncAllowance = originalClientBlockSyncAllowance
		types.GlobalViperConfig.ClientSessionSyncAllowance = originalClientSessionSyncAllowance
		gock.Off() // Flush pending mocks after test execution
	})

	mockCtx := new(Ctx)
	mockCtx.On("KVStore", kvkeys["pos"]).Return(ctx.KVStore(kvkeys["pos"]))
	mockCtx.On("KVStore", kvkeys["params"]).Return(ctx.KVStore(kvkeys["params"]))
	mockCtx.On("BlockHeight").Return(ctx.BlockHeight())
	mockCtx.On("Logger").Return(ctx.Logger())
	mockCtx.On("PrevCtx", nodeBlockHeight).Return(ctx, nil)

	allSessionRangesTests := 4 // The range of block heights we will mock

	// Set up mocks for heights we'll query later.
	for i := int64(1); i <= blocksPerSesssion*int64(allSessionRangesTests); i++ {
		mockCtx.On("PrevCtx", nodeBlockHeight-i).Return(ctx, nil)
		mockCtx.On("PrevCtx", nodeBlockHeight+i).Return(ctx, nil)
	}
	fmt.Println("node bh:", nodeBlockHeight)

	// Client is synced with Node --> Success
	resp, err := testRelayAt(
		t,
		mockCtx,
		keeper,
		nodeBlockHeight,
		clientPrivateKey,
		appPrivateKey,
		nodePubKey,
		chain,
		geoZone,
		numServicers,
	)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp)
	assert.Equal(t, resp.Response, "bar")

	// TC 1:
	// Client is behind or advanced beyond Node's height with ClientSessionSyncAllowance 0
	// --> CodeInvalidBlockHeightError
	types.GlobalViperConfig.ClientSessionSyncAllowance = 0
	for i := 1; i <= allSessionRangesTests; i++ {
		resp, err = testRelayAt(
			t,
			mockCtx,
			keeper,
			latestSessionHeight-blocksPerSesssion*int64(i),
			clientPrivateKey,
			appPrivateKey,
			nodePubKey,
			chain,
			geoZone,
			numServicers,
		)
		assert.Nil(t, resp)
		assert.NotNil(t, err)
		assert.Equal(t, err.Codespace(), sdk.CodespaceType(types.ModuleName))
		assert.Equal(t, err.Code(), sdk.CodeType(types.CodeInvalidBlockHeightError))
		resp, err = testRelayAt(
			t,
			mockCtx,
			keeper,
			latestSessionHeight+blocksPerSesssion*int64(i),
			clientPrivateKey,
			appPrivateKey,
			nodePubKey,
			chain,
			geoZone,
			numServicers,
		)
		assert.Nil(t, resp)
		assert.NotNil(t, err)
		assert.Equal(t, err.Codespace(), sdk.CodespaceType(types.ModuleName))
		assert.Equal(t, err.Code(), sdk.CodeType(types.CodeInvalidBlockHeightError))
	}

	// TC2:
	// Test a relay while one session behind and forward,
	// while ClientSessionSyncAllowance = 1
	// --> Success on one session behind
	// --> InvalidBlockHeightError on one session forward
	sessionRangeTc := 1
	types.GlobalViperConfig.ClientSessionSyncAllowance = int64(sessionRangeTc)

	// First test the minimum boundary
	resp, err = testRelayAt(
		t,
		mockCtx,
		keeper,
		latestSessionHeight-blocksPerSesssion*int64(sessionRangeTc),
		clientPrivateKey,
		appPrivateKey,
		nodePubKey,
		chain,
		geoZone,
		numServicers,
	)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp)
	assert.Equal(t, resp.Response, "bar")

	// Second test the maximum boundary - Error
	resp, err = testRelayAt(
		t,
		mockCtx,
		keeper,
		latestSessionHeight+blocksPerSesssion*int64(sessionRangeTc),
		clientPrivateKey,
		appPrivateKey,
		nodePubKey,
		chain,
		geoZone,
		numServicers,
	)
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, err.Codespace(), sdk.CodespaceType(types.ModuleName))
	assert.Equal(t, err.Code(), sdk.CodeType(types.CodeInvalidBlockHeightError))

	// TC2:
	// Test a relay while two sessions behind and forward,
	// while ClientSessionSyncAllowance = 1
	// --> InvalidBlockHeightError on two sessions behind and forwards
	sessionRangeTc = 2
	types.GlobalViperConfig.ClientSessionSyncAllowance = 1

	// First test two sessions back - Error
	resp, err = testRelayAt(
		t,
		mockCtx,
		keeper,
		latestSessionHeight-blocksPerSesssion*int64(sessionRangeTc),
		clientPrivateKey,
		appPrivateKey,
		nodePubKey,
		chain,
		geoZone,
		numServicers,
	)
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, err.Codespace(), sdk.CodespaceType(types.ModuleName))
	assert.Equal(t, err.Code(), sdk.CodeType(types.CodeInvalidBlockHeightError))

	// Second test two sessions forward - Error
	resp, err = testRelayAt(
		t,
		mockCtx,
		keeper,
		latestSessionHeight+blocksPerSesssion*int64(sessionRangeTc),
		clientPrivateKey,
		appPrivateKey,
		nodePubKey,
		chain,
		geoZone,
		numServicers,
	)
	assert.Nil(t, resp)
	assert.NotNil(t, err)
	assert.Equal(t, err.Codespace(), sdk.CodespaceType(types.ModuleName))
	assert.Equal(t, err.Code(), sdk.CodeType(types.CodeInvalidBlockHeightError))
}

func setupHandleWebsocketRelayTest(t *testing.T) (
	ctx sdk.Ctx,
	keeper Keeper,
	kvkeys map[string]*sdk.KVStoreKey,
	clientPrivateKey, appPrivateKey crypto.Ed25519PrivateKey,
	nodePubKey crypto.PublicKey,
	chain string,
	geoZone string, numServicers int64,
) {
	var kb keys.Keybase
	ctx, _, _, _, keeper, kvkeys, kb = createTestInput(t, false)

	chain = hex.EncodeToString([]byte{01})
	geoZone = hex.EncodeToString([]byte{01})
	clientPrivateKey = getRandomPrivateKey()

	kp, _ := kb.GetCoinbase()
	nodePubKey = kp.PublicKey
	numServicers = 5
	appPrivateKey = getRandomPrivateKey()

	appPubKey := appPrivateKey.PublicKey()
	app := requestorsTypes.NewRequestor(
		sdk.Address(appPubKey.Address()),
		appPubKey,
		[]string{chain},
		sdk.NewInt(1000000000),
		[]string{geoZone},
		numServicers,
	)

	// Stake app
	ak := keeper.requestorKeeper.(requestorsKeeper.Keeper)
	app.MaxRelays = ak.CalculateRequestorRelays(ctx, app)
	ak.SetRequestor(ctx, app)

	return
}

func testWebsocketRelayAt(
	t *testing.T,
	ctx sdk.Ctx,
	keeper Keeper,
	clientBlockHeight int64,
	clientPrivateKey, appPrivateKey crypto.Ed25519PrivateKey,
	nodePubKey crypto.PublicKey,
	chain string,
	geozone string,
	numServicers int64,
) (chan *types.RelayResponse, error) {
	clientPubKey := clientPrivateKey.PublicKey()
	appPubKey := appPrivateKey.PublicKey()
	blocksPerSesssion := keeper.BlocksPerSession(ctx)
	clientSessionHeight :=
		((clientBlockHeight-1)/blocksPerSesssion)*blocksPerSesssion + 1

	validWebsocketRelay := types.Relay{
		Payload: types.Payload{
			Data: `{
			"jsonrpc":"2.0",
			"method":"web3_clientVersion",
			"params":[],
			"id":67
		}`,
			Method:  "",
			Path:    "",
			Headers: nil,
		},
		Meta: types.RelayMeta{BlockHeight: clientSessionHeight},
		Proof: types.RelayProof{
			Entropy:            rand.Int63(),
			SessionBlockHeight: clientSessionHeight,
			ServicerPubKey:     nodePubKey.RawString(),
			Blockchain:         chain,
			Token: types.AAT{
				Version:            "0.0.1",
				RequestorPublicKey: appPubKey.RawString(),
				ClientPublicKey:    clientPubKey.RawString(),
				RequestorSignature: "",
			},
			Signature:    "",
			GeoZone:      geozone,
			NumServicers: numServicers,
		},
	}

	validWebsocketRelay.Proof.RequestHash = validWebsocketRelay.RequestHashString()
	appSig, er := appPrivateKey.Sign(validWebsocketRelay.Proof.Token.Hash())
	if er != nil {
		t.Fatalf(er.Error())
	}
	validWebsocketRelay.Proof.Token.RequestorSignature = hex.EncodeToString(appSig)

	clientSig, er := clientPrivateKey.Sign(validWebsocketRelay.Proof.Hash())
	if er != nil {
		t.Fatalf(er.Error())
	}
	assert.Nil(t, er)
	validWebsocketRelay.Proof.Signature = hex.EncodeToString(clientSig)

	return keeper.HandleWebsocketRelay(ctx, validWebsocketRelay)
}

func TestKeeper_HandleWebsocketRelay(t *testing.T) {
	ctx, keeper, kvkeys, clientPrivateKey, appPrivateKey, nodePubKey, chain, geoZone, numServicers :=
		setupHandleWebsocketRelayTest(t)

	// Store the original allowances to clean up at the end of this test
	originalClientBlockSyncAllowance := types.GlobalViperConfig.ClientBlockSyncAllowance
	originalClientSessionSyncAllowance := types.GlobalViperConfig.ClientSessionSyncAllowance

	// Eliminate the impact of ClientBlockSyncAllowance to avoid any relay meta validation errors (OutOfSyncError)
	types.GlobalViperConfig.ClientBlockSyncAllowance = 10000

	nodeBlockHeight := ctx.BlockHeight()
	blocksPerSesssion := keeper.BlocksPerSession(ctx)
	latestSessionHeight := keeper.GetLatestSessionBlockHeight(ctx)

	t.Cleanup(func() {
		types.GlobalViperConfig.ClientBlockSyncAllowance = originalClientBlockSyncAllowance
		types.GlobalViperConfig.ClientSessionSyncAllowance = originalClientSessionSyncAllowance
		gock.Off() // Flush pending mocks after test execution
	})

	mockCtx := new(Ctx)
	mockCtx.On("KVStore", kvkeys["pos"]).Return(ctx.KVStore(kvkeys["pos"]))
	mockCtx.On("KVStore", kvkeys["params"]).Return(ctx.KVStore(kvkeys["params"]))
	mockCtx.On("BlockHeight").Return(ctx.BlockHeight())
	mockCtx.On("Logger").Return(ctx.Logger())
	mockCtx.On("PrevCtx", nodeBlockHeight).Return(ctx, nil)

	allSessionRangesTests := 4 // The range of block heights we will mock

	// Set up mocks for heights we'll query later.
	for i := int64(1); i <= blocksPerSesssion*int64(allSessionRangesTests); i++ {
		mockCtx.On("PrevCtx", nodeBlockHeight-i).Return(ctx, nil)
		mockCtx.On("PrevCtx", nodeBlockHeight+i).Return(ctx, nil)
	}
	fmt.Println("node bh:", nodeBlockHeight)

	// Case 1: Client is synced with Node --> Success
	resChan, err := testWebsocketRelayAt(
		t,
		mockCtx,
		keeper,
		nodeBlockHeight,
		clientPrivateKey,
		appPrivateKey,
		nodePubKey,
		chain,
		geoZone,
		numServicers,
	)
	assert.Nil(t, err)
	assert.NotNil(t, resChan)

	// Case 2: Test when the relay is out of sync with the client's session
	types.GlobalViperConfig.ClientSessionSyncAllowance = 0
	_, err = testWebsocketRelayAt(
		t,
		mockCtx,
		keeper,
		latestSessionHeight-blocksPerSesssion,
		clientPrivateKey,
		appPrivateKey,
		nodePubKey,
		chain,
		geoZone,
		numServicers,
	)
	assert.NotNil(t, err)

	// Case 3: Test when the relay is within the client's session sync allowance
	types.GlobalViperConfig.ClientSessionSyncAllowance = 1
	resChan, err = testWebsocketRelayAt(
		t,
		mockCtx,
		keeper,
		latestSessionHeight-blocksPerSesssion,
		clientPrivateKey,
		appPrivateKey,
		nodePubKey,
		chain,
		geoZone,
		numServicers,
	)
	assert.Nil(t, err)
	assert.NotNil(t, resChan)

}
