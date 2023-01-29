package pos

import (
	"fmt"
	"reflect"
	"testing"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/providers/types"
)

func TestPos_InitGenesis(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "set init genesis"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, keeper, supplyKeeper, posKeeper := createTestInput(t, true)
			state := types.DefaultGenesisState()
			InitGenesis(context, keeper, supplyKeeper, posKeeper, state)
			if got := keeper.GetParams(context); got != state.Params {
				t.Errorf("InitGenesis()= got %v, want %v", got, state.Params)
			}
		})
	}
}
func TestPos_ExportGenesis(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "get genesis from provider"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, keeper, supplyKeeper, posKeeper := createTestInput(t, true)
			state := types.DefaultGenesisState()
			InitGenesis(context, keeper, supplyKeeper, posKeeper, state)
			state.Exported = true // Export genesis returns an exported state
			if got := ExportGenesis(context, keeper); !reflect.DeepEqual(got, state) {
				t.Errorf("\nExportGenesis()=\nGot-> %v\nWant-> %v", got, state.Params)
			}
		})
	}
}
func TestPos_ValidateGeneis(t *testing.T) {
	provider := getProvider()

	jailedProvider := getProvider()
	jailedProvider.Jailed = true

	zeroStakeProvider := getProvider()
	zeroStakeProvider.StakedTokens = sdk.NewInt(0)

	singleTokenProvider := getProvider()
	singleTokenProvider.StakedTokens = sdk.NewInt(1)
	tests := []struct {
		name      string
		state     types.GenesisState
		providers types.Providers
		params    bool
		want      interface{}
	}{
		{
			name:      "valdiates genesis for provider",
			providers: types.Providers{provider},
			want:      nil,
		},
		{
			name:      "errs if invalid params",
			providers: types.Providers{provider},
			params:    true,
			want:      fmt.Errorf("staking parameter StakeMimimum must be a positive integer"),
		},
		{
			name:      "errs if dupplicate provider in geneiss state",
			providers: types.Providers{provider, provider},
			want:      fmt.Errorf("duplicate provider in genesis state: address %v", provider.GetAddress()),
		},
		{
			name:      "errs if jailed provider staked",
			providers: types.Providers{jailedProvider},
			want:      fmt.Errorf("provider is staked and jailed in genesis state: address %v", jailedProvider.GetAddress()),
		},
		{
			name:      "errs if staked with zero tokens",
			providers: types.Providers{zeroStakeProvider},
			want:      fmt.Errorf("staked/unstaked genesis provider cannot have zero stake, provider: %v", zeroStakeProvider),
		},
		{
			name:      "errs if lower or equal than minimum stake ",
			providers: types.Providers{singleTokenProvider},
			want:      fmt.Errorf("provider has less than minimum stake: %v", singleTokenProvider),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := types.DefaultGenesisState()
			state.Providers = tt.providers
			if tt.params {
				state.Params.MinProviderStake = 0
			}
			if got := ValidateGenesis(state); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidateGenesis()= got %v, want %v", got, tt.want)
			}
		})
	}
}
