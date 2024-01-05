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

func TestKeeper_GetRequestorsStakedTokens(t *testing.T) {
	ctx, _, requestors, _, k, _, _ := createTestInput(t, false)
	assert.NotZero(t, len(requestors))
	tokens := requestors[0].StakedTokens
	assert.Equal(t, k.GetRequestorStakedTokens(ctx), tokens.Mul(types.NewInt(int64(len(requestors)))))
}

func TestKeeper_GetTotalStakedTokens(t *testing.T) {
	ctx, vals, requestors, _, k, _, _ := createTestInput(t, false)
	assert.NotZero(t, len(requestors))
	requestorToken := requestors[0].StakedTokens
	requestorTokens := requestorToken.Mul(types.NewInt(int64(len(requestors))))
	valToken := vals[0].StakedTokens
	valTokens := valToken.Mul(types.NewInt(int64(len(vals))))
	assert.Equal(t, k.GetTotalStakedTokens(ctx), requestorTokens.Add(valTokens))
}
