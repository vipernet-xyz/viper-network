package keeper

import (
	"testing"

	"github.com/vipernet-xyz/viper-network/codec"

	sdk "github.com/vipernet-xyz/viper-network/types"
	appsTypes "github.com/vipernet-xyz/viper-network/x/apps/types"
	"github.com/vipernet-xyz/viper-network/x/nodes/types"

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

func TestSetandGetApplication(t *testing.T) {
	application := getStakedApplication()
	consAddress := application.GetAddress()

	tests := []struct {
		name            string
		args            args
		expectedAddress sdk.Address
	}{
		{
			name:            "can set the application",
			args:            args{consAddress: consAddress},
			expectedAddress: consAddress,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)

			keeper.SetApplicationKey(context, test.args.consAddress)
			receivedAddress := keeper.GetApplication(context)
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
		ctx              sdk.Context
		previousProposer sdk.Address
		application      sdk.Address
		Output           sdk.Address
		aOutput          sdk.Address
		Amount           sdk.BigInt
	}
	stakedValidator := getStakedValidator()
	stakedApplication := getStakedApplication()
	stakedValidator.OutputAddress = getRandomValidatorAddress()
	stakedApplication.Address = getRandomApplicationAddress()
	codec.UpgradeFeatureMap[codec.RSCALKey] = 0
	codec.TestMode = -3
	amount := sdk.NewInt(10000)
	fees := sdk.NewCoins(sdk.NewCoin("uvip", amount))
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
				application:      stakedApplication.GetAddress(),
				Output:           stakedValidator.OutputAddress,
				aOutput:          stakedApplication.Address,
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.fields.keeper
			ctx := tt.args.ctx
			k.blockReward(tt.args.ctx, tt.args.previousProposer)
			acc := k.GetAccount(ctx, tt.args.Output)
			acc1 := k.GetAccount(ctx, tt.args.aOutput)
			assert.False(t, acc.Coins.IsZero())
			assert.True(t, acc.Coins.IsEqual(sdk.NewCoins(sdk.NewCoin("uvip", sdk.NewInt(910)))))
			acc = k.GetAccount(ctx, tt.args.previousProposer)
			acc1 = k.GetAccount(ctx, tt.args.application)
			assert.True(t, acc.Coins.IsZero())
			assert.True(t, acc1.Coins.IsZero())
		})
	}

}

func getRandomApplicationAddress() sdk.Address {
	return sdk.Address(getRandomPubKey().Address())
}

func GetApplication() appsTypes.Application {
	pub := getRandomPubKey()
	return appsTypes.Application{
		Address:      sdk.Address(pub.Address()),
		StakedTokens: sdk.NewInt(100000000000),
		PublicKey:    pub,
		Jailed:       false,
		Status:       sdk.Staked,
		MaxRelays:    sdk.NewInt(100000000000),
		Chains:       []string{"0001"},
	}
}
func getStakedApplication() appsTypes.Application {
	return GetApplication()
}

