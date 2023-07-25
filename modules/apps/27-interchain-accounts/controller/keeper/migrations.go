package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/vipernet-xyz/viper-network/types"
	capabilitytypes "github.com/vipernet-xyz/viper-network/x/capability/types"

	controllertypes "github.com/vipernet-xyz/viper-network/modules/apps/27-interchain-accounts/controller/types"
	icatypes "github.com/vipernet-xyz/viper-network/modules/apps/27-interchain-accounts/types"
	host "github.com/vipernet-xyz/viper-network/modules/core/24-host"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper *Keeper
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper *Keeper) Migrator {
	return Migrator{keeper: keeper}
}

// AssertChannelCapabilityMigrations checks that all channel capabilities generated using the interchain accounts controller port prefix
// are owned by the controller submodule and ibc.
func (m Migrator) AssertChannelCapabilityMigrations(ctx sdk.Ctx) error {
	if m.keeper != nil {
		logger := m.keeper.Logger(ctx)
		filteredChannels := m.keeper.channelKeeper.GetAllChannelsWithPortPrefix(ctx, icatypes.ControllerPortPrefix)
		for _, ch := range filteredChannels {
			name := host.ChannelCapabilityPath(ch.PortId, ch.ChannelId)
			capability, found := m.keeper.scopedKeeper.GetCapability(ctx, name)
			if !found {
				logger.Error(fmt.Sprintf("failed to find capability: %s", name))
				return errorsmod.Wrapf(capabilitytypes.ErrCapabilityNotFound, "failed to find capability: %s", name)
			}

			isAuthenticated := m.keeper.scopedKeeper.AuthenticateCapability(ctx, capability, name)
			if !isAuthenticated {
				logger.Error(fmt.Sprintf("expected capability owner: %s", controllertypes.SubModuleName))
				return errorsmod.Wrapf(capabilitytypes.ErrCapabilityNotOwned, "expected capability owner: %s", controllertypes.SubModuleName)
			}

			m.keeper.SetMiddlewareEnabled(ctx, ch.PortId, ch.ConnectionHops[0])
			logger.Info("successfully migrated channel capability", "name", name)
		}
	}
	return nil
}
