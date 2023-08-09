package keeper

import (
	"encoding/hex"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/servicers/exported"
	"github.com/vipernet-xyz/viper-network/x/vipernet/types"
)

// "HandleDispatch" - Handles a client request for their session information
func (k Keeper) HandleDispatch(ctx sdk.Ctx, header types.SessionHeader) (*types.DispatchResponse, sdk.Error) {
	// retrieve the latest session block height
	latestSessionBlockHeight := k.GetLatestSessionBlockHeight(ctx)
	// set the session block height
	header.SessionBlockHeight = latestSessionBlockHeight
	// validate the header
	err := header.ValidateHeader()
	if err != nil {
		return nil, err
	}
	// get the session context
	sessionCtx, er := ctx.PrevCtx(latestSessionBlockHeight)
	if er != nil {
		return nil, sdk.ErrInternal(er.Error())
	}
	// check cache
	session, found := types.GetSession(header, types.GlobalSessionCache)
	// if not found generate the session
	if !found {
		var err sdk.Error
		blockHashBz, er := sessionCtx.BlockHash(k.Cdc, sessionCtx.BlockHeight())
		if er != nil {
			return nil, sdk.ErrInternal(er.Error())
		}
		session, err = types.NewSession(sessionCtx, ctx, k.posKeeper, header, hex.EncodeToString(blockHashBz))
		if err != nil {
			return nil, err
		}
		// add to cache
		types.SetSession(session, types.GlobalSessionCache)
	}
	actualServicers := make([]exported.ValidatorI, len(session.SessionServicers))
	for i, addr := range session.SessionServicers {
		actualServicers[i], _ = k.GetNode(sessionCtx, addr)
	}
	actualFishermen := make([]exported.ValidatorI, len(session.SessionFishermen))
	for i, addr := range session.SessionFishermen {
		actualFishermen[i], _ = k.GetNode(sessionCtx, addr)
	}
	return &types.DispatchResponse{Session: types.DispatchSession{
		SessionHeader:    session.SessionHeader,
		SessionKey:       session.SessionKey,
		SessionServicers: actualServicers,
		SessionFishermen: actualFishermen,
	}, BlockHeight: ctx.BlockHeight()}, nil
}

// "IsSessionBlock" - Returns true if current block, is a session block (beginning of a session)
func (k Keeper) IsSessionBlock(ctx sdk.Ctx) bool {
	return ctx.BlockHeight()%k.posKeeper.BlocksPerSession(ctx) == 1
}

// "GetLatestSessionBlockHeight" - Returns the latest session block height (first block of the session, (see blocksPerSession))
func (k Keeper) GetLatestSessionBlockHeight(ctx sdk.Ctx) (sessionBlockHeight int64) {
	// get the latest block height
	blockHeight := ctx.BlockHeight()
	// get the blocks per session
	blocksPerSession := k.posKeeper.BlocksPerSession(ctx)
	// if block height / blocks per session remainder is zero, just subtract blocks per session and add 1
	if blockHeight%blocksPerSession == 0 {
		sessionBlockHeight = blockHeight - k.posKeeper.BlocksPerSession(ctx) + 1
	} else {
		// calculate the latest session block height by dividing the current block height by the blocksPerSession
		sessionBlockHeight = (blockHeight/blocksPerSession)*blocksPerSession + 1
	}
	return
}

// "IsViperSupportedBlockchain" - Returns true if network identifier param is supported by viper
func (k Keeper) IsViperSupportedBlockchain(ctx sdk.Ctx, chain string) bool {
	// loop through supported blockchains (network identifiers)
	for _, c := range k.SupportedBlockchains(ctx) {
		// if contains chain return true
		if c == chain {
			return true
		}
	}
	// else return false
	return false
}

func (Keeper) ClearSessionCache() {
	types.ClearSessionCache(types.GlobalSessionCache)
}
