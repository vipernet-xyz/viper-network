package keeper

import (
	"reflect"
	"testing"

	"github.com/vipernet-xyz/viper-network/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/platforms/types"
)

func TestPlatformStateChange_ValidatePlatformlicaitonBeginUnstaking(t *testing.T) {
	tests := []struct {
		name     string
		platform types.Platform
		hasError bool
		want     interface{}
	}{
		{
			name:     "validates platform",
			platform: getStakedPlatform(),
			want:     nil,
		},
		{
			name:     "errors if platform not staked",
			platform: getUnstakedPlatform(),
			want:     types.ErrPlatformStatus("platforms"),
			hasError: true,
		},
		{
			name:     "validates platform",
			platform: getStakedPlatform(),
			hasError: true,
			want:     "should not hplatformen: platform trying to begin unstaking has less than the minimum stake",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			switch tt.hasError {
			case true:
				tt.platform.StakedTokens = sdk.NewInt(-1)
				_ = keeper.ValidatePlatformBeginUnstaking(context, tt.platform)
			default:
				if got := keeper.ValidatePlatformBeginUnstaking(context, tt.platform); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("PlatformStateChange.ValidatePlatformBeginUnstaking() = got %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestPlatformStateChange_ValidatePlatformlicaitonStaking(t *testing.T) {
	tests := []struct {
		name                 string
		platform             types.Platform
		panics               bool
		amount               sdk.BigInt
		stakedPlatformsCount int
		isAfterUpgrade       bool
		want                 interface{}
	}{
		{
			name:                 "validates platform",
			stakedPlatformsCount: 0,
			platform:             getUnstakedPlatform(),
			amount:               sdk.NewInt(1000000),
			want:                 nil,
		},
		{
			name:                 "errors if below minimum stake",
			platform:             getUnstakedPlatform(),
			stakedPlatformsCount: 0,
			amount:               sdk.NewInt(0),
			want:                 types.ErrMinimumStake("platforms"),
		},
		{
			name:                 "errors bank does not have enough coins",
			platform:             getUnstakedPlatform(),
			stakedPlatformsCount: 0,
			amount:               sdk.NewInt(1000000000000000000),
			want:                 types.ErrNotEnoughCoins("platforms"),
		},
		{
			name:                 "errors if max platforms hit",
			platform:             getUnstakedPlatform(),
			stakedPlatformsCount: 5,
			amount:               sdk.NewInt(1000000),
			want:                 types.ErrMaxPlatforms("platforms"),
			isAfterUpgrade:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			p := keeper.GetParams(context)
			p.MaxPlatforms = 5
			keeper.SetParams(context, p)
			for i := 0; i < tt.stakedPlatformsCount; i++ {
				pk := getRandomPubKey()
				keeper.SetStakedPlatform(context, types.Platform{
					Address:      sdk.Address(pk.Address()),
					PublicKey:    pk,
					Jailed:       false,
					Status:       2,
					Chains:       []string{"0021"},
					StakedTokens: sdk.NewInt(10000000),
				})
			}
			if tt.isAfterUpgrade {
				codec.UpgradeHeight = -1
			}
			addMintedCoinsToModule(t, context, &keeper, types.StakedPoolName)
			sendFromModuleToAccount(t, context, &keeper, types.StakedPoolName, tt.platform.Address, sdk.NewInt(100000000000))
			if got := keeper.ValidatePlatformStaking(context, tt.platform, tt.amount); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PlatformStateChange.ValidatePlatformStaking() = got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlatformStateChange_JailPlatform(t *testing.T) {
	jailedPlatform := getStakedPlatform()
	jailedPlatform.Jailed = true
	tests := []struct {
		name     string
		platform types.Platform
		hasError bool
		want     interface{}
	}{
		{
			name:     "jails platform",
			platform: getStakedPlatform(),
			want:     true,
		},
		{
			name:     "already jailed platform ",
			platform: jailedPlatform,
			hasError: true,
			want:     "cannot jail already jailed platform, platform:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			keeper.SetPlatform(context, tt.platform)
			keeper.SetStakedPlatform(context, tt.platform)

			switch tt.hasError {
			case true:
				keeper.JailPlatform(context, tt.platform.GetAddress())
			default:
				keeper.JailPlatform(context, tt.platform.GetAddress())
				if got, _ := keeper.GetPlatform(context, tt.platform.GetAddress()); got.Jailed != tt.want {
					t.Errorf("PlatformStateChange.ValidatePlatformBeginUnstaking() = got %v, want %v", tt.platform.Jailed, tt.want)
				}
			}

		})
	}
}

func TestPlatformStateChange_UnjailPlatform(t *testing.T) {
	jailedPlatform := getStakedPlatform()
	jailedPlatform.Jailed = true
	tests := []struct {
		name     string
		platform types.Platform
		hasError bool
		want     interface{}
	}{
		{
			name:     "unjails platform",
			platform: jailedPlatform,
			want:     false,
		},
		{
			name:     "already jailed platform ",
			platform: getStakedPlatform(),
			hasError: true,
			want:     "cannot unjail already unjailed platform, platform:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			keeper.SetPlatform(context, tt.platform)
			keeper.SetStakedPlatform(context, tt.platform)

			switch tt.hasError {
			case true:
				keeper.UnjailPlatform(context, tt.platform.GetAddress())
			default:
				keeper.UnjailPlatform(context, tt.platform.GetAddress())
				if got, _ := keeper.GetPlatform(context, tt.platform.GetAddress()); got.Jailed != tt.want {
					t.Errorf("PlatformStateChange.ValidatePlatformBeginUnstaking() = got %v, want %v", tt.platform.Jailed, tt.want)
				}
			}

		})
	}
}

func TestPlatformStateChange_StakePlatform(t *testing.T) {
	tests := []struct {
		name     string
		platform types.Platform
		amount   sdk.BigInt
	}{
		{
			name:     "name registers platforms",
			platform: getUnstakedPlatform(),
			amount:   sdk.NewInt(100000000000),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			addMintedCoinsToModule(t, context, &keeper, types.StakedPoolName)
			sendFromModuleToAccount(t, context, &keeper, types.StakedPoolName, tt.platform.Address, sdk.NewInt(100000000000))
			_ = keeper.StakePlatform(context, tt.platform, tt.amount)
			got, found := keeper.GetPlatform(context, tt.platform.Address)
			if !found {
				t.Errorf("PlatformStateChanges.RegisterPlatform() = Did not register platform")
			}
			if !got.StakedTokens.Equal(tt.amount.Add(sdk.NewInt(100000000000))) {
				t.Errorf("PlatformStateChanges.RegisterPlatform() = Did not register platform %v", got.StakedTokens)
			}

		})

	}
}

func TestPlatformStateChange_EditAndValidateStakePlatform(t *testing.T) {
	stakeAmount := sdk.NewInt(100000000000)
	accountAmount := sdk.NewInt(1000000000000).Add(stakeAmount)
	bumpStakeAmount := sdk.NewInt(1000000000000)
	newChains := []string{"0021"}
	platform := getUnstakedPlatform()
	platform.StakedTokens = sdk.ZeroInt()
	// updatedStakeAmount
	updateStakeAmountPlatform := platform
	updateStakeAmountPlatform.StakedTokens = bumpStakeAmount
	// updatedStakeAmountFail
	updateStakeAmountPlatformFail := platform
	updateStakeAmountPlatformFail.StakedTokens = stakeAmount.Sub(sdk.OneInt())
	// updatedStakeAmountNotEnoughCoins
	notEnoughCoinsAccount := stakeAmount
	// updateChains
	updateChainsPlatform := platform
	updateChainsPlatform.StakedTokens = stakeAmount
	updateChainsPlatform.Chains = newChains
	//same platform no change no fail
	updateNothingPlatform := platform
	updateNothingPlatform.StakedTokens = stakeAmount
	tests := []struct {
		name          string
		accountAmount sdk.BigInt
		origPlatform  types.Platform
		amount        sdk.BigInt
		want          types.Platform
		err           sdk.Error
	}{
		{
			name:          "edit stake amount of existing platform",
			accountAmount: accountAmount,
			origPlatform:  platform,
			amount:        stakeAmount,
			want:          updateStakeAmountPlatform,
		},
		{
			name:          "FAIL edit stake amount of existing platform",
			accountAmount: accountAmount,
			origPlatform:  platform,
			amount:        stakeAmount,
			want:          updateStakeAmountPlatformFail,
			err:           types.ErrMinimumEditStake("platforms"),
		},
		{
			name:          "edit stake the chains of the platform",
			accountAmount: accountAmount,
			origPlatform:  platform,
			amount:        stakeAmount,
			want:          updateChainsPlatform,
		},
		{
			name:          "FAIL not enough coins to bump stake amount of existing platform",
			accountAmount: notEnoughCoinsAccount,
			origPlatform:  platform,
			amount:        stakeAmount,
			want:          updateStakeAmountPlatform,
			err:           types.ErrNotEnoughCoins("platforms"),
		},
		{
			name:          "update nothing for the platform",
			accountAmount: accountAmount,
			origPlatform:  platform,
			amount:        stakeAmount,
			want:          updateNothingPlatform,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// test setup
			codec.UpgradeHeight = -1
			context, _, keeper := createTestInput(t, true)
			coins := sdk.NewCoins(sdk.NewCoin(keeper.StakeDenom(context), tt.accountAmount))
			err := keeper.AccountKeeper.MintCoins(context, types.StakedPoolName, coins)
			if err != nil {
				t.Fail()
			}
			err = keeper.AccountKeeper.SendCoinsFromModuleToAccount(context, types.StakedPoolName, tt.origPlatform.Address, coins)
			if err != nil {
				t.Fail()
			}
			err = keeper.StakePlatform(context, tt.origPlatform, tt.amount)
			if err != nil {
				t.Fail()
			}
			// test begins here
			err = keeper.ValidatePlatformStaking(context, tt.want, tt.want.StakedTokens)
			if err != nil {
				if tt.err.Error() != err.Error() {
					t.Fatalf("Got error %s wanted error %s", err, tt.err)
				}
				return
			}
			// edit stake
			_ = keeper.StakePlatform(context, tt.want, tt.want.StakedTokens)
			tt.want.MaxRelays = keeper.CalculatePlatformRelays(context, tt.want)
			tt.want.Status = sdk.Staked
			// see if the changes stuck
			got, _ := keeper.GetPlatform(context, tt.origPlatform.Address)
			if !got.Equals(tt.want) {
				t.Fatalf("Got platform %s\nWanted platform %s", got.String(), tt.want.String())
			}
		})

	}
}

func TestPlatformStateChange_BeginUnstakingPlatform(t *testing.T) {
	tests := []struct {
		name     string
		platform types.Platform
		want     sdk.StakeStatus
	}{
		{
			name:     "name registers platforms",
			platform: getStakedPlatform(),
			want:     sdk.Unstaking,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			addMintedCoinsToModule(t, context, &keeper, types.StakedPoolName)
			sendFromModuleToAccount(t, context, &keeper, types.StakedPoolName, tt.platform.Address, sdk.NewInt(100000000000))
			keeper.BeginUnstakingPlatform(context, tt.platform)
			got, found := keeper.GetPlatform(context, tt.platform.Address)
			if !found {
				t.Errorf("PlatformStateChanges.RegisterPlatform() = Did not register platform")
			}
			if got.Status != tt.want {
				t.Errorf("PlatformStateChanges.RegisterPlatform() = Did not register platform %v", got.StakedTokens)
			}
		})
	}
}
