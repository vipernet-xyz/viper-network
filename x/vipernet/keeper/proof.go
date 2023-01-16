package keeper

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"reflect"

	"github.com/vipernet-xyz/viper-network/codec"
	"github.com/vipernet-xyz/viper-network/crypto"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/authentication"
	"github.com/vipernet-xyz/viper-network/x/authentication/util"
	vc "github.com/vipernet-xyz/viper-network/x/vipernet/types"

	"github.com/tendermint/tendermint/rpc/client"
)

// auto sends a proof transaction for the claim
func (k Keeper) SendProofTx(ctx sdk.Ctx, n client.Client, proofTx func(cliCtx util.CLIContext, txBuilder authentication.TxBuilder, merkleProof vc.MerkleProof, leafNode vc.Proof, evidenceType vc.EvidenceType) (*sdk.TxResponse, error)) {
	kp, err := k.GetPKFromFile(ctx)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("an error occured retrieving the pk from the file for the Proof Transaction:\n%v", err))
		return
	}
	// get the self address
	addr := sdk.Address(kp.PublicKey().Address())
	// get all mature (waiting period has passed) claims for your address
	claims, err := k.GetMatureClaims(ctx, addr)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("an error occured getting the mature claims in the Proof Transaction:\n%v", err))
		return
	}
	// for every claim of the mature set
	for _, claim := range claims {
		// check to see if evidence is stored in cache
		evidence, err := vc.GetEvidence(claim.SessionHeader, claim.EvidenceType, sdk.ZeroInt())
		if err != nil || evidence.Proofs == nil || len(evidence.Proofs) == 0 {
			ctx.Logger().Info(fmt.Sprintf("the evidence object for evidence is not found, ignoring pending claim for platform: %s, at sessionHeight: %d", claim.SessionHeader.PlatformPubKey, claim.SessionHeader.SessionBlockHeight))
			continue
		}
		if ctx.BlockHeight()-claim.SessionHeader.SessionBlockHeight > int64(vc.GlobalViperConfig.MaxClaimAgeForProofRetry) {
			err := vc.DeleteEvidence(claim.SessionHeader, claim.EvidenceType)
			ctx.Logger().Error(fmt.Sprintf("deleting evidence older than MaxClaimAgeForProofRetry"))
			if err != nil {
				ctx.Logger().Error(fmt.Sprintf("unable to delete evidence that is older than 32 blocks: %s", err.Error()))
			}
			continue
		}
		if !evidence.IsSealed() {
			err := vc.DeleteEvidence(claim.SessionHeader, claim.EvidenceType)
			ctx.Logger().Error(fmt.Sprintf("evidence is not sealed, could cause a relay leak:"))
			if err != nil {
				ctx.Logger().Error(fmt.Sprintf("could not delete evidence is not sealed, could cause a relay leak: %s", err.Error()))
			}
		}
		if evidence.NumOfProofs != claim.TotalProofs {
			err := vc.DeleteEvidence(claim.SessionHeader, claim.EvidenceType)
			ctx.Logger().Error(fmt.Sprintf("evidence num of proofs does not equal claim total proofs... possible relay leak"))
			if err != nil {
				ctx.Logger().Error(fmt.Sprintf("evidence num of proofs does not equal claim total proofs... possible relay leak: %s", err.Error()))
			}
		}
		// get the session context
		sessionCtx, err := ctx.PrevCtx(claim.SessionHeader.SessionBlockHeight)
		if err != nil {
			ctx.Logger().Info(fmt.Sprintf("could not get Session Context, ignoring pending claim for platform: %s, at sessionHeight: %d", claim.SessionHeader.PlatformPubKey, claim.SessionHeader.SessionBlockHeight))
			continue
		}
		// generate the needed pseudorandom index using the information found in the first transaction
		index, err := k.getPseudorandomIndex(ctx, claim.TotalProofs, claim.SessionHeader, sessionCtx)
		if err != nil {
			ctx.Logger().Error(err.Error())
			continue
		}
		platform, found := k.GetPlatformFromPublicKey(sessionCtx, claim.SessionHeader.PlatformPubKey)
		if !found {
			ctx.Logger().Error(fmt.Sprintf("an error occurred creating the proof transaction with platform %s not found with evidence %v", evidence.PlatformPubKey, evidence))
		}
		// get the merkle proof object for the pseudorandom index
		mProof, leaf := evidence.GenerateMerkleProof(claim.SessionHeader.SessionBlockHeight, int(index), vc.MaxPossibleRelays(platform, k.SessionNodeCount(sessionCtx)).Int64())
		// if prevalidation on, then pre-validate
		if vc.GlobalViperConfig.ProofPrevalidation {
			// validate level count on claim by total relays
			levelCount := len(mProof.HashRanges)
			if levelCount != int(math.Ceil(math.Log2(float64(claim.TotalProofs)))) {
				ctx.Logger().Error(fmt.Sprintf("produced invalid proof for pending claim for platform: %s, at sessionHeight: %d, level count", claim.SessionHeader.PlatformPubKey, claim.SessionHeader.SessionBlockHeight))
				continue
			}
			if isValid, _ := mProof.Validate(claim.SessionHeader.SessionBlockHeight, claim.MerkleRoot, leaf, levelCount); !isValid {
				ctx.Logger().Error(fmt.Sprintf("produced invalid proof for pending claim for platform: %s, at sessionHeight: %d", claim.SessionHeader.PlatformPubKey, claim.SessionHeader.SessionBlockHeight))
				continue
			}
		}
		// generate the auto txbuilder and clictx
		txBuilder, cliCtx, err := newTxBuilderAndCliCtx(ctx, &vc.MsgProof{}, n, kp, k)
		if err != nil {
			ctx.Logger().Error(fmt.Sprintf("an error occured in the transaction process of the Proof Transaction:\n%v", err))
			return
		}
		// send the proof TX
		_, err = proofTx(cliCtx, txBuilder, mProof, leaf, evidence.EvidenceType)
		if err != nil {
			ctx.Logger().Error(err.Error())
		}
	}
}

