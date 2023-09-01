package keeper

import (
	"encoding/hex"
	"fmt"
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

	// If everything has gone well so far, call HandleFishermanRelay.
	err := k.StartServicersSampling(ctx, trigger)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Rewrite the func, according to the specs
func (k Keeper) StartServicersSampling(ctx sdk.Ctx, trigger vc.FishermenTrigger) sdk.Error {
	// Get session information from the relay
	sessionHeader := vc.SessionHeader{
		ProviderPubKey:     trigger.Proof.Token.ProviderPublicKey,
		Chain:              trigger.Proof.Blockchain,
		GeoZone:            trigger.Proof.GeoZone,
		NumServicers:       int8(trigger.Proof.NumServicers),
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

	sampleRelayCount := 0
	go func() {
		ticker := time.NewTicker(time.Duration(10+rand.Intn(25)) * time.Second)
		defer ticker.Stop() // Ensure to stop the ticker once done

		for {
			select {
			case <-ticker.C: // On each tick

				// Check end conditions
				if (ctx.BlockHeight()+int64(fishermanAddress[0]))%blocksPerSession == 1 && ctx.BlockHeight() != 1 {
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

				sampleRelayCount++

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
