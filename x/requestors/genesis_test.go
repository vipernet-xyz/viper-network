package pos

import (
	"fmt"
	"reflect"
	"testing"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/requestors/types"
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
		{name: "get genesis from requestor"},
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
	requestor := getRequestor()

	jailedRequestor := getRequestor()
	jailedRequestor.Jailed = true

	zeroStakeRequestor := getRequestor()
	zeroStakeRequestor.StakedTokens = sdk.NewInt(0)

	singleTokenRequestor := getRequestor()
	singleTokenRequestor.StakedTokens = sdk.NewInt(1)
	tests := []struct {
		name       string
		state      types.GenesisState
		requestors types.Requestors
		params     bool
		want       interface{}
	}{
		{
			name:       "valdiates genesis for requestor",
			requestors: types.Requestors{requestor},
			want:       nil,
		},
		{
			name:       "errs if invalid params",
			requestors: types.Requestors{requestor},
			params:     true,
			want:       fmt.Errorf("staking parameter StakeMimimum must be a positive integer"),
		},
		{
			name:       "errs if dupplicate requestor in geneiss state",
			requestors: types.Requestors{requestor, requestor},
			want:       fmt.Errorf("duplicate requestor in genesis state: address %v", requestor.GetAddress()),
		},
		{
			name:       "errs if jailed requestor staked",
			requestors: types.Requestors{jailedRequestor},
			want:       fmt.Errorf("requestor is staked and jailed in genesis state: address %v", jailedRequestor.GetAddress()),
		},
		{
			name:       "errs if staked with zero tokens",
			requestors: types.Requestors{zeroStakeRequestor},
			want:       fmt.Errorf("staked/unstaked genesis requestor cannot have zero stake, requestor: %v", zeroStakeRequestor),
		},
		{
			name:       "errs if lower or equal than minimum stake ",
			requestors: types.Requestors{singleTokenRequestor},
			want:       fmt.Errorf("requestor has less than minimum stake: %v", singleTokenRequestor),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := types.DefaultGenesisState()
			state.Requestors = tt.requestors
			if tt.params {
				state.Params.MinRequestorStake = 0
			}
			if got := ValidateGenesis(state); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidateGenesis()= got %v, want %v", got, tt.want)
			}
		})
	}
}
