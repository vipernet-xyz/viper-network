package keeper

import (
	"fmt"
	"time"

	"github.com/tendermint/tendermint/libs/strings"

	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/requestors/types"
)

// ValidateRequestorStaking - Check requestor before staking
func (k Keeper) ValidateRequestorStaking(ctx sdk.Ctx, requestor types.Requestor, amount sdk.BigInt) sdk.Error {
	// convert the amount to sdk.Coin
	coin := sdk.NewCoins(sdk.NewCoin(k.StakeDenom(ctx), amount))
	if int64(len(requestor.Chains)) > k.MaxChains(ctx) {
		return types.ErrTooManyChains(types.ModuleName)
	}
	if requestor.NumServicers < int64(k.MinNumServicers(ctx)) || requestor.NumServicers > int64(k.MaxNumServicers(ctx)) {
		return types.ErrNumServicers(types.ModuleName)
	}
	// attempt to get the requestor from the world state
	app, found := k.GetRequestor(ctx, requestor.Address)
	// if the requestor exists
	if found {
		// edit stake in 6.X upgrade
		if app.IsStaked() {
			return k.ValidateEditStake(ctx, app, amount)
		}
		if !app.IsUnstaked() { // unstaking or already staked but before the upgrade
			return types.ErrRequestorStatus(k.codespace)
		}
	} else {
		// ensure public key type is supported
		if ctx.ConsensusParams() != nil {
			tmPubKey, err := crypto.CheckConsensusPubKey(requestor.PublicKey.PubKey())
			if err != nil {
				return types.ErrRequestorPubKeyTypeNotSupported(k.Codespace(),
					err.Error(),
					ctx.ConsensusParams().Validator.PubKeyTypes)
			}
			if !strings.StringInSlice(tmPubKey.Type, ctx.ConsensusParams().Validator.PubKeyTypes) {
				return types.ErrRequestorPubKeyTypeNotSupported(k.Codespace(),
					tmPubKey.Type,
					ctx.ConsensusParams().Validator.PubKeyTypes)
			}
		}
	}
	// ensure the amount they are staking is < the minimum stake amount
	if amount.LT(sdk.NewInt(k.MinimumStake(ctx))) {
		return types.ErrMinimumStake(k.codespace)
	}
	if !k.AccountKeeper.HasCoins(ctx, requestor.Address, coin) {
		return types.ErrNotEnoughCoins(k.codespace)
	}

	if k.getStakedRequestorsCount(ctx) >= k.MaxRequestors(ctx) {
		return types.ErrMaxRequestors(k.codespace)
	}

	return nil
}

// ValidateEditStake - Validate the updates to a current staked validator
func (k Keeper) ValidateEditStake(ctx sdk.Ctx, currentApp types.Requestor, amount sdk.BigInt) sdk.Error {
	// ensure not staking less
	diff := amount.Sub(currentApp.StakedTokens)
	if diff.IsNegative() {
		return types.ErrMinimumEditStake(k.codespace)
	}
	// if stake bump
	if !diff.IsZero() {
		// ensure account has enough coins for bump
		coin := sdk.NewCoins(sdk.NewCoin(k.StakeDenom(ctx), diff))
		if !k.AccountKeeper.HasCoins(ctx, currentApp.Address, coin) {
			return types.ErrNotEnoughCoins(k.Codespace())
		}
	}
	return nil
}

// StakeRequestor - Store ops when a requestor stakes
func (k Keeper) StakeRequestor(ctx sdk.Ctx, requestor types.Requestor, amount sdk.BigInt) sdk.Error {
	// edit stake
	// get Validator to see if edit stake
	curApp, found := k.GetRequestor(ctx, requestor.Address)
	if found && curApp.IsStaked() {
		return k.EditStakeRequestor(ctx, curApp, requestor, amount)
	}
	// send the coins from address to staked module account
	err := k.coinsFromUnstakedToStaked(ctx, requestor, amount)
	if err != nil {
		return err
	}
	// add coins to the staked field
	requestor, er := requestor.AddStakedTokens(amount)
	if er != nil {
		return sdk.ErrInternal(er.Error())
	}
	// calculate relays
	requestor.MaxRelays = k.CalculateRequestorRelays(ctx, requestor)
	// set the status to staked
	requestor = requestor.UpdateStatus(sdk.Staked)
	// save in the requestor store
	k.SetRequestor(ctx, requestor)
	return nil
}

