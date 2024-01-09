package keeper

import (
	"testing"

	"github.com/vipernet-xyz/viper-network/crypto/keys/mintkey"

	"github.com/stretchr/testify/assert"
)

func TestAATGeneration(t *testing.T) {
	passphrase := "test"
	kb := NewTestKeybase()
	kp, err := kb.Create(passphrase)
	assert.Nil(t, err)
	privkey, err := mintkey.UnarmorDecryptPrivKey(kp.PrivKeyArmor, passphrase)
	assert.Nil(t, err)
	requestorPubKey := kp.PublicKey
	res, err := AATGeneration(requestorPubKey.RawString(), requestorPubKey.RawString(), privkey)
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Nil(t, res.Validate())
}