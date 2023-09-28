package rootmulti

import (
	"fmt"
	"io"
	"log"
	"strings"

	sdk "github.com/vipernet-xyz/viper-network/types"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/merkle"
	"github.com/tendermint/tendermint/crypto/tmhash"
	dbm "github.com/tendermint/tm-db"

	"github.com/vipernet-xyz/viper-network/store/cachemulti"
	"github.com/vipernet-xyz/viper-network/store/dbadapter"
	"github.com/vipernet-xyz/viper-network/store/errors"
	"github.com/vipernet-xyz/viper-network/store/iavl"
	"github.com/vipernet-xyz/viper-network/store/mem"
	"github.com/vipernet-xyz/viper-network/store/rootmulti/heightcache"
	"github.com/vipernet-xyz/viper-network/store/tracekv"
	"github.com/vipernet-xyz/viper-network/store/transient"
	"github.com/vipernet-xyz/viper-network/store/types"
)

const (
	latestVersionKey = "s/latest"
	commitInfoKeyFmt = "s/%d" // s/<version>
)

// Store is composed of many CommitStores. Name contrasts with
// cacheMultiStore which is for cache-wrapping other MultiStores. It implements
// the CommitMultiStore interface.
type Store struct {
	DB           dbm.DB
	Cache        types.MultiStoreCache
	lastCommitID types.CommitID
	pruningOpts  types.PruningOptions
	storesParams map[types.StoreKey]storeParams
	stores       map[types.StoreKey]types.CommitStore
	keysByName   map[string]types.StoreKey
	lazyLoading  bool

	traceWriter   io.Writer
	traceContext  types.TraceContext
	iavlCacheSize int64
}

func (rs *Store) CopyStore() *types.Store {
	newParams := make(map[types.StoreKey]storeParams)
	for k, v := range rs.storesParams {
		newParams[k] = v
	}
	newStores := make(map[types.StoreKey]types.CommitStore)
	for k, v := range rs.stores {
		newStores[k] = v
	}
	newKeysByName := make(map[string]types.StoreKey)
	for k, v := range rs.keysByName {
		newKeysByName[k] = v
	}
	newTraceCtx := map[string]interface{}{}
	for k, v := range rs.traceContext {
		newTraceCtx[k] = v
	}
	s := types.Store(&Store{
		DB:           rs.DB,
		Cache:        rs.Cache,
		lastCommitID: rs.lastCommitID,
		pruningOpts:  rs.pruningOpts,
		storesParams: newParams,
		stores:       newStores,
		keysByName:   newKeysByName,
		lazyLoading:  rs.lazyLoading,
		traceWriter:  rs.traceWriter,
		traceContext: newTraceCtx,
	})
	return &s
}

var _ types.CommitMultiStore = (*Store)(nil)
var _ types.Queryable = (*Store)(nil)

const MemoryCacheCapacity = 12

func NewStore(db dbm.DB, cache bool, iavlCacheSize int64) *Store {
	var multiStoreCache types.MultiStoreCache
	if cache {
		multiStoreCache = heightcache.NewMultiStoreMemoryCache(MemoryCacheCapacity)
	} else {
		multiStoreCache = heightcache.NewMultiStoreInvalidCache()
	}

	return &Store{
		DB:            db,
		Cache:         multiStoreCache,
		storesParams:  make(map[types.StoreKey]storeParams),
		stores:        make(map[types.StoreKey]types.CommitStore),
		keysByName:    make(map[string]types.StoreKey),
		iavlCacheSize: iavlCacheSize,
	}
}

// Implements CommitMultiStore
func (rs *Store) SetPruning(pruningOpts types.PruningOptions) {
	rs.pruningOpts = pruningOpts
	for _, substore := range rs.stores {
		substore.SetPruning(pruningOpts)
	}
}

// SetLazyLoading sets if the iavl store should be loaded lazily or not
func (rs *Store) SetLazyLoading(lazyLoading bool) {
	rs.lazyLoading = lazyLoading
}

