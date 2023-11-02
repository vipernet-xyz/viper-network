package keeper

import (
	"reflect"
	"testing"
	"time"

	"github.com/vipernet-xyz/viper-network/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/servicers/types"

	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestKeeper_FinishUnstakingValidator(t *testing.T) {
	type fields struct {
		keeper Keeper
	}

	type args struct {
		ctx       sdk.Ctx
		validator types.Validator
	}

	validator := getStakedValidator()
	validator.StakedTokens = sdk.NewInt(0)
	context, _, keeper := createTestInput(t, true)

	tests := []struct {
		name   string
		fields fields
		args   args
		want   sdk.Error
	}{
		{"Test FinishUnstakingValidator", fields{keeper: keeper}, args{
			ctx:       context,
			validator: validator,
		}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.fields.keeper
			// todo: add more tests scenarios
			k.FinishUnstakingValidator(tt.args.ctx, tt.args.validator)
		})
	}
}

func TestValidatorStateChange_EditAndValidateStakeValidator(t *testing.T) {
	stakeAmount := sdk.NewInt(100000000000)
	accountAmount := sdk.NewInt(1000000000000).Add(stakeAmount)
	bumpStakeAmount := sdk.NewInt(1000000000000)
	newChains := []string{"0021"}
	newGeoZone := []string{"0003"}
	val := getUnstakedValidator()
	val.StakedTokens = sdk.ZeroInt()
	val.OutputAddress = val.Address
	// updatedStakeAmount
	updateStakeAmountProvider := val
	updateStakeAmountProvider.StakedTokens = bumpStakeAmount
	// updatedStakeAmountFail
	updateStakeAmountProviderFail := val
	updateStakeAmountProviderFail.StakedTokens = stakeAmount.Sub(sdk.OneInt())
	// updatedStakeAmountNotEnoughCoins
	notEnoughCoinsAccount := stakeAmount
	// updateChains
	updateChainsVal := val
	updateChainsVal.StakedTokens = stakeAmount
	updateChainsVal.Chains = newChains
	// updateServiceURL
	updateServiceURL := val
	updateServiceURL.StakedTokens = stakeAmount
	updateServiceURL.Chains = newChains
	updateServiceURL.ServiceURL = "https://newServiceUrl.com"
	// updateGeoZones
	updateGeoZonesVal := val
	updateGeoZonesVal.StakedTokens = stakeAmount
	updateGeoZonesVal.GeoZone = newGeoZone
	// nil output addresss
	nilOutputAddress := val
	nilOutputAddress.OutputAddress = nil
	nilOutputAddress.StakedTokens = stakeAmount
	//same provider no change no fail
	updateNothingval := val
	updateNothingval.StakedTokens = stakeAmount
	//new staked amount doesn't push into the next bin
	fail := val
	fail.StakedTokens = sdk.NewInt(29999000000)
	//New staked amount does push into the next bin
	passNextBin := val
	passNextBin.StakedTokens = sdk.NewInt(30001000000)
	//All updates should pass above the ceiling
	passAboveCeil := val
	passAboveCeil.StakedTokens = sdk.NewInt(60000000000).Add(sdk.OneInt())

	tests := []struct {
		name          string
		accountAmount sdk.BigInt
		origProvider  types.Validator
		amount        sdk.BigInt
		want          types.Validator
		err           sdk.Error
		Edit          bool
	}{
		{
			name:          "edit stake amount of existing validator",
			accountAmount: accountAmount,
			origProvider:  val,
			amount:        stakeAmount,
			want:          updateStakeAmountProvider,
			Edit:          true,
		},
		{
			name:          "FAIL edit stake amount of existing validator",
			accountAmount: accountAmount,
			origProvider:  val,
			amount:        stakeAmount,
			want:          updateStakeAmountProviderFail,
			err:           types.ErrMinimumEditStake("pos"),
			Edit:          false,
		},
		{
			name:          "edit stake the chains of the validator",
			accountAmount: accountAmount,
			origProvider:  val,
			amount:        stakeAmount,
			want:          updateChainsVal,
			Edit:          false,
		},
		{
			name:          "edit stake the serviceurl of the validator",
			accountAmount: accountAmount,
			origProvider:  val,
			amount:        stakeAmount,
			want:          updateChainsVal,
			Edit:          false,
		},
		{
			name:          "FAIL not enough coins to bump stake amount of existing validator",
			accountAmount: notEnoughCoinsAccount,
			origProvider:  val,
			amount:        stakeAmount,
			want:          updateStakeAmountProvider,
			err:           types.ErrNotEnoughCoins("pos"),
			Edit:          false,
		},
		{
			name:          "update nothing for the validator",
			accountAmount: accountAmount,
			origProvider:  val,
			amount:        stakeAmount,
			want:          updateNothingval,
			Edit:          false,
		},
		{
			name:          " not enough to bump bin",
			accountAmount: accountAmount,
			origProvider:  val,
			amount:        sdk.NewInt(15001000000),
			want:          fail,
			err:           types.ErrSameBinEditStake("pos"),
			Edit:          true,
		},
		{
			name:          " update to next bin",
			accountAmount: accountAmount,
			origProvider:  val,
			amount:        sdk.NewInt(15001000000),
			want:          passNextBin,
			Edit:          true,
		},
		{
			name:          " above ceil",
			accountAmount: accountAmount,
			origProvider:  val,
			amount:        sdk.NewInt(60000000000),
			want:          passAboveCeil,
			Edit:          true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// test setup
			codec.UpgradeHeight = -1
			if tt.Edit {
				codec.UpgradeFeatureMap[codec.VEDITKey] = -1
			} else {
				codec.UpgradeFeatureMap[codec.VEDITKey] = 0
			}
			context, _, keeper := createTestInput(t, true)
			coins := sdk.NewCoins(sdk.NewCoin(keeper.StakeDenom(context), tt.accountAmount))
			err := keeper.AccountKeeper.MintCoins(context, types.StakedPoolName, coins)
			if err != nil {
				t.Fail()
			}
			err = keeper.AccountKeeper.SendCoinsFromModuleToAccount(context, types.StakedPoolName, tt.origProvider.Address, coins)
			if err != nil {
				t.Fail()
			}
			err = keeper.StakeValidator(context, tt.origProvider, tt.amount, tt.origProvider.PublicKey)
			if err != nil {
				t.Fail()
			}
			// test begins here
			err = keeper.ValidateValidatorStaking(context, tt.want, tt.want.StakedTokens, sdk.Address(tt.origProvider.PublicKey.Address()))
			if err != nil {
				if tt.err.Error() != err.Error() {
					t.Fatalf("Got error %s wanted error %s", err, tt.err)
				}
				return
			}
			// edit stake
			_ = keeper.StakeValidator(context, tt.want, tt.want.StakedTokens, tt.want.PublicKey)
			tt.want.Status = sdk.Staked
			// see if the changes stuck
			got, _ := keeper.GetValidator(context, tt.origProvider.Address)
			if !got.Equals(tt.want) {
				t.Fatalf("Got provider %s\nWanted provider %s", got.String(), tt.want.String())
			}
		})

	}
}
func TestValidatorStateChange_EditAndValidateStakeValidatorAfterNonCustodialUpgrade(t *testing.T) {
	stakeAmount := sdk.NewInt(100000000000)
	accountAmount := sdk.NewInt(1000000000000).Add(stakeAmount)
	bumpStakeAmount := sdk.NewInt(1000000000000)
	newChains := []string{"0021"}
	newGeoZone := []string{"0003"}
	val := getUnstakedValidator()
	val.StakedTokens = sdk.ZeroInt()
	val.OutputAddress = val.Address
	// updatedStakeAmount
	updateStakeAmountProvider := val
	updateStakeAmountProvider.StakedTokens = bumpStakeAmount
	// updatedStakeAmountFail
	updateStakeAmountProviderFail := val
	updateStakeAmountProviderFail.StakedTokens = stakeAmount.Sub(sdk.OneInt())
	// updatedStakeAmountNotEnoughCoins
	notEnoughCoinsAccount := stakeAmount
	// updateChains
	updateChainsVal := val
	updateChainsVal.StakedTokens = stakeAmount
	updateChainsVal.Chains = newChains
	// updateServiceURL
	updateServiceURL := val
	updateServiceURL.StakedTokens = stakeAmount
	updateServiceURL.Chains = newChains
	updateServiceURL.ServiceURL = "https://newServiceUrl.com"
	// updateGeoZones
	updateGeoZonesVal := val
	updateGeoZonesVal.StakedTokens = stakeAmount
	updateGeoZonesVal.GeoZone = newGeoZone
	// nil output addresss
	nilOutputAddress := val
	nilOutputAddress.OutputAddress = nil
	nilOutputAddress.StakedTokens = stakeAmount
	//same provider no change no fail
	updateNothingval := val
	updateNothingval.StakedTokens = stakeAmount
	tests := []struct {
		name          string
		accountAmount sdk.BigInt
		origProvider  types.Validator
		amount        sdk.BigInt
		want          types.Validator
		err           sdk.Error
	}{
		{
			name:          "edit stake amount of existing validator",
			accountAmount: accountAmount,
			origProvider:  val,
			amount:        stakeAmount,
			want:          updateStakeAmountProvider,
		},
		{
			name:          "FAIL edit stake amount of existing validator",
			accountAmount: accountAmount,
			origProvider:  val,
			amount:        stakeAmount,
			want:          updateStakeAmountProviderFail,
			err:           types.ErrMinimumEditStake("pos"),
		},
		{
			name:          "edit stake the chains of the validator",
			accountAmount: accountAmount,
			origProvider:  val,
			amount:        stakeAmount,
			want:          updateChainsVal,
		},
		{
			name:          "edit stake the serviceurl of the validator",
			accountAmount: accountAmount,
			origProvider:  val,
			amount:        stakeAmount,
			want:          updateChainsVal,
		},
		{
			name:          "FAIL not enough coins to bump stake amount of existing validator",
			accountAmount: notEnoughCoinsAccount,
			origProvider:  val,
			amount:        stakeAmount,
			want:          updateStakeAmountProvider,
			err:           types.ErrNotEnoughCoins("pos"),
		},
		{
			name:          "FAIL nil output address",
			accountAmount: notEnoughCoinsAccount,
			origProvider:  val,
			amount:        stakeAmount,
			want:          nilOutputAddress,
			err:           types.ErrNilOutputAddr("pos"),
		},
		{
			name:          "update nothing for the validator",
			accountAmount: accountAmount,
			origProvider:  val,
			amount:        stakeAmount,
			want:          updateNothingval,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// test setup
			codec.TestMode = -3
			codec.UpgradeHeight = -1
			context, _, keeper := createTestInput(t, true)
			coins := sdk.NewCoins(sdk.NewCoin(keeper.StakeDenom(context), tt.accountAmount))
			err := keeper.AccountKeeper.MintCoins(context, types.StakedPoolName, coins)
			if err != nil {
				t.Fail()
			}
			err = keeper.AccountKeeper.SendCoinsFromModuleToAccount(context, types.StakedPoolName, tt.origProvider.Address, coins)
			if err != nil {
				t.Fail()
			}
			err = keeper.StakeValidator(context, tt.origProvider, tt.amount, tt.origProvider.PublicKey)
			if err != nil {
				t.Fail()
			}
			// test begins here
			err = keeper.ValidateValidatorStaking(context, tt.want, tt.want.StakedTokens, sdk.Address(tt.origProvider.PublicKey.Address()))
			if err != nil {
				if tt.err.Error() != err.Error() {
					t.Fatalf("Got error %s wanted error %s", err, tt.err)
				}
				return
			}
			// edit stake
			_ = keeper.StakeValidator(context, tt.want, tt.want.StakedTokens, tt.want.PublicKey)
			tt.want.Status = sdk.Staked
			// see if the changes stuck
			got, _ := keeper.GetValidator(context, tt.origProvider.Address)
			if !got.Equals(tt.want) {
				t.Fatalf("Got provider %s\nWanted provider %s", got.String(), tt.want.String())
			}
		})

	}
}

