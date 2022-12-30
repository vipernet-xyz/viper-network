package store

import (
	dbm "github.com/tendermint/tm-db"

	"github.com/vipernet-xyz/viper-network/store/rootmulti"
	"github.com/vipernet-xyz/viper-network/store/types"
)

func NewCommitMultiStore(db dbm.DB, cache bool, iavlCacheSize int64) types.CommitMultiStore {
	return rootmulti.NewStore(db, cache, iavlCacheSize)
}

func NewPruningOptionsFromString(strategy string) (opt PruningOptions) {
	switch strategy {
	case "nothing":
		opt = PruneNothing
	case "everything":
		opt = PruneEverything
	case "syncable":
		opt = PruneSyncable
	default:
		opt = PruneSyncable
	}
	return
}
