package keeper

import (
	"encoding/hex"
	"fmt"

	"github.com/vipernet-xyz/viper-network/codec"

	"github.com/vipernet-xyz/viper-network/crypto"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/authentication"
	"github.com/vipernet-xyz/viper-network/x/authentication/util"
	vc "github.com/vipernet-xyz/viper-network/x/vipernet/types"

	"github.com/tendermint/tendermint/rpc/client"
)

// "SendClaimTx" - Automatically sends a claim of work/challenge based on relays or challenges stored.
func (k Keeper) SendClaimTx(ctx sdk.Ctx, keeper Keeper, n client.Client, claimTx func(pk crypto.PrivateKey, cliCtx util.CLIContext, txBuilder authentication.TxBuilder, header vc.SessionHeader, totalProofs int64, root vc.HashRange, evidenceType vc.EvidenceType) (*sdk.TxResponse, error)) {
	// get the private val key (main) account from the keybase
	kp, err := k.GetPKFromFile(ctx)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("an error occured retrieving the private key from file for the claim transaction:\n%s", err.Error()))
		return
	}
	// retrieve the iterator to go through each piece of evidence in storage
	iter := vc.EvidenceIterator()
	defer iter.Close()
	// loop through each evidence
	for ; iter.Valid(); iter.Next() {
		evidence := iter.Value()
		// if the number of proofs in the evidence object is zero
		if evidence.NumOfProofs == 0 {
			ctx.Logger().Error("evidence of length zero was found in evidence storage")
			continue
		}
		// get the type of the first piece of evidence to know if we are dealing with challenge or relays
		evidenceType := evidence.EvidenceType
		// get the session context
		sessionCtx, er := ctx.PrevCtx(evidence.SessionHeader.SessionBlockHeight)
		if er != nil {
			ctx.Logger().Info("could not get sessionCtx in auto send claim tx, could be due to relay timing before commit is in store: " + er.Error())
			continue
		}
		// if the evidence length is less than minimum, it would not satisfy our merkle tree needs
		if evidence.NumOfProofs < keeper.MinimumNumberOfProofs(sessionCtx) {
			if err := vc.DeleteEvidence(evidence.SessionHeader, evidenceType); err != nil {
				ctx.Logger().Debug(err.Error())
			}
			continue
		}
		if ctx.BlockHeight() <= evidence.SessionBlockHeight+k.BlocksPerSession(sessionCtx)-1 { // ensure session is over
			ctx.Logger().Info("the session is ongoing, so will not send the claim-tx yet")
			continue
		}
		// if the blockchain in the evidence is not supported then delete it because providers don't get paid/challenged for unsupported blockchains
		if !k.IsViperSupportedBlockchain(sessionCtx.WithBlockHeight(evidence.SessionHeader.SessionBlockHeight), evidence.SessionHeader.Chain) {
			ctx.Logger().Info(fmt.Sprintf("claim for %s blockchain isn't viper supported, so will not send. Deleting evidence\n", evidence.SessionHeader.Chain))
			if err := vc.DeleteEvidence(evidence.SessionHeader, evidenceType); err != nil {
				ctx.Logger().Debug(err.Error())
			}
			continue
		}
		// check the current state to see if the unverified evidence has already been sent and processed (if so, then skip this evidence)
		if _, found := k.GetClaim(ctx, sdk.Address(kp.PublicKey().Address()), evidence.SessionHeader, evidenceType); found {
			continue
		}
		// if the claim is mature, delete it because we cannot submit a mature claim
		if k.ClaimIsMature(ctx, evidence.SessionBlockHeight) {
			if err := vc.DeleteEvidence(evidence.SessionHeader, evidenceType); err != nil {
				ctx.Logger().Debug(err.Error())
			}
			continue
		}
		platform, found := k.GetPlatformFromPublicKey(sessionCtx, evidence.PlatformPubKey)
		if !found {
			ctx.Logger().Error(fmt.Sprintf("an error occurred creating the claim transaction with platform %s not found with evidence %v", evidence.PlatformPubKey, evidence))
		}
		// generate the merkle root for this evidence
		root := evidence.GenerateMerkleRoot(evidence.SessionHeader.SessionBlockHeight, vc.MaxPossibleRelays(platform, k.SessionNodeCount(sessionCtx)).Int64())
		// generate the auto txbuilder and clictx
		txBuilder, cliCtx, err := newTxBuilderAndCliCtx(ctx, &vc.MsgClaim{}, n, kp, k)
		if err != nil {
			ctx.Logger().Error(fmt.Sprintf("an error occured creating the tx builder for the claim tx:\n%s", err.Error()))
			return
		}
		// send in the evidence header, the total relays completed, and the merkle root (ensures data integrity)
		if _, err := claimTx(kp, cliCtx, txBuilder, evidence.SessionHeader, evidence.NumOfProofs, root, evidenceType); err != nil {
			ctx.Logger().Error(fmt.Sprintf("an error occured executing the claim transaciton: \n%s", err.Error()))
		}
	}
}

