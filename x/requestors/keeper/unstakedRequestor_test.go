package keeper

import (
	"testing"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/requestors/types"

	"github.com/stretchr/testify/assert"
)

func TestRequestorUnstaked_GetAndSetlUnstaking(t *testing.T) {
	stakedRequestor := getStakedRequestor()
	unstaking := getUnstakingRequestor()

	type want struct {
		requestors       []types.Requestor
		stakedRequestors bool
		length           int
	}
	type args struct {
		requestors      []types.Requestor
		stakedRequestor types.Requestor
	}
	tests := []struct {
		name       string
		requestor  types.Requestor
		requestors []types.Requestor
		want
		args
	}{
		{
			name: "gets requestors",
			args: args{requestors: []types.Requestor{unstaking}},
			want: want{requestors: []types.Requestor{unstaking}, length: 1, stakedRequestors: false},
		},
		{
			name: "gets emtpy slice of requestors",
			want: want{length: 0, stakedRequestors: true},
			args: args{stakedRequestor: stakedRequestor},
		},
		{
			name:       "only gets unstaking requestors",
			requestors: []types.Requestor{stakedRequestor, unstaking},
			want:       want{length: 1, stakedRequestors: true},
			args:       args{stakedRequestor: stakedRequestor, requestors: []types.Requestor{unstaking}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			for _, requestor := range tt.args.requestors {
				keeper.SetRequestor(context, requestor)
			}
			if tt.want.stakedRequestors {
				keeper.SetRequestor(context, tt.args.stakedRequestor)
			}
			requestors := keeper.getAllUnstakingRequestors(context)
			if len(requestors) != tt.want.length {
				t.Errorf("requestorUnstaked.GetRequestors() = %v, want %v", len(requestors), tt.want.length)
			}
		})
	}
}

func TestRequestorUnstaked_DeleteUnstakingRequestor(t *testing.T) {
	stakedRequestor := getStakedRequestor()
	secondStakedRequestor := getStakedRequestor()

	type want struct {
		stakedRequestors bool
		length           int
	}
	type args struct {
		requestors []types.Requestor
	}
	tests := []struct {
		name       string
		requestor  types.Requestor
		requestors []types.Requestor
		sets       bool
		want
		args
	}{
		{
			name: "deletes",
			args: args{requestors: []types.Requestor{stakedRequestor, secondStakedRequestor}},
			sets: false,
			want: want{length: 1, stakedRequestors: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			for _, requestor := range tt.args.requestors {
				keeper.SetRequestor(context, requestor)
			}
			keeper.SetUnstakingRequestor(context, tt.args.requestors[0])
			_ = keeper.getAllUnstakingRequestors(context)

			keeper.deleteUnstakingRequestor(context, tt.args.requestors[1])

			if got := keeper.getAllUnstakingRequestors(context); len(got) != tt.want.length {
				t.Errorf("KeeperCoins.BurnStakedTokens()= %v, want %v", len(got), tt.want.length)
			}
		})
	}
}

func TestRequestorUnstaked_DeleteUnstakingRequestors(t *testing.T) {
	stakedRequestor := getStakedRequestor()
	secondaryStakedRequestor := getStakedRequestor()

	type want struct {
		stakedRequestors bool
		length           int
	}
	type args struct {
		requestors []types.Requestor
	}
	tests := []struct {
		name       string
		requestor  types.Requestor
		requestors []types.Requestor
		want
		args
	}{
		{
			name: "deletes all unstaking requestor",
			args: args{requestors: []types.Requestor{stakedRequestor, secondaryStakedRequestor}},
			want: want{length: 0, stakedRequestors: false},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			for _, requestor := range test.args.requestors {
				keeper.SetRequestor(context, requestor)
				keeper.SetUnstakingRequestor(context, requestor)
				keeper.deleteUnstakingRequestors(context, requestor.UnstakingCompletionTime)
			}

			requestors := keeper.getAllUnstakingRequestors(context)

			assert.Equalf(t, test.want.length, len(requestors), "length of the requestors does not match want on %v", test.name)
		})
	}
}

func TestRequestorUnstaked_GetAllMatureRequestors(t *testing.T) {
	stakingRequestor := getUnstakingRequestor()

	type want struct {
		requestors       []types.Requestor
		stakedRequestors bool
		length           int
	}
	type args struct {
		requestors []types.Requestor
	}
	tests := []struct {
		name       string
		requestor  types.Requestor
		requestors []types.Requestor
		want
		args
	}{
		{
			name: "gets all mature requestors",
			args: args{requestors: []types.Requestor{stakingRequestor}},
			want: want{requestors: []types.Requestor{stakingRequestor}, length: 1, stakedRequestors: false},
		},
		{
			name: "gets empty slice if no mature requestors",
			args: args{requestors: []types.Requestor{}},
			want: want{requestors: []types.Requestor{stakingRequestor}, length: 0, stakedRequestors: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			for _, requestor := range tt.args.requestors {
				keeper.SetRequestor(context, requestor)
			}
			if got := keeper.getMatureRequestors(context); len(got) != tt.want.length {
				t.Errorf("requestorUnstaked.unstakeAllMatureRequestors()= %v, want %v", len(got), tt.want.length)
			}
		})
	}
}

//func TestRequestorUnstaked_UnstakeAllMatureRequestors(t *testing.T) {
//	stakingRequestor := getUnstakingRequestor()
//
//	type want struct {
//		requestors       []types.Requestor
//		stakedRequestors bool
//		length             int
//	}
//	type args struct {
//		stakedVal         types.Requestor
//		requestors      []types.Requestor
//		stakedRequestor types.Requestor
//	}
//	tests := []struct {
//		name         string
//		requestor  types.Requestor
//		requestors []types.Requestor
//		want
//		args
//	}{
//		{
//			name: "unstake mature requestors",
//			args: args{requestors: []types.Requestor{stakingRequestor}},
//			want: want{requestors: []types.Requestor{stakingRequestor}, length: 0, stakedRequestors: false},
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			context, _, keeper := createTestInput(t, true)
//			for _, requestor := range tt.args.requestors {
//				keeper.SetRequestor(context, requestor)
//				keeper.SetUnstakingRequestor(context, requestor)
//			}
//			keeper.unstakeAllMatureRequestors(context)
//			if got := keeper.getAllUnstakingRequestors(context); len(got) != tt.want.length {
//				t.Errorf("requestorUnstaked.unstakeAllMatureRequestors()= %v, want %v", len(got), tt.want.length)
//			}
//		})
//	}
//}

func TestRequestorUnstaked_UnstakingRequestorsIterator(t *testing.T) {
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context, _, keeper := createTestInput(t, true)
			for _, requestor := range tt.requestors {
				keeper.SetRequestor(context, requestor)
				keeper.SetStakedRequestor(context, requestor)
			}

			it, _ := keeper.unstakingRequestorsIterator(context, context.BlockHeader().Time)
			if v, ok := it.(sdk.Iterator); !ok {
				t.Errorf("requestorUnstaked.UnstakingRequestorsIterator()= %v does not implement sdk.Iterator", v)
			}
		})
	}
}
