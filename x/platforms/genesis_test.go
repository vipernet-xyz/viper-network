package pos

import (
	"fmt"
	"reflect"
	"testing"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/platforms/types"
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
		{name: "get genesis from platform"},
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
	platform := getPlatform()

	jailedPlatform := getPlatform()
	jailedPlatform.Jailed = true

	zeroStakePlatform := getPlatform()
	zeroStakePlatform.StakedTokens = sdk.NewInt(0)

	singleTokenPlatform := getPlatform()
	singleTokenPlatform.StakedTokens = sdk.NewInt(1)
	tests := []struct {
		name      string
		state     types.GenesisState
		platforms types.Platforms
		params    bool
		want      interface{}
	}{
		{
			name:      "valdiates genesis for platform",
			platforms: types.Platforms{platform},
			want:      nil,
		},
		{
			name:      "errs if invalid params",
			platforms: types.Platforms{platform},
			params:    true,
			want:      fmt.Errorf("staking parameter StakeMimimum must be a positive integer"),
		},
		{
			name:      "errs if dupplicate platform in geneiss state",
			platforms: types.Platforms{platform, platform},
			want:      fmt.Errorf("duplicate platform in genesis state: address %v", platform.GetAddress()),
		},
		{
			name:      "errs if jailed platform staked",
			platforms: types.Platforms{jailedPlatform},
			want:      fmt.Errorf("platform is staked and jailed in genesis state: address %v", jailedPlatform.GetAddress()),
		},
		{
			name:      "errs if staked with zero tokens",
			platforms: types.Platforms{zeroStakePlatform},
			want:      fmt.Errorf("staked/unstaked genesis platform cannot have zero stake, platform: %v", zeroStakePlatform),
		},
		{
			name:      "errs if lower or equal than minimum stake ",
			platforms: types.Platforms{singleTokenPlatform},
			want:      fmt.Errorf("platform has less than minimum stake: %v", singleTokenPlatform),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := types.DefaultGenesisState()
			state.Platforms = tt.platforms
			if tt.params {
				state.Params.MinPlatformStake = 0
			}
			if got := ValidateGenesis(state); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidateGenesis()= got %v, want %v", got, tt.want)
			}
		})
	}
}
