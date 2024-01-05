package keeper

import (
	"testing"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/requestors/types"

	"github.com/stretchr/testify/assert"
)

func TestGetAndSetStakedRequestor(t *testing.T) {
	stakedRequestor := getStakedRequestor()
	unstakedRequestor := getUnstakedRequestor()
	jailedRequestor := getStakedRequestor()
	jailedRequestor.Jailed = true

	type want struct {
		requestors []types.Requestor
		length     int
	}
	tests := []struct {
		name       string
		requestor  types.Requestor
		requestors []types.Requestor
		want       want
	}{
		{
			name:       "gets requestors",
			requestors: []types.Requestor{stakedRequestor},
			want:       want{requestors: []types.Requestor{stakedRequestor}, length: 1},
		},
		{
			name:       "gets emtpy slice of requestors",
			requestors: []types.Requestor{unstakedRequestor},
			want:       want{requestors: []types.Requestor{}, length: 0},
		},
		{
			name:       "gets emtpy slice of requestors",
			requestors: []types.Requestor{jailedRequestor},
			want:       want{requestors: []types.Requestor{}, length: 0},
		},
		{
			name:       "only gets staked requestors",
			requestors: []types.Requestor{stakedRequestor, unstakedRequestor},
			want:       want{requestors: []types.Requestor{stakedRequestor}, length: 1},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			for _, requestor := range test.requestors {
				keeper.SetRequestor(context, requestor)
				if requestor.IsStaked() {
					keeper.SetStakedRequestor(context, requestor)
				}
			}
			requestors := keeper.getStakedRequestors(context)
			if equal := assert.ObjectsAreEqualValues(requestors, test.want.requestors); !equal { // note ObjectsAreEqualValues does not assert, manual verification is required
				t.FailNow()
			}
			assert.Equalf(t, len(requestors), test.want.length, "length of the requestors does not match want on %v", test.name)
		})
	}
}

func TestRemoveStakedRequestorTokens(t *testing.T) {
	stakedRequestor := getStakedRequestor()

	type want struct {
		tokens     sdk.BigInt
		requestors []types.Requestor
		hasError   bool
	}
	tests := []struct {
		name      string
		requestor types.Requestor
		panics    bool
		amount    sdk.BigInt
		want
	}{
		{
			name:      "removes tokens from requestor requestors",
			requestor: stakedRequestor,
			amount:    sdk.NewInt(5),
			panics:    false,
			want:      want{tokens: sdk.NewInt(99999999995), requestors: []types.Requestor{}},
		},
		{
			name:      "removes tokens from requestor requestors",
			requestor: stakedRequestor,
			amount:    sdk.NewInt(-5),
			panics:    true,
			want:      want{tokens: sdk.NewInt(99999999995), requestors: []types.Requestor{}, hasError: true},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			keeper.SetRequestor(context, test.requestor)
			keeper.SetStakedRequestor(context, test.requestor)
			requestor, err := keeper.removeRequestorTokens(context, test.requestor, test.amount)
			if err != nil {
				assert.True(t, test.want.hasError)
				return
			}
			assert.True(t, requestor.StakedTokens.Equal(test.want.tokens), "requestor staked tokens is not as want")
			store := context.KVStore(keeper.storeKey)
			sg, _ := store.Get(types.KeyForRequestorInStakingSet(requestor))
			assert.NotNil(t, sg)

		})
	}
}

func TestRemoveDeleteFromStakingSet(t *testing.T) {
	stakedRequestor := getStakedRequestor()
	unstakedRequestor := getUnstakedRequestor()

	tests := []struct {
		name       string
		requestors []types.Requestor
		panics     bool
		amount     sdk.BigInt
	}{
		{
			name:       "removes requestors from set",
			requestors: []types.Requestor{stakedRequestor, unstakedRequestor},
			panics:     false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			for _, requestor := range test.requestors {
				keeper.SetRequestor(context, requestor)
				keeper.SetStakedRequestor(context, requestor)
			}
			for _, requestor := range test.requestors {
				keeper.deleteRequestorFromStakingSet(context, requestor)
			}

			requestors := keeper.getStakedRequestors(context)
			assert.Empty(t, requestors, "there should not be any requestors in the set")
		})
	}
}

func TestGetValsIterator(t *testing.T) {
	stakedRequestor := getStakedRequestor()
	unstakedRequestor := getUnstakedRequestor()

	tests := []struct {
		name       string
		requestors []types.Requestor
		panics     bool
		amount     sdk.BigInt
	}{
		{
			name:       "recieves a valid iterator",
			requestors: []types.Requestor{stakedRequestor, unstakedRequestor},
			panics:     false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			for _, requestor := range test.requestors {
				keeper.SetRequestor(context, requestor)
				keeper.SetStakedRequestor(context, requestor)
			}

			it, _ := keeper.stakedRequestorsIterator(context)
			assert.Implements(t, (*sdk.Iterator)(nil), it, "does not implement interface")
		})
	}
}

func TestRequestorStaked_IterateAndExecuteOverStakedRequestors(t *testing.T) {
	stakedRequestor := getStakedRequestor()
	secondStakedRequestor := getStakedRequestor()

	tests := []struct {
		name       string
		requestor  types.Requestor
		requestors []types.Requestor
		want       int
	}{
		{
			name:       "iterates over requestors",
			requestors: []types.Requestor{stakedRequestor, secondStakedRequestor},
			want:       2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			for _, requestor := range tt.requestors {
				keeper.SetRequestor(context, requestor)
				keeper.SetStakedRequestor(context, requestor)
			}
			got := 0
			fn := modifyFn(&got)

			keeper.IterateAndExecuteOverStakedRequestors(context, fn)

			if got != tt.want {
				t.Errorf("requestorStaked.IterateAndExecuteOverRequestors() = got %v, want %v", got, tt.want)
			}
		})
	}
}
