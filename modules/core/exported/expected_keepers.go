package exported

import (
	sdk "github.com/vipernet-xyz/viper-network/types"
	capabilitytypes "github.com/vipernet-xyz/viper-network/x/capability/types"
)

// ScopedKeeper defines the expected x/capability scoped keeper interface
type ScopedKeeper interface {
	NewCapability(ctx sdk.Ctx, name string) (*capabilitytypes.Capability, error)
	GetCapability(ctx sdk.Ctx, name string) (*capabilitytypes.Capability, bool)
	AuthenticateCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) bool
	LookupModules(ctx sdk.Context, name string) ([]string, *capabilitytypes.Capability, error)
	ClaimCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) error
}
