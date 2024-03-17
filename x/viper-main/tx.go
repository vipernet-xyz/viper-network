package vipernet

import (
	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/authentication"
	"github.com/vipernet-xyz/viper-network/x/authentication/util"
	"github.com/vipernet-xyz/viper-network/x/viper-main/types"
)

// "ClaimTx" - A transaction that sends the total number of proofs (claim), the merkle root (for data integrity), and the header (for identification)
func ClaimTx(pk crypto.PrivateKey, cliCtx util.CLIContext, txBuilder authentication.TxBuilder, header types.SessionHeader, totalProofs int64, root types.HashRange, evidenceType types.EvidenceType) (*sdk.TxResponse, error) {
	msg := types.MsgClaim{
		SessionHeader:    header,
		TotalProofs:      totalProofs,
		MerkleRoot:       root,
		FromAddress:      sdk.Address(pk.PublicKey().Address()),
		EvidenceType:     evidenceType,
		ExpirationHeight: 0, // leave as zero
	}
	err := msg.ValidateBasic()
	if err != nil {
		return nil, err
	}

	return util.CompleteAndBroadcastTxCLI(txBuilder, cliCtx, &msg, false)
}

// "ProofTx" - A transaction to prove the claim that was previously sent (Merkle Proofs and leaf/cousin)
func ProofTx(cliCtx util.CLIContext, txBuilder authentication.TxBuilder, claimMerkleProof types.MerkleProof, claimLeafNode types.Proof, claimEvidenceType types.EvidenceType, reportMerkleProof types.MerkleProof, reportLeafNode types.Test, reportEvidenceType types.EvidenceType) (*sdk.TxResponse, error) {
	msg := types.MsgProof{
		ClaimMerkleProof:   claimMerkleProof,
		ClaimLeaf:          claimLeafNode,
		ClaimEvidenceType:  claimEvidenceType,
		ReportMerkleProof:  reportMerkleProof,
		ReportLeaf:         reportLeafNode,
		ReportEvidenceType: reportEvidenceType,
	}
	err := msg.ValidateBasic()
	if err != nil {
		return nil, err
	}
	return util.CompleteAndBroadcastTxCLI(txBuilder, cliCtx, &msg, false)
}

func ReportCardTx(pk crypto.PrivateKey, cliCtx util.CLIContext, txBuilder authentication.TxBuilder, header types.SessionHeader, servicerAddr sdk.Address, reportCard types.ViperQoSReport, merkleProof types.MerkleProof, leafNode types.TestI, numOfTestResults int64, evidenceType types.EvidenceType) (*sdk.TxResponse, error) {
	msg := types.MsgSubmitQoSReport{
		SessionHeader:    header,
		ServicerAddress:  servicerAddr,
		FishermanAddress: sdk.Address(pk.PublicKey().Address()),
		Report:           reportCard,
		EvidenceType:     evidenceType,
		MerkleProof:      merkleProof,
		Leaf:             leafNode,
		NumOfTestResults: numOfTestResults,
	}
	err := msg.ValidateBasic()
	if err != nil {
		return nil, err
	}

	return util.CompleteAndBroadcastTxCLI(txBuilder, cliCtx, &msg, false)
}
