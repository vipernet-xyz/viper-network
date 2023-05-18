package keeper

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"strings"

	errorsmod "cosmossdk.io/errors"
	"github.com/cometbft/cometbft/light"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/vipernet-xyz/viper-network/codec"
	ibcerrors "github.com/vipernet-xyz/viper-network/internal/errors"
	"github.com/vipernet-xyz/viper-network/modules/core/02-client/types"
	upgradetypes "github.com/vipernet-xyz/viper-network/modules/core/02-client/types"
	commitmenttypes "github.com/vipernet-xyz/viper-network/modules/core/23-commitment/types"
	host "github.com/vipernet-xyz/viper-network/modules/core/24-host"
	"github.com/vipernet-xyz/viper-network/modules/core/exported"
	ibctm "github.com/vipernet-xyz/viper-network/modules/light-clients/07-tendermint"
	"github.com/vipernet-xyz/viper-network/store/prefix"
	storetypes "github.com/vipernet-xyz/viper-network/store/types"
	paramtypes "github.com/vipernet-xyz/viper-network/types"
	sdk "github.com/vipernet-xyz/viper-network/types"
	sdkerrors "github.com/vipernet-xyz/viper-network/types/errors"
	"github.com/vipernet-xyz/viper-network/types/kv"
)

// Keeper represents a type that grants read and write permissions to any client
// state information
type Keeper struct {
	storeKey      storetypes.StoreKey
	cdc           codec.BinaryCodec
	paramSpace    paramtypes.Subspace
	stakingKeeper types.StakingKeeper
	upgradeKeeper types.UpgradeKeeper
}

// NewKeeper creates a new NewKeeper instance
func NewKeeper(cdc codec.BinaryCodec, key storetypes.StoreKey, paramSpace paramtypes.Subspace, uk types.UpgradeKeeper) Keeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeKey:      key,
		cdc:           cdc,
		paramSpace:    paramSpace,
		upgradeKeeper: uk,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Ctx) log.Logger {
	return ctx.Logger().With("module", "x/"+exported.ModuleName+"/"+types.SubModuleName)
}

// GenerateClientIdentifier returns the next client identifier.
func (k Keeper) GenerateClientIdentifier(ctx sdk.Ctx, clientType string) string {
	nextClientSeq := k.GetNextClientSequence(ctx)
	clientID := types.FormatClientIdentifier(clientType, nextClientSeq)

	nextClientSeq++
	k.SetNextClientSequence(ctx, nextClientSeq)
	return clientID
}

// GetClientState gets a particular client from the store
func (k Keeper) GetClientState(ctx sdk.Ctx, clientID string) (exported.ClientState, bool) {
	store := k.ClientStore(ctx, clientID)
	bz, _ := store.Get(host.ClientStateKey())
	if len(bz) == 0 {
		return nil, false
	}

	clientState := k.MustUnmarshalClientState(bz)
	return clientState, true
}

// SetClientState sets a particular Client to the store
func (k Keeper) SetClientState(ctx sdk.Ctx, clientID string, clientState exported.ClientState) {
	store := k.ClientStore(ctx, clientID)
	store.Set(host.ClientStateKey(), k.MustMarshalClientState(clientState))
}

// GetClientConsensusState gets the stored consensus state from a client at a given height.
func (k Keeper) GetClientConsensusState(ctx sdk.Ctx, clientID string, height exported.Height) (exported.ConsensusState, bool) {
	store := k.ClientStore(ctx, clientID)
	bz, _ := store.Get(host.ConsensusStateKey(height))
	if len(bz) == 0 {
		return nil, false
	}

	consensusState := k.MustUnmarshalConsensusState(bz)
	return consensusState, true
}

// SetClientConsensusState sets a ConsensusState to a particular client at the given
// height
func (k Keeper) SetClientConsensusState(ctx sdk.Ctx, clientID string, height exported.Height, consensusState exported.ConsensusState) {
	store := k.ClientStore(ctx, clientID)
	store.Set(host.ConsensusStateKey(height), k.MustMarshalConsensusState(consensusState))
}

