package keeper

import (
	"fmt"
	"time"

	"github.com/tendermint/tendermint/libs/strings"

	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/providers/types"
)

// ValidateProviderStaking - Check provider before staking
func (k Keeper) ValidateProviderStaking(ctx sdk.Ctx, provider types.Provider, amount sdk.BigInt) sdk.Error {
	// convert the amount to sdk.Coin
	coin := sdk.NewCoins(sdk.NewCoin(k.StakeDenom(ctx), amount))
	if int64(len(provider.Chains)) > k.MaxChains(ctx) {
		return types.ErrTooManyChains(types.ModuleName)
	}
	// attempt to get the provider from the world state
	app, found := k.GetProvider(ctx, provider.Address)
	// if the provider exists
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
			tmPubKey, err := crypto.CheckConsensusPubKey(provider.PublicKey.PubKey())
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
	if !k.AccountKeeper.HasCoins(ctx, provider.Address, coin) {
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

// StakeProvider - Store ops when a provider stakes
func (k Keeper) StakeProvider(ctx sdk.Ctx, provider types.Provider, amount sdk.BigInt) sdk.Error {
	// edit stake
	if ctx.IsAfterUpgradeHeight() {
		// get Validator to see if edit stake
		curApp, found := k.GetProvider(ctx, provider.Address)
		if found && curApp.IsStaked() {
			return k.EditStakeProvider(ctx, curApp, provider, amount)
		}
	}
	// send the coins from address to staked module account
	err := k.coinsFromUnstakedToStaked(ctx, provider, amount)
	if err != nil {
		return err
	}
	// add coins to the staked field
	provider, er := provider.AddStakedTokens(amount)
	if er != nil {
		return sdk.ErrInternal(er.Error())
	}
	// calculate relays
	provider.MaxRelays = k.CalculateProviderRelays(ctx, provider)
	// set the status to staked
	provider = provider.UpdateStatus(sdk.Staked)
	// save in the provider store
	k.SetProvider(ctx, provider)
	return nil
}

func (k Keeper) EditStakeProvider(ctx sdk.Ctx, provider, updatedProvider types.Provider, amount sdk.BigInt) sdk.Error {
	origAppForDeletion := provider
	// get the difference in coins
	diff := amount.Sub(provider.StakedTokens)
	// if they bumped the stake amount
	if diff.IsPositive() {
		// send the coins from address to staked module account
		err := k.coinsFromUnstakedToStaked(ctx, provider, diff)
		if err != nil {
			return err
		}
		var er error
		// add coins to the staked field
		provider, er = provider.AddStakedTokens(diff)
		if er != nil {
			return sdk.ErrInternal(er.Error())
		}
		// update apps max relays
		provider.MaxRelays = k.CalculateProviderRelays(ctx, provider)
	}
	// update chains
	provider.Chains = updatedProvider.Chains
	// delete the validator from the staking set
	k.deleteProviderFromStakingSet(ctx, origAppForDeletion)
	// delete in main store
	k.DeleteProvider(ctx, origAppForDeletion.Address)
	// save in the app store
	k.SetProvider(ctx, provider)
	// save the app by chains
	k.SetStakedProvider(ctx, provider)
	// clear session cache
	k.ViperKeeper.ClearSessionCache()
	// log success
	ctx.Logger().Info("Successfully updated staked provider: " + provider.Address.String())
	return nil
}

// ValidateProviderBeginUnstaking - Check for validator status
func (k Keeper) ValidateProviderBeginUnstaking(ctx sdk.Ctx, provider types.Provider) sdk.Error {
	// must be staked to begin unstaking
	if !provider.IsStaked() {
		return sdk.ErrInternal(types.ErrProviderStatus(k.codespace).Error())
	}
	if provider.IsJailed() {
		return sdk.ErrInternal(types.ErrProviderJailed(k.codespace).Error())
	}
	return nil
}

// BeginUnstakingProvider - Store ops when provider begins to unstake -> starts the unstaking timer
func (k Keeper) BeginUnstakingProvider(ctx sdk.Ctx, provider types.Provider) {
	// get params
	params := k.GetParams(ctx)
	// delete the provider from the staking set, as it is technically staked but not going to participate
	k.deleteProviderFromStakingSet(ctx, provider)
	// set the status
	provider = provider.UpdateStatus(sdk.Unstaking)
	// set the unstaking completion time and completion height appropriately
	if provider.UnstakingCompletionTime.IsZero() {
		provider.UnstakingCompletionTime = ctx.BlockHeader().Time.Add(params.UnstakingTime)
	}
	// save the now unstaked provider record and power index
	k.SetProvider(ctx, provider)
	ctx.Logger().Info("Began unstaking App " + provider.Address.String())
}

// ValidateProviderFinishUnstaking - Check if provider can finish unstaking
func (k Keeper) ValidateProviderFinishUnstaking(ctx sdk.Ctx, provider types.Provider) sdk.Error {
	if !provider.IsUnstaking() {
		return types.ErrProviderStatus(k.codespace)
	}
	if provider.IsJailed() {
		return types.ErrProviderJailed(k.codespace)
	}
	return nil
}

// FinishUnstakingProvider - Store ops to unstake a client -> called after unstaking time is up
func (k Keeper) FinishUnstakingProvider(ctx sdk.Ctx, provider types.Provider) {
	// delete the provider from the unstaking queue
	k.deleteUnstakingProvider(ctx, provider)
	// amount unstaked = stakedTokens
	amount := provider.StakedTokens
	// send the tokens from staking module account to provider account
	err := k.coinsFromStakedToUnstaked(ctx, provider)
	if err != nil {
		k.Logger(ctx).Error("could not move coins from staked to unstaked for applications module" + err.Error() + "for this app address: " + provider.Address.String())
		// continue with the unstaking
	}
	// removed the staked tokens field from provider structure
	provider, er := provider.RemoveStakedTokens(amount)
	if er != nil {
		k.Logger(ctx).Error("could not remove tokens from unstaking provider: " + er.Error())
		// continue with the unstaking
	}
	// update the status to unstaked
	provider = provider.UpdateStatus(sdk.Unstaked)
	// reset app relays
	provider.MaxRelays = sdk.ZeroInt()
	// update the unstaking time
	provider.UnstakingCompletionTime = time.Time{}
	// update the provider in the main store
	k.SetProvider(ctx, provider)
	ctx.Logger().Info("Finished unstaking provider " + provider.Address.String())
	// create the event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeUnstake,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, provider.Address.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, provider.Address.String()),
		),
	})
}

