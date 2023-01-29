package keeper

import (
	"fmt"
	"time"

	"github.com/tendermint/tendermint/libs/strings"

	"github.com/vipernet-xyz/viper-network/crypto"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/providers/types"
)

// ValidateProviderStaking - Check application before staking
func (k Keeper) ValidateProviderStaking(ctx sdk.Ctx, application types.Provider, amount sdk.BigInt) sdk.Error {
	// convert the amount to sdk.Coin
	coin := sdk.NewCoins(sdk.NewCoin(k.StakeDenom(ctx), amount))
	if int64(len(application.Chains)) > k.MaxChains(ctx) {
		return types.ErrTooManyChains(types.ModuleName)
	}
	// attempt to get the application from the world state
	app, found := k.GetProvider(ctx, application.Address)
	// if the application exists
	if found {
		// edit stake in 6.X upgrade
		if ctx.IsAfterUpgradeHeight() && app.IsStaked() {
			return k.ValidateEditStake(ctx, app, amount)
		}
		if !app.IsUnstaked() { // unstaking or already staked but before the upgrade
			return types.ErrProviderStatus(k.codespace)
		}
	} else {
		// ensure public key type is supported
		if ctx.ConsensusParams() != nil {
			tmPubKey, err := crypto.CheckConsensusPubKey(application.PublicKey.PubKey())
			if err != nil {
				return types.ErrProviderPubKeyTypeNotSupported(k.Codespace(),
					err.Error(),
					ctx.ConsensusParams().Validator.PubKeyTypes)
			}
			if !strings.StringInSlice(tmPubKey.Type, ctx.ConsensusParams().Validator.PubKeyTypes) {
				return types.ErrProviderPubKeyTypeNotSupported(k.Codespace(),
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
		if k.getStakedProvidersCount(ctx) >= k.MaxProviders(ctx) {
			return types.ErrMaxProviders(k.codespace)
		}
	}
	return nil
}

// ValidateEditStake - Validate the updates to a current staked validator
func (k Keeper) ValidateEditStake(ctx sdk.Ctx, currentApp types.Provider, amount sdk.BigInt) sdk.Error {
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

// StakeProvider - Store ops when a application stakes
func (k Keeper) StakeProvider(ctx sdk.Ctx, application types.Provider, amount sdk.BigInt) sdk.Error {
	// edit stake
	if ctx.IsAfterUpgradeHeight() {
		// get Validator to see if edit stake
		curApp, found := k.GetProvider(ctx, application.Address)
		if found && curApp.IsStaked() {
			return k.EditStakeProvider(ctx, curApp, application, amount)
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
	application.MaxRelays = k.CalculateProviderRelays(ctx, application)
	// set the status to staked
	application = application.UpdateStatus(sdk.Staked)
	// save in the application store
	k.SetProvider(ctx, application)
	return nil
}

func (k Keeper) EditStakeProvider(ctx sdk.Ctx, application, updatedProvider types.Provider, amount sdk.BigInt) sdk.Error {
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
		application.MaxRelays = k.CalculateProviderRelays(ctx, application)
	}
	// update chains
	application.Chains = updatedProvider.Chains
	// delete the validator from the staking set
	k.deleteProviderFromStakingSet(ctx, origAppForDeletion)
	// delete in main store
	k.DeleteProvider(ctx, origAppForDeletion.Address)
	// save in the app store
	k.SetProvider(ctx, application)
	// save the app by chains
	k.SetStakedProvider(ctx, application)
	// clear session cache
	k.ViperKeeper.ClearSessionCache()
	// log success
	ctx.Logger().Info("Successfully updated staked application: " + application.Address.String())
	return nil
}

// ValidateProviderBeginUnstaking - Check for validator status
func (k Keeper) ValidateProviderBeginUnstaking(ctx sdk.Ctx, application types.Provider) sdk.Error {
	// must be staked to begin unstaking
	if !application.IsStaked() {
		return sdk.ErrInternal(types.ErrProviderStatus(k.codespace).Error())
	}
	if application.IsJailed() {
		return sdk.ErrInternal(types.ErrProviderJailed(k.codespace).Error())
	}
	return nil
}

// BeginUnstakingProvider - Store ops when application begins to unstake -> starts the unstaking timer
func (k Keeper) BeginUnstakingProvider(ctx sdk.Ctx, application types.Provider) {
	// get params
	params := k.GetParams(ctx)
	// delete the application from the staking set, as it is technically staked but not going to participate
	k.deleteProviderFromStakingSet(ctx, application)
	// set the status
	application = application.UpdateStatus(sdk.Unstaking)
	// set the unstaking completion time and completion height appropriately
	if application.UnstakingCompletionTime.IsZero() {
		application.UnstakingCompletionTime = ctx.BlockHeader().Time.Add(params.UnstakingTime)
	}
	// save the now unstaked application record and power index
	k.SetProvider(ctx, application)
	ctx.Logger().Info("Began unstaking App " + application.Address.String())
}

// ValidateProviderFinishUnstaking - Check if application can finish unstaking
func (k Keeper) ValidateProviderFinishUnstaking(ctx sdk.Ctx, application types.Provider) sdk.Error {
	if !application.IsUnstaking() {
		return types.ErrProviderStatus(k.codespace)
	}
	if application.IsJailed() {
		return types.ErrProviderJailed(k.codespace)
	}
	return nil
}

// FinishUnstakingProvider - Store ops to unstake a application -> called after unstaking time is up
func (k Keeper) FinishUnstakingProvider(ctx sdk.Ctx, application types.Provider) {
	// delete the application from the unstaking queue
	k.deleteUnstakingProvider(ctx, application)
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
	k.SetProvider(ctx, application)
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

// LegacyForceProviderUnstake - Coerce unstake (called when slashed below the minimum)
func (k Keeper) LegacyForceProviderUnstake(ctx sdk.Ctx, application types.Provider) sdk.Error {
	// delete the application from staking set as they are unstaked
	k.deleteProviderFromStakingSet(ctx, application)
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
	k.SetProvider(ctx, application)
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
func (k Keeper) ForceProviderUnstake(ctx sdk.Ctx, application types.Provider) sdk.Error {
	if !ctx.IsAfterUpgradeHeight() {
		return k.LegacyForceProviderUnstake(ctx, application)
	}
	switch application.Status {
	case sdk.Staked:
		k.deleteProviderFromStakingSet(ctx, application)
	case sdk.Unstaking:
		k.deleteUnstakingProvider(ctx, application)
		k.DeleteProvider(ctx, application.Address)
	default:
		k.DeleteProvider(ctx, application.Address)
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
		k.SetProvider(ctx, validator)
	}
	ctx.Logger().Info("Force Unstaked validator " + application.Address.String())
	return nil
}

// JailProvider - Send a application to jail
func (k Keeper) JailProvider(ctx sdk.Ctx, addr sdk.Address) {
	application, found := k.GetProvider(ctx, addr)
	if !found {
		k.Logger(ctx).Error(fmt.Errorf("application %s is attempted jailed but not found in all applications store", addr).Error())
		return
	}
	if application.Jailed {
		k.Logger(ctx).Error(fmt.Sprintf("cannot jail already jailed application, application: %v\n", application))
		return
	}
	application.Jailed = true
	k.SetProvider(ctx, application)
	k.deleteProviderFromStakingSet(ctx, application)
	logger := k.Logger(ctx)
	logger.Info(fmt.Sprintf("application %s jailed", addr))
}

func (k Keeper) IncrementJailedProviders(ctx sdk.Ctx) {
	// TODO
}

// ValidateUnjailMessage - Check unjail message
func (k Keeper) ValidateUnjailMessage(ctx sdk.Ctx, msg types.MsgUnjail) (addr sdk.Address, err sdk.Error) {
	application, found := k.GetProvider(ctx, msg.ProviderAddr)
	if !found {
		return nil, types.ErrNoProviderForAddress(k.Codespace())
	}
	// cannot be unjailed if not staked
	stake := application.GetTokens()
	if stake == sdk.ZeroInt() {
		return nil, types.ErrMissingProviderStake(k.Codespace())
	}
	if application.GetTokens().LT(sdk.NewInt(k.MinimumStake(ctx))) { // TODO look into this state change (stuck in jail)
		return nil, types.ErrStakeTooLow(k.Codespace())
	}
	// cannot be unjailed if not jailed
	if !application.IsJailed() {
		return nil, types.ErrProviderNotJailed(k.Codespace())
	}
	return
}

// UnjailProvider - Remove a application from jail
func (k Keeper) UnjailProvider(ctx sdk.Ctx, addr sdk.Address) {
	application, found := k.GetProvider(ctx, addr)
	if !found {
		k.Logger(ctx).Error(fmt.Errorf("application %s is attempted jailed but not found in all applications store", addr).Error())
		return
	}
	if !application.Jailed {
		k.Logger(ctx).Error(fmt.Sprintf("cannot unjail already unjailed application, application: %v\n", application))
		return
	}
	application.Jailed = false
	k.SetProvider(ctx, application)
	logger := k.Logger(ctx)
	logger.Info(fmt.Sprintf("application %s unjailed", addr))
}
