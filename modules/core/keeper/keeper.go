package keeper

import (
	"fmt"
	"reflect"

	abciTypes "github.com/tendermint/tendermint/abci/types"
	"github.com/vipernet-xyz/viper-network/codec"
	paramtypes "github.com/vipernet-xyz/viper-network/types"
	sdk "github.com/vipernet-xyz/viper-network/types"
	capabilitykeeper "github.com/vipernet-xyz/viper-network/x/capability/keeper"
	viperTypes "github.com/vipernet-xyz/viper-network/x/vipernet/types"

	clientkeeper "github.com/vipernet-xyz/viper-network/modules/core/02-client/keeper"
	clienttypes "github.com/vipernet-xyz/viper-network/modules/core/02-client/types"
	connectionkeeper "github.com/vipernet-xyz/viper-network/modules/core/03-connection/keeper"
	connectiontypes "github.com/vipernet-xyz/viper-network/modules/core/03-connection/types"
	channelkeeper "github.com/vipernet-xyz/viper-network/modules/core/04-channel/keeper"
	portkeeper "github.com/vipernet-xyz/viper-network/modules/core/05-port/keeper"
	porttypes "github.com/vipernet-xyz/viper-network/modules/core/05-port/types"
	"github.com/vipernet-xyz/viper-network/modules/core/types"
)

var _ types.QueryServer = (*Keeper)(nil)

// Keeper defines each ICS keeper for IBC
type Keeper struct {
	// implements gRPC QueryServer interface
	types.QueryServer

	cdc *codec.Codec

	ClientKeeper     clientkeeper.Keeper
	ConnectionKeeper connectionkeeper.Keeper
	ChannelKeeper    channelkeeper.Keeper
	PortKeeper       portkeeper.Keeper
	Router           *porttypes.Router
}

// NewKeeper creates a new ibc Keeper
func NewKeeper(
	cdc *codec.Codec, key sdk.StoreKey, paramSpace paramtypes.Subspace, stakingKeeper viperTypes.PosKeeper,
	scopedKeeper capabilitykeeper.ScopedKeeper,
) *Keeper {
	// register paramSpace at top level keeper
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		keyTable := clienttypes.ParamKeyTable()
		keyTable.RegisterParamSet(&connectiontypes.Params{})
		paramSpace = paramSpace.WithKeyTable(keyTable)
	}
	if isEmpty(stakingKeeper) {
		panic(fmt.Errorf("cannot initialize IBC keeper: empty staking keeper"))
	}

	if reflect.DeepEqual(capabilitykeeper.ScopedKeeper{}, scopedKeeper) {
		panic(fmt.Errorf("cannot initialize IBC keeper: empty scoped keeper"))
	}

	clientKeeper := clientkeeper.NewKeeper(cdc, key, paramSpace, stakingKeeper)
	connectionKeeper := connectionkeeper.NewKeeper(cdc, key, paramSpace, clientKeeper)
	portKeeper := portkeeper.NewKeeper(scopedKeeper)
	channelKeeper := channelkeeper.NewKeeper(cdc, key, clientKeeper, connectionKeeper, portKeeper, scopedKeeper)

	return &Keeper{
		cdc:              cdc,
		ClientKeeper:     clientKeeper,
		ConnectionKeeper: connectionKeeper,
		ChannelKeeper:    channelKeeper,
		PortKeeper:       portKeeper,
	}
}

// Codec returns the IBC module codec.
func (k Keeper) Codec() *codec.Codec {
	return k.cdc
}

// SetRouter sets the Router in IBC Keeper and seals it. The method panics if
// there is an existing router that's already sealed.
func (k *Keeper) SetRouter(rtr *porttypes.Router) {
	if k.Router != nil && k.Router.Sealed() {
		panic("cannot reset a sealed router")
	}

	k.PortKeeper.Router = rtr
	k.Router = rtr
	k.Router.Seal()
}

// isEmpty checks if the interface is an empty struct or a pointer pointing
// to an empty struct
func isEmpty(keeper interface{}) bool {
	switch reflect.TypeOf(keeper).Kind() {
	case reflect.Ptr:
		if reflect.ValueOf(keeper).Elem().IsZero() {
			return true
		}
	default:
		if reflect.ValueOf(keeper).IsZero() {
			return true
		}
	}
	return false
}

// creates a querier for staking REST endpoints
func NewQuerier(k *Keeper) sdk.Querier {
	return func(ctx sdk.Ctx, path []string, req abciTypes.RequestQuery) (res []byte, err sdk.Error) {

		return nil, sdk.ErrUnknownRequest("unknown governance query endpoint")
	}
}

func (k Keeper) UpgradeCodec(ctx sdk.Ctx) {
	if ctx.IsOnUpgradeHeight() {
		k.ConvertState(ctx)
	}
}

func (k Keeper) ConvertState(ctx sdk.Ctx) {
}