// Implements Store.
func (rs *Store) GetStoreType() types.StoreType {
	return types.StoreTypeMulti
}

// Implements CommitMultiStore.
func (rs *Store) MountStoreWithDB(key types.StoreKey, typ types.StoreType, db dbm.DB) {
	if key == nil {
		panic("MountIAVLStore() Key cannot be nil")
	}
	if _, ok := rs.storesParams[key]; ok {
		panic(fmt.Sprintf("Store duplicate store Key %v", key))
	}
	if _, ok := rs.keysByName[key.Name()]; ok {
		panic(fmt.Sprintf("Store duplicate store Key name %v", key))
	}
	rs.storesParams[key] = storeParams{
		key: key,
		typ: typ,
		db:  db,
	}
	rs.keysByName[key.Name()] = key
}

// Implements CommitMultiStore.
func (rs *Store) GetCommitStore(key types.StoreKey) types.CommitStore {
	return rs.stores[key]
}

// Implements CommitMultiStore.
func (rs *Store) GetCommitKVStore(key types.StoreKey) types.CommitKVStore {
	return rs.stores[key].(types.CommitKVStore)
}

// Implements CommitMultiStore.
func (rs *Store) LoadLatestVersion() error {
	ver := getLatestVersion(rs.DB)
	return rs.LoadVersion(ver)
}

// Implements CommitMultiStore.
func (rs *Store) RollbackVersion(height int64) error {
	// ensure latest version is greater than rollback height
	ver := getLatestVersion(rs.DB)
	if height >= ver {
		return fmt.Errorf("the rollback height: %d must be less than the actual app height: %d", height, ver)
	}
	// create a new db
	b := rs.DB.NewBatch()
	// set the latest version so that we always start from this height
	setLatestVersion(b, height)
	// get commit info
	cInfo, err := getCommitInfo(rs.DB, ver)
	if err != nil {
		return err
	}
	// convert StoreInfos slice to map
	infos := make(map[types.StoreKey]StoreInfo)
	for _, storeInfo := range cInfo.StoreInfos {
		infos[rs.nameToKey(storeInfo.Name)] = storeInfo
	}
	// for each store
	for key, storeParams := range rs.storesParams {
		var id types.CommitID
		info, ok := infos[key]
		if ok {
			id = info.Core.CommitID
		}
		// load the kvStore using the commit id
		store, err := rs.loadCommitStoreFromParams(key, id, storeParams)
		if err != nil {
			return fmt.Errorf("failed to load Store: %v", err)
		}
		if store.GetStoreType() == types.StoreTypeTransient {
			continue
		}
		// convert to iavl
		s, ok := store.(*iavl.Store)
		if !ok {
			panic("store type not supported for rollback, (iavl or transient) type: " + fmt.Sprintf("%v", s.GetStoreType()))
		}
		// rollback to a certain height
		err = s.Rollback(height)
		if err != nil {
			return err
		}
		rs.stores[key] = s
	}
	// delete commit info from height + 1
	for i := height + 1; i <= ver; i++ {
		cInfoKey := fmt.Sprintf(commitInfoKeyFmt, i)
		b.Delete([]byte(cInfoKey))
	}
	// write to db
	_ = b.Write()
	return nil
}

