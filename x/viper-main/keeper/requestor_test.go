package keeper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRequestor(t *testing.T) {
	ctx, _, requestors, _, keeper, _, _ := createTestInput(t, false)
	a, found := keeper.GetRequestor(ctx, requestors[0].Address)
	assert.True(t, found)
	assert.Equal(t, a, requestors[0])
	randomAddr := getRandomValidatorAddress()
	_, found = keeper.GetRequestor(ctx, randomAddr)
	assert.False(t, found)
}

func TestGetRequestorFromPublicKey(t *testing.T) {
	ctx, _, requestors, _, keeper, _, _ := createTestInput(t, false)
	pk := requestors[0].PublicKey.RawString()
	a, found := keeper.GetRequestorFromPublicKey(ctx, pk)
	assert.True(t, found)
	assert.Equal(t, a, requestors[0])
	randomPubKey := getRandomPubKey().String()
	_, found = keeper.GetRequestorFromPublicKey(ctx, randomPubKey)
	assert.False(t, found)
}