// GetNextClientSequence gets the next client sequence from the store.
func (k Keeper) GetNextClientSequence(ctx sdk.Ctx) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz, _ := store.Get([]byte(types.KeyNextClientSequence))
	if len(bz) == 0 {
		panic("next client sequence is nil")
	}

	return sdk.BigEndianToUint64(bz)
}

// SetNextClientSequence sets the next client sequence to the store.
func (k Keeper) SetNextClientSequence(ctx sdk.Ctx, sequence uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := sdk.Uint64ToBigEndian(sequence)
	store.Set([]byte(types.KeyNextClientSequence), bz)
}

// IterateConsensusStates provides an iterator over all stored consensus states.
// objects. For each State object, cb will be called. If the cb returns true,
// the iterator will close and stop.
func (k Keeper) IterateConsensusStates(ctx sdk.Ctx, cb func(clientID string, cs types.ConsensusStateWithHeight) bool) {
	store := ctx.KVStore(k.storeKey)
	iterator, _ := sdk.KVStorePrefixIterator(store, host.KeyClientStorePrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		keySplit := strings.Split(string(iterator.Key()), "/")
		// consensus key is in the format "clients/<clientID>/consensusStates/<height>"
		if len(keySplit) != 4 || keySplit[2] != string(host.KeyConsensusStatePrefix) {
			continue
		}
		clientID := keySplit[1]
		height := types.MustParseHeight(keySplit[3])
		consensusState := k.MustUnmarshalConsensusState(iterator.Value())

		consensusStateWithHeight := types.NewConsensusStateWithHeight(height, consensusState)

		if cb(clientID, consensusStateWithHeight) {
			break
		}
	}
}

// GetAllGenesisClients returns all the clients in state with their client ids returned as IdentifiedClientState
func (k Keeper) GetAllGenesisClients(ctx sdk.Ctx) types.IdentifiedClientStates {
	var genClients types.IdentifiedClientStates
	k.IterateClientStates(ctx, nil, func(clientID string, cs exported.ClientState) bool {
		genClients = append(genClients, types.NewIdentifiedClientState(clientID, cs))
		return false
	})

	return genClients.Sort()
}

// GetAllClientMetadata will take a list of IdentifiedClientState and return a list
// of IdentifiedGenesisMetadata necessary for exporting and importing client metadata
// into the client store.
func (k Keeper) GetAllClientMetadata(ctx sdk.Ctx, genClients []types.IdentifiedClientState) ([]types.IdentifiedGenesisMetadata, error) {
	genMetadata := make([]types.IdentifiedGenesisMetadata, 0)
	for _, ic := range genClients {
		cs, err := types.UnpackClientState(ic.ClientState)
		if err != nil {
			return nil, err
		}
		gms := cs.ExportMetadata(k.ClientStore(ctx, ic.ClientId))
		if len(gms) == 0 {
			continue
		}
		clientMetadata := make([]types.GenesisMetadata, len(gms))
		for i, metadata := range gms {
			cmd, ok := metadata.(types.GenesisMetadata)
			if !ok {
				return nil, errorsmod.Wrapf(types.ErrInvalidClientMetadata, "expected metadata type: %T, got: %T",
					types.GenesisMetadata{}, cmd)
			}
			clientMetadata[i] = cmd
		}
		genMetadata = append(genMetadata, types.NewIdentifiedGenesisMetadata(
			ic.ClientId,
			clientMetadata,
		))
	}
	return genMetadata, nil
}

// SetAllClientMetadata takes a list of IdentifiedGenesisMetadata and stores all of the metadata in the client store at the appropriate paths.
func (k Keeper) SetAllClientMetadata(ctx sdk.Ctx, genMetadata []types.IdentifiedGenesisMetadata) {
	for _, igm := range genMetadata {
		// create client store
		store := k.ClientStore(ctx, igm.ClientId)
		// set all metadata kv pairs in client store
		for _, md := range igm.ClientMetadata {
			store.Set(md.GetKey(), md.GetValue())
		}
	}
}

