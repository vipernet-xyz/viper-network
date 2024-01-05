package keeper

import (
	"testing"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/requestors/types"
)

func TestRequestor_SetAndGetRequestor(t *testing.T) {
	requestor := getStakedRequestor()

	tests := []struct {
		name      string
		requestor types.Requestor
		want      bool
	}{
		{
			name:      "get and set requestor",
			requestor: requestor,
			want:      true,
		},
		{
			name:      "not found",
			requestor: requestor,
			want:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)

			if tt.want {
				keeper.SetRequestor(context, tt.requestor)
			}

			if _, found := keeper.GetRequestor(context, tt.requestor.Address); found != tt.want {
				t.Errorf("Requestor.GetRequestor() = got %v, want %v", found, tt.want)
			}
		})
	}
}

func TestRequestor_CalculateRequestorRelays(t *testing.T) {
	requestor := getStakedRequestor()

	tests := []struct {
		name      string
		requestor types.Requestor
		want      sdk.BigInt
	}{
		{
			name:      "calculates Requestor relays",
			requestor: requestor,
			want:      sdk.NewInt(200000000),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)

			if got := keeper.CalculateRequestorRelays(context, tt.requestor); !got.Equal(tt.want) {
				t.Errorf("Requestor.CalculateRequestorRelays() = got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRequestor_IterateAndExecuteOverRequestors(t *testing.T) {
	requestor := getStakedRequestor()
	secondRequestor := getStakedRequestor()

	tests := []struct {
		name            string
		requestor       types.Requestor
		secondRequestor types.Requestor
		want            int
	}{
		{
			name:            "iterates over all requestors",
			requestor:       requestor,
			secondRequestor: secondRequestor,
			want:            2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)

			keeper.SetRequestor(context, tt.requestor)
			keeper.SetRequestor(context, tt.secondRequestor)
			got := 0
			fn := modifyFn(&got)
			keeper.IterateAndExecuteOverRequestors(context, fn)
			if got != tt.want {
				t.Errorf("Requestor.IterateAndExecuteOverRequestors() = got %v, want %v", got, tt.want)
			}
		})
	}
}