// "ValidateClaim" - Validates a claim message and returns an sdk error if invalid
func (k Keeper) ValidateClaim(ctx sdk.Ctx, claim vc.MsgClaim) (err sdk.Error) {
	// check to see if evidence type is included in the message
	if claim.EvidenceType == 0 {
		return vc.NewNoEvidenceTypeErr(vc.ModuleName)
	}
	// get the session context (state info at the beginning of the session)
	sessionContext, er := ctx.PrevCtx(claim.SessionHeader.SessionBlockHeight)
	if er != nil {
		return sdk.ErrInternal(er.Error())
	}
	// ensure that session ended
	sessionEndHeight := claim.SessionHeader.SessionBlockHeight + k.BlocksPerSession(sessionContext) - 1
	if ctx.BlockHeight() <= sessionEndHeight {
		return vc.NewInvalidBlockHeightError(vc.ModuleName)
	}
	if claim.TotalProofs < k.MinimumNumberOfProofs(sessionContext) {
		return vc.NewInvalidProofsError(vc.ModuleName)
	}
	// if is not a viper supported blockchain then return not supported error
	if !k.IsViperSupportedBlockchain(sessionContext, claim.SessionHeader.Chain) {
		return vc.NewChainNotSupportedErr(vc.ModuleName)
	}
	// get the node from the keeper (at the state of the start of the session)
	_, found := k.GetNode(sessionContext, claim.FromAddress)
	// if not found return not found error
	if !found {
		return vc.NewNodeNotFoundErr(vc.ModuleName)
	}
	// get the platformlication (at the state of the start of the session)
	platform, found := k.GetPlatformFromPublicKey(sessionContext, claim.SessionHeader.PlatformPubKey)
	// if not found return not found error
	if !found {
		return vc.NewPlatformNotFoundError(vc.ModuleName)
	}
	if vc.ModuleCdc.IsAfterNamedFeatureActivationHeight(ctx.BlockHeight(), codec.MaxRelayProtKey) {
		if vc.MaxPossibleRelays(platform, k.SessionNodeCount(sessionContext)).LT(sdk.NewInt(claim.TotalProofs)) {
			return vc.NewOverServiceError(vc.ModuleName)
		}
	}
	// get the session node count for the time of the session
	sessionNodeCount := int(k.SessionNodeCount(sessionContext))
	// check cache
	session, found := vc.GetSession(claim.SessionHeader)
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
		session, err = vc.NewSession(sessionContext, sessionEndCtx, k.posKeeper, claim.SessionHeader, hex.EncodeToString(hash), sessionNodeCount)
		if err != nil {
			ctx.Logger().Error(fmt.Errorf("could not generate session with public key: %s, for chain: %s", platform.GetPublicKey().RawString(), claim.SessionHeader.Chain).Error())
			return err
		}
	}
	// validate the session
	err = session.Validate(claim.FromAddress, platform, sessionNodeCount)
	if err != nil {
		return err
	}
	// check if the proof is ready to be claimed, if it's already ready to be claimed, then it's too late to submit cause the secret is revealed
	if k.ClaimIsMature(ctx, claim.SessionHeader.SessionBlockHeight) {
		return vc.NewExpiredProofsSubmissionError(vc.ModuleName)
	}
	return nil
}

// "SetClaim" - Sets the claim message in the state storage
func (k Keeper) SetClaim(ctx sdk.Ctx, msg vc.MsgClaim) error {
	// retrieve the store
	store := ctx.KVStore(k.storeKey)
	// generate the store key
	key, err := vc.KeyForClaim(ctx, msg.FromAddress, msg.SessionHeader, msg.EvidenceType)
	if err != nil {
		return err
	}
	// generate the expiration height upon setting
	if msg.ExpirationHeight == 0 {
		sessionCtx, err := ctx.PrevCtx(msg.SessionHeader.SessionBlockHeight)
		if err != nil {
			return err
		}
		msg.ExpirationHeight = ctx.BlockHeight() + k.ClaimExpiration(sessionCtx)*k.BlocksPerSession(sessionCtx)
	}
	// marshal the message into amino
	bz, err := k.Cdc.MarshalBinaryBare(&msg, ctx.BlockHeight())
	if err != nil {
		panic(err)
	}
	// set in the store
	_ = store.Set(key, bz)
	return nil
}

// "GetClaim" - Retrieves the claim message from the store, requires the evidence type and header to return the proper claim message
func (k Keeper) GetClaim(ctx sdk.Ctx, address sdk.Address, header vc.SessionHeader, evidenceType vc.EvidenceType) (msg vc.MsgClaim, found bool) {
	// retrieve the store
	store := ctx.KVStore(k.storeKey)
	// generate the store key
	key, err := vc.KeyForClaim(ctx, address, header, evidenceType)
	if err != nil {
		ctx.Logger().Error("an error occured getting the claim:\n", msg)
		return vc.MsgClaim{}, false
	}
	// get the claim msg from the store
	res, _ := store.Get(key)
	if res == nil {
		return vc.MsgClaim{}, false
	}
	// unmarshal into message object
	err = k.Cdc.UnmarshalBinaryBare(res, &msg, ctx.BlockHeight())
	if err != nil {
		panic(err)
	}
	// return the object
	return msg, true
}