// GetAllConsensusStates returns all stored client consensus states.
func (k Keeper) GetAllConsensusStates(ctx sdk.Ctx) types.ClientsConsensusStates {
	clientConsStates := make(types.ClientsConsensusStates, 0)
	mapClientIDToConsStateIdx := make(map[string]int)

	k.IterateConsensusStates(ctx, func(clientID string, cs types.ConsensusStateWithHeight) bool {
		idx, ok := mapClientIDToConsStateIdx[clientID]
		if ok {
			clientConsStates[idx].ConsensusStates = append(clientConsStates[idx].ConsensusStates, cs)
			return false
		}

		clientConsState := types.ClientConsensusStates{
			ClientId:        clientID,
			ConsensusStates: []types.ConsensusStateWithHeight{cs},
		}

		clientConsStates = append(clientConsStates, clientConsState)
		mapClientIDToConsStateIdx[clientID] = len(clientConsStates) - 1
		return false
	})

	return clientConsStates.Sort()
}

// HasClientConsensusState returns if keeper has a ConsensusState for a particular
// client at the given height
func (k Keeper) HasClientConsensusState(ctx sdk.Ctx, clientID string, height exported.Height) bool {
	store := k.ClientStore(ctx, clientID)
	b, _ := store.Has(host.ConsensusStateKey(height))
	return b
}

// GetLatestClientConsensusState gets the latest ConsensusState stored for a given client
func (k Keeper) GetLatestClientConsensusState(ctx sdk.Context, clientID string) (exported.ConsensusState, bool) {
	clientState, ok := k.GetClientState(ctx, clientID)
	if !ok {
		return nil, false
	}
	return k.GetClientConsensusState(ctx, clientID, clientState.GetLatestHeight())
}

// GetSelfConsensusState introspects the (self) past historical info at a given height
// and returns the expected consensus state at that height.
// For now, can only retrieve self consensus states for the current revision
func (k Keeper) GetSelfConsensusState(ctx sdk.Ctx, height exported.Height) (exported.ConsensusState, error) {
	selfHeight, ok := height.(types.Height)
	if !ok {
		return nil, errorsmod.Wrapf(ibcerrors.ErrInvalidType, "expected %T, got %T", types.Height{}, height)
	}
	// check that height revision matches chainID revision
	revision := types.ParseChainID(ctx.ChainID())
	if revision != height.GetRevisionNumber() {
		return nil, errorsmod.Wrapf(types.ErrInvalidHeight, "chainID revision number does not match height revision number: expected %d, got %d", revision, height.GetRevisionNumber())
	}
	histInfo, found := k.stakingKeeper.GetHistoricalInfo(ctx, int64(selfHeight.RevisionHeight))
	if !found {
		return nil, errorsmod.Wrapf(ibcerrors.ErrNotFound, "no historical info found at height %d", selfHeight.RevisionHeight)
	}

	consensusState := &ibctm.ConsensusState{
		Timestamp:          histInfo.Header.Time,
		Root:               commitmenttypes.NewMerkleRoot(histInfo.Header.GetAppHash()),
		NextValidatorsHash: histInfo.Header.NextValidatorsHash,
	}
	return consensusState, nil
}