// Implements CommitMultiStore.
func (rs *Store) LoadVersion(ver int64) error {
	if ver == 0 {
		// Special logic for version 0 where there is no need to get commit
		// information.
		for key, storeParams := range rs.storesParams {
			store, err := rs.loadCommitStoreFromParams(key, types.CommitID{}, storeParams)
			if err != nil {
				return fmt.Errorf("failed to load Store: %v", err)
			}

			rs.stores[key] = store
		}

		rs.lastCommitID = types.CommitID{}
		return nil
	}

	cInfo, err := getCommitInfo(rs.DB, ver)
	if err != nil {
		return err
	}

	// convert StoreInfos slice to map
	infos := make(map[types.StoreKey]StoreInfo)
	for _, storeInfo := range cInfo.StoreInfos {
		infos[rs.nameToKey(storeInfo.Name)] = storeInfo
	}

	// load each Store
	var newStores = make(map[types.StoreKey]types.CommitStore)
	for key, storeParams := range rs.storesParams {
		var id types.CommitID

		info, ok := infos[key]
		if ok {
			id = info.Core.CommitID
		}

		store, err := rs.loadCommitStoreFromParams(key, id, storeParams)
		if err != nil {
			return fmt.Errorf("failed to load Store: %v", err)
		}

		newStores[key] = store
	}

	rs.lastCommitID = cInfo.CommitID()
	rs.stores = newStores

	return nil
}

func (rs *Store) LoadLazyVersion(ver int64) (*types.Store, error) {
	newStores := make(map[types.StoreKey]types.CommitStore)
	for k, v := range rs.stores {
		a, ok := (v).(*iavl.Store)
		if !ok {
			if _, ok := (v).(*transient.Store); ok {
				continue
			}
			if _, ok := (v).(*mem.Store); ok {
				continue
			}
			return nil, fmt.Errorf("cannot convert store into iavl store in get immutable")
		}
		fmt.Println("version:", ver, "Store:", k)
		s, err := a.LazyLoadStore(ver, rs.Cache.GetSingleStoreCache(k))
		if err != nil {
			return nil, fmt.Errorf("error loading store: %s, in LoadLazyVersion: %s", k, err.Error())
		}
		newStores[k] = s
	}
	newParams := make(map[types.StoreKey]storeParams)
	for k, v := range rs.storesParams {
		newParams[k] = v
	}
	newKeysByName := make(map[string]types.StoreKey)
	for k, v := range rs.keysByName {
		newKeysByName[k] = v
	}
	newTraceCtx := map[string]interface{}{}
	for k, v := range rs.traceContext {
		newTraceCtx[k] = v
	}
	s := types.Store(&Store{
		DB:           rs.DB,
		lastCommitID: rs.lastCommitID,
		pruningOpts:  rs.pruningOpts,
		storesParams: newParams,
		stores:       newStores,
		keysByName:   newKeysByName,
		lazyLoading:  rs.lazyLoading,
		traceWriter:  rs.traceWriter,
		traceContext: newTraceCtx,
		Cache:        rs.Cache,
	})
	return &s, nil
}

// SetTracer sets the tracer for the MultiStore that the underlying
// stores will utilize to trace operations. A MultiStore is returned.
func (rs *Store) SetTracer(w io.Writer) types.MultiStore {
	rs.traceWriter = w
	return rs
}

// SetTracingContext updates the tracing context for the MultiStore by merging
// the given context with the existing context by Key. Any existing keys will
// be overwritten. It is implied that the caller should update the context when
// necessary between tracing operations. It returns a modified MultiStore.
func (rs *Store) SetTracingContext(tc types.TraceContext) types.MultiStore {
	if rs.traceContext != nil {
		for k, v := range tc {
			rs.traceContext[k] = v
		}
	} else {
		rs.traceContext = tc
	}

	return rs
}

// TracingEnabled returns if tracing is enabled for the MultiStore.
func (rs *Store) TracingEnabled() bool {
	return rs.traceWriter != nil
}

//----------------------------------------
// +CommitStore

// Implements Committer/CommitStore.
func (rs *Store) LastCommitID() types.CommitID {
	return rs.lastCommitID
}

// Implements Committer/CommitStore.
func (rs *Store) Commit() types.CommitID {

	// Commit stores.
	version := rs.lastCommitID.Version + 1
	commitInfo := commitStores(version, rs.stores)

	// Need to update atomically.
	batch := rs.DB.NewBatch()
	defer batch.Close()
	setCommitInfo(batch, version, commitInfo)
	setLatestVersion(batch, version)
	_ = batch.Write()

	// Prepare for next version.
	commitID := types.CommitID{
		Version: version,
		Hash:    commitInfo.Hash(),
	}
	rs.lastCommitID = commitID
	return commitID
}