// LegacyForceProviderUnstake - Coerce unstake (called when slashed below the minimum)
func (k Keeper) LegacyForceProviderUnstake(ctx sdk.Ctx, provider types.Provider) sdk.Error {
	// delete the provider from staking set as they are unstaked
	k.deleteProviderFromStakingSet(ctx, provider)
	// amount unstaked = stakedTokens
	err := k.burnStakedTokens(ctx, provider.StakedTokens)
	if err != nil {
		return err
	}
	// remove their tokens from the field
	provider, er := provider.RemoveStakedTokens(provider.StakedTokens)
	if er != nil {
		return sdk.ErrInternal(er.Error())
	}
	// update their status to unstaked
	provider = provider.UpdateStatus(sdk.Unstaked)
	// reset app relays
	provider.MaxRelays = sdk.ZeroInt()
	// set the provider in store
	k.SetProvider(ctx, provider)
	ctx.Logger().Info("Force Unstaked provider " + provider.Address.String())
	// create the event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeUnstake,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, provider.Address.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, provider.Address.String()),
		),
	})
	return nil
}

// ForceValidatorUnstake - Coerce unstake (called when slashed below the minimum)
func (k Keeper) ForceProviderUnstake(ctx sdk.Ctx, provider types.Provider) sdk.Error {
	if !ctx.IsAfterUpgradeHeight() {
		return k.LegacyForceProviderUnstake(ctx, provider)
	}
	switch provider.Status {
	case sdk.Staked:
		k.deleteProviderFromStakingSet(ctx, provider)
	case sdk.Unstaking:
		k.deleteUnstakingProvider(ctx, provider)
		k.DeleteProvider(ctx, provider.Address)
	default:
		k.DeleteProvider(ctx, provider.Address)
		return sdk.ErrInternal("should not happen: trying to force unstake an already unstaked provider: " + provider.Address.String())
	}
	// amount unstaked = stakedTokens
	err := k.burnStakedTokens(ctx, provider.StakedTokens)
	if err != nil {
		return err
	}
	if provider.IsStaked() {
		// remove their tokens from the field
		validator, er := provider.RemoveStakedTokens(provider.StakedTokens)
		if er != nil {
			return sdk.ErrInternal(er.Error())
		}
		// update their status to unstaked
		validator = validator.UpdateStatus(sdk.Unstaked)
		// set the validator in store
		k.SetProvider(ctx, validator)
	}
	ctx.Logger().Info("Force Unstaked validator " + provider.Address.String())
	return nil
}

