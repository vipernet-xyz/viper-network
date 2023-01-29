package keeper

import (
	"testing"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/providers/types"

	"github.com/stretchr/testify/assert"
)

func TestProviderUnstaked_GetAndSetlUnstaking(t *testing.T) {
	stakedProvider := getStakedProvider()
	unstaking := getUnstakingProvider()

	type want struct {
		providers       []types.Provider
		stakedProviders bool
		length          int
	}
	type args struct {
		providers      []types.Provider
		stakedProvider types.Provider
	}
	tests := []struct {
		name      string
		provider  types.Provider
		providers []types.Provider
		want
		args
	}{
		{
			name: "gets providers",
			args: args{providers: []types.Provider{unstaking}},
			want: want{providers: []types.Provider{unstaking}, length: 1, stakedProviders: false},
		},
		{
			name: "gets emtpy slice of providers",
			want: want{length: 0, stakedProviders: true},
			args: args{stakedProvider: stakedProvider},
		},
		{
			name:      "only gets unstaking providers",
			providers: []types.Provider{stakedProvider, unstaking},
			want:      want{length: 1, stakedProviders: true},
			args:      args{stakedProvider: stakedProvider, providers: []types.Provider{unstaking}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			for _, provider := range tt.args.providers {
				keeper.SetProvider(context, provider)
			}
			if tt.want.stakedProviders {
				keeper.SetProvider(context, tt.args.stakedProvider)
			}
			providers := keeper.getAllUnstakingProviders(context)
			if len(providers) != tt.want.length {
				t.Errorf("providerUnstaked.GetProviders() = %v, want %v", len(providers), tt.want.length)
			}
		})
	}
}

func TestProviderUnstaked_DeleteUnstakingProvider(t *testing.T) {
	stakedProvider := getStakedProvider()
	secondStakedProvider := getStakedProvider()

	type want struct {
		stakedProviders bool
		length          int
	}
	type args struct {
		providers []types.Provider
	}
	tests := []struct {
		name      string
		provider  types.Provider
		providers []types.Provider
		sets      bool
		want
		args
	}{
		{
			name: "deletes",
			args: args{providers: []types.Provider{stakedProvider, secondStakedProvider}},
			sets: false,
			want: want{length: 1, stakedProviders: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			for _, provider := range tt.args.providers {
				keeper.SetProvider(context, provider)
			}
			keeper.SetUnstakingProvider(context, tt.args.providers[0])
			_ = keeper.getAllUnstakingProviders(context)

			keeper.deleteUnstakingProvider(context, tt.args.providers[1])

			if got := keeper.getAllUnstakingProviders(context); len(got) != tt.want.length {
				t.Errorf("KeeperCoins.BurnStakedTokens()= %v, want %v", len(got), tt.want.length)
			}
		})
	}
}

func TestProviderUnstaked_DeleteUnstakingProviders(t *testing.T) {
	stakedProvider := getStakedProvider()
	secondaryStakedProvider := getStakedProvider()

	type want struct {
		stakedProviders bool
		length          int
	}
	type args struct {
		providers []types.Provider
	}
	tests := []struct {
		name      string
		provider  types.Provider
		providers []types.Provider
		want
		args
	}{
		{
			name: "deletes all unstaking provider",
			args: args{providers: []types.Provider{stakedProvider, secondaryStakedProvider}},
			want: want{length: 0, stakedProviders: false},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			for _, provider := range test.args.providers {
				keeper.SetProvider(context, provider)
				keeper.SetUnstakingProvider(context, provider)
				keeper.deleteUnstakingProviders(context, provider.UnstakingCompletionTime)
			}

			providers := keeper.getAllUnstakingProviders(context)

			assert.Equalf(t, test.want.length, len(providers), "length of the providers does not match want on %v", test.name)
		})
	}
}

func TestProviderUnstaked_GetAllMatureProviders(t *testing.T) {
	stakingProvider := getUnstakingProvider()

	type want struct {
		providers       []types.Provider
		stakedProviders bool
		length          int
	}
	type args struct {
		providers []types.Provider
	}
	tests := []struct {
		name      string
		provider  types.Provider
		providers []types.Provider
		want
		args
	}{
		{
			name: "gets all mature providers",
			args: args{providers: []types.Provider{stakingProvider}},
			want: want{providers: []types.Provider{stakingProvider}, length: 1, stakedProviders: false},
		},
		{
			name: "gets empty slice if no mature providers",
			args: args{providers: []types.Provider{}},
			want: want{providers: []types.Provider{stakingProvider}, length: 0, stakedProviders: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			for _, provider := range tt.args.providers {
				keeper.SetProvider(context, provider)
			}
			if got := keeper.getMatureProviders(context); len(got) != tt.want.length {
				t.Errorf("providerUnstaked.unstakeAllMatureProviders()= %v, want %v", len(got), tt.want.length)
			}
		})
	}
}

//func TestProviderUnstaked_UnstakeAllMatureProviders(t *testing.T) {
//	stakingProvider := getUnstakingProvider()
//
//	type want struct {
//		providers       []types.Provider
//		stakedProviders bool
//		length             int
//	}
//	type args struct {
//		stakedVal         types.Provider
//		providers      []types.Provider
//		stakedProvider types.Provider
//	}
//	tests := []struct {
//		name         string
//		provider  types.Provider
//		providers []types.Provider
//		want
//		args
//	}{
//		{
//			name: "unstake mature providers",
//			args: args{providers: []types.Provider{stakingProvider}},
//			want: want{providers: []types.Provider{stakingProvider}, length: 0, stakedProviders: false},
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			context, _, keeper := createTestInput(t, true)
//			for _, provider := range tt.args.providers {
//				keeper.SetProvider(context, provider)
//				keeper.SetUnstakingProvider(context, provider)
//			}
//			keeper.unstakeAllMatureProviders(context)
//			if got := keeper.getAllUnstakingProviders(context); len(got) != tt.want.length {
//				t.Errorf("providerUnstaked.unstakeAllMatureProviders()= %v, want %v", len(got), tt.want.length)
//			}
//		})
//	}
//}

func TestProviderUnstaked_UnstakingProvidersIterator(t *testing.T) {
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			for _, provider := range tt.providers {
				keeper.SetProvider(context, provider)
				keeper.SetStakedProvider(context, provider)
			}

			it, _ := keeper.unstakingProvidersIterator(context, context.BlockHeader().Time)
			if v, ok := it.(sdk.Iterator); !ok {
				t.Errorf("providerUnstaked.UnstakingProvidersIterator()= %v does not implement sdk.Iterator", v)
			}
		})
	}
}
