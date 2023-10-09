package mem

import (
	dbm "github.com/tendermint/tm-db"

	"github.com/vipernet-xyz/viper-network/store/types"

	"github.com/vipernet-xyz/viper-network/store/dbadapter"
	pruningtypes "github.com/vipernet-xyz/viper-network/store/pruning/types"
)

var (
	_ types.KVStore   = (*Store)(nil)
	_ types.Committer = (*Store)(nil)
)

// Store implements an in-memory only KVStore. Entries are persisted between
// commits and thus between blocks. State in Memory store is not committed as part of app state but maintained privately by each node
type Store struct {
	dbadapter.Store
}

func NewStore() *Store {
	return NewStoreWithDB(dbm.NewMemDB())
}

func NewStoreWithDB(db *dbm.MemDB) *Store { //nolint: interfacer
	return &Store{Store: dbadapter.Store{DB: db}}
}

// GetStoreType returns the Store's type.
func (s Store) GetStoreType() types.StoreType {
	return types.StoreTypeMemory
}

// Commit performs a no-op as entries are persistent between commitments.
func (s *Store) Commit() (id types.CommitID) { return }

// Implements CommitStore
func (s *Store) SetPruning(pruning types.PruningOptions) {
}
func (s *Store) GetPruning() pruningtypes.PruningOptions {
	return pruningtypes.NewPruningOptions(pruningtypes.PruningUndefined)
}
func (s Store) LastCommitID() (id types.CommitID) { return }
