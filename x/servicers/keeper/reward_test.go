package keeper

import (
	"fmt"
	"testing"

	"github.com/vipernet-xyz/viper-network/codec"

	sdk "github.com/vipernet-xyz/viper-network/types"
	providersTypes "github.com/vipernet-xyz/viper-network/x/providers/types"
	"github.com/vipernet-xyz/viper-network/x/servicers/types"
	viperTypes "github.com/vipernet-xyz/viper-network/x/vipernet/types"

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

func TestSetandGetProvider(t *testing.T) {
	provider := getStakedProvider()
	consAddress := provider.GetAddress()

	tests := []struct {
		name            string
		args            args
		expectedAddress sdk.Address
	}{
		{
			name:            "can set the provider",
			args:            args{consAddress: consAddress},
			expectedAddress: consAddress,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)

			keeper.SetProviderKey(context, test.args.consAddress)
			receivedAddress := keeper.GetProvider(context)
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

func TestKeeper_rewardFromFees(t *testing.T) {
	type fields struct {
		keeper Keeper
	}

	type args struct {
		ctx              sdk.Ctx
		previousProposer sdk.Address
		provider         sdk.Address
		Output           sdk.Address
		aOutput          sdk.Address
		Amount           sdk.BigInt
	}
	stakedValidator := getStakedValidator()
	stakedProvider := getStakedProvider()
	stakedValidator.OutputAddress = getRandomValidatorAddress()
	stakedProvider.Address = getRandomApplicationAddress()
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
				provider:         stakedProvider.GetAddress(),
				Output:           stakedValidator.OutputAddress,
				aOutput:          stakedProvider.Address,
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.fields.keeper
			ctx := tt.args.ctx
			k.blockReward(tt.args.ctx, tt.args.previousProposer)
			acc := k.GetAccount(ctx, tt.args.Output)
			assert.False(t, acc.Coins.IsZero())
			fmt.Println(acc.Coins)
			assert.True(t, acc.Coins.IsEqual(sdk.NewCoins(sdk.NewCoin("uvipr", sdk.NewInt(3334)))))
			acc = k.GetAccount(ctx, tt.args.previousProposer)
			assert.True(t, acc.Coins.IsZero())
		})
	}

}

func getRandomApplicationAddress() sdk.Address {
	return sdk.Address(getRandomPubKey().Address())
}

func GetProvider() providersTypes.Provider {
	pub := getRandomPubKey()
	return providersTypes.Provider{
		Address:      sdk.Address(pub.Address()),
		StakedTokens: sdk.NewInt(100000000000),
		PublicKey:    pub,
		Jailed:       false,
		Status:       sdk.Staked,
		MaxRelays:    sdk.NewInt(100000000000),
		Chains:       []string{"0001"},
	}
}
func getStakedProvider() providersTypes.Provider {
	return GetProvider()
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
	stakedValidator := getStakedValidator()
	stakedValidatorNoOutput := getStakedValidator()
	stakedValidatorNoOutput.OutputAddress = nil
	stakedValidator.OutputAddress = getRandomValidatorAddress()
	codec.TestMode = -3
	context, _, keeper := createTestInput(t, true)
	keeper.SetValidator(context, stakedValidator)
	keeper.SetValidator(context, stakedValidatorNoOutput)
	// Create a sample report card
	reportCard := viperTypes.MsgSubmitReportCard{
		Report: viperTypes.ViperQoSReport{
			LatencyScore:      sdk.NewDecWithPrec(8, 1),
			AvailabilityScore: sdk.NewDecWithPrec(7, 1),
			ReliabilityScore:  sdk.NewDecWithPrec(9, 1),
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
			p := k.providerKeeper.Provider(ctx, tt.args.ctx.BlockHeader().ApplicationAddress)
			p1 := p.(providersTypes.Provider)
			k.RewardForRelays(tt.args.ctx, reportCard, sdk.NewInt(10000), tt.args.validator, p1)
			acc := k.GetAccount(ctx, tt.args.Output)
			acc1 := k.GetAccount(ctx, tt.args.ctx.BlockHeader().ApplicationAddress)
			assert.False(t, acc.Coins.IsZero())
			// Update the expected coin amount based on your actual calculations
			assert.True(t, acc.Coins.IsEqual(sdk.NewCoins(sdk.NewCoin("uvipr", sdk.NewInt(8000000)))))
			assert.False(t, acc1.Coins.IsZero())
			// Update the expected coin amount based on your actual calculations
			assert.True(t, acc1.Coins.IsEqual(sdk.NewCoins(sdk.NewCoin("uvipr", sdk.NewInt(500000)))))
			acc = k.GetAccount(ctx, tt.args.validator)
			assert.True(t, acc.Coins.IsZero())

			// no output now
			k.RewardForRelays(tt.args.ctx, reportCard, sdk.NewInt(10000), tt.args.validatorNoOutput, p1)
			acc = k.GetAccount(ctx, tt.args.OutputNoOutput)
			assert.False(t, acc.Coins.IsZero())
			// Update the expected coin amount based on your actual calculations
			assert.True(t, acc.Coins.IsEqual(sdk.NewCoins(sdk.NewCoin("uvipr", sdk.NewInt(8000000)))))
			acc2 := k.GetAccount(ctx, tt.args.validatorNoOutput)
			assert.Equal(t, acc, acc2)
		})
	}
}

