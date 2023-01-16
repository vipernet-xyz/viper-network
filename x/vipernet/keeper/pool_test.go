package keeper

import (
	"testing"

	"github.com/vipernet-xyz/viper-network/types"

	"github.com/stretchr/testify/assert"
)

func TestKeeper_StakeDenom(t *testing.T) {
	ctx, _, _, _, k, _, _ := createTestInput(t, false)
	stakeDenom := types.DefaultStakeDenom
	assert.Equal(t, stakeDenom, k.posKeeper.StakeDenom(ctx))
}

func TestKeeper_GetNodesStakedTokens(t *testing.T) {
	ctx, vals, _, _, k, _, _ := createTestInput(t, false)
	assert.NotZero(t, len(vals))
	tokens := vals[0].StakedTokens
	assert.Equal(t, k.GetNodesStakedTokens(ctx), tokens.Mul(types.NewInt(int64(len(vals)))))
}

func TestKeeper_GetPlatformsStakedTokens(t *testing.T) {
	ctx, _, platforms, _, k, _, _ := createTestInput(t, false)
	assert.NotZero(t, len(platforms))
	tokens := platforms[0].StakedTokens
	assert.Equal(t, k.GetPlatformStakedTokens(ctx), tokens.Mul(types.NewInt(int64(len(platforms)))))
}

func TestKeeper_GetTotalStakedTokens(t *testing.T) {
	ctx, vals, platforms, _, k, _, _ := createTestInput(t, false)
	assert.NotZero(t, len(platforms))
	platformToken := platforms[0].StakedTokens
	platformTokens := platformToken.Mul(types.NewInt(int64(len(platforms))))
	valToken := vals[0].StakedTokens
	valTokens := valToken.Mul(types.NewInt(int64(len(vals))))
	assert.Equal(t, k.GetTotalStakedTokens(ctx), platformTokens.Add(valTokens))
}
