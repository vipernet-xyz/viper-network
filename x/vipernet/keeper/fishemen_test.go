package keeper

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	sdk "github.com/vipernet-xyz/viper-network/types"
	vc "github.com/vipernet-xyz/viper-network/x/vipernet/types"
)

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

	fmt.Println(latencyScores)
}