func TestKeeper_rewardFromRelaysNoEXP(t *testing.T) {
	type fields struct {
		keeper Keeper
	}
	type args struct {
		ctx        sdk.Ctx
		baseReward sdk.BigInt
		relays     int64
		validator1 types.Validator
		validator2 types.Validator
		validator3 types.Validator
		validator4 types.Validator
	}

	context, _, keeper := createTestInput(t, true)
	context = context.WithBlockHeight(3)
	p := keeper.GetParams(context)
	keeper.SetParams(context, p)

	stakedValidator := getStakedValidator()

	numRelays := int64(10000)
	base := keeper.TokenRewardFactor(context)

	keeper.SetValidator(context, stakedValidator)
	reportCard := viperTypes.MsgSubmitReportCard{
		Report: viperTypes.ViperQoSReport{
			LatencyScore:      sdk.NewDecWithPrec(8, 1),
			AvailabilityScore: sdk.NewDecWithPrec(7, 1),
			ReliabilityScore:  sdk.NewDecWithPrec(9, 1),
		},
		FishermanAddress: getRandomValidatorAddress(),
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{"Test RelayReward", fields{keeper: keeper},
			args{
				ctx:        context,
				baseReward: base,
				relays:     numRelays,
				validator1: stakedValidator,
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.fields.keeper
			ctx := tt.args.ctx
			p := k.providerKeeper.Provider(ctx, tt.args.ctx.BlockHeader().ApplicationAddress)
			p1 := p.(providersTypes.Provider)
			k.RewardForRelays(tt.args.ctx, reportCard, sdk.NewInt(tt.args.relays), tt.args.validator1.GetAddress(), p1)
			acc := k.GetAccount(ctx, tt.args.validator1.GetAddress())
			assert.False(t, acc.Coins.IsZero())
			assert.True(t, acc.Coins.IsEqual(sdk.NewCoins(sdk.NewCoin("uvipr", tt.args.baseReward))))
			k.RewardForRelays(tt.args.ctx, reportCard, sdk.NewInt(tt.args.relays), tt.args.validator2.GetAddress(), p1)
			acc = k.GetAccount(ctx, tt.args.validator2.GetAddress())
			assert.False(t, acc.Coins.IsZero())
			assert.True(t, acc.Coins.IsEqual(sdk.NewCoins(sdk.NewCoin("uvipr", tt.args.baseReward.Mul(sdk.NewInt(2))))))
			k.RewardForRelays(tt.args.ctx, reportCard, sdk.NewInt(tt.args.relays), tt.args.validator3.GetAddress(), p1)
			acc = k.GetAccount(ctx, tt.args.validator3.GetAddress())
			assert.False(t, acc.Coins.IsZero())
			assert.True(t, acc.Coins.IsEqual(sdk.NewCoins(sdk.NewCoin("uvipr", tt.args.baseReward.Mul(sdk.NewInt(3))))))
			k.RewardForRelays(tt.args.ctx, reportCard, sdk.NewInt(tt.args.relays), tt.args.validator4.GetAddress(), p1)
			acc = k.GetAccount(ctx, tt.args.validator4.GetAddress())
			assert.False(t, acc.Coins.IsZero())
			assert.True(t, acc.Coins.IsEqual(sdk.NewCoins(sdk.NewCoin("uvipr", tt.args.baseReward.Mul(sdk.NewInt(4))))))
		})
	}
}

