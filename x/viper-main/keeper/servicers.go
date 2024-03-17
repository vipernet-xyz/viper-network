package keeper

import (
	sdk "github.com/vipernet-xyz/viper-network/types"
	requestorsTypes "github.com/vipernet-xyz/viper-network/x/requestors/types"
	"github.com/vipernet-xyz/viper-network/x/servicers/exported"
	vc "github.com/vipernet-xyz/viper-network/x/viper-main/types"
)

// "GetNode" - Gets a servicer from the state storage
func (k Keeper) GetNode(ctx sdk.Ctx, address sdk.Address) (n exported.ValidatorI, found bool) {
	n = k.posKeeper.Validator(ctx, address)
	if n == nil {
		return n, false
	}
	return n, true
}

// "AwardCoinsForRelays" - Award coins to servicers for relays completed using the servicers keeper
func (k Keeper) AwardCoinsForRelays(ctx sdk.Ctx, reportCard vc.MsgSubmitQoSReport, relays int64, toAddr sdk.Address, requestor requestorsTypes.Requestor) sdk.BigInt {
	tokens := k.posKeeper.RewardForRelays(ctx, reportCard, sdk.NewInt(relays), toAddr, requestor)
	return tokens
}

// "BurnCoinsForChallenges" - Executes the burn for challenge function in the servicers module
func (k Keeper) BurnCoinsForChallenges(ctx sdk.Ctx, relays int64, toAddr sdk.Address) {
	k.posKeeper.BurnForChallenge(ctx, sdk.NewInt(relays), toAddr)
}
