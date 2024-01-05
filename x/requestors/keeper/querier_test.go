package keeper

import (
	"reflect"
	"testing"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/requestors/types"

	"github.com/tendermint/go-amino"
	abci "github.com/tendermint/tendermint/abci/types"
)

func Test_queryRequestors(t *testing.T) {
	type args struct {
		ctx sdk.Ctx
		req abci.RequestQuery
		k   Keeper
	}
	context, _, keeper := createTestInput(t, true)
	jsondata, _ := amino.MarshalJSON(types.QueryRequestorsWithOpts{
		Page:  1,
		Limit: 1,
	})

	expectedRequestorsPage := types.RequestorsPage{Result: []types.Requestor{}, Total: 1, Page: 1}
	jsonresponse, _ := amino.MarshalJSONIndent(expectedRequestorsPage, "", "  ")
	tests := []struct {
		name  string
		args  args
		want  []byte
		want1 sdk.Error
	}{
		{"Test query requestorlicaitons", args{
			ctx: context,
			req: abci.RequestQuery{Data: jsondata, Path: "unstaking_validators"},
			k:   keeper,
		}, jsonresponse, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := queryRequestors(tt.args.ctx, tt.args.req, tt.args.k)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("queryRequestors() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("queryRequestors() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_queryRequestor(t *testing.T) {
	type args struct {
		ctx sdk.Ctx
		req abci.RequestQuery
		k   Keeper
	}

	context, _, keeper := createTestInput(t, true)
	addr := getRandomRequestorAddress()
	jsondata, _ := amino.MarshalJSON(types.QueryRequestorParams{Address: addr})
	var jsonresponse []byte

	tests := []struct {
		name  string
		args  args
		want  []byte
		want1 sdk.Error
	}{
		{"Test query requestorlicaiton", args{
			ctx: context,
			req: abci.RequestQuery{Data: jsondata, Path: "unstaking_validators"},
			k:   keeper,
		}, jsonresponse, types.ErrNoRequestorFound(types.DefaultCodespace)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := queryRequestor(tt.args.ctx, tt.args.req, tt.args.k)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("queryUnstakingValidators() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("queryUnstakingValidators() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_queryParameters(t *testing.T) {
	type args struct {
		ctx sdk.Ctx
		k   Keeper
	}
	context, _, keeper := createTestInput(t, true)
	jsonresponse, _ := amino.MarshalJSONIndent(keeper.GetParams(context), "", "  ")
	tests := []struct {
		name  string
		args  args
		want  []byte
		want1 sdk.Error
	}{
		{"Test Queryparameters", args{
			ctx: context,
			k:   keeper,
		}, jsonresponse, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := queryParameters(tt.args.ctx, tt.args.k)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("queryParameters() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("queryParameters() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_queryStakedPool(t *testing.T) {
	type args struct {
		ctx sdk.Ctx
		k   Keeper
	}
	context, _, keeper := createTestInput(t, true)
	jsonresponse, _ := amino.MarshalJSONIndent(sdk.ZeroInt(), "", "  ")
	tests := []struct {
		name  string
		args  args
		want  []byte
		want1 sdk.Error
	}{
		{"test QueryStakedPool", args{
			ctx: context,
			k:   keeper,
		}, jsonresponse, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := queryStakedPool(tt.args.ctx, tt.args.k)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("queryStakedPool() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("queryStakedPool() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_NewQuerier(t *testing.T) {
	type args struct {
		ctx  sdk.Ctx
		req  abci.RequestQuery
		path []string
		k    Keeper
	}
	context, _, keeper := createTestInput(t, true)
	jsondata, _ := amino.MarshalJSON(types.QueryRequestorsWithOpts{
		Page:  1,
		Limit: 1,
	})
	expectedRequestorsPage := types.RequestorsPage{Result: []types.Requestor{}, Total: 1, Page: 1}
	jsonresponse, _ := amino.MarshalJSONIndent(expectedRequestorsPage, "", "  ")
	jsonresponseForParams, _ := amino.MarshalJSONIndent(keeper.GetParams(context), "", "  ")
	tests := []struct {
		name  string
		args  args
		want  []byte
		want1 sdk.Error
	}{
		{
			name: "Test queryParams",
			args: args{
				ctx:  context,
				req:  abci.RequestQuery{Data: jsondata, Path: "unstaking_validators"},
				path: []string{types.QueryParameters},
				k:    keeper,
			},
			want:  jsonresponseForParams,
			want1: nil,
		},
		{
			name: "Test queryRequestors",
			args: args{
				ctx:  context,
				req:  abci.RequestQuery{Data: jsondata, Path: "unstaking_validators"},
				path: []string{types.QueryRequestors},
				k:    keeper,
			},
			want:  jsonresponse,
			want1: nil,
		},
		{
			name: "Test query requestor",
			args: args{
				ctx:  context,
				req:  abci.RequestQuery{Data: jsondata, Path: "unstaking_validators"},
				path: []string{types.QueryRequestor},
				k:    keeper,
			},
			want:  []byte(nil),
			want1: types.ErrNoRequestorFound(types.DefaultCodespace),
		},
		{
			name: "Test default querier",
			args: args{
				ctx:  context,
				req:  abci.RequestQuery{Data: jsondata, Path: "unstaking_validators"},
				path: []string{"query"},
				k:    keeper,
			},
			want:  []byte(nil),
			want1: sdk.ErrUnknownRequest("unknown staking query endpoint"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := NewQuerier(tt.args.k)
			got, got1 := fn(tt.args.ctx, tt.args.path, tt.args.req)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("queryUnstakingValidators() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("queryUnstakingValidators() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