func TestKeeper_rewardFromRelays(t *testing.T) {
	type fields struct {
		keeper Keeper
	}
	type args struct {
		ctx               sdk.Context
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
	codec.UpgradeFeatureMap[codec.RSCALKey] = 0
	context, _, keeper := createTestInput(t, true)
	keeper.SetValidator(context, stakedValidator)
	keeper.SetValidator(context, stakedValidatorNoOutput)
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{"Test RelayReward", fields{keeper: keeper},
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
			k.RewardForRelays(tt.args.ctx, sdk.NewInt(10000), tt.args.validator, tt.args.ctx.BlockHeader().ApplicationAddress)
			acc := k.GetAccount(ctx, tt.args.Output)
			acc1 := k.GetAccount(ctx, tt.args.ctx.BlockHeader().ApplicationAddress)
			assert.False(t, acc.Coins.IsZero())
			assert.True(t, acc.Coins.IsEqual(sdk.NewCoins(sdk.NewCoin("uvip", sdk.NewInt(8900000)))))
			assert.False(t, acc1.Coins.IsZero())
			assert.True(t, acc1.Coins.IsEqual(sdk.NewCoins(sdk.NewCoin("uvip", sdk.NewInt(200000)))))
			acc = k.GetAccount(ctx, tt.args.validator)
			assert.True(t, acc.Coins.IsZero())
			// no output now
			k.RewardForRelays(tt.args.ctx, sdk.NewInt(10000), tt.args.validatorNoOutput, tt.args.ctx.BlockHeader().ApplicationAddress)
			acc = k.GetAccount(ctx, tt.args.OutputNoOutput)
			assert.False(t, acc.Coins.IsZero())
			assert.True(t, acc.Coins.IsEqual(sdk.NewCoins(sdk.NewCoin("uvip", sdk.NewInt(8900000)))))
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
		ctx        sdk.Context
		baseReward sdk.BigInt
		relays     int64
		validator1 types.Validator
		validator2 types.Validator
		validator3 types.Validator
		validator4 types.Validator
	}

	codec.UpgradeFeatureMap[codec.RSCALKey] = 3
	context, _, keeper := createTestInput(t, true)
	context = context.WithBlockHeight(3)
	p := keeper.GetParams(context)
	p.ServicerStakeFloorMultiplier = types.DefaultServicerStakeFloorMultiplier
	p.ServicerStakeWeightMultiplier = types.DefaultServicerStakeWeightMultiplier
	p.ServicerStakeFloorMultiplierExponent = sdk.NewDec(1)
	p.ServicerStakeWeightCeiling = 60000000000
	keeper.SetParams(context, p)

	stakedValidatorBin1 := getStakedValidator()
	stakedValidatorBin1.StakedTokens = keeper.ServicerStakeFloorMultiplier(context)
	stakedValidatorBin2 := getStakedValidator()
	stakedValidatorBin2.StakedTokens = keeper.ServicerStakeFloorMultiplier(context).Mul(sdk.NewInt(2))
	stakedValidatorBin3 := getStakedValidator()
	stakedValidatorBin3.StakedTokens = keeper.ServicerStakeFloorMultiplier(context).Mul(sdk.NewInt(3))
	stakedValidatorBin4 := getStakedValidator()
	stakedValidatorBin4.StakedTokens = keeper.ServicerStakeFloorMultiplier(context).Mul(sdk.NewInt(4))

	numRelays := int64(10000)
	base := sdk.NewDec(1).Quo(keeper.ServicerStakeWeightMultiplier(context)).Mul(sdk.NewDec(numRelays)).Mul(sdk.NewDecWithPrec(89, 2)).TruncateInt().Mul(keeper.RelaysToTokensMultiplier(context))

	keeper.SetValidator(context, stakedValidatorBin1)
	keeper.SetValidator(context, stakedValidatorBin2)
	keeper.SetValidator(context, stakedValidatorBin3)
	keeper.SetValidator(context, stakedValidatorBin4)
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
				validator1: stakedValidatorBin1,
				validator2: stakedValidatorBin2,
				validator3: stakedValidatorBin3,
				validator4: stakedValidatorBin4,
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.fields.keeper
			ctx := tt.args.ctx
			k.RewardForRelays(tt.args.ctx, sdk.NewInt(tt.args.relays), tt.args.validator1.GetAddress(), tt.args.ctx.BlockHeader().ApplicationAddress)
			acc := k.GetAccount(ctx, tt.args.validator1.GetAddress())
			assert.False(t, acc.Coins.IsZero())
			assert.True(t, acc.Coins.IsEqual(sdk.NewCoins(sdk.NewCoin("uvip", tt.args.baseReward))))
			k.RewardForRelays(tt.args.ctx, sdk.NewInt(tt.args.relays), tt.args.validator2.GetAddress(), tt.args.ctx.BlockHeader().ApplicationAddress)
			acc = k.GetAccount(ctx, tt.args.validator2.GetAddress())
			assert.False(t, acc.Coins.IsZero())
			assert.True(t, acc.Coins.IsEqual(sdk.NewCoins(sdk.NewCoin("uvip", tt.args.baseReward.Mul(sdk.NewInt(2))))))
			k.RewardForRelays(tt.args.ctx, sdk.NewInt(tt.args.relays), tt.args.validator3.GetAddress(), tt.args.ctx.BlockHeader().ApplicationAddress)
			acc = k.GetAccount(ctx, tt.args.validator3.GetAddress())
			assert.False(t, acc.Coins.IsZero())
			assert.True(t, acc.Coins.IsEqual(sdk.NewCoins(sdk.NewCoin("uvip", tt.args.baseReward.Mul(sdk.NewInt(3))))))
			k.RewardForRelays(tt.args.ctx, sdk.NewInt(tt.args.relays), tt.args.validator4.GetAddress(), tt.args.ctx.BlockHeader().ApplicationAddress)
			acc = k.GetAccount(ctx, tt.args.validator4.GetAddress())
			assert.False(t, acc.Coins.IsZero())
			assert.True(t, acc.Coins.IsEqual(sdk.NewCoins(sdk.NewCoin("uvip", tt.args.baseReward.Mul(sdk.NewInt(4))))))
		})
	}
}

