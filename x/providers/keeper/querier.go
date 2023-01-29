package keeper

import (
	"fmt"
	"math"

	"github.com/vipernet-xyz/viper-network/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/authentication/util"
	"github.com/vipernet-xyz/viper-network/x/providers/types"

	abci "github.com/tendermint/tendermint/abci/types"
)

func paginate(page, limit int, validators []types.Provider, MaxValidators int) types.ProvidersPage {
	validatorsLen := len(validators)
	start, end := util.Paginate(validatorsLen, page, limit, MaxValidators)

	if start < 0 || end < 0 {
		validators = []types.Provider{}
	} else {
		validators = validators[start:end]
	}
	totalPages := int(math.Ceil(float64(validatorsLen) / float64(end-start)))
	if totalPages < 1 {
		totalPages = 1
	}
	providersPage := types.ProvidersPage{Result: validators, Total: totalPages, Page: page}
	return providersPage
}

// NewQuerier - creates a query router for staking REST endpoints
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Ctx, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case types.QueryProviders:
			return queryProviders(ctx, req, k)
		case types.QueryProvider:
			return queryProvider(ctx, req, k)
		case types.QueryParameters:
			return queryParameters(ctx, k)
		case types.QueryProviderStakedPool:
			return queryStakedPool(ctx, k)
		default:
			return nil, sdk.ErrUnknownRequest("unknown staking query endpoint")
		}
	}
}

func queryProviders(ctx sdk.Ctx, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryProvidersWithOpts
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}
	providers := k.GetAllProvidersWithOpts(ctx, params)
	providersPage := paginate(params.Page, params.Limit, providers, int(k.GetParams(ctx).MaxProviders))
	res, err := codec.MarshalJSONIndent(types.ModuleCdc, providersPage)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to JSON marshal result: %s", err.Error()))
	}

	return res, nil
}

func queryProvider(ctx sdk.Ctx, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryProviderParams
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}
	provider, found := k.GetProvider(ctx, params.Address)
	if !found {
		return nil, types.ErrNoProviderFound(types.DefaultCodespace)
	}
	res, err := codec.MarshalJSONIndent(types.ModuleCdc, provider)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return res, nil
}

func queryStakedPool(ctx sdk.Ctx, k Keeper) ([]byte, sdk.Error) {
	stakedTokens := k.GetStakedTokens(ctx)
	res, err := codec.MarshalJSONIndent(types.ModuleCdc, stakedTokens)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return res, nil
}

func queryParameters(ctx sdk.Ctx, k Keeper) ([]byte, sdk.Error) {
	params := k.GetParams(ctx)
	res, err := codec.MarshalJSONIndent(types.ModuleCdc, params)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return res, nil
}
