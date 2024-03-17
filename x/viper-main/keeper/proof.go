package keeper

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"time"

	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/authentication"
	"github.com/vipernet-xyz/viper-network/x/authentication/util"
	requestorsType "github.com/vipernet-xyz/viper-network/x/requestors/types"
	servicersTypes "github.com/vipernet-xyz/viper-network/x/servicers/types"
	vc "github.com/vipernet-xyz/viper-network/x/viper-main/types"

	"github.com/tendermint/tendermint/rpc/client"
)

// auto sends a proof transaction for the claim
func (k Keeper) SendProofTx(ctx sdk.Ctx, n client.Client, node *vc.ViperNode, proofTx func(cliCtx util.CLIContext, txBuilder authentication.TxBuilder, claimMerkleProof vc.MerkleProof, claimLeafNode vc.Proof, claimEvidenceType vc.EvidenceType, reportMerkleProof vc.MerkleProof, reportLeafNode vc.Test, reportEvidenceType vc.EvidenceType) (*sdk.TxResponse, error)) {
	addr := node.GetAddress()
	now := time.Now()
	// Get all mature claims for the address
	claims, err := k.GetMatureClaims(ctx, addr)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("an error occurred getting the mature claims in the Proof Transaction:\n%v", err))
		return
	}
	// Get all mature report cards for the address
	reportCards, err := k.GetMatureReportCards(ctx, addr)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("an error occurred getting the mature report cards in the Proof Transaction:\n%v", err))
		return
	}
	// Map to store claims and report cards for each session header
	sessionClaimsReportCards := make(map[vc.SessionHeader][]struct {
		claim      vc.MsgClaim
		reportCard vc.MsgSubmitQoSReport
	})

	// Group claims and report cards by session header
	for _, claim := range claims {
		found := false
		for _, reportCard := range reportCards {
			if reportCard.SessionHeader.Equal(claim.SessionHeader) {
				sessionClaimsReportCards[claim.SessionHeader] = append(sessionClaimsReportCards[claim.SessionHeader], struct {
					claim      vc.MsgClaim
					reportCard vc.MsgSubmitQoSReport
				}{claim: claim, reportCard: reportCard})
				found = true
			}
		}
		// If no report card was found for the claim, append an empty report card
		if !found {
			sessionClaimsReportCards[claim.SessionHeader] = append(sessionClaimsReportCards[claim.SessionHeader], struct {
				claim      vc.MsgClaim
				reportCard vc.MsgSubmitQoSReport
			}{claim: claim, reportCard: vc.MsgSubmitQoSReport{}})
		}
	}

	// Process claims and report cards for each session header
	for sessionHeader, claimsReportCards := range sessionClaimsReportCards {
		for _, claimReportCard := range claimsReportCards {
			claim := claimReportCard.claim
			reportCard := claimReportCard.reportCard

			// Retrieve evidence object
			evidence, err := vc.GetEvidence(sessionHeader, claim.EvidenceType, sdk.ZeroInt(), node.EvidenceStore)
			if err != nil || evidence.Proofs == nil || len(evidence.Proofs) == 0 {
				ctx.Logger().Info(fmt.Sprintf("the evidence object for evidence is not found, ignoring pending claim for req: %s, at sessionHeight: %d", claim.SessionHeader.RequestorPubKey, claim.SessionHeader.SessionBlockHeight))
				continue
			}

			if ctx.BlockHeight()-sessionHeader.SessionBlockHeight > int64(vc.GlobalViperConfig.MaxClaimAgeForProofRetry) {
				err := vc.DeleteEvidence(claim.SessionHeader, claim.EvidenceType, node.EvidenceStore)
				ctx.Logger().Error(fmt.Sprintf("deleting evidence older than MaxClaimAgeForProofRetry"))
				if err != nil {
					ctx.Logger().Error(fmt.Sprintf("unable to delete evidence that is older than 32 blocks: %s", err.Error()))
				}
				continue
			}

			if !node.EvidenceStore.IsSealed(evidence) {
				err := vc.DeleteEvidence(claim.SessionHeader, claim.EvidenceType, node.EvidenceStore)
				ctx.Logger().Error(fmt.Sprintf("evidence is not sealed, could cause a relay leak:"))
				if err != nil {
					ctx.Logger().Error(fmt.Sprintf("could not delete evidence is not sealed, could cause a relay leak: %s", err.Error()))
				}
			}

			if evidence.NumOfProofs != claim.TotalProofs {
				err := vc.DeleteEvidence(sessionHeader, claim.EvidenceType, node.EvidenceStore)
				ctx.Logger().Error(fmt.Sprintf("evidence num of proofs does not equal claim total proofs... possible relay leak"))
				if err != nil {
					ctx.Logger().Error(fmt.Sprintf("evidence num of proofs does not equal claim total proofs... possible relay leak: %s", err.Error()))
				}
			}

			// Validate session context
			sessionCtx, err := ctx.PrevCtx(sessionHeader.SessionBlockHeight)
			if err != nil {
				ctx.Logger().Info(fmt.Sprintf("could not get Session Context, ignoring pending claim and report card for req: %s, at sessionHeight: %d", sessionHeader.RequestorPubKey, sessionHeader.SessionBlockHeight))
				continue
			}

			// Get the pseudorandom index
			index, err := k.getPseudorandomIndex(ctx, claim.TotalProofs, sessionHeader, sessionCtx)
			if err != nil {
				ctx.Logger().Error(err.Error())
				continue
			}
			app, found := k.GetRequestorFromPublicKey(sessionCtx, sessionHeader.RequestorPubKey)
			if !found {
				ctx.Logger().Error(fmt.Sprintf("an error occurred creating the proof transaction with req %s not found with evidence %v", evidence.RequestorPubKey, evidence))
			}
			// Get the Merkle proof object for the claim
			claimMProof, claimLeaf := evidence.GenerateMerkleProof(sessionHeader.SessionBlockHeight, int(index), vc.MaxPossibleRelays(app, int64(app.GetNumServicers())).Int64())
			reportMProof := reportCard.MerkleProof
			reportLeaf := reportCard.Leaf.FromProto()

			// If prevalidation is enabled, validate the Merkle proofs
			if vc.GlobalViperConfig.ProofPrevalidation {
				//claim
				levelCount := len(claimMProof.HashRanges)
				if levelCount != int(math.Ceil(math.Log2(float64(claim.TotalProofs)))) {
					ctx.Logger().Error(fmt.Sprintf("produced invalid proof for pending claim for req: %s, at sessionHeight: %d, level count", claim.SessionHeader.RequestorPubKey, claim.SessionHeader.SessionBlockHeight))
					continue
				}

				if isValid, _ := claimMProof.Validate(sessionHeader.SessionBlockHeight, claim.MerkleRoot, claimLeaf, len(claimMProof.HashRanges)); !isValid {
					ctx.Logger().Error(fmt.Sprintf("produced invalid proof for pending claim for req: %s, at sessionHeight: %d", claim.SessionHeader.RequestorPubKey, claim.SessionHeader.SessionBlockHeight))
					continue
				}

				//report card
				levelCount = len(reportMProof.HashRanges)
				if levelCount != int(math.Ceil(math.Log2(float64(reportCard.NumOfTestResults)))) {
					ctx.Logger().Error(fmt.Sprintf("produced invalid proof for pending  for req: %s, at sessionHeight: %d, level count", reportCard.SessionHeader.RequestorPubKey, reportCard.SessionHeader.SessionBlockHeight))
					continue
				}

				if isValid, _ := reportMProof.ValidateTR(sessionHeader.SessionBlockHeight, reportCard.Report.SampleRoot, reportLeaf, len(reportMProof.HashRanges)); !isValid {
					ctx.Logger().Error(fmt.Sprintf("produced invalid proof for report card validation, at sessionHeight: %d", reportCard.SessionHeader.SessionBlockHeight))
					continue
				}
			}
			proofTxTotalTime := float64(time.Since(now))
			go func() {
				vc.GlobalServiceMetric().AddProofTiming(sessionHeader.Chain, proofTxTotalTime, &addr)
			}()
			// Generate the auto txbuilder and clictx
			txBuilder, cliCtx, err := newTxBuilderAndCliCtx(ctx, &vc.MsgProof{}, n, node.PrivateKey, k)
			if err != nil {
				ctx.Logger().Error(fmt.Sprintf("an error occurred in the transaction process of the Proof Transaction:\n%v", err))
				return
			}
			// Send the proof TX
			_, err = proofTx(cliCtx, txBuilder, claimMProof, claimLeaf, claim.EvidenceType, reportMProof, reportLeaf, reportCard.EvidenceType)
			if err != nil {
				ctx.Logger().Error(err.Error())
			}
		}
	}
}

