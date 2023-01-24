package keeper

import (
	"reflect"
	"testing"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/platforms/types"

	"github.com/tendermint/go-amino"
	abci "github.com/tendermint/tendermint/abci/types"
)

func Test_queryPlatforms(t *testing.T) {
	type args struct {
		ctx sdk.Context
		req abci.RequestQuery
		k   Keeper
	}
	context, _, keeper := createTestInput(t, true)
	jsondata, _ := amino.MarshalJSON(types.QueryPlatformsWithOpts{
		Page:  1,
		Limit: 1,
	})

	expectedPlatformsPage := types.PlatformsPage{Result: []types.Platform{}, Total: 1, Page: 1}
	jsonresponse, _ := amino.MarshalJSONIndent(expectedPlatformsPage, "", "  ")
	tests := []struct {
		name  string
		args  args
		want  []byte
		want1 sdk.Error
	}{
		{"Test query platformlicaitons", args{
			ctx: context,
			req: abci.RequestQuery{Data: jsondata, Path: "unstaking_validators"},
			k:   keeper,
		}, jsonresponse, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := queryPlatforms(tt.args.ctx, tt.args.req, tt.args.k)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("queryPlatforms() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("queryPlatforms() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_queryPlatform(t *testing.T) {
	type args struct {
		ctx sdk.Context
		req abci.RequestQuery
		k   Keeper
	}

	context, _, keeper := createTestInput(t, true)
	addr := getRandomPlatformAddress()
	jsondata, _ := amino.MarshalJSON(types.QueryPlatformParams{Address: addr})
	var jsonresponse []byte

	tests := []struct {
		name  string
		args  args
		want  []byte
		want1 sdk.Error
	}{
		{"Test query platformlicaiton", args{
			ctx: context,
			req: abci.RequestQuery{Data: jsondata, Path: "unstaking_validators"},
			k:   keeper,
		}, jsonresponse, types.ErrNoPlatformFound(types.DefaultCodespace)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := queryPlatform(tt.args.ctx, tt.args.req, tt.args.k)
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
		ctx sdk.Context
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
		ctx sdk.Context
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
		ctx  sdk.Context
		req  abci.RequestQuery
		path []string
		k    Keeper
	}
	context, _, keeper := createTestInput(t, true)
	jsondata, _ := amino.MarshalJSON(types.QueryPlatformsWithOpts{
		Page:  1,
		Limit: 1,
	})
	expectedPlatformsPage := types.PlatformsPage{Result: []types.Platform{}, Total: 1, Page: 1}
	jsonresponse, _ := amino.MarshalJSONIndent(expectedPlatformsPage, "", "  ")
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
			name: "Test queryPlatforms",
			args: args{
				ctx:  context,
				req:  abci.RequestQuery{Data: jsondata, Path: "unstaking_validators"},
				path: []string{types.QueryPlatforms},
				k:    keeper,
			},
			want:  jsonresponse,
			want1: nil,
		},
		{
			name: "Test query platform",
			args: args{
				ctx:  context,
				req:  abci.RequestQuery{Data: jsondata, Path: "unstaking_validators"},
				path: []string{types.QueryPlatform},
				k:    keeper,
			},
			want:  []byte(nil),
			want1: types.ErrNoPlatformFound(types.DefaultCodespace),
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
