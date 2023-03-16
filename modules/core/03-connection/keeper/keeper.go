package keeper

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/vipernet-xyz/viper-network/codec"
	storetypes "github.com/vipernet-xyz/viper-network/store/types"
	paramtypes "github.com/vipernet-xyz/viper-network/types"
	sdk "github.com/vipernet-xyz/viper-network/types"

	clienttypes "github.com/vipernet-xyz/viper-network/modules/core/02-client/types"
	"github.com/vipernet-xyz/viper-network/modules/core/03-connection/types"
	commitmenttypes "github.com/vipernet-xyz/viper-network/modules/core/23-commitment/types"
	host "github.com/vipernet-xyz/viper-network/modules/core/24-host"
	"github.com/vipernet-xyz/viper-network/modules/core/exported"
)

// Keeper defines the IBC connection keeper
type Keeper struct {
	// implements gRPC QueryServer interface
	types.QueryServer

	storeKey     storetypes.StoreKey
	paramSpace   paramtypes.Subspace
	cdc          codec.BinaryCodec
	clientKeeper types.ClientKeeper
}

// NewKeeper creates a new IBC connection Keeper instance
func NewKeeper(cdc codec.BinaryCodec, key storetypes.StoreKey, paramSpace paramtypes.Subspace, ck types.ClientKeeper) Keeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeKey:     key,
		cdc:          cdc,
		paramSpace:   paramSpace,
		clientKeeper: ck,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Ctx) log.Logger {
	return ctx.Logger().With("module", "x/"+exported.ModuleName+"/"+types.SubModuleName)
}

// GetCommitmentPrefix returns the IBC connection store prefix as a commitment
// Prefix
func (k Keeper) GetCommitmentPrefix() exported.Prefix {
	return commitmenttypes.NewMerklePrefix([]byte(k.storeKey.Name()))
}

// GenerateConnectionIdentifier returns the next connection identifier.
func (k Keeper) GenerateConnectionIdentifier(ctx sdk.Context) string {
	nextConnSeq := k.GetNextConnectionSequence(ctx)
	connectionID := types.FormatConnectionIdentifier(nextConnSeq)

	nextConnSeq++
	k.SetNextConnectionSequence(ctx, nextConnSeq)
	return connectionID
}

// GetConnection returns a connection with a particular identifier
func (k Keeper) GetConnection(ctx sdk.Ctx, connectionID string) (types.ConnectionEnd, bool) {
	store := ctx.KVStore(k.storeKey)
	bz, _ := store.Get(host.ConnectionKey(connectionID))
	if len(bz) == 0 {
		return types.ConnectionEnd{}, false
	}

	var connection types.ConnectionEnd
	k.cdc.MustUnmarshal(bz, &connection)

	return connection, true
}

// HasConnection returns a true if the connection with the given identifier
// exists in the store.
func (k Keeper) HasConnection(ctx sdk.Context, connectionID string) bool {
	store := ctx.KVStore(k.storeKey)
	b, _ := store.Has(host.ConnectionKey(connectionID))
	return b
}

// SetConnection sets a connection to the store
func (k Keeper) SetConnection(ctx sdk.Ctx, connectionID string, connection types.ConnectionEnd) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&connection)
	store.Set(host.ConnectionKey(connectionID), bz)
}

// GetTimestampAtHeight returns the timestamp in nanoseconds of the consensus state at the
// given height.
func (k Keeper) GetTimestampAtHeight(ctx sdk.Ctx, connection types.ConnectionEnd, height exported.Height) (uint64, error) {
	clientState, found := k.clientKeeper.GetClientState(ctx, connection.GetClientID())
	if !found {
		return 0, errorsmod.Wrapf(
			clienttypes.ErrClientNotFound, "clientID (%s)", connection.GetClientID(),
		)
	}

	timestamp, err := clientState.GetTimestampAtHeight(ctx, k.clientKeeper.ClientStore(ctx, connection.GetClientID()), k.cdc, height)
	if err != nil {
		return 0, err
	}

	return timestamp, nil
}

