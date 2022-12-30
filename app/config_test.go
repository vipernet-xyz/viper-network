package app

import (
	"testing"

	"github.com/vipernet-xyz/viper-network/types"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	c := types.DefaultConfig("~/.viper")

	// Check default Tx indexing params
	assert.EqualValues(t, types.DefaultTxIndexer, c.TendermintConfig.TxIndex.Indexer)
	assert.EqualValues(t, types.DefaultTxIndexTags, c.TendermintConfig.TxIndex.IndexKeys)
}
