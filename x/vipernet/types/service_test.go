package types

import (
	"encoding/hex"
	"reflect"
	"testing"
	"time"

	"github.com/vipernet-xyz/viper-network/codec"
	types2 "github.com/vipernet-xyz/viper-network/codec/types"
	"github.com/vipernet-xyz/viper-network/crypto"
	"github.com/vipernet-xyz/viper-network/x/authentication"
	"github.com/vipernet-xyz/viper-network/x/governance"
	exported2 "github.com/vipernet-xyz/viper-network/x/platforms/exported"

	sdk "github.com/vipernet-xyz/viper-network/types"
	platformsType "github.com/vipernet-xyz/viper-network/x/platforms/types"
	"github.com/vipernet-xyz/viper-network/x/providers/exported"
	providersTypes "github.com/vipernet-xyz/viper-network/x/providers/types"

	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestRelay_Validate(t *testing.T) { // TODO add overservice, and not unique relay here
	clientPrivateKey := GetRandomPrivateKey()
	clientPubKey := clientPrivateKey.PublicKey().RawString()
	platformPrivateKey := GetRandomPrivateKey()
	platformPubKey := platformPrivateKey.PublicKey().RawString()
	npk := getRandomPubKey()
	providerPubKey := npk.RawString()
	ethereum := hex.EncodeToString([]byte{01})
	bitcoin := hex.EncodeToString([]byte{02})
	p := Payload{
		Data:    "{\"jsonrpc\":\"2.0\",\"method\":\"web3_clientVersion\",\"params\":[],\"id\":67}",
		Method:  "",
		Path:    "",
		Headers: nil,
	}
	validRelay := Relay{
		Payload: p,
		Meta:    RelayMeta{BlockHeight: 1},
		Proof: RelayProof{
			Entropy:            1,
			SessionBlockHeight: 1,
			ServicerPubKey:     providerPubKey,
			Blockchain:         ethereum,
			Token: AAT{
				Version:           "0.0.1",
				PlatformPublicKey: platformPubKey,
				ClientPublicKey:   clientPubKey,
				PlatformSignature: "",
			},
			Signature: "",
		},
	}
	validRelay.Proof.RequestHash = validRelay.RequestHashString()
	platformSig, er := platformPrivateKey.Sign(validRelay.Proof.Token.Hash())
	if er != nil {
		t.Fatalf(er.Error())
	}
	validRelay.Proof.Token.PlatformSignature = hex.EncodeToString(platformSig)
	clientSig, er := clientPrivateKey.Sign(validRelay.Proof.Hash())
	if er != nil {
		t.Fatalf(er.Error())
	}
	validRelay.Proof.Signature = hex.EncodeToString(clientSig)
	// invalid payload empty data and method
	invalidPayloadEmpty := validRelay
	invalidPayloadEmpty.Payload.Data = ""
	selfNode := providersTypes.Validator{
		Address:                 sdk.Address(npk.Address()),
		PublicKey:               npk,
		Jailed:                  false,
		Status:                  sdk.Staked,
		Chains:                  []string{ethereum, bitcoin},
		ServiceURL:              "https://www.google.com:443",
		StakedTokens:            sdk.NewInt(100000),
		UnstakingCompletionTime: time.Time{},
	}
	var noEthereumProviders []exported.ValidatorI
	for i := 0; i < 4; i++ {
		pubKey := getRandomPubKey()
		noEthereumProviders = append(noEthereumProviders, providersTypes.Validator{
			Address:                 sdk.Address(pubKey.Address()),
			PublicKey:               pubKey,
			Jailed:                  false,
			Status:                  sdk.Staked,
			Chains:                  []string{bitcoin},
			ServiceURL:              "https://www.google.com:443",
			StakedTokens:            sdk.NewInt(100000),
			UnstakingCompletionTime: time.Time{},
		})
	}
	noEthereumProviders = append(noEthereumProviders, selfNode)
	hb := HostedBlockchains{
		M: map[string]HostedBlockchain{ethereum: {
			ID:  ethereum,
			URL: "https://www.google.com:443",
		}},
	}
	pubKey := getRandomPubKey()
	platform := platformsType.Platform{
		Address:                 sdk.Address(pubKey.Address()),
		PublicKey:               pubKey,
		Jailed:                  false,
		Status:                  sdk.Staked,
		Chains:                  []string{ethereum},
		StakedTokens:            sdk.NewInt(1000),
		MaxRelays:               sdk.NewInt(1000),
		UnstakingCompletionTime: time.Time{},
	}
	tests := []struct {
		name         string
		relay        Relay
		provider     providersTypes.Validator
		platform     platformsType.Platform
		allProviders []exported.ValidatorI
		hb           *HostedBlockchains
		hasError     bool
	}{
		{
			name:         "invalid relay: not enough service providers",
			relay:        validRelay,
			provider:     selfNode,
			platform:     platform,
			allProviders: noEthereumProviders,
			hb:           &hb,
			hasError:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			k := MockPosKeeper{Validators: tt.allProviders}
			k2 := MockPlatformsKeeper{Platforms: []exported2.PlatformI{tt.platform}}
			k3 := MockViperKeeper{}
			_, err := tt.relay.Validate(newContext(t, false).WithAppVersion("0.0.0"), k, k2, k3, tt.provider.Address, tt.hb, 1)
			assert.Equal(t, err != nil, tt.hasError)
		})
		ClearSessionCache()
	}
}