func (k Keeper) ValidateProof(ctx sdk.Ctx, proof vc.MsgProof) (servicerAddr sdk.Address, claim vc.MsgClaim, sdkError sdk.Error) {
	// get the public key from the claim
	servicerAddr = proof.GetSigners()[0]
	// get the claim for the address
	claim, found := k.GetClaim(ctx, servicerAddr, proof.GetLeaf().SessionHeader(), proof.EvidenceType)
	// if the claim is not found for this claim
	if !found {
		return servicerAddr, claim, vc.NewClaimNotFoundError(vc.ModuleName)
	}
	// validate level count on claim by total relays
	levelCount := len(proof.MerkleProof.HashRanges)
	if levelCount != int(math.Ceil(math.Log2(float64(claim.TotalProofs)))) {
		return servicerAddr, claim, vc.NewInvalidProofsError(vc.ModuleName)
	}
	var hasMatch bool
	for _, m := range proof.MerkleProof.HashRanges {
		if claim.MerkleRoot.Range.Upper == m.Range.Upper {
			hasMatch = true
			break
		}
	}
	if !hasMatch && proof.MerkleProof.Target.Range.Upper != claim.MerkleRoot.Range.Upper {
		return servicerAddr, claim, vc.NewInvalidMerkleVerifyError(vc.ModuleName)
	}
	// get the session context
	sessionCtx, err := ctx.PrevCtx(claim.SessionHeader.SessionBlockHeight)
	if err != nil {
		return servicerAddr, claim, sdk.ErrInternal(err.Error())
	}
	// validate the proof
	ctx.Logger().Info(fmt.Sprintf("Generate psuedorandom proof with %d proofs, at session height of %d, for platform: %s", claim.TotalProofs, claim.SessionHeader.SessionBlockHeight, claim.SessionHeader.PlatformPubKey))
	reqProof, err := k.getPseudorandomIndex(ctx, claim.TotalProofs, claim.SessionHeader, sessionCtx)
	if err != nil {
		return servicerAddr, claim, sdk.ErrInternal(err.Error())
	}
	// if the required proof message index does not match the leaf node index
	if reqProof != int64(proof.MerkleProof.TargetIndex) {
		return servicerAddr, claim, vc.NewInvalidProofsError(vc.ModuleName)
	}
	// validate the merkle proofs
	isValid, isReplayAttack := proof.MerkleProof.Validate(claim.SessionHeader.SessionBlockHeight, claim.MerkleRoot, proof.GetLeaf(), levelCount)
	// if is not valid for other reasons
	if !isValid {
		if isReplayAttack && k.Cdc.IsAfterNamedFeatureActivationHeight(ctx.BlockHeight(), codec.ReplayBurnKey) {
			return servicerAddr, claim, vc.NewReplayAttackError(vc.ModuleName)
		}
		return servicerAddr, claim, vc.NewInvalidMerkleVerifyError(vc.ModuleName)
	}
	// get the platformlication
	platformlication, found := k.GetPlatformFromPublicKey(sessionCtx, claim.SessionHeader.PlatformPubKey)
	if !found {
		return servicerAddr, claim, vc.NewPlatformNotFoundError(vc.ModuleName)
	}
	// validate the proof depending on the type of proof it is
	er := proof.GetLeaf().Validate(platformlication.GetChains(), int(k.SessionNodeCount(sessionCtx)), claim.SessionHeader.SessionBlockHeight)
	if er != nil {
		return nil, claim, er
	}
	// return the needed info to the handler
	return servicerAddr, claim, nil
}