// GetClientConnectionPaths returns all the connection paths stored under a
// particular client
func (k Keeper) GetClientConnectionPaths(ctx sdk.Ctx, clientID string) ([]string, bool) {
	store := ctx.KVStore(k.storeKey)
	bz, _ := store.Get(host.ClientConnectionsKey(clientID))
	if len(bz) == 0 {
		return nil, false
	}

	var clientPaths types.ClientPaths
	k.cdc.MustUnmarshal(bz, &clientPaths)
	return clientPaths.Paths, true
}

// SetClientConnectionPaths sets the connections paths for client
func (k Keeper) SetClientConnectionPaths(ctx sdk.Context, clientID string, paths []string) {
	store := ctx.KVStore(k.storeKey)
	clientPaths := types.ClientPaths{Paths: paths}
	bz := k.cdc.MustMarshal(&clientPaths)
	store.Set(host.ClientConnectionsKey(clientID), bz)
}

// GetNextConnectionSequence gets the next connection sequence from the store.
func (k Keeper) GetNextConnectionSequence(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz, _ := store.Get([]byte(types.KeyNextConnectionSequence))
	if len(bz) == 0 {
		panic("next connection sequence is nil")
	}

	return sdk.BigEndianToUint64(bz)
}

// SetNextConnectionSequence sets the next connection sequence to the store.
func (k Keeper) SetNextConnectionSequence(ctx sdk.Context, sequence uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := sdk.Uint64ToBigEndian(sequence)
	store.Set([]byte(types.KeyNextConnectionSequence), bz)
}

// GetAllClientConnectionPaths returns all stored clients connection id paths. It
// will ignore the clients that haven't initialized a connection handshake since
// no paths are stored.
func (k Keeper) GetAllClientConnectionPaths(ctx sdk.Ctx) []types.ConnectionPaths {
	var allConnectionPaths []types.ConnectionPaths
	k.clientKeeper.IterateClientStates(ctx, nil, func(clientID string, cs exported.ClientState) bool {
		paths, found := k.GetClientConnectionPaths(ctx, clientID)
		if !found {
			// continue when connection handshake is not initialized
			return false
		}
		connPaths := types.NewConnectionPaths(clientID, paths)
		allConnectionPaths = append(allConnectionPaths, connPaths)
		return false
	})

	return allConnectionPaths
}

// IterateConnections provides an iterator over all ConnectionEnd objects.
// For each ConnectionEnd, cb will be called. If the cb returns true, the
// iterator will close and stop.
func (k Keeper) IterateConnections(ctx sdk.Context, cb func(types.IdentifiedConnection) bool) {
	store := ctx.KVStore(k.storeKey)
	iterator, _ := sdk.KVStorePrefixIterator(store, []byte(host.KeyConnectionPrefix))

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var connection types.ConnectionEnd
		k.cdc.MustUnmarshal(iterator.Value(), &connection)

		connectionID := host.MustParseConnectionPath(string(iterator.Key()))
		identifiedConnection := types.NewIdentifiedConnection(connectionID, connection)
		if cb(identifiedConnection) {
			break
		}
	}
}

// GetAllConnections returns all stored ConnectionEnd objects.
func (k Keeper) GetAllConnections(ctx sdk.Context) (connections []types.IdentifiedConnection) {
	k.IterateConnections(ctx, func(connection types.IdentifiedConnection) bool {
		connections = append(connections, connection)
		return false
	})
	return connections
}

// addConnectionToClient is used to add a connection identifier to the set of
// connections associated with a client.
func (k Keeper) addConnectionToClient(ctx sdk.Context, clientID, connectionID string) error {
	_, found := k.clientKeeper.GetClientState(ctx, clientID)
	if !found {
		return errorsmod.Wrap(clienttypes.ErrClientNotFound, clientID)
	}

	conns, found := k.GetClientConnectionPaths(ctx, clientID)
	if !found {
		conns = []string{}
	}

	conns = append(conns, connectionID)
	k.SetClientConnectionPaths(ctx, clientID, conns)
	return nil
}
