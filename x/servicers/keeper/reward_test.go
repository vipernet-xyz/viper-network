package keeper

import (
	"testing"

	"github.com/vipernet-xyz/viper-network/codec"

	sdk "github.com/vipernet-xyz/viper-network/types"
	requestorsTypes "github.com/vipernet-xyz/viper-network/x/requestors/types"
	viperTypes "github.com/vipernet-xyz/viper-network/x/viper-main/types"

	"github.com/stretchr/testify/assert"
)

type args struct {
	consAddress sdk.Address
}

func TestSetAndGetProposer(t *testing.T) {
	validator := getStakedValidator()
	consAddress := validator.GetAddress()

	tests := []struct {
		name            string
		args            args
		expectedAddress sdk.Address
	}{
		{
			name:            "can set the preivous proposer",
			args:            args{consAddress: consAddress},
			expectedAddress: consAddress,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)

			keeper.SetPreviousProposer(context, test.args.consAddress)
			receivedAddress := keeper.GetPreviousProposer(context)
			assert.True(t, test.expectedAddress.Equals(receivedAddress), "addresses do not match ")
		})
	}
}

func TestSetandGetRequestor(t *testing.T) {
	requestor := getStakedRequestor()
	consAddress := requestor.GetAddress()

	tests := []struct {
		name            string
		args            args
		expectedAddress sdk.Address
	}{
		{
			name:            "can set the requestor",
			args:            args{consAddress: consAddress},
			expectedAddress: consAddress,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)

			keeper.SetRequestorKey(context, test.args.consAddress)
			receivedAddress := keeper.GetRequestor(context)
			assert.True(t, test.expectedAddress.Equals(receivedAddress), "addresses do not match ")
		})
	}
}
func TestMint(t *testing.T) {
	validator := getStakedValidator()
	validatorAddress := validator.Address

	tests := []struct {
		name     string
		amount   sdk.BigInt
		expected string
		address  sdk.Address
		panics   bool
	}{
		{
			name:     "mints a coin",
			amount:   sdk.NewInt(90),
			expected: "a reward of ",
			address:  validatorAddress,
			panics:   false,
		},
		{
			name:     "errors invalid ammount of coins",
			amount:   sdk.NewInt(-1),
			expected: "negative coin amount: -1",
			address:  validatorAddress,
			panics:   true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)

			switch test.panics {
			case true:
				defer func() {
					err := recover().(error)
					assert.Contains(t, err.Error(), test.expected, "error does not match")
				}()
				_ = keeper.mint(context, test.amount, test.address)
			default:
				result := keeper.mint(context, test.amount, test.address)
				assert.Contains(t, result.Log, test.expected, "does not contain message")
				coins := keeper.AccountKeeper.GetCoins(context, sdk.Address(test.address))
				assert.True(t, sdk.NewCoins(sdk.NewCoin(keeper.StakeDenom(context), test.amount)).IsEqual(coins), "coins should match")
			}
		})
	}
}

func TestBurn(t *testing.T) {
	requestor := getStakedRequestor()

	tests := []struct {
		name      string
		amount    sdk.BigInt
		expected  string
		requestor requestorsTypes.Requestor
		panics    bool
	}{
		{
			name:      "burns a coin",
			amount:    sdk.NewInt(90),
			expected:  "an amount of ",
			requestor: requestor,
			panics:    false,
		},
		{
			name:      "errors invalid amount of coins",
			amount:    sdk.NewInt(-1),
			expected:  "negative coin amount: -1",
			requestor: requestor,
			panics:    true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, _, keeper := createTestInput(t, true)

			// Set the requestor in the keeper
			keeper.RequestorKeeper.SetRequestor(ctx, test.requestor)

			switch test.panics {
			case true:
				defer func() {
					if r := recover(); r != nil {
						err, ok := r.(error)
						assert.True(t, ok, "panic value is not an error")
						assert.Contains(t, err.Error(), test.expected, "error does not match")
					}
				}()
				_, _ = keeper.burn(ctx, test.amount, test.requestor, 5)
			default:
				result, err := keeper.burn(ctx, test.amount, test.requestor, 5)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				assert.Contains(t, result.Log, test.expected, "does not contain message")
				updatedRequestor, _ := keeper.RequestorKeeper.GetRequestor(ctx, test.requestor.Address)
				expectedTokens := sdk.MaxInt(sdk.NewInt(0), requestor.StakedTokens.Sub(test.amount))
				assert.True(t, expectedTokens.Equal(updatedRequestor.StakedTokens), "tokens should match")
			}
		})
	}
}