func (k Keeper) ExecuteProof(ctx sdk.Ctx, proof vc.MsgProof, claim vc.MsgClaim) (tokens sdk.BigInt, err sdk.Error) {
	//platformlication address
	platformAddress := sdk.Address(claim.SessionHeader.PlatformPubKey)
	// convert to value for switch consistency
	l := proof.GetLeaf()
	if reflect.ValueOf(l).Kind() == reflect.Ptr {
		l = reflect.Indirect(reflect.ValueOf(l)).Interface().(vc.Proof)
	}
	switch l.(type) {
	case vc.RelayProof:
		ctx.Logger().Info(fmt.Sprintf("reward coins to %s, for %d relays", claim.FromAddress.String(), claim.TotalProofs))
		tokens = k.AwardCoinsForRelays(ctx, claim.TotalProofs, claim.FromAddress, platformAddress)
		err := k.DeleteClaim(ctx, claim.FromAddress, claim.SessionHeader, vc.RelayEvidence)
		if err != nil {
			return tokens, sdk.ErrInternal(err.Error())
		}
	case vc.ChallengeProofInvalidData:
		ctx.Logger().Info(fmt.Sprintf("burning coins from %s, for %d valid challenges", claim.FromAddress.String(), claim.TotalProofs))
		proof, ok := proof.GetLeaf().(vc.ChallengeProofInvalidData)
		if !ok {
			return sdk.ZeroInt(), vc.NewInvalidProofsError(vc.ModuleName)
		}
		pk := proof.MinorityResponse.Proof.ServicerPubKey
		pubKey, err := crypto.NewPublicKey(pk)
		if err != nil {
			return sdk.ZeroInt(), sdk.ErrInvalidPubKey(err.Error())
		}
		k.BurnCoinsForChallenges(ctx, claim.TotalProofs, sdk.Address(pubKey.Address()))
		err = k.DeleteClaim(ctx, claim.FromAddress, claim.SessionHeader, vc.ChallengeEvidence)
		if err != nil {
			return sdk.ZeroInt(), sdk.ErrInternal(err.Error())
		}
		// small reward for the challenge proof invalid data
		tokens = k.AwardCoinsForRelays(ctx, claim.TotalProofs/100, claim.FromAddress, platformAddress)
	}
	return tokens, nil
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
	pk, err := k.GetPKFromFile(ctx)
	if err != nil {
		return txBuilder, cliCtx, err
	}
	cliCtx.PrivateKey = pk
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