func TestKeeper_checkCheckCeiling(t *testing.T) {
	type fields struct {
		keeper Keeper
	}
	type args struct {
		ctx        sdk.Context
		baseReward sdk.BigInt
		relays     int64
		validator1 types.Validator
		validator2 types.Validator
	}

	codec.UpgradeFeatureMap[codec.RSCALKey] = 3
	context, _, keeper := createTestInput(t, true)
	context = context.WithBlockHeight(3)
	p := keeper.GetParams(context)
	p.ServicerStakeFloorMultiplier = types.DefaultServicerStakeFloorMultiplier
	p.ServicerStakeWeightMultiplier = types.DefaultServicerStakeWeightMultiplier
	p.ServicerStakeFloorMultiplierExponent = sdk.NewDec(1)
	p.ServicerStakeWeightCeiling = 15000000000
	keeper.SetParams(context, p)

	stakedValidatorBin1 := getStakedValidator()
	stakedValidatorBin1.StakedTokens = keeper.ServicerStakeFloorMultiplier(context)
	stakedValidatorBin2 := getStakedValidator()
	stakedValidatorBin2.StakedTokens = keeper.ServicerStakeFloorMultiplier(context).Mul(sdk.NewInt(2))

	numRelays := int64(10000)
	base := sdk.NewDec(1).Quo(keeper.ServicerStakeWeightMultiplier(context)).Mul(sdk.NewDec(numRelays)).Mul(sdk.NewDecWithPrec(89, 2)).TruncateInt().Mul(keeper.RelaysToTokensMultiplier(context))

	keeper.SetValidator(context, stakedValidatorBin1)
	keeper.SetValidator(context, stakedValidatorBin2)
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
				validator1: stakedValidatorBin1,
				validator2: stakedValidatorBin2,
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.fields.keeper
			ctx := tt.args.ctx
			k.RewardForRelays(tt.args.ctx, sdk.NewInt(tt.args.relays), tt.args.validator1.GetAddress(), tt.args.ctx.BlockHeader().ApplicationAddress)
			acc := k.GetAccount(ctx, tt.args.validator1.GetAddress())
			assert.False(t, acc.Coins.IsZero())
			assert.True(t, acc.Coins.IsEqual(sdk.NewCoins(sdk.NewCoin("uvip", tt.args.baseReward))))
			k.RewardForRelays(tt.args.ctx, sdk.NewInt(tt.args.relays), tt.args.validator2.GetAddress(), tt.args.ctx.BlockHeader().ApplicationAddress)
			acc = k.GetAccount(ctx, tt.args.validator2.GetAddress())
			assert.False(t, acc.Coins.IsZero())
			assert.True(t, acc.Coins.IsEqual(sdk.NewCoins(sdk.NewCoin("uvip", tt.args.baseReward))))
		})
	}
}