func (k Keeper) ValidateProof(ctx sdk.Ctx, proof vc.MsgProof) (servicerAddr sdk.Address, reportCard vc.MsgSubmitQoSReport, claim vc.MsgClaim, sdkError sdk.Error, errorType int64) {
	// get the public key from the claim
	servicerAddr = proof.GetSigners()[0]
	// get the claim for the address
	claim, found := k.GetClaim(ctx, servicerAddr, proof.GetClaimLeaf().SessionHeader(), proof.ClaimEvidenceType)
	// if the claim is not found for this claim
	if !found {
		return servicerAddr, reportCard, claim, vc.NewClaimNotFoundError(vc.ModuleName), 1
	}
	// validate level count on claim by total relays
	levelCount := len(proof.ClaimMerkleProof.HashRanges)
	if levelCount != int(math.Ceil(math.Log2(float64(claim.TotalProofs)))) {
		return servicerAddr, reportCard, claim, vc.NewInvalidProofsError(vc.ModuleName), 1
	}
	var hasMatch bool
	for _, m := range proof.ClaimMerkleProof.HashRanges {
		if claim.MerkleRoot.Range.Upper == m.Range.Upper {
			hasMatch = true
			break
		}
	}
	if !hasMatch && proof.ClaimMerkleProof.Target.Range.Upper != claim.MerkleRoot.Range.Upper {
		return servicerAddr, reportCard, claim, vc.NewInvalidClaimMerkleVerifyError(vc.ModuleName), 1
	}

	// get the session context
	sessionCtx, err := ctx.PrevCtx(claim.SessionHeader.SessionBlockHeight)
	if err != nil {
		return servicerAddr, reportCard, claim, sdk.ErrInternal(err.Error()), 1
	}
	// validate the proof
	ctx.Logger().Info(fmt.Sprintf("Generate psuedorandom proof with %d proofs for claim, at session height of %d, for requestor: %s", claim.TotalProofs, claim.SessionHeader.SessionBlockHeight, claim.SessionHeader.RequestorPubKey))
	reqProof, err := k.getPseudorandomIndex(ctx, claim.TotalProofs, claim.SessionHeader, sessionCtx)
	if err != nil {
		return servicerAddr, reportCard, claim, sdk.ErrInternal(err.Error()), 1
	}
	// if the required proof message index does not match the leaf servicer index
	if reqProof != int64(proof.ClaimMerkleProof.TargetIndex) {
		return servicerAddr, reportCard, claim, vc.NewInvalidProofsError(vc.ModuleName), 1
	}
	// validate the merkle proofs
	isValid, _ := proof.ClaimMerkleProof.Validate(claim.SessionHeader.SessionBlockHeight, claim.MerkleRoot, proof.GetClaimLeaf(), levelCount)
	// if is not valid for other reasons
	if !isValid {
		return servicerAddr, reportCard, claim, vc.NewReplayAttackError(vc.ModuleName), 1
	}
	// get the requestor
	requestor, found := k.GetRequestorFromPublicKey(sessionCtx, claim.SessionHeader.RequestorPubKey)
	if !found {
		return servicerAddr, reportCard, claim, vc.NewRequestorNotFoundError(vc.ModuleName), 1
	}
	// validate the proof depending on the type of proof it is
	er := proof.GetClaimLeaf().Validate(requestor.GetChains(), int(requestor.GetNumServicers()), claim.SessionHeader.SessionBlockHeight)
	if er != nil {
		return nil, reportCard, claim, er, 1
	}
	if len(proof.ReportMerkleProof.HashRanges) == 0 || proof.ReportLeaf == nil {
		return servicerAddr, reportCard, claim, vc.NewNoReportCardError(vc.ModuleName), 2
	}

	// Fetch the report card for the respective node and session
	reportCard, Found := k.GetReportCard(ctx, servicerAddr, proof.GetClaimLeaf().SessionHeader(), vc.FishermanTestEvidence)
	if !Found {
		return servicerAddr, reportCard, claim, vc.NewReportCardNotFoundError(vc.ModuleName), 2
	}
	// Verify the signature on the report card against the Fisherman's signature
	if valid, _ := k.verifyReportCardSignature(ctx, reportCard, reportCard.Report.Signature); !valid {
		return servicerAddr, reportCard, claim, vc.NewInvalidSignatureError(vc.ModuleName), 2
	}

	levelCount = len(proof.ReportMerkleProof.HashRanges)
	if levelCount != int(math.Ceil(math.Log2(float64(reportCard.NumOfTestResults)))) {
		return servicerAddr, reportCard, claim, vc.NewInvalidProofsError(vc.ModuleName), 2
	}

	for _, m := range proof.ReportMerkleProof.HashRanges {
		if reportCard.Report.SampleRoot.Range.Upper == m.Range.Upper {
			hasMatch = true
			break
		}
	}
	if !hasMatch && proof.ReportMerkleProof.Target.Range.Upper != reportCard.Report.SampleRoot.Range.Upper {
		return servicerAddr, reportCard, claim, vc.NewInvalidReportMerkleVerifyError(vc.ModuleName), 2
	}
	ctx.Logger().Info(fmt.Sprintf("Generate psuedorandom proof with %d proofs for report card, at session height of %d, for requestor: %s", reportCard.NumOfTestResults, reportCard.SessionHeader.SessionBlockHeight, reportCard.SessionHeader.RequestorPubKey))
	reqProof, err = k.getPseudorandomIndexForRC(ctx, reportCard.NumOfTestResults, reportCard.SessionHeader, sessionCtx)
	if err != nil {
		return servicerAddr, reportCard, claim, sdk.ErrInternal(err.Error()), 1
	}
	// if the required proof message index does not match the leaf servicer index
	if reqProof != int64(proof.ReportMerkleProof.TargetIndex) {
		return servicerAddr, reportCard, claim, vc.NewInvalidProofsError(vc.ModuleName), 2
	}
	// validate the merkle proofs
	isValid, _ = proof.ReportMerkleProof.ValidateTR(reportCard.SessionHeader.SessionBlockHeight, reportCard.Report.SampleRoot, proof.GetReportLeaf(), len(proof.ReportMerkleProof.HashRanges))
	// if is not valid for other reasons
	if !isValid {
		return servicerAddr, reportCard, claim, vc.NewReplayAttackError(vc.ModuleName), 2
	}
	// validate the proof depending on the type of proof it is
	er1 := proof.GetReportLeaf().ValidateBasic()
	if er1 != nil {
		return nil, reportCard, claim, er1, 2
	}
	// return the needed info to the handler
	return servicerAddr, reportCard, claim, nil, 0
}

