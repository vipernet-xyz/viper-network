package types

import (
	sdk "github.com/vipernet-xyz/viper-network/types"
)

// query endpoints supported by the staking Querier
const (
	QueryProviders            = "providers"
	QueryProvider             = "provider"
	QueryProviderStakedPool   = "providerStakedPool"
	QueryProviderUnstakedPool = "providerUnstakedPool"
	QueryParameters           = "parameters"
)

type QueryProviderParams struct {
	Address sdk.Address
}

func NewQueryProviderParams(providerAddr sdk.Address) QueryProviderParams {
	return QueryProviderParams{
		Address: providerAddr,
	}
}

type QueryProvidersParams struct {
	Page, Limit int
}

func NewQueryProvidersParams(page, limit int) QueryProvidersParams {
	return QueryProvidersParams{page, limit}
}

type QueryUnstakingProvidersParams struct {
	Page, Limit int
}

type QueryProvidersWithOpts struct {
	Page          int             `json:"page"`
	Limit         int             `json:"per_page"`
	StakingStatus sdk.StakeStatus `json:"staking_status"`
	Blockchain    string          `json:"blockchain"`
	GeoZone       string          `json:"geo_zone"`
}

func (opts QueryProvidersWithOpts) IsValid(provider Provider) bool {
	if opts.StakingStatus != 0 {
		if opts.StakingStatus != provider.Status {
			return false
		}
	}
	if opts.Blockchain != "" {
		var contains bool
		for _, chain := range provider.Chains {
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
		for _, geozone := range provider.GeoZones {
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

type QueryStakedProvidersParams struct {
	Page, Limit int
}
