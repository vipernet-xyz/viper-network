package types

import (
	sdk "github.com/vipernet-xyz/viper-network/types"
)

// query endpoints supported by the staking Querier
const (
	QueryPlatforms            = "platforms"
	QueryPlatform             = "platform"
	QueryPlatformStakedPool   = "platformStakedPool"
	QueryPlatformUnstakedPool = "platformUnstakedPool"
	QueryParameters           = "parameters"
)

type QueryPlatformParams struct {
	Address sdk.Address
}

func NewQueryPlatformParams(platformAddr sdk.Address) QueryPlatformParams {
	return QueryPlatformParams{
		Address: platformAddr,
	}
}

type QueryPlatformsParams struct {
	Page, Limit int
}

func NewQueryPlatformsParams(page, limit int) QueryPlatformsParams {
	return QueryPlatformsParams{page, limit}
}

type QueryUnstakingPlatformsParams struct {
	Page, Limit int
}

type QueryPlatformsWithOpts struct {
	Page          int             `json:"page"`
	Limit         int             `json:"per_page"`
	StakingStatus sdk.StakeStatus `json:"staking_status"`
	Blockchain    string          `json:"blockchain"`
}

func (opts QueryPlatformsWithOpts) IsValid(platform Platform) bool {
	if opts.StakingStatus != 0 {
		if opts.StakingStatus != platform.Status {
			return false
		}
	}
	if opts.Blockchain != "" {
		var contains bool
		for _, chain := range platform.Chains {
			if chain == opts.Blockchain {
				contains = true
				break
			}
		}
		if !contains {
			return false
		}
	}
	return true
}

type QueryStakedPlatformsParams struct {
	Page, Limit int
}
