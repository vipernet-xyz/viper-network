package keeper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/vipernet/types"
)

func TestKeeper_GetSetReportCard(t *testing.T) {
	ctx, _, _, _, keeper, _, _ := createTestInput(t, false)

	// Simulate some relay proofs for the test
	npk, fpk, header, _ := simulateTestRelays(t, keeper, &ctx, 5)
	result, err := types.GetResult(header, types.FishermanTestEvidence, sdk.Address(npk.Address()), types.GlobalTestCache)
	assert.Nil(t, err)

	// Create a MsgSubmitReportCard with sample data
	reportCard := types.MsgSubmitReportCard{
		SessionHeader:    header,
		ServicerAddress:  sdk.Address(npk.Address()),
		FishermanAddress: sdk.Address(fpk.Address()),
		Report: types.ViperQoSReport{
			FirstSampleTimestamp: time.Now().UTC(),
			BlockHeight:          1000,
			ServicerAddress:      sdk.Address(npk.Address()),
			LatencyScore:         sdk.NewDecWithPrec(123, 2),
			AvailabilityScore:    sdk.NewDecWithPrec(456, 2),
			ReliabilityScore:     sdk.NewDecWithPrec(789, 2),
			SampleRoot:           result.GenerateSampleMerkleRoot(0, types.GlobalTestCache),
			Nonce:                42,
			Signature:            "signature",
		},
		EvidenceType: types.RelayEvidence,
	}

	// Mock the context for testing
	mockCtx := new(Ctx)
	mockCtx.On("KVStore", keeper.storeKey).Return(ctx.KVStore(keeper.storeKey))
	mockCtx.On("BlockHeight").Return(int64(1))
	mockCtx.On("PrevCtx", header.SessionBlockHeight).Return(ctx, nil)
	mockCtx.On("BlockHash", header.SessionBlockHeight).Return(types.Hash([]byte("fake")), nil)
	mockCtx.On("Logger").Return(ctx.Logger())

	// Set the report card
	err = keeper.SetReportCard(mockCtx, reportCard)
	assert.Nil(t, err)

	// Get the report card
	queriedReportCard, found := keeper.GetReportCard(mockCtx, reportCard.ServicerAddress, reportCard.FishermanAddress, header)
	assert.True(t, found)

	// Assertions to check the retrieved report card matches the original one
	assert.Equal(t, reportCard.Report.FirstSampleTimestamp, queriedReportCard.Report.FirstSampleTimestamp)
	assert.Equal(t, reportCard.Report.BlockHeight, queriedReportCard.Report.BlockHeight)
	assert.Equal(t, reportCard.Report.ServicerAddress, queriedReportCard.Report.ServicerAddress)
	assert.Equal(t, reportCard.Report.LatencyScore, queriedReportCard.Report.LatencyScore)
	assert.Equal(t, reportCard.Report.AvailabilityScore, queriedReportCard.Report.AvailabilityScore)
	assert.Equal(t, reportCard.Report.ReliabilityScore, queriedReportCard.Report.ReliabilityScore)
	assert.Equal(t, reportCard.Report.SampleRoot, queriedReportCard.Report.SampleRoot)
	assert.Equal(t, reportCard.Report.Nonce, queriedReportCard.Report.Nonce)
	assert.Equal(t, reportCard.Report.Signature, queriedReportCard.Report.Signature)
}

func TestKeeper_GetSetDeleteReportCards(t *testing.T) {
	ctx, _, _, _, keeper, _, _ := createTestInput(t, false)
	var reportCards []types.MsgSubmitReportCard
	var servicerAddrs []sdk.Address
	var fishermanAddrs []sdk.Address

	for i := 0; i < 2; i++ {
		npk, fpk, header, _ := simulateTestRelays(t, keeper, &ctx, 5)
		result, _ := types.GetResult(header, types.FishermanTestEvidence, sdk.Address(npk.Address()), types.GlobalTestCache)
		reportCard := types.MsgSubmitReportCard{
			SessionHeader:    header,
			ServicerAddress:  sdk.Address(sdk.Address(npk.Address())),
			FishermanAddress: sdk.Address(sdk.Address(fpk.Address())),
			Report: types.ViperQoSReport{
				FirstSampleTimestamp: time.Now(),
				BlockHeight:          int64(i),
				ServicerAddress:      sdk.Address(sdk.Address(npk.Address())),
				LatencyScore:         sdk.NewDecWithPrec(12345, 6),
				AvailabilityScore:    sdk.NewDecWithPrec(67890, 6),
				ReliabilityScore:     sdk.NewDecWithPrec(54321, 6),
				SampleRoot:           result.GenerateSampleMerkleRoot(0, types.GlobalTestCache),
				Nonce:                int64(42),
				Signature:            "sample_signature",
			},
			EvidenceType: types.FishermanTestEvidence,
		}
		reportCards = append(reportCards, reportCard)
		servicerAddrs = append(servicerAddrs, sdk.Address(sdk.Address(npk.Address())))
		fishermanAddrs = append(fishermanAddrs, sdk.Address(sdk.Address(fpk.Address())))
	}

	mockCtx := new(Ctx)
	mockCtx.On("KVStore", keeper.storeKey).Return(ctx.KVStore(keeper.storeKey))
	mockCtx.On("Logger").Return(ctx.Logger())

	keeper.SetReportCards(mockCtx, reportCards)
	rc := keeper.GetAllReportCards(mockCtx)
	assert.Len(t, rc, 2)

	_ = keeper.DeleteReportCard(mockCtx, servicerAddrs[0], fishermanAddrs[0], reportCards[0].SessionHeader)
	_, err := keeper.GetReportCard(ctx, servicerAddrs[0], fishermanAddrs[0], reportCards[0].SessionHeader)
	assert.NotNil(t, err)
}
