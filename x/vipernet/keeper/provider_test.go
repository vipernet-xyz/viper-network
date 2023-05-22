package keeper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetProvider(t *testing.T) {
	ctx, _, providers, _, keeper, _, _ := createTestInput(t, false)
	a, found := keeper.GetProvider(ctx, providers[0].Address)
	assert.True(t, found)
	assert.Equal(t, a, providers[0])
	randomAddr := getRandomValidatorAddress()
	_, found = keeper.GetProvider(ctx, randomAddr)
	assert.False(t, found)
}

func TestGetProviderFromPublicKey(t *testing.T) {
	ctx, _, providers, _, keeper, _, _ := createTestInput(t, false)
	pk := providers[0].PublicKey.RawString()
	a, found := keeper.GetProviderFromPublicKey(ctx, pk)
	assert.True(t, found)
	assert.Equal(t, a, providers[0])
	randomPubKey := getRandomPubKey().String()
	_, found = keeper.GetProviderFromPublicKey(ctx, randomPubKey)
	assert.False(t, found)
}
