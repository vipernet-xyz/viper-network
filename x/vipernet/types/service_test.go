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
	exported2 "github.com/vipernet-xyz/viper-network/x/providers/exported"

	sdk "github.com/vipernet-xyz/viper-network/types"
	providersType "github.com/vipernet-xyz/viper-network/x/providers/types"
	"github.com/vipernet-xyz/viper-network/x/servicers/exported"
	servicersTypes "github.com/vipernet-xyz/viper-network/x/servicers/types"

	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestRelay_Validate(t *testing.T) { // TODO add overservice, and not unique relay here
	clientPrivateKey := GetRandomPrivateKey()
	clientPubKey := clientPrivateKey.PublicKey().RawString()
	providerPrivateKey := GetRandomPrivateKey()
	providerPubKey := providerPrivateKey.PublicKey().RawString()
	npk := getRandomPubKey()
	servicerPubKey := npk.RawString()
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
			ServicerPubKey:     servicerPubKey,
			Blockchain:         ethereum,
			Token: AAT{
				Version:           "0.0.1",
				ProviderPublicKey: providerPubKey,
				ClientPublicKey:   clientPubKey,
				ProviderSignature: "",
			},
			Signature: "",
		},
	}
	validRelay.Proof.RequestHash = validRelay.RequestHashString()
	providerSig, er := providerPrivateKey.Sign(validRelay.Proof.Token.Hash())
	if er != nil {
		t.Fatalf(er.Error())
	}
	validRelay.Proof.Token.ProviderSignature = hex.EncodeToString(providerSig)
	clientSig, er := clientPrivateKey.Sign(validRelay.Proof.Hash())
	if er != nil {
		t.Fatalf(er.Error())
	}
	validRelay.Proof.Signature = hex.EncodeToString(clientSig)
	// invalid payload empty data and method
	invalidPayloadEmpty := validRelay
	invalidPayloadEmpty.Payload.Data = ""
	selfNode := servicersTypes.Validator{
		Address:                 sdk.Address(npk.Address()),
		PublicKey:               npk,
		Jailed:                  false,
		Status:                  sdk.Staked,
		Chains:                  []string{ethereum, bitcoin},
		ServiceURL:              "https://www.google.com:443",
		StakedTokens:            sdk.NewInt(100000),
		UnstakingCompletionTime: time.Time{},
	}
	var noEthereumServicers []exported.ValidatorI
	for i := 0; i < 4; i++ {
		pubKey := getRandomPubKey()
		noEthereumServicers = append(noEthereumServicers, servicersTypes.Validator{
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
	noEthereumServicers = append(noEthereumServicers, selfNode)
	hb := HostedBlockchains{
		M: map[string]HostedBlockchain{ethereum: {
			ID:  ethereum,
			URL: "https://www.google.com:443",
		}},
	}
	pubKey := getRandomPubKey()
	provider := providersType.Provider{
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
		servicer     servicersTypes.Validator
		provider     providersType.Provider
		allServicers []exported.ValidatorI
		hb           *HostedBlockchains
		hasError     bool
	}{
		{
			name:         "invalid relay: not enough service servicers",
			relay:        validRelay,
			servicer:     selfNode,
			provider:     provider,
			allServicers: noEthereumServicers,
			hb:           &hb,
			hasError:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			k := MockPosKeeper{Validators: tt.allServicers}
			k2 := MockProvidersKeeper{Providers: []exported2.ProviderI{tt.provider}}
			k3 := MockViperKeeper{}
			_, err := tt.relay.Validate(newContext(t, false).WithAppVersion("0.0.0"), k, k2, k3, tt.servicer.Address, tt.hb, 1)
			assert.Equal(t, err != nil, tt.hasError)
		})
		ClearSessionCache()
	}
}

func TestRelay_Execute(t *testing.T) {
	clientPrivateKey := GetRandomPrivateKey()
	clientPubKey := clientPrivateKey.PublicKey().RawString()
	providerPrivateKey := GetRandomPrivateKey()
	providerPubKey := providerPrivateKey.PublicKey().RawString()
	npk := getRandomPubKey()
	servicerPubKey := npk.RawString()
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
			ServicerPubKey:     servicerPubKey,
			Blockchain:         ethereum,
			Token: AAT{
				Version:           "0.0.1",
				ProviderPublicKey: providerPubKey,
				ClientPublicKey:   clientPubKey,
				ProviderSignature: "",
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
	providerPrivateKey := GetRandomPrivateKey()
	providerPubKey := providerPrivateKey.PublicKey().RawString()
	npk := getRandomPubKey()
	servicerPubKey := npk.RawString()
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
			ServicerPubKey:     servicerPubKey,
			Blockchain:         ethereum,
			Token: AAT{
				Version:           "0.0.1",
				ProviderPublicKey: providerPubKey,
				ClientPublicKey:   clientPubKey,
				ProviderSignature: "",
			},
			Signature: "",
		},
	}
	validRelay.Proof.RequestHash = validRelay.RequestHashString()
	validRelay.Proof.Store(sdk.NewInt(100000))
	res := GetProof(SessionHeader{
		ProviderPubKey:     providerPubKey,
		Chain:              ethereum,
		SessionBlockHeight: 1,
	}, RelayEvidence, 0)
	assert.True(t, reflect.DeepEqual(validRelay.Proof, res))
}

