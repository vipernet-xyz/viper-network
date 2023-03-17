package fee

import (
	//"fmt"
	"reflect"

	"github.com/vipernet-xyz/viper-network/crypto"

	"github.com/vipernet-xyz/viper-network/modules/apps/29-fee/keeper"
	sdk "github.com/vipernet-xyz/viper-network/types"
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
