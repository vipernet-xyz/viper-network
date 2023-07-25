package keeper

import (
	"bytes"
	"fmt"
	"strings"

	errorsmod "cosmossdk.io/errors"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/vipernet-xyz/viper-network/codec"
	storetypes "github.com/vipernet-xyz/viper-network/store/types"
	paramtypes "github.com/vipernet-xyz/viper-network/types"
	sdk "github.com/vipernet-xyz/viper-network/types"
	capabilitytypes "github.com/vipernet-xyz/viper-network/x/capability/types"

	"github.com/vipernet-xyz/viper-network/modules/apps/27-interchain-accounts/controller/types"
	genesistypes "github.com/vipernet-xyz/viper-network/modules/apps/27-interchain-accounts/genesis/types"
	icatypes "github.com/vipernet-xyz/viper-network/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/vipernet-xyz/viper-network/modules/core/04-channel/types"
	porttypes "github.com/vipernet-xyz/viper-network/modules/core/05-port/types"
	host "github.com/vipernet-xyz/viper-network/modules/core/24-host"
	"github.com/vipernet-xyz/viper-network/modules/core/exported"
)

// Keeper defines the IBC interchain accounts controller keeper
type Keeper struct {
	storeKey   storetypes.StoreKey
	cdc        codec.BinaryCodec
	paramSpace paramtypes.Subspace

	ics4Wrapper   porttypes.ICS4Wrapper
	channelKeeper icatypes.ChannelKeeper
	portKeeper    icatypes.PortKeeper

	scopedKeeper exported.ScopedKeeper

	msgRouter icatypes.MessageRouter
}

// NewKeeper creates a new interchain accounts controller Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec, key storetypes.StoreKey, paramSpace paramtypes.Subspace,
	ics4Wrapper porttypes.ICS4Wrapper, channelKeeper icatypes.ChannelKeeper, portKeeper icatypes.PortKeeper,
	scopedKeeper exported.ScopedKeeper, msgRouter icatypes.MessageRouter,
) Keeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeKey:      key,
		cdc:           cdc,
		paramSpace:    paramSpace,
		ics4Wrapper:   ics4Wrapper,
		channelKeeper: channelKeeper,
		portKeeper:    portKeeper,
		scopedKeeper:  scopedKeeper,
		msgRouter:     msgRouter,
	}
}

// Logger returns the application logger, scoped to the associated module
func (k Keeper) Logger(ctx sdk.Ctx) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s-%s", exported.ModuleName, icatypes.ModuleName))
}

// GetConnectionID returns the connection id for the given port and channelIDs.
func (k Keeper) GetConnectionID(ctx sdk.Ctx, portID, channelID string) (string, error) {
	channel, found := k.channelKeeper.GetChannel(ctx, portID, channelID)
	if !found {
		return "", errorsmod.Wrapf(channeltypes.ErrChannelNotFound, "port ID (%s) channel ID (%s)", portID, channelID)
	}
	return channel.ConnectionHops[0], nil
}

// GetAllPorts returns all ports to which the interchain accounts controller module is bound. Used in ExportGenesis
func (k Keeper) GetAllPorts(ctx sdk.Ctx) []string {
	store := ctx.KVStore(k.storeKey)
	iterator, _ := sdk.KVStorePrefixIterator(store, []byte(icatypes.PortKeyPrefix))
	defer iterator.Close()

	var ports []string
	for ; iterator.Valid(); iterator.Next() {
		keySplit := strings.Split(string(iterator.Key()), "/")

		ports = append(ports, keySplit[1])
	}

	return ports
}

// BindPort stores the provided portID and binds to it, returning the associated capability
func (k Keeper) BindPort(ctx sdk.Ctx, portID string) *capabilitytypes.Capability {
	store := ctx.KVStore(k.storeKey)
	store.Set(icatypes.KeyPort(portID), []byte{0x01})

	return k.portKeeper.BindPort(ctx, portID)
}

