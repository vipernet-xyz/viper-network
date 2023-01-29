package types

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAAT_VersionIsIncluded(t *testing.T) {
	providerPrivKey := GetRandomPrivateKey()
	clientPrivKey := GetRandomPrivateKey()
	var AATNoVersion = AAT{
		Version:           "",
		ProviderPublicKey: providerPrivKey.PublicKey().RawString(),
		ClientPublicKey:   clientPrivKey.PublicKey().RawString(),
		ProviderSignature: "",
	}
	var AATWithVersion = AAT{
		Version:           "0.0.1",
		ProviderPublicKey: providerPrivKey.PublicKey().RawString(),
		ClientPublicKey:   clientPrivKey.PublicKey().RawString(),
		ProviderSignature: "",
	}
	tests := []struct {
		name     string
		aat      AAT
		expected bool
	}{
		{
			name:     "AAT is missing the version",
			aat:      AATNoVersion,
			expected: false,
		},
		{
			name:     "AAT has the version",
			aat:      AATWithVersion,
			expected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.aat.VersionIsIncluded(), tt.expected)
		})
	}
}

func TestAAT_VersionIsSupported(t *testing.T) {
	providerPrivKey := GetRandomPrivateKey()
	clientPrivKey := GetRandomPrivateKey()
	var AATNotSupportedVersion = AAT{
		Version:           "0.0.11",
		ProviderPublicKey: providerPrivKey.PublicKey().RawString(),
		ClientPublicKey:   clientPrivKey.PublicKey().RawString(),
		ProviderSignature: "",
	}
	var AATSupported = AAT{
		Version:           "0.0.1",
		ProviderPublicKey: providerPrivKey.PublicKey().RawString(),
		ClientPublicKey:   clientPrivKey.PublicKey().RawString(),
		ProviderSignature: "",
	}
	tests := []struct {
		name     string
		aat      AAT
		expected bool
	}{
		{
			name:     "AAT doesn't not have a supported version",
			aat:      AATNotSupportedVersion,
			expected: false,
		},
		{
			name:     "AAT has a supported version",
			aat:      AATSupported,
			expected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.aat.VersionIsSupported(), tt.expected)
		})
	}
}

func TestAAT_ValidateVersion(t *testing.T) {
	providerPrivKey := GetRandomPrivateKey()
	clientPrivKey := GetRandomPrivateKey()
	var AATVersionMissing = AAT{
		Version:           "",
		ProviderPublicKey: providerPrivKey.PublicKey().RawString(),
		ClientPublicKey:   clientPrivKey.PublicKey().RawString(),
		ProviderSignature: "",
	}
	var AATNotSupportedVersion = AAT{
		Version:           "0.0.11",
		ProviderPublicKey: providerPrivKey.PublicKey().RawString(),
		ClientPublicKey:   clientPrivKey.PublicKey().RawString(),
		ProviderSignature: "",
	}
	var AATSupported = AAT{
		Version:           "0.0.1",
		ProviderPublicKey: providerPrivKey.PublicKey().RawString(),
		ClientPublicKey:   clientPrivKey.PublicKey().RawString(),
		ProviderSignature: "",
	}
	tests := []struct {
		name     string
		aat      AAT
		hasError bool
	}{
		{
			name:     "AAT is missing the version",
			aat:      AATVersionMissing,
			hasError: true,
		},
		{
			name:     "AAT doesn't not have a supported version",
			aat:      AATNotSupportedVersion,
			hasError: true,
		},
		{
			name:     "AAT has a supported version",
			aat:      AATSupported,
			hasError: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.aat.ValidateVersion() != nil, tt.hasError)
		})
	}
}

