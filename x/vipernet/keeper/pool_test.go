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

func TestKeeper_GetServicersStakedTokens(t *testing.T) {
	ctx, vals, _, _, k, _, _ := createTestInput(t, false)
	assert.NotZero(t, len(vals))
	tokens := vals[0].StakedTokens
	assert.Equal(t, k.GetServicersStakedTokens(ctx), tokens.Mul(types.NewInt(int64(len(vals)))))
}

func TestKeeper_GetProvidersStakedTokens(t *testing.T) {
	ctx, _, providers, _, k, _, _ := createTestInput(t, false)
	assert.NotZero(t, len(providers))
	tokens := providers[0].StakedTokens
	assert.Equal(t, k.GetProviderStakedTokens(ctx), tokens.Mul(types.NewInt(int64(len(providers)))))
}

func TestKeeper_GetTotalStakedTokens(t *testing.T) {
	ctx, vals, providers, _, k, _, _ := createTestInput(t, false)
	assert.NotZero(t, len(providers))
	providerToken := providers[0].StakedTokens
	providerTokens := providerToken.Mul(types.NewInt(int64(len(providers))))
	valToken := vals[0].StakedTokens
	valTokens := valToken.Mul(types.NewInt(int64(len(vals))))
	assert.Equal(t, k.GetTotalStakedTokens(ctx), providerTokens.Add(valTokens))
}
