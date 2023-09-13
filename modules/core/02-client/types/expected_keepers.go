package types

import (
	"time"

	sdk "github.com/vipernet-xyz/viper-network/types"
	stakingTypes "github.com/vipernet-xyz/viper-network/x/servicers/types"
)

// StakingKeeper expected staking keeper
type StakingKeeper interface {
	GetHistoricalInfo(ctx sdk.Ctx, height int64) (stakingTypes.HistoricalInfo, bool)
	UnbondingTime(ctx sdk.Ctx) time.Duration
}
