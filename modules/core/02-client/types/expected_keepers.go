package types

import (
	"time"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	sdk "github.com/vipernet-xyz/viper-network/types"
)

// StakingKeeper expected staking keeper
type StakingKeeper interface {
	GetHistoricalInfo(ctx sdk.Ctx, height int64) (stakingtypes.HistoricalInfo, bool)
	UnbondingTime(ctx sdk.Ctx) time.Duration
}

// UpgradeKeeper expected upgrade keeper
type UpgradeKeeper interface {
	ClearIBCState(ctx sdk.Context, lastHeight int64)
	GetUpgradePlan(ctx sdk.Ctx) (plan upgradetypes.Plan, havePlan bool)
	GetUpgradedClient(ctx sdk.Ctx, height int64) ([]byte, bool)
	SetUpgradedClient(ctx sdk.Context, planHeight int64, bz []byte) error
	GetUpgradedConsensusState(ctx sdk.Context, lastHeight int64) ([]byte, bool)
	SetUpgradedConsensusState(ctx sdk.Ctx, planHeight int64, bz []byte) error
	ScheduleUpgrade(ctx sdk.Context, plan upgradetypes.Plan) error
}