// Implements CacheWrapper/Store/CommitStore.
func (rs *Store) CacheWrap() types.CacheWrap {
	return rs.CacheMultiStore().(types.CacheWrap)
}

// CacheWrapWithTrace implements the CacheWrapper interface.
func (rs *Store) CacheWrapWithTrace(_ io.Writer, _ types.TraceContext) types.CacheWrap {
	return rs.CacheWrap()
}

//----------------------------------------
// +MultiStore

// Implements MultiStore.
func (rs *Store) CacheMultiStore() types.CacheMultiStore {
	stores := make(map[types.StoreKey]types.CacheWrapper)
	for k, v := range rs.stores {
		stores[k] = v
	}

	return cachemulti.NewStore(rs.DB, stores, rs.keysByName, rs.traceWriter, rs.traceContext)
}

// CacheMultiStoreWithVersion is analogous to CacheMultiStore except that it
// attempts to load stores at a given version (height). An error is returned if
// any store cannot be loaded. This should only be used for querying and
// iterating at past heights.
func (rs *Store) CacheMultiStoreWithVersion(version int64) (types.CacheMultiStore, error) {
	cachedStores := make(map[types.StoreKey]types.CacheWrapper)
	for key, store := range rs.stores {
		switch store.GetStoreType() {
		case types.StoreTypeIAVL:
			// Attempt to lazy-load an already saved IAVL store version. If the
			// version does not exist or is pruned, an error should be returned.
			iavlStore, err := store.(*iavl.Store).LazyLoadStore(version, rs.Cache.GetSingleStoreCache(key))
			if err != nil {
				return nil, err
			}

			cachedStores[key] = iavlStore

		default:
			cachedStores[key] = store
		}
	}

	return cachemulti.NewStore(rs.DB, cachedStores, rs.keysByName, rs.traceWriter, rs.traceContext), nil
}

// Implements MultiStore.
// If the store does not exist, panics.
func (rs *Store) GetStore(key types.StoreKey) types.Store {
	store := rs.stores[key]
	if store == nil {
		panic("Could not load store " + key.String())
	}
	return store
}

// GetKVStore implements the MultiStore interface. If tracing is enabled on the
// Store, a wrapped TraceKVStore will be returned with the given
// tracer, otherwise, the original KVStore will be returned.
// If the store does not exist, panics.
func (rs *Store) GetKVStore(key types.StoreKey) types.KVStore {
	store := rs.stores[key].(types.KVStore)

	if rs.TracingEnabled() {
		store = tracekv.NewStore(store, rs.traceWriter, rs.traceContext)
	}

	return store
}

// Implements MultiStore

// getStoreByName will first convert the original name to
// a special Key, before looking up the CommitStore.
// This is not exposed to the extensions (which will need the
// storeKey), but is useful in main, and particularly app.Query,
// in order to convert human strings into CommitStores.
func (rs *Store) getStoreByName(name string) types.Store {
	key := rs.keysByName[name]
	if key == nil {
		return nil
	}
	return rs.stores[key]
}

//---------------------- Query ------------------

