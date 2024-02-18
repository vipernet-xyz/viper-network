package types

import (
	"encoding/hex"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vipernet-xyz/utils-go/client"
	sdk "github.com/vipernet-xyz/viper-network/types"
	servicerTypes "github.com/vipernet-xyz/viper-network/x/servicers/types"
)

func TestCalculateQoSForServicer(t *testing.T) {
	privateKey := GetRandomPrivateKey()
	PubKey := privateKey.PublicKey()

	// Sample test data for the ServicerResults
	result := &ServicerResults{
		Timestamps:      []time.Time{time.Now(), time.Now()},
		Availabilities:  []bool{true, false, false, true, true},
		Latencies:       []time.Duration{time.Millisecond * 10, time.Millisecond * 20, time.Millisecond * 30, time.Millisecond * 40, time.Millisecond * 80},
		Reliabilities:   []bool{true, true, true, false, true},
		ServicerAddress: sdk.Address(PubKey.Address()),
	}

	// Block height for testing
	blockHeight := int64(1000)
	latencyScore := sdk.NewDecWithPrec(7, 1)
	// Calculate the QoS report
	report, err := CalculateQoSForServicer(result, blockHeight, latencyScore)
	if err != nil {
		t.Errorf("CalculateQoSForServicer returned an error: %v", err)
	}
	// Test the expected QoS report fields
	assert.Equal(t, report.BlockHeight, blockHeight)
	assert.Equal(t, report.ServicerAddress, result.ServicerAddress)
	assert.NotNil(t, report.FirstSampleTimestamp)
	assert.True(t, report.AvailabilityScore.GT(sdk.ZeroDec()))
	assert.True(t, report.ReliabilityScore.GT(sdk.ZeroDec()))
}

func GetValidator() servicerTypes.Validator {
	pub := getRandomPubKey()
	return servicerTypes.Validator{
		Address:      sdk.Address(pub.Address()),
		StakedTokens: sdk.NewInt(100000000000),
		PublicKey:    pub,
		Jailed:       false,
		Status:       sdk.Staked,
		Chains:       []string{"0002", "0003"},
		ServiceURL:   "https://mynode.com:8082",
		GeoZone:      []string{"US"},
	}
}

func TestRelayer_SendSampleRelay(t *testing.T) {
	c := require.New(t)

	// Mocking for Sender.Relay
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Set up a responder for the POST request to "https://mynode.com:8082/v1/client/relay"
	httpmock.RegisterResponder(
		http.MethodPost,
		"https://mynode.com:8082/v1/client/relay",
		httpmock.NewStringResponder(http.StatusOK, `{
			"response": "{\"id\":3905054414,\"jsonrpc\":\"2.0\",\"result\":\"0xdd03e4\"}",
			"signature": "abfabfabfabfabfabfabfabfabfabfabfabfabfabfabfabfabfabfabfabfabfabf"
		  }`),
	)

	reqPubKey := getRandomPubKey()
	clientPubKey := getRandomPubKey()
	fishermanPrivKey := GetRandomPrivateKey()
	servicer := GetValidator()

	// Initialize a sample signer and sender
	signer := Signer{
		address:    fishermanPrivKey.PubKey().Address().String(),
		publicKey:  string(fishermanPrivKey.PubKey().Bytes()),
		privateKey: hex.EncodeToString(fishermanPrivKey[:]),
	}

	fisherman := servicerTypes.Validator{
		Address:      sdk.Address(fishermanPrivKey.PublicKey().Address()),
		StakedTokens: sdk.NewInt(100000000000),
		PublicKey:    fishermanPrivKey.PublicKey(),
		Jailed:       false,
		Status:       sdk.Staked,
		Chains:       []string{"0002", "0003"},
		ServiceURL:   "https://mynode.com:8082",
		GeoZone:      []string{"US"},
	}

	sender := Sender{
		rpcURL: "https://mynode.com:8082",
		client: client.NewDefaultClient(),
	}

	// Create a relayer instance
	relayer := NewRelayer(signer, sender)

	// Call SendSampleRelay
	output, err := relayer.SendSampleRelay(123, "0002", FishermenTrigger{
		Proof: RelayProof{
			RequestHash:        hex.EncodeToString(Hash([]byte("fake"))),
			Entropy:            rand.Int63(),
			SessionBlockHeight: 123,
			ServicerPubKey:     servicer.PublicKey.String(),
			Blockchain:         "0002",
			Token: AAT{
				Version:            "0.0.1",
				RequestorPublicKey: reqPubKey.RawString(),
				ClientPublicKey:    clientPubKey.RawString(),
				RequestorSignature: "",
			},
			Signature:    "",
			GeoZone:      "US",
			NumServicers: 5,
		},
	}, servicer, fisherman, &HostedBlockchains{})

	// Assertions
	c.NoError(err)
	c.NotNil(output)
	c.NotNil(output.RelayOutput)
	c.NotNil(output.Proof)
	c.NotNil(output.Latency)
	c.NotNil(output.Reliability)

	// Mock verification
	calls := httpmock.GetCallCountInfo()

	// Verify the number of calls to Sender.Relay
	c.Equal(1, calls["POST https://mynode.com:8082/v1/client/relay"])

	// Reset the mocks
	httpmock.Reset()
}
