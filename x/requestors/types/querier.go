package types

import (
	sdk "github.com/vipernet-xyz/viper-network/types"
)

// query endpoints supported by the staking Querier
const (
	QueryRequestors            = "requestors"
	QueryRequestor             = "requestor"
	QueryRequestorStakedPool   = "requestorStakedPool"
	QueryRequestorUnstakedPool = "requestorUnstakedPool"
	QueryParameters            = "parameters"
)

type QueryRequestorParams struct {
	Address sdk.Address
}

func NewQueryRequestorParams(requestorAddr sdk.Address) QueryRequestorParams {
	return QueryRequestorParams{
		Address: requestorAddr,
	}
}

type QueryRequestorsParams struct {
	Page, Limit int
}

func NewQueryRequestorsParams(page, limit int) QueryRequestorsParams {
	return QueryRequestorsParams{page, limit}
}

type QueryUnstakingRequestorsParams struct {
	Page, Limit int
}

type QueryRequestorsWithOpts struct {
	Page          int             `json:"page"`
	Limit         int             `json:"per_page"`
	StakingStatus sdk.StakeStatus `json:"staking_status"`
	Blockchain    string          `json:"blockchain"`
	GeoZone       string          `json:"geo_zone"`
}

func (opts QueryRequestorsWithOpts) IsValid(requestor Requestor) bool {
	if opts.StakingStatus != 0 {
		if opts.StakingStatus != requestor.Status {
			return false
		}
	}
	if opts.Blockchain != "" {
		var contains bool
		for _, chain := range requestor.Chains {
			if chain == opts.Blockchain {
				contains = true
				break
			}
		}
		if !contains {
			return false
		}
	}
	if opts.GeoZone != "" {
		var contains bool
		for _, geozone := range requestor.GeoZones {
			if geozone == opts.GeoZone {
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

type QueryStakedRequestorsParams struct {
	Page, Limit int
}