func TestRelay_Execute(t *testing.T) {
	clientPrivateKey := GetRandomPrivateKey()
	clientPubKey := clientPrivateKey.PublicKey().RawString()
	platformPrivateKey := GetRandomPrivateKey()
	platformPubKey := platformPrivateKey.PublicKey().RawString()
	npk := getRandomPubKey()
	providerPubKey := npk.RawString()
	ethereum := hex.EncodeToString([]byte{01})
	p := Payload{
		Data:    "foo",
		Method:  "POST",
		Path:    "",
		Headers: nil,
	}
	validRelay := Relay{
		Payload: p,
		Proof: RelayProof{
			Entropy:            1,
			SessionBlockHeight: 1,
			ServicerPubKey:     providerPubKey,
			Blockchain:         ethereum,
			Token: AAT{
				Version:           "0.0.1",
				PlatformPublicKey: platformPubKey,
				ClientPublicKey:   clientPubKey,
				PlatformSignature: "",
			},
			Signature: "",
		},
	}
	validRelay.Proof.RequestHash = validRelay.RequestHashString()
	defer gock.Off() // Flush pending mocks after test execution

	gock.New("https://server.com").
		Post("/relay").
		Reply(200).
		BodyString("bar")

	hb := HostedBlockchains{
		M: map[string]HostedBlockchain{ethereum: {
			ID:  ethereum,
			URL: "https://server.com/relay/",
		}},
	}
	response, err := validRelay.Execute(&hb)
	assert.True(t, err == nil)
	assert.Equal(t, response, "bar")
}

func TestRelay_HandleProof(t *testing.T) {
	clientPrivateKey := GetRandomPrivateKey()
	clientPubKey := clientPrivateKey.PublicKey().RawString()
	platformPrivateKey := GetRandomPrivateKey()
	platformPubKey := platformPrivateKey.PublicKey().RawString()
	npk := getRandomPubKey()
	providerPubKey := npk.RawString()
	ethereum := hex.EncodeToString([]byte{01})
	p := Payload{
		Data:    "foo",
		Method:  "POST",
		Path:    "",
		Headers: nil,
	}
	validRelay := Relay{
		Payload: p,
		Proof: RelayProof{
			Entropy:            1,
			SessionBlockHeight: 1,
			ServicerPubKey:     providerPubKey,
			Blockchain:         ethereum,
			Token: AAT{
				Version:           "0.0.1",
				PlatformPublicKey: platformPubKey,
				ClientPublicKey:   clientPubKey,
				PlatformSignature: "",
			},
			Signature: "",
		},
	}
	validRelay.Proof.RequestHash = validRelay.RequestHashString()
	validRelay.Proof.Store(sdk.NewInt(100000))
	res := GetProof(SessionHeader{
		PlatformPubKey:     platformPubKey,
		Chain:              ethereum,
		SessionBlockHeight: 1,
	}, RelayEvidence, 0)
	assert.True(t, reflect.DeepEqual(validRelay.Proof, res))
}

func TestRelayResponse_BytesAndHash(t *testing.T) {
	providerPrivKey := GetRandomPrivateKey()
	providerPubKey := providerPrivKey.PublicKey().RawString()
	platformPrivKey := GetRandomPrivateKey()
	platformPublicKey := platformPrivKey.PublicKey().RawString()
	cliPrivKey := GetRandomPrivateKey()
	cliPublicKey := cliPrivKey.PublicKey().RawString()
	relayResp := RelayResponse{
		Signature: "",
		Response:  "foo",
		Proof: RelayProof{
			Entropy:            230942034,
			SessionBlockHeight: 1,
			RequestHash:        providerPubKey,
			ServicerPubKey:     providerPubKey,
			Blockchain:         hex.EncodeToString(merkleHash([]byte("foo"))),
			Token: AAT{
				Version:           "0.0.1",
				PlatformPublicKey: platformPublicKey,
				ClientPublicKey:   cliPublicKey,
				PlatformSignature: "",
			},
			Signature: "",
		},
	}
	platformSig, err := platformPrivKey.Sign(relayResp.Proof.Token.Hash())
	if err != nil {
		t.Fatalf(err.Error())
	}
	relayResp.Proof.Token.PlatformSignature = hex.EncodeToString(platformSig)
	assert.NotNil(t, relayResp.Hash())
	assert.Equal(t, hex.EncodeToString(relayResp.Hash()), relayResp.HashString())
	storedHashString := relayResp.HashString()
	providerSig, err := providerPrivKey.Sign(relayResp.Hash())
	if err != nil {
		t.Fatalf(err.Error())
	}
	relayResp.Signature = hex.EncodeToString(providerSig)
	assert.Equal(t, storedHashString, relayResp.HashString())
}

