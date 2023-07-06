package keeper

import (
	"encoding/hex"
	"testing"

	sdk "github.com/vipernet-xyz/viper-network/types"
	providersKeeper "github.com/vipernet-xyz/viper-network/x/providers/keeper"
	providersTypes "github.com/vipernet-xyz/viper-network/x/providers/types"
	"github.com/vipernet-xyz/viper-network/x/vipernet/types"

	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestKeeper_HandleRelay(t *testing.T) {
	ethereum := hex.EncodeToString([]byte{01})
	ctx, _, _, _, keeper, keys, kb := createTestInput(t, false)
	mockCtx := new(Ctx)
	ak := keeper.providerKeeper.(providersKeeper.Keeper)
	clientPrivateKey := getRandomPrivateKey()
	clientPubKey := clientPrivateKey.PublicKey().RawString()
	providerPrivateKey := getRandomPrivateKey()
	apk := providerPrivateKey.PublicKey()
	providerPubKey := apk.RawString()
	// add provider to world state
	provider := providersTypes.NewProvider(sdk.Address(apk.Address()), apk, []string{ethereum}, sdk.NewInt(10000000), []string{"0001"}, 5)
	// calculate relays
	provider.MaxRelays = ak.CalculateProviderRelays(ctx, provider)
	// set the vals from the data
	ak.SetProvider(ctx, provider)
	ak.SetStakedProvider(ctx, provider)
	kp, _ := kb.GetCoinbase()
	npk := kp.PublicKey
	servicerPubKey := npk.RawString()
	p := types.Payload{
		Data:    "{\"jsonrpc\":\"2.0\",\"method\":\"web3_clientVersion\",\"params\":[],\"id\":67}",
		Method:  "",
		Path:    "",
		Headers: nil,
	}
	validRelay := types.Relay{
		Payload: p,
		Meta:    types.RelayMeta{BlockHeight: 976},
		Proof: types.RelayProof{
			Entropy:            1,
			SessionBlockHeight: 976,
			ServicerPubKey:     servicerPubKey,
			Blockchain:         ethereum,
			Token: types.AAT{
				Version:           "0.0.1",
				ProviderPublicKey: providerPubKey,
				ClientPublicKey:   clientPubKey,
				ProviderSignature: "",
			},
			Signature: "",
		},
	}
	validRelay.Proof.RequestHash = validRelay.RequestHashString()
	providerSig, er := providerPrivateKey.Sign(validRelay.Proof.Token.Hash())
	if er != nil {
		t.Fatalf(er.Error())
	}
	validRelay.Proof.Token.ProviderSignature = hex.EncodeToString(providerSig)
	clientSig, er := clientPrivateKey.Sign(validRelay.Proof.Hash())
	if er != nil {
		t.Fatalf(er.Error())
	}
	validRelay.Proof.Signature = hex.EncodeToString(clientSig)
	defer gock.Off() // Flush pending mocks after test execution

	gock.New("https://www.google.com:443").
		Post("/").
		Reply(200).
		BodyString("bar")

	mockCtx.On("KVStore", keeper.storeKey).Return(ctx.KVStore(keeper.storeKey))
	mockCtx.On("KVStore", keys["pos"]).Return(ctx.KVStore(keys["pos"]))
	mockCtx.On("KVStore", keys["params"]).Return(ctx.KVStore(keys["params"]))
	mockCtx.On("KVStore", keys["application"]).Return(ctx.KVStore(keys["application"]))
	mockCtx.On("BlockHeight").Return(ctx.BlockHeight())
	mockCtx.On("PrevCtx", int64(976)).Return(ctx, nil)
	mockCtx.On("PrevCtx", keeper.GetLatestSessionBlockHeight(mockCtx)).Return(ctx, nil)
	mockCtx.On("Logger").Return(ctx.Logger())

	resp, err := keeper.HandleRelay(mockCtx, validRelay)
	assert.Nil(t, err, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp)
	assert.Equal(t, resp.Response, "bar")
}