func TestKeeper_JailValidator(t *testing.T) {
	type fields struct {
		keeper Keeper
	}
	type args struct {
		ctx  sdk.Ctx
		addr sdk.Address
	}

	validator := getStakedValidator()
	context, _, keeper := createTestInput(t, true)
	keeper.SetValidator(context, validator)

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{"Test JailValidator", fields{keeper: keeper}, args{
			ctx:  context,
			addr: validator.GetAddress(),
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.fields.keeper
			k.JailValidator(tt.args.ctx, tt.args.addr)
		})
	}
}

func TestKeeper_ReleaseWaitingValidators(t *testing.T) {
	type fields struct {
		keeper Keeper
	}
	type args struct {
		ctx sdk.Ctx
	}

	validator := getUnstakingValidator()
	context, _, keeper := createTestInput(t, true)

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{"Test ReleaseWaitingValidators", fields{keeper: keeper}, args{ctx: context}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.fields.keeper
			k.SetWaitingValidator(tt.args.ctx, validator)
			k.ReleaseWaitingValidators(tt.args.ctx)
		})
	}
}

func TestKeeper_StakeValidator(t *testing.T) {
	type fields struct {
		keeper Keeper
	}
	type args struct {
		ctx       sdk.Ctx
		validator types.Validator
		amount    sdk.BigInt
	}

	validator := getStakedValidator()
	context, _, keeper := createTestInput(t, true)

	tests := []struct {
		name   string
		fields fields
		args   args
		want   sdk.Error
	}{
		{"Test StakeValidator", fields{keeper: keeper}, args{
			ctx:       context,
			validator: validator,
			amount:    sdk.ZeroInt(),
		}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.fields.keeper
			if got := k.StakeValidator(tt.args.ctx, tt.args.validator, tt.args.amount, tt.args.validator.PublicKey); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StakeValidator() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeeper_UnjailValidator(t *testing.T) {
	type fields struct {
		keeper Keeper
	}
	type args struct {
		ctx  sdk.Ctx
		addr sdk.Address
	}
	validator := getStakedValidator()
	context, _, keeper := createTestInput(t, true)
	validator.Jailed = true
	keeper.SetValidator(context, validator)

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{"Test UnjailValidator", fields{keeper: keeper}, args{
			ctx:  context,
			addr: validator.GetAddress(),
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.fields.keeper
			k.UnjailValidator(tt.args.ctx, tt.args.addr)
		})
	}
}

func TestKeeper_UpdateTendermintValidators(t *testing.T) {
	type fields struct {
		keeper Keeper
	}
	type args struct {
		ctx sdk.Ctx
	}

	//validator := getStakedValidator()
	context, _, keeper := createTestInput(t, true)

	tests := []struct {
		name        string
		fields      fields
		args        args
		wantUpdates []abci.ValidatorUpdate
	}{
		{"Test UpdateTenderMintValidators", fields{keeper: keeper}, args{ctx: context},
			[]abci.ValidatorUpdate{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.fields.keeper
			if gotUpdates := k.UpdateTendermintValidators(tt.args.ctx); !assert.True(t, len(gotUpdates) == len(tt.wantUpdates)) {
				t.Errorf("UpdateTendermintValidators() = %v, want %v", gotUpdates, tt.wantUpdates)
			}
		})
	}
}

func TestKeeper_ValidateValidatorBeginUnstaking(t *testing.T) {
	type fields struct {
		keeper Keeper
	}
	type args struct {
		ctx       sdk.Ctx
		validator types.Validator
	}

	validator := getStakedValidator()
	context, _, keeper := createTestInput(t, true)

	tests := []struct {
		name   string
		fields fields
		args   args
		want   sdk.Error
	}{
		{"Test ValidateValidatorBeginUnstaking", fields{keeper: keeper}, args{
			ctx:       context,
			validator: validator,
		}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.fields.keeper
			if got := k.ValidateValidatorBeginUnstaking(tt.args.ctx, tt.args.validator); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidateValidatorBeginUnstaking() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeeper_ValidateValidatorFinishUnstaking(t *testing.T) {
	type fields struct {
		keeper Keeper
	}
	type args struct {
		ctx       sdk.Ctx
		validator types.Validator
	}

	validator := getUnstakingValidator()
	context, _, keeper := createTestInput(t, true)

	tests := []struct {
		name   string
		fields fields
		args   args
		want   sdk.Error
	}{
		{"Test ValidateValidatorFinishUnstaking", fields{keeper: keeper}, args{
			ctx:       context,
			validator: validator,
		}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.fields.keeper
			if got := k.ValidateValidatorFinishUnstaking(tt.args.ctx, tt.args.validator); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidateValidatorFinishUnstaking() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeeper_ValidateValidatorStaking(t *testing.T) {
	type fields struct {
		keeper Keeper
	}
	type args struct {
		ctx       sdk.Ctx
		validator types.Validator
		amount    sdk.BigInt
	}

	validator := getUnstakedValidator()
	context, _, keeper := createTestInput(t, true)

	tests := []struct {
		name   string
		fields fields
		args   args
		want   sdk.Error
	}{
		{
			name:   "Test ValidateValidatorStaking - Not Enough Coins",
			fields: fields{keeper: keeper},
			args: args{
				ctx:       context,
				validator: validator,
				amount:    sdk.NewInt(1000000), // This should be greater than the balance in the account
			},
			want: types.ErrNotEnoughCoins(types.ModuleName),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.fields.keeper
			codec.TestMode = -2

			// Setting balance to a value less than the staking amount but greater than the minimum stake.
			k.AccountKeeper.SetCoins(tt.args.ctx, tt.args.validator.Address, sdk.NewCoins(sdk.NewCoin(k.StakeDenom(tt.args.ctx), sdk.NewInt(500000))))

			if got := k.ValidateValidatorStaking(tt.args.ctx, tt.args.validator, tt.args.amount, sdk.Address(tt.args.validator.PublicKey.Address())); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidateValidatorStaking() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeeper_WaitToBeginUnstakingValidator(t *testing.T) {
	type fields struct {
		keeper Keeper
	}
	type args struct {
		ctx       sdk.Ctx
		validator types.Validator
	}

	validator := getStakedValidator()
	context, _, keeper := createTestInput(t, true)

	tests := []struct {
		name   string
		fields fields
		args   args
		want   sdk.Error
	}{
		{"Test WaitToBeginUnstakingValidator", fields{keeper: keeper}, args{
			ctx:       context,
			validator: validator,
		}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := tt.fields.keeper
			if got := k.WaitToBeginUnstakingValidator(tt.args.ctx, tt.args.validator); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WaitToBeginUnstakingValidator() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKeeper_ValidateUnjailMessage(t *testing.T) {
	type args struct {
		ctx sdk.Ctx
		k   Keeper
		v   types.Validator
		msg types.MsgUnjail
	}
	unauthSigner := getRandomValidatorAddress()
	validator := getStakedValidator()
	validator.Jailed = true
	validator.OutputAddress = getRandomValidatorAddress()
	validatorNoOuptut := validator
	validatorNoOuptut.OutputAddress = nil
	context, _, keeper := createTestInput(t, true)
	msgUnjailAuthorizedByValidator := types.MsgUnjail{
		ValidatorAddr: validator.Address,
		Signer:        validator.Address,
	}
	msgUnjailAuthorizedByOutput := types.MsgUnjail{
		ValidatorAddr: validator.Address,
		Signer:        validator.OutputAddress,
	}
	msgUnjailUnauthorizedSigner := types.MsgUnjail{
		ValidatorAddr: validator.Address,
		Signer:        unauthSigner,
	}
	tests := []struct {
		name string
		args args
		want error
	}{
		{"Test ValidateUnjailMessage With Output Address & AuthorizedByValidator", args{
			ctx: context,
			k:   keeper,
			v:   validator,
			msg: msgUnjailAuthorizedByValidator,
		}, nil},
		{"Test ValidateUnjailMessage With Output Address & AuthorizedByOutput", args{
			ctx: context,
			k:   keeper,
			v:   validator,
			msg: msgUnjailAuthorizedByOutput,
		}, nil},
		{"Test ValidateUnjailMessage Without Output Address & AuthorizedByValidator", args{
			ctx: context,
			k:   keeper,
			v:   validatorNoOuptut,
			msg: msgUnjailAuthorizedByValidator,
		}, nil},
		{"Test ValidateUnjailMessage Without Output Address & AuthroizedByOutput", args{
			ctx: context,
			k:   keeper,
			v:   validatorNoOuptut,
			msg: msgUnjailAuthorizedByOutput,
		}, types.ErrUnauthorizedSigner("pos")},
		{"Test ValidateUnjailMessage Without Output Address & Unauthorized", args{
			ctx: context,
			k:   keeper,
			v:   validatorNoOuptut,
			msg: msgUnjailUnauthorizedSigner,
		}, types.ErrUnauthorizedSigner("pos")},

		{"Test ValidateUnjailMessage With Output Address & Unauthorized", args{
			ctx: context,
			k:   keeper,
			v:   validator,
			msg: msgUnjailUnauthorizedSigner,
		}, types.ErrUnauthorizedSigner("pos")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keeper.SetValidator(tt.args.ctx, tt.args.v)
			keeper.SetValidatorSigningInfo(tt.args.ctx, tt.args.v.Address, types.ValidatorSigningInfo{
				Address:             tt.args.v.Address,
				StartHeight:         0,
				Index:               0,
				JailedUntil:         time.Time{},
				MissedBlocksCounter: 0,
				JailedBlocksCounter: 0,
			})
			_, err := tt.args.k.ValidateUnjailMessage(tt.args.ctx, tt.args.msg)
			assert.Equal(t, tt.want, err)
		})
	}
}

func TestKeeper_ValidatePauseNodeMessage(t *testing.T) {
	type args struct {
		ctx sdk.Ctx
		k   Keeper
		v   types.Validator
		msg types.MsgPause
	}

	unauthSigner := getRandomValidatorAddress()
	validator := getStakedValidator()
	validator.Paused = false
	validator.OutputAddress = getRandomValidatorAddress()

	validatorNoOutput := validator
	validatorNoOutput.Paused = false
	validatorNoOutput.OutputAddress = nil

	context, _, keeper := createTestInput(t, true)

	msgPauseAuthorizedByValidator := types.MsgPause{
		ValidatorAddr: validator.Address,
		Signer:        validator.Address,
	}

	msgPauseAuthorizedByOutput := types.MsgPause{
		ValidatorAddr: validator.Address,
		Signer:        validator.OutputAddress,
	}

	msgPauseUnauthorizedSigner := types.MsgPause{
		ValidatorAddr: validator.Address,
		Signer:        unauthSigner,
	}

	tests := []struct {
		name string
		args args
		want sdk.Error
	}{
		{"Test ValidatePauseNodeMessage With Output Address & AuthorizedByValidator", args{
			ctx: context,
			k:   keeper,
			v:   validator,
			msg: msgPauseAuthorizedByValidator,
		}, nil},
		{"Test ValidatePauseNodeMessage With Output Address & AuthorizedByOutput", args{
			ctx: context,
			k:   keeper,
			v:   validator,
			msg: msgPauseAuthorizedByOutput,
		}, nil},
		{"Test ValidatePauseNodeMessage Without Output Address & AuthorizedByValidator", args{
			ctx: context,
			k:   keeper,
			v:   validatorNoOutput,
			msg: msgPauseAuthorizedByValidator,
		}, nil},
		{"Test ValidatePauseNodeMessage Without Output Address & AuthorizedByOutput", args{
			ctx: context,
			k:   keeper,
			v:   validatorNoOutput,
			msg: msgPauseAuthorizedByOutput,
		}, types.ErrUnauthorizedSigner("pos")},
		{"Test ValidatePauseNodeMessage Without Output Address & Unauthorized", args{
			ctx: context,
			k:   keeper,
			v:   validatorNoOutput,
			msg: msgPauseUnauthorizedSigner,
		}, types.ErrUnauthorizedSigner("pos")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keeper.SetValidator(tt.args.ctx, tt.args.v)
			err := tt.args.k.ValidatePauseNodeMessage(tt.args.ctx, tt.args.msg)
			assert.Equal(t, tt.want, err)
		})
	}
}

func TestKeeper_ValidateUnpauseNodeMessage(t *testing.T) {
	type args struct {
		ctx sdk.Ctx
		k   Keeper
		v   types.Validator
		msg types.MsgUnpause
	}
	unauthSigner := getRandomValidatorAddress()
	validator := getStakedValidator()
	validator.Paused = true
	validator.OutputAddress = getRandomValidatorAddress()
	validatorNoOutput := validator
	validatorNoOutput.Paused = true
	validatorNoOutput.OutputAddress = nil
	context, _, keeper := createTestInput(t, true)
	msgUnpauseAuthorizedByValidator := types.MsgUnpause{
		ValidatorAddr: validator.Address,
		Signer:        validator.Address,
	}
	msgUnpauseAuthorizedByOutput := types.MsgUnpause{
		ValidatorAddr: validator.Address,
		Signer:        validator.OutputAddress,
	}
	msgUnpauseUnauthorizedSigner := types.MsgUnpause{
		ValidatorAddr: validator.Address,
		Signer:        unauthSigner,
	}
	tests := []struct {
		name string
		args args
		want error
	}{
		{"Test ValidateUnpauseNodeMessage With Output Address & AuthorizedByValidator", args{
			ctx: context,
			k:   keeper,
			v:   validator,
			msg: msgUnpauseAuthorizedByValidator,
		}, nil},
		{"Test ValidateUnpauseNodeMessage With Output Address & AuthorizedByOutput", args{
			ctx: context,
			k:   keeper,
			v:   validator,
			msg: msgUnpauseAuthorizedByOutput,
		}, nil},
		{"Test ValidateUnpauseNodeMessage Without Output Address & AuthorizedByValidator", args{
			ctx: context,
			k:   keeper,
			v:   validatorNoOutput,
			msg: msgUnpauseAuthorizedByValidator,
		}, nil},
		{"Test ValidateUnpauseNodeMessage Without Output Address & AuthorizedByOutput", args{
			ctx: context,
			k:   keeper,
			v:   validatorNoOutput,
			msg: msgUnpauseAuthorizedByOutput,
		}, types.ErrUnauthorizedSigner("pos")},
		{"Test ValidateUnpauseNodeMessage Without Output Address & Unauthorized", args{
			ctx: context,
			k:   keeper,
			v:   validatorNoOutput,
			msg: msgUnpauseUnauthorizedSigner,
		}, types.ErrUnauthorizedSigner("pos")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keeper.SetValidator(tt.args.ctx, tt.args.v)
			keeper.SetValidatorSigningInfo(tt.args.ctx, tt.args.v.Address, types.ValidatorSigningInfo{
				Address:             tt.args.v.Address,
				StartHeight:         0,
				Index:               0,
				JailedUntil:         time.Time{},
				MissedBlocksCounter: 0,
				JailedBlocksCounter: 0,
			})
			_, err := tt.args.k.ValidateUnpauseNodeMessage(tt.args.ctx, tt.args.msg)
			assert.Equal(t, tt.want, err)
		})
	}
}
