package keeper

import (
	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
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

func (k Keeper) GetSelfAddress(ctx sdk.Ctx) sdk.Address {
	kp, err := k.GetPKFromFile(ctx)
	if err != nil {
		ctx.Logger().Error("Unable to retrieve selfAddress: " + err.Error())
		return nil
	}
	return sdk.Address(kp.PublicKey().Address())
}

func (k Keeper) GetSelfPrivKey(ctx sdk.Ctx) (crypto.PrivateKey, sdk.Error) {
	// get the private key from the private validator file
	pk, er := k.GetPKFromFile(ctx)
	if er != nil {
		return nil, vc.NewKeybaseError(vc.ModuleName, er)
	}
	return pk, nil
}

// "GetSelfNode" - Gets self servicer (private val key) from the world state
func (k Keeper) GetSelfNode(ctx sdk.Ctx) (servicer exported.ValidatorI, er sdk.Error) {
	// get the servicer from the world state
	self, found := k.GetNode(ctx, k.GetSelfAddress(ctx))
	if !found {
		er = vc.NewSelfNotFoundError(vc.ModuleName)
		return nil, er
	}
	return self, nil
}

// "AwardCoinsForRelays" - Award coins to servicers for relays completed using the servicers keeper
func (k Keeper) AwardCoinsForRelays(ctx sdk.Ctx, reportCard vc.MsgSubmitReportCard, relays int64, toAddr sdk.Address, requestor requestorsTypes.Requestor) sdk.BigInt {
	return k.posKeeper.RewardForRelays(ctx, reportCard, sdk.NewInt(relays), toAddr, requestor)
}

// "BurnCoinsForChallenges" - Executes the burn for challenge function in the servicers module
func (k Keeper) BurnCoinsForChallenges(ctx sdk.Ctx, relays int64, toAddr sdk.Address) {
	k.posKeeper.BurnForChallenge(ctx, sdk.NewInt(relays), toAddr)
}
