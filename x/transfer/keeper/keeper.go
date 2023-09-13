package keeper

import (
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/vipernet-xyz/viper-network/codec"
	"github.com/vipernet-xyz/viper-network/store/prefix"
	storetypes "github.com/vipernet-xyz/viper-network/store/types"
	paramtypes "github.com/vipernet-xyz/viper-network/types"
	sdk "github.com/vipernet-xyz/viper-network/types"
	capabilitytypes "github.com/vipernet-xyz/viper-network/x/capability/types"

	abci "github.com/tendermint/tendermint/abci/types"
	porttypes "github.com/vipernet-xyz/viper-network/modules/core/05-port/types"
	host "github.com/vipernet-xyz/viper-network/modules/core/24-host"
	"github.com/vipernet-xyz/viper-network/x/transfer/exported"
	"github.com/vipernet-xyz/viper-network/x/transfer/types"
)

// Keeper defines the IBC fungible transfer keeper
type Keeper struct {
	storeKey   sdk.StoreKey
	cdc        *codec.Codec
	Cdc        codec.BinaryCodec
	paramSpace paramtypes.Subspace

	ics4Wrapper   porttypes.ICS4Wrapper
	channelKeeper types.ChannelKeeper
	portKeeper    types.PortKeeper
	authKeeper    types.AccountKeeper
	bankKeeper    types.BankKeeper
	scopedKeeper  exported.ScopedKeeper
}

// NewKeeper creates a new IBC transfer Keeper instance
func NewKeeper(
	cdc *codec.Codec, key storetypes.StoreKey, paramSpace paramtypes.Subspace,
	ics4Wrapper porttypes.ICS4Wrapper, channelKeeper types.ChannelKeeper, portKeeper types.PortKeeper,
	authKeeper types.AccountKeeper, bankKeeper types.BankKeeper, scopedKeeper exported.ScopedKeeper,
) Keeper {
	// ensure ibc transfer module account is set
	if addr := authKeeper.GetModuleAddress(types.ModuleName); addr == nil {
		panic("the IBC transfer module account has not been set")
	}

	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:           cdc,
		storeKey:      key,
		paramSpace:    paramSpace,
		ics4Wrapper:   ics4Wrapper,
		channelKeeper: channelKeeper,
		portKeeper:    portKeeper,
		authKeeper:    authKeeper,
		bankKeeper:    bankKeeper,
		scopedKeeper:  scopedKeeper,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Ctx) log.Logger {
	return ctx.Logger().With("module", "x/"+exported.ModuleName+"-"+types.ModuleName)
}

// IsBound checks if the transfer module is already bound to the desired port
func (k Keeper) IsBound(ctx sdk.Ctx, portID string) bool {
	_, ok := k.scopedKeeper.GetCapability(ctx, host.PortPath(portID))
	return ok
}

// BindPort defines a wrapper function for the ort Keeper's function in
// order to expose it to module's InitGenesis function
func (k Keeper) BindPort(ctx sdk.Ctx, portID string) error {
	cap := k.portKeeper.BindPort(ctx, portID)
	return k.ClaimCapability(ctx, cap, host.PortPath(portID))
}

// GetPort returns the portID for the transfer module. Used in ExportGenesis
func (k Keeper) GetPort(ctx sdk.Ctx) string {
	store := ctx.KVStore(k.storeKey)
	p, _ := store.Get(types.PortKey)
	return string(p)
}

// SetPort sets the portID for the transfer module. Used in InitGenesis
func (k Keeper) SetPort(ctx sdk.Ctx, portID string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.PortKey, []byte(portID))
}

// GetDenomTrace retreives the full identifiers trace and base denomination from the store.
func (k Keeper) GetDenomTrace(ctx sdk.Ctx, denomTraceHash tmbytes.HexBytes) (types.DenomTrace, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DenomTraceKey)
	bz, _ := store.Get(denomTraceHash)
	if len(bz) == 0 {
		return types.DenomTrace{}, false
	}

	denomTrace := k.MustUnmarshalDenomTrace(bz)
	return denomTrace, true
}

// HasDenomTrace checks if a the key with the given denomination trace hash exists on the store.
func (k Keeper) HasDenomTrace(ctx sdk.Ctx, denomTraceHash tmbytes.HexBytes) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DenomTraceKey)
	h, _ := store.Has(denomTraceHash)
	return h
}

// SetDenomTrace sets a new {trace hash -> denom trace} pair to the store.
func (k Keeper) SetDenomTrace(ctx sdk.Ctx, denomTrace types.DenomTrace) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DenomTraceKey)
	bz := k.MustMarshalDenomTrace(denomTrace)
	store.Set(denomTrace.Hash(), bz)
}

// GetAllDenomTraces returns the trace information for all the denominations.
func (k Keeper) GetAllDenomTraces(ctx sdk.Ctx) types.Traces {
	traces := types.Traces{}
	k.IterateDenomTraces(ctx, func(denomTrace types.DenomTrace) bool {
		traces = append(traces, denomTrace)
		return false
	})

	return traces.Sort()
}

// IterateDenomTraces iterates over the denomination traces in the store
// and performs a callback function.
func (k Keeper) IterateDenomTraces(ctx sdk.Ctx, cb func(denomTrace types.DenomTrace) bool) {
	store := ctx.KVStore(k.storeKey)
	iterator, _ := sdk.KVStorePrefixIterator(store, types.DenomTraceKey)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {

		denomTrace := k.MustUnmarshalDenomTrace(iterator.Value())
		if cb(denomTrace) {
			break
		}
	}
}

// AuthenticateCapability wraps the scopedKeeper's AuthenticateCapability function
func (k Keeper) AuthenticateCapability(ctx sdk.Ctx, cap *capabilitytypes.Capability, name string) bool {
	return k.scopedKeeper.AuthenticateCapability(ctx, cap, name)
}

// ClaimCapability allows the transfer module that can claim a capability that IBC module
// passes to it
func (k Keeper) ClaimCapability(ctx sdk.Ctx, cap *capabilitytypes.Capability, name string) error {
	return k.scopedKeeper.ClaimCapability(ctx, cap, name)
}

func (k Keeper) UpgradeCodec(ctx sdk.Ctx) {
	if ctx.IsOnUpgradeHeight() {
		k.ConvertState(ctx)
	}
}

func (k Keeper) ConvertState(ctx sdk.Ctx) {
	k.cdc.SetUpgradeOverride(false)
	params := k.GetParams(ctx)
	k.SetParams(ctx, params)
	k.cdc.DisableUpgradeOverride()
}

// creates a querier for staking REST endpoints
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Ctx, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {

		return nil, sdk.ErrUnknownRequest("unknown governance query endpoint")
	}
}