// IsBound checks if the interchain account controller module is already bound to the desired port
func (k Keeper) IsBound(ctx sdk.Ctx, portID string) bool {
	_, ok := k.scopedKeeper.GetCapability(ctx, host.PortPath(portID))
	return ok
}

// AuthenticateCapability wraps the scopedKeeper's AuthenticateCapability function
func (k Keeper) AuthenticateCapability(ctx sdk.Ctx, cap *capabilitytypes.Capability, name string) bool {
	return k.scopedKeeper.AuthenticateCapability(ctx, cap, name)
}

// ClaimCapability wraps the scopedKeeper's ClaimCapability function
func (k Keeper) ClaimCapability(ctx sdk.Ctx, cap *capabilitytypes.Capability, name string) error {
	return k.scopedKeeper.ClaimCapability(ctx, cap, name)
}

// GetAppVersion calls the ICS4Wrapper GetAppVersion function.
func (k Keeper) GetAppVersion(ctx sdk.Ctx, portID, channelID string) (string, bool) {
	return k.ics4Wrapper.GetAppVersion(ctx, portID, channelID)
}

// GetActiveChannelID retrieves the active channelID from the store, keyed by the provided connectionID and portID
func (k Keeper) GetActiveChannelID(ctx sdk.Ctx, connectionID, portID string) (string, bool) {
	store := ctx.KVStore(k.storeKey)
	key := icatypes.KeyActiveChannel(portID, connectionID)
	a, _ := store.Has(key)
	if !a {
		return "", false
	}
	b, _ := store.Get(key)
	return string(b), true
}

// GetOpenActiveChannel retrieves the active channelID from the store, keyed by the provided connectionID and portID & checks if the channel in question is in state OPEN
func (k Keeper) GetOpenActiveChannel(ctx sdk.Ctx, connectionID, portID string) (string, bool) {
	channelID, found := k.GetActiveChannelID(ctx, connectionID, portID)
	if !found {
		return "", false
	}

	channel, found := k.channelKeeper.GetChannel(ctx, portID, channelID)

	if found && channel.State == channeltypes.OPEN {
		return channelID, true
	}

	return "", false
}

// IsActiveChannelClosed retrieves the active channel from the store and returns true if the channel state is CLOSED, otherwise false
func (k Keeper) IsActiveChannelClosed(ctx sdk.Ctx, connectionID, portID string) bool {
	channelID, found := k.GetActiveChannelID(ctx, connectionID, portID)
	if !found {
		return false
	}

	channel, found := k.channelKeeper.GetChannel(ctx, portID, channelID)
	return found && channel.State == channeltypes.CLOSED
}

// GetAllActiveChannels returns a list of all active interchain accounts controller channels and their associated connection and port identifiers
func (k Keeper) GetAllActiveChannels(ctx sdk.Ctx) []genesistypes.ActiveChannel {
	store := ctx.KVStore(k.storeKey)
	iterator, _ := sdk.KVStorePrefixIterator(store, []byte(icatypes.ActiveChannelKeyPrefix))
	defer iterator.Close()

	var activeChannels []genesistypes.ActiveChannel
	for ; iterator.Valid(); iterator.Next() {
		keySplit := strings.Split(string(iterator.Key()), "/")

		portID := keySplit[1]
		connectionID := keySplit[2]
		channelID := string(iterator.Value())

		ch := genesistypes.ActiveChannel{
			ConnectionId:        connectionID,
			PortId:              portID,
			ChannelId:           channelID,
			IsMiddlewareEnabled: k.IsMiddlewareEnabled(ctx, portID, connectionID),
		}

		activeChannels = append(activeChannels, ch)
	}

	return activeChannels
}

// SetActiveChannelID stores the active channelID, keyed by the provided connectionID and portID
func (k Keeper) SetActiveChannelID(ctx sdk.Ctx, connectionID, portID, channelID string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(icatypes.KeyActiveChannel(portID, connectionID), []byte(channelID))
}

// IsActiveChannel returns true if there exists an active channel for the provided connectionID and portID, otherwise false
func (k Keeper) IsActiveChannel(ctx sdk.Ctx, connectionID, portID string) bool {
	_, ok := k.GetActiveChannelID(ctx, connectionID, portID)
	return ok
}