// Query calls substore.Query with the same `req` where `req.Path` is
// modified to remove the substore prefix.
// Ie. `req.Path` here is `/<substore>/<path>`, and trimmed to `/<path>` for the substore.
// TODO: add proof for `multistore -> substore`.
func (rs *Store) Query(req abci.RequestQuery) abci.ResponseQuery {
	// Query just routes this to a substore.
	path := req.Path
	storeName, subpath, err := parsePath(path)
	if err != nil {
		return err.QueryResult()
	}

	store := rs.getStoreByName(storeName)
	if store == nil {
		msg := fmt.Sprintf("no such store: %s", storeName)
		return errors.ErrUnknownRequest(msg).QueryResult()
	}

	queryable, ok := store.(types.Queryable)
	if !ok {
		msg := fmt.Sprintf("store %s doesn't support queries", storeName)
		return errors.ErrUnknownRequest(msg).QueryResult()
	}

	// trim the path and make the query
	req.Path = subpath
	res := queryable.Query(req)

	if !req.Prove || !RequireProof(subpath) {
		return res
	}

	if res.Proof == nil || len(res.Proof.Ops) == 0 {
		return errors.ErrInternal("proof is unexpectedly empty; ensure height has not been pruned").QueryResult()
	}

	commitInfo, errMsg := getCommitInfo(rs.DB, res.Height)
	if errMsg != nil {
		return errors.ErrInternal(errMsg.Error()).QueryResult()
	}

	proofOp := NewMultiStoreProofOp(
		[]byte(storeName),
		NewMultiStoreProof(commitInfo.StoreInfos),
	).ProofOp()
	// Restore origin path and append proof op.
	res.Proof.Ops = append(res.Proof.Ops, proofOp)

	// TODO: handle in another TM v0.26 update PR
	// res.Proof = buildMultiStoreProof(res.Proof, storeName, commitInfo.StoreInfos)
	return res
}

// parsePath expects a format like /<storeName>[/<subpath>]
// Must start with /, subpath may be empty
// Returns error if it doesn't start with /
func parsePath(path string) (storeName string, subpath string, err errors.Error) {
	if !strings.HasPrefix(path, "/") {
		err = errors.ErrUnknownRequest(fmt.Sprintf("invalid path: %s", path))
		return
	}

	paths := strings.SplitN(path[1:], "/", 2)
	storeName = paths[0]

	if len(paths) == 2 {
		subpath = "/" + paths[1]
	}

	return
}

//----------------------------------------

func (rs *Store) loadCommitStoreFromParams(key types.StoreKey, id types.CommitID, params storeParams) (store types.CommitStore, err error) {
	var db dbm.DB

	if params.db != nil {
		db = dbm.NewPrefixDB(params.db, []byte("s/_/"))
	} else {
		db = dbm.NewPrefixDB(rs.DB, []byte("s/k:"+params.key.Name()+"/"))
	}

	switch params.typ {
	case types.StoreTypeMulti:
		panic("recursive MultiStores not yet supported")

	case types.StoreTypeIAVL:
		cacheForStore := rs.Cache.GetSingleStoreCache(key)
		if cacheForStore.IsValid() {
			log.Printf("Warming up cache for %s\n", key.Name())
		}
		return iavl.LoadStore(db, id, rs.pruningOpts, rs.lazyLoading, cacheForStore, rs.iavlCacheSize)

	case types.StoreTypeDB:
		return commitDBStoreAdapter{dbadapter.Store{DB: db}}, nil

	case types.StoreTypeTransient:
		_, ok := key.(*types.TransientStoreKey)
		if !ok {
			return store, fmt.Errorf("invalid storeKey for StoreTypeTransient: %s", key.String())
		}
		return transient.NewStore(), nil
	case types.StoreTypeMemory:
		if _, ok := key.(*types.MemoryStoreKey); !ok {
			return nil, fmt.Errorf("unexpected key type for a MemoryStoreKey; got: %s", key.String())
		}
		return mem.NewStore(), nil

	default:
		panic(fmt.Sprintf("unrecognized store type %v", params.typ))
	}
}

func (rs *Store) nameToKey(name string) types.StoreKey {
	for key := range rs.storesParams {
		if key.Name() == name {
			return key
		}
	}
	panic("Unknown name " + name)
}

//----------------------------------------
// storeParams

type storeParams struct {
	key types.StoreKey
	db  dbm.DB
	typ types.StoreType
}

//----------------------------------------
// commitInfo

// NOTE: Keep commitInfo a simple immutable struct.
// type commitInfo struct {
// 	// Version
// 	Version int64

// 	// Store info for
// 	StoreInfos []StoreInfo
// }

