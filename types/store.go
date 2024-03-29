package types

import (
	"github.com/vipernet-xyz/viper-network/store/types"

	"github.com/tendermint/tendermint/libs/kv"
)

// nolint - reexport
type (
	PruningOptions = types.PruningOptions
)

// nolint - reexport
type (
	Store            = types.Store
	Committer        = types.Committer
	CommitStore      = types.CommitStore
	Queryable        = types.Queryable
	MultiStore       = types.MultiStore
	CacheMultiStore  = types.CacheMultiStore
	CommitMultiStore = types.CommitMultiStore
	KVStore          = types.KVStore
	Iterator         = types.Iterator
)

// Iterator over all the keys with a certain prefix in ascending order
func KVStorePrefixIterator(kvs KVStore, prefix []byte) (Iterator, error) {
	return types.KVStorePrefixIterator(kvs, prefix)
}

// Iterator over all the keys with a certain prefix in descending order.
func KVStoreReversePrefixIterator(kvs KVStore, prefix []byte) (Iterator, error) {
	return types.KVStoreReversePrefixIterator(kvs, prefix)
}

// Compare two KVstores, return either the first key/value pair
// at which they differ and whether or not they are equal, skipping
// value comparison for a set of provided prefixes
func DiffKVStores(a KVStore, b KVStore, prefixesToSkip [][]byte) (kvA kv.Pair, kvB kv.Pair, count int64, equal bool) {
	return types.DiffKVStores(a, b, prefixesToSkip)
}

// nolint - reexport
type (
	CacheKVStore  = types.CacheKVStore
	CommitKVStore = types.CommitKVStore
	CacheWrap     = types.CacheWrap
	CacheWrapper  = types.CacheWrapper
	CommitID      = types.CommitID
)

// nolint - reexport
type StoreType = types.StoreType

// nolint - reexport
const (
	StoreTypeMulti     = types.StoreTypeMulti
	StoreTypeDB        = types.StoreTypeDB
	StoreTypeIAVL      = types.StoreTypeIAVL
	StoreTypeTransient = types.StoreTypeTransient
	StoreTypeMemory    = types.StoreTypeMemory
)

// nolint - reexport
type (
	StoreKey          = types.StoreKey
	KVStoreKey        = types.KVStoreKey
	TransientStoreKey = types.TransientStoreKey
	MemoryStoreKey    = types.MemoryStoreKey
)

// NewKVStoreKey returns a new pointer to a KVStoreKey.
// Use a pointer so keys don't collide.
func NewKVStoreKey(name string) *KVStoreKey {
	return types.NewKVStoreKey(name)
}

// NewKVStoreKeys returns a map of new  pointers to KVStoreKey's.
// Uses pointers so keys don't collide.
func NewKVStoreKeys(names ...string) map[string]*KVStoreKey {
	keys := make(map[string]*KVStoreKey)
	for _, name := range names {
		keys[name] = NewKVStoreKey(name)
	}
	return keys
}

// Constructs new TransientStoreKey
// Must return a pointer according to the ocap principle
func NewTransientStoreKey(name string) *TransientStoreKey {
	return types.NewTransientStoreKey(name)
}

// NewTransientStoreKeys constructs a new map of TransientStoreKey's
// Must return pointers according to the ocap principle
func NewTransientStoreKeys(names ...string) map[string]*TransientStoreKey {
	keys := make(map[string]*TransientStoreKey)
	for _, name := range names {
		keys[name] = NewTransientStoreKey(name)
	}
	return keys
}

func NewMemoryStoreKey(name string) *MemoryStoreKey {
	return types.NewMemoryStoreKey(name)
}

func NewMemoryStoreKeys(names ...string) map[string]*MemoryStoreKey {
	keys := make(map[string]*MemoryStoreKey)
	for _, name := range names {
		keys[name] = NewMemoryStoreKey(name)
	}
	return keys
}

// PrefixEndBytes returns the []byte that would end a
// range query for all []byte with a certain prefix
// Deals with last byte of prefix being FF without overflowing
func PrefixEndBytes(prefix []byte) []byte {
	return types.PrefixEndBytes(prefix)
}

// InclusiveEndBytes returns the []byte that would end a
// range query such that the input would be included
func InclusiveEndBytes(inclusiveBytes []byte) (exclusiveBytes []byte) {
	return types.InclusiveEndBytes(inclusiveBytes)
}

//----------------------------------------

// key-value result for iterator queries
type KVPair = types.KVPair

type KvPairs []types.KVPair

//----------------------------------------

// TraceContext contains TraceKVStore context data. It will be written with
// every trace operation.
type TraceContext = types.TraceContext

// --------------------------------------

// nolint - reexport
type (
	Gas       = types.Gas
	GasMeter  = types.GasMeter
	GasConfig = types.GasConfig
)

// nolint - reexport
func NewGasMeter(limit Gas) GasMeter {
	return types.NewGasMeter(limit)
}

// nolint - reexport
type (
	ErrorOutOfGas    = types.ErrorOutOfGas
	ErrorGasOverflow = types.ErrorGasOverflow
)

// nolint - reexport
func NewInfiniteGasMeter() GasMeter {
	return types.NewInfiniteGasMeter()
}
