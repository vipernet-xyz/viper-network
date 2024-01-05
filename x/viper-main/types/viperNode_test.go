package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	sdk "github.com/vipernet-xyz/viper-network/types"
)

func TestViperNodeAdd(t *testing.T) {
	key := GetRandomPrivateKey()
	address := sdk.GetAddress(key.PublicKey())
	AddViperNode(key, log.NewNopLogger())
	_, ok := GlobalViperNodes[address.String()]
	assert.True(t, ok)
}

func TestViperNodeGetByAddress(t *testing.T) {
	key := GetRandomPrivateKey()
	address := sdk.GetAddress(key.PublicKey())
	AddViperNode(key, log.NewNopLogger())
	node, err := GetViperNodeByAddress(&address)
	assert.Nil(t, err)
	assert.NotNil(t, node)
}

func TestViperNodeGet(t *testing.T) {
	key := GetRandomPrivateKey()
	AddViperNode(key, log.NewNopLogger())
	node := GetViperNode()
	assert.NotNil(t, node)
}

func TestViperNodeCleanCache(t *testing.T) {
	key := GetRandomPrivateKey()
	AddViperNode(key, log.NewNopLogger())
	CleanViperNodes()
	assert.Nil(t, GlobalSessionCache)
	assert.Nil(t, GlobalEvidenceCache)
	assert.EqualValues(t, 0, len(GlobalViperNodes))
}

func TestViperNodeInitCache(t *testing.T) {
	CleanViperNodes()
	key := GetRandomPrivateKey()
	testingConfig := sdk.DefaultTestingViperConfig()
	AddViperNode(key, log.NewNopLogger())
	InitViperNodeCaches(testingConfig, log.NewNopLogger())
	address := sdk.GetAddress(key.PublicKey())
	node, err := GetViperNodeByAddress(&address)
	assert.NotNil(t, GlobalSessionCache)
	assert.NotNil(t, GlobalEvidenceCache)
	assert.EqualValues(t, 1, len(GlobalViperNodes))
	assert.Nil(t, err)
	assert.NotNil(t, node.EvidenceStore)
	assert.NotNil(t, node.SessionStore)
}

func TestViperNodeInitCaches(t *testing.T) {
	CleanViperNodes()
	key := GetRandomPrivateKey()
	key2 := GetRandomPrivateKey()
	key3 := GetRandomPrivateKey()
	logger := log.NewNopLogger()
	testingConfig := sdk.DefaultTestingViperConfig()
	testingConfig.ViperConfig.LeanViper = true
	AddViperNode(key, logger)
	AddViperNode(key2, logger)
	AddViperNode(key3, logger)
	InitViperNodeCaches(testingConfig, logger)
	assert.NotNil(t, GlobalSessionCache)
	assert.NotNil(t, GlobalEvidenceCache)
	assert.NotNil(t, GlobalTestCache)
	assert.EqualValues(t, 3, len(GlobalViperNodes))
	addresses := []sdk.Address{sdk.GetAddress(key.PublicKey()), sdk.GetAddress(key2.PublicKey()), sdk.GetAddress(key3.PublicKey())}
	for _, address := range addresses {
		node, err := GetViperNodeByAddress(&address)
		assert.Nil(t, err)
		assert.NotNil(t, node.EvidenceStore)
		assert.NotNil(t, node.SessionStore)
		assert.NotNil(t, node.TestStore)
	}
}
