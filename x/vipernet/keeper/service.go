package keeper

import (
	"encoding/hex"
	"fmt"
	"time"

	sdk "github.com/vipernet-xyz/viper-network/types"
	vc "github.com/vipernet-xyz/viper-network/x/vipernet/types"
)

// "HandleRelay" - Handles an api (read/write) request to a non-native (external) blockchain
func (k Keeper) HandleRelay(ctx sdk.Ctx, relay vc.Relay) (*vc.RelayResponse, sdk.Error) {
	relayTimeStart := time.Now()
	// get the latest session block height because this relay will correspond with the latest session
	sessionBlockHeight := k.GetLatestSessionBlockHeight(ctx)
	// get self node (your validator) from the current state
	pk, err := k.GetSelfPrivKey(ctx)
	if err != nil {
		return nil, err
	}
	selfAddr := sdk.Address(pk.PublicKey().Address())
	// retrieve the nonNative blockchains your node is hosting
	hostedBlockchains := k.GetHostedBlockchains()
	// ensure the validity of the relay
	maxPossibleRelays, err := relay.Validate(ctx, k.posKeeper, k.platformKeeper, k, selfAddr, hostedBlockchains, sessionBlockHeight)
	if err != nil {
		if vc.GlobalViperConfig.RelayErrors {
			ctx.Logger().Error(
				fmt.Sprintf("could not validate relay for platform: %s for chainID: %v with error: %s",
					relay.Proof.ServicerPubKey,
					relay.Proof.Blockchain,
					err.Error(),
				),
			)
			ctx.Logger().Debug(
				fmt.Sprintf(
					"could not validate relay for platform: %s, for chainID %v on node %s, at session height: %v, with error: %s",
					relay.Proof.ServicerPubKey,
					relay.Proof.Blockchain,
					selfAddr.String(),
					sessionBlockHeight,
					err.Error(),
				),
			)
		}
		return nil, err
	}
	// store the proof before execution, because the proof corresponds to the previous relay
	relay.Proof.Store(maxPossibleRelays)
	// attempt to execute
	respPayload, err := relay.Execute(hostedBlockchains)
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
	sig, er := pk.Sign(resp.Hash())
	if er != nil {
		ctx.Logger().Error(
			fmt.Sprintf("could not sign response for address: %s with hash: %v, with error: %s",
				selfAddr.String(), resp.HashString(), er.Error()),
		)
		return nil, vc.NewKeybaseError(vc.ModuleName, er)
	}
	// attach the signature in hex to the response
	resp.Signature = hex.EncodeToString(sig)
	// track the relay time
	relayTime := time.Since(relayTimeStart)
	// add to metrics
	vc.GlobalServiceMetric().AddRelayTimingFor(relay.Proof.Blockchain, float64(relayTime.Milliseconds()))
	vc.GlobalServiceMetric().AddRelayFor(relay.Proof.Blockchain)
	return resp, nil
}

// "HandleChallenge" - Handles a client relay response challenge request
func (k Keeper) HandleChallenge(ctx sdk.Ctx, challenge vc.ChallengeProofInvalidData) (*vc.ChallengeResponse, sdk.Error) {
	// get self node (your validator) from the current state
	selfNode := k.GetSelfAddress(ctx)
	sessionBlkHeight := k.GetLatestSessionBlockHeight(ctx)
	// get the session context
	sessionCtx, er := ctx.PrevCtx(sessionBlkHeight)
	if er != nil {
		return nil, sdk.ErrInternal(er.Error())
	}
	// get the platformlication that staked on behalf of the client
	platform, found := k.GetPlatformFromPublicKey(sessionCtx, challenge.MinorityResponse.Proof.Token.PlatformPublicKey)
	if !found {
		return nil, vc.NewPlatformNotFoundError(vc.ModuleName)
	}
	// generate header
	header := vc.SessionHeader{
		PlatformPubKey:     challenge.MinorityResponse.Proof.Token.PlatformPublicKey,
		Chain:              challenge.MinorityResponse.Proof.Blockchain,
		SessionBlockHeight: sessionCtx.BlockHeight(),
	}
	// check cache
	session, found := vc.GetSession(header)
	// if not found generate the session
	if !found {
		var err sdk.Error
		blockHashBz, er := sessionCtx.BlockHash(k.Cdc, sessionCtx.BlockHeight())
		if er != nil {
			return nil, sdk.ErrInternal(er.Error())
		}
		session, err = vc.NewSession(sessionCtx, ctx, k.posKeeper, header, hex.EncodeToString(blockHashBz), int(k.SessionNodeCount(sessionCtx)))
		if err != nil {
			return nil, err
		}
		// add to cache
		vc.SetSession(session)
	}
	// validate the challenge
	err := challenge.ValidateLocal(header, platform.GetMaxRelays(), platform.GetChains(), int(k.SessionNodeCount(sessionCtx)), session.SessionNodes, selfNode)
	if err != nil {
		return nil, err
	}
	// store the challenge in memory
	challenge.Store(platform.GetMaxRelays())
	// update metric
	vc.GlobalServiceMetric().AddChallengeFor(header.Chain)
	return &vc.ChallengeResponse{Response: fmt.Sprintf("successfully stored challenge proof for %s", challenge.MinorityResponse.Proof.ServicerPubKey)}, nil
}
