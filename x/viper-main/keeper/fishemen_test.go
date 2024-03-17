package keeper

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/viper-main/types"
	vc "github.com/vipernet-xyz/viper-network/x/viper-main/types"
)

func TestHandleFishermanTrigger(t *testing.T) {
	ctx, servicers, _, _, keeper, _, _ := createTestInput(t, false)
	requestorPrivateKey := getRandomPrivateKey()
	requestorPubKey := requestorPrivateKey.PublicKey().RawString()
	clientPrivateKey := getRandomPrivateKey()
	numServicers := 5
	var servicerAddrs []sdk.Address
	for i := 0; i < int(numServicers) && i < len(servicers); i++ {
		servicerAddrs = append(servicerAddrs, servicers[i].GetAddress())
	}
	fisherman := append(servicerAddrs, sdk.Address(getRandomPubKey().Address()))
	// Create a FishermenTrigger
	trigger := vc.FishermenTrigger{
		Proof: types.RelayProof{
			Entropy:            rand.Int63(),
			SessionBlockHeight: ctx.BlockHeight(),
			ServicerPubKey:     servicers[0].PublicKey.RawString(),
			Blockchain:         "ethereum",
			Token: types.AAT{
				Version:            "0.0.1",
				RequestorPublicKey: requestorPubKey,
				ClientPublicKey:    clientPrivateKey.PublicKey().RawString(),
				RequestorSignature: "",
			},
			Signature:    "",
			GeoZone:      "US",
			NumServicers: 5,
		},
	}

	// Set up a session
	sessionHeader := vc.SessionHeader{
		RequestorPubKey:    requestorPubKey,
		Chain:              "ethereum",
		GeoZone:            "US",
		NumServicers:       5,
		SessionBlockHeight: ctx.BlockHeight(),
	}
	session := vc.Session{
		SessionHeader:    sessionHeader,
		SessionKey:       []byte("session_key"),
		SessionServicers: servicerAddrs,
		SessionFishermen: fisherman,
	}

	// Store the session using your keeper
	vc.SetSession(session, vc.GlobalSessionCache)

	// Call HandleFishermanTrigger
	resp, err := keeper.HandleFishermanTrigger(ctx, trigger)

	// Check the result
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "fisherman triggered", resp.Response)
	assert.NotEmpty(t, resp.Proof)
}

func TestCalculateLatencyScores(t *testing.T) {
	// Create a sample dataset of servicer results
	results := map[string]*vc.ServicerResults{
		"servicerA": {
			Latencies: []time.Duration{time.Millisecond * 100, time.Millisecond * 200},
		},
		"servicerB": {
			Latencies: []time.Duration{time.Millisecond * 50, time.Millisecond * 150},
		},
		"servicerC": {
			Latencies: []time.Duration{time.Millisecond * 10, time.Millisecond * 20},
		},
		"servicerD": {
			Latencies: []time.Duration{},
		},
	}

	// Calculate latency scores
	latencyScores := CalculateLatencyScores(results)

	// Assert that the map contains the correct number of scores
	assert.Equal(t, 4, len(latencyScores))
	// Test specific scores for each servicer
	assert.True(t, latencyScores["servicerA"].GT(sdk.ZeroDec())) // Expecting a positive score
	assert.True(t, latencyScores["servicerB"].GT(sdk.ZeroDec())) // Expecting a positive score
	assert.True(t, latencyScores["servicerC"].GT(sdk.ZeroDec())) // Expecting a positive score

	// Ensure that servicerB has a higher score than servicerA (since servicerB has lower latency)
	assert.True(t, latencyScores["servicerB"].GT(latencyScores["servicerA"]))

	// Ensure that servicerC has the highest score (since servicerC has the lowest latency)
	assert.True(t, latencyScores["servicerC"].GT(latencyScores["servicerB"]))

	// Test a servicer with no latencies (should have a score of 0)
	servicerXScore, exists := latencyScores["servicerD"]
	assert.True(t, exists)                  // Check if "servicerD" key exists
	assert.True(t, servicerXScore.IsZero()) // Expecting a score of 0 for "servicerD"

}
