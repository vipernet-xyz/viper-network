package types

import (
	"time"

	sdk "github.com/vipernet-xyz/viper-network/types"
)

// StakingKeeper expected staking keeper
type StakingKeeper interface {
	GetHistoricalInfo(ctx sdk.Ctx, height int64) (HistoricalInfo, bool)
	UnbondingTime(ctx sdk.Ctx) time.Duration
}

// UpgradeKeeper expected upgrade keeper
type UpgradeKeeper interface {
	//ClearIBCState(ctx sdk.Ctx, lastHeight int64)
	GetUpgradePlan(ctx sdk.Ctx) (plan Plan, havePlan bool)
	GetUpgradedClient(ctx sdk.Ctx, height int64) ([]byte, bool)
	SetUpgradedClient(ctx sdk.Ctx, planHeight int64, bz []byte) error
	GetUpgradedConsensusState(ctx sdk.Ctx, lastHeight int64) ([]byte, bool)
	SetUpgradedConsensusState(ctx sdk.Ctx, planHeight int64, bz []byte) error
	ScheduleUpgrade(ctx sdk.Ctx, plan Plan) error
}
