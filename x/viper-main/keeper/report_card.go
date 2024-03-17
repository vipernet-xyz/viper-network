package keeper

import (
	rand1 "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"time"

	"github.com/tendermint/tendermint/rpc/client"
	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	auth "github.com/vipernet-xyz/viper-network/x/authentication"
	"github.com/vipernet-xyz/viper-network/x/authentication/util"
	servicersTypes "github.com/vipernet-xyz/viper-network/x/servicers/types"
	vc "github.com/vipernet-xyz/viper-network/x/viper-main/types"
)

func (k Keeper) SendReportCardTx(ctx sdk.Ctx, keeper Keeper, n client.Client, node *vc.ViperNode, reportCardTx func(pk crypto.PrivateKey, cliCtx util.CLIContext, txBuilder auth.TxBuilder, header vc.SessionHeader, servicerAddr sdk.Address, reportCard vc.ViperQoSReport, reportMerkleProof vc.MerkleProof, reportLeafNode vc.TestI, numOfTestResults int64, evidenceType vc.EvidenceType) (*sdk.TxResponse, error)) {
	// Iterate through the result iterator
	iter := vc.ResultIterator(node.TestStore)
	defer iter.Close()

	// Map to store servicer results for each session header and servicer address
	sessionServicerResults := make(map[vc.SessionHeader]map[string]*vc.ServicerResults)
	// Map to store merkle roots for each servicer
	merkleroot := make(map[string]vc.HashRange)
	// Iterate over test results and group them by session header and servicer address
	for ; iter.Valid(); iter.Next() {
		// Get the result
		result := iter.Value()
		// Get the session header and servicer address
		sessionHeader := result.SessionHeader
		servicerAddr := result.ServicerAddr.String()

		sessionCtx, er := ctx.PrevCtx(result.SessionBlockHeight)
		if er != nil {
			ctx.Logger().Info("could not get sessionCtx in auto send claim tx, could be due to relay timing before commit is in store: " + er.Error())
			continue
		}

		// Check if the current session is still ongoing
		if ctx.BlockHeight() <= result.SessionHeader.SessionBlockHeight+k.BlocksPerSession(sessionCtx)-1 {
			ctx.Logger().Info("The session is ongoing, so will not send the report card yet.")
			continue
		}

		// if the blockchain in the evidence is not supported then delete it because nodes don't get paid/challenged for unsupported blockchains
		if !k.IsViperSupportedBlockchain(sessionCtx.WithBlockHeight(result.SessionHeader.SessionBlockHeight), result.SessionHeader.Chain) {
			ctx.Logger().Info(fmt.Sprintf("report card for %s blockchain isn't viper supported, so will not send. Deleting reportcard\n", result.SessionHeader.Chain))
			if err := vc.DeleteResult(result.SessionHeader, result.EvidenceType, result.ServicerAddr, node.TestStore); err != nil {
				ctx.Logger().Debug(err.Error())
			}
			continue
		}
		if !k.IsViperSupportedGeoZone(sessionCtx.WithBlockHeight(result.SessionHeader.SessionBlockHeight), result.SessionHeader.GeoZone) {
			ctx.Logger().Info(fmt.Sprintf("report card for %s Geozone isn't viper supported, so will not send. Deleting reportcard\n", result.SessionHeader.GeoZone))
			if err := vc.DeleteResult(result.SessionHeader, result.EvidenceType, result.ServicerAddr, node.TestStore); err != nil {
				ctx.Logger().Debug(err.Error())
			}
			continue
		}

		// Check the current state to see if the report card has already been sent and processed (if so, then return)
		if _, found := k.GetReportCard(ctx, result.ServicerAddr, result.SessionHeader, result.EvidenceType); found {
			continue
		}

		// Check if the report card has expired
		if k.ReportCardIsExpired(ctx, result.SessionHeader.SessionBlockHeight) {
			// Delete the result since we cannot submit an expired report card
			if err := vc.DeleteResult(result.SessionHeader, result.EvidenceType, result.ServicerAddr, node.TestStore); err != nil {
				ctx.Logger().Debug(err.Error())
			}
			continue
		}

		merkleroot[servicerAddr] = result.GenerateSampleMerkleRoot(result.SessionHeader.SessionBlockHeight, node.TestStore)

		// Initialize or get the map for the session header
		sessionResults, found := sessionServicerResults[sessionHeader]
		if !found {
			sessionResults = make(map[string]*vc.ServicerResults)
			sessionServicerResults[sessionHeader] = sessionResults
		}

		// Initialize or get the ServicerResults for the servicer address
		sr, found := sessionResults[servicerAddr]
		if !found {
			sr = &vc.ServicerResults{
				ServicerAddress: result.ServicerAddr,
				Timestamps:      make([]time.Time, 0),
				Latencies:       make([]time.Duration, 0),
				Availabilities:  make([]bool, 0),
				Reliabilities:   make([]bool, 0),
			}
			sessionResults[servicerAddr] = sr
		}

		// Loop through the test results and populate ServicerResults
		for _, testResult := range result.TestResults {
			bz := testResult.Bytes()
			var tr vc.TestResult
			json.Unmarshal(bz, &tr)
			sr.Timestamps = append(sr.Timestamps, tr.Timestamp)
			sr.Latencies = append(sr.Latencies, tr.Latency)
			sr.Availabilities = append(sr.Availabilities, tr.IsAvailable)
			sr.Reliabilities = append(sr.Reliabilities, tr.IsReliable)
		}
	}

	// Process each session's test results
	for sessionHeader, results := range sessionServicerResults {
		// Calculate latency scores for the session

		latencyScores := CalculateLatencyScores(results)

		// Process each servicer's test results in the session
		for servicerAddr, sr := range results {
			// Get latency score for the servicer
			latencyScore := latencyScores[servicerAddr]
			// Calculate availability, reliability, and other scores
			qosReport, err := vc.CalculateQoSForServicer(results[servicerAddr], latencyScore)
			if err != nil {
				ctx.Logger().Error(fmt.Sprintf("QoS Report could not be created for %s", servicerAddr))
			}
			qosReport.BlockHeight = sessionHeader.SessionBlockHeight
			qosReport.ServicerAddress = sr.ServicerAddress
			qosReport.SampleRoot = merkleroot[servicerAddr]
			nonce, _ := rand1.Int(rand1.Reader, big.NewInt(math.MaxInt64))
			qosReport.Nonce = nonce.Int64()

			selfPk := node.PrivateKey.RawString()
			signer, err := vc.NewSignerFromPrivateKey(selfPk)
			if err != nil {
				ctx.Logger().Error("Error creating signer")
			}

			signature, err := signer.GetSignedReportBytes(qosReport)
			if err != nil {
				ctx.Logger().Error(fmt.Sprintf("QoS Report could not be signed:%s", err))
			}
			qosReport.Signature = signature

			// Get the Merkle proof object for the report card
			reportMProof, reportLeaf, numOfTestResults, err := k.validateReportCardAgainstRoot(ctx, *qosReport, sessionHeader)
			if err != nil {
				ctx.Logger().Error(fmt.Sprintf("could not validate reportcard againts root:%s", err))
			}

			// Generate the auto tx builder and cli ctx
			txBuilder, cliCtx, err := newTxBuilderAndCliCtx(ctx, &vc.MsgSubmitQoSReport{}, n, node.PrivateKey, k)
			if err != nil {
				ctx.Logger().Error(fmt.Sprintf("An error occurred creating the tx builder for the report card tx:\n%s", err.Error()))
				continue
			}

			// Send in the report card
			if _, err := reportCardTx(node.PrivateKey, cliCtx, txBuilder, sessionHeader, sr.ServicerAddress, *qosReport, reportMProof, reportLeaf, numOfTestResults, vc.FishermanTestEvidence); err != nil {
				ctx.Logger().Error(fmt.Sprintf("An error occurred executing the report card transaction: \n%s", err.Error()))
			}
		}
	}
}