func TestKeeper_rewardFromFees(t *testing.T) {
	type fields struct {
		keeper Keeper
	}

	type args struct {
		ctx              sdk.Ctx
		previousProposer sdk.Address
		requestor        sdk.Address
		Output           sdk.Address
		aOutput          sdk.Address
		Amount           sdk.BigInt
	}
	stakedValidator := getStakedValidator()
	stakedRequestor := getStakedRequestor()
	stakedValidator.OutputAddress = getRandomValidatorAddress()
	stakedRequestor.Address = getRandomRequestorAddress()
	codec.TestMode = -3
	amount := sdk.NewInt(10000)
	fees := sdk.NewCoins(sdk.NewCoin("uvipr", amount))
	context, _, keeper := createTestInput(t, true)
	fp := keeper.getFeePool(context)
	keeper.AccountKeeper.SetCoins(context, fp.GetAddress(), fees)
	fp = keeper.getFeePool(context)
	keeper.SetValidator(context, stakedValidator)
	assert.Equal(t, fees, fp.GetCoins())
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{"Test blockReward", fields{keeper: keeper},
			args{
				ctx:              context,
				previousProposer: stakedValidator.GetAddress(),
				requestor:        stakedRequestor.GetAddress(),
				Output:           stakedValidator.OutputAddress,
				aOutput:          stakedRequestor.Address,
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.fields.keeper
			ctx := tt.args.ctx
			k.blockReward(tt.args.ctx, tt.args.previousProposer)
			acc := k.GetAccount(ctx, tt.args.Output)
			assert.False(t, acc.Coins.IsZero())
			assert.True(t, acc.Coins.IsEqual(sdk.NewCoins(sdk.NewCoin("uvipr", sdk.NewInt(3334)))))
			acc = k.GetAccount(ctx, tt.args.previousProposer)
			assert.True(t, acc.Coins.IsZero())
		})
	}

}

func getRandomRequestorAddress() sdk.Address {
	return sdk.Address(getRandomPubKey().Address())
}

func GetRequestor() requestorsTypes.Requestor {
	pub := getRandomPubKey()
	return requestorsTypes.Requestor{
		Address:      sdk.Address(pub.Address()),
		StakedTokens: sdk.NewInt(100000000000),
		PublicKey:    pub,
		Jailed:       false,
		Status:       sdk.Staked,
		MaxRelays:    sdk.NewInt(100000000000),
		Chains:       []string{"0001"},
		GeoZones:     []string{"0001"},
		NumServicers: 5,
	}
}
func getStakedRequestor() requestorsTypes.Requestor {
	return GetRequestor()
}

func TestKeeper_rewardFromRelays(t *testing.T) {
	type fields struct {
		keeper Keeper
	}
	type args struct {
		ctx               sdk.Ctx
		validator         sdk.Address
		Output            sdk.Address
		validatorNoOutput sdk.Address
		OutputNoOutput    sdk.Address
	}
	originalTestMode := codec.TestMode

	t.Cleanup(func() {
		codec.TestMode = originalTestMode
	})
	stakedValidator := getStakedValidator()
	stakedValidatorNoOutput := getStakedValidator()
	stakedValidatorNoOutput.OutputAddress = nil
	stakedValidator.OutputAddress = getRandomValidatorAddress()
	codec.TestMode = -3
	context, _, keeper := createTestInput(t, true)
	context = context.WithBlockHeight(11000)
	keeper.SetValidator(context, stakedValidator)
	keeper.SetValidator(context, stakedValidatorNoOutput)
	latencyScore := sdk.NewDecWithPrec(8, 1)
	availabilityScore := sdk.NewDecWithPrec(7, 1)
	reliabilityScore := sdk.NewDecWithPrec(9, 1)

	reportCard := viperTypes.MsgSubmitQoSReport{
		Report: viperTypes.ViperQoSReport{
			LatencyScore:      latencyScore,
			AvailabilityScore: availabilityScore,
			ReliabilityScore:  reliabilityScore,
		},
		FishermanAddress: getRandomValidatorAddress(),
	}

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{"Test RewardForRelays", fields{keeper: keeper},
			args{
				ctx:               context,
				validator:         stakedValidator.GetAddress(),
				Output:            stakedValidator.OutputAddress,
				validatorNoOutput: stakedValidatorNoOutput.GetAddress(),
				OutputNoOutput:    stakedValidatorNoOutput.GetAddress(),
			}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.fields.keeper
			ctx := tt.args.ctx
			p := getStakedRequestor()

			// Reward for relays with output
			relays := sdk.NewInt(10000)
			k.RewardForRelays(ctx, reportCard, relays, p)
			// Check the rewards
			acc := k.GetAccount(ctx, tt.args.Output)
			assert.False(t, acc.Coins.IsZero())
			assert.True(t, acc.Coins.IsEqual(sdk.NewCoins(sdk.NewCoin("uvipr", sdk.NewInt(6000000)))))

			// Check that the validator account is zeroed out
			acc = k.GetAccount(ctx, tt.args.validator)
			assert.True(t, acc.Coins.IsZero())

			// Reward for relays without output
			k.RewardForRelays(ctx, reportCard, relays, p)

			// Check the rewards
			acc = k.GetAccount(ctx, tt.args.OutputNoOutput)
			assert.False(t, acc.Coins.IsZero())
			assert.True(t, acc.Coins.IsEqual(sdk.NewCoins(sdk.NewCoin("uvipr", sdk.NewInt(6000000)))))

			// Check that the validator accounts are the same (since there's no output address)
			acc2 := k.GetAccount(ctx, tt.args.validatorNoOutput)
			assert.Equal(t, acc, acc2)
		})
	}
}
