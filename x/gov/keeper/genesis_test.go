package keeper

import (
	"testing"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/gov/types"

	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestInitGenesis(t *testing.T) {
	gs := types.GenesisState{
		Params: types.Params{
			ACL:      createTestACL(),
			DAOOwner: getRandomValidatorAddress(),
			Upgrade:  types.Upgrade{},
		},
		DAOTokens: sdk.NewInt(1000),
	}
	ctx, k := createTestKeeperAndContext(t, false)
	assert.Equal(t, k.InitGenesis(ctx, gs), []abci.ValidatorUpdate{})
}

func TestExportGenesis(t *testing.T) {
	ctx, k := createTestKeeperAndContext(t, false)
	d := types.DefaultGenesisState()
	d.Params.ACL = createTestACL()
	d.Params.Upgrade = types.Upgrade{}
	assert.Equal(t, k.ExportGenesis(ctx).Params.ACL.String(), d.Params.ACL.String())
	assert.Equal(t, k.ExportGenesis(ctx).DAOTokens.Int64(), d.DAOTokens.Int64())
}