func (k Keeper) EditStakeRequestor(ctx sdk.Ctx, requestor, updatedRequestor types.Requestor, amount sdk.BigInt) sdk.Error {
	origAppForDeletion := requestor
	// get the difference in coins
	diff := amount.Sub(requestor.StakedTokens)
	// if they bumped the stake amount
	if diff.IsPositive() {
		// send the coins from address to staked module account
		err := k.coinsFromUnstakedToStaked(ctx, requestor, diff)
		if err != nil {
			return err
		}
		var er error
		// add coins to the staked field
		requestor, er = requestor.AddStakedTokens(diff)
		if er != nil {
			return sdk.ErrInternal(er.Error())
		}
		// update apps max relays
		requestor.MaxRelays = k.CalculateRequestorRelays(ctx, requestor)
	}
	// update chains
	requestor.Chains = updatedRequestor.Chains
	// update geozones
	requestor.GeoZones = updatedRequestor.GeoZones
	// update numservicers
	requestor.NumServicers = updatedRequestor.NumServicers
	// delete the validator from the staking set
	k.deleteRequestorFromStakingSet(ctx, origAppForDeletion)
	// delete in main store
	k.DeleteRequestor(ctx, origAppForDeletion.Address)
	// save in the app store
	k.SetRequestor(ctx, requestor)
	// save the app by chains
	k.SetStakedRequestor(ctx, requestor)
	// clear session cache
	k.ViperKeeper.ClearSessionCache()
	// log success
	ctx.Logger().Info("Successfully updated staked requestor: " + requestor.Address.String())
	return nil
}

// ValidateRequestorBeginUnstaking - Check for validator status
func (k Keeper) ValidateRequestorBeginUnstaking(ctx sdk.Ctx, requestor types.Requestor) sdk.Error {
	// must be staked to begin unstaking
	if !requestor.IsStaked() {
		return sdk.ErrInternal(types.ErrRequestorStatus(k.codespace).Error())
	}
	if requestor.IsJailed() {
		return sdk.ErrInternal(types.ErrRequestorJailed(k.codespace).Error())
	}
	return nil
}

// BeginUnstakingRequestor - Store ops when requestor begins to unstake -> starts the unstaking timer
func (k Keeper) BeginUnstakingRequestor(ctx sdk.Ctx, requestor types.Requestor) {
	// get params
	params := k.GetParams(ctx)
	// delete the requestor from the staking set, as it is technically staked but not going to participate
	k.deleteRequestorFromStakingSet(ctx, requestor)
	// set the status
	requestor = requestor.UpdateStatus(sdk.Unstaking)
	// set the unstaking completion time and completion height appropriately
	if requestor.UnstakingCompletionTime.IsZero() {
		requestor.UnstakingCompletionTime = ctx.BlockHeader().Time.Add(params.UnstakingTime)
	}
	// save the now unstaked requestor record and power index
	k.SetRequestor(ctx, requestor)
	ctx.Logger().Info("Began unstaking App " + requestor.Address.String())
}

// ValidateRequestorFinishUnstaking - Check if requestor can finish unstaking
func (k Keeper) ValidateRequestorFinishUnstaking(ctx sdk.Ctx, requestor types.Requestor) sdk.Error {
	if !requestor.IsUnstaking() {
		return types.ErrRequestorStatus(k.codespace)
	}
	if requestor.IsJailed() {
		return types.ErrRequestorJailed(k.codespace)
	}
	return nil
}

