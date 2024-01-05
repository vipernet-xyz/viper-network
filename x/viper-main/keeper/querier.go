package keeper

import (
	"fmt"

	"github.com/vipernet-xyz/viper-network/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/viper-main/types"

	abci "github.com/tendermint/tendermint/abci/types"
)

// "NewQuerier" - Creates an sdk.Querier for the viper core module
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Ctx, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		// query viper supported supported non-native blockchains
		case types.QuerySupportedBlockchains:
			return querySupportedBlockchains(ctx, req, k)
		// query viper supported geozones
		case types.QuerySupportedGeoZones:
			return querySupportedGeoZones(ctx, req, k)
		// query the parameters of the vipernet module
		case types.QueryParameters:
			return queryParameters(ctx, k)
		// endpoint allowing a client to query a relay to a non-native blockchain
		case types.QueryRelay:
			return queryRelay(ctx, req, k)
		// endpoint allowing a client to receive the servicers for their session
		case types.QueryDispatch:
			return queryDispatch(ctx, req, k)
		// endpoint allowing a client to submit a challenge for an invalid relay-response
		case types.QueryChallenge:
			return queryChallenge(ctx, req, k)
		default:
			return nil, sdk.ErrUnknownRequest("unknown vipernet query endpoint")
		}
	}
}

// "queryChallenge" - Is a handler for the challenge query
// The challenge query allows clients to submit a challenge for invalid relay responses
func queryChallenge(ctx sdk.Ctx, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	// unmarshal data into a query params object
	var params types.QueryChallengeParams
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}
	// handle the challenge from the params
	response, er := k.HandleChallenge(ctx, params.Challenge)
	if er != nil {
		return nil, er
	}
	// marshal the response data into amino-json
	res, err := codec.MarshalJSONIndent(types.ModuleCdc, response)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to JSON marshal result: %s", err.Error()))
	}
	return res, nil
}

// "queryRelay" - Is a handler for the relay query
// The relay query allows clients to submit a request to a non-native blockchain
func queryRelay(ctx sdk.Ctx, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	// unmarshal data into a query params object
	var params types.QueryRelayParams
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}
	// handle the relay from the params
	response, er := k.HandleRelay(ctx, params.Relay)
	if er != nil {
		return nil, er
	}
	// marshals the response data into amino-json
	res, err := codec.MarshalJSONIndent(types.ModuleCdc, response)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to JSON marshal result: %s", err.Error()))
	}
	return res, nil
}

// "queryDispatch" - Is a handler for the dispatch query
// The dispatch query allows clients retrieve their session information
func queryDispatch(ctx sdk.Ctx, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	// unmarshal data into a query params object
	var params types.QueryDispatchParams
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}
	// handle the dispatch from the params
	response, er := k.HandleDispatch(ctx, params.SessionHeader)
	if er != nil {
		return nil, er
	}
	// marshals the response data into amino-json
	res, err := codec.MarshalJSONIndent(types.ModuleCdc, *response)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to JSON marshal result: %s", err.Error()))
	}
	return res, nil
}

// "queryParameters" - Is a handler for the parameters query
// Returns all the parameters in the module
func queryParameters(ctx sdk.Ctx, k Keeper) ([]byte, sdk.Error) {
	// get the params
	params := k.GetParams(ctx)
	// marshal response data into amino-json
	res, err := codec.MarshalJSONIndent(types.ModuleCdc, params)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}
	return res, nil
}

// "querySupportedBlockchains" - Is a handler for the supported blockchains query
// Returns the non native chains supported on viper network
func querySupportedBlockchains(ctx sdk.Ctx, _ abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	// marshal supported blockchains into amino-json
	res, err := codec.MarshalJSONIndent(types.ModuleCdc, k.SupportedBlockchains(ctx))
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to JSON marshal result: %s", err.Error()))
	}
	return res, nil
}

// "querySupportedGeopZones" - Is a handler for the supported geozones query
// Returns the geozones supported on viper network
func querySupportedGeoZones(ctx sdk.Ctx, _ abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	res, err := codec.MarshalJSONIndent(types.ModuleCdc, k.SupportedGeoZones(ctx))
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to JSON marshal result: %s", err.Error()))
	}
	return res, nil
}
