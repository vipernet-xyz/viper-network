package types

import (
	"crypto/rand"
	"reflect"
	"testing"
	"time"

	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"

	"github.com/stretchr/testify/assert"
)

func TestLegacyValidator_ToFromValidator(t *testing.T) {
	var pub crypto.Ed25519PublicKey
	_, err := rand.Read(pub[:])
	if err != nil {
		_ = err
	}
	lv := LegacyValidator{
		Address:                 sdk.Address(pub.Address()),
		PublicKey:               pub,
		Jailed:                  false,
		Status:                  sdk.Staked,
		Chains:                  []string{"0001"},
		ServiceURL:              "foo.bar",
		StakedTokens:            sdk.OneInt(),
		UnstakingCompletionTime: time.Now(),
	}
	validator := lv.ToValidator()
	lv2 := validator.ToLegacy()
	assert.True(t, reflect.DeepEqual(lv, lv2))
}
