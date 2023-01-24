package keeper

import (
	"fmt"
	"time"

	"github.com/tendermint/tendermint/libs/strings"

	"github.com/vipernet-xyz/viper-network/crypto"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/platforms/types"
)

// ValidatePlatformStaking - Check application before staking
func (k Keeper) ValidatePlatformStaking(ctx sdk.Ctx, application types.Platform, amount sdk.BigInt) sdk.Error {
	// convert the amount to sdk.Coin
	coin := sdk.NewCoins(sdk.NewCoin(k.StakeDenom(ctx), amount))
	if int64(len(application.Chains)) > k.MaxChains(ctx) {
		return types.ErrTooManyChains(types.ModuleName)
	}
	// attempt to get the application from the world state
	app, found := k.GetPlatform(ctx, application.Address)
	// if the application exists
	if found {
		// edit stake in 6.X upgrade
		if ctx.IsAfterUpgradeHeight() && app.IsStaked() {
			return k.ValidateEditStake(ctx, app, amount)
		}
		if !app.IsUnstaked() { // unstaking or already staked but before the upgrade
			return types.ErrPlatformStatus(k.codespace)
		}
	} else {
		// ensure public key type is supported
		if ctx.ConsensusParams() != nil {
			tmPubKey, err := crypto.CheckConsensusPubKey(application.PublicKey.PubKey())
			if err != nil {
				return types.ErrPlatformPubKeyTypeNotSupported(k.Codespace(),
					err.Error(),
					ctx.ConsensusParams().Validator.PubKeyTypes)
			}
			if !strings.StringInSlice(tmPubKey.Type, ctx.ConsensusParams().Validator.PubKeyTypes) {
				return types.ErrPlatformPubKeyTypeNotSupported(k.Codespace(),
					tmPubKey.Type,
					ctx.ConsensusParams().Validator.PubKeyTypes)
			}
		}
	}
	// ensure the amount they are staking is < the minimum stake amount
	if amount.LT(sdk.NewInt(k.MinimumStake(ctx))) {
		return types.ErrMinimumStake(k.codespace)
	}
	if !k.AccountKeeper.HasCoins(ctx, application.Address, coin) {
		return types.ErrNotEnoughCoins(k.codespace)
	}
	if ctx.IsAfterUpgradeHeight() {
		if k.getStakedPlatformsCount(ctx) >= k.MaxPlatforms(ctx) {
			return types.ErrMaxPlatforms(k.codespace)
		}
	}
	return nil
}

