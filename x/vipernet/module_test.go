package vipernet

import (
	"encoding/hex"
	"reflect"
	"testing"

	"github.com/vipernet-xyz/viper-network/x/vipernet/keeper"
	"github.com/vipernet-xyz/viper-network/x/vipernet/types"

	"github.com/stretchr/testify/assert"
)

func TestProviderModule_Name(t *testing.T) {
	_, _, _, k, _ := createTestInput(t, false)
	pm := NewProviderModule(k)
	assert.Equal(t, pm.Name(), types.ModuleName)
	assert.Equal(t, pm.Name(), types.ModuleName)
}

func TestProviderModule_InitExportGenesis(t *testing.T) {
	p := types.Params{
		SessionNodeCount:      10,
		ClaimSubmissionWindow: 22,
		SupportedBlockchains:  []string{"eth"},
		ClaimExpiration:       55,
	}
	genesisState := types.GenesisState{
		Params: p,
		Claims: []types.MsgClaim(nil),
	}
	ctx, _, _, k, _ := createTestInput(t, false)
	pm := NewProviderModule(k)
	data, err := types.ModuleCdc.MarshalJSON(genesisState)
	assert.Nil(t, err)
	pm.InitGenesis(ctx, data)
	genesisbz := pm.ExportGenesis(ctx)
	var genesis types.GenesisState
	err = types.ModuleCdc.UnmarshalJSON(genesisbz, &genesis)
	assert.Nil(t, err)
	assert.Equal(t, genesis, genesisState)
	pm.InitGenesis(ctx, nil)
	var genesis2 types.GenesisState
	genesis2bz := pm.ExportGenesis(ctx)
	err = types.ModuleCdc.UnmarshalJSON(genesis2bz, &genesis2)
	assert.Equal(t, genesis2, types.DefaultGenesisState())
	assert.Nil(t, err)
}

func TestProviderModule_NewQuerierHandler(t *testing.T) {
	_, _, _, k, _ := createTestInput(t, false)
	pm := NewProviderModule(k)
	assert.Equal(t, reflect.ValueOf(keeper.NewQuerier(k)).String(), reflect.ValueOf(pm.NewQuerierHandler()).String())
}

func TestProviderModule_Route(t *testing.T) {
	_, _, _, k, _ := createTestInput(t, false)
	pm := NewProviderModule(k)
	assert.Equal(t, pm.Route(), types.RouterKey)
}

func TestProviderModule_QuerierRoute(t *testing.T) {
	_, _, _, k, _ := createTestInput(t, false)
	pm := NewProviderModule(k)
	assert.Equal(t, pm.QuerierRoute(), types.ModuleName)
}

//func TestProviderModule_EndBlock(t *testing.T) {
//	ctx, _, _, k, _ := createTestInput(t, false)
//	pm := NewProviderModule(k)
//	assert.Equal(t, pm.EndBlock(ctx, abci.RequestEndBlock{}), []abci.ValidatorUpdate{})
//}

func TestProviderModuleBasic_DefaultGenesis(t *testing.T) {
	_, _, _, k, _ := createTestInput(t, false)
	pm := NewProviderModule(k)
	assert.Equal(t, []byte(pm.DefaultGenesis()), []byte(types.ModuleCdc.MustMarshalJSON(types.DefaultGenesisState())))
}

func TestProviderModuleBasic_ValidateGenesis(t *testing.T) {
	_, _, _, k, _ := createTestInput(t, false)
	pm := NewProviderModule(k)
	p := types.Params{
		SessionNodeCount:      10,
		ClaimSubmissionWindow: 22,
		SupportedBlockchains:  []string{hex.EncodeToString([]byte{01})},
		ClaimExpiration:       55,
	}
	genesisState := types.GenesisState{
		Params: p,
		Claims: []types.MsgClaim(nil),
	}
	p2 := types.Params{
		SessionNodeCount:      -1,
		ClaimSubmissionWindow: 22,
		SupportedBlockchains:  []string{"eth"},
		ClaimExpiration:       55,
	}
	genesisState2 := types.GenesisState{
		Params: p2,
		Claims: []types.MsgClaim(nil),
	}
	validBz, err := types.ModuleCdc.MarshalJSON(genesisState)
	assert.Nil(t, err)
	invalidBz, err := types.ModuleCdc.MarshalJSON(genesisState2)
	assert.Nil(t, err)
	assert.True(t, nil == pm.ValidateGenesis(validBz))
	assert.False(t, nil == pm.ValidateGenesis(invalidBz))
}
