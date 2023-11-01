package vipernet

import (
	"testing"

	"github.com/vipernet-xyz/viper-network/x/vipernet/types"

	"github.com/stretchr/testify/assert"
)

func TestInitExportGenesis(t *testing.T) {
	ctx, _, _, k, _ := createTestInput(t, false)
	p := types.Params{
		ClaimSubmissionWindow: 22,
		SupportedBlockchains:  []string{"eth"},
		ClaimExpiration:       55,
		MinimumNumberOfProofs: int64(5),
	}
	genesisState := types.GenesisState{
		Params:      p,
		Claims:      []types.MsgClaim(nil),
		ReportCards: []types.MsgSubmitReportCard(nil),
	}
	InitGenesis(ctx, k, genesisState)
	assert.Equal(t, k.GetParams(ctx), p)
	gen := ExportGenesis(ctx, k)
	assert.Equal(t, genesisState, gen)
}
