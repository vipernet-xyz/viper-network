package types

import (
	"fmt"

	sdk "github.com/vipernet-xyz/viper-network/types"
)

// names used as root for pool module accounts:
// StakingPool -> "provider_staked_tokens_pool"
const (
	StakedPoolName = "provider_staked_tokens_pool"
)

// Tokens - tracking staked token supply
type StakingPool Pool

// NewPool creates a new Tokens instance used for queries
func NewPool(tokens sdk.BigInt) Pool {
	return Pool{
		Tokens: tokens,
	}
}

// String returns a human readable string representation of a pool.
func (bp StakingPool) String() string {
	return fmt.Sprintf(`Staked Tokens:      %s`, bp.Tokens)
}