// JailProvider - Send a provider to jail
func (k Keeper) JailProvider(ctx sdk.Ctx, addr sdk.Address) {
	provider, found := k.GetProvider(ctx, addr)
	if !found {
		k.Logger(ctx).Error(fmt.Errorf("provider %s is attempted jailed but not found in all applications store", addr).Error())
		return
	}
	if provider.Jailed {
		k.Logger(ctx).Error(fmt.Sprintf("cannot jail already jailed provider, provider: %v\n", provider))
		return
	}
	provider.Jailed = true
	k.SetProvider(ctx, provider)
	k.deleteProviderFromStakingSet(ctx, provider)
	logger := k.Logger(ctx)
	logger.Info(fmt.Sprintf("provider %s jailed", addr))
}

func (k Keeper) IncrementJailedProviders(ctx sdk.Ctx) {
	// TODO
}

// ValidateUnjailMessage - Check unjail message
func (k Keeper) ValidateUnjailMessage(ctx sdk.Ctx, msg types.MsgUnjail) (addr sdk.Address, err sdk.Error) {
	provider, found := k.GetProvider(ctx, msg.ProviderAddr)
	if !found {
		return nil, types.ErrNoProviderForAddress(k.Codespace())
	}
	// cannot be unjailed if not staked
	stake := provider.GetTokens()
	if stake == sdk.ZeroInt() {
		return nil, types.ErrMissingProviderStake(k.Codespace())
	}
	if provider.GetTokens().LT(sdk.NewInt(k.MinimumStake(ctx))) { // TODO look into this state change (stuck in jail)
		return nil, types.ErrStakeTooLow(k.Codespace())
	}
	// cannot be unjailed if not jailed
	if !provider.IsJailed() {
		return nil, types.ErrProviderNotJailed(k.Codespace())
	}
	return
}

// UnjailProvider - Remove a provider from jail
func (k Keeper) UnjailProvider(ctx sdk.Ctx, addr sdk.Address) {
	provider, found := k.GetProvider(ctx, addr)
	if !found {
		k.Logger(ctx).Error(fmt.Errorf("provider %s is attempted jailed but not found in all applications store", addr).Error())
		return
	}
	if !provider.Jailed {
		k.Logger(ctx).Error(fmt.Sprintf("cannot unjail already unjailed provider, provider: %v\n", provider))
		return
	}
	provider.Jailed = false
	k.SetProvider(ctx, provider)
	logger := k.Logger(ctx)
	logger.Info(fmt.Sprintf("provider %s unjailed", addr))
}
