package keeper

import (
	"fmt"
	"time"

	"github.com/tendermint/tendermint/libs/strings"

	"github.com/vipernet-xyz/viper-network/crypto"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/platforms/types"
)

// ValidatePlatformStaking - Check platform before staking
func (k Keeper) ValidatePlatformStaking(ctx sdk.Ctx, platform types.Platform, amount sdk.BigInt) sdk.Error {
	// convert the amount to sdk.Coin
	coin := sdk.NewCoins(sdk.NewCoin(k.StakeDenom(ctx), amount))
	if int64(len(platform.Chains)) > k.MaxChains(ctx) {
		return types.ErrTooManyChains(types.ModuleName)
	}
	// attempt to get the platform from the world state
	platform, found := k.GetPlatform(ctx, platform.Address)
	// if the platform exists
	if found {
		// edit stake in 6.X upgrade
		if ctx.IsAfterUpgradeHeight() && platform.IsStaked() {
			return k.ValidateEditStake(ctx, platform, amount)
		}
		if !platform.IsUnstaked() { // unstaking or already staked but before the upgrade
			return types.ErrPlatformStatus(k.codespace)
		}
	} else {
		// ensure public key type is supported
		if ctx.ConsensusParams() != nil {
			tmPubKey, err := crypto.CheckConsensusPubKey(platform.PublicKey.PubKey())
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
	if !k.AccountKeeper.HasCoins(ctx, platform.Address, coin) {
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
func (k Keeper) ValidateEditStake(ctx sdk.Ctx, currentPlatform types.Platform, amount sdk.BigInt) sdk.Error {
	// ensure not staking less
	diff := amount.Sub(currentPlatform.StakedTokens)
	if diff.IsNegative() {
		return types.ErrMinimumEditStake(k.codespace)
	}
	// if stake bump
	if !diff.IsZero() {
		// ensure account has enough coins for bump
		coin := sdk.NewCoins(sdk.NewCoin(k.StakeDenom(ctx), diff))
		if !k.AccountKeeper.HasCoins(ctx, currentPlatform.Address, coin) {
			return types.ErrNotEnoughCoins(k.Codespace())
		}
	}
	return nil
}

// StakePlatform - Store ops when a platform stakes
func (k Keeper) StakePlatform(ctx sdk.Ctx, platform types.Platform, amount sdk.BigInt) sdk.Error {
	// edit stake
	if ctx.IsAfterUpgradeHeight() {
		// get Validator to see if edit stake
		curPlatform, found := k.GetPlatform(ctx, platform.Address)
		if found && curPlatform.IsStaked() {
			return k.EditStakePlatform(ctx, curPlatform, platform, amount)
		}
	}
	// send the coins from address to staked module account
	err := k.coinsFromUnstakedToStaked(ctx, platform, amount)
	if err != nil {
		return err
	}
	// add coins to the staked field
	platform, er := platform.AddStakedTokens(amount)
	if er != nil {
		return sdk.ErrInternal(er.Error())
	}
	// calculate relays
	platform.MaxRelays = k.CalculatePlatformRelays(ctx, platform)
	// set the status to staked
	platform = platform.UpdateStatus(sdk.Staked)
	// save in the platform store
	k.SetPlatform(ctx, platform)
	return nil
}

func (k Keeper) EditStakePlatform(ctx sdk.Ctx, platform, updatedPlatform types.Platform, amount sdk.BigInt) sdk.Error {
	origPlatformForDeletion := platform
	// get the difference in coins
	diff := amount.Sub(platform.StakedTokens)
	// if they bumped the stake amount
	if diff.IsPositive() {
		// send the coins from address to staked module account
		err := k.coinsFromUnstakedToStaked(ctx, platform, diff)
		if err != nil {
			return err
		}
		var er error
		// add coins to the staked field
		platform, er = platform.AddStakedTokens(diff)
		if er != nil {
			return sdk.ErrInternal(er.Error())
		}
		// update platforms max relays
		platform.MaxRelays = k.CalculatePlatformRelays(ctx, platform)
	}
	// update chains
	platform.Chains = updatedPlatform.Chains
	// delete the validator from the staking set
	k.deletePlatformFromStakingSet(ctx, origPlatformForDeletion)
	// delete in main store
	k.DeletePlatform(ctx, origPlatformForDeletion.Address)
	// save in the platform store
	k.SetPlatform(ctx, platform)
	// save the platform by chains
	k.SetStakedPlatform(ctx, platform)
	// clear session cache
	k.ViperKeeper.ClearSessionCache()
	// log success
	ctx.Logger().Info("Successfully updated staked platform: " + platform.Address.String())
	return nil
}

// ValidatePlatformBeginUnstaking - Check for validator status
func (k Keeper) ValidatePlatformBeginUnstaking(ctx sdk.Ctx, platform types.Platform) sdk.Error {
	// must be staked to begin unstaking
	if !platform.IsStaked() {
		return sdk.ErrInternal(types.ErrPlatformStatus(k.codespace).Error())
	}
	if platform.IsJailed() {
		return sdk.ErrInternal(types.ErrPlatformJailed(k.codespace).Error())
	}
	return nil
}

// BeginUnstakingPlatform - Store ops when platform begins to unstake -> starts the unstaking timer
func (k Keeper) BeginUnstakingPlatform(ctx sdk.Ctx, platform types.Platform) {
	// get params
	params := k.GetParams(ctx)
	// delete the platform from the staking set, as it is technically staked but not going to participate
	k.deletePlatformFromStakingSet(ctx, platform)
	// set the status
	platform = platform.UpdateStatus(sdk.Unstaking)
	// set the unstaking completion time and completion height platformropriately
	if platform.UnstakingCompletionTime.IsZero() {
		platform.UnstakingCompletionTime = ctx.BlockHeader().Time.Add(params.UnstakingTime)
	}
	// save the now unstaked platform record and power index
	k.SetPlatform(ctx, platform)
	ctx.Logger().Info("Began unstaking Platform " + platform.Address.String())
}

// ValidatePlatformFinishUnstaking - Check if platform can finish unstaking
func (k Keeper) ValidatePlatformFinishUnstaking(ctx sdk.Ctx, platform types.Platform) sdk.Error {
	if !platform.IsUnstaking() {
		return types.ErrPlatformStatus(k.codespace)
	}
	if platform.IsJailed() {
		return types.ErrPlatformJailed(k.codespace)
	}
	return nil
}

// FinishUnstakingPlatform - Store ops to unstake a platform -> called after unstaking time is up
func (k Keeper) FinishUnstakingPlatform(ctx sdk.Ctx, platform types.Platform) {
	// delete the platform from the unstaking queue
	k.deleteUnstakingPlatform(ctx, platform)
	// amount unstaked = stakedTokens
	amount := platform.StakedTokens
	// send the tokens from staking module account to platform account
	err := k.coinsFromStakedToUnstaked(ctx, platform)
	if err != nil {
		k.Logger(ctx).Error("could not move coins from staked to unstaked for platforms module" + err.Error() + "for this platform address: " + platform.Address.String())
		// continue with the unstaking
	}
	// removed the staked tokens field from platform structure
	platform, er := platform.RemoveStakedTokens(amount)
	if er != nil {
		k.Logger(ctx).Error("could not remove tokens from unstaking platform: " + er.Error())
		// continue with the unstaking
	}
	// update the status to unstaked
	platform = platform.UpdateStatus(sdk.Unstaked)
	// reset platform relays
	platform.MaxRelays = sdk.ZeroInt()
	// update the unstaking time
	platform.UnstakingCompletionTime = time.Time{}
	// update the platform in the main store
	k.SetPlatform(ctx, platform)
	ctx.Logger().Info("Finished unstaking platform " + platform.Address.String())
	// create the event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeUnstake,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, platform.Address.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, platform.Address.String()),
		),
	})
}

// LegacyForcePlatformUnstake - Coerce unstake (called when slashed below the minimum)
func (k Keeper) LegacyForcePlatformUnstake(ctx sdk.Ctx, platform types.Platform) sdk.Error {
	// delete the platform from staking set as they are unstaked
	k.deletePlatformFromStakingSet(ctx, platform)
	// amount unstaked = stakedTokens
	err := k.burnStakedTokens(ctx, platform.StakedTokens)
	if err != nil {
		return err
	}
	// remove their tokens from the field
	platform, er := platform.RemoveStakedTokens(platform.StakedTokens)
	if er != nil {
		return sdk.ErrInternal(er.Error())
	}
	// update their status to unstaked
	platform = platform.UpdateStatus(sdk.Unstaked)
	// reset platform relays
	platform.MaxRelays = sdk.ZeroInt()
	// set the platform in store
	k.SetPlatform(ctx, platform)
	ctx.Logger().Info("Force Unstaked platform " + platform.Address.String())
	// create the event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeUnstake,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, platform.Address.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, platform.Address.String()),
		),
	})
	return nil
}