func TestKeeper_rewardFromRelaysEXP(t *testing.T) {
	type fields struct {
		keeper Keeper
	}
	type args struct {
		ctx        sdk.Context
		validator1 types.Validator
		validator2 types.Validator
		validator3 types.Validator
		validator4 types.Validator
	}

	codec.UpgradeFeatureMap[codec.RSCALKey] = 3
	context, _, keeper := createTestInput(t, true)
	context = context.WithBlockHeight(3)
	p := keeper.GetParams(context)
	p.ServicerStakeFloorMultiplier = types.DefaultServicerStakeFloorMultiplier
	p.ServicerStakeWeightMultiplier = types.DefaultServicerStakeWeightMultiplier
	p.ServicerStakeFloorMultiplierExponent = sdk.NewDecWithPrec(50, 2)
	p.ServicerStakeWeightMultiplier = sdk.NewDec(1)
	p.ServicerStakeWeightCeiling = 60000000000
	keeper.SetParams(context, p)

	stakedValidatorBin1 := getStakedValidator()
	stakedValidatorBin1.StakedTokens = keeper.ServicerStakeFloorMultiplier(context)
	stakedValidatorBin2 := getStakedValidator()
	stakedValidatorBin2.StakedTokens = keeper.ServicerStakeFloorMultiplier(context).Mul(sdk.NewInt(2))
	stakedValidatorBin3 := getStakedValidator()
	stakedValidatorBin3.StakedTokens = keeper.ServicerStakeFloorMultiplier(context).Mul(sdk.NewInt(3))
	stakedValidatorBin4 := getStakedValidator()
	stakedValidatorBin4.StakedTokens = keeper.ServicerStakeFloorMultiplier(context).Mul(sdk.NewInt(4))

	keeper.SetValidator(context, stakedValidatorBin1)
	keeper.SetValidator(context, stakedValidatorBin2)
	keeper.SetValidator(context, stakedValidatorBin3)
	keeper.SetValidator(context, stakedValidatorBin4)
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{"Test RelayReward", fields{keeper: keeper},
			args{
				ctx:        context,
				validator1: stakedValidatorBin1,
				validator2: stakedValidatorBin2,
				validator3: stakedValidatorBin3,
				validator4: stakedValidatorBin4,
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.fields.keeper
			ctx := tt.args.ctx
			k.RewardForRelays(tt.args.ctx, sdk.NewInt(1000), tt.args.validator1.GetAddress(), tt.args.ctx.BlockHeader().ApplicationAddress)
			acc := k.GetAccount(ctx, tt.args.validator1.GetAddress())
			assert.False(t, acc.Coins.IsZero())
			assert.True(t, acc.Coins.IsEqual(sdk.NewCoins(sdk.NewCoin("uvip", sdk.NewInt(890000)))))
			k.RewardForRelays(tt.args.ctx, sdk.NewInt(1000), tt.args.validator2.GetAddress(), tt.args.ctx.BlockHeader().ApplicationAddress)
			acc = k.GetAccount(ctx, tt.args.validator2.GetAddress())
			assert.False(t, acc.Coins.IsZero())
			assert.True(t, acc.Coins.IsEqual(sdk.NewCoins(sdk.NewCoin("uvip", sdk.NewInt(1258650)))))
			k.RewardForRelays(tt.args.ctx, sdk.NewInt(1000), tt.args.validator3.GetAddress(), tt.args.ctx.BlockHeader().ApplicationAddress)
			acc = k.GetAccount(ctx, tt.args.validator3.GetAddress())
			assert.False(t, acc.Coins.IsZero())
			assert.True(t, acc.Coins.IsEqual(sdk.NewCoins(sdk.NewCoin("uvip", sdk.NewInt(1541525)))))
			k.RewardForRelays(tt.args.ctx, sdk.NewInt(1000), tt.args.validator4.GetAddress(), tt.args.ctx.BlockHeader().ApplicationAddress)
			acc = k.GetAccount(ctx, tt.args.validator4.GetAddress())
			assert.False(t, acc.Coins.IsZero())
			assert.True(t, acc.Coins.IsEqual(sdk.NewCoins(sdk.NewCoin("uvip", sdk.NewInt(1780000)))))
		})
	}
}