// ValidateSelfClient validates the client parameters for a client of the running chain
// This function is only used to validate the client state the counterparty stores for this chain
// Client must be in same revision as the executing chain
func (k Keeper) ValidateSelfClient(ctx sdk.Ctx, clientState exported.ClientState) error {
	tmClient, ok := clientState.(*ibctm.ClientState)
	if !ok {
		return errorsmod.Wrapf(types.ErrInvalidClient, "client must be a Tendermint client, expected: %T, got: %T",
			&ibctm.ClientState{}, tmClient)
	}

	if !tmClient.FrozenHeight.IsZero() {
		return types.ErrClientFrozen
	}

	if ctx.ChainID() != tmClient.ChainId {
		return errorsmod.Wrapf(types.ErrInvalidClient, "invalid chain-id. expected: %s, got: %s",
			ctx.ChainID(), tmClient.ChainId)
	}

	revision := types.ParseChainID(ctx.ChainID())

	// client must be in the same revision as executing chain
	if tmClient.LatestHeight.RevisionNumber != revision {
		return errorsmod.Wrapf(types.ErrInvalidClient, "client is not in the same revision as the chain. expected revision: %d, got: %d",
			tmClient.LatestHeight.RevisionNumber, revision)
	}

	selfHeight := types.NewHeight(revision, uint64(ctx.BlockHeight()))
	if tmClient.LatestHeight.GTE(selfHeight) {
		return errorsmod.Wrapf(types.ErrInvalidClient, "client has LatestHeight %d greater than or equal to chain height %d",
			tmClient.LatestHeight, selfHeight)
	}

	expectedProofSpecs := commitmenttypes.GetSDKSpecs()
	if !reflect.DeepEqual(expectedProofSpecs, tmClient.ProofSpecs) {
		return errorsmod.Wrapf(types.ErrInvalidClient, "client has invalid proof specs. expected: %v got: %v",
			expectedProofSpecs, tmClient.ProofSpecs)
	}

	if err := light.ValidateTrustLevel(tmClient.TrustLevel.ToTendermint()); err != nil {
		return errorsmod.Wrapf(types.ErrInvalidClient, "trust-level invalid: %v", err)
	}

	expectedUbdPeriod := k.stakingKeeper.UnbondingTime(ctx)
	if expectedUbdPeriod != tmClient.UnbondingPeriod {
		return errorsmod.Wrapf(types.ErrInvalidClient, "invalid unbonding period. expected: %s, got: %s",
			expectedUbdPeriod, tmClient.UnbondingPeriod)
	}

	if tmClient.UnbondingPeriod < tmClient.TrustingPeriod {
		return errorsmod.Wrapf(types.ErrInvalidClient, "unbonding period must be greater than trusting period. unbonding period (%d) < trusting period (%d)",
			tmClient.UnbondingPeriod, tmClient.TrustingPeriod)
	}

	if len(tmClient.UpgradePath) != 0 {
		// For now, SDK IBC implementation assumes that upgrade path (if defined) is defined by SDK upgrade module
		expectedUpgradePath := []string{"upgrade", upgradetypes.KeyUpgradedIBCState}
		if !reflect.DeepEqual(expectedUpgradePath, tmClient.UpgradePath) {
			return errorsmod.Wrapf(types.ErrInvalidClient, "upgrade path must be the upgrade path defined by upgrade module. expected %v, got %v",
				expectedUpgradePath, tmClient.UpgradePath)
		}
	}
	return nil
}

// GetUpgradePlan executes the upgrade keeper GetUpgradePlan function.
func (k Keeper) GetUpgradePlan(ctx sdk.Ctx) (plan upgradetypes.Plan, havePlan bool) {
	store := ctx.KVStore(k.storeKey)
	bz, _ := store.Get(types.PlanKey())
	if bz == nil {
		return plan, false
	}

	k.cdc.MustUnmarshal(bz, &plan)
	return plan, true
}

// GetUpgradedClient executes the upgrade keeper GetUpgradeClient function.
func (k Keeper) GetUpgradedClient(ctx sdk.Ctx, planHeight int64) ([]byte, bool) {
	return k.upgradeKeeper.GetUpgradedClient(ctx, planHeight)
}

// GetUpgradedConsensusState returns the upgraded consensus state
func (k Keeper) GetUpgradedConsensusState(ctx sdk.Ctx, planHeight int64) ([]byte, bool) {
	store := ctx.KVStore(k.storeKey)
	bz, _ := store.Get(types.UpgradedConsStateKey(planHeight))
	if len(bz) == 0 {
		return nil, false
	}

	return bz, true
}

// SetUpgradedConsensusState executes the upgrade keeper SetUpgradedConsensusState function.
func (k Keeper) SetUpgradedConsensusState(ctx sdk.Ctx, planHeight int64, bz []byte) error {
	return k.upgradeKeeper.SetUpgradedConsensusState(ctx, planHeight, bz)
}

