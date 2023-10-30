package keeper

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/vipernet-xyz/viper-network/x/vipernet/types"

	"github.com/stretchr/testify/assert"
)

func TestKeeper_Dispatch(t *testing.T) {
	ctx, _, _, _, keeper, keys, _ := createTestInput(t, false)
	providerPrivateKey := getRandomPrivateKey()
	providerPubKey := providerPrivateKey.PublicKey().RawString()
	ethereum := hex.EncodeToString([]byte{01})
	bitcoin := hex.EncodeToString([]byte{02})
	US := hex.EncodeToString([]byte{01})
	EU := hex.EncodeToString([]byte{03})
	// create a session header
	validHeader := types.SessionHeader{
		ProviderPubKey:     providerPubKey,
		Chain:              ethereum,
		SessionBlockHeight: 977,
		GeoZone:            US,
		NumServicers:       5,
	}
	// create an invalid session header
	invalidHeader := types.SessionHeader{
		ProviderPubKey:     providerPubKey,
		Chain:              bitcoin,
		SessionBlockHeight: 977,
		GeoZone:            EU,
		NumServicers:       7,
	}
	mockCtx := new(Ctx)
	mockCtx.On("KVStore", keeper.storeKey).Return(ctx.KVStore(keeper.storeKey))
	mockCtx.On("KVStore", keys["pos"]).Return(ctx.KVStore(keys["pos"]))
	mockCtx.On("KVStore", keys["params"]).Return(ctx.KVStore(keys["params"]))
	mockCtx.On("PrevCtx", validHeader.SessionBlockHeight).Return(ctx, nil)
	mockCtx.On("BlockHeight").Return(ctx.BlockHeight())
	mockCtx.On("Logger").Return(ctx.Logger())
	res, err := keeper.HandleDispatch(mockCtx, validHeader)
	assert.Nil(t, err)
	assert.Equal(t, res.Session.SessionHeader.Chain, ethereum)
	assert.Equal(t, res.Session.SessionHeader.SessionBlockHeight, int64(977))
	assert.Equal(t, res.Session.SessionHeader.ProviderPubKey, providerPubKey)
	assert.Equal(t, res.Session.SessionHeader, validHeader)
	assert.Len(t, res.Session.SessionServicers, 5)
	_, err = keeper.HandleDispatch(mockCtx, invalidHeader)
	assert.NotNil(t, err)
}

func TestKeeper_IsSessionBlock(t *testing.T) {
	notSessionContext, _, _, _, keeper, _, _ := createTestInput(t, false)
	fmt.Println(t, keeper.IsSessionBlock(notSessionContext.WithBlockHeight(977)))
	//assert.False(t, keeper.IsSessionBlock(notSessionContext.WithBlockHeight(977)))
}

func TestKeeper_IsViperSupportedBlockchain(t *testing.T) {
	ctx, _, _, _, keeper, _, _ := createTestInput(t, false)
	sb := []string{"ethereum"}
	notSB := "bitcoin"
	p := types.Params{
		SupportedBlockchains: sb,
	}
	keeper.SetParams(ctx, p)
	assert.True(t, keeper.IsViperSupportedBlockchain(ctx, "ethereum"))
	assert.False(t, keeper.IsViperSupportedBlockchain(ctx, notSB))
}
