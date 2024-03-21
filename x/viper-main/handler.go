package vipernet

import (
	"fmt"
	"reflect"
	"time"

	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/viper-main/keeper"
	"github.com/vipernet-xyz/viper-network/x/viper-main/types"
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
		case types.MsgSubmitQoSReport:
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
	// validate the proof
	addr, reportCard, claim, err, errorType := k.ValidateProof(ctx, proof)
	if err != nil && errorType == 1 {
		if err.Code() == types.CodeInvalidClaimMerkleVerifyError && !claim.IsEmpty() {
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
			err = k.DeleteReportCard(ctx, addr, reportCard.FishermanAddress, reportCard.SessionHeader, reportCard.EvidenceType)
			if err != nil {
				ctx.Logger().Error("Could not delete report card from world state after replay attack detected", "Address", reportCard.ServicerAddress)
			}
		}
		return err.Result()
	}
	if err != nil && errorType == 2 {
		var report types.ViperQoSReport
		report.LatencyScore = sdk.NewDec(1)
		report.AvailabilityScore = sdk.NewDec(1)
		report.ReliabilityScore = sdk.NewDec(1)

		var qos types.MsgSubmitQoSReport
		qos.SessionHeader = claim.SessionHeader
		qos.ServicerAddress = claim.FromAddress
		qos.Report = report
		qos.EvidenceType = types.FishermanTestEvidence
		// Set report card with max score of 1

		k.SetReportCard(ctx, qos)
		tokens, _, err := k.ExecuteProof(ctx, proof, qos, claim)
		if err != nil {
			return err.Result()
		}
		k.HandleFishermanSlash(ctx, claim.SessionHeader, ctx.BlockHeight())
		processSelf(ctx, proof.GetSigners()[0], claim.SessionHeader, claim.EvidenceType, tokens)
		ctx.EventManager().EmitEvents(sdk.Events{
			sdk.NewEvent(
				types.EventTypeProof,
				sdk.NewAttribute(types.AttributeKeyValidator, addr.String()),
			),
		})
		return sdk.Result{Events: ctx.EventManager().Events()}
	}
	// valid claim message so execute according to type
	tokens, _, err := k.ExecuteProof(ctx, proof, reportCard, claim)
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

// "handleSubmitReportCardMsg" - General handler for the MsgSubmitReport message
func handleSubmitReportCardMsg(ctx sdk.Ctx, k keeper.Keeper, msg types.MsgSubmitQoSReport) sdk.Result {
	defer sdk.TimeTrack(time.Now())
	// validate the report card submission message
	if err := k.ValidateSumbitReportCard(ctx, msg); err != nil {
		return err.Result()
	}

	// Set the valid report card
	err := k.SetReportCard(ctx, msg)
	if err != nil {
		return sdk.ErrInternal(err.Error()).Result()
	}

	processResult(ctx, msg.FishermanAddress, msg.SessionHeader, msg.EvidenceType, msg.Report)
	// create the event
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeSubmitReportCard,
			sdk.NewAttribute(types.AttributeKeyValidator, msg.ServicerAddress.String()),
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
			go types.GlobalServiceMetric().AddUVIPREarnedFor(header.Chain, float64(tokens.Int64()), &signer)
		} else {
			types.GlobalServiceMetric().AddUVIPREarnedFor(header.Chain, float64(tokens.Int64()), &signer)
		}
	}
}

func processResult(ctx sdk.Ctx, signer sdk.Address, header types.SessionHeader, evidenceType types.EvidenceType, reportCard types.ViperQoSReport) {
	node, ok := types.GlobalViperNodes[signer.String()]
	if !ok {
		return
	}
	testStore := node.TestStore
	err := types.DeleteResult(header, evidenceType, reportCard.ServicerAddress, testStore)
	if err != nil {
		ctx.Logger().Error("Unable to delete result: " + err.Error())
	}

	if !reportCard.ServicerAddress.Empty() {
		// Convert BigDec to float64
		latencyScoreFloat64 := bigDecToFloat64(reportCard.LatencyScore)
		availabilityScoreFloat64 := bigDecToFloat64(reportCard.AvailabilityScore)
		reliabilityScoreFloat64 := bigDecToFloat64(reportCard.ReliabilityScore)

		if types.GlobalViperConfig.LeanViper {
			go types.GlobalServiceMetric().AddReportCardMetric(
				header.Chain,
				latencyScoreFloat64,
				availabilityScoreFloat64,
				reliabilityScoreFloat64,
				&reportCard.ServicerAddress,
			)
		} else {
			types.GlobalServiceMetric().AddReportCardMetric(
				header.Chain,
				latencyScoreFloat64,
				availabilityScoreFloat64,
				reliabilityScoreFloat64,
				&reportCard.ServicerAddress,
			)
		}
	}
}

func bigDecToFloat64(value sdk.BigDec) float64 {
	roundedScore := value.RoundInt()
	return float64(roundedScore.Int64())
}