func (k Keeper) ValidateSumbitReportCard(ctx sdk.Ctx, submitReportcard vc.MsgSubmitQoSReport) (err sdk.Error) {
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
	app, found := k.GetRequestorFromPublicKey(sessionContext, submitReportcard.SessionHeader.RequestorPubKey)
	// if not found return not found error
	if !found {
		return vc.NewRequestorNotFoundError(vc.ModuleName)
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
			ctx.Logger().Error(fmt.Errorf("could not generate session with public key: %s, for chain: %s", app.GetPublicKey().RawString(), &submitReportcard.SessionHeader.Chain).Error())
			return err
		}
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

func (k Keeper) UpdateReportCard(ctx sdk.Ctx, servicerAddr sdk.Address, reportCard vc.MsgSubmitQoSReport, evidenceType vc.EvidenceType) servicersTypes.ReportCard {
	// Update the report crd
	updatedRC := k.posKeeper.UpdateValidatorReportCard(ctx, servicerAddr, reportCard.Report)

	// Delete the report card
	k.DeleteReportCard(ctx, servicerAddr, reportCard.FishermanAddress, reportCard.SessionHeader, evidenceType)

	return updatedRC
}

// "SetReportCard" - Sets the report card in the state storage
func (k Keeper) SetReportCard(ctx sdk.Ctx, msg vc.MsgSubmitQoSReport) error {
	// retrieve the store
	store := ctx.KVStore(k.storeKey)
	// generate the store key for the report card
	key, err := vc.KeyForReportCard(ctx, msg.ServicerAddress, msg.SessionHeader, msg.EvidenceType)
	if err != nil {
		return err
	}
	// marshal the report card into amino (or the appropriate codec)
	bz, err := k.Cdc.MarshalBinaryBare(&msg)
	if err != nil {
		panic(err)
	}
	err = store.Set(key, bz)
	if err != nil {
		return err
	}
	return nil
}

// GetReportCard retrieves the ReportCard message from the store.
func (k Keeper) GetReportCard(ctx sdk.Ctx, servicerAddr sdk.Address, header vc.SessionHeader, evidenceType vc.EvidenceType) (msg vc.MsgSubmitQoSReport, found bool) {
	// Get the store.
	store := ctx.KVStore(k.storeKey)

	// Generate the key for the ReportCard.
	key, err := vc.KeyForReportCard(ctx, servicerAddr, header, evidenceType)

	if err != nil {
		ctx.Logger().Error("Error generating key for report card:", err)
		return vc.MsgSubmitQoSReport{}, false
	}

	// Get the report card from the store using the generated key.
	res, _ := store.Get(key)
	if res == nil {
		return vc.MsgSubmitQoSReport{}, false
	}

	// Unmarshal the data into the report card object.
	err = k.Cdc.UnmarshalBinaryBare(res, &msg)
	if err != nil {
		panic(err)
	}

	return msg, true
}

func (k Keeper) SetReportCards(ctx sdk.Ctx, reportCards []vc.MsgSubmitQoSReport) {
	for _, msg := range reportCards {
		err := k.SetReportCard(ctx, msg)
		if err != nil {
			ctx.Logger().Error("an error occurred setting the report card:\n", msg)
		}
	}
}

func (k Keeper) GetReportCards(ctx sdk.Ctx, servicerAddress sdk.Address) (reportCards []vc.MsgSubmitQoSReport, err error) {
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
		var reportCard vc.MsgSubmitQoSReport
		err = k.Cdc.UnmarshalBinaryBare(iterator.Value(), &reportCard)
		if err != nil {
			panic(err)
		}
		reportCards = append(reportCards, reportCard)
	}
	return
}

