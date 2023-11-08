package types

import (
	"encoding/hex"
	"fmt"

	"github.com/vipernet-xyz/viper-network/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
)

// RouterKey is the module name router key
const (
	RouterKey               = ModuleName // router name is module name
	MsgClaimName            = "claim"    // name for the claim message
	MsgProofName            = "proof"    // name for the proof message
	MsgSubmitReportCardName = "submitReportCard"
)

// "GetFee" - Returns the fee (sdk.BigInt) of the message type
func (msg MsgClaim) GetFee() sdk.BigInt {
	return sdk.NewInt(ViperFeeMap[msg.Type()])
}

// "Route" - Returns module router key
func (msg MsgClaim) Route() string { return RouterKey }

// "Type" - Returns message name
func (msg MsgClaim) Type() string { return MsgClaimName }

// "ValidateBasic" - Storeless validity check for claim message
func (msg MsgClaim) ValidateBasic() sdk.Error {
	// validate a non empty chain
	if msg.SessionHeader.Chain == "" {
		return NewEmptyChainError(ModuleName)
	}
	// basic validation on the session block height
	if msg.SessionHeader.SessionBlockHeight < 1 {
		return NewEmptyBlockIDError(ModuleName)
	}
	// validate greater than 5 relays (need 5 for the tree structure)
	if msg.TotalProofs < 5 {
		return NewEmptyProofsError(ModuleName)
	}
	// validate the public key format
	if err := PubKeyVerification(msg.SessionHeader.ProviderPubKey); err != nil {
		return NewPubKeyError(ModuleName, err)
	}
	// validate the address format
	if err := AddressVerification(msg.FromAddress.String()); err != nil {
		return NewInvalidHashError(ModuleName, err, msg.FromAddress.String())
	}
	// validate the root format
	if err := HashVerification(hex.EncodeToString(msg.MerkleRoot.Hash)); err != nil {
		return err
	}
	// ensure non zero root upper range
	if !msg.MerkleRoot.isValidRange() {
		return NewInvalidMerkleRangeError(ModuleName)
	}
	// ensure zero root lower range
	if msg.MerkleRoot.Range.Lower != 0 {
		return NewInvalidRootError(ModuleName)
	}
	// ensure non zero GOBEvidence
	if msg.EvidenceType == 0 {
		return NewNoEvidenceTypeErr(ModuleName)
	}
	if msg.EvidenceType != RelayEvidence && msg.EvidenceType != ChallengeEvidence {
		return NewInvalidEvidenceErr(ModuleName)
	}
	if msg.ExpirationHeight != 0 {
		return NewInvalidExpirationHeightErr(ModuleName)
	}
	return nil
}

// "GetSignBytes" - Encodes the message for signing
func (msg MsgClaim) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// "GetSigners" - Defines whose signature is required
func (msg MsgClaim) GetSigners() []sdk.Address {
	return []sdk.Address{msg.FromAddress}
}

// "GetSigners" - Defines whose signature is required
func (msg MsgClaim) GetRecipient() sdk.Address {
	return nil
}

// "IsEmpty" - Returns true if the EvidenceType == 0, this should only happen on initialization and MsgClaim{} calls
func (msg MsgClaim) IsEmpty() bool {
	return msg.EvidenceType == 0
}

// ---------------------------------------------------------------------------------------------------------------------
// "MsgProof" - Proves the previous claim by providing the merkle Proof and the leaf servicer
type MsgProof struct {
	MerkleProof  MerkleProof  `json:"merkle_proofs"` // the merkleProof needed to verify the proofs
	Leaf         Proof        `json:"leaf"`          // the needed to verify the Proof
	EvidenceType EvidenceType `json:"evidence_type"` // the type of GOBEvidence
}

var _ codec.ProtoMarshaler = &MsgProof{}

func (msg *MsgProof) Marshal() ([]byte, error) {
	m := msg.ToProto()
	return m.Marshal()
}

func (msg *MsgProof) MarshalTo(data []byte) (n int, err error) {
	m := msg.ToProto()
	return m.MarshalTo(data)
}

func (msg *MsgProof) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	m := msg.ToProto()
	return m.MarshalToSizedBuffer(dAtA)
}

func (msg *MsgProof) Size() int {
	m := msg.ToProto()
	return m.Size()
}

func (msg *MsgProof) Unmarshal(data []byte) error {
	var m MsgProtoProof
	err := m.Unmarshal(data)
	if err != nil {
		return err
	}
	*msg = MsgProof{
		MerkleProof:  m.MerkleProof,
		Leaf:         m.Leaf.FromProto(),
		EvidenceType: m.EvidenceType,
	}
	return nil
}