func TestRelayResponse_BytesAndHash(t *testing.T) {
	servicerPrivKey := GetRandomPrivateKey()
	servicerPubKey := servicerPrivKey.PublicKey().RawString()
	providerPrivKey := GetRandomPrivateKey()
	providerPublicKey := providerPrivKey.PublicKey().RawString()
	cliPrivKey := GetRandomPrivateKey()
	cliPublicKey := cliPrivKey.PublicKey().RawString()
	relayResp := RelayResponse{
		Signature: "",
		Response:  "foo",
		Proof: RelayProof{
			Entropy:            230942034,
			SessionBlockHeight: 1,
			RequestHash:        servicerPubKey,
			ServicerPubKey:     servicerPubKey,
			Blockchain:         hex.EncodeToString(merkleHash([]byte("foo"))),
			Token: AAT{
				Version:           "0.0.1",
				ProviderPublicKey: providerPublicKey,
				ClientPublicKey:   cliPublicKey,
				ProviderSignature: "",
			},
			Signature: "",
		},
	}
	providerSig, err := providerPrivKey.Sign(relayResp.Proof.Token.Hash())
	if err != nil {
		t.Fatalf(err.Error())
	}
	relayResp.Proof.Token.ProviderSignature = hex.EncodeToString(providerSig)
	assert.NotNil(t, relayResp.Hash())
	assert.Equal(t, hex.EncodeToString(relayResp.Hash()), relayResp.HashString())
	storedHashString := relayResp.HashString()
	servicerSig, err := servicerPrivKey.Sign(relayResp.Hash())
	if err != nil {
		t.Fatalf(err.Error())
	}
	relayResp.Signature = hex.EncodeToString(servicerSig)
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

type MockProvidersKeeper struct {
	Providers []exported2.ProviderI
}

func (m MockProvidersKeeper) GetStakedTokens(ctx sdk.Ctx) sdk.BigInt {
	panic("implement me")
}

func (m MockProvidersKeeper) Provider(ctx sdk.Ctx, addr sdk.Address) exported2.ProviderI {
	for _, v := range m.Providers {
		if v.GetAddress().Equals(addr) {
			return v
		}
	}
	return nil
}

func (m MockProvidersKeeper) AllProviders(ctx sdk.Ctx) (providers []exported2.ProviderI) {
	panic("implement me")
}

func (m MockProvidersKeeper) TotalTokens(ctx sdk.Ctx) sdk.BigInt {
	panic("implement me")
}

func (m MockProvidersKeeper) JailProvider(ctx sdk.Ctx, addr sdk.Address) {
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

func (m MockPosKeeper) RewardForRelays(ctx sdk.Ctx, relays sdk.BigInt, address sdk.Address, providerAdddress sdk.Address) sdk.BigInt {
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
