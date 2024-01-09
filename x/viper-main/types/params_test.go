package types

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParams_Equal(t *testing.T) {
	p1 := DefaultParams()
	p2 := DefaultParams()
	p3 := DefaultParams()
	p3.MinimumSampleRelays = 50
	assert.True(t, p1.Equal(p2))
	assert.False(t, p2.Equal(p3))
}

func TestParams_Validate(t *testing.T) {
	ethereum := hex.EncodeToString([]byte{01})
	validParams := DefaultParams()
	validParams.SupportedBlockchains = []string{ethereum}
	// invalid waiting period
	invalidParamsWaitingPeriod := validParams
	invalidParamsWaitingPeriod.ClaimSubmissionWindow = -1
	// invalid supported chains
	invalidParamsSupported := validParams
	invalidParamsSupported.SupportedBlockchains = []string{"invalid"}
	// invalid claim expiration
	invalidParamsClaims := validParams
	invalidParamsClaims.ClaimExpiration = -1
	tests := []struct {
		name     string
		params   Params
		hasError bool
	}{
		{
			name:     "Invalid Params, session waiting period",
			params:   invalidParamsWaitingPeriod,
			hasError: true,
		},
		{
			name:     "Invalid Params, supported chains",
			params:   invalidParamsSupported,
			hasError: true,
		},
		{
			name:     "Invalid Params, claims",
			params:   invalidParamsClaims,
			hasError: true,
		},
		{
			name:     "Valid Params",
			params:   validParams,
			hasError: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.params.Validate() != nil, tt.hasError)
		})
	}
}

func TestDefaultParams(t *testing.T) {
	assert.True(t, Params{
		ClaimSubmissionWindow:      DefaultClaimSubmissionWindow,
		SupportedBlockchains:       DefaultSupportedBlockchains,
		ClaimExpiration:            DefaultClaimExpiration,
		ReplayAttackBurnMultiplier: DefaultReplayAttackBurnMultiplier,
		MinimumNumberOfProofs:      DefaultMinimumNumberOfProofs,
		BlockByteSize:              DefaultBlockByteSize,
		SupportedGeoZones:          nil,
		MinimumSampleRelays:        DefaultMinimumSampleRelays,
		ReportCardSubmissionWindow: DefaultReportCardSubmissionWindow,
	}.Equal(DefaultParams()))
}

func TestParams_ParamSetPairs(t *testing.T) {
	df := DefaultParams()
	assert.NotPanics(t, func() { df.ParamSetPairs() })
}

func TestParams_String(t *testing.T) {
	df := DefaultParams()
	assert.NotPanics(t, func() { _ = df.String() })
}