// GetInterchainAccountAddress retrieves the InterchainAccount address from the store associated with the provided connectionID and portID
func (k Keeper) GetInterchainAccountAddress(ctx sdk.Ctx, connectionID, portID string) (string, bool) {
	store := ctx.KVStore(k.storeKey)
	key := icatypes.KeyOwnerAccount(portID, connectionID)
	a, _ := store.Has(key)
	if !a {
		return "", false
	}
	b, _ := store.Get(key)
	return string(b), true
}

// GetAllInterchainAccounts returns a list of all registered interchain account addresses and their associated connection and controller port identifiers
func (k Keeper) GetAllInterchainAccounts(ctx sdk.Ctx) []genesistypes.RegisteredInterchainAccount {
	store := ctx.KVStore(k.storeKey)
	iterator, _ := sdk.KVStorePrefixIterator(store, []byte(icatypes.OwnerKeyPrefix))

	var interchainAccounts []genesistypes.RegisteredInterchainAccount
	for ; iterator.Valid(); iterator.Next() {
		keySplit := strings.Split(string(iterator.Key()), "/")

		acc := genesistypes.RegisteredInterchainAccount{
			ConnectionId:   keySplit[2],
			PortId:         keySplit[1],
			AccountAddress: string(iterator.Value()),
		}

		interchainAccounts = append(interchainAccounts, acc)
	}

	return interchainAccounts
}

// SetInterchainAccountAddress stores the InterchainAccount address, keyed by the associated connectionID and portID
func (k Keeper) SetInterchainAccountAddress(ctx sdk.Ctx, connectionID, portID, address string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(icatypes.KeyOwnerAccount(portID, connectionID), []byte(address))
}

// IsMiddlewareEnabled returns true if the underlying application callbacks are enabled for given port and connection identifier pair, otherwise false
func (k Keeper) IsMiddlewareEnabled(ctx sdk.Ctx, portID, connectionID string) bool {
	store := ctx.KVStore(k.storeKey)
	a, _ := store.Get(icatypes.KeyIsMiddlewareEnabled(portID, connectionID))
	return bytes.Equal(icatypes.MiddlewareEnabled, a)
}

// IsMiddlewareDisabled returns true if the underlying application callbacks are disabled for the given port and connection identifier pair, otherwise false
func (k Keeper) IsMiddlewareDisabled(ctx sdk.Ctx, portID, connectionID string) bool {
	store := ctx.KVStore(k.storeKey)
	a, _ := store.Get(icatypes.KeyIsMiddlewareEnabled(portID, connectionID))
	return bytes.Equal(icatypes.MiddlewareDisabled, a)
}

// SetMiddlewareEnabled stores a flag to indicate that the underlying application callbacks should be enabled for the given port and connection identifier pair
func (k Keeper) SetMiddlewareEnabled(ctx sdk.Ctx, portID, connectionID string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(icatypes.KeyIsMiddlewareEnabled(portID, connectionID), icatypes.MiddlewareEnabled)
}

// SetMiddlewareDisabled stores a flag to indicate that the underlying application callbacks should be disabled for the given port and connection identifier pair
func (k Keeper) SetMiddlewareDisabled(ctx sdk.Ctx, portID, connectionID string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(icatypes.KeyIsMiddlewareEnabled(portID, connectionID), icatypes.MiddlewareDisabled)
}

// DeleteMiddlewareEnabled deletes the middleware enabled flag stored in state
func (k Keeper) DeleteMiddlewareEnabled(ctx sdk.Ctx, portID, connectionID string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(icatypes.KeyIsMiddlewareEnabled(portID, connectionID))
}

func (k Keeper) UpgradeCodec(ctx sdk.Ctx) {
	if ctx.IsOnUpgradeHeight() {
		k.ConvertState(ctx)
	}
}

func (k Keeper) ConvertState(ctx sdk.Ctx) {
}
