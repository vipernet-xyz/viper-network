package keeper

import (
	"fmt"
	"strings"

	"github.com/tendermint/tendermint/libs/log"
	"github.com/vipernet-xyz/viper-network/codec"
	storetypes "github.com/vipernet-xyz/viper-network/store/types"
	paramtypes "github.com/vipernet-xyz/viper-network/types"
	sdk "github.com/vipernet-xyz/viper-network/types"
	capabilitytypes "github.com/vipernet-xyz/viper-network/x/capability/types"

	genesistypes "github.com/vipernet-xyz/viper-network/modules/apps/27-interchain-accounts/genesis/types"
	"github.com/vipernet-xyz/viper-network/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/vipernet-xyz/viper-network/modules/apps/27-interchain-accounts/types"
	channeltypes "github.com/vipernet-xyz/viper-network/modules/core/04-channel/types"
	porttypes "github.com/vipernet-xyz/viper-network/modules/core/05-port/types"
	host "github.com/vipernet-xyz/viper-network/modules/core/24-host"
	"github.com/vipernet-xyz/viper-network/modules/core/exported"
)

// Keeper defines the IBC interchain accounts host keeper
type Keeper struct {
	storeKey   storetypes.StoreKey
	cdc        *codec.Codec
	paramSpace paramtypes.Subspace

	ics4Wrapper   porttypes.ICS4Wrapper
	channelKeeper icatypes.ChannelKeeper
	portKeeper    icatypes.PortKeeper
	accountKeeper icatypes.AccountKeeper

	scopedKeeper exported.ScopedKeeper

	msgRouter icatypes.MessageRouter
}

// NewKeeper creates a new interchain accounts host Keeper instance
func NewKeeper(
	cdc *codec.Codec, key storetypes.StoreKey, paramSpace paramtypes.Subspace,
	ics4Wrapper porttypes.ICS4Wrapper, channelKeeper icatypes.ChannelKeeper, portKeeper icatypes.PortKeeper,
	accountKeeper icatypes.AccountKeeper, scopedKeeper exported.ScopedKeeper, msgRouter icatypes.MessageRouter,
) Keeper {
	// ensure ibc interchain accounts module account is set
	if addr := accountKeeper.GetModuleAddress(icatypes.ModuleName); addr == nil {
		panic("the Interchain Accounts module account has not been set")
	}

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
		accountKeeper: accountKeeper,
		scopedKeeper:  scopedKeeper,
		msgRouter:     msgRouter,
	}
}

// Logger returns the application logger, scoped to the associated module
func (k Keeper) Logger(ctx sdk.Ctx) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s-%s", exported.ModuleName, icatypes.ModuleName))
}

// BindPort stores the provided portID and binds to it, returning the associated capability
func (k Keeper) BindPort(ctx sdk.Ctx, portID string) *capabilitytypes.Capability {
	store := ctx.KVStore(k.storeKey)
	store.Set(icatypes.KeyPort(portID), []byte{0x01})

	return k.portKeeper.BindPort(ctx, portID)
}

// IsBound checks if the interchain account host module is already bound to the desired port
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

// GetActiveChannelID retrieves the active channelID from the store keyed by the provided connectionID and portID
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

// GetAllActiveChannels returns a list of all active interchain accounts host channels and their associated connection and port identifiers
func (k Keeper) GetAllActiveChannels(ctx sdk.Ctx) []genesistypes.ActiveChannel {
	store := ctx.KVStore(k.storeKey)
	iterator, _ := sdk.KVStorePrefixIterator(store, []byte(icatypes.ActiveChannelKeyPrefix))
	defer iterator.Close()

	var activeChannels []genesistypes.ActiveChannel
	for ; iterator.Valid(); iterator.Next() {
		keySplit := strings.Split(string(iterator.Key()), "/")

		ch := genesistypes.ActiveChannel{
			ConnectionId: keySplit[2],
			PortId:       keySplit[1],
			ChannelId:    string(iterator.Value()),
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