func (k Keeper) ExecuteProof(ctx sdk.Ctx, proof vc.MsgProof, reportCard vc.MsgSubmitQoSReport, claim vc.MsgClaim) (tokens sdk.BigInt, updatedReportCard servicersTypes.ReportCard, err sdk.Error) {
	//requestor address
	pk, _ := crypto.NewPublicKey(claim.SessionHeader.RequestorPubKey)
	requestorAddress := pk.Address().Bytes()
	p := k.requestorKeeper.Requestor(ctx, requestorAddress)
	requestor := p.(requestorsType.Requestor)
	// convert to value for switch consistency
	l := proof.GetClaimLeaf()
	if reflect.ValueOf(l).Kind() == reflect.Ptr {
		l = reflect.Indirect(reflect.ValueOf(l)).Interface().(vc.Proof)
	}
	switch l.(type) {
	case vc.RelayProof:
		ctx.Logger().Info(fmt.Sprintf("reward coins to %s, for %d relays", claim.FromAddress.String(), claim.TotalProofs))
		tokens = k.AwardCoinsForRelays(ctx, reportCard, claim.TotalProofs, claim.FromAddress, requestor)
		err := k.DeleteClaim(ctx, claim.FromAddress, claim.SessionHeader, vc.RelayEvidence)
		updatedReportCard = k.UpdateReportCard(ctx, reportCard.ServicerAddress, reportCard, vc.FishermanTestEvidence)
		if err != nil {
			return tokens, updatedReportCard, sdk.ErrInternal(err.Error())
		}
	case vc.ChallengeProofInvalidData:
		ctx.Logger().Info(fmt.Sprintf("burning coins from %s, for %d valid challenges", claim.FromAddress.String(), claim.TotalProofs))
		proof, ok := proof.GetClaimLeaf().(vc.ChallengeProofInvalidData)
		if !ok {
			return sdk.ZeroInt(), servicersTypes.ReportCard{}, vc.NewInvalidProofsError(vc.ModuleName)
		}
		pk := proof.MinorityResponse.Proof.ServicerPubKey
		pubKey, err := crypto.NewPublicKey(pk)
		if err != nil {
			return sdk.ZeroInt(), servicersTypes.ReportCard{}, sdk.ErrInvalidPubKey(err.Error())
		}
		k.BurnCoinsForChallenges(ctx, claim.TotalProofs, sdk.Address(pubKey.Address()))
		err = k.DeleteClaim(ctx, claim.FromAddress, claim.SessionHeader, vc.ChallengeEvidence)
		if err != nil {
			return sdk.ZeroInt(), servicersTypes.ReportCard{}, sdk.ErrInternal(err.Error())
		}
		// small reward for the challenge proof invalid data
		tokens = k.AwardCoinsForRelays(ctx, reportCard, claim.TotalProofs/100, claim.FromAddress, requestor)
	}
	return tokens, updatedReportCard, nil
}

