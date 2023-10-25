package servicers

import (
	"fmt"
	"reflect"
	"time"

	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/servicers/keeper"
	"github.com/vipernet-xyz/viper-network/x/servicers/types"
)

func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Ctx, msg sdk.Msg, signer crypto.PublicKey) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		// convert to value for switch consistency
		if reflect.ValueOf(msg).Kind() == reflect.Ptr {
			msg = reflect.Indirect(reflect.ValueOf(msg)).Interface().(sdk.Msg)
		}
		{
			switch msg := msg.(type) {
			case types.MsgBeginUnstake:
				return handleMsgBeginUnstake(ctx, msg, k)
			case types.MsgUnjail:
				return handleMsgUnjail(ctx, msg, k)
			case types.MsgSend:
				return handleMsgSend(ctx, msg, k)
			case types.MsgStake:
				return handleStake(ctx, msg, k, signer)
			case types.MsgPause:
				return handleMsgPause(ctx, msg, k)
			case types.MsgUnpause:
				return handleMsgUnpause(ctx, msg, k)
			default:
				errMsg := fmt.Sprintf("unrecognized staking message type: %T", msg)
				return sdk.ErrUnknownRequest(errMsg).Result()
			}
		}
	}
}

func handleStake(ctx sdk.Ctx, msg types.MsgStake, k keeper.Keeper, signer crypto.PublicKey) sdk.Result {
	defer sdk.TimeTrack(time.Now())

	err := msg.CheckServiceUrlLength(msg.ServiceUrl)
	if err != nil {
		return err.Result()
	}

	pk := msg.PublicKey
	addr := pk.Address()
	// create validator object using the message fields
	validator := types.NewValidator(sdk.Address(addr), pk, msg.Chains, msg.ServiceUrl, sdk.ZeroInt(), msg.GeoZone, msg.Output, types.ReportCard{})
	// check if they can stake
	if err := k.ValidateValidatorStaking(ctx, validator, msg.Value, sdk.Address(signer.Address())); err != nil {
		if sdk.ShowTimeTrackData {
			result := err.Result()
			fmt.Println(result.String())
		}
		return err.Result()
	}
	// change the validator state to staked
	err1 := k.StakeValidator(ctx, validator, msg.Value, signer)
	if err1 != nil {
		if sdk.ShowTimeTrackData {
			result := err.Result()
			fmt.Println(result.String())
		}
		return err.Result()
	}
	// create the event
	ctx.EventManager().EmitEvents(sdk.Events{
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
	defer sdk.TimeTrack(time.Now())

	ctx.Logger().Info("Begin Unstaking Message received from " + msg.Address.String())
	// move coins from the msg.Address account to a (self-delegation) delegator account
	// the validator account and global shares are updated within here
	validator, found := k.GetValidator(ctx, msg.Address)
	if !found {
		return types.ErrNoValidatorFound(k.Codespace()).Result()
	}
	err, valid := keeper.ValidateValidatorMsgSigner(validator, msg.Signer, k)
	if !valid {
		return err.Result()
	}

	if err := k.ValidateValidatorBeginUnstaking(ctx, validator); err != nil {
		return err.Result()
	}
	if err := k.WaitToBeginUnstakingValidator(ctx, validator); err != nil {
		return err.Result()
	}
	// create the event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeWaitingToBeginUnstaking,
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

// Validators must submit a transaction to unjail itself after todo
// having been jailed (and thus unstaked) for downtime
func handleMsgUnjail(ctx sdk.Ctx, msg types.MsgUnjail, k keeper.Keeper) sdk.Result {
	defer sdk.TimeTrack(time.Now())

	ctx.Logger().Info("Unjail Message received from " + msg.ValidatorAddr.String())
	addr, err := k.ValidateUnjailMessage(ctx, msg)
	if err != nil {
		return err.Result()
	}
	k.UnjailValidator(ctx, addr)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.ValidatorAddr.String()),
		),
	)
	return sdk.Result{Events: ctx.EventManager().Events()}
}

func handleMsgSend(ctx sdk.Ctx, msg types.MsgSend, k keeper.Keeper) sdk.Result {
	defer sdk.TimeTrack(time.Now())

	ctx.Logger().Info("Send Message from " + msg.FromAddress.String() + " received")
	err := k.SendCoins(ctx, msg.FromAddress, msg.ToAddress, msg.Amount)
	if err != nil {
		return err.Result()
	}
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	)
	return sdk.Result{Events: ctx.EventManager().Events()}
}

func handleMsgPause(ctx sdk.Ctx, msg types.MsgPause, k keeper.Keeper) sdk.Result {
	defer sdk.TimeTrack(time.Now())

	ctx.Logger().Info("Pause Node Message received from " + msg.ValidatorAddr.String())

	// Validate the PauseNode message
	addr, err := k.ValidatePauseNodeMessage(ctx, msg)
	if err != nil {
		return err.Result()
	}

	k.PauseNode(ctx, addr)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.ValidatorAddr.String()),
		),
	)
	return sdk.Result{Events: ctx.EventManager().Events()}
}

func handleMsgUnpause(ctx sdk.Ctx, msg types.MsgUnpause, k keeper.Keeper) sdk.Result {
	defer sdk.TimeTrack(time.Now())

	ctx.Logger().Info("Unpause Node Message received from " + msg.ValidatorAddr.String())

	// Validate the unpause message
	addr, err := k.ValidateUnpauseNodeMessage(ctx, msg)
	if err != nil {
		return err.Result()
	}

	// Only unpause if the message was valid
	k.UnpauseNode(ctx, addr)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.ValidatorAddr.String()),
		),
	)
	return sdk.Result{Events: ctx.EventManager().Events()}
}
