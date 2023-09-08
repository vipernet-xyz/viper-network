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
	"time"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/servicers/exported"
	vc "github.com/vipernet-xyz/viper-network/x/vipernet/types"
)

func (k Keeper) HandleFishermanTrigger(ctx sdk.Ctx, trigger vc.FishermenTrigger) (*vc.RelayResponse, sdk.Error) {
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

	// If everything has gone well so far, call StartServicersSampling.
	err := k.StartServicersSampling(ctx, trigger)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (k Keeper) StartServicersSampling(ctx sdk.Ctx, trigger vc.FishermenTrigger) sdk.Error {
	// Get session information from the relay
	sessionHeader := vc.SessionHeader{
		ProviderPubKey:     trigger.Proof.Token.ProviderPublicKey,
		Chain:              trigger.Proof.Blockchain,
		GeoZone:            trigger.Proof.GeoZone,
		NumServicers:       int32(trigger.Proof.NumServicers),
		SessionBlockHeight: trigger.Proof.SessionBlockHeight,
	}

	latestSessionBlockHeight := k.GetLatestSessionBlockHeight(ctx)
	// get the session context
	sessionCtx, er := ctx.PrevCtx(latestSessionBlockHeight)
	if er != nil {
		return sdk.ErrInternal(er.Error())
	}
	// Retrieve the session
	session, found := vc.GetSession(sessionHeader, vc.GlobalSessionCache)
	if !found {
		return sdk.ErrInternal("Session not found")
	}

	// Get actual servicers from the session
	actualServicers := make([]exported.ValidatorI, len(session.SessionServicers))
	for i, addr := range session.SessionServicers {
		actualServicers[i], _ = k.GetNode(sessionCtx, addr)
	}
	blocksPerSession := k.posKeeper.BlocksPerSession(ctx)

	var fisherman *vc.ViperNode
	var fishermanAddress sdk.Address
	fisherman = vc.GetViperNode()
	fishermanAddress = fisherman.GetAddress()
	fishermanValidator, _ := k.GetSelfNode(ctx)

	// Map to hold results for all servicers
	results := make(map[string]*vc.ServicerResults)

	// Initialize the results map
	for _, servicer := range actualServicers {
		servicerAddrStr := servicer.GetAddress().String()
		results[servicerAddrStr] = &vc.ServicerResults{}
	}
	go func() {
		ticker := time.NewTicker(time.Duration(10+rand.Intn(25)) * time.Second)
		defer ticker.Stop() // Ensure to stop the ticker once done

		for {
			select {
			case <-ticker.C: // On each tick

				// Check end conditions
				if (ctx.BlockHeight()+int64(fishermanAddress[0]))%blocksPerSession == 1 && ctx.BlockHeight() != 1 {
					// Calculate and send QoS for each servicer after breaking the loop
					for _, servicer := range actualServicers {
						servicerResult := results[servicer.GetAddress().String()]

						// Get proofs for the current servicer
						proofs, err := k.GetProofsForServicer(ctx, sessionHeader, servicer.GetAddress(), fisherman.TestStore)
						if err != nil {
							fmt.Errorf("Sample Relay Proofs could not be fetched for %s", servicer)
						}

						// Use time and block height as seed for random selection
						seed := time.Now().UnixNano() + int64(ctx.BlockHeight())
						rng := rand.New(rand.NewSource(seed))
						subsetSize := int(float64(len(proofs)) * 0.20) // Take 20% of total relays

						if len(proofs) > int(subsetSize) {
							vc.Shuffle(proofs, rng)
							proofs = proofs[:subsetSize]
						}
						// Convert proofs into a Result structure for Merkle Root generation
						resultForMerkle := &vc.Result{
							SessionHeader:    sessionHeader,
							ServicerAddr:     servicer.GetAddress(),
							NumOfTestResults: int64(len(proofs)),
							TestResults:      proofs,
							EvidenceType:     vc.FishermanTestEvidence,
						}

						// Generate Merkle Root using the Result structure
						merkleRoot := resultForMerkle.GenerateSampleMerkleRoot(sessionHeader.SessionBlockHeight, fisherman.TestStore)

						// Assuming your QoS report has a field for the Merkle Root
						qos, err := vc.CalculateQoSForServicer(servicerResult, sessionHeader.SessionBlockHeight)
						qos.SampleRoot = merkleRoot
						if err != nil {
							fmt.Errorf("QoS Report could not be created for %s", servicer)
						}
						//Generate nonce
						nonce, err := rand1.Int(rand1.Reader, big.NewInt(math.MaxInt64))
						if err != nil {
							return
						}
						qos.Nonce = nonce.Int64()
						signer, _ := vc.NewSigner(fishermanValidator)
						// Sign the QoS report
						qos.Signature, err = k.SignQoSReport(signer, qos)
						if err != nil {
							fmt.Errorf("QoS Report could not be signed")
						}

						// Send the QoS to the servicer.
						k.sendQoSToServicer(ctx, servicer, qos)
					}

					return
				}

				// Loop over each servicer in the session
				for _, servicer := range actualServicers {
					startTime := time.Now()
					Blockchain := trigger.Proof.Blockchain
					resp, err := vc.SendSampleRelay(Blockchain, trigger, servicer, fishermanValidator)

					latency := time.Since(startTime)

					isAvailable := err == nil && resp.Proof.Signature != ""

					servicerResult := results[servicer.GetAddress().String()]
					servicerResult.Timestamps = append(servicerResult.Timestamps, startTime)
					servicerResult.Latencies = append(servicerResult.Latencies, latency)
					servicerResult.Availabilities = append(servicerResult.Availabilities, isAvailable)

					// If the last 5 results show the servicer missed signing 5 times consecutively, pause the node.
					if len(servicerResult.Availabilities) >= 5 && !anyTrue(servicerResult.Availabilities[len(servicerResult.Availabilities)-5:]) {
						k.posKeeper.BurnforNoActivity(ctx, servicer.GetAddress())
						k.posKeeper.PauseNode(ctx, servicer.GetAddress())
					}

					// Store the test result after generating it
					testResult := vc.TestResult{
						ServicerAddress: servicer.GetAddress(),
						Timestamp:       startTime,
						Latency:         latency,
						IsAvailable:     isAvailable,
					}

					testResult.Store(sessionHeader, fisherman.TestStore)
				}
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
	keyPrefix, err := vc.KeyForTestResult(header, vc.FishermanTestEvidence, servicerAddr)
	if err != nil {
		return nil, err
	}
	const maxResultSize = 1024
	iterator, _ := testStore.Iterator()
	for ; iterator.Valid(); iterator.Next() {
		// Check if key starts with keyPrefix
		if bytes.HasPrefix(iterator.Key(), keyPrefix) {
			var result vc.Result
			err := vc.ModuleCdc.UnmarshalBinaryBare(iterator.Value(), &result, maxResultSize)
			if err == nil {
				proofs = append(proofs, result.TestResults...)
			}
		}
	}
	return proofs, nil
}

func (k Keeper) sendQoSToServicer(ctx sdk.Ctx, servicer exported.ValidatorI, qos *vc.ViperQoSReport) error {
	// Logic to send QoS to the servicer. This will be chain specific.
	// For now, a placeholder. You would replace this with actual logic.
	return nil
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
