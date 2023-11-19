package keeper

import (
	"encoding/hex"
	"fmt"
	"math"

	"github.com/tendermint/tendermint/rpc/client"
	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	auth "github.com/vipernet-xyz/viper-network/x/authentication"
	"github.com/vipernet-xyz/viper-network/x/authentication/util"
	vc "github.com/vipernet-xyz/viper-network/x/vipernet/types"
)

func (k Keeper) SendReportCardTx(ctx sdk.Ctx, keeper Keeper, n client.Client, node *vc.ViperNode, servicerAddr sdk.Address, sessionHeader vc.SessionHeader, evidenceType vc.EvidenceType, qosReport vc.ViperQoSReport, reportCardTx func(pk crypto.PrivateKey, cliCtx util.CLIContext, txBuilder auth.TxBuilder, header vc.SessionHeader, servicerAddr sdk.Address, reportCard vc.ViperQoSReport) (*sdk.TxResponse, error)) {

	// Use GetResult to fetch the result for the given session, servicer address and evidence type
	result, err := vc.GetResult(sessionHeader, evidenceType, servicerAddr, node.TestStore)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("An error occurred retrieving the result: \n%s", err.Error()))
		return
	}

	// If the blockchain in the evidence is not supported then return.
	if !k.IsViperSupportedBlockchain(ctx, sessionHeader.Chain) {
		ctx.Logger().Info(fmt.Sprintf("Report card for %s blockchain isn't viper supported, so will not send.", sessionHeader.Chain))
		return
	}

	if !k.IsViperSupportedGeoZone(ctx, sessionHeader.Chain) {
		ctx.Logger().Info(fmt.Sprintf("Report card for %s geozone isn't viper supported, so will not send.", sessionHeader.GeoZone))
		return
	}

	// Check if the current session is still ongoing
	if ctx.BlockHeight() <= sessionHeader.SessionBlockHeight+k.BlocksPerSession(ctx)-1 {
		ctx.Logger().Info("The session is ongoing, so will not send the report card yet.")
		return
	}

	// Check the current state to see if the report card has already been sent and processed (if so, then return)
	if rc, found := k.GetReportCard(ctx, qosReport.ServicerAddress, sessionHeader); found {
		ctx.Logger().Info(fmt.Sprintf("Report card already found for session: %v", rc.SessionHeader))
		return
	}

	// Check if the report card has expired
	if k.ReportCardIsExpired(ctx, result.SessionBlockHeight) {
		// Delete the result since we cannot submit an expired report card
		if err := vc.DeleteResult(sessionHeader, evidenceType, node.TestStore); err != nil {
			ctx.Logger().Debug(err.Error())
		}
		return
	}

	// Check if proof count is above minimum
	if result.NumOfTestResults < k.MinimumSampleRelays(ctx) {
		ctx.Logger().Info("Number of proofs is below the required minimum, will not send report card.")
		return
	}

	// Validate against the root
	if valid, err := k.validateReportCardAgainstRoot(ctx, qosReport, sessionHeader); !valid {
		ctx.Logger().Error(fmt.Sprintf("Report card validation against root failed. Error: %v", err))
		return
	}

	// Generate the auto txbuilder and clictx
	txBuilder, cliCtx, err := newTxBuilderAndCliCtx(ctx, &vc.MsgSubmitReportCard{}, n, node.PrivateKey, k)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("An error occurred creating the tx builder for the report card tx:\n%s", err.Error()))
		return
	}

	// Send in the report card
	if _, err := reportCardTx(node.PrivateKey, cliCtx, txBuilder, sessionHeader, servicerAddr, qosReport); err != nil {
		ctx.Logger().Error(fmt.Sprintf("An error occurred executing the report card transaction: \n%s", err.Error()))
	}
}

