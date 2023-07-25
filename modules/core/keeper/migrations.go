package keeper

import (
	sdk "github.com/vipernet-xyz/viper-network/types"

	clientkeeper "github.com/vipernet-xyz/viper-network/modules/core/02-client/keeper"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper Keeper
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper Keeper) Migrator {
	return Migrator{keeper: keeper}
}

// Migrate2to3 migrates from version 2 to 3. See 02-client keeper function Migrate2to3.
func (m Migrator) Migrate2to3(ctx sdk.Ctx) error {
	clientMigrator := clientkeeper.NewMigrator(m.keeper.ClientKeeper)
	if err := clientMigrator.Migrate2to3(ctx); err != nil {
		return err
	}

	return nil
}
