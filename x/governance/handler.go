package governance

import (
	"fmt"
	"reflect"

	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/governance/keeper"
	"github.com/vipernet-xyz/viper-network/x/governance/types"
)

func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Ctx, msg sdk.Msg, _ crypto.PublicKey) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())
		// convert to value for switch consistency
		if reflect.ValueOf(msg).Kind() == reflect.Ptr {
			msg = reflect.Indirect(reflect.ValueOf(msg)).Interface().(sdk.Msg)
		}
		switch msg := msg.(type) {
		case types.MsgChangeParam:
			return handleMsgChangeParam(ctx, msg, k)
		case types.MsgDAOTransfer:
			return handleMsgDaoTransfer(ctx, msg, k)
		case types.MsgUpgrade:
			return handleMsgUpgrade(ctx, msg, k)
		case types.MsgGenerateDiscountKey:
			return handleMsgGenerateDiscountKey(ctx, k, msg)
		default:
			errMsg := fmt.Sprintf("unrecognized governance message type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgChangeParam(ctx sdk.Ctx, msg types.MsgChangeParam, k keeper.Keeper) sdk.Result {
	return k.ModifyParam(ctx, msg.ParamKey, msg.ParamVal, msg.FromAddress)
}

func handleMsgDaoTransfer(ctx sdk.Ctx, msg types.MsgDAOTransfer, k keeper.Keeper) sdk.Result {
	da, err := types.DAOActionFromString(msg.Action)
	if err != nil {
		return err.Result()
	}
	switch da {
	case types.DAOTransfer:
		return k.DAOTransferFrom(ctx, msg.FromAddress, msg.ToAddress, msg.Amount)
	case types.DAOBurn:
		return k.DAOBurn(ctx, msg.FromAddress, msg.Amount)
	}
	return sdk.Result{}
}

func handleMsgUpgrade(ctx sdk.Ctx, msg types.MsgUpgrade, k keeper.Keeper) sdk.Result {
	return k.HandleUpgrade(ctx, types.NewACLKey(ModuleName, string(types.UpgradeKey)), msg.Upgrade, msg.Address)
}

func handleMsgGenerateDiscountKey(ctx sdk.Ctx, k keeper.Keeper, msg types.MsgGenerateDiscountKey) sdk.Result {
	// Check if a discount key already exists for the given address
	if k.HasDiscountKey(ctx, msg.ToAddress) {
		existingKey := k.GetDiscountKey(ctx, msg.ToAddress) // Fetch the existing discount key
		ctx.Logger().Info(fmt.Sprintf("Discount Key already exists for address %s: %s", msg.ToAddress, existingKey))
		return sdk.Result{
			Events: ctx.EventManager().ABCIEvents(),
		}
	}

	// Store the generated discount key in the state using the keeper
	err := k.SetDiscountKey(ctx, msg.ToAddress, msg.DiscountKey)
	if err != nil {
		return sdk.ErrInternal(fmt.Sprintf("Failed to set discount key: %s", err.Error())).Result()
	}

	ctx.Logger().Info(fmt.Sprintf("New Discount Key set for address %s: %s", msg.ToAddress, msg.DiscountKey))

	return sdk.Result{
		Events: ctx.EventManager().ABCIEvents(),
	}
}

// Content defines an interface that a proposal must implement. It contains
// information such as the title and description along with the type and routing
// information for the appropriate handler to process the proposal. Content can
// have additional fields, which will handled by a proposal's Handler.
type Content interface {
	GetTitle() string
	GetDescription() string
	ProposalRoute() string
	ProposalType() string
	ValidateBasic() error
	String() string
}

// Handler defines a function that handles a proposal after it has passed the
// governance process.
type Handler func(ctx sdk.Ctx, content Content) error