func TestAAT_ValidateMessage(t *testing.T) {
	providerPrivKey := GetRandomPrivateKey()
	clientPubKey := getRandomPubKey()
	var AATInvalidProviderPubKey = AAT{
		Version:           "0.0.1",
		ProviderPublicKey: providerPrivKey.PubKey().Address().String(),
		ClientPublicKey:   clientPubKey.RawString(),
		ProviderSignature: "",
	}
	var AATInvalidClientPubKey = AAT{
		Version:           "0.0.1",
		ProviderPublicKey: providerPrivKey.PublicKey().RawString(),
		ClientPublicKey:   clientPubKey.Address().String(),
		ProviderSignature: "",
	}
	var AATValidMessage = AAT{
		Version:           "0.0.1",
		ProviderPublicKey: providerPrivKey.PublicKey().RawString(),
		ClientPublicKey:   clientPubKey.RawString(),
		ProviderSignature: "",
	}
	tests := []struct {
		name     string
		aat      AAT
		hasError bool
	}{
		{
			name:     "AAT doesn't have a valid provider pub key",
			aat:      AATInvalidProviderPubKey,
			hasError: true,
		},
		{
			name:     "AAT doesn't have a valid client pub key",
			aat:      AATInvalidClientPubKey,
			hasError: true,
		},
		{
			name:     "AAT has a valid message",
			aat:      AATValidMessage,
			hasError: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.aat.ValidateMessage() != nil, tt.hasError)
		})
	}
}

func TestAAT_ValidateSignature(t *testing.T) {
	providerPrivKey := GetRandomPrivateKey()
	clientPrivKey := GetRandomPrivateKey()
	var AATMissingSignature = AAT{
		Version:           "0.0.1",
		ProviderPublicKey: providerPrivKey.PublicKey().RawString(),
		ClientPublicKey:   clientPrivKey.PublicKey().RawString(),
		ProviderSignature: "",
	}
	var AATInvalidSignature = AAT{
		Version:           "0.0.1",
		ProviderPublicKey: providerPrivKey.PublicKey().RawString(),
		ClientPublicKey:   clientPrivKey.PublicKey().RawString(),
		ProviderSignature: "",
	}
	// sign with the client (invalid)
	clientSignature, err := clientPrivKey.Sign(AATInvalidSignature.Hash())
	if err != nil {
		t.Fatalf(err.Error())
	}
	AATInvalidSignature.ProviderSignature = hex.EncodeToString(clientSignature)
	// sign with the providerlication
	var AATValidSignature = AAT{
		Version:           "0.0.1",
		ProviderPublicKey: providerPrivKey.PublicKey().RawString(),
		ClientPublicKey:   clientPrivKey.PublicKey().RawString(),
		ProviderSignature: "",
	}
	providerSignature, err := providerPrivKey.Sign(AATValidSignature.Hash())
	if err != nil {
		t.Fatalf(err.Error())
	}
	AATValidSignature.ProviderSignature = hex.EncodeToString(providerSignature)
	tests := []struct {
		name     string
		aat      AAT
		hasError bool
	}{
		{
			name:     "AAT doesn't have a signature",
			aat:      AATMissingSignature,
			hasError: true,
		},
		{
			name:     "AAT doesn't have a valid signature",
			aat:      AATInvalidSignature,
			hasError: true,
		},
		{
			name:     "AAT has a valid signature",
			aat:      AATValidSignature,
			hasError: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.aat.ValidateSignature() != nil, tt.hasError)
		})
	}
}

func TestAAT_HashString(t *testing.T) {
	providerPrivKey := GetRandomPrivateKey()
	clientPrivKey := GetRandomPrivateKey()
	var AAT = AAT{
		Version:           "0.0.1",
		ProviderPublicKey: providerPrivKey.PublicKey().RawString(),
		ClientPublicKey:   clientPrivKey.PublicKey().RawString(),
		ProviderSignature: "",
	}
	assert.True(t, len(AAT.Hash()) == HashLength)
	assert.True(t, HashVerification(AAT.HashString()) == nil)
}

func TestAAT_Validate(t *testing.T) {
	providerPrivKey := GetRandomPrivateKey()
	clientPrivKey := GetRandomPrivateKey()
	var AAT = AAT{
		Version:           "0.0.1",
		ProviderPublicKey: providerPrivKey.PublicKey().RawString(),
		ClientPublicKey:   clientPrivKey.PublicKey().RawString(),
		ProviderSignature: "",
	}
	// sign with the client (invalid)
	providerlicationSignature, err := providerPrivKey.Sign(AAT.Hash())
	if err != nil {
		t.Fatalf(err.Error())
	}
	AAT.ProviderSignature = hex.EncodeToString(providerlicationSignature)
	assert.Nil(t, AAT.Validate())
}
