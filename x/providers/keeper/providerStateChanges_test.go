package keeper

import (
	"reflect"
	"testing"

	"github.com/vipernet-xyz/viper-network/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/providers/types"
)

func TestProviderStateChange_ValidateProviderBeginUnstaking(t *testing.T) {
	tests := []struct {
		name     string
		provider types.Provider
		hasError bool
		want     interface{}
	}{
		{
			name:     "validates provider",
			provider: getStakedProvider(),
			want:     nil,
		},
		{
			name:     "errors if provider not staked",
			provider: getUnstakedProvider(),
			want:     types.ErrProviderStatus("providers"),
			hasError: true,
		},
		{
			name:     "validates provider",
			provider: getStakedProvider(),
			hasError: true,
			want:     "should not happen: provider trying to begin unstaking has less than the minimum stake",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			switch tt.hasError {
			case true:
				tt.provider.StakedTokens = sdk.NewInt(-1)
				_ = keeper.ValidateProviderBeginUnstaking(context, tt.provider)
			default:
				if got := keeper.ValidateProviderBeginUnstaking(context, tt.provider); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ProviderStateChange.ValidateProviderBeginUnstaking() = got %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestAppStateChange_ValidateProviderStaking(t *testing.T) {
	tests := []struct {
		name            string
		provider        types.Provider
		panics          bool
		amount          sdk.BigInt
		stakedAppsCount int
		want            interface{}
	}{
		{
			name:            "validates provider",
			stakedAppsCount: 0,
			provider:        getUnstakedProvider(),
			amount:          sdk.NewInt(1000000),
			want:            nil,
		},
		{
			name:            "errors if below minimum stake",
			provider:        getUnstakedProvider(),
			stakedAppsCount: 0,
			amount:          sdk.NewInt(0),
			want:            types.ErrMinimumStake("apps"),
		},
		{
			name:            "errors bank does not have enough coins",
			provider:        getUnstakedProvider(),
			stakedAppsCount: 0,
			amount:          sdk.NewInt(1000000000000000000),
			want:            types.ErrNotEnoughCoins("apps"),
		},
		{
			name:            "errors if max applications hit",
			provider:        getUnstakedProvider(),
			stakedAppsCount: 5,
			amount:          sdk.NewInt(1000000),
			want:            types.ErrMaxProviders("apps"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			p := keeper.GetParams(context)
			p.MaxProviders = 5
			keeper.SetParams(context, p)
			for i := 0; i < tt.stakedAppsCount; i++ {
				pk := getRandomPubKey()
				keeper.SetStakedProvider(context, types.Provider{
					Address:      sdk.Address(pk.Address()),
					PublicKey:    pk,
					Jailed:       false,
					Status:       2,
					Chains:       []string{"0021"},
					GeoZones:     []string{"0001"},
					NumServicers: 10,
					StakedTokens: sdk.NewInt(10000000),
				})
			}
			addMintedCoinsToModule(t, context, &keeper, types.StakedPoolName)
			sendFromModuleToAccount(t, context, &keeper, types.StakedPoolName, tt.provider.Address, sdk.NewInt(100000000000))
			if got := keeper.ValidateProviderStaking(context, tt.provider, tt.amount); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AppStateChange.ValidateApplicationStaking() = got %v, want %v", got, tt.want)
			}
		})
	}
}
func TestProviderStateChange_JailProvider(t *testing.T) {
	jailedProvider := getStakedProvider()
	jailedProvider.Jailed = true
	tests := []struct {
		name     string
		provider types.Provider
		hasError bool
		want     interface{}
	}{
		{
			name:     "jails provider",
			provider: getStakedProvider(),
			want:     true,
		},
		{
			name:     "already jailed provider ",
			provider: jailedProvider,
			hasError: true,
			want:     "cannot jail already jailed provider, provider:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			keeper.SetProvider(context, tt.provider)
			keeper.SetStakedProvider(context, tt.provider)

			switch tt.hasError {
			case true:
				keeper.JailProvider(context, tt.provider.GetAddress())
			default:
				keeper.JailProvider(context, tt.provider.GetAddress())
				if got, _ := keeper.GetProvider(context, tt.provider.GetAddress()); got.Jailed != tt.want {
					t.Errorf("ProviderStateChange.ValidateProviderBeginUnstaking() = got %v, want %v", tt.provider.Jailed, tt.want)
				}
			}

		})
	}
}

func TestProviderStateChange_UnjailProvider(t *testing.T) {
	jailedProvider := getStakedProvider()
	jailedProvider.Jailed = true
	tests := []struct {
		name     string
		provider types.Provider
		hasError bool
		want     interface{}
	}{
		{
			name:     "unjails provider",
			provider: jailedProvider,
			want:     false,
		},
		{
			name:     "already jailed provider ",
			provider: getStakedProvider(),
			hasError: true,
			want:     "cannot unjail already unjailed provider, provider:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			keeper.SetProvider(context, tt.provider)
			keeper.SetStakedProvider(context, tt.provider)

			switch tt.hasError {
			case true:
				keeper.UnjailProvider(context, tt.provider.GetAddress())
			default:
				keeper.UnjailProvider(context, tt.provider.GetAddress())
				if got, _ := keeper.GetProvider(context, tt.provider.GetAddress()); got.Jailed != tt.want {
					t.Errorf("ProviderStateChange.ValidateProviderBeginUnstaking() = got %v, want %v", tt.provider.Jailed, tt.want)
				}
			}

		})
	}
}

func TestProviderStateChange_StakeProvider(t *testing.T) {
	tests := []struct {
		name     string
		provider types.Provider
		amount   sdk.BigInt
	}{
		{
			name:     "name registers providers",
			provider: getUnstakedProvider(),
			amount:   sdk.NewInt(100000000000),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			addMintedCoinsToModule(t, context, &keeper, types.StakedPoolName)
			sendFromModuleToAccount(t, context, &keeper, types.StakedPoolName, tt.provider.Address, sdk.NewInt(100000000000))
			_ = keeper.StakeProvider(context, tt.provider, tt.amount)
			got, found := keeper.GetProvider(context, tt.provider.Address)
			if !found {
				t.Errorf("ProviderStateChanges.RegisterProvider() = Did not register provider")
			}
			if !got.StakedTokens.Equal(tt.amount.Add(sdk.NewInt(100000000000))) {
				t.Errorf("ProviderStateChanges.RegisterProvider() = Did not register provider %v", got.StakedTokens)
			}

		})

	}
}