func TestKeeper_checkCheckCeiling(t *testing.T) {
	type fields struct {
		keeper Keeper
	}
	type args struct {
		ctx        sdk.Ctx
		baseReward sdk.BigInt
		relays     int64
		validator1 types.Validator
		validator2 types.Validator
	}

	context, _, keeper := createTestInput(t, true)
	context = context.WithBlockHeight(3)
	p := keeper.GetParams(context)
	keeper.SetParams(context, p)

	stakedValidator := getStakedValidator()

	numRelays := int64(10000)
	base := keeper.TokenRewardFactor(context)

	keeper.SetValidator(context, stakedValidator)
	reportCard := viperTypes.MsgSubmitReportCard{
		Report: viperTypes.ViperQoSReport{
			LatencyScore:      sdk.NewDecWithPrec(8, 1),
			AvailabilityScore: sdk.NewDecWithPrec(7, 1),
			ReliabilityScore:  sdk.NewDecWithPrec(9, 1),
		},
		FishermanAddress: getRandomValidatorAddress(),
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{"Test RelayReward", fields{keeper: keeper},
			args{
				ctx:        context,
				baseReward: base,
				relays:     numRelays,
				validator1: stakedValidator,
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.fields.keeper
			ctx := tt.args.ctx
			p := k.providerKeeper.Provider(ctx, tt.args.ctx.BlockHeader().ApplicationAddress)
			p1 := p.(providersTypes.Provider)
			k.RewardForRelays(tt.args.ctx, reportCard, sdk.NewInt(tt.args.relays), tt.args.validator1.GetAddress(), p1)
			acc := k.GetAccount(ctx, tt.args.validator1.GetAddress())
			assert.False(t, acc.Coins.IsZero())
			assert.True(t, acc.Coins.IsEqual(sdk.NewCoins(sdk.NewCoin("uvipr", tt.args.baseReward))))
			k.RewardForRelays(tt.args.ctx, reportCard, sdk.NewInt(tt.args.relays), tt.args.validator2.GetAddress(), p1)
			acc = k.GetAccount(ctx, tt.args.validator2.GetAddress())
			assert.False(t, acc.Coins.IsZero())
			assert.True(t, acc.Coins.IsEqual(sdk.NewCoins(sdk.NewCoin("uvipr", tt.args.baseReward))))
		})
	}
}

func TestKeeper_rewardFromRelaysEXP(t *testing.T) {
	type fields struct {
		keeper Keeper
	}
	type args struct {
		ctx        sdk.Ctx
		validator1 types.Validator
		validator2 types.Validator
		validator3 types.Validator
		validator4 types.Validator
	}

	context, _, keeper := createTestInput(t, true)
	context = context.WithBlockHeight(3)
	p := keeper.GetParams(context)
	keeper.SetParams(context, p)

	stakedValidator := getStakedValidator()

	keeper.SetValidator(context, stakedValidator)
	reportCard := viperTypes.MsgSubmitReportCard{
		Report: viperTypes.ViperQoSReport{
			LatencyScore:      sdk.NewDecWithPrec(8, 1),
			AvailabilityScore: sdk.NewDecWithPrec(7, 1),
			ReliabilityScore:  sdk.NewDecWithPrec(9, 1),
		},
		FishermanAddress: getRandomValidatorAddress(),
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{"Test RelayReward", fields{keeper: keeper},
			args{
				ctx:        context,
				validator1: stakedValidator,
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.fields.keeper
			ctx := tt.args.ctx
			p := k.providerKeeper.Provider(ctx, tt.args.ctx.BlockHeader().ApplicationAddress)
			p1 := p.(providersTypes.Provider)
			k.RewardForRelays(tt.args.ctx, reportCard, sdk.NewInt(1000), tt.args.validator1.GetAddress(), p1)
			acc := k.GetAccount(ctx, tt.args.validator1.GetAddress())
			assert.False(t, acc.Coins.IsZero())
			assert.True(t, acc.Coins.IsEqual(sdk.NewCoins(sdk.NewCoin("uvipr", sdk.NewInt(800000)))))
			k.RewardForRelays(tt.args.ctx, reportCard, sdk.NewInt(1000), tt.args.validator2.GetAddress(), p1)
			acc = k.GetAccount(ctx, tt.args.validator2.GetAddress())
			assert.False(t, acc.Coins.IsZero())
			assert.True(t, acc.Coins.IsEqual(sdk.NewCoins(sdk.NewCoin("uvipr", sdk.NewInt(1131372)))))
			k.RewardForRelays(tt.args.ctx, reportCard, sdk.NewInt(1000), tt.args.validator3.GetAddress(), p1)
			acc = k.GetAccount(ctx, tt.args.validator3.GetAddress())
			assert.False(t, acc.Coins.IsZero())
			assert.True(t, acc.Coins.IsEqual(sdk.NewCoins(sdk.NewCoin("uvipr", sdk.NewInt(1385641)))))
			k.RewardForRelays(tt.args.ctx, reportCard, sdk.NewInt(1000), tt.args.validator4.GetAddress(), p1)
			acc = k.GetAccount(ctx, tt.args.validator4.GetAddress())
			assert.False(t, acc.Coins.IsZero())
			assert.True(t, acc.Coins.IsEqual(sdk.NewCoins(sdk.NewCoin("uvipr", sdk.NewInt(1600000)))))
		})
	}
}
