package keeper

import (
	"testing"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/providers/types"

	"github.com/stretchr/testify/assert"
)

func TestGetAndSetStakedProvider(t *testing.T) {
	stakedProvider := getStakedProvider()
	unstakedProvider := getUnstakedProvider()
	jailedProvider := getStakedProvider()
	jailedProvider.Jailed = true

	type want struct {
		providers []types.Provider
		length    int
	}
	tests := []struct {
		name      string
		provider  types.Provider
		providers []types.Provider
		want      want
	}{
		{
			name:      "gets providers",
			providers: []types.Provider{stakedProvider},
			want:      want{providers: []types.Provider{stakedProvider}, length: 1},
		},
		{
			name:      "gets emtpy slice of providers",
			providers: []types.Provider{unstakedProvider},
			want:      want{providers: []types.Provider{}, length: 0},
		},
		{
			name:      "gets emtpy slice of providers",
			providers: []types.Provider{jailedProvider},
			want:      want{providers: []types.Provider{}, length: 0},
		},
		{
			name:      "only gets staked providers",
			providers: []types.Provider{stakedProvider, unstakedProvider},
			want:      want{providers: []types.Provider{stakedProvider}, length: 1},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			for _, provider := range test.providers {
				keeper.SetProvider(context, provider)
				if provider.IsStaked() {
					keeper.SetStakedProvider(context, provider)
				}
			}
			providers := keeper.getStakedProviders(context)
			if equal := assert.ObjectsAreEqualValues(providers, test.want.providers); !equal { // note ObjectsAreEqualValues does not assert, manual verification is required
				t.FailNow()
			}
			assert.Equalf(t, len(providers), test.want.length, "length of the providers does not match want on %v", test.name)
		})
	}
}

func TestRemoveStakedProviderTokens(t *testing.T) {
	stakedProvider := getStakedProvider()

	type want struct {
		tokens    sdk.BigInt
		providers []types.Provider
		hasError  bool
	}
	tests := []struct {
		name     string
		provider types.Provider
		panics   bool
		amount   sdk.BigInt
		want
	}{
		{
			name:     "removes tokens from provider providers",
			provider: stakedProvider,
			amount:   sdk.NewInt(5),
			panics:   false,
			want:     want{tokens: sdk.NewInt(99999999995), providers: []types.Provider{}},
		},
		{
			name:     "removes tokens from provider providers",
			provider: stakedProvider,
			amount:   sdk.NewInt(-5),
			panics:   true,
			want:     want{tokens: sdk.NewInt(99999999995), providers: []types.Provider{}, hasError: true},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			keeper.SetProvider(context, test.provider)
			keeper.SetStakedProvider(context, test.provider)
			provider, err := keeper.removeProviderTokens(context, test.provider, test.amount)
			if err != nil {
				assert.True(t, test.want.hasError)
				return
			}
			assert.True(t, provider.StakedTokens.Equal(test.want.tokens), "provider staked tokens is not as want")
			store := context.KVStore(keeper.storeKey)
			sg, _ := store.Get(types.KeyForProviderInStakingSet(provider))
			assert.NotNil(t, sg)

		})
	}
}

func TestRemoveDeleteFromStakingSet(t *testing.T) {
	stakedProvider := getStakedProvider()
	unstakedProvider := getUnstakedProvider()

	tests := []struct {
		name      string
		providers []types.Provider
		panics    bool
		amount    sdk.BigInt
	}{
		{
			name:      "removes providers from set",
			providers: []types.Provider{stakedProvider, unstakedProvider},
			panics:    false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			for _, provider := range test.providers {
				keeper.SetProvider(context, provider)
				keeper.SetStakedProvider(context, provider)
			}
			for _, provider := range test.providers {
				keeper.deleteProviderFromStakingSet(context, provider)
			}

			providers := keeper.getStakedProviders(context)
			assert.Empty(t, providers, "there should not be any providers in the set")
		})
	}
}

func TestGetValsIterator(t *testing.T) {
	stakedProvider := getStakedProvider()
	unstakedProvider := getUnstakedProvider()

	tests := []struct {
		name      string
		providers []types.Provider
		panics    bool
		amount    sdk.BigInt
	}{
		{
			name:      "recieves a valid iterator",
			providers: []types.Provider{stakedProvider, unstakedProvider},
			panics:    false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			for _, provider := range test.providers {
				keeper.SetProvider(context, provider)
				keeper.SetStakedProvider(context, provider)
			}

			it, _ := keeper.stakedProvidersIterator(context)
			assert.Implements(t, (*sdk.Iterator)(nil), it, "does not implement interface")
		})
	}
}

func TestProviderStaked_IterateAndExecuteOverStakedProviders(t *testing.T) {
	stakedProvider := getStakedProvider()
	secondStakedProvider := getStakedProvider()

	tests := []struct {
		name      string
		provider  types.Provider
		providers []types.Provider
		want      int
	}{
		{
			name:      "iterates over providers",
			providers: []types.Provider{stakedProvider, secondStakedProvider},
			want:      2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			for _, provider := range tt.providers {
				keeper.SetProvider(context, provider)
				keeper.SetStakedProvider(context, provider)
			}
			got := 0
			fn := modifyFn(&got)

			keeper.IterateAndExecuteOverStakedProviders(context, fn)

			if got != tt.want {
				t.Errorf("providerStaked.IterateAndExecuteOverProviders() = got %v, want %v", got, tt.want)
			}
		})
	}
}