// Hash returns the simple merkle root hash of the stores sorted by name.
func (ci *CommitInfo) Hash() []byte {
	// TODO: cache to ci.hash []byte
	m := make(map[string][]byte, len(ci.StoreInfos))
	for _, storeInfo := range ci.StoreInfos {
		m[storeInfo.Name] = storeInfo.Hash()
	}

	return merkle.SimpleHashFromMap(m)
}

func (ci *CommitInfo) CommitID() types.CommitID {
	return types.CommitID{
		Version: ci.Version,
		Hash:    ci.Hash(),
	}
}

//----------------------------------------
// StoreInfo

// StoreInfo contains the name and core reference for an
// underlying store.  It is the leaf of the Stores top
// level simple merkle tree.
//type StoreInfo struct {
//	Name string
//	Core StoreCore
//}
//
//type StoreCore struct {
//	// StoreType StoreType
//	CommitID types.CommitID
//	// ... maybe add more state
//}

// Implements merkle.Hasher.
func (si StoreInfo) Hash() []byte {
	// Doesn't write Name, since merkle.SimpleHashFromMap() will
	// include them via the keys.
	bz := si.Core.CommitID.Hash
	hasher := tmhash.New()

	_, err := hasher.Write(bz)
	if err != nil {
		// TODO: Handle with #870
		panic(err)
	}

	return hasher.Sum(nil)
}

//----------------------------------------
// Misc.

func getLatestVersion(db dbm.DB) int64 {
	var latest sdk.Int64
	latestBytes, _ := db.Get([]byte(latestVersionKey))
	if latestBytes == nil {
		return 0
	}
	err := cdc.LegacyUnmarshalBinaryLengthPrefixed(latestBytes, &latest)
	if err != nil {
		panic(err)
	}

	return int64(latest)
}

// Set the latest version.
func setLatestVersion(batch dbm.Batch, version int64) {
	v := sdk.Int64(version)
	latestBytes, _ := cdc.LegacyMarshalBinaryLengthPrefixed(&v)
	batch.Set([]byte(latestVersionKey), latestBytes)
}

// Commits each store and returns a new commitInfo.
func commitStores(version int64, storeMap map[types.StoreKey]types.CommitStore) CommitInfo {
	storeInfos := make([]StoreInfo, 0, len(storeMap))

	for key, store := range storeMap {
		// Commit
		commitID := store.Commit()

		if store.GetStoreType() == types.StoreTypeTransient || store.GetStoreType() == types.StoreTypeMemory {
			continue
		}

		// Record CommitID
		si := StoreInfo{}
		si.Name = key.Name()
		si.Core.CommitID = commitID
		// si.Core.StoreType = store.GetStoreType()
		storeInfos = append(storeInfos, si)
	}

	ci := CommitInfo{
		Version:    version,
		StoreInfos: storeInfos,
	}
	return ci
}

// Gets commitInfo from disk.
func getCommitInfo(db dbm.DB, ver int64) (CommitInfo, error) {

	// Get from DB.
	cInfoKey := fmt.Sprintf(commitInfoKeyFmt, ver)
	cInfoBytes, _ := db.Get([]byte(cInfoKey))
	if cInfoBytes == nil {
		return CommitInfo{}, fmt.Errorf("failed to get Store: no data")
	}

	var cInfo CommitInfo

	err := cdc.LegacyUnmarshalBinaryLengthPrefixed(cInfoBytes, &cInfo)
	if err != nil {
		return CommitInfo{}, fmt.Errorf("failed to get Store: %v", err)
	}

	return cInfo, nil
}

// Set a commitInfo for given version.
func setCommitInfo(batch dbm.Batch, version int64, cInfo CommitInfo) {
	cInfoBytes, err := cdc.LegacyMarshalBinaryLengthPrefixed(&cInfo)
	if err != nil {
		panic(err)
	}
	cInfoKey := fmt.Sprintf(commitInfoKeyFmt, version)
	batch.Set([]byte(cInfoKey), cInfoBytes)
}
