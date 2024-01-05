package types

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAAT_VersionIsIncluded(t *testing.T) {
	requestorPrivKey := GetRandomPrivateKey()
	clientPrivKey := GetRandomPrivateKey()
	var AATNoVersion = AAT{
		Version:            "",
		RequestorPublicKey: requestorPrivKey.PublicKey().RawString(),
		ClientPublicKey:    clientPrivKey.PublicKey().RawString(),
		RequestorSignature: "",
	}
	var AATWithVersion = AAT{
		Version:            "0.0.1",
		RequestorPublicKey: requestorPrivKey.PublicKey().RawString(),
		ClientPublicKey:    clientPrivKey.PublicKey().RawString(),
		RequestorSignature: "",
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
	requestorPrivKey := GetRandomPrivateKey()
	clientPrivKey := GetRandomPrivateKey()
	var AATNotSupportedVersion = AAT{
		Version:            "0.0.11",
		RequestorPublicKey: requestorPrivKey.PublicKey().RawString(),
		ClientPublicKey:    clientPrivKey.PublicKey().RawString(),
		RequestorSignature: "",
	}
	var AATSupported = AAT{
		Version:            "0.0.1",
		RequestorPublicKey: requestorPrivKey.PublicKey().RawString(),
		ClientPublicKey:    clientPrivKey.PublicKey().RawString(),
		RequestorSignature: "",
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
	requestorPrivKey := GetRandomPrivateKey()
	clientPrivKey := GetRandomPrivateKey()
	var AATVersionMissing = AAT{
		Version:            "",
		RequestorPublicKey: requestorPrivKey.PublicKey().RawString(),
		ClientPublicKey:    clientPrivKey.PublicKey().RawString(),
		RequestorSignature: "",
	}
	var AATNotSupportedVersion = AAT{
		Version:            "0.0.11",
		RequestorPublicKey: requestorPrivKey.PublicKey().RawString(),
		ClientPublicKey:    clientPrivKey.PublicKey().RawString(),
		RequestorSignature: "",
	}
	var AATSupported = AAT{
		Version:            "0.0.1",
		RequestorPublicKey: requestorPrivKey.PublicKey().RawString(),
		ClientPublicKey:    clientPrivKey.PublicKey().RawString(),
		RequestorSignature: "",
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
	requestorPrivKey := GetRandomPrivateKey()
	clientPubKey := getRandomPubKey()
	var AATInvalidRequestorPubKey = AAT{
		Version:            "0.0.1",
		RequestorPublicKey: requestorPrivKey.PubKey().Address().String(),
		ClientPublicKey:    clientPubKey.RawString(),
		RequestorSignature: "",
	}
	var AATInvalidClientPubKey = AAT{
		Version:            "0.0.1",
		RequestorPublicKey: requestorPrivKey.PublicKey().RawString(),
		ClientPublicKey:    clientPubKey.Address().String(),
		RequestorSignature: "",
	}
	var AATValidMessage = AAT{
		Version:            "0.0.1",
		RequestorPublicKey: requestorPrivKey.PublicKey().RawString(),
		ClientPublicKey:    clientPubKey.RawString(),
		RequestorSignature: "",
	}
	tests := []struct {
		name     string
		aat      AAT
		hasError bool
	}{
		{
			name:     "AAT doesn't have a valid requestor pub key",
			aat:      AATInvalidRequestorPubKey,
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
	requestorPrivKey := GetRandomPrivateKey()
	clientPrivKey := GetRandomPrivateKey()
	var AATMissingSignature = AAT{
		Version:            "0.0.1",
		RequestorPublicKey: requestorPrivKey.PublicKey().RawString(),
		ClientPublicKey:    clientPrivKey.PublicKey().RawString(),
		RequestorSignature: "",
	}
	var AATInvalidSignature = AAT{
		Version:            "0.0.1",
		RequestorPublicKey: requestorPrivKey.PublicKey().RawString(),
		ClientPublicKey:    clientPrivKey.PublicKey().RawString(),
		RequestorSignature: "",
	}
	// sign with the client (invalid)
	clientSignature, err := clientPrivKey.Sign(AATInvalidSignature.Hash())
	if err != nil {
		t.Fatalf(err.Error())
	}
	AATInvalidSignature.RequestorSignature = hex.EncodeToString(clientSignature)
	// sign with the requestor
	var AATValidSignature = AAT{
		Version:            "0.0.1",
		RequestorPublicKey: requestorPrivKey.PublicKey().RawString(),
		ClientPublicKey:    clientPrivKey.PublicKey().RawString(),
		RequestorSignature: "",
	}
	requestorSignature, err := requestorPrivKey.Sign(AATValidSignature.Hash())
	if err != nil {
		t.Fatalf(err.Error())
	}
	AATValidSignature.RequestorSignature = hex.EncodeToString(requestorSignature)
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
	requestorPrivKey := GetRandomPrivateKey()
	clientPrivKey := GetRandomPrivateKey()
	var AAT = AAT{
		Version:            "0.0.1",
		RequestorPublicKey: requestorPrivKey.PublicKey().RawString(),
		ClientPublicKey:    clientPrivKey.PublicKey().RawString(),
		RequestorSignature: "",
	}
	assert.True(t, len(AAT.Hash()) == HashLength)
	assert.True(t, HashVerification(AAT.HashString()) == nil)
}

func TestAAT_Validate(t *testing.T) {
	requestorPrivKey := GetRandomPrivateKey()
	clientPrivKey := GetRandomPrivateKey()
	var AAT = AAT{
		Version:            "0.0.1",
		RequestorPublicKey: requestorPrivKey.PublicKey().RawString(),
		ClientPublicKey:    clientPrivKey.PublicKey().RawString(),
		RequestorSignature: "",
	}
	// sign with the client (invalid)
	requestorSignature, err := requestorPrivKey.Sign(AAT.Hash())
	if err != nil {
		t.Fatalf(err.Error())
	}
	AAT.RequestorSignature = hex.EncodeToString(requestorSignature)
	assert.Nil(t, AAT.Validate())
}
