package types

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSessionKey(t *testing.T) {
	providerPubKey := getRandomPubKey()
	ctx := newContext(t, false).WithAppVersion("0.0.0")
	blockhash := hex.EncodeToString(ctx.BlockHeader().LastBlockId.Hash)
	ethereum := hex.EncodeToString([]byte{01})
	bitcoin := hex.EncodeToString([]byte{02})
	key1, err := NewSessionKey(providerPubKey.RawString(), ethereum, blockhash)
	assert.Nil(t, err)
	assert.NotNil(t, key1)
	assert.NotEmpty(t, key1)
	assert.Nil(t, HashVerification(hex.EncodeToString(key1)))
	key2, err := NewSessionKey(providerPubKey.RawString(), bitcoin, blockhash)
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

//func TestNewSessionServicers(t *testing.T) {
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
//	var allServicers []exported.ValidatorI
//	servicer12 := servicersTypes.Validator{
//		Address:                 sdk.Address(fakePubKey12.Address()),
//		PublicKey:               fakePubKey12,
//		Jailed:                  false,
//		Status:                  sdk.Staked,
//		Chains:                  []string{ethereum},
//		ServiceUrl:              "https://www.google.com:443",
//		StakedTokens:            sdk.NewInt(100000),
//		UnstakingCompletionTime: time.Time{},
//	}
//	servicer1 := servicersTypes.Validator{
//		Address:                 sdk.Address(fakePubKey1.Address()),
//		PublicKey:               (fakePubKey1),
//		Jailed:                  false,
//		Status:                  sdk.Staked,
//		Chains:                  []string{ethereum},
//		ServiceUrl:              "https://www.google.com:443",
//		StakedTokens:            sdk.NewInt(100000),
//		UnstakingCompletionTime: time.Time{},
//	}
//	servicer2 := servicersTypes.Validator{
//		Address:                 sdk.Address(fakePubKey2.Address()),
//		PublicKey:               (fakePubKey2),
//		Jailed:                  false,
//		Status:                  sdk.Staked,
//		Chains:                  []string{ethereum},
//		ServiceUrl:              "https://www.google.com:443",
//		StakedTokens:            sdk.NewInt(100000),
//		UnstakingCompletionTime: time.Time{},
//	}
//	servicer3 := servicersTypes.Validator{
//		Address:                 sdk.Address(fakePubKey3.Address()),
//		PublicKey:               (fakePubKey3),
//		Jailed:                  false,
//		Status:                  sdk.Staked,
//		Chains:                  []string{ethereum},
//		ServiceUrl:              "https://www.google.com:443",
//		StakedTokens:            sdk.NewInt(100000),
//		UnstakingCompletionTime: time.Time{},
//	}
//	servicer4 := servicersTypes.Validator{
//		Address:                 sdk.Address(fakePubKey4.Address()),
//		PublicKey:               (fakePubKey4),
//		Jailed:                  false,
//		Status:                  sdk.Staked,
//		Chains:                  []string{ethereum},
//		ServiceUrl:              "https://www.google.com:443",
//		StakedTokens:            sdk.NewInt(100000),
//		UnstakingCompletionTime: time.Time{},
//	}
//	servicer5 := servicersTypes.Validator{
//		Address:                 sdk.Address(fakePubKey5.Address()),
//		PublicKey:               (fakePubKey5),
//		Jailed:                  false,
//		Status:                  sdk.Staked,
//		Chains:                  []string{ethereum},
//		ServiceUrl:              "https://www.google.com:443",
//		StakedTokens:            sdk.NewInt(100000),
//		UnstakingCompletionTime: time.Time{},
//	}
//	servicer6 := servicersTypes.Validator{
//		Address:                 sdk.Address(fakePubKey6.Address()),
//		PublicKey:               (fakePubKey6),
//		Jailed:                  false,
//		Status:                  sdk.Staked,
//		Chains:                  []string{ethereum},
//		ServiceUrl:              "https://www.google.com:443",
//		StakedTokens:            sdk.NewInt(100000),
//		UnstakingCompletionTime: time.Time{},
//	}
//	servicer7 := servicersTypes.Validator{
//		Address:                 sdk.Address(fakePubKey7.Address()),
//		PublicKey:               (fakePubKey7),
//		Jailed:                  false,
//		Status:                  sdk.Staked,
//		Chains:                  []string{ethereum},
//		ServiceUrl:              "https://www.google.com:443",
//		StakedTokens:            sdk.NewInt(100000),
//		UnstakingCompletionTime: time.Time{},
//	}
//	servicer8 := servicersTypes.Validator{
//		Address:                 sdk.Address(fakePubKey8.Address()),
//		PublicKey:               (fakePubKey8),
//		Jailed:                  false,
//		Status:                  sdk.Staked,
//		Chains:                  []string{ethereum},
//		ServiceUrl:              "https://www.google.com:443",
//		StakedTokens:            sdk.NewInt(100000),
//		UnstakingCompletionTime: time.Time{},
//	}
//	servicer9 := servicersTypes.Validator{
//		Address:                 sdk.Address(fakePubKey9.Address()),
//		PublicKey:               (fakePubKey9),
//		Jailed:                  false,
//		Status:                  sdk.Staked,
//		Chains:                  []string{ethereum},
//		ServiceUrl:              "https://www.google.com:443",
//		StakedTokens:            sdk.NewInt(100000),
//		UnstakingCompletionTime: time.Time{},
//	}
//	servicer10 := servicersTypes.Validator{
//		Address:                 sdk.Address(fakePubKey10.Address()),
//		PublicKey:               (fakePubKey10),
//		Jailed:                  false,
//		Status:                  sdk.Staked,
//		Chains:                  []string{ethereum},
//		ServiceUrl:              "https://www.google.com:443",
//		StakedTokens:            sdk.NewInt(100000),
//		UnstakingCompletionTime: time.Time{},
//	}
//	servicer11 := servicersTypes.Validator{
//		Address:                 sdk.Address(fakePubKey11.Address()),
//		PublicKey:               (fakePubKey11),
//		Jailed:                  false,
//		Status:                  sdk.Staked,
//		Chains:                  []string{ethereum},
//		ServiceUrl:              "https://www.google.com:443",
//		StakedTokens:            sdk.NewInt(100000),
//		UnstakingCompletionTime: time.Time{},
//	}
//	allServicers = make([]exported.ValidatorI, 12)
//	allServicers[0] = servicer12
//	allServicers[1] = servicer1
//	allServicers[2] = servicer2
//	allServicers[3] = servicer3
//	allServicers[4] = servicer4
//	allServicers[5] = servicer5
//	allServicers[6] = servicer6
//	allServicers[7] = servicer7
//	allServicers[8] = servicer8
//	allServicers[9] = servicer9
//	allServicers[10] = servicer10
//	allServicers[11] = servicer11
//	k := MockPosKeeper{Validators: allServicers}
//	sessionServicers, err := NewSessionServicers(newContext(t, false).WithProviderVersion("0.0.0"), newContext(t, false).WithProviderVersion("0.0.0"), k, ethereum, fakeSessionKey, 5)
//	assert.Nil(t, err)
//	assert.Len(t, sessionServicers, 5)
//	assert.Contains(t, sessionServicers, allServicers[0].(servicersTypes.Validator))
//	assert.Contains(t, sessionServicers, allServicers[1].(servicersTypes.Validator))
//	assert.NotContains(t, sessionServicers, allServicers[2].(servicersTypes.Validator))
//	assert.NotContains(t, sessionServicers, allServicers[3].(servicersTypes.Validator))
//	assert.Contains(t, sessionServicers, allServicers[4].(servicersTypes.Validator))
//	assert.NotContains(t, sessionServicers, allServicers[5].(servicersTypes.Validator))
//	assert.NotContains(t, sessionServicers, allServicers[6].(servicersTypes.Validator))
//	assert.Contains(t, sessionServicers, allServicers[7].(servicersTypes.Validator))
//	assert.Contains(t, sessionServicers, allServicers[8].(servicersTypes.Validator))
//	assert.NotContains(t, sessionServicers, allServicers[9].(servicersTypes.Validator))
//	assert.NotContains(t, sessionServicers, allServicers[10].(servicersTypes.Validator))
//	assert.NotContains(t, sessionServicers, allServicers[11].(servicersTypes.Validator))
//	assert.True(t, sessionServicers.Contains(servicer12))
//	assert.True(t, sessionServicers.Contains(servicer8))
//	assert.True(t, sessionServicers.Contains(servicer7))
//	assert.True(t, sessionServicers.Contains(servicer4))
//	assert.True(t, sessionServicers.Contains(servicer1))
//	assert.False(t, sessionServicers.Contains(servicer2))
//	assert.Nil(t, sessionServicers.Validate(5))
//	assert.NotNil(t, SessionServicers(make([]exported.ValidatorI, 5)).Validate(5))
//}
