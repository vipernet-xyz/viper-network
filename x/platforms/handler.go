package pos

import (
	"fmt"
	"reflect"

	"github.com/vipernet-xyz/viper-network/crypto"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/platforms/keeper"
	"github.com/vipernet-xyz/viper-network/x/platforms/types"
)

func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Ctx, msg sdk.Msg, _ crypto.PublicKey) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		// convert to value for switch consistency
		if reflect.ValueOf(msg).Kind() == reflect.Ptr {
			msg = reflect.Indirect(reflect.ValueOf(msg)).Interface().(sdk.Msg)
		}
		switch msg := msg.(type) {
		case types.MsgStake:
			return handleStake(ctx, msg, k)
		case types.MsgBeginUnstake:
			return handleMsgBeginUnstake(ctx, msg, k)
		case types.MsgUnjail:
			return handleMsgUnjail(ctx, msg, k)
		default:
			errMsg := fmt.Sprintf("unrecognized staking message type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleStake(ctx sdk.Ctx, msg types.MsgStake, k keeper.Keeper) sdk.Result {
	pk := msg.PubKey
	addr := pk.Address()
	ctx.Logger().Info("Begin Staking Platform Message received from " + sdk.Address(pk.Address()).String())
	// create platform object using the message fields
	platform := types.NewPlatform(sdk.Address(addr), pk, msg.Chains, sdk.ZeroInt())
	ctx.Logger().Info("Validate Platform Can Stake " + sdk.Address(addr).String())
	// check if they can stake
	if err := k.ValidatePlatformStaking(ctx, platform, msg.Value); err != nil {
		ctx.Logger().Error(fmt.Sprintf("Validate Platform Can Stake Error, at height: %d with address: %s", ctx.BlockHeight(), sdk.Address(addr).String()))
		return err.Result()
	}
	ctx.Logger().Info("Change Platform state to Staked " + sdk.Address(addr).String())
	// change the platform state to staked
	err := k.StakePlatform(ctx, platform, msg.Value)
	if err != nil {
		return err.Result()
	}
	// create the event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreatePlatform,
			sdk.NewAttribute(types.AttributeKeyPlatform, sdk.Address(addr).String()),
		),
		sdk.NewEvent(
			types.EventTypeStake,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, sdk.Address(addr).String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Value.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, sdk.Address(addr).String()),
		),
	})
	return sdk.Result{Events: ctx.EventManager().Events()}
}

func handleMsgBeginUnstake(ctx sdk.Ctx, msg types.MsgBeginUnstake, k keeper.Keeper) sdk.Result {
	platform, found := k.GetPlatform(ctx, msg.Address)
	if !found {
		ctx.Logger().Error(fmt.Sprintf("Platform Not Found at height: %d", ctx.BlockHeight()) + msg.Address.String())
		return types.ErrNoPlatformFound(k.Codespace()).Result()
	}
	if err := k.ValidatePlatformBeginUnstaking(ctx, platform); err != nil {
		ctx.Logger().Error(fmt.Sprintf("Platform Unstake Validation Not Successful, at height: %d", ctx.BlockHeight()) + msg.Address.String())
		return err.Result()
	}
	ctx.Logger().Info("Starting to Unstake Platform " + msg.Address.String())
	k.BeginUnstakingPlatform(ctx, platform)
	// create the event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeBeginUnstake,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Address.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Address.String()),
		),
	})
	return sdk.Result{Events: ctx.EventManager().Events()}
}

// Platforms must submit a transaction to unjail itself after todo
// having been jailed (and thus unstaked) for downtime
func handleMsgUnjail(ctx sdk.Ctx, msg types.MsgUnjail, k keeper.Keeper) sdk.Result {
	consAddr, err := k.ValidateUnjailMessage(ctx, msg)
	if err != nil {
		return err.Result()
	}
	k.UnjailPlatform(ctx, consAddr)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.PlatformAddr.String()),
		),
	)
	return sdk.Result{Events: ctx.EventManager().Events()}
}
