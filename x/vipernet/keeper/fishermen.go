package keeper

import (
	"encoding/hex"
	"fmt"
	"time"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/servicers/exported"
	vc "github.com/vipernet-xyz/viper-network/x/vipernet/types"
)

func (k Keeper) HandleFishermanTrigger(ctx sdk.Ctx, relay vc.Relay) (*vc.RelayResponse, sdk.Error) {
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
	err := k.HandleFishermanRelay(ctx, relay)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (k Keeper) HandleFishermanRelay(ctx sdk.Ctx, relay vc.Relay) sdk.Error {
	// Get session information from the relay
	sessionHeader := vc.SessionHeader{
		ProviderPubKey:     relay.Proof.Token.ProviderPublicKey,
		Chain:              relay.Proof.Blockchain,
		GeoZone:            relay.Proof.GeoZone,
		NumServicers:       int8(relay.Proof.NumServicers),
		SessionBlockHeight: relay.Proof.SessionBlockHeight,
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
	// Run this loop continuously
	for {
		relayData := []vc.FishermanRelay{}

		if (ctx.BlockHeight()+int64(fishermanAddress[0]))%blocksPerSession == 1 && ctx.BlockHeight() != 1 {
			break
		}

		// Loop over each servicer in the session
		for _, servicer := range actualServicers {

			// Send the relay
			startTime := time.Now()
			// write the sendRelay function
			//
			resp, err := sendRelay(node, relay)

			latency := time.Since(startTime)

			// Check the response
			isAvailable := false
			if err == nil {
				isAvailable = resp.Signature != ""
			}

			//write code to store param as relay proof,
			relayData = append(relayData, vc.FishermanRelay{
				ServicerAddr: servicer.GetAddress(),
				Latency:      latency,
				IsSigned:     isAvailable,
			})
		}

		// Pause for 1 minute
		time.Sleep(1 * time.Minute)
	}

	return nil
}