// "SetClaims" - Sets all the claim messages in the state storage.
// (Needed for genesis initializing)
func (k Keeper) SetClaims(ctx sdk.Ctx, claims []vc.MsgClaim) {
	// loop through all of the claim messages one by one and set them
	for _, msg := range claims {
		err := k.SetClaim(ctx, msg)
		if err != nil {
			ctx.Logger().Error("an error occurred setting the claim:\n", msg)
		}
	}
}

// "GetClaims" - Gets all of the claim messages in the state storage for an address
func (k Keeper) GetClaims(ctx sdk.Ctx, address sdk.Address) (claims []vc.MsgClaim, err error) {
	// retrieve the store
	store := ctx.KVStore(k.storeKey)
	// generate the key for the claims
	key, err := vc.KeyForClaims(address)
	if err != nil {
		return nil, err
	}
	// iterate through all of the kv pairs and unmarshal into claim objects
	iterator, _ := sdk.KVStorePrefixIterator(store, key)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var claim vc.MsgClaim
		err = k.Cdc.UnmarshalBinaryBare(iterator.Value(), &claim, ctx.BlockHeight())
		if err != nil {
			panic(err)
		}
		claims = append(claims, claim)
	}
	return
}

// "GetAllClaims" - Gets all of the claim messages held in the state storage.
func (k Keeper) GetAllClaims(ctx sdk.Ctx) (claims []vc.MsgClaim) {
	// retrieve the store
	store := ctx.KVStore(k.storeKey)
	// iterate through the kv in the state and unmarshal into claim objects
	iterator, _ := sdk.KVStorePrefixIterator(store, vc.ClaimKey)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var claim vc.MsgClaim
		err := k.Cdc.UnmarshalBinaryBare(iterator.Value(), &claim, ctx.BlockHeight())
		if err != nil {
			panic(err)
		}
		claims = append(claims, claim)
	}
	return
}

// "DeleteClaim" - Removes a claim object for a certain key
func (k Keeper) DeleteClaim(ctx sdk.Ctx, address sdk.Address, header vc.SessionHeader, evidenceType vc.EvidenceType) error {
	// retrieve the store
	store := ctx.KVStore(k.storeKey)
	// generate the key for the claim
	key, err := vc.KeyForClaim(ctx, address, header, evidenceType)
	if err != nil {
		return err
	}
	// delete it from the state storage
	_ = store.Delete(key)
	return nil
}

// "GetMatureClaims" - Returns the mature (ready to be proved, past its security waiting period)
func (k Keeper) GetMatureClaims(ctx sdk.Ctx, address sdk.Address) (matureProofs []vc.MsgClaim, err error) {
	// retrieve the store
	store := ctx.KVStore(k.storeKey)
	// generate the key for the claim
	key, err := vc.KeyForClaims(address)
	if err != nil {
		return nil, err
	}
	// iterate through all kv and see if the claim is mature for each
	iterator, _ := sdk.KVStorePrefixIterator(store, key)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var msg vc.MsgClaim
		err = k.Cdc.UnmarshalBinaryBare(iterator.Value(), &msg, ctx.BlockHeight())
		if err != nil {
			panic(err)
		}
		// if the claim is mature, add it to the list
		if k.ClaimIsMature(ctx, msg.SessionHeader.SessionBlockHeight) {
			matureProofs = append(matureProofs, msg)
		}
	}
	return
}

// "ClaimIsMature" - Returns if the claim is past its security waiting period
func (k Keeper) ClaimIsMature(ctx sdk.Ctx, sessionBlockHeight int64) bool {
	waitingPeriodInBlocks := k.ClaimSubmissionWindow(ctx) * k.BlocksPerSession(ctx)
	return ctx.BlockHeight() > waitingPeriodInBlocks+sessionBlockHeight
}

// "DeleteExpiredClaims" - Deletes the expired (claim expiration > # of session passed since claim genesis) claims
func (k Keeper) DeleteExpiredClaims(ctx sdk.Ctx) {
	var msg = vc.MsgClaim{}
	store := ctx.KVStore(k.storeKey)
	iterator, _ := sdk.KVStorePrefixIterator(store, vc.ClaimKey)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		err := k.Cdc.UnmarshalBinaryBare(iterator.Value(), &msg, ctx.BlockHeight())
		if err != nil {
			panic(err)
		}
		// if more sessions has passed than the expiration of the claim's genesis, delete it from the set
		if msg.ExpirationHeight <= ctx.BlockHeight() {
			_ = store.Delete(iterator.Key())
		}
	}
}
