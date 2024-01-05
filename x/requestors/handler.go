package pos

import (
	"fmt"
	"reflect"

	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/requestors/keeper"
	"github.com/vipernet-xyz/viper-network/x/requestors/types"
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
	ctx.Logger().Info("Begin Staking Requestor Message received from " + sdk.Address(pk.Address()).String())
	// create requestor object using the message fields
	requestor := types.NewRequestor(sdk.Address(addr), pk, msg.Chains, sdk.ZeroInt(), msg.GeoZones, msg.NumServicers)
	ctx.Logger().Info("Validate Requestor Can Stake " + sdk.Address(addr).String())
	// check if they can stake
	if err := k.ValidateRequestorStaking(ctx, requestor, msg.Value); err != nil {
		ctx.Logger().Error(fmt.Sprintf("Validate Requestor Can Stake Error, at height: %d with address: %s", ctx.BlockHeight(), sdk.Address(addr).String()))
		return err.Result()
	}
	ctx.Logger().Info("Change Requestor state to Staked " + sdk.Address(addr).String())
	// change the requestor state to staked
	err := k.StakeRequestor(ctx, requestor, msg.Value)
	if err != nil {
		return err.Result()
	}
	// create the event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreateRequestor,
			sdk.NewAttribute(types.AttributeKeyRequestor, sdk.Address(addr).String()),
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
	requestor, found := k.GetRequestor(ctx, msg.Address)
	if !found {
		ctx.Logger().Error(fmt.Sprintf("Requestor Not Found at height: %d", ctx.BlockHeight()) + msg.Address.String())
		return types.ErrNoRequestorFound(k.Codespace()).Result()
	}
	if err := k.ValidateRequestorBeginUnstaking(ctx, requestor); err != nil {
		ctx.Logger().Error(fmt.Sprintf("Requestor Unstake Validation Not Successful, at height: %d", ctx.BlockHeight()) + msg.Address.String())
		return err.Result()
	}
	ctx.Logger().Info("Starting to Unstake Requestor " + msg.Address.String())
	k.BeginUnstakingRequestor(ctx, requestor)
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

// Requestors must submit a transaction to unjail itself after todo
// having been jailed (and thus unstaked) for downtime
func handleMsgUnjail(ctx sdk.Ctx, msg types.MsgUnjail, k keeper.Keeper) sdk.Result {
	consAddr, err := k.ValidateUnjailMessage(ctx, msg)
	if err != nil {
		return err.Result()
	}
	k.UnjailRequestor(ctx, consAddr)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.RequestorAddr.String()),
		),
	)
	return sdk.Result{Events: ctx.EventManager().Events()}
}
