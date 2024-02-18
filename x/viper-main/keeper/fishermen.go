package keeper

import (
	"bytes"
	rand1 "crypto/rand"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"sort"
	"time"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/servicers/exported"
	vc "github.com/vipernet-xyz/viper-network/x/viper-main/types"
)

func (k Keeper) HandleFishermanTrigger(ctx sdk.Ctx, trigger vc.FishermenTrigger) (*vc.RelayResponse, sdk.Error) {
	// Check if the trigger is empty, and if so, return the response without triggering sampling.
	if triggerIsEmpty(trigger) {
		resp := &vc.RelayResponse{
			Response: "fisherman could not be triggered",
		}
		return resp, nil
	}

	// Start by creating the response.
	resp := &vc.RelayResponse{
		Response: "fisherman triggered",
	}

	// Attempt to sign the response.
	node := vc.GetViperNode()
	sig, er := node.PrivateKey.Sign(resp.Hash())
	if er != nil {
		ctx.Logger().Error(
			fmt.Sprintf("could not sign response for address: %s with hash: %v, with error: %s",
				node.GetAddress().String(), resp.HashString(), er.Error()),
		)
		return nil, vc.NewKeybaseError(vc.ModuleName, er)
	}

	// Attach the signature in hex to the response.
	resp.Signature = hex.EncodeToString(sig)
	resp.Proof = trigger.Proof

	// Introduce a longer delay before triggering sampling.
	minDelay := 5000
	maxDelay := 8000
	delay := time.Duration(rand.Intn(maxDelay-minDelay) + minDelay)
	time.Sleep(delay * time.Millisecond)

	// If everything has gone well so far, call StartServicersSampling.
	err := k.StartServicersSampling(ctx, trigger)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func triggerIsEmpty(trigger vc.FishermenTrigger) bool {
	// Check if trigger fields are empty or zero values based on your trigger definition.
	return trigger.Proof.SessionBlockHeight == 0 &&
		trigger.Proof.NumServicers == 0 &&
		trigger.Proof.Blockchain == "" &&
		trigger.Proof.ServicerPubKey == "" &&
		trigger.Proof.Token.RequestorPublicKey == "" &&
		trigger.Proof.Token.ClientPublicKey == "" &&
		trigger.Proof.Token.RequestorSignature == ""
}

func (k Keeper) StartServicersSampling(ctx sdk.Ctx, trigger vc.FishermenTrigger) sdk.Error {
	sessionHeader := vc.SessionHeader{
		RequestorPubKey:    trigger.Proof.Token.RequestorPublicKey,
		Chain:              trigger.Proof.Blockchain,
		GeoZone:            trigger.Proof.GeoZone,
		NumServicers:       trigger.Proof.NumServicers,
		SessionBlockHeight: trigger.Proof.SessionBlockHeight,
	}

	latestSessionBlockHeight := k.GetLatestSessionBlockHeight(ctx)
	sessionCtx, er := ctx.PrevCtx(latestSessionBlockHeight)
	if er != nil {
		return sdk.ErrInternal(er.Error())
	}
	session, found := vc.GetSession(sessionHeader, vc.GlobalSessionCache)
	if !found {
		return sdk.ErrInternal("Session not found")
	}

	actualServicers := make([]exported.ValidatorI, len(session.SessionServicers))
	for i, addr := range session.SessionServicers {
		actualServicers[i], _ = k.GetNode(sessionCtx, addr)
	}

	fisherman := vc.GetViperNode()
	fishermanAddr := fisherman.GetAddress()

	hostedBlockchains := k.GetHostedBlockchains()
	fishermanValidator, _ := k.GetNode(ctx, fishermanAddr)
	results := make(map[string]*vc.ServicerResults)

	rpcURL := fishermanValidator.GetServiceURL()

	sender := vc.NewSender(rpcURL, []string{rpcURL})
	selfPk := fisherman.PrivateKey.RawString()
	signer, err := vc.NewSignerFromPrivateKey(selfPk)
	if err != nil {
		return sdk.ErrInternal("Error creating signer")
	}

	for _, servicer := range actualServicers {
		servicerAddrStr := servicer.GetAddress().String()
		results[servicerAddrStr] = &vc.ServicerResults{}
	}

	// Function to send sample relays
	sendSampleRelays := func() {
		for _, servicer := range actualServicers {
			relayer := vc.NewRelayer(*signer, *sender)
			startTime := time.Now()
			Blockchain := trigger.Proof.Blockchain

			resp, err := relayer.SendSampleRelay(sessionHeader.SessionBlockHeight, Blockchain, trigger, servicer, fishermanValidator, hostedBlockchains)
			latency := resp.Latency
			isAvailable := err == nil && resp.Proof.Signature != ""
			isReliable := resp.Reliability

			servicerResult := results[servicer.GetAddress().String()]
			servicerResult.Timestamps = append(servicerResult.Timestamps, startTime)
			servicerResult.Latencies = append(servicerResult.Latencies, latency)
			servicerResult.Availabilities = append(servicerResult.Availabilities, isAvailable)
			servicerResult.Reliabilities = append(servicerResult.Reliabilities, isReliable)

			if len(servicerResult.Availabilities) >= 5 && !anyTrue(servicerResult.Availabilities[len(servicerResult.Availabilities)-5:]) {
				k.posKeeper.BurnforNoActivity(ctx, ctx.BlockHeight(), servicer.GetAddress())
				k.posKeeper.PauseNode(ctx, servicer.GetAddress())
			}

			testResult := vc.TestResult{
				ServicerAddress: servicer.GetAddress(),
				Timestamp:       startTime,
				Latency:         latency,
				IsAvailable:     isAvailable,
				IsReliable:      isReliable,
			}
			testResult.Store(sessionHeader, fisherman.TestStore)
		}
	}

	// Function to send report card tx
	sendReportCardTx := func() {
		latencyScores := CalculateLatencyScores(results)
		for _, servicer := range actualServicers {
			servicerResult := results[servicer.GetAddress().String()]
			proofs, err := k.GetProofsForServicer(ctx, sessionHeader, servicer.GetAddress(), fisherman.TestStore)
			if err != nil {
				fmt.Errorf("Sample Relay Proofs could not be fetched for %s", servicer)
			}

			seed := time.Now().UnixNano() + int64(ctx.BlockHeight())
			rng := rand.New(rand.NewSource(seed))
			subsetSize := int(float64(len(proofs)) * 0.20)
			if len(proofs) > int(subsetSize) {
				vc.Shuffle(proofs, rng)
				proofs = proofs[:subsetSize]
			}

			resultForMerkle := &vc.Result{
				SessionHeader:    sessionHeader,
				ServicerAddr:     servicer.GetAddress(),
				NumOfTestResults: int64(len(proofs)),
				TestResults:      proofs,
				EvidenceType:     vc.FishermanTestEvidence,
			}

			merkleRoot := resultForMerkle.GenerateSampleMerkleRoot(sessionHeader.SessionBlockHeight, fisherman.TestStore)
			qos, err := vc.CalculateQoSForServicer(servicerResult, sessionHeader.SessionBlockHeight, latencyScores[servicer.GetAddress().String()])
			if err != nil {
				fmt.Errorf("QoS Report could not be created for %s", servicer)
			}

			qos.SampleRoot = merkleRoot

			nonce, err := rand1.Int(rand1.Reader, big.NewInt(math.MaxInt64))
			if err != nil {
				return
			}
			qos.Nonce = nonce.Int64()
			qos.Signature, err = k.SignQoSReport(signer, qos)
			if err != nil {
				fmt.Errorf("QoS Report could not be signed")
			}

			k.SendReportCardTx(ctx, k, k.TmNode, fisherman, qos.ServicerAddress, sessionHeader, resultForMerkle.EvidenceType, *qos, vc.SendReportCardTx)
		}
	}

	// Loop to send sample relays based on time triggers
	go func() {
		ticker := time.NewTicker(time.Duration(10+rand.Intn(20)) * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			blockHeight, _ := sender.GetBlockHeight()
			if int64(blockHeight) >= (trigger.Proof.SessionBlockHeight + k.posKeeper.BlocksPerSession(ctx)) {
				sendReportCardTx()
				return
			} else {
				sendSampleRelays()
			}
		}
	}()

	return nil
}

// Utility function to check if any value in the provided slice is true
func anyTrue(booleans []bool) bool {
	for _, b := range booleans {
		if b {
			return true
		}
	}
	return false
}

func (k Keeper) GetProofsForServicer(ctx sdk.Ctx, header vc.SessionHeader, servicerAddr sdk.Address, testStore *vc.CacheStorage) ([]vc.Test, error) {
	var proofs []vc.Test
	fisherman := vc.GetViperNode()
	iter := vc.ResultIterator(fisherman.TestStore)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		result := iter.Value()
		fmt.Println("result:", result)

		// Check if the result belongs to the current servicer and session
		if result.ServicerAddr.Equals(servicerAddr) && result.SessionHeader == header {
			proofs = append(proofs, result.TestResults...)
		}
	}
	return proofs, nil
}