// FinishUnstakingRequestor - Store ops to unstake a client -> called after unstaking time is up
func (k Keeper) FinishUnstakingRequestor(ctx sdk.Ctx, requestor types.Requestor) {
	// delete the requestor from the unstaking queue
	k.deleteUnstakingRequestor(ctx, requestor)
	// amount unstaked = stakedTokens
	amount := requestor.StakedTokens
	// send the tokens from staking module account to requestor account
	err := k.coinsFromStakedToUnstaked(ctx, requestor)
	if err != nil {
		k.Logger(ctx).Error("could not move coins from staked to unstaked for applications module" + err.Error() + "for this app address: " + requestor.Address.String())
		// continue with the unstaking
	}
	// removed the staked tokens field from requestor structure
	requestor, er := requestor.RemoveStakedTokens(amount)
	if er != nil {
		k.Logger(ctx).Error("could not remove tokens from unstaking requestor: " + er.Error())
		// continue with the unstaking
	}
	// update the status to unstaked
	requestor = requestor.UpdateStatus(sdk.Unstaked)
	// reset app relays
	requestor.MaxRelays = sdk.ZeroInt()
	// update the unstaking time
	requestor.UnstakingCompletionTime = time.Time{}
	// update the requestor in the main store
	k.SetRequestor(ctx, requestor)
	ctx.Logger().Info("Finished unstaking requestor " + requestor.Address.String())
	// create the event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeUnstake,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, requestor.Address.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, requestor.Address.String()),
		),
	})
}

// LegacyForceRequestorUnstake - Coerce unstake (called when slashed below the minimum)
func (k Keeper) LegacyForceRequestorUnstake(ctx sdk.Ctx, requestor types.Requestor) sdk.Error {
	// delete the requestor from staking set as they are unstaked
	k.deleteRequestorFromStakingSet(ctx, requestor)
	// amount unstaked = stakedTokens
	err := k.burnStakedTokens(ctx, requestor.StakedTokens)
	if err != nil {
		return err
	}
	// remove their tokens from the field
	requestor, er := requestor.RemoveStakedTokens(requestor.StakedTokens)
	if er != nil {
		return sdk.ErrInternal(er.Error())
	}
	// update their status to unstaked
	requestor = requestor.UpdateStatus(sdk.Unstaked)
	// reset app relays
	requestor.MaxRelays = sdk.ZeroInt()
	// set the requestor in store
	k.SetRequestor(ctx, requestor)
	ctx.Logger().Info("Force Unstaked requestor " + requestor.Address.String())
	// create the event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeUnstake,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, requestor.Address.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, requestor.Address.String()),
		),
	})
	return nil
}

// ForceValidatorUnstake - Coerce unstake (called when slashed below the minimum)
func (k Keeper) ForceRequestorUnstake(ctx sdk.Ctx, requestor types.Requestor) sdk.Error {
	switch requestor.Status {
	case sdk.Staked:
		k.deleteRequestorFromStakingSet(ctx, requestor)
	case sdk.Unstaking:
		k.deleteUnstakingRequestor(ctx, requestor)
		k.DeleteRequestor(ctx, requestor.Address)
	default:
		k.DeleteRequestor(ctx, requestor.Address)
		return sdk.ErrInternal("should not happen: trying to force unstake an already unstaked requestor: " + requestor.Address.String())
	}
	// amount unstaked = stakedTokens
	err := k.burnStakedTokens(ctx, requestor.StakedTokens)
	if err != nil {
		return err
	}
	if requestor.IsStaked() {
		// remove their tokens from the field
		validator, er := requestor.RemoveStakedTokens(requestor.StakedTokens)
		if er != nil {
			return sdk.ErrInternal(er.Error())
		}
		// update their status to unstaked
		validator = validator.UpdateStatus(sdk.Unstaked)
		// set the validator in store
		k.SetRequestor(ctx, validator)
	}
	ctx.Logger().Info("Force Unstaked validator " + requestor.Address.String())
	return nil
}