// ValidateEditStake - Validate the updates to a current staked validator
func (k Keeper) ValidateEditStake(ctx sdk.Ctx, currentApp types.Platform, amount sdk.BigInt) sdk.Error {
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

// StakePlatform - Store ops when a application stakes
func (k Keeper) StakePlatform(ctx sdk.Ctx, application types.Platform, amount sdk.BigInt) sdk.Error {
	// edit stake
	if ctx.IsAfterUpgradeHeight() {
		// get Validator to see if edit stake
		curApp, found := k.GetPlatform(ctx, application.Address)
		if found && curApp.IsStaked() {
			return k.EditStakePlatform(ctx, curApp, application, amount)
		}
	}
	// send the coins from address to staked module account
	err := k.coinsFromUnstakedToStaked(ctx, application, amount)
	if err != nil {
		return err
	}
	// add coins to the staked field
	application, er := application.AddStakedTokens(amount)
	if er != nil {
		return sdk.ErrInternal(er.Error())
	}
	// calculate relays
	application.MaxRelays = k.CalculatePlatformRelays(ctx, application)
	// set the status to staked
	application = application.UpdateStatus(sdk.Staked)
	// save in the application store
	k.SetPlatform(ctx, application)
	return nil
}

func (k Keeper) EditStakePlatform(ctx sdk.Ctx, application, updatedPlatform types.Platform, amount sdk.BigInt) sdk.Error {
	origAppForDeletion := application
	// get the difference in coins
	diff := amount.Sub(application.StakedTokens)
	// if they bumped the stake amount
	if diff.IsPositive() {
		// send the coins from address to staked module account
		err := k.coinsFromUnstakedToStaked(ctx, application, diff)
		if err != nil {
			return err
		}
		var er error
		// add coins to the staked field
		application, er = application.AddStakedTokens(diff)
		if er != nil {
			return sdk.ErrInternal(er.Error())
		}
		// update apps max relays
		application.MaxRelays = k.CalculatePlatformRelays(ctx, application)
	}
	// update chains
	application.Chains = updatedPlatform.Chains
	// delete the validator from the staking set
	k.deletePlatformFromStakingSet(ctx, origAppForDeletion)
	// delete in main store
	k.DeletePlatform(ctx, origAppForDeletion.Address)
	// save in the app store
	k.SetPlatform(ctx, application)
	// save the app by chains
	k.SetStakedPlatform(ctx, application)
	// clear session cache
	k.ViperKeeper.ClearSessionCache()
	// log success
	ctx.Logger().Info("Successfully updated staked application: " + application.Address.String())
	return nil
}

// ValidatePlatformBeginUnstaking - Check for validator status
func (k Keeper) ValidatePlatformBeginUnstaking(ctx sdk.Ctx, application types.Platform) sdk.Error {
	// must be staked to begin unstaking
	if !application.IsStaked() {
		return sdk.ErrInternal(types.ErrPlatformStatus(k.codespace).Error())
	}
	if application.IsJailed() {
		return sdk.ErrInternal(types.ErrPlatformJailed(k.codespace).Error())
	}
	return nil
}

// BeginUnstakingPlatform - Store ops when application begins to unstake -> starts the unstaking timer
func (k Keeper) BeginUnstakingPlatform(ctx sdk.Ctx, application types.Platform) {
	// get params
	params := k.GetParams(ctx)
	// delete the application from the staking set, as it is technically staked but not going to participate
	k.deletePlatformFromStakingSet(ctx, application)
	// set the status
	application = application.UpdateStatus(sdk.Unstaking)
	// set the unstaking completion time and completion height appropriately
	if application.UnstakingCompletionTime.IsZero() {
		application.UnstakingCompletionTime = ctx.BlockHeader().Time.Add(params.UnstakingTime)
	}
	// save the now unstaked application record and power index
	k.SetPlatform(ctx, application)
	ctx.Logger().Info("Began unstaking App " + application.Address.String())
}

// ValidatePlatformFinishUnstaking - Check if application can finish unstaking
func (k Keeper) ValidatePlatformFinishUnstaking(ctx sdk.Ctx, application types.Platform) sdk.Error {
	if !application.IsUnstaking() {
		return types.ErrPlatformStatus(k.codespace)
	}
	if application.IsJailed() {
		return types.ErrPlatformJailed(k.codespace)
	}
	return nil
}

// FinishUnstakingPlatform - Store ops to unstake a application -> called after unstaking time is up
func (k Keeper) FinishUnstakingPlatform(ctx sdk.Ctx, application types.Platform) {
	// delete the application from the unstaking queue
	k.deleteUnstakingPlatform(ctx, application)
	// amount unstaked = stakedTokens
	amount := application.StakedTokens
	// send the tokens from staking module account to application account
	err := k.coinsFromStakedToUnstaked(ctx, application)
	if err != nil {
		k.Logger(ctx).Error("could not move coins from staked to unstaked for applications module" + err.Error() + "for this app address: " + application.Address.String())
		// continue with the unstaking
	}
	// removed the staked tokens field from application structure
	application, er := application.RemoveStakedTokens(amount)
	if er != nil {
		k.Logger(ctx).Error("could not remove tokens from unstaking application: " + er.Error())
		// continue with the unstaking
	}
	// update the status to unstaked
	application = application.UpdateStatus(sdk.Unstaked)
	// reset app relays
	application.MaxRelays = sdk.ZeroInt()
	// update the unstaking time
	application.UnstakingCompletionTime = time.Time{}
	// update the application in the main store
	k.SetPlatform(ctx, application)
	ctx.Logger().Info("Finished unstaking application " + application.Address.String())
	// create the event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeUnstake,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, application.Address.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, application.Address.String()),
		),
	})
}

// LegacyForcePlatformUnstake - Coerce unstake (called when slashed below the minimum)
func (k Keeper) LegacyForcePlatformUnstake(ctx sdk.Ctx, application types.Platform) sdk.Error {
	// delete the application from staking set as they are unstaked
	k.deletePlatformFromStakingSet(ctx, application)
	// amount unstaked = stakedTokens
	err := k.burnStakedTokens(ctx, application.StakedTokens)
	if err != nil {
		return err
	}
	// remove their tokens from the field
	application, er := application.RemoveStakedTokens(application.StakedTokens)
	if er != nil {
		return sdk.ErrInternal(er.Error())
	}
	// update their status to unstaked
	application = application.UpdateStatus(sdk.Unstaked)
	// reset app relays
	application.MaxRelays = sdk.ZeroInt()
	// set the application in store
	k.SetPlatform(ctx, application)
	ctx.Logger().Info("Force Unstaked application " + application.Address.String())
	// create the event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeUnstake,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, application.Address.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, application.Address.String()),
		),
	})
	return nil
}

