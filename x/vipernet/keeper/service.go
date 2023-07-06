package keeper

import (
	"encoding/hex"
	"fmt"
	"time"

	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	vc "github.com/vipernet-xyz/viper-network/x/vipernet/types"
)

// HandleRelay handles an api (read/write) request to a non-native (external) blockchain
func (k Keeper) HandleRelay(ctx sdk.Ctx, relay vc.Relay) (*vc.RelayResponse, sdk.Error) {
	relayTimeStart := time.Now()
	// get the latest session block height because this relay will correspond with the latest session
	sessionBlockHeight := k.GetLatestSessionBlockHeight(ctx)
	var node *vc.ViperNode
	// There is reference to node address so that way we don't have to recreate address twice for pre-leanvipr
	var nodeAddress sdk.Address

	if vc.GlobalViperConfig.LeanViper {
		// if lean viper enabled, grab the targeted servicer through the relay proof
		servicerRelayPublicKey, err := crypto.NewPublicKey(relay.Proof.ServicerPubKey)
		if err != nil {
			return nil, sdk.ErrInternal("Could not convert servicer hex to public key")
		}
		nodeAddress = sdk.GetAddress(servicerRelayPublicKey)
		node, err = vc.GetViperNodeByAddress(&nodeAddress)
		if err != nil {
			return nil, sdk.ErrInternal("Failed to find correct servicer PK")
		}
	} else {
		// get self node (your validator) from the current state
		node = vc.GetViperNode()
		nodeAddress = node.GetAddress()
	}

	// retrieve the nonNative blockchains your node is hosting
	hostedBlockchains := k.GetHostedBlockchains()
	// ensure the validity of the relay
	maxPossibleRelays, err := relay.Validate(ctx, k.posKeeper, k.providerKeeper, k, hostedBlockchains, sessionBlockHeight, node)
	if err != nil {
		if vc.GlobalViperConfig.RelayErrors {
			ctx.Logger().Error(
				fmt.Sprintf("could not validate relay for app: %s for chainID: %v with error: %s",
					relay.Proof.ServicerPubKey,
					relay.Proof.Blockchain,
					err.Error(),
				),
			)
			ctx.Logger().Debug(
				fmt.Sprintf(
					"could not validate relay for app: %s, for chainID %v on node %s, at session height: %v, with error: %s",
					relay.Proof.ServicerPubKey,
					relay.Proof.Blockchain,
					nodeAddress.String(),
					sessionBlockHeight,
					err.Error(),
				),
			)
		}
		return nil, err
	}
	// store the proof before execution, because the proof corresponds to the previous relay
	relay.Proof.Store(maxPossibleRelays, node.EvidenceStore)
	// attempt to execute
	respPayload, err := relay.Execute(hostedBlockchains, &nodeAddress)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("could not send relay with error: %s", err.Error()))
		return nil, err
	}
	// generate response object
	resp := &vc.RelayResponse{
		Response: respPayload,
		Proof:    relay.Proof,
	}
	// sign the response
	sig, er := node.PrivateKey.Sign(resp.Hash())
	if er != nil {
		ctx.Logger().Error(
			fmt.Sprintf("could not sign response for address: %s with hash: %v, with error: %s",
				nodeAddress.String(), resp.HashString(), er.Error()),
		)
		return nil, vc.NewKeybaseError(vc.ModuleName, er)
	}
	// attach the signature in hex to the response
	resp.Signature = hex.EncodeToString(sig)
	// track the relay time
	relayTime := time.Since(relayTimeStart)
	// add to metrics
	addRelayMetricsFunc := func() {
		vc.GlobalServiceMetric().AddRelayTimingFor(relay.Proof.Blockchain, float64(relayTime.Milliseconds()), &nodeAddress)
		vc.GlobalServiceMetric().AddRelayFor(relay.Proof.Blockchain, &nodeAddress)
	}
	if vc.GlobalViperConfig.LeanViper {
		go addRelayMetricsFunc()
	} else {
		addRelayMetricsFunc()
	}
	return resp, nil
}

// "HandleChallenge" - Handles a client relay response challenge request
func (k Keeper) HandleChallenge(ctx sdk.Ctx, challenge vc.ChallengeProofInvalidData) (*vc.ChallengeResponse, sdk.Error) {

	var node *vc.ViperNode
	// There is reference to self address so that way we don't have to recreate address twice for pre-leanvipr
	var nodeAddress sdk.Address

	if vc.GlobalViperConfig.LeanViper {
		// try to retrieve a ViperNode that was part of session
		for _, r := range challenge.MajorityResponses {
			servicerRelayPublicKey, err := crypto.NewPublicKey(r.Proof.ServicerPubKey)
			if err != nil {
				continue
			}
			potentialNodeAddress := sdk.GetAddress(servicerRelayPublicKey)
			potentialNode, err := vc.GetViperNodeByAddress(&nodeAddress)
			if err != nil || potentialNode == nil {
				continue
			}
			node = potentialNode
			nodeAddress = potentialNodeAddress
			break
		}
		if node == nil {
			return nil, vc.NewNodeNotInSessionError(vc.ModuleName)
		}
	} else {
		node = vc.GetViperNode()
		nodeAddress = node.GetAddress()
	}

	sessionBlkHeight := k.GetLatestSessionBlockHeight(ctx)
	// get the session context
	sessionCtx, er := ctx.PrevCtx(sessionBlkHeight)
	if er != nil {
		return nil, sdk.ErrInternal(er.Error())
	}
	// get the application that staked on behalf of the client
	app, found := k.GetProviderFromPublicKey(sessionCtx, challenge.MinorityResponse.Proof.Token.ProviderPublicKey)
	if !found {
		return nil, vc.NewProviderNotFoundError(vc.ModuleName)
	}
	// generate header
	header := vc.SessionHeader{
		ProviderPubKey:     challenge.MinorityResponse.Proof.Token.ProviderPublicKey,
		Chain:              challenge.MinorityResponse.Proof.Blockchain,
		SessionBlockHeight: sessionCtx.BlockHeight(),
	}
	// check cache
	session, found := vc.GetSession(header, node.SessionStore)
	// if not found generate the session
	if !found {
		var err sdk.Error
		blockHashBz, er := sessionCtx.BlockHash(k.Cdc, sessionCtx.BlockHeight())
		if er != nil {
			return nil, sdk.ErrInternal(er.Error())
		}
		session, err = vc.NewSession(sessionCtx, ctx, k.posKeeper, header, hex.EncodeToString(blockHashBz))
		if err != nil {
			return nil, err
		}
		// add to cache
		vc.SetSession(session, node.SessionStore)
	}
	// validate the challenge
	err := challenge.ValidateLocal(header, app.GetMaxRelays(), app.GetChains(), int(app.GetNumServicers()), vc.SessionServicers(session.SessionServicers), nodeAddress, node.EvidenceStore)
	if err != nil {
		return nil, err
	}
	// store the challenge in memory
	challenge.Store(app.GetMaxRelays(), node.EvidenceStore)
	// update metric

	if vc.GlobalViperConfig.LeanViper {
		go vc.GlobalServiceMetric().AddChallengeFor(header.Chain, &nodeAddress)
	} else {
		vc.GlobalServiceMetric().AddChallengeFor(header.Chain, &nodeAddress)
	}

	return &vc.ChallengeResponse{Response: fmt.Sprintf("successfully stored challenge proof for %s", challenge.MinorityResponse.Proof.ServicerPubKey)}, nil
}