func init() {
	gob.Register(vc.ViperQoSReport{})
}

func (k Keeper) SignQoSReport(signer *vc.Signer, report *vc.ViperQoSReport) (string, error) {

	// Create a bytes.Buffer to hold our encoded data
	var buf bytes.Buffer
	// Create a new GOB encoder that writes to the buffer
	enc := gob.NewEncoder(&buf)

	// Encode the report
	err := enc.Encode(report)
	if err != nil {
		return "", err
	}

	return signer.Sign(buf.Bytes())
}

func CalculateLatencyScores(results map[string]*vc.ServicerResults) map[string]sdk.BigDec {
	latencyScores := make(map[string]sdk.BigDec)

	// Collect latencies and servicer addresses from results
	var latencies []sdk.BigDec
	var servicerAddresses []string

	for servicerAddr, result := range results {
		if len(result.Latencies) > 0 {
			// Check if any latency value is nil or zero
			if hasNilOrZeroValue(result.Latencies) {
				latencies = append(latencies, sdk.ZeroDec())
			} else {
				// Calculate the average latency for each servicer
				totalLatency := sdk.ZeroDec()

				for _, latency := range result.Latencies {
					totalLatency = totalLatency.Add(sdk.NewDec(int64(latency.Milliseconds())))
				}

				averageLatency := totalLatency.Quo(sdk.NewDec(int64(len(result.Latencies))))
				latencies = append(latencies, averageLatency)
			}

			servicerAddresses = append(servicerAddresses, servicerAddr)
		} else {
			latencyScores[servicerAddr] = sdk.ZeroDec()
		}
	}

	// Rank servicers by latency (lower latency gets a higher rank)
	rankedLatencies := rankLatencies(latencies, servicerAddresses)

	// Assign scores based on rankings
	maxRank := len(rankedLatencies)

	for servicerAddr, rank := range rankedLatencies {
		if maxRank == 0 {
			latencyScores[servicerAddr] = sdk.ZeroDec()
		} else {
			// Assign scores inversely proportional to rank
			score := sdk.NewDec(int64(maxRank - rank + 1)).Quo(sdk.NewDec(int64(maxRank)))
			latencyScores[servicerAddr] = score
		}
	}

	return latencyScores
}

// Function to check if Latencies contain a nil or zero value
func hasNilOrZeroValue(latencies []time.Duration) bool {
	for _, latency := range latencies {
		if latency == 0 || latency == 0*time.Second || latencies == nil {
			return true
		}
	}
	return false
}

func rankLatencies(latencies []sdk.BigDec, servicerNames []string) map[string]int {
	ranks := make(map[string]int)
	servicerRanks := make(map[string]int)

	for i, servicerName := range servicerNames {
		servicerRanks[servicerName] = i
	}

	sort.Slice(servicerNames, func(i, j int) bool {
		nameI := servicerNames[i]
		nameJ := servicerNames[j]
		rankI, okI := servicerRanks[nameI]
		rankJ, okJ := servicerRanks[nameJ]

		if okI && okJ {
			return latencies[rankI].LT(latencies[rankJ])
		} else if okI {
			return true
		} else {
			return false
		}
	})

	for i, servicerName := range servicerNames {
		ranks[servicerName] = i + 1
	}

	return ranks
}
