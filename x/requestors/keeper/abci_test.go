package keeper

import (
	"testing"

	sdk "github.com/vipernet-xyz/viper-network/types"

	abci "github.com/tendermint/tendermint/abci/types"
)

func TestBeginBlocker(t *testing.T) {
	type args struct {
		ctx sdk.Ctx
		req abci.RequestBeginBlock
		k   Keeper
	}
	context, _, keeper := createTestInput(t, true)
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Test BeginBlocker",
			args: args{
				ctx: context,
				req: abci.RequestBeginBlock{
					Hash:                 []byte{0x51, 0x51, 0x51},
					Header:               abci.Header{},
					LastCommitInfo:       abci.LastCommitInfo{},
					ByzantineValidators:  nil,
					XXX_NoUnkeyedLiteral: struct{}{},
					XXX_unrecognized:     nil,
					XXX_sizecache:        0,
				},
				k: keeper,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			BeginBlocker(tt.args.ctx, tt.args.req, tt.args.k)
		})
	}
}

//func TestEndBlocker(t *testing.T) {
//	type args struct {
//		ctx  sdk.Ctx
//		k    Keeper
//		requestors []types.Requestor
//	}
//	context, _, keeper := createTestInput(t, true)
//	requestor := getUnstakingRequestor()
//
//	keeper.SetRequestor(context, requestor)
//	keeper.SetUnstakingRequestor(context, requestor)
//
//	tests := []struct {
//		name string
//		args args
//		want []abci.ValidatorUpdate
//	}{
//		{
//			name: "Test EndBlocker",
//			args: args{
//				ctx:  context,
//				k:    keeper,
//				requestors: []types.Requestor{requestor},
//			},
//			want: []abci.ValidatorUpdate{},
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if got := EndBlocker(tt.args.ctx, tt.args.k); !assert.True(t, len(got) == len(tt.want)) {
//				t.Errorf("EndBlocker() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
