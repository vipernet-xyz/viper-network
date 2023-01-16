package types

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSessionKey(t *testing.T) {
	platformPubKey := getRandomPubKey()
	ctx := newContext(t, false).WithAppVersion("0.0.0")
	blockhash := hex.EncodeToString(ctx.BlockHeader().LastBlockId.Hash)
	ethereum := hex.EncodeToString([]byte{01})
	bitcoin := hex.EncodeToString([]byte{02})
	key1, err := NewSessionKey(platformPubKey.RawString(), ethereum, blockhash)
	assert.Nil(t, err)
	assert.NotNil(t, key1)
	assert.NotEmpty(t, key1)
	assert.Nil(t, HashVerification(hex.EncodeToString(key1)))
	key2, err := NewSessionKey(platformPubKey.RawString(), bitcoin, blockhash)
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

//func TestNewSessionProviders(t *testing.T) {
//	fakeSessionKey, err := hex.DecodeString("36f028580bb02cc8272a9a020f4200e346e276ae664e45ee80745574e2f5ab80")
//	if err != nil {
//		t.Fatalf(err.Error())
//	}
//	fakePubKey1, err := crypto.NewPublicKey("36f028580bb02cc8272a9a020f4200e346e276ae664e45ee80745574e2f5ab81")
//	if err != nil {
//		t.Fatalf(err.Error())
//	}
//	fakePubKey2, err := crypto.NewPublicKey("36f028580bb02cc8272a9a020f4200e346e276ae664e45ee80745574e2f5ab82")
//	if err != nil {
//		t.Fatalf(err.Error())
//	}
//	fakePubKey3, err := crypto.NewPublicKey("36f028580bb02cc8272a9a020f4200e346e276ae664e45ee80745574e2f5ab83")
//	if err != nil {
//		t.Fatalf(err.Error())
//	}
//	fakePubKey4, err := crypto.NewPublicKey("36f028580bb02cc8272a9a020f4200e346e276ae664e45ee80745574e2f5ab84")
//	if err != nil {
//		t.Fatalf(err.Error())
//	}
//	fakePubKey5, err := crypto.NewPublicKey("36f028580bb02cc8272a9a020f4200e346e276ae664e45ee80745574e2f5ab85")
//	if err != nil {
//		t.Fatalf(err.Error())
//	}
//	fakePubKey6, err := crypto.NewPublicKey("36f028580bb02cc8272a9a020f4200e346e276ae664e45ee80745574e2f5ab86")
//	if err != nil {
//		t.Fatalf(err.Error())
//	}
//	fakePubKey7, err := crypto.NewPublicKey("36f028580bb02cc8272a9a020f4200e346e276ae664e45ee80745574e2f5ab87")
//	if err != nil {
//		t.Fatalf(err.Error())
//	}
//	fakePubKey8, err := crypto.NewPublicKey("36f028580bb02cc8272a9a020f4200e346e276ae664e45ee80745574e2f5ab88")
//	if err != nil {
//		t.Fatalf(err.Error())
//	}
//	fakePubKey9, err := crypto.NewPublicKey("36f028580bb02cc8272a9a020f4200e346e276ae664e45ee80745574e2f5ab89")
//	if err != nil {
//		t.Fatalf(err.Error())
//	}
//	fakePubKey10, err := crypto.NewPublicKey("36f028580bb02cc8272a9a020f4200e346e276ae664e45ee80745574e2f5ab8A")
//	if err != nil {
//		t.Fatalf(err.Error())
//	}
//	fakePubKey11, err := crypto.NewPublicKey("36f028580bb02cc8272a9a020f4200e346e276ae664e45ee80745574e2f5ab8B")
//	if err != nil {
//		t.Fatalf(err.Error())
//	}
//	fakePubKey12, err := crypto.NewPublicKey("36f028580bb02cc8272a9a020f4200e346e276ae664e45ee80745574e2f5ab8C")
//	if err != nil {
//		t.Fatalf(err.Error())
//	}
//	ethereum := hex.EncodeToString([]byte{01})
//	var allProviders []exported.ValidatorI
//	provider12 := providersTypes.Validator{
//		Address:                 sdk.Address(fakePubKey12.Address()),
//		PublicKey:               fakePubKey12,
//		Jailed:                  false,
//		Status:                  sdk.Staked,
//		Chains:                  []string{ethereum},
//		ServiceUrl:              "https://www.google.com:443",
//		StakedTokens:            sdk.NewInt(100000),
//		UnstakingCompletionTime: time.Time{},
//	}
//	provider1 := providersTypes.Validator{
//		Address:                 sdk.Address(fakePubKey1.Address()),
//		PublicKey:               (fakePubKey1),
//		Jailed:                  false,
//		Status:                  sdk.Staked,
//		Chains:                  []string{ethereum},
//		ServiceUrl:              "https://www.google.com:443",
//		StakedTokens:            sdk.NewInt(100000),
//		UnstakingCompletionTime: time.Time{},
//	}
//	provider2 := providersTypes.Validator{
//		Address:                 sdk.Address(fakePubKey2.Address()),
//		PublicKey:               (fakePubKey2),
//		Jailed:                  false,
//		Status:                  sdk.Staked,
//		Chains:                  []string{ethereum},
//		ServiceUrl:              "https://www.google.com:443",
//		StakedTokens:            sdk.NewInt(100000),
//		UnstakingCompletionTime: time.Time{},
//	}
//	provider3 := providersTypes.Validator{
//		Address:                 sdk.Address(fakePubKey3.Address()),
//		PublicKey:               (fakePubKey3),
//		Jailed:                  false,
//		Status:                  sdk.Staked,
//		Chains:                  []string{ethereum},
//		ServiceUrl:              "https://www.google.com:443",
//		StakedTokens:            sdk.NewInt(100000),
//		UnstakingCompletionTime: time.Time{},
//	}
//	provider4 := providersTypes.Validator{
//		Address:                 sdk.Address(fakePubKey4.Address()),
//		PublicKey:               (fakePubKey4),
//		Jailed:                  false,
//		Status:                  sdk.Staked,
//		Chains:                  []string{ethereum},
//		ServiceUrl:              "https://www.google.com:443",
//		StakedTokens:            sdk.NewInt(100000),
//		UnstakingCompletionTime: time.Time{},
//	}
//	provider5 := providersTypes.Validator{
//		Address:                 sdk.Address(fakePubKey5.Address()),
//		PublicKey:               (fakePubKey5),
//		Jailed:                  false,
//		Status:                  sdk.Staked,
//		Chains:                  []string{ethereum},
//		ServiceUrl:              "https://www.google.com:443",
//		StakedTokens:            sdk.NewInt(100000),
//		UnstakingCompletionTime: time.Time{},
//	}
//	provider6 := providersTypes.Validator{
//		Address:                 sdk.Address(fakePubKey6.Address()),
//		PublicKey:               (fakePubKey6),
//		Jailed:                  false,
//		Status:                  sdk.Staked,
//		Chains:                  []string{ethereum},
//		ServiceUrl:              "https://www.google.com:443",
//		StakedTokens:            sdk.NewInt(100000),
//		UnstakingCompletionTime: time.Time{},
//	}
//	provider7 := providersTypes.Validator{
//		Address:                 sdk.Address(fakePubKey7.Address()),
//		PublicKey:               (fakePubKey7),
//		Jailed:                  false,
//		Status:                  sdk.Staked,
//		Chains:                  []string{ethereum},
//		ServiceUrl:              "https://www.google.com:443",
//		StakedTokens:            sdk.NewInt(100000),
//		UnstakingCompletionTime: time.Time{},
//	}
//	provider8 := providersTypes.Validator{
//		Address:                 sdk.Address(fakePubKey8.Address()),
//		PublicKey:               (fakePubKey8),
//		Jailed:                  false,
//		Status:                  sdk.Staked,
//		Chains:                  []string{ethereum},
//		ServiceUrl:              "https://www.google.com:443",
//		StakedTokens:            sdk.NewInt(100000),
//		UnstakingCompletionTime: time.Time{},
//	}
//	provider9 := providersTypes.Validator{
//		Address:                 sdk.Address(fakePubKey9.Address()),
//		PublicKey:               (fakePubKey9),
//		Jailed:                  false,
//		Status:                  sdk.Staked,
//		Chains:                  []string{ethereum},
//		ServiceUrl:              "https://www.google.com:443",
//		StakedTokens:            sdk.NewInt(100000),
//		UnstakingCompletionTime: time.Time{},
//	}
//	provider10 := providersTypes.Validator{
//		Address:                 sdk.Address(fakePubKey10.Address()),
//		PublicKey:               (fakePubKey10),
//		Jailed:                  false,
//		Status:                  sdk.Staked,
//		Chains:                  []string{ethereum},
//		ServiceUrl:              "https://www.google.com:443",
//		StakedTokens:            sdk.NewInt(100000),
//		UnstakingCompletionTime: time.Time{},
//	}
//	provider11 := providersTypes.Validator{
//		Address:                 sdk.Address(fakePubKey11.Address()),
//		PublicKey:               (fakePubKey11),
//		Jailed:                  false,
//		Status:                  sdk.Staked,
//		Chains:                  []string{ethereum},
//		ServiceUrl:              "https://www.google.com:443",
//		StakedTokens:            sdk.NewInt(100000),
//		UnstakingCompletionTime: time.Time{},
//	}
//	allProviders = make([]exported.ValidatorI, 12)
//	allProviders[0] = provider12
//	allProviders[1] = provider1
//	allProviders[2] = provider2
//	allProviders[3] = provider3
//	allProviders[4] = provider4
//	allProviders[5] = provider5
//	allProviders[6] = provider6
//	allProviders[7] = provider7
//	allProviders[8] = provider8
//	allProviders[9] = provider9
//	allProviders[10] = provider10
//	allProviders[11] = provider11
//	k := MockPosKeeper{Validators: allProviders}
//	sessionProviders, err := NewSessionProviders(newContext(t, false).WithPlatformVersion("0.0.0"), newContext(t, false).WithPlatformVersion("0.0.0"), k, ethereum, fakeSessionKey, 5)
//	assert.Nil(t, err)
//	assert.Len(t, sessionProviders, 5)
//	assert.Contains(t, sessionProviders, allProviders[0].(providersTypes.Validator))
//	assert.Contains(t, sessionProviders, allProviders[1].(providersTypes.Validator))
//	assert.NotContains(t, sessionProviders, allProviders[2].(providersTypes.Validator))
//	assert.NotContains(t, sessionProviders, allProviders[3].(providersTypes.Validator))
//	assert.Contains(t, sessionProviders, allProviders[4].(providersTypes.Validator))
//	assert.NotContains(t, sessionProviders, allProviders[5].(providersTypes.Validator))
//	assert.NotContains(t, sessionProviders, allProviders[6].(providersTypes.Validator))
//	assert.Contains(t, sessionProviders, allProviders[7].(providersTypes.Validator))
//	assert.Contains(t, sessionProviders, allProviders[8].(providersTypes.Validator))
//	assert.NotContains(t, sessionProviders, allProviders[9].(providersTypes.Validator))
//	assert.NotContains(t, sessionProviders, allProviders[10].(providersTypes.Validator))
//	assert.NotContains(t, sessionProviders, allProviders[11].(providersTypes.Validator))
//	assert.True(t, sessionProviders.Contains(provider12))
//	assert.True(t, sessionProviders.Contains(provider8))
//	assert.True(t, sessionProviders.Contains(provider7))
//	assert.True(t, sessionProviders.Contains(provider4))
//	assert.True(t, sessionProviders.Contains(provider1))
//	assert.False(t, sessionProviders.Contains(provider2))
//	assert.Nil(t, sessionProviders.Validate(5))
//	assert.NotNil(t, SessionProviders(make([]exported.ValidatorI, 5)).Validate(5))
//}
