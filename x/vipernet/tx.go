package vipernet

import (
	"github.com/vipernet-xyz/viper-network/codec"
	"github.com/vipernet-xyz/viper-network/crypto"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/auth"
	"github.com/vipernet-xyz/viper-network/x/auth/util"
	"github.com/vipernet-xyz/viper-network/x/vipernet/types"
)

// "ClaimTx" - A transaction that sends the total number of proofs (claim), the merkle root (for data integrity), and the header (for identification)
func ClaimTx(kp crypto.PrivateKey, cliCtx util.CLIContext, txBuilder auth.TxBuilder, header types.SessionHeader, totalProofs int64, root types.HashRange, evidenceType types.EvidenceType) (*sdk.TxResponse, error) {
	msg := types.MsgClaim{
		SessionHeader:    header,
		TotalProofs:      totalProofs,
		MerkleRoot:       root,
		FromAddress:      sdk.Address(kp.PublicKey().Address()),
		EvidenceType:     evidenceType,
		ExpirationHeight: 0, // leave as zero
	}
	err := msg.ValidateBasic()
	if err != nil {
		return nil, err
	}
	var legacyCodec bool
	if cliCtx.Height < codec.GetCodecUpgradeHeight() {
		legacyCodec = true
	}
	return util.CompleteAndBroadcastTxCLI(txBuilder, cliCtx, &msg, legacyCodec)
}

// "ProofTx" - A transaction to prove the claim that was previously sent (Merkle Proofs and leaf/cousin)
func ProofTx(cliCtx util.CLIContext, txBuilder auth.TxBuilder, merkleProof types.MerkleProof, leafNode types.Proof, evidenceType types.EvidenceType) (*sdk.TxResponse, error) {
	msg := types.MsgProof{
		MerkleProof:  merkleProof,
		Leaf:         leafNode,
		EvidenceType: evidenceType,
	}
	err := msg.ValidateBasic()
	if err != nil {
		return nil, err
	}
	var legacyCodec bool
	if cliCtx.Height < codec.GetCodecUpgradeHeight() {
		legacyCodec = true
	}
	return util.CompleteAndBroadcastTxCLI(txBuilder, cliCtx, &msg, legacyCodec)
}