// "GetAllReportCards" - Gets all of the submit report card messages held in the state storage.
func (k Keeper) GetAllReportCards(ctx sdk.Ctx) (reportCards []vc.MsgSubmitQoSReport) {
	// retrieve the store
	store := ctx.KVStore(k.storeKey)

	iterator, _ := sdk.KVStorePrefixIterator(store, vc.ReportCardKey)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var reportCard vc.MsgSubmitQoSReport
		err := k.Cdc.UnmarshalBinaryBare(iterator.Value(), &reportCard)
		if err != nil {
			panic(err)
		}
		reportCards = append(reportCards, reportCard)
	}
	return
}

func (k Keeper) DeleteReportCard(ctx sdk.Ctx, servicerAddr sdk.Address, fishermanAddr sdk.Address, header vc.SessionHeader, evidenceType vc.EvidenceType) error {
	// retrieve the store
	store := ctx.KVStore(k.storeKey)
	// generate the key for the claim
	key, err := vc.KeyForReportCard(ctx, servicerAddr, header, evidenceType)
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
func (k Keeper) validateReportCardAgainstRoot(ctx sdk.Ctx, reportCard vc.ViperQoSReport, sessionHeader vc.SessionHeader) (vc.MerkleProof, vc.TestI, int64, error) {

	node := vc.GetViperNode()

	sessionCtx, _ := ctx.PrevCtx(sessionHeader.SessionBlockHeight)
	// Retrieve the evidence object
	result, err := vc.GetResult(sessionHeader, vc.FishermanTestEvidence, reportCard.ServicerAddress, node.TestStore)
	if err != nil {
		return vc.MerkleProof{}, vc.TestI{}, 0, err
	}
	// Get the Merkle proof object for the report card
	index, err := k.getPseudorandomIndexForRC(ctx, result.NumOfTestResults, sessionHeader, sessionCtx)
	if err != nil {
		return vc.MerkleProof{}, vc.TestI{}, 0, err
	}

	// Generate the Merkle proof object for the index
	mProof, leaf := result.GenerateMerkleProof(sessionHeader.SessionBlockHeight, int(index))

	// Validate the Merkle proof
	levelCount := len(mProof.HashRanges)
	if levelCount != int(math.Ceil(math.Log2(float64(result.NumOfTestResults)))) {

		return vc.MerkleProof{}, vc.TestI{}, 0, err
	}
	if isValid, _ := mProof.ValidateTR(sessionHeader.SessionBlockHeight, reportCard.SampleRoot, leaf, levelCount); !isValid {
		return vc.MerkleProof{}, vc.TestI{}, 0, err
	}

	return mProof, leaf.ToProto(), result.NumOfTestResults, nil
}

func (k Keeper) HandleFishermanSlash(ctx sdk.Ctx, sessionHeader vc.SessionHeader, height int64) {
	node := vc.GetViperNode()
	session, found := vc.GetSession(sessionHeader, node.SessionStore)
	if !found {
		// get the session context (state info at the beginning of the session)
		sessionContext, _ := ctx.PrevCtx(sessionHeader.SessionBlockHeight)
		sessionEndHeight := sessionHeader.SessionBlockHeight + k.BlocksPerSession(sessionContext) - 1
		// use the session end context to ensure that people who were jailed mid session do not get to submit claims
		sessionEndCtx, _ := ctx.PrevCtx(sessionEndHeight)

		app, _ := k.GetRequestorFromPublicKey(sessionContext, sessionHeader.RequestorPubKey)
		hash, _ := sessionContext.BlockHash(k.Cdc, sessionContext.BlockHeight())

		// create a new session to validate
		session, err := vc.NewSession(sessionContext, sessionEndCtx, k.posKeeper, sessionHeader, hex.EncodeToString(hash))
		if err != nil {
			ctx.Logger().Error(fmt.Errorf("could not generate session with public key: %s, for chain: %s", app.GetPublicKey().RawString(), sessionHeader.Chain).Error())
			return
		}
		k.posKeeper.SlashFisherman(ctx, height, session.SessionFishermen[0])
	}
	k.posKeeper.SlashFisherman(ctx, height, session.SessionFishermen[0])
}

func (k Keeper) GetMatureReportCards(ctx sdk.Ctx, address sdk.Address) (matureReportCards []vc.MsgSubmitQoSReport, err error) {
	// retrieve the store
	store := ctx.KVStore(k.storeKey)
	// generate the key for the claim
	key, err := vc.KeyForReportCards(address)
	if err != nil {
		return nil, err
	}
	iterator, _ := sdk.KVStorePrefixIterator(store, key)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var msg vc.MsgSubmitQoSReport
		err = k.Cdc.UnmarshalBinaryBare(iterator.Value(), &msg)
		if err != nil {
			panic(err)
		}
		matureReportCards = append(matureReportCards, msg)
		if k.ReportCardIsMature(ctx, msg.SessionHeader.SessionBlockHeight) {
			matureReportCards = append(matureReportCards, msg)
		}
	}
	return
}

func (k Keeper) ReportCardIsMature(ctx sdk.Ctx, sessionBlockHeight int64) bool {
	waitingPeriodInBlocks := k.ReportCardSubmissionWindow(ctx) * k.BlocksPerSession(ctx)
	return ctx.BlockHeight() > waitingPeriodInBlocks+sessionBlockHeight
}

// generates the required pseudorandom index for the zero knowledge proof
func (k Keeper) getPseudorandomIndexForRC(ctx sdk.Ctx, totalSampleRelays int64, header vc.SessionHeader, sessionCtx sdk.Ctx) (int64, error) {
	rcHeight := header.SessionBlockHeight + k.BlocksPerSession(sessionCtx)
	// get the pseudorandomGenerator json bytes
	blockHashBz, err := ctx.GetPrevBlockHash(rcHeight)
	if err != nil {
		return 0, err
	}
	headerHash := header.HashString()
	pseudoGenerator := pseudorandomGenerator{hex.EncodeToString(blockHashBz), headerHash}
	r, err := json.Marshal(pseudoGenerator)
	if err != nil {
		return 0, err
	}
	return vc.PseudorandomSelection(sdk.NewInt(totalSampleRelays), vc.Hash(r)).Int64(), nil
}
