package keeper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPlatform(t *testing.T) {
	ctx, _, platforms, _, keeper, _, _ := createTestInput(t, false)
	a, found := keeper.GetPlatform(ctx, platforms[0].Address)
	assert.True(t, found)
	assert.Equal(t, a, platforms[0])
	randomAddr := getRandomValidatorAddress()
	_, found = keeper.GetPlatform(ctx, randomAddr)
	assert.False(t, found)
}

func TestGetPlatformFromPublicKey(t *testing.T) {
	ctx, _, platforms, _, keeper, _, _ := createTestInput(t, false)
	pk := platforms[0].PublicKey.RawString()
	a, found := keeper.GetPlatformFromPublicKey(ctx, pk)
	assert.True(t, found)
	assert.Equal(t, a, platforms[0])
	randomPubKey := getRandomPubKey().String()
	_, found = keeper.GetPlatformFromPublicKey(ctx, randomPubKey)
	assert.False(t, found)
}