// ForceValidatorUnstake - Coerce unstake (called when slashed below the minimum)
func (k Keeper) ForcePlatformUnstake(ctx sdk.Ctx, application types.Platform) sdk.Error {
	if !ctx.IsAfterUpgradeHeight() {
		return k.LegacyForcePlatformUnstake(ctx, application)
	}
	switch application.Status {
	case sdk.Staked:
		k.deletePlatformFromStakingSet(ctx, application)
	case sdk.Unstaking:
		k.deleteUnstakingPlatform(ctx, application)
		k.DeletePlatform(ctx, application.Address)
	default:
		k.DeletePlatform(ctx, application.Address)
		return sdk.ErrInternal("should not happen: trying to force unstake an already unstaked application: " + application.Address.String())
	}
	// amount unstaked = stakedTokens
	err := k.burnStakedTokens(ctx, application.StakedTokens)
	if err != nil {
		return err
	}
	if application.IsStaked() {
		// remove their tokens from the field
		validator, er := application.RemoveStakedTokens(application.StakedTokens)
		if er != nil {
			return sdk.ErrInternal(er.Error())
		}
		// update their status to unstaked
		validator = validator.UpdateStatus(sdk.Unstaked)
		// set the validator in store
		k.SetPlatform(ctx, validator)
	}
	ctx.Logger().Info("Force Unstaked validator " + application.Address.String())
	return nil
}

// JailPlatform - Send a application to jail
func (k Keeper) JailPlatform(ctx sdk.Ctx, addr sdk.Address) {
	application, found := k.GetPlatform(ctx, addr)
	if !found {
		k.Logger(ctx).Error(fmt.Errorf("application %s is attempted jailed but not found in all applications store", addr).Error())
		return
	}
	if application.Jailed {
		k.Logger(ctx).Error(fmt.Sprintf("cannot jail already jailed application, application: %v\n", application))
		return
	}
	application.Jailed = true
	k.SetPlatform(ctx, application)
	k.deletePlatformFromStakingSet(ctx, application)
	logger := k.Logger(ctx)
	logger.Info(fmt.Sprintf("application %s jailed", addr))
}

func (k Keeper) IncrementJailedPlatforms(ctx sdk.Ctx) {
	// TODO
}

// ValidateUnjailMessage - Check unjail message
func (k Keeper) ValidateUnjailMessage(ctx sdk.Ctx, msg types.MsgUnjail) (addr sdk.Address, err sdk.Error) {
	application, found := k.GetPlatform(ctx, msg.PlatformAddr)
	if !found {
		return nil, types.ErrNoPlatformForAddress(k.Codespace())
	}
	// cannot be unjailed if not staked
	stake := application.GetTokens()
	if stake == sdk.ZeroInt() {
		return nil, types.ErrMissingPlatformStake(k.Codespace())
	}
	if application.GetTokens().LT(sdk.NewInt(k.MinimumStake(ctx))) { // TODO look into this state change (stuck in jail)
		return nil, types.ErrStakeTooLow(k.Codespace())
	}
	// cannot be unjailed if not jailed
	if !application.IsJailed() {
		return nil, types.ErrPlatformNotJailed(k.Codespace())
	}
	return
}

// UnjailPlatform - Remove a application from jail
func (k Keeper) UnjailPlatform(ctx sdk.Ctx, addr sdk.Address) {
	application, found := k.GetPlatform(ctx, addr)
	if !found {
		k.Logger(ctx).Error(fmt.Errorf("application %s is attempted jailed but not found in all applications store", addr).Error())
		return
	}
	if !application.Jailed {
		k.Logger(ctx).Error(fmt.Sprintf("cannot unjail already unjailed application, application: %v\n", application))
		return
	}
	application.Jailed = false
	k.SetPlatform(ctx, application)
	logger := k.Logger(ctx)
	logger.Info(fmt.Sprintf("application %s unjailed", addr))
}
