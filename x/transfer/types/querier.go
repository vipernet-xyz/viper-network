package types

import (
	sdk "github.com/vipernet-xyz/viper-network/types"
)

// query endpoints supported by the staking Querier
const (
	QueryStakedPool     = "stakedPool"
	QueryUnstakedPool   = "unstakedPool"
	QueryParameters     = "parameters"
	QueryTotalSupply    = "total_supply"
	QuerySigningInfo    = "signingInfo"
	QuerySigningInfos   = "signingInfos"
	QueryAccountBalance = "account_balance"
	QueryAccount        = "account"
)

type QueryValidatorParams struct {
	Address sdk.Address
}

func NewQueryValidatorParams(validatorAddr sdk.Address) QueryValidatorParams {
	return QueryValidatorParams{
		Address: validatorAddr,
	}
}

type QueryValidatorsParams struct {
	StakingStatus sdk.StakeStatus `json:"staking_status"`
	JailedStatus  int             `json:"jailed_status"`
	Blockchain    string          `json:"blockchain"`
	Page          int             `json:"page"`
	Limit         int             `json:"per_page"`
}

type QueryAccountBalanceParams struct {
	sdk.Address
}

type QueryAccountParams struct {
	sdk.Address
}

type QueryUnstakingValidatorsParams struct {
	Page, Limit int
}

func NewQueryUnstakingValidatorsParams(page, limit int) QueryUnstakingValidatorsParams {
	return QueryUnstakingValidatorsParams{page, limit}
}

type QueryStakedValidatorsParams struct {
	Page, Limit int
}

func NewQueryStakedValidatorsParams(page, limit int) QueryStakedValidatorsParams {
	return QueryStakedValidatorsParams{page, limit}
}

type QueryUnstakedValidatorsParams struct {
	Page, Limit int
}

func NewQueryUnstakedValidatorsParams(page, limit int) QueryUnstakedValidatorsParams {
	return QueryUnstakedValidatorsParams{page, limit}
}

type QuerySigningInfoParams struct {
	Address sdk.Address
}

func NewQuerySigningInfoParams(consAddr sdk.Address) QuerySigningInfoParams {
	return QuerySigningInfoParams{consAddr}
}

type QuerySigningInfosParams struct {
	Page, Limit int
}

func NewQuerySigningInfosParams(page, limit int) QuerySigningInfosParams {
	return QuerySigningInfosParams{page, limit}
}
