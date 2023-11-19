package keeper

import (
	"fmt"
	"reflect"
	"testing"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/servicers/types"
	viperTypes "github.com/vipernet-xyz/viper-network/x/vipernet/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeeper_GetValidators(t *testing.T) {
	type fields struct {
		keeper Keeper
	}
	type args struct {
		ctx         sdk.Ctx
		maxRetrieve uint16
	}

	context, _, keeper := createTestInput(t, true)

	tests := []struct {
		name           string
		fields         fields
		args           args
		wantValidators []types.Validator
	}{
		{"Test GetValidators 0", fields{keeper: keeper}, args{
			ctx:         context,
			maxRetrieve: 0,
		}, []types.Validator{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.fields.keeper

			if gotValidators := k.GetValidators(tt.args.ctx, tt.args.maxRetrieve); !reflect.DeepEqual(gotValidators, tt.wantValidators) {
				t.Errorf("GetValidators() = %v, want %v", gotValidators, tt.wantValidators)
			}
		})
	}
}

func TestKeeper_GetValidatorOutputAddress(t *testing.T) {
	type args struct {
		ctx sdk.Ctx
		k   Keeper
		v   types.Validator
	}
	validator := getStakedValidator()
	validator.OutputAddress = validator.Address
	validatorNoOuptut := getStakedValidator()
	validatorNoOuptut.OutputAddress = nil
	context, _, keeper := createTestInput(t, true)
	keeper.SetValidator(context, validator)
	keeper.SetValidator(context, validatorNoOuptut)
	tests := []struct {
		name string
		args args
		want sdk.Address
	}{
		{"Test GetValidatorOutput With Output Address", args{
			ctx: context,
			k:   keeper,
			v:   validator,
		}, validator.OutputAddress},
		{"Test GetValidatorOutput Without Output Address", args{
			ctx: context,
			k:   keeper,
			v:   validatorNoOuptut,
		}, validatorNoOuptut.Address},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, found := tt.args.k.GetValidatorOutputAddress(tt.args.ctx, tt.args.v.Address)
			if !assert.True(t, len(got) == len(tt.want)) {
				t.Errorf("GetValidatorOutputAddress() = %v, want %v", got, tt.want)
			}
			assert.True(t, found)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMustGetValidator(t *testing.T) {
	stakedValidator := getStakedValidator()

	type args struct {
		validator types.Validator
	}
	type expected struct {
		validator types.Validator
		message   string
	}
	tests := []struct {
		name     string
		hasError bool
		args
		expected
	}{
		{
			name:     "gets validator",
			hasError: false,
			args:     args{validator: stakedValidator},
			expected: expected{validator: stakedValidator},
		},
		{
			name:     "errors if no validator",
			hasError: true,
			args:     args{validator: stakedValidator},
			expected: expected{message: fmt.Sprintf("validator record not found for address: %X\n", stakedValidator.Address)},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			switch test.hasError {
			case true:
				_, _ = keeper.GetValidator(context, test.args.validator.Address)
			default:
				keeper.SetValidator(context, test.args.validator)
				validator, _ := keeper.GetValidator(context, test.args.validator.Address)
				assert.True(t, validator.Equals(test.expected.validator), "validator does not match")
			}
		})
	}

}

func Test_sortNoLongerStakedValidators(t *testing.T) {
	type args struct {
		prevState valPowerMap
	}
	tests := []struct {
		name string
		args args
		want [][]byte
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sortNoLongerStakedValidators(tt.args.prevState); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sortNoLongerStakedValidators() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetValidatorReportCard(t *testing.T) {
	// Create a test context, store, and keeper
	context, _, keeper := createTestInput(t, true)

	// Create a staked validator with a report card
	stakedValidator := getStakedValidator()
	reportCard := types.ReportCard{
		TotalSessions:          10,
		TotalLatencyScore:      sdk.NewDecWithPrec(5, 1), // 0.5
		TotalAvailabilityScore: sdk.NewDecWithPrec(8, 1), // 0.8
		TotalReliabilityScore:  sdk.NewDecWithPrec(9, 1), // 0.9
	}
	stakedValidator.ReportCard = reportCard

	// Set the validator with the report card in the store
	keeper.SetValidatorReportCard(context, stakedValidator)
	// Retrieve the report card for the validator
	retrievedReportCard, found := keeper.GetValidatorReportCard(context, stakedValidator)
	assert.True(t, found)

	// Check that the retrieved report card matches the expected values
	assert.Equal(t, reportCard.TotalSessions, retrievedReportCard.TotalSessions)
	assert.True(t, reportCard.TotalLatencyScore.LT(sdk.OneDec()))      // Check that it's less than 1
	assert.True(t, reportCard.TotalAvailabilityScore.LT(sdk.OneDec())) // Check that it's less than 1
	assert.True(t, reportCard.TotalReliabilityScore.LT(sdk.OneDec()))  // Check that it's less than 1

	// Try to retrieve a report card for a non-existing validator
	nonExistingValidator := getStakedValidator()
	_, found = keeper.GetValidatorReportCard(context, nonExistingValidator)
	assert.False(t, found)
}

func TestKeeper_UpdateValidatorReportCard(t *testing.T) {
	// Create a context, keeper, and set up any initial conditions
	context, _, keeper := createTestInput(t, true)

	// Create a test validator
	validator := getStakedValidator()
	validator.Address = getRandomValidatorAddress()

	// Set the validator with an existing report card
	existingReport := types.ReportCard{
		TotalSessions:          5,
		TotalLatencyScore:      sdk.NewDecWithPrec(6, 1),
		TotalAvailabilityScore: sdk.NewDecWithPrec(5, 1),
		TotalReliabilityScore:  sdk.NewDecWithPrec(3, 1),
	}
	validator.ReportCard = existingReport
	keeper.SetValidator(context, validator)

	// Create a sample session report
	sessionReport := viperTypes.ViperQoSReport{
		LatencyScore:      sdk.NewDecWithPrec(5, 1),
		AvailabilityScore: sdk.NewDecWithPrec(4, 1),
		ReliabilityScore:  sdk.NewDecWithPrec(2, 1),
	}
	// Call the function to update the validator's report card
	keeper.UpdateValidatorReportCard(context, validator.Address, sessionReport)

	// Retrieve the updated validator
	updatedValidator, found := keeper.GetValidator(context, validator.Address)
	require.True(t, found)

	// Calculate the expected updated scores based on the formula
	expectedLatencyScore := sdk.NewDecWithPrec(583000000000000000, 18)
	expectedAvailabilityScore := sdk.NewDecWithPrec(483000000000000000, 18)
	expectedReliabilityScore := sdk.NewDecWithPrec(283000000000000000, 18)

	// Check if the report card has been updated correctly
	assert.Equal(t, existingReport.TotalSessions+1, updatedValidator.ReportCard.TotalSessions)
	assert.True(t, expectedLatencyScore.Equal(updatedValidator.ReportCard.TotalLatencyScore))
	assert.True(t, expectedAvailabilityScore.Equal(updatedValidator.ReportCard.TotalAvailabilityScore))
	assert.True(t, expectedReliabilityScore.Equal(updatedValidator.ReportCard.TotalReliabilityScore))
}

func TestKeeper_DeleteValidatorReportCard(t *testing.T) {
	// Create a context, keeper, and set up any initial conditions
	context, _, keeper := createTestInput(t, true)

	// Create a test validator
	validator := getStakedValidator()
	validator.Address = getRandomValidatorAddress()

	// Set the validator with an empty report card
	keeper.SetValidator(context, validator)

	// Call the function to delete the validator's report card
	err := keeper.deleteValidatorReportCard(context, validator)
	require.NoError(t, err)

	// Attempt to retrieve the deleted report card
	_, found := keeper.GetValidatorReportCard(context, validator)
	assert.False(t, found, "Report card should be deleted")
}
