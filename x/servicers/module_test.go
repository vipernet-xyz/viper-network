package servicers

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/vipernet-xyz/viper-network/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/servicers/keeper"
	"github.com/vipernet-xyz/viper-network/x/servicers/types"

	abci "github.com/tendermint/tendermint/abci/types"
)

func TestProviderModuleBasic_DefaultGenesis(t *testing.T) {
	tests := []struct {
		name string
		want json.RawMessage
	}{
		{"Test DefaultGenesis", types.ModuleCdc.MustMarshalJSON(types.DefaultGenesisState())},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ap := ProviderModuleBasic{}
			if got := ap.DefaultGenesis(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DefaultGenesis() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProviderModuleBasic_Name(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ap := ProviderModuleBasic{}
			if got := ap.Name(); got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProviderModuleBasic_RegisterCodec(t *testing.T) {
	type args struct {
		cdc *codec.Codec
	}
	tests := []struct {
		name string
		args args
	}{
		{"Test RegisterCodec", args{cdc: makeTestCodec()}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ap := ProviderModuleBasic{}
			ap.RegisterCodec(tt.args.cdc)
		})
	}
}

func TestProviderModuleBasic_ValidateGenesis(t *testing.T) {
	type args struct {
		bz json.RawMessage
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Test ValidateGenesis", args{bz: types.ModuleCdc.MustMarshalJSON(types.DefaultGenesisState())}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ap := ProviderModuleBasic{}
			if err := ap.ValidateGenesis(tt.args.bz); (err != nil) != tt.wantErr {
				t.Errorf("ValidateGenesis() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProviderModule_BeginBlock(t *testing.T) {
	type fields struct {
		ProviderModuleBasic ProviderModuleBasic
		keeper              keeper.Keeper
		accountKeeper       types.AuthKeeper
		supplyKeeper        types.AuthKeeper
	}
	type args struct {
		ctx sdk.Context
		req abci.RequestBeginBlock
	}

	ctx, _, k := createTestInput(t, true)
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{"Test BeginBlock", fields{
			ProviderModuleBasic: ProviderModuleBasic{},
			keeper:              k,
			accountKeeper:       nil,
			supplyKeeper:        nil,
		}, args{
			ctx: ctx,
			req: abci.RequestBeginBlock{},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := ProviderModule{
				ProviderModuleBasic: tt.fields.ProviderModuleBasic,
				keeper:              tt.fields.keeper,
			}
			pm.BeginBlock(tt.args.ctx, tt.args.req)
		})
	}
}

func TestProviderModule_EndBlock(t *testing.T) {
	type fields struct {
		ProviderModuleBasic ProviderModuleBasic
		keeper              keeper.Keeper
		accountKeeper       types.AuthKeeper
		supplyKeeper        types.AuthKeeper
	}
	type args struct {
		ctx sdk.Context
		in1 abci.RequestEndBlock
	}

	ctx, _, k := createTestInput(t, true)

	tests := []struct {
		name   string
		fields fields
		args   args
		want   []abci.ValidatorUpdate
	}{
		{"Test EndBlock", fields{
			ProviderModuleBasic: ProviderModuleBasic{},
			keeper:              k,
			accountKeeper:       nil,
			supplyKeeper:        nil,
		}, args{
			ctx: ctx,
			in1: abci.RequestEndBlock{},
		}, []abci.ValidatorUpdate{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := ProviderModule{
				ProviderModuleBasic: tt.fields.ProviderModuleBasic,
				keeper:              tt.fields.keeper,
			}
			if got := pm.EndBlock(tt.args.ctx, tt.args.in1); !reflect.DeepEqual(len(got), len(tt.want)) {
				t.Errorf("EndBlock() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProviderModule_ExportGenesis(t *testing.T) {
	type fields struct {
		ProviderModuleBasic ProviderModuleBasic
		keeper              keeper.Keeper
		accountKeeper       types.AuthKeeper
		supplyKeeper        types.AuthKeeper
	}
	context, _, k := createTestInput(t, true)

	k.SetPreviousProposer(context, sdk.GetAddress(getRandomPubKey()))
	type args struct {
		ctx sdk.Context
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   json.RawMessage
	}{
		{"Test Export Genesis", fields{
			ProviderModuleBasic: ProviderModuleBasic{},
			keeper:              k,
			accountKeeper:       nil,
			supplyKeeper:        nil,
		}, args{ctx: context}, types.ModuleCdc.MustMarshalJSON(ExportGenesis(context, k))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := ProviderModule{
				ProviderModuleBasic: tt.fields.ProviderModuleBasic,
				keeper:              tt.fields.keeper,
			}
			if got := pm.ExportGenesis(tt.args.ctx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExportGenesis() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestProviderModule_InitGenesis(t *testing.T) {
	type fields struct {
		ProviderModuleBasic ProviderModuleBasic
		keeper              keeper.Keeper
	}
	type args struct {
		ctx  sdk.Context
		data json.RawMessage
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []abci.ValidatorUpdate
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := ProviderModule{
				ProviderModuleBasic: tt.fields.ProviderModuleBasic,
				keeper:              tt.fields.keeper,
			}
			if got := pm.InitGenesis(tt.args.ctx, tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InitGenesis() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProviderModule_Name(t *testing.T) {
	type fields struct {
		ProviderModuleBasic ProviderModuleBasic
		keeper              keeper.Keeper
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ap := ProviderModule{
				ProviderModuleBasic: tt.fields.ProviderModuleBasic,
				keeper:              tt.fields.keeper,
			}
			if got := ap.Name(); got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProviderModule_NewHandler(t *testing.T) {
	type fields struct {
		ProviderModuleBasic ProviderModuleBasic
		keeper              keeper.Keeper
		accountKeeper       types.AuthKeeper
		supplyKeeper        types.AuthKeeper
	}

	_, _, k := createTestInput(t, true)

	tests := []struct {
		name   string
		fields fields
		want   sdk.Handler
	}{
		{"Test NewHandler", fields{
			ProviderModuleBasic: ProviderModuleBasic{},
			keeper:              k,
			accountKeeper:       nil,
			supplyKeeper:        nil,
		}, NewHandler(k)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := ProviderModule{
				ProviderModuleBasic: tt.fields.ProviderModuleBasic,
				keeper:              tt.fields.keeper,
			}
			pm.NewHandler()
		})
	}
}

func TestProviderModule_NewQuerierHandler(t *testing.T) {
	type fields struct {
		ProviderModuleBasic ProviderModuleBasic
		keeper              keeper.Keeper
		accountKeeper       types.AuthKeeper
		supplyKeeper        types.AuthKeeper
	}
	tests := []struct {
		name   string
		fields fields
		want   sdk.Querier
	}{
		{"Test Querier Handler", fields{
			ProviderModuleBasic: ProviderModuleBasic{},
			keeper:              keeper.Keeper{},
			accountKeeper:       nil,
			supplyKeeper:        nil,
		}, keeper.NewQuerier(keeper.Keeper{})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := ProviderModule{
				ProviderModuleBasic: tt.fields.ProviderModuleBasic,
				keeper:              tt.fields.keeper,
			}
			pm.NewQuerierHandler()
		})
	}
}

func TestProviderModule_QuerierRoute(t *testing.T) {
	type fields struct {
		ProviderModuleBasic ProviderModuleBasic
		keeper              keeper.Keeper
		accountKeeper       types.AuthKeeper
		supplyKeeper        types.AuthKeeper
	}

	_, _, k := createTestInput(t, true)

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"Test QuerierRoute", fields{
			ProviderModuleBasic: ProviderModuleBasic{},
			keeper:              k,
			accountKeeper:       nil,
			supplyKeeper:        nil,
		}, types.ModuleName},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ap := ProviderModule{
				ProviderModuleBasic: tt.fields.ProviderModuleBasic,
				keeper:              tt.fields.keeper,
			}
			if got := ap.QuerierRoute(); got != tt.want {
				t.Errorf("QuerierRoute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProviderModule_Route(t *testing.T) {
	type fields struct {
		ProviderModuleBasic ProviderModuleBasic
		keeper              keeper.Keeper
		accountKeeper       types.AuthKeeper
		supplyKeeper        types.AuthKeeper
	}
	_, _, keeper := createTestInput(t, true)
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"test Route", fields{
			ProviderModuleBasic: ProviderModuleBasic{},
			keeper:              keeper,
			accountKeeper:       nil,
			supplyKeeper:        nil,
		}, types.ModuleName},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ap := ProviderModule{
				ProviderModuleBasic: tt.fields.ProviderModuleBasic,
				keeper:              tt.fields.keeper,
			}
			if got := ap.Route(); got != tt.want {
				t.Errorf("Route() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewProviderModule(t *testing.T) {
	type args struct {
		keeper keeper.Keeper
	}
	tests := []struct {
		name string
		args args
		want ProviderModule
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewProviderModule(tt.args.keeper); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewProviderModule() = %v, want %v", got, tt.want)
			}
		})
	}
}