// JailRequestor - Send a requestor to jail
func (k Keeper) JailRequestor(ctx sdk.Ctx, addr sdk.Address) {
	requestor, found := k.GetRequestor(ctx, addr)
	if !found {
		k.Logger(ctx).Error(fmt.Errorf("requestor %s is attempted jailed but not found in all applications store", addr).Error())
		return
	}
	if requestor.Jailed {
		k.Logger(ctx).Error(fmt.Sprintf("cannot jail already jailed requestor, requestor: %v\n", requestor))
		return
	}
	requestor.Jailed = true
	k.SetRequestor(ctx, requestor)
	k.deleteRequestorFromStakingSet(ctx, requestor)
	logger := k.Logger(ctx)
	logger.Info(fmt.Sprintf("requestor %s jailed", addr))
}

func (k Keeper) IncrementJailedRequestors(ctx sdk.Ctx) {
	// TODO
}

// ValidateUnjailMessage - Check unjail message
func (k Keeper) ValidateUnjailMessage(ctx sdk.Ctx, msg types.MsgUnjail) (addr sdk.Address, err sdk.Error) {
	requestor, found := k.GetRequestor(ctx, msg.RequestorAddr)
	if !found {
		return nil, types.ErrNoRequestorForAddress(k.Codespace())
	}
	// cannot be unjailed if not staked
	stake := requestor.GetTokens()
	if stake == sdk.ZeroInt() {
		return nil, types.ErrMissingRequestorStake(k.Codespace())
	}
	if requestor.GetTokens().LT(sdk.NewInt(k.MinimumStake(ctx))) { // TODO look into this state change (stuck in jail)
		return nil, types.ErrStakeTooLow(k.Codespace())
	}
	// cannot be unjailed if not jailed
	if !requestor.IsJailed() {
		return nil, types.ErrRequestorNotJailed(k.Codespace())
	}
	return
}

// UnjailRequestor - Remove a requestor from jail
func (k Keeper) UnjailRequestor(ctx sdk.Ctx, addr sdk.Address) {
	requestor, found := k.GetRequestor(ctx, addr)
	if !found {
		k.Logger(ctx).Error(fmt.Errorf("requestor %s is attempted jailed but not found in all applications store", addr).Error())
		return
	}
	if !requestor.Jailed {
		k.Logger(ctx).Error(fmt.Sprintf("cannot unjail already unjailed requestor, requestor: %v\n", requestor))
		return
	}
	requestor.Jailed = false
	k.SetRequestor(ctx, requestor)
	logger := k.Logger(ctx)
	logger.Info(fmt.Sprintf("requestor %s unjailed", addr))
}

func (k Keeper) BurnRequestorStake(ctx sdk.Ctx, requestor types.Requestor, amount sdk.BigInt) {
	logger := k.Logger(ctx)
	tokensToBurn := sdk.MinInt(amount, requestor.StakedTokens)
	tokensToBurn = sdk.MaxInt(tokensToBurn, sdk.ZeroInt())
	requestor, err := k.removeRequestorTokens(ctx, requestor, amount)
	if err != nil {
		k.Logger(ctx).Error("could not remove staked tokens: " + err.Error() + "\nfor requestor " + requestor.Address.String())
		return
	}
	err = k.burnStakedTokens(ctx, tokensToBurn)
	if err != nil {
		k.Logger(ctx).Error("could not burn staked tokens in burn: " + err.Error() + "\nfor requestor " + requestor.Address.String())
		return
	}
	// if falls below minimum force burn all of the stake
	if requestor.GetTokens().LT(sdk.NewInt(k.MinimumStake(ctx))) {
		err = k.ForceRequestorUnstake(ctx, requestor)

		if err != nil {
			k.Logger(ctx).Error("could not force unstake in burn: " + err.Error() + "\nfor requestor " + requestor.Address.String())
			return
		}
	}
	// Log that burn occured
	logger.Info(fmt.Sprintf("requestor %s burned by amount of %s; burned %v tokens",
		requestor.GetAddress(), amount.String(), tokensToBurn))
}