func (msg *MsgProof) Reset() {
	*msg = MsgProof{}
}

func (msg *MsgProof) ProtoMessage() {
	m := msg.ToProto()
	m.ProtoMessage()
}

func (msg MsgProof) String() string {
	return fmt.Sprintf("MerkleProof: %s\nLeaf: %v\nEvidenceType: %d\n", msg.MerkleProof.String(), msg.Leaf, msg.EvidenceType)
}

func (msg MsgProof) ToProto() MsgProtoProof {
	return MsgProtoProof{
		MerkleProof:  msg.MerkleProof,
		Leaf:         msg.Leaf.ToProto(),
		EvidenceType: msg.EvidenceType,
	}
}

// "GetFee" - Returns the fee (sdk.BigInt) of the message type
func (msg MsgProof) GetFee() sdk.BigInt {
	return sdk.NewInt(ViperFeeMap[msg.Type()])
}

// "Route" - Returns module router key
func (msg MsgProof) Route() string { return RouterKey }

// "Type" - Returns message name
func (msg MsgProof) Type() string { return MsgProofName }

// "ValidateBasic" - Storeless validity check for proof message
func (msg MsgProof) ValidateBasic() sdk.Error {
	// verify valid number of levels for merkle proofs
	if len(msg.MerkleProof.HashRanges) < 3 {
		return NewInvalidLeafCousinProofsComboError(ModuleName)
	}
	// validate the target range
	if !msg.MerkleProof.Target.isValidRange() {
		return NewInvalidMerkleRangeError(ModuleName)
	}
	// validate the leaf
	if err := msg.Leaf.ValidateBasic(); err != nil {
		return err
	}
	if _, err := msg.EvidenceType.Byte(); err != nil {
		return NewInvalidEvidenceErr(ModuleName)
	}
	return nil
}

// "GetSignBytes" - Encodes the message for signing
func (msg MsgProof) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners defines whose signature is required
func (msg MsgProof) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Leaf.GetSigner()}
}

// "GetSigners" - Defines whose signature is required
func (msg MsgProof) GetRecipient() sdk.Address {
	return nil
}

func (msg MsgProof) GetLeaf() Proof {
	return msg.Leaf
}

// ---------------------------------------------------------------------------------------------------------------------
// "MsgSubmitReportCard"

// "GetFee" - Returns the fee (sdk.BigInt) of the message type
func (msg MsgSubmitReportCard) GetFee() sdk.BigInt {
	return sdk.NewInt(ViperFeeMap[msg.Type()])
}

// "Route" - Returns module router key
func (msg MsgSubmitReportCard) Route() string { return RouterKey }

// "Type" - Returns message name
func (msg MsgSubmitReportCard) Type() string { return MsgClaimName }

func (msg MsgSubmitReportCard) ValidateBasic() sdk.Error {
	// Validate non-empty servicer address
	if msg.ServicerAddress.Empty() {
		return sdk.ErrInvalidAddress("Servicer address cannot be empty")
	}

	// Validate the report
	report := msg.Report

	// Ensure the block height is positive
	if report.BlockHeight < 1 {
		return sdk.ErrInvalidSequence("Block height must be positive")
	}

	// Ensure the LatencyScore, AvailabilityScore, and ReliabilityScore are within acceptable ranges
	// You can adjust these checks based on your specific requirements
	if report.LatencyScore.IsNegative() || report.AvailabilityScore.IsNegative() || report.ReliabilityScore.IsNegative() {
		return sdk.ErrInternal("Scores cannot be negative")
	}

	// Validate SampleRoot
	// Assuming the HashRange struct has a method to validate itself called IsValid()
	if !report.SampleRoot.isValidRange() {
		return sdk.ErrInternal("Invalid Sample Root")
	}

	// Ensure nonce is positive
	if report.Nonce < 1 {
		return sdk.ErrInvalidSequence("Nonce must be positive")
	}

	// Validate the signature is not empty (and potentially other signature checks if needed)
	if report.Signature == "" {
		return sdk.ErrUnauthorized("Missing signature")
	}

	if err := AddressVerification(msg.ServicerAddress.String()); err != nil {
		return NewInvalidHashError(ModuleName, err, msg.ServicerAddress.String())
	}
	return nil
}

// "GetSignBytes" - Encodes the message for signing
func (msg MsgSubmitReportCard) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// "GetSigners" - Defines whose signature is required
func (msg MsgSubmitReportCard) GetSigners() []sdk.Address {
	return []sdk.Address{msg.FishermanAddress}
}

// "GetSigners" - Defines whose signature is required
func (msg MsgSubmitReportCard) GetRecipient() sdk.Address {
	return nil
}