// ForceValidatorUnstake - Coerce unstake (called when slashed below the minimum)
func (k Keeper) ForcePlatformUnstake(ctx sdk.Ctx, platform types.Platform) sdk.Error {
	if !ctx.IsAfterUpgradeHeight() {
		return k.LegacyForcePlatformUnstake(ctx, platform)
	}
	switch platform.Status {
	case sdk.Staked:
		k.deletePlatformFromStakingSet(ctx, platform)
	case sdk.Unstaking:
		k.deleteUnstakingPlatform(ctx, platform)
		k.DeletePlatform(ctx, platform.Address)
	default:
		k.DeletePlatform(ctx, platform.Address)
		return sdk.ErrInternal("should not hplatformen: trying to force unstake an already unstaked platform: " + platform.Address.String())
	}
	// amount unstaked = stakedTokens
	err := k.burnStakedTokens(ctx, platform.StakedTokens)
	if err != nil {
		return err
	}
	if platform.IsStaked() {
		// remove their tokens from the field
		validator, er := platform.RemoveStakedTokens(platform.StakedTokens)
		if er != nil {
			return sdk.ErrInternal(er.Error())
		}
		// update their status to unstaked
		validator = validator.UpdateStatus(sdk.Unstaked)
		// set the validator in store
		k.SetPlatform(ctx, validator)
	}
	ctx.Logger().Info("Force Unstaked validator " + platform.Address.String())
	return nil
}