// struct used for creating the psuedorandom index
type pseudorandomGenerator struct {
	BlockHash string
	Header    string
}

// generates the required pseudorandom index for the zero knowledge proof
func (k Keeper) getPseudorandomIndex(ctx sdk.Ctx, totalRelays int64, header vc.SessionHeader, sessionCtx sdk.Ctx) (int64, error) {
	// get the context for the proof (the proof context is X sessions after the session began)
	proofHeight := header.SessionBlockHeight + k.ClaimSubmissionWindow(sessionCtx)*k.BlocksPerSession(sessionCtx) // next session block hash
	// get the pseudorandomGenerator json bytes
	blockHashBz, err := ctx.GetPrevBlockHash(proofHeight)
	if err != nil {
		return 0, err
	}
	headerHash := header.HashString()
	pseudoGenerator := pseudorandomGenerator{hex.EncodeToString(blockHashBz), headerHash}
	r, err := json.Marshal(pseudoGenerator)
	if err != nil {
		return 0, err
	}
	return vc.PseudorandomSelection(sdk.NewInt(totalRelays), vc.Hash(r)).Int64(), nil
}

func (k Keeper) HandleReplayAttack(ctx sdk.Ctx, address sdk.Address, numberOfChallenges sdk.BigInt) {
	ctx.Logger().Error(fmt.Sprintf("Replay Attack Detected: By %s, for %v proofs", address.String(), numberOfChallenges))
	k.posKeeper.BurnForChallenge(ctx, numberOfChallenges.Mul(sdk.NewInt(k.ReplayAttackBurnMultiplier(ctx))), address)
}

