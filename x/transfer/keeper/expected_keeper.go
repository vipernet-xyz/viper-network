package keeper

import (
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	sdk "github.com/vipernet-xyz/viper-network/types"
)

// ScopedKeeper defines the expected x/capability scoped keeper interface
type ScopedKeeper interface {
	NewCapability(ctx sdk.Ctx, name string) (*capabilitytypes.Capability, error)
	GetCapability(ctx sdk.Ctx, name string) (*capabilitytypes.Capability, bool)
	AuthenticateCapability(ctx sdk.Ctx, cap *capabilitytypes.Capability, name string) bool
	LookupModules(ctx sdk.Ctx, name string) ([]string, *capabilitytypes.Capability, error)
	ClaimCapability(ctx sdk.Ctx, cap *capabilitytypes.Capability, name string) error
}