func TestProviderStateChange_EditAndValidateStakeProvider(t *testing.T) {
	stakeAmount := sdk.NewInt(100000000000)
	accountAmount := sdk.NewInt(1000000000000).Add(stakeAmount)
	bumpStakeAmount := sdk.NewInt(1000000000000)
	newChains := []string{"0021"}
	provider := getUnstakedProvider()
	provider.StakedTokens = sdk.ZeroInt()
	// updatedStakeAmount
	updateStakeAmountProvider := provider
	updateStakeAmountProvider.StakedTokens = bumpStakeAmount
	// updatedStakeAmountFail
	updateStakeAmountProviderFail := provider
	updateStakeAmountProviderFail.StakedTokens = stakeAmount.Sub(sdk.OneInt())
	// updatedStakeAmountNotEnoughCoins
	notEnoughCoinsAccount := stakeAmount
	// updateChains
	updateChainsProvider := provider
	updateChainsProvider.StakedTokens = stakeAmount
	updateChainsProvider.Chains = newChains
	//same provider no change no fail
	updateNothingProvider := provider
	updateNothingProvider.StakedTokens = stakeAmount
	tests := []struct {
		name          string
		accountAmount sdk.BigInt
		origApp       types.Provider
		amount        sdk.BigInt
		want          types.Provider
		err           sdk.Error
	}{
		{
			name:          "edit stake amount of existing provider",
			accountAmount: accountAmount,
			origApp:       provider,
			amount:        stakeAmount,
			want:          updateStakeAmountProvider,
		},
		{
			name:          "FAIL edit stake amount of existing provider",
			accountAmount: accountAmount,
			origApp:       provider,
			amount:        stakeAmount,
			want:          updateStakeAmountProviderFail,
			err:           types.ErrMinimumEditStake("apps"),
		},
		{
			name:          "edit stake the chains of the provider",
			accountAmount: accountAmount,
			origApp:       provider,
			amount:        stakeAmount,
			want:          updateChainsProvider,
		},
		{
			name:          "FAIL not enough coins to bump stake amount of existing provider",
			accountAmount: notEnoughCoinsAccount,
			origApp:       provider,
			amount:        stakeAmount,
			want:          updateStakeAmountProvider,
			err:           types.ErrNotEnoughCoins("apps"),
		},
		{
			name:          "update nothing for the provider",
			accountAmount: accountAmount,
			origApp:       provider,
			amount:        stakeAmount,
			want:          updateNothingProvider,
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
			err = keeper.AccountKeeper.SendCoinsFromModuleToAccount(context, types.StakedPoolName, tt.origApp.Address, coins)
			if err != nil {
				t.Fail()
			}
			err = keeper.StakeProvider(context, tt.origApp, tt.amount)
			if err != nil {
				t.Fail()
			}
			// test begins here
			err = keeper.ValidateProviderStaking(context, tt.want, tt.want.StakedTokens)
			if err != nil {
				if tt.err.Error() != err.Error() {
					t.Fatalf("Got error %s wanted error %s", err, tt.err)
				}
				return
			}
			// edit stake
			_ = keeper.StakeProvider(context, tt.want, tt.want.StakedTokens)
			tt.want.MaxRelays = keeper.CalculateProviderRelays(context, tt.want)
			tt.want.Status = sdk.Staked
			// see if the changes stuck
			got, _ := keeper.GetProvider(context, tt.origApp.Address)
			if !got.Equals(tt.want) {
				t.Fatalf("Got app %s\nWanted app %s", got.String(), tt.want.String())
			}
		})

	}
}

func TestProviderStateChange_BeginUnstakingProvider(t *testing.T) {
	tests := []struct {
		name     string
		provider types.Provider
		want     sdk.StakeStatus
	}{
		{
			name:     "name registers providers",
			provider: getStakedProvider(),
			want:     sdk.Unstaking,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			addMintedCoinsToModule(t, context, &keeper, types.StakedPoolName)
			sendFromModuleToAccount(t, context, &keeper, types.StakedPoolName, tt.provider.Address, sdk.NewInt(100000000000))
			keeper.BeginUnstakingProvider(context, tt.provider)
			got, found := keeper.GetProvider(context, tt.provider.Address)
			if !found {
				t.Errorf("ProviderStateChanges.RegisterProvider() = Did not register provider")
			}
			if got.Status != tt.want {
				t.Errorf("ProviderStateChanges.RegisterProvider() = Did not register provider %v", got.StakedTokens)
			}
		})
	}
}