func (k Keeper) ValidateSumbitReportCard(ctx sdk.Ctx, submitReportcard vc.MsgSubmitReportCard) (err sdk.Error) {
	// check to see if evidence type is included in the message
	if submitReportcard.EvidenceType == 0 {
		return vc.NewNoEvidenceTypeErr(vc.ModuleName)
	}
	// get the session context (state info at the beginning of the session)
	sessionContext, er := ctx.PrevCtx(submitReportcard.SessionHeader.SessionBlockHeight)
	if er != nil {
		return sdk.ErrInternal(er.Error())
	}
	// ensure that session ended
	sessionEndHeight := submitReportcard.SessionHeader.SessionBlockHeight + k.BlocksPerSession(sessionContext) - 1
	if ctx.BlockHeight() <= sessionEndHeight {
		return vc.NewInvalidBlockHeightError(vc.ModuleName)
	}
	node := vc.GetViperNode()
	result, er := vc.GetResult(submitReportcard.SessionHeader, submitReportcard.EvidenceType, submitReportcard.ServicerAddress, node.TestStore)
	if er != nil {
		ctx.Logger().Error(fmt.Sprintf("An error occurred retrieving the result: \n%s", err.Error()))
		return
	}
	if result.NumOfTestResults < k.MinimumSampleRelays(sessionContext) {
		return vc.NewInvalidTestsError(vc.ModuleName)
	}
	// if is not a viper supported blockchain then return not supported error
	if !k.IsViperSupportedBlockchain(sessionContext, submitReportcard.SessionHeader.Chain) {
		return vc.NewChainNotSupportedErr(vc.ModuleName)
	}

	if !k.IsViperSupportedGeoZone(sessionContext, submitReportcard.SessionHeader.GeoZone) {
		return vc.NewGeoZoneNotSupportedErr(vc.ModuleName)
	}
	// get the node from the keeper (at the state of the start of the session)
	_, found := k.GetNode(sessionContext, submitReportcard.FishermanAddress)
	// if not found return not found error
	if !found {
		return vc.NewNodeNotFoundErr(vc.ModuleName)
	}
	// get the application (at the state of the start of the session)
	app, found := k.GetProviderFromPublicKey(sessionContext, submitReportcard.SessionHeader.ProviderPubKey)
	// if not found return not found error
	if !found {
		return vc.NewProviderNotFoundError(vc.ModuleName)
	}
	// get the session node count for the time of the session
	sessionNodeCount := int(app.GetNumServicers())
	// check cache
	session, found := vc.GetSession(submitReportcard.SessionHeader, vc.GlobalSessionCache)
	if !found {
		// use the session end context to ensure that people who were jailed mid session do not get to submit claims
		sessionEndCtx, er := ctx.PrevCtx(sessionEndHeight)
		if er != nil {
			return sdk.ErrInternal("could not get prev context: " + er.Error())
		}
		hash, er := sessionContext.BlockHash(k.Cdc, sessionContext.BlockHeight())
		if er != nil {
			return sdk.ErrInternal(er.Error())
		}
		// create a new session to validate
		session, err = vc.NewSession(sessionContext, sessionEndCtx, k.posKeeper, submitReportcard.SessionHeader, hex.EncodeToString(hash))
		if err != nil {
			ctx.Logger().Error(fmt.Errorf("could not generate session with public key: %s, for chain: %s", app.GetPublicKey().RawString(), submitReportcard.SessionHeader.Chain).Error())
			return err
		}
	}
	// Validate against the root
	valid, _ := k.validateReportCardAgainstRoot(ctx, submitReportcard.Report, submitReportcard.SessionHeader)
	if !valid {
		ctx.Logger().Error(fmt.Sprintf("Report card validation against root failed"))
		return vc.NewInvalidRCMerkleVerifyError(vc.ModuleName)
	}

	// validate the session
	err = session.Validate(submitReportcard.ServicerAddress, app, sessionNodeCount)
	if err != nil {
		return err
	}
	// check if the proof is ready to be claimed, if it's already ready to be claimed, then it's too late to submit cause the secret is revealed
	if k.ReportCardIsExpired(ctx, submitReportcard.SessionHeader.SessionBlockHeight) {
		return vc.NewExpiredReportSubmissionError(vc.ModuleName)
	}
	return nil
}

func (k Keeper) ExecuteReportCard(ctx sdk.Ctx, servicerAddr sdk.Address, reportCard vc.MsgSubmitReportCard) {

	// Check if the report card has expired
	if k.ReportCardIsExpired(ctx, reportCard.SessionHeader.SessionBlockHeight) {
		ctx.Logger().Info(fmt.Sprintf("Report card for validator %s in session %v has expired", servicerAddr.String(), reportCard.SessionHeader))
		return
	}

	// Update the report card
	k.posKeeper.UpdateValidatorReportCard(ctx, servicerAddr, reportCard.Report)

	// Delete the report card
	k.DeleteReportCard(ctx, servicerAddr, reportCard.FishermanAddress, reportCard.SessionHeader)
}

// "SetReportCard" - Sets the report card in the state storage
func (k Keeper) SetReportCard(ctx sdk.Ctx, msg vc.MsgSubmitReportCard) error {
	// retrieve the store
	store := ctx.KVStore(k.storeKey)

	// generate the store key for the report card. Here, I'm assuming a function `KeyForReportCard` similar to `KeyForClaim`.
	key, err := vc.KeyForReportCard(ctx, msg.ServicerAddress, msg.SessionHeader)
	if err != nil {
		return err
	}

	// marshal the report card into amino (or the appropriate codec)
	bz, err := k.Cdc.MarshalBinaryBare(&msg)
	if err != nil {
		panic(err)
	}

	// set in the store
	_ = store.Set(key, bz)
	return nil
}