// JailPlatform - Send a platform to jail
func (k Keeper) JailPlatform(ctx sdk.Ctx, addr sdk.Address) {
	platform, found := k.GetPlatform(ctx, addr)
	if !found {
		k.Logger(ctx).Error(fmt.Errorf("platform %s is attempted jailed but not found in all platforms store", addr).Error())
		return
	}
	if platform.Jailed {
		k.Logger(ctx).Error(fmt.Sprintf("cannot jail already jailed platform, platform: %v\n", platform))
		return
	}
	platform.Jailed = true
	k.SetPlatform(ctx, platform)
	k.deletePlatformFromStakingSet(ctx, platform)
	logger := k.Logger(ctx)
	logger.Info(fmt.Sprintf("platform %s jailed", addr))
}

func (k Keeper) IncrementJailedPlatforms(ctx sdk.Ctx) {
	// TODO
}

// ValidateUnjailMessage - Check unjail message
func (k Keeper) ValidateUnjailMessage(ctx sdk.Ctx, msg types.MsgUnjail) (addr sdk.Address, err sdk.Error) {
	platform, found := k.GetPlatform(ctx, msg.PlatformAddr)
	if !found {
		return nil, types.ErrNoPlatformForAddress(k.Codespace())
	}
	// cannot be unjailed if not staked
	stake := platform.GetTokens()
	if stake == sdk.ZeroInt() {
		return nil, types.ErrMissingPlatformStake(k.Codespace())
	}
	if platform.GetTokens().LT(sdk.NewInt(k.MinimumStake(ctx))) { // TODO look into this state change (stuck in jail)
		return nil, types.ErrStakeTooLow(k.Codespace())
	}
	// cannot be unjailed if not jailed
	if !platform.IsJailed() {
		return nil, types.ErrPlatformNotJailed(k.Codespace())
	}
	return
}

// UnjailPlatform - Remove a platform from jail
func (k Keeper) UnjailPlatform(ctx sdk.Ctx, addr sdk.Address) {
	platform, found := k.GetPlatform(ctx, addr)
	if !found {
		k.Logger(ctx).Error(fmt.Errorf("platform %s is attempted jailed but not found in all platforms store", addr).Error())
		return
	}
	if !platform.Jailed {
		k.Logger(ctx).Error(fmt.Sprintf("cannot unjail already unjailed platform, platform: %v\n", platform))
		return
	}
	platform.Jailed = false
	k.SetPlatform(ctx, platform)
	logger := k.Logger(ctx)
	logger.Info(fmt.Sprintf("platform %s unjailed", addr))
}
