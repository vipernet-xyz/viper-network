package keeper

import (
	"reflect"
	"testing"

	"github.com/vipernet-xyz/viper-network/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/requestors/types"
)

func TestRequestorStateChange_ValidateRequestorBeginUnstaking(t *testing.T) {
	tests := []struct {
		name      string
		requestor types.Requestor
		hasError  bool
		want      interface{}
	}{
		{
			name:      "validates requestor",
			requestor: getStakedRequestor(),
			want:      nil,
		},
		{
			name:      "errors if requestor not staked",
			requestor: getUnstakedRequestor(),
			want:      types.ErrRequestorStatus("requestors"),
			hasError:  true,
		},
		{
			name:      "validates requestor",
			requestor: getStakedRequestor(),
			hasError:  true,
			want:      "should not happen: requestor trying to begin unstaking has less than the minimum stake",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			switch tt.hasError {
			case true:
				tt.requestor.StakedTokens = sdk.NewInt(-1)
				_ = keeper.ValidateRequestorBeginUnstaking(context, tt.requestor)
			default:
				if got := keeper.ValidateRequestorBeginUnstaking(context, tt.requestor); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("RequestorStateChange.ValidateRequestorBeginUnstaking() = got %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestAppStateChange_ValidateRequestorStaking(t *testing.T) {
	tests := []struct {
		name            string
		requestor       types.Requestor
		panics          bool
		amount          sdk.BigInt
		stakedAppsCount int
		want            interface{}
	}{
		{
			name:            "validates requestor",
			stakedAppsCount: 0,
			requestor:       getUnstakedRequestor(),
			amount:          sdk.NewInt(1000000),
			want:            nil,
		},
		{
			name:            "errors if below minimum stake",
			requestor:       getUnstakedRequestor(),
			stakedAppsCount: 0,
			amount:          sdk.NewInt(0),
			want:            types.ErrMinimumStake("apps"),
		},
		{
			name:            "errors bank does not have enough coins",
			requestor:       getUnstakedRequestor(),
			stakedAppsCount: 0,
			amount:          sdk.NewInt(1000000000000000000),
			want:            types.ErrNotEnoughCoins("apps"),
		},
		{
			name:            "errors if max applications hit",
			requestor:       getUnstakedRequestor(),
			stakedAppsCount: 5,
			amount:          sdk.NewInt(1000000),
			want:            types.ErrMaxRequestors("apps"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			p := keeper.GetParams(context)
			p.MaxRequestors = 5
			keeper.SetParams(context, p)
			for i := 0; i < tt.stakedAppsCount; i++ {
				pk := getRandomPubKey()
				keeper.SetStakedRequestor(context, types.Requestor{
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
			sendFromModuleToAccount(t, context, &keeper, types.StakedPoolName, tt.requestor.Address, sdk.NewInt(100000000000))
			if got := keeper.ValidateRequestorStaking(context, tt.requestor, tt.amount); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AppStateChange.ValidateApplicationStaking() = got %v, want %v", got, tt.want)
			}
		})
	}
}
func TestRequestorStateChange_JailRequestor(t *testing.T) {
	jailedRequestor := getStakedRequestor()
	jailedRequestor.Jailed = true
	tests := []struct {
		name      string
		requestor types.Requestor
		hasError  bool
		want      interface{}
	}{
		{
			name:      "jails requestor",
			requestor: getStakedRequestor(),
			want:      true,
		},
		{
			name:      "already jailed requestor ",
			requestor: jailedRequestor,
			hasError:  true,
			want:      "cannot jail already jailed requestor, requestor:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			keeper.SetRequestor(context, tt.requestor)
			keeper.SetStakedRequestor(context, tt.requestor)

			switch tt.hasError {
			case true:
				keeper.JailRequestor(context, tt.requestor.GetAddress())
			default:
				keeper.JailRequestor(context, tt.requestor.GetAddress())
				if got, _ := keeper.GetRequestor(context, tt.requestor.GetAddress()); got.Jailed != tt.want {
					t.Errorf("RequestorStateChange.ValidateRequestorBeginUnstaking() = got %v, want %v", tt.requestor.Jailed, tt.want)
				}
			}

		})
	}
}

func TestRequestorStateChange_UnjailRequestor(t *testing.T) {
	jailedRequestor := getStakedRequestor()
	jailedRequestor.Jailed = true
	tests := []struct {
		name      string
		requestor types.Requestor
		hasError  bool
		want      interface{}
	}{
		{
			name:      "unjails requestor",
			requestor: jailedRequestor,
			want:      false,
		},
		{
			name:      "already jailed requestor ",
			requestor: getStakedRequestor(),
			hasError:  true,
			want:      "cannot unjail already unjailed requestor, requestor:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			keeper.SetRequestor(context, tt.requestor)
			keeper.SetStakedRequestor(context, tt.requestor)

			switch tt.hasError {
			case true:
				keeper.UnjailRequestor(context, tt.requestor.GetAddress())
			default:
				keeper.UnjailRequestor(context, tt.requestor.GetAddress())
				if got, _ := keeper.GetRequestor(context, tt.requestor.GetAddress()); got.Jailed != tt.want {
					t.Errorf("RequestorStateChange.ValidateRequestorBeginUnstaking() = got %v, want %v", tt.requestor.Jailed, tt.want)
				}
			}

		})
	}
}

func TestRequestorStateChange_StakeRequestor(t *testing.T) {
	tests := []struct {
		name      string
		requestor types.Requestor
		amount    sdk.BigInt
	}{
		{
			name:      "name registers requestors",
			requestor: getUnstakedRequestor(),
			amount:    sdk.NewInt(100000000000),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			addMintedCoinsToModule(t, context, &keeper, types.StakedPoolName)
			sendFromModuleToAccount(t, context, &keeper, types.StakedPoolName, tt.requestor.Address, sdk.NewInt(100000000000))
			_ = keeper.StakeRequestor(context, tt.requestor, tt.amount)
			got, found := keeper.GetRequestor(context, tt.requestor.Address)
			if !found {
				t.Errorf("RequestorStateChanges.RegisterRequestor() = Did not register requestor")
			}
			if !got.StakedTokens.Equal(tt.amount.Add(sdk.NewInt(100000000000))) {
				t.Errorf("RequestorStateChanges.RegisterRequestor() = Did not register requestor %v", got.StakedTokens)
			}

		})

	}
}

func TestRequestorStateChange_EditAndValidateStakeRequestor(t *testing.T) {
	stakeAmount := sdk.NewInt(100000000000)
	accountAmount := sdk.NewInt(1000000000000).Add(stakeAmount)
	bumpStakeAmount := sdk.NewInt(1000000000000)
	newChains := []string{"0021"}
	newGeoZones := []string{"0003"}
	requestor := getUnstakedRequestor()
	requestor.StakedTokens = sdk.ZeroInt()
	// updatedStakeAmount
	updateStakeAmountRequestor := requestor
	updateStakeAmountRequestor.StakedTokens = bumpStakeAmount
	// updatedStakeAmountFail
	updateStakeAmountRequestorFail := requestor
	updateStakeAmountRequestorFail.StakedTokens = stakeAmount.Sub(sdk.OneInt())
	// updatedStakeAmountNotEnoughCoins
	notEnoughCoinsAccount := stakeAmount
	// updateChains
	updateChainsRequestor := requestor
	updateChainsRequestor.StakedTokens = stakeAmount
	updateChainsRequestor.Chains = newChains
	// updateGeoZones
	updateGeoZonesRequestor := requestor
	updateGeoZonesRequestor.StakedTokens = stakeAmount
	updateGeoZonesRequestor.GeoZones = newGeoZones
	//updateNumServicers
	updateNumServicersRequestor := requestor
	updateNumServicersRequestor.StakedTokens = stakeAmount
	updateNumServicersRequestor.NumServicers = 7
	//same requestor no change no fail
	updateNothingRequestor := requestor
	updateNothingRequestor.StakedTokens = stakeAmount
	tests := []struct {
		name          string
		accountAmount sdk.BigInt
		origApp       types.Requestor
		amount        sdk.BigInt
		want          types.Requestor
		err           sdk.Error
	}{
		{
			name:          "edit stake amount of existing requestor",
			accountAmount: accountAmount,
			origApp:       requestor,
			amount:        stakeAmount,
			want:          updateStakeAmountRequestor,
		},
		{
			name:          "FAIL edit stake amount of existing requestor",
			accountAmount: accountAmount,
			origApp:       requestor,
			amount:        stakeAmount,
			want:          updateStakeAmountRequestorFail,
			err:           types.ErrMinimumEditStake("apps"),
		},
		{
			name:          "edit stake the chains of the requestor",
			accountAmount: accountAmount,
			origApp:       requestor,
			amount:        stakeAmount,
			want:          updateChainsRequestor,
		},
		{
			name:          "FAIL not enough coins to bump stake amount of existing requestor",
			accountAmount: notEnoughCoinsAccount,
			origApp:       requestor,
			amount:        stakeAmount,
			want:          updateStakeAmountRequestor,
			err:           types.ErrNotEnoughCoins("apps"),
		},
		{
			name:          "update nothing for the requestor",
			accountAmount: accountAmount,
			origApp:       requestor,
			amount:        stakeAmount,
			want:          updateNothingRequestor,
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
			err = keeper.StakeRequestor(context, tt.origApp, tt.amount)
			if err != nil {
				t.Fail()
			}
			// test begins here
			err = keeper.ValidateRequestorStaking(context, tt.want, tt.want.StakedTokens)
			if err != nil {
				if tt.err.Error() != err.Error() {
					t.Fatalf("Got error %s wanted error %s", err, tt.err)
				}
				return
			}
			// edit stake
			_ = keeper.StakeRequestor(context, tt.want, tt.want.StakedTokens)
			tt.want.MaxRelays = keeper.CalculateRequestorRelays(context, tt.want)
			tt.want.Status = sdk.Staked
			// see if the changes stuck
			got, _ := keeper.GetRequestor(context, tt.origApp.Address)
			if !got.Equals(tt.want) {
				t.Fatalf("Got app %s\nWanted app %s", got.String(), tt.want.String())
			}
		})

	}
}

func TestRequestorStateChange_BeginUnstakingRequestor(t *testing.T) {
	tests := []struct {
		name      string
		requestor types.Requestor
		want      sdk.StakeStatus
	}{
		{
			name:      "name registers requestors",
			requestor: getStakedRequestor(),
			want:      sdk.Unstaking,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			addMintedCoinsToModule(t, context, &keeper, types.StakedPoolName)
			sendFromModuleToAccount(t, context, &keeper, types.StakedPoolName, tt.requestor.Address, sdk.NewInt(100000000000))
			keeper.BeginUnstakingRequestor(context, tt.requestor)
			got, found := keeper.GetRequestor(context, tt.requestor.Address)
			if !found {
				t.Errorf("RequestorStateChanges.RegisterRequestor() = Did not register requestor")
			}
			if got.Status != tt.want {
				t.Errorf("RequestorStateChanges.RegisterRequestor() = Did not register requestor %v", got.StakedTokens)
			}
		})
	}
}
