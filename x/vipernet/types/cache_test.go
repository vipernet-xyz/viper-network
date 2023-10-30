package types

import (
	"encoding/hex"
	"os"
	"reflect"
	"testing"

	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/vipernet-xyz/viper-network/types"

	"github.com/stretchr/testify/assert"
)

func InitCacheTest() {
	logger := log.NewNopLogger()
	testingConfig := sdk.DefaultTestingViperConfig()
	CleanViperNodes()
	AddViperNode(GetRandomPrivateKey(), log.NewNopLogger())
	InitConfig(&HostedBlockchains{
		M: make(map[string]HostedBlockchain),
	}, &HostedGeoZones{
		M: make(map[string]GeoZone),
	}, logger, testingConfig)
}

func TestMain(m *testing.M) {
	InitCacheTest()
	m.Run()
	err := os.RemoveAll("data")
	if err != nil {
		panic(err)
	}
	os.Exit(0)
}

func TestIsUniqueProof(t *testing.T) {
	h := SessionHeader{
		ProviderPubKey:     "0",
		Chain:              "0001",
		SessionBlockHeight: 0,
	}
	e, _ := GetEvidence(h, RelayEvidence, sdk.NewInt(100000), GlobalEvidenceCache)
	p := RelayProof{
		Entropy: 1,
	}
	p1 := RelayProof{
		Entropy: 2,
	}
	assert.True(t, IsUniqueProof(p, e), "p is unique")
	e.AddProof(p)
	SetEvidence(e, GlobalEvidenceCache)
	e, err := GetEvidence(h, RelayEvidence, sdk.ZeroInt(), GlobalEvidenceCache)
	assert.Nil(t, err)
	assert.False(t, IsUniqueProof(p, e), "p is no longer unique")
	assert.True(t, IsUniqueProof(p1, e), "p is unique")
}

func TestAllEvidence_AddGetEvidence(t *testing.T) {
	appPubKey := getRandomPubKey().RawString()
	servicerPubKey := getRandomPubKey().RawString()
	clientPubKey := getRandomPubKey().RawString()
	ethereum := hex.EncodeToString([]byte{0001})
	header := SessionHeader{
		ProviderPubKey:     appPubKey,
		Chain:              ethereum,
		SessionBlockHeight: 1,
	}
	proof := RelayProof{
		Entropy:            0,
		RequestHash:        header.HashString(), // fake
		SessionBlockHeight: 1,
		ServicerPubKey:     servicerPubKey,
		Blockchain:         ethereum,
		Token: AAT{
			Version:           "0.0.1",
			ProviderPublicKey: appPubKey,
			ClientPublicKey:   clientPubKey,
			ProviderSignature: "",
		},
		Signature: "",
	}
	SetProof(header, RelayEvidence, proof, sdk.NewInt(100000), GlobalEvidenceCache)
	assert.True(t, reflect.DeepEqual(GetProof(header, RelayEvidence, 0, GlobalEvidenceCache), proof))
}

func TestAllEvidence_Iterator(t *testing.T) {
	ClearEvidence(GlobalEvidenceCache)
	appPubKey := getRandomPubKey().RawString()
	servicerPubKey := getRandomPubKey().RawString()
	clientPubKey := getRandomPubKey().RawString()
	ethereum := hex.EncodeToString([]byte{0001})
	header := SessionHeader{
		ProviderPubKey:     appPubKey,
		Chain:              ethereum,
		SessionBlockHeight: 1,
	}
	proof := RelayProof{
		Entropy:            0,
		RequestHash:        header.HashString(), // fake
		SessionBlockHeight: 1,
		ServicerPubKey:     servicerPubKey,
		Blockchain:         ethereum,
		Token: AAT{
			Version:           "0.0.1",
			ProviderPublicKey: appPubKey,
			ClientPublicKey:   clientPubKey,
			ProviderSignature: "",
		},
		Signature: "",
	}
	SetProof(header, RelayEvidence, proof, sdk.NewInt(100000), GlobalEvidenceCache)
	iter := EvidenceIterator(GlobalEvidenceCache)
	var count = 0
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		count++
	}
	assert.Equal(t, 1, int(count))
}

func TestAllEvidence_DeleteEvidence(t *testing.T) {
	appPubKey := getRandomPubKey().RawString()
	servicerPubKey := getRandomPubKey().RawString()
	clientPubKey := getRandomPubKey().RawString()
	ethereum := hex.EncodeToString([]byte{0001})
	header := SessionHeader{
		ProviderPubKey:     appPubKey,
		Chain:              ethereum,
		SessionBlockHeight: 1,
	}
	proof := RelayProof{
		Entropy:            0,
		SessionBlockHeight: 1,
		ServicerPubKey:     servicerPubKey,
		RequestHash:        header.HashString(), // fake
		Blockchain:         ethereum,
		Token: AAT{
			Version:           "0.0.1",
			ProviderPublicKey: appPubKey,
			ClientPublicKey:   clientPubKey,
			ProviderSignature: "",
		},
		Signature: "",
	}
	SetProof(header, RelayEvidence, proof, sdk.NewInt(100000), GlobalEvidenceCache)
	assert.True(t, reflect.DeepEqual(GetProof(header, RelayEvidence, 0, GlobalEvidenceCache), proof))
	GetProof(header, RelayEvidence, 0, GlobalEvidenceCache)
	_ = DeleteEvidence(header, RelayEvidence, GlobalEvidenceCache)
	assert.Empty(t, GetProof(header, RelayEvidence, 0, GlobalEvidenceCache))
}