// IterateClientStates provides an iterator over all stored light client State
// objects. For each State object, cb will be called. If the cb returns true,
// the iterator will close and stop.
func (k Keeper) IterateClientStates(ctx sdk.Ctx, prefix []byte, cb func(clientID string, cs exported.ClientState) bool) {
	store := ctx.KVStore(k.storeKey)
	iterator, _ := sdk.KVStorePrefixIterator(store, host.PrefixedClientStoreKey(prefix))

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		path := string(iterator.Key())
		if !strings.Contains(path, host.KeyClientState) {
			// skip non client state keys
			continue
		}

		clientID := host.MustParseClientStatePath(path)
		clientState := k.MustUnmarshalClientState(iterator.Value())

		if cb(clientID, clientState) {
			break
		}
	}
}

// GetAllClients returns all stored light client State objects.
func (k Keeper) GetAllClients(ctx sdk.Context) []exported.ClientState {
	var states []exported.ClientState
	k.IterateClientStates(ctx, nil, func(_ string, state exported.ClientState) bool {
		states = append(states, state)
		return false
	})

	return states
}

// ClientStore returns isolated prefix store for each client so they can read/write in separate
// namespace without being able to read/write other client's data
func (k Keeper) ClientStore(ctx sdk.Ctx, clientID string) sdk.KVStore {
	clientPrefix := []byte(fmt.Sprintf("%s/%s/", host.KeyClientStorePrefix, clientID))
	return prefix.NewStore(ctx.KVStore(k.storeKey), clientPrefix)
}

// ScheduleUpgrade schedules an upgrade based on the specified plan.
// If there is another Plan already scheduled, it will cancel and overwrite it.
// ScheduleUpgrade will also write the upgraded IBC ClientState to the upgraded client
// path if it is specified in the plan.
func (k Keeper) ScheduleUpgrade(ctx sdk.Ctx, plan upgradetypes.Plan) error {
	if err := plan.ValidateBasic(); err != nil {
		return err
	}

	// NOTE: allow for the possibility of chains to schedule upgrades in begin block of the same block
	// as a strategy for emergency hard fork recoveries
	if plan.Height < ctx.BlockHeight() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "upgrade cannot be scheduled in the past")
	}

	if k.GetDoneHeight(ctx, plan.Name) != 0 {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "upgrade with name %s has already been completed", plan.Name)
	}

	store := ctx.KVStore(k.storeKey)

	// clear any old IBC state stored by previous plan
	oldPlan, found := k.GetUpgradePlan(ctx)
	if found {
		k.ClearIBCState(ctx, oldPlan.Height)
	}

	bz := k.cdc.MustMarshal(&plan)
	store.Set(types.PlanKey(), bz)

	return nil
}

// GetDoneHeight returns the height at which the given upgrade was executed
func (k Keeper) GetDoneHeight(ctx sdk.Ctx, name string) int64 {
	iter, _ := storetypes.KVStorePrefixIterator(ctx.KVStore(k.storeKey), []byte{types.DoneByte})
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		upgradeName, height := parseDoneKey(iter.Key())
		if upgradeName == name {
			return height
		}
	}
	return 0
}

// parseDoneKey - split upgrade name and height from the done key
func parseDoneKey(key []byte) (string, int64) {
	// 1 byte for the DoneByte + 8 bytes height + at least 1 byte for the name
	kv.AssertKeyAtLeastLength(key, 10)
	height := binary.BigEndian.Uint64(key[1:9])
	return string(key[9:]), int64(height)
}

// ClearIBCState clears any planned IBC state
func (k Keeper) ClearIBCState(ctx sdk.Ctx, lastHeight int64) {
	// delete IBC client and consensus state from store if this is IBC plan
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.UpgradedClientKey(lastHeight))
	store.Delete(types.UpgradedConsStateKey(lastHeight))
}

// SetUpgradedClient sets the expected upgraded client for the next version of this chain at the last height the current chain will commit.
func (k Keeper) SetUpgradedClient(ctx sdk.Ctx, planHeight int64, bz []byte) error {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.UpgradedClientKey(planHeight), bz)
	return nil
}