func newTxBuilderAndCliCtx(ctx sdk.Ctx, msg sdk.ProtoMsg, n client.Client, key crypto.PrivateKey, k Keeper) (txBuilder authentication.TxBuilder, cliCtx util.CLIContext, err error) {
	// get the from address from the pkf
	fromAddr := sdk.Address(key.PublicKey().Address())
	// create a client context for sending
	cliCtx = util.NewCLIContext(n, fromAddr, "").WithCodec(k.Cdc).WithHeight(ctx.BlockHeight())

	cliCtx.PrivateKey = key
	// broadcast synchronously
	cliCtx.BroadcastMode = util.BroadcastSync
	// get the account to ensure balance
	// retrieve the account for a balance check (and ensure it exists)
	account := k.authKeeper.GetAccount(ctx, fromAddr)
	if account == nil {
		return txBuilder, cliCtx, fmt.Errorf("unable to locate an account at address: %s", fromAddr)
	}
	// check the fee amount
	fee := k.authKeeper.GetFee(ctx, msg)
	if account.GetCoins().AmountOf(k.posKeeper.StakeDenom(ctx)).LT(fee) {
		return txBuilder, cliCtx, fmt.Errorf("insufficient funds for the auto %s transaction: the fee needed is %v ", msg.Type(), fee)
	}
	// ensure that the tx builder has the correct tx encoder, chainID, fee
	txBuilder = authentication.NewTxBuilder(
		authentication.DefaultTxEncoder(k.Cdc),
		authentication.DefaultTxDecoder(k.Cdc),
		ctx.ChainID(),
		"",
		sdk.NewCoins(sdk.NewCoin(k.posKeeper.StakeDenom(ctx), fee)),
	)
	return
}

// verifyReportCardSignature verifies the signature on the report card against the Fisherman's signature
func (k Keeper) verifyReportCardSignature(ctx sdk.Ctx, reportCard vc.MsgSubmitQoSReport, fishermanSignature string) (bool, error) {
	fisherman := k.posKeeper.Validator(ctx, reportCard.FishermanAddress)
	fishermanPK := fisherman.GetPublicKey().RawString()
	if err := vc.SignatureVerification(fishermanPK, reportCard.Report.HashString(), fishermanSignature); err != nil {
		return false, err
	}

	return true, nil
}
