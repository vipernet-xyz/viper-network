package keeper

import (
	"testing"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/providers/types"
)

func TestProvider_SetAndGetProvider(t *testing.T) {
	provider := getStakedProvider()

	tests := []struct {
		name     string
		provider types.Provider
		want     bool
	}{
		{
			name:     "get and set provider",
			provider: provider,
			want:     true,
		},
		{
			name:     "not found",
			provider: provider,
			want:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)

			if tt.want {
				keeper.SetProvider(context, tt.provider)
			}

			if _, found := keeper.GetProvider(context, tt.provider.Address); found != tt.want {
				t.Errorf("Provider.GetProvider() = got %v, want %v", found, tt.want)
			}
		})
	}
}

func TestProvider_CalculateProviderRelays(t *testing.T) {
	provider := getStakedProvider()

	tests := []struct {
		name     string
		provider types.Provider
		want     sdk.BigInt
	}{
		{
			name:     "calculates Provider relays",
			provider: provider,
			want:     sdk.NewInt(200000000),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)

			if got := keeper.CalculateProviderRelays(context, tt.provider); !got.Equal(tt.want) {
				t.Errorf("Provider.CalculateProviderRelays() = got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProvider_IterateAndExecuteOverProviders(t *testing.T) {
	provider := getStakedProvider()
	secondProvider := getStakedProvider()

	tests := []struct {
		name           string
		provider       types.Provider
		secondProvider types.Provider
		want           int
	}{
		{
			name:           "iterates over all providers",
			provider:       provider,
			secondProvider: secondProvider,
			want:           2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)

			keeper.SetProvider(context, tt.provider)
			keeper.SetProvider(context, tt.secondProvider)
			got := 0
			fn := modifyFn(&got)
			keeper.IterateAndExecuteOverProviders(context, fn)
			if got != tt.want {
				t.Errorf("Provider.IterateAndExecuteOverProviders() = got %v, want %v", got, tt.want)
			}
		})
	}
}