// GetReportCard retrieves the ReportCard message from the store.
func (k Keeper) GetReportCard(ctx sdk.Ctx, servicerAddr sdk.Address, header vc.SessionHeader) (msg vc.MsgSubmitReportCard, found bool) {
	// Get the store.
	store := ctx.KVStore(k.storeKey)

	// Generate the key for the ReportCard.
	key, err := vc.KeyForReportCard(ctx, servicerAddr, header)
	if err != nil {
		ctx.Logger().Error("Error generating key for report card:", err)
		return vc.MsgSubmitReportCard{}, false
	}

	// Get the report card from the store using the generated key.
	res, _ := store.Get(key)
	if res == nil {
		return vc.MsgSubmitReportCard{}, false
	}

	// Unmarshal the data into the report card object.
	err = k.Cdc.UnmarshalBinaryBare(res, &msg)
	if err != nil {
		panic(err)
	}

	return msg, true
}

func (k Keeper) SetReportCards(ctx sdk.Ctx, reportCards []vc.MsgSubmitReportCard) {
	for _, msg := range reportCards {
		err := k.SetReportCard(ctx, msg)
		if err != nil {
			ctx.Logger().Error("an error occurred setting the report card:\n", msg)
		}
	}
}

func (k Keeper) GetReportCards(ctx sdk.Ctx, servicerAddress sdk.Address) (reportCards []vc.MsgSubmitReportCard, err error) {
	// retrieve the store
	store := ctx.KVStore(k.storeKey)

	key, err := vc.KeyForReportCards(servicerAddress)
	if err != nil {
		return nil, err
	}
	// iterate through all of the kv pairs and unmarshal into claim objects
	iterator, _ := sdk.KVStorePrefixIterator(store, key)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var reportCard vc.MsgSubmitReportCard
		err = k.Cdc.UnmarshalBinaryBare(iterator.Value(), &reportCard)
		if err != nil {
			panic(err)
		}
		reportCards = append(reportCards, reportCard)
	}
	return
}

// "GetAllClaims" - Gets all of the submit report card messages held in the state storage.
func (k Keeper) GetAllReportCards(ctx sdk.Ctx) (reportCards []vc.MsgSubmitReportCard) {
	// retrieve the store
	store := ctx.KVStore(k.storeKey)

	iterator, _ := sdk.KVStorePrefixIterator(store, vc.ReportCardKey)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var reportCard vc.MsgSubmitReportCard
		err := k.Cdc.UnmarshalBinaryBare(iterator.Value(), &reportCard)
		if err != nil {
			panic(err)
		}
		reportCards = append(reportCards, reportCard)
	}
	return
}

func (k Keeper) DeleteReportCard(ctx sdk.Ctx, servicerAddr sdk.Address, fishermanAddr sdk.Address, header vc.SessionHeader) error {
	// retrieve the store
	store := ctx.KVStore(k.storeKey)
	// generate the key for the claim
	key, err := vc.KeyForReportCard(ctx, servicerAddr, header)
	if err != nil {
		return err
	}
	// delete it from the state storage
	_ = store.Delete(key)
	return nil
}

func (k Keeper) ReportCardIsExpired(ctx sdk.Ctx, sessionBlockHeight int64) bool {
	expirationWindowInBlocks := k.ReportCardSubmissionWindow(ctx) * k.BlocksPerSession(ctx)
	return ctx.BlockHeight() > expirationWindowInBlocks+sessionBlockHeight
}

// Function to validate the report card against the root
func (k Keeper) validateReportCardAgainstRoot(ctx sdk.Ctx, reportCard vc.ViperQoSReport, sessionHeader vc.SessionHeader) (bool, error) {

	node := vc.GetViperNode()

	sessionCtx, _ := ctx.PrevCtx(sessionHeader.SessionBlockHeight)
	// Retrieve the evidence object
	result, err := vc.GetResult(sessionHeader, vc.FishermanTestEvidence, reportCard.ServicerAddress, node.TestStore)
	if err != nil {
		return false, fmt.Errorf("error retrieving evidence: %v", err)
	}

	// Get the Merkle proof object for the report card
	index, err := k.getPseudorandomIndex(ctx, result.NumOfTestResults, sessionHeader, sessionCtx)
	if err != nil {
		return false, fmt.Errorf("error getting pseudorandom index: %v", err)
	}

	// Generate the Merkle proof object for the index
	mProof, leaf := result.GenerateMerkleProof(sessionHeader.SessionBlockHeight, int(index))

	// Validate the Merkle proof
	levelCount := len(mProof.HashRanges)
	if levelCount != int(math.Ceil(math.Log2(float64(result.NumOfTestResults)))) {
		return false, fmt.Errorf("produced invalid proof for report card validation, level count")
	}

	if isValid, _ := mProof.ValidateTR(sessionHeader.SessionBlockHeight, reportCard.SampleRoot, leaf, levelCount); !isValid {
		return false, fmt.Errorf("produced invalid proof for report card validation")
	}

	return true, nil
}

func (k Keeper) HandleFishermanSlash(ctx sdk.Ctx, address sdk.Address) {
	ctx.Logger().Error(fmt.Sprintf("Invalid Report Card Detected: By %s", address.String()))
	k.posKeeper.SlashFisherman(ctx, address)
}