func TestSortJSON(t *testing.T) {
	// out of order json arrays
	j1 := `{"foo":0,"bar":1}`
	j2 := `{"bar":1,"foo":0}`
	// sort
	objs := sortJSONResponse(j1)
	objs2 := sortJSONResponse(j2)
	// compare
	assert.Equal(t, objs, objs2)
}

type MockValidatorI interface {
	IsStaked() bool                 // check if has a staked status
	IsUnstaked() bool               // check if has status unstaked
	IsUnstaking() bool              // check if has status unstaking
	IsJailed() bool                 // whether the validator is jailed
	GetStatus() sdk.StakeStatus     // status of the validator
	GetAddress() sdk.Address        // operator address to receive/return validators coins
	GetPublicKey() crypto.PublicKey // validation consensus pubkey
	GetTokens() sdk.BigInt          // validation tokens
	GetConsensusPower() int64       // validation power in tendermint
	GetChains() []string
}

type MockPlatformsKeeper struct {
	Platforms []exported2.PlatformI
}

func (m MockPlatformsKeeper) GetStakedTokens(ctx sdk.Ctx) sdk.BigInt {
	panic("implement me")
}

func (m MockPlatformsKeeper) Platform(ctx sdk.Ctx, addr sdk.Address) exported2.PlatformI {
	for _, v := range m.Platforms {
		if v.GetAddress().Equals(addr) {
			return v
		}
	}
	return nil
}

func (m MockPlatformsKeeper) AllPlatforms(ctx sdk.Ctx) (platformlications []exported2.PlatformI) {
	panic("implement me")
}

func (m MockPlatformsKeeper) TotalTokens(ctx sdk.Ctx) sdk.BigInt {
	panic("implement me")
}

func (m MockPlatformsKeeper) JailPlatform(ctx sdk.Ctx, addr sdk.Address) {
	panic("implement me")
}

type MockPosKeeper struct {
	Validators []exported.ValidatorI
}

type MockViperKeeper struct{}

func (m MockViperKeeper) Codec() *codec.Codec {
	return makeTestCodec()
}

func (m MockViperKeeper) SessionNodeCount(ctx sdk.Ctx) (res int64) {
	return 5
}

func (m MockPosKeeper) GetValidatorsByChain(ctx sdk.Ctx, networkID string) (validators []sdk.Address, total int) {
	for _, v := range m.Validators {
		s := v.(MockValidatorI)
		chains := s.GetChains()
		for _, c := range chains {
			if c == networkID {
				total++
				validators = append(validators, v.GetAddress())
			}
		}
	}
	return
}

func (m MockPosKeeper) RewardForRelays(ctx sdk.Ctx, relays sdk.BigInt, address sdk.Address, platformAdddress sdk.Address) sdk.BigInt {
	panic("implement me")
}

func (m MockPosKeeper) GetStakedTokens(ctx sdk.Ctx) sdk.BigInt {
	panic("implement me")
}

func (m MockPosKeeper) Validator(ctx sdk.Ctx, addr sdk.Address) exported.ValidatorI {
	for _, v := range m.Validators {
		if addr.Equals(v.GetAddress()) {
			return v
		}
	}
	return nil
}

func (m MockPosKeeper) TotalTokens(ctx sdk.Ctx) sdk.BigInt {
	panic("implement me")
}

func (m MockPosKeeper) BurnForChallenge(ctx sdk.Ctx, challenges sdk.BigInt, address sdk.Address) {
	panic("implement me")
}

func (m MockPosKeeper) JailValidator(ctx sdk.Ctx, addr sdk.Address) {
	panic("implement me")
}

func (m MockPosKeeper) AllValidators(ctx sdk.Ctx) (validators []exported.ValidatorI) {
	return m.Validators
}

func (m MockPosKeeper) GetStakedValidators(ctx sdk.Ctx) (validators []exported.ValidatorI) {
	return m.Validators
}

func (m MockPosKeeper) BlocksPerSession(ctx sdk.Ctx) (res int64) {
	panic("implement me")
}

func (m MockPosKeeper) StakeDenom(ctx sdk.Ctx) (res string) {
	panic("implement me")
}

func makeTestCodec() *codec.Codec {
	var cdc = codec.NewCodec(types2.NewInterfaceRegistry())
	authentication.RegisterCodec(cdc)
	governance.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	crypto.RegisterAmino(cdc.AminoCodec().Amino)
	return cdc
}
