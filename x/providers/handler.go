package pos

import (
	"fmt"
	"reflect"

	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/providers/keeper"
	"github.com/vipernet-xyz/viper-network/x/providers/types"
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
		case types.MsgStakingKey:
			return HandleMsgStoreStakingKey(ctx, k, msg)
		default:
			errMsg := fmt.Sprintf("unrecognized staking message type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleStake(ctx sdk.Ctx, msg types.MsgStake, k keeper.Keeper) sdk.Result {
	pk := msg.PubKey
	addr := pk.Address()
	ctx.Logger().Info("Begin Staking Provider Message received from " + sdk.Address(pk.Address()).String())
	// create provider object using the message fields
	provider := types.NewProvider(sdk.Address(addr), pk, msg.Chains, sdk.ZeroInt())
	ctx.Logger().Info("Validate Provider Can Stake " + sdk.Address(addr).String())
	// check if they can stake
	if err := k.ValidateProviderStaking(ctx, provider, msg.Value); err != nil {
		ctx.Logger().Error(fmt.Sprintf("Validate Provider Can Stake Error, at height: %d with address: %s", ctx.BlockHeight(), sdk.Address(addr).String()))
		return err.Result()
	}
	ctx.Logger().Info("Change Provider state to Staked " + sdk.Address(addr).String())
	// change the provider state to staked
	err := k.StakeProvider(ctx, provider, msg.Value)
	if err != nil {
		return err.Result()
	}
	// create the event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreateProvider,
			sdk.NewAttribute(types.AttributeKeyProvider, sdk.Address(addr).String()),
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
	provider, found := k.GetProvider(ctx, msg.Address)
	if !found {
		ctx.Logger().Error(fmt.Sprintf("Provider Not Found at height: %d", ctx.BlockHeight()) + msg.Address.String())
		return types.ErrNoProviderFound(k.Codespace()).Result()
	}
	if err := k.ValidateProviderBeginUnstaking(ctx, provider); err != nil {
		ctx.Logger().Error(fmt.Sprintf("Provider Unstake Validation Not Successful, at height: %d", ctx.BlockHeight()) + msg.Address.String())
		return err.Result()
	}
	ctx.Logger().Info("Starting to Unstake Provider " + msg.Address.String())
	k.BeginUnstakingProvider(ctx, provider)
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

// Providers must submit a transaction to unjail itself after todo
// having been jailed (and thus unstaked) for downtime
func handleMsgUnjail(ctx sdk.Ctx, msg types.MsgUnjail, k keeper.Keeper) sdk.Result {
	consAddr, err := k.ValidateUnjailMessage(ctx, msg)
	if err != nil {
		return err.Result()
	}
	k.UnjailProvider(ctx, consAddr)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.ProviderAddr.String()),
		),
	)
	return sdk.Result{Events: ctx.EventManager().Events()}
}

// HandleMsgStoreStakingKey handles the MsgStoreStakingKey message.
func HandleMsgStoreStakingKey(ctx sdk.Ctx, k keeper.Keeper, msg types.MsgStakingKey) sdk.Result {
	// Store the staking key in the module's state
	k.SetStakingKey(ctx, msg.Address, msg.StakingKey)

	return sdk.Result{}
}
