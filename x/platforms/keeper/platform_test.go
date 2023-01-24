package keeper

import (
	"reflect"
	"testing"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/platforms/types"
)

func TestPlatform_SetAndGetPlatform(t *testing.T) {
	platform := getStakedPlatform()

	tests := []struct {
		name     string
		platform types.Platform
		want     bool
	}{
		{
			name:     "get and set platform",
			platform: platform,
			want:     true,
		},
		{
			name:     "not found",
			platform: platform,
			want:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)

			if tt.want {
				keeper.SetPlatform(context, tt.platform)
			}

			if _, found := keeper.GetPlatform(context, tt.platform.Address); found != tt.want {
				t.Errorf("Platform.GetPlatform() = got %v, want %v", found, tt.want)
			}
		})
	}
}

func TestPlatform_CalculatePlatformRelays(t *testing.T) {
	platform := getStakedPlatform()

	tests := []struct {
		name     string
		platform types.Platform
		want     sdk.BigInt
	}{
		{
			name:     "calculates Platform relays",
			platform: platform,
			want:     sdk.NewInt(200000000),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)

			if got := keeper.CalculatePlatformRelays(context, tt.platform); !got.Equal(tt.want) {
				t.Errorf("Platform.CalculatePlatformRelays() = got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlatform_GetAllAplications(t *testing.T) {
	platform := getStakedPlatform()

	tests := []struct {
		name     string
		platform types.Platform
		want     types.Platforms
	}{
		{
			name:     "gets all platforms",
			platform: platform,
			want:     types.Platforms([]types.Platform{platform}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)

			keeper.SetPlatform(context, tt.platform)

			if got := keeper.GetAllPlatforms(context); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Platform.GetAllPlatforms() = got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlatform_GetAplications(t *testing.T) {
	platform := getStakedPlatform()

	tests := []struct {
		name        string
		platform    types.Platform
		maxRetrieve uint16
		want        types.Platforms
	}{
		{
			name:        "gets all platforms",
			platform:    platform,
			maxRetrieve: 2,
			want:        types.Platforms([]types.Platform{platform}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)

			keeper.SetPlatform(context, tt.platform)

			if got := keeper.GetPlatforms(context, tt.maxRetrieve); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Platform.GetAllPlatforms() = got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlatform_IterateAndExecuteOverPlatforms(t *testing.T) {
	platform := getStakedPlatform()
	secondPlatform := getStakedPlatform()

	tests := []struct {
		name           string
		platform       types.Platform
		secondPlatform types.Platform
		want           int
	}{
		{
			name:           "iterates over all platforms",
			platform:       platform,
			secondPlatform: secondPlatform,
			want:           2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)

			keeper.SetPlatform(context, tt.platform)
			keeper.SetPlatform(context, tt.secondPlatform)
			got := 0
			fn := modifyFn(&got)
			keeper.IterateAndExecuteOverPlatforms(context, fn)
			if got != tt.want {
				t.Errorf("Platform.IterateAndExecuteOverPlatforms() = got %v, want %v", got, tt.want)
			}
		})
	}
}
