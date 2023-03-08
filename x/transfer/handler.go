package transfer

import (
	//"fmt"
	"reflect"

	"github.com/vipernet-xyz/viper-network/crypto"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/transfer/keeper"
	//"github.com/vipernet-xyz/viper-network/x/transfer/types"
)

func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Ctx, msg sdk.Msg, _ crypto.PublicKey) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		// convert to value for switch consistency
		if reflect.ValueOf(msg).Kind() == reflect.Ptr {
			msg = reflect.Indirect(reflect.ValueOf(msg)).Interface().(sdk.Msg)
		}
		//switch msg := msg.(type)
		{
		}
		return sdk.Result{}
	}
}

/*func handleMsgTransfer(ctx sdk.Ctx, msg types.MsgTransfer, k keeper.Keeper) sdk.Result {
	da, err := types.IBCActionFromString(msg.Action)
	if err != nil {
		return err.Result()
	}
	switch da {
	case types.Transfer:
		return k.IBCTransferFrom(ctx, msg.FromAddress, msg.ToAddress, msg.Amount)
	case types.Burn:
		return k.Burn(ctx, msg.FromAddress, msg.Amount)
	}
	return sdk.Result{}
}*/
