package types

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	sdk "github.com/vipernet-xyz/viper-network/types"
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
	fmt.Println("report:", report)
	// Test the expected QoS report fields
	assert.Equal(t, report.BlockHeight, blockHeight)
	assert.Equal(t, report.ServicerAddress, result.ServicerAddress)
	assert.NotNil(t, report.FirstSampleTimestamp)
	assert.True(t, report.AvailabilityScore.GT(sdk.ZeroDec()))
	assert.True(t, report.ReliabilityScore.GT(sdk.ZeroDec()))
}
