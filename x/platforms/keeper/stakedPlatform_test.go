package keeper

import (
	"testing"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/platforms/types"

	"github.com/stretchr/testify/assert"
)

func TestGetAndSetStakedPlatform(t *testing.T) {
	stakedPlatform := getStakedPlatform()
	unstakedPlatform := getUnstakedPlatform()
	jailedPlatform := getStakedPlatform()
	jailedPlatform.Jailed = true

	type want struct {
		platforms []types.Platform
		length    int
	}
	tests := []struct {
		name      string
		platform  types.Platform
		platforms []types.Platform
		want      want
	}{
		{
			name:      "gets platforms",
			platforms: []types.Platform{stakedPlatform},
			want:      want{platforms: []types.Platform{stakedPlatform}, length: 1},
		},
		{
			name:      "gets emtpy slice of platforms",
			platforms: []types.Platform{unstakedPlatform},
			want:      want{platforms: []types.Platform{}, length: 0},
		},
		{
			name:      "gets emtpy slice of platforms",
			platforms: []types.Platform{jailedPlatform},
			want:      want{platforms: []types.Platform{}, length: 0},
		},
		{
			name:      "only gets staked platforms",
			platforms: []types.Platform{stakedPlatform, unstakedPlatform},
			want:      want{platforms: []types.Platform{stakedPlatform}, length: 1},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			for _, platform := range test.platforms {
				keeper.SetPlatform(context, platform)
				if platform.IsStaked() {
					keeper.SetStakedPlatform(context, platform)
				}
			}
			platforms := keeper.getStakedPlatforms(context)
			if equal := assert.ObjectsAreEqualValues(platforms, test.want.platforms); !equal { // note ObjectsAreEqualValues does not assert, manual verification is required
				t.FailNow()
			}
			assert.Equalf(t, len(platforms), test.want.length, "length of the platforms does not match want on %v", test.name)
		})
	}
}

func TestRemoveStakedPlatformTokens(t *testing.T) {
	stakedPlatform := getStakedPlatform()

	type want struct {
		tokens    sdk.BigInt
		platforms []types.Platform
		hasError  bool
	}
	tests := []struct {
		name     string
		platform types.Platform
		panics   bool
		amount   sdk.BigInt
		want
	}{
		{
			name:     "removes tokens from platform platforms",
			platform: stakedPlatform,
			amount:   sdk.NewInt(5),
			panics:   false,
			want:     want{tokens: sdk.NewInt(99999999995), platforms: []types.Platform{}},
		},
		{
			name:     "removes tokens from platform platforms",
			platform: stakedPlatform,
			amount:   sdk.NewInt(-5),
			panics:   true,
			want:     want{tokens: sdk.NewInt(99999999995), platforms: []types.Platform{}, hasError: true},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			keeper.SetPlatform(context, test.platform)
			keeper.SetStakedPlatform(context, test.platform)
			platform, err := keeper.removePlatformTokens(context, test.platform, test.amount)
			if err != nil {
				assert.True(t, test.want.hasError)
				return
			}
			assert.True(t, platform.StakedTokens.Equal(test.want.tokens), "platform staked tokens is not as want")
			store := context.KVStore(keeper.storeKey)
			sg, _ := store.Get(types.KeyForPlatformInStakingSet(platform))
			assert.NotNil(t, sg)

		})
	}
}

func TestRemoveDeleteFromStakingSet(t *testing.T) {
	stakedPlatform := getStakedPlatform()
	unstakedPlatform := getUnstakedPlatform()

	tests := []struct {
		name      string
		platforms []types.Platform
		panics    bool
		amount    sdk.BigInt
	}{
		{
			name:      "removes platforms from set",
			platforms: []types.Platform{stakedPlatform, unstakedPlatform},
			panics:    false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			for _, platform := range test.platforms {
				keeper.SetPlatform(context, platform)
				keeper.SetStakedPlatform(context, platform)
			}
			for _, platform := range test.platforms {
				keeper.deletePlatformFromStakingSet(context, platform)
			}

			platforms := keeper.getStakedPlatforms(context)
			assert.Empty(t, platforms, "there should not be any platforms in the set")
		})
	}
}

func TestGetValsIterator(t *testing.T) {
	stakedPlatform := getStakedPlatform()
	unstakedPlatform := getUnstakedPlatform()

	tests := []struct {
		name      string
		platforms []types.Platform
		panics    bool
		amount    sdk.BigInt
	}{
		{
			name:      "recieves a valid iterator",
			platforms: []types.Platform{stakedPlatform, unstakedPlatform},
			panics:    false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			for _, platform := range test.platforms {
				keeper.SetPlatform(context, platform)
				keeper.SetStakedPlatform(context, platform)
			}

			it, _ := keeper.stakedPlatformsIterator(context)
			assert.Implements(t, (*sdk.Iterator)(nil), it, "does not implement interface")
		})
	}
}

func TestPlatformStaked_IterateAndExecuteOverStakedPlatforms(t *testing.T) {
	stakedPlatform := getStakedPlatform()
	secondStakedPlatform := getStakedPlatform()

	tests := []struct {
		name      string
		platform  types.Platform
		platforms []types.Platform
		want      int
	}{
		{
			name:      "iterates over platforms",
			platforms: []types.Platform{stakedPlatform, secondStakedPlatform},
			want:      2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			for _, platform := range tt.platforms {
				keeper.SetPlatform(context, platform)
				keeper.SetStakedPlatform(context, platform)
			}
			got := 0
			fn := modifyFn(&got)

			keeper.IterateAndExecuteOverStakedPlatforms(context, fn)

			if got != tt.want {
				t.Errorf("platformStaked.IterateAndExecuteOverPlatforms() = got %v, want %v", got, tt.want)
			}
		})
	}
}
