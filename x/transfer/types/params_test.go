package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/vipernet-xyz/viper-network/x/transfer/types"
)

func TestValidateParams(t *testing.T) {
	require.NoError(t, types.DefaultParams().Validate())
	require.NoError(t, types.NewParams(true, false).Validate())
}