func TestAllEvidence_GetTotalProofs(t *testing.T) {
	appPubKey := getRandomPubKey().RawString()
	servicerPubKey := getRandomPubKey().RawString()
	clientPubKey := getRandomPubKey().RawString()
	ethereum := hex.EncodeToString([]byte{0001})
	header := SessionHeader{
		ProviderPubKey:     appPubKey,
		Chain:              ethereum,
		SessionBlockHeight: 1,
	}
	header2 := SessionHeader{
		ProviderPubKey:     appPubKey,
		Chain:              ethereum,
		SessionBlockHeight: 101,
	}
	proof := RelayProof{
		Entropy:            0,
		SessionBlockHeight: 1,
		ServicerPubKey:     servicerPubKey,
		RequestHash:        header.HashString(), // fake
		Blockchain:         ethereum,
		Token: AAT{
			Version:           "0.0.1",
			ProviderPublicKey: appPubKey,
			ClientPublicKey:   clientPubKey,
			ProviderSignature: "",
		},
		Signature: "",
	}
	proof2 := RelayProof{
		Entropy:            0,
		SessionBlockHeight: 1,
		ServicerPubKey:     servicerPubKey,
		RequestHash:        header.HashString(), // fake
		Blockchain:         ethereum,
		Token: AAT{
			Version:           "0.0.1",
			ProviderPublicKey: appPubKey,
			ClientPublicKey:   clientPubKey,
			ProviderSignature: "",
		},
		Signature: "",
	}
	SetProof(header, RelayEvidence, proof, sdk.NewInt(100000), GlobalEvidenceCache)
	SetProof(header, RelayEvidence, proof2, sdk.NewInt(100000), GlobalEvidenceCache)
	SetProof(header2, RelayEvidence, proof2, sdk.NewInt(100000), GlobalEvidenceCache) // different header so shouldn't be counted
	_, totalRelays := GetTotalProofs(header, RelayEvidence, sdk.NewInt(100000), GlobalEvidenceCache)
	assert.Equal(t, totalRelays, int64(2))
}

func TestSetGetSession(t *testing.T) {
	session := NewTestSession(t, hex.EncodeToString(Hash([]byte("foo"))), hex.EncodeToString(Hash([]byte("ind"))))
	session2 := NewTestSession(t, hex.EncodeToString(Hash([]byte("bar"))), hex.EncodeToString(Hash([]byte("ind"))))
	SetSession(session, GlobalSessionCache)
	s, found := GetSession(session.SessionHeader, GlobalSessionCache)
	assert.True(t, found)
	assert.Equal(t, s, session)
	_, found = GetSession(session2.SessionHeader, GlobalSessionCache)
	assert.False(t, found)
	SetSession(session2, GlobalSessionCache)
	s, found = GetSession(session2.SessionHeader, GlobalSessionCache)
	assert.True(t, found)
	assert.Equal(t, s, session2)
}

func TestDeleteSession(t *testing.T) {
	session := NewTestSession(t, hex.EncodeToString(Hash([]byte("foo"))), hex.EncodeToString(Hash([]byte("ind"))))
	SetSession(session, GlobalSessionCache)
	DeleteSession(session.SessionHeader, GlobalSessionCache)
	_, found := GetSession(session.SessionHeader, GlobalSessionCache)
	assert.False(t, found)
}

func TestClearCache(t *testing.T) {
	session := NewTestSession(t, hex.EncodeToString(Hash([]byte("foo"))), hex.EncodeToString(Hash([]byte("ind"))))
	SetSession(session, GlobalSessionCache)
	ClearSessionCache(GlobalSessionCache)
	iter := SessionIterator(GlobalSessionCache)
	var count = 0
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		count++
	}
	assert.Zero(t, count)
}

func NewTestSession(t *testing.T, chain string, geoZone string) Session {
	appPubKey := getRandomPubKey()
	var vals []sdk.Address
	for i := 0; i < 5; i++ {
		nodePubKey := getRandomPubKey()
		vals = append(vals, sdk.Address(nodePubKey.Address()))
	}
	fisherman := append(vals, sdk.Address(getRandomPubKey().Address()))
	return Session{
		SessionHeader: SessionHeader{
			ProviderPubKey:     appPubKey.RawString(),
			Chain:              chain,
			SessionBlockHeight: 1,
			GeoZone:            geoZone,
			NumServicers:       5,
		},
		SessionKey:       appPubKey.RawBytes(),
		SessionServicers: vals,
		SessionFishermen: fisherman,
	}
}
