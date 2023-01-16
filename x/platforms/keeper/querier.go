package keeper

import (
	"fmt"
	"math"

	"github.com/vipernet-xyz/viper-network/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/authentication/util"
	"github.com/vipernet-xyz/viper-network/x/platforms/types"

	abci "github.com/tendermint/tendermint/abci/types"
)

func paginate(page, limit int, validators []types.Platform, MaxValidators int) types.PlatformsPage {
	validatorsLen := len(validators)
	start, end := util.Paginate(validatorsLen, page, limit, MaxValidators)

	if start < 0 || end < 0 {
		validators = []types.Platform{}
	} else {
		validators = validators[start:end]
	}
	totalPages := int(math.Ceil(float64(validatorsLen) / float64(end-start)))
	if totalPages < 1 {
		totalPages = 1
	}
	platformsPage := types.PlatformsPage{Result: validators, Total: totalPages, Page: page}
	return platformsPage
}

// NewQuerier - creates a query router for staking REST endpoints
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Ctx, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case types.QueryPlatforms:
			return queryPlatforms(ctx, req, k)
		case types.QueryPlatform:
			return queryPlatform(ctx, req, k)
		case types.QueryParameters:
			return queryParameters(ctx, k)
		case types.QueryPlatformStakedPool:
			return queryStakedPool(ctx, k)
		default:
			return nil, sdk.ErrUnknownRequest("unknown staking query endpoint")
		}
	}
}

func queryPlatforms(ctx sdk.Ctx, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryPlatformsWithOpts
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}
	platforms := k.GetAllPlatformsWithOpts(ctx, params)
	platformsPage := paginate(params.Page, params.Limit, platforms, int(k.GetParams(ctx).MaxPlatforms))
	res, err := codec.MarshalJSONIndent(types.ModuleCdc, platformsPage)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to JSON marshal result: %s", err.Error()))
	}

	return res, nil
}

func queryPlatform(ctx sdk.Ctx, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryPlatformParams
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}
	platform, found := k.GetPlatform(ctx, params.Address)
	if !found {
		return nil, types.ErrNoPlatformFound(types.DefaultCodespace)
	}
	res, err := codec.MarshalJSONIndent(types.ModuleCdc, platform)
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
