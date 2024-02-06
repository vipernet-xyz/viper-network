package types

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSessionKey(t *testing.T) {
	requestorPubKey := getRandomPubKey()
	ctx := newContext(t, false).WithAppVersion("0.0.0")
	blockhash := hex.EncodeToString(ctx.BlockHeader().LastBlockId.Hash)
	ethereum := hex.EncodeToString([]byte{01})
	bitcoin := hex.EncodeToString([]byte{02})
	key1, err := NewSessionKey(requestorPubKey.RawString(), ethereum, blockhash)
	assert.Nil(t, err)
	assert.NotNil(t, key1)
	assert.NotEmpty(t, key1)
	assert.Nil(t, HashVerification(hex.EncodeToString(key1)))
	key2, err := NewSessionKey(requestorPubKey.RawString(), bitcoin, blockhash)
	assert.Nil(t, err)
	assert.NotNil(t, key2)
	assert.NotEmpty(t, key2)
	assert.Nil(t, HashVerification(hex.EncodeToString(key2)))
	assert.Equal(t, len(key1), len(key2))
	assert.NotEqual(t, key1, key2)
}

func TestSessionKey_Validate(t *testing.T) {
	fakeKey1 := SessionKey([]byte("fakekey"))
	fakeKey2 := SessionKey([]byte(""))
	realKey := SessionKey(merkleHash([]byte("validKey")))
	assert.NotNil(t, fakeKey1.Validate())
	assert.NotNil(t, fakeKey2.Validate())
	assert.Nil(t, realKey.Validate())
}
