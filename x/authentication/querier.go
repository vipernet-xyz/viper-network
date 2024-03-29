package authentication

import (
	"fmt"

	"github.com/vipernet-xyz/viper-network/x/authentication/keeper"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/vipernet-xyz/viper-network/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/authentication/types"
)

// NewQuerier creates a querier for authentication REST endpoints
func NewQuerier(keeper keeper.Keeper) sdk.Querier {
	return func(ctx sdk.Ctx, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
		switch path[0] {
		case types.QueryAccount:
			return queryAccount(ctx, req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown authentication query endpoint")
		}
	}
}

func queryAccount(ctx sdk.Ctx, req abci.RequestQuery, keeper keeper.Keeper) ([]byte, sdk.Error) {
	var params types.QueryAccountParams
	if err := types.ModuleCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	account := keeper.GetAccount(ctx, params.Address)
	if account == nil {
		allAccs := keeper.GetAllAccounts(ctx)
		_ = allAccs
		return nil, sdk.ErrUnknownAddress(fmt.Sprintf("account %s does not exist", params.Address))
	}

	bz, err := codec.MarshalJSONIndent(types.ModuleCdc, account)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return bz, nil
}
