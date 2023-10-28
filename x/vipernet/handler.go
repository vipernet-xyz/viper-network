package vipernet

import (
	"fmt"
	"reflect"
	"time"

	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/vipernet/keeper"
	"github.com/vipernet-xyz/viper-network/x/vipernet/types"
)

// "NewHandler" - Returns a handler for "vipernet" type messages.
func NewHandler(keeper keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Ctx, msg sdk.Msg, _ crypto.PublicKey) sdk.Result {

		ctx = ctx.WithEventManager(sdk.NewEventManager())

		// convert to value for switch consistency
		if reflect.ValueOf(msg).Kind() == reflect.Ptr {
			msg = reflect.Indirect(reflect.ValueOf(msg)).Interface().(sdk.Msg)
		}
		switch msg := msg.(type) {
		// handle claim message
		case types.MsgClaim:
			return handleClaimMsg(ctx, keeper, msg)
		// handle legacy proof message
		case types.MsgProof:
			return handleProofMsg(ctx, keeper, msg)
		case types.MsgSubmitReportCard:
			return handleSubmitReportCardMsg(ctx, keeper, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized vipernet ProtoMsg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// "handleClaimMsg" - General handler for the claim message
func handleClaimMsg(ctx sdk.Ctx, k keeper.Keeper, msg types.MsgClaim) sdk.Result {
	defer sdk.TimeTrack(time.Now())
	// validate the claim message
	if err := k.ValidateClaim(ctx, msg); err != nil {
		return err.Result()
	}
	// set the claim in the world state
	err := k.SetClaim(ctx, msg)
	if err != nil {
		return sdk.ErrInternal(err.Error()).Result()
	}
	// create the event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeClaim,
			sdk.NewAttribute(types.AttributeKeyValidator, msg.FromAddress.String()),
		),
	})
	return sdk.Result{Events: ctx.EventManager().Events()}
}

// "handleProofMsg" - General handler for the proof message
func handleProofMsg(ctx sdk.Ctx, k keeper.Keeper, proof types.MsgProof) sdk.Result {
	defer sdk.TimeTrack(time.Now())

	// validate the claim claim
	addr, claim, err := k.ValidateProof(ctx, proof)
	if err != nil {
		if err.Code() == types.CodeInvalidMerkleVerifyError && !claim.IsEmpty() {
			// delete local evidence
			processSelf(ctx, proof.GetSigners()[0], claim.SessionHeader, claim.EvidenceType, sdk.ZeroInt())
			return err.Result()
		}
		if err.Code() == types.CodeReplayAttackError && !claim.IsEmpty() {
			// delete local evidence
			processSelf(ctx, proof.GetSigners()[0], claim.SessionHeader, claim.EvidenceType, sdk.ZeroInt())
			// if is a replay attack, handle accordingly
			k.HandleReplayAttack(ctx, addr, sdk.NewInt(claim.TotalProofs))
			err := k.DeleteClaim(ctx, addr, claim.SessionHeader, claim.EvidenceType)
			if err != nil {
				ctx.Logger().Error("Could not delete claim from world state after replay attack detected", "Address", claim.FromAddress)
			}
		}
		return err.Result()
	}
	// valid claim message so execute according to type
	tokens, err := k.ExecuteProof(ctx, proof, claim)
	if err != nil {
		return err.Result()
	}
	// delete local evidence
	processSelf(ctx, proof.GetSigners()[0], claim.SessionHeader, claim.EvidenceType, tokens)
	// create the event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeProof,
			sdk.NewAttribute(types.AttributeKeyValidator, addr.String()),
		),
	})
	return sdk.Result{Events: ctx.EventManager().Events()}
}

func processSelf(ctx sdk.Ctx, signer sdk.Address, header types.SessionHeader, evidenceType types.EvidenceType, tokens sdk.BigInt) {
	node, ok := types.GlobalViperNodes[signer.String()]
	if !ok {
		return
	}
	evidenceStore := node.EvidenceStore
	err := types.DeleteEvidence(header, evidenceType, evidenceStore)
	if err != nil {
		ctx.Logger().Error("Unable to delete evidence: " + err.Error())
	}
	if !tokens.IsZero() {
		if types.GlobalViperConfig.LeanViper {
			go types.GlobalServiceMetric().AdduviprEarnedFor(header.Chain, float64(tokens.Int64()), &signer)
		} else {
			types.GlobalServiceMetric().AdduviprEarnedFor(header.Chain, float64(tokens.Int64()), &signer)
		}
	}
}

// "handleSubmitReportCardMsg" - General handler for the MsgSubmitReportCard message
func handleSubmitReportCardMsg(ctx sdk.Ctx, k keeper.Keeper, msg types.MsgSubmitReportCard) sdk.Result {
	defer sdk.TimeTrack(time.Now())

	// validate the report card submission message
	if err := msg.ValidateBasic(); err != nil {
		return err.Result()
	}

	err := k.SetReportCard(ctx, msg)
	if err != nil {
		return sdk.ErrInternal(err.Error()).Result()
	}

	// create the event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeSubmitReportCard,
			sdk.NewAttribute(types.AttributeKeyValidator, msg.ServicerAddress.String()),
		),
	})

	return sdk.Result{Events: ctx.EventManager().Events()}
}
