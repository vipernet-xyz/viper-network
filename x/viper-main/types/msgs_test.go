package types

import (
	"encoding/hex"
	"reflect"
	"testing"
	"time"

	"github.com/vipernet-xyz/viper-network/types"

	"github.com/stretchr/testify/assert"
)

func TestMsgClaim_Route(t *testing.T) {
	assert.Equal(t, MsgClaim{}.Route(), RouterKey)
}

func TestMsgClaim_Type(t *testing.T) {
	assert.Equal(t, MsgClaim{}.Type(), MsgClaimName)
}

func TestMsgClaim_GetSigners(t *testing.T) {
	addr := getRandomValidatorAddress()
	signers := MsgClaim{
		SessionHeader: SessionHeader{},
		MerkleRoot:    HashRange{},
		TotalProofs:   0,
		FromAddress:   addr,
	}.GetSigners()
	assert.True(t, reflect.DeepEqual(signers, []types.Address{addr}))
}

func TestMsgClaim_ValidateBasic(t *testing.T) {
	requestorPubKey := getRandomPubKey().RawString()
	servicerAddress := getRandomValidatorAddress()
	ethereum := hex.EncodeToString([]byte{01})
	rootHash := Hash([]byte("fakeRoot"))
	root := HashRange{
		Hash:  rootHash,
		Range: Range{Upper: 100},
	}
	invalidClaimMessageSH := MsgClaim{
		SessionHeader: SessionHeader{
			RequestorPubKey:    "",
			Chain:              ethereum,
			SessionBlockHeight: 1,
		},
		MerkleRoot:   root,
		TotalProofs:  100,
		FromAddress:  servicerAddress,
		EvidenceType: RelayEvidence,
	}
	invalidClaimMessageRoot := MsgClaim{
		SessionHeader: SessionHeader{
			RequestorPubKey:    requestorPubKey,
			Chain:              ethereum,
			SessionBlockHeight: 1,
		},
		MerkleRoot: HashRange{
			Hash: []byte("bad_root"),
		},
		TotalProofs:  100,
		FromAddress:  servicerAddress,
		EvidenceType: RelayEvidence,
	}
	invalidClaimMessageRelays := MsgClaim{
		SessionHeader: SessionHeader{
			RequestorPubKey:    requestorPubKey,
			Chain:              ethereum,
			SessionBlockHeight: 1,
		},
		MerkleRoot:   root,
		TotalProofs:  -1,
		FromAddress:  servicerAddress,
		EvidenceType: RelayEvidence,
	}
	invalidClaimMessageFromAddress := MsgClaim{
		SessionHeader: SessionHeader{
			RequestorPubKey:    requestorPubKey,
			Chain:              ethereum,
			SessionBlockHeight: 1,
		},
		MerkleRoot:   root,
		TotalProofs:  -1,
		FromAddress:  types.Address{},
		EvidenceType: RelayEvidence,
	}
	invalidClaimMessageNoEvidence := MsgClaim{
		SessionHeader: SessionHeader{
			RequestorPubKey:    requestorPubKey,
			Chain:              ethereum,
			SessionBlockHeight: 1,
		},
		MerkleRoot:  root,
		TotalProofs: 100,
		FromAddress: servicerAddress,
	}
	validClaimMessage := MsgClaim{
		SessionHeader: SessionHeader{
			RequestorPubKey:    requestorPubKey,
			Chain:              ethereum,
			SessionBlockHeight: 1,
		},
		MerkleRoot:   root,
		TotalProofs:  100,
		FromAddress:  servicerAddress,
		EvidenceType: RelayEvidence,
	}
	tests := []struct {
		name     string
		msg      MsgClaim
		hasError bool
	}{
		{
			name:     "Invalid Claim Message, session header",
			msg:      invalidClaimMessageSH,
			hasError: true,
		},
		{
			name:     "Invalid Claim Message, root",
			msg:      invalidClaimMessageRoot,
			hasError: true,
		},
		{
			name:     "Invalid Claim Message, relays",
			msg:      invalidClaimMessageRelays,
			hasError: true,
		},
		{
			name:     "Invalid Claim Message, From Address",
			msg:      invalidClaimMessageFromAddress,
			hasError: true,
		},
		{
			name:     "Invalid Claim Message, No Evidence",
			msg:      invalidClaimMessageNoEvidence,
			hasError: true,
		},
		{
			name:     "Valid Claim Message",
			msg:      validClaimMessage,
			hasError: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.msg.ValidateBasic() != nil, tt.hasError)
		})
	}
}

func TestMsgClaim_GetSignBytes(t *testing.T) {
	assert.NotPanics(t, func() { MsgClaim{}.GetSignBytes() })
}

func TestMsgProof_Route(t *testing.T) {
	assert.Equal(t, MsgProof{}.Route(), RouterKey)
}

func TestMsgProof_Type(t *testing.T) {
	assert.Equal(t, MsgProof{}.Type(), MsgProofName)
}

func TestMsgProof_GetSigners(t *testing.T) {
	pk := getRandomPubKey()
	addr := types.Address(pk.Address())
	signers := MsgProof{
		ClaimMerkleProof: MerkleProof{},
		ClaimLeaf: RelayProof{
			Entropy:            0,
			RequestHash:        pk.RawString(), // fake
			SessionBlockHeight: 0,
			ServicerPubKey:     pk.RawString(),
			Blockchain:         "",
			Token:              AAT{},
			Signature:          "",
		},
	}.GetSigners()
	assert.True(t, reflect.DeepEqual(signers, []types.Address{addr}))
}

func TestMsgProof_ValidateBasic(t *testing.T) {
	ethereum := hex.EncodeToString([]byte{01})
	US := hex.EncodeToString([]byte{01})
	servicerPubKey := getRandomPubKey().RawString()
	clientPrivKey := GetRandomPrivateKey()
	clientPubKey := clientPrivKey.PublicKey().RawString()
	requestorPrivKey := GetRandomPrivateKey()
	requestorPubKey := requestorPrivKey.PublicKey().RawString()
	hash1 := merkleHash([]byte("fake1"))
	hash2 := merkleHash([]byte("fake2"))
	hash3 := merkleHash([]byte("fake3"))
	hash4 := merkleHash([]byte("fake4"))
	validProofMessage := MsgProof{
		ClaimMerkleProof: MerkleProof{
			TargetIndex: 0,
			HashRanges: []HashRange{
				{
					Hash:  hash1,
					Range: Range{0, 1},
				},
				{
					Hash:  hash2,
					Range: Range{1, 2},
				},
				{
					Hash:  hash3,
					Range: Range{2, 3},
				},
			},
			Target: HashRange{Hash: hash4, Range: Range{3, 4}},
		},
		ClaimLeaf: RelayProof{
			Entropy:            1,
			SessionBlockHeight: 1,
			ServicerPubKey:     servicerPubKey,
			Blockchain:         ethereum,
			RequestHash:        servicerPubKey, // fake
			Token: AAT{
				Version:            "0.0.1",
				RequestorPublicKey: requestorPubKey,
				ClientPublicKey:    clientPubKey,
				RequestorSignature: "",
			},
			Signature:    "",
			GeoZone:      US,
			NumServicers: 5,
		},
		ClaimEvidenceType: RelayEvidence,
	}
	vprLeaf := validProofMessage.ClaimLeaf.(RelayProof)
	signature, er := requestorPrivKey.Sign(vprLeaf.Token.Hash())
	if er != nil {
		t.Fatalf(er.Error())
	}
	vprLeaf.Token.RequestorSignature = hex.EncodeToString(signature)
	clientSig, er := clientPrivKey.Sign(validProofMessage.ClaimLeaf.(RelayProof).Hash())
	if er != nil {
		t.Fatalf(er.Error())
	}
	vprLeaf.Signature = hex.EncodeToString(clientSig)
	validProofMessage.ClaimLeaf = vprLeaf
	// invalid entropy
	invalidProofMsgIndex := validProofMessage
	//vprLeaf = validProofMessage.Leaf.LegacyFromProto().(*RelayProof)
	vprLeaf.Entropy = 0
	invalidProofMsgIndex.ClaimLeaf = vprLeaf
	// invalid merkleHash sum
	invalidProofMsgHashes := validProofMessage
	invalidProofMsgHashes.ClaimMerkleProof.HashRanges = []HashRange{}
	// invalid session block height
	invalidProofMsgSessionBlkHeight := validProofMessage
	//vprLeaf = validProofMessage.Leaf.LegacyFromProto().(*RelayProof)
	vprLeaf.SessionBlockHeight = -1
	invalidProofMsgSessionBlkHeight.ClaimLeaf = vprLeaf
	// invalid token
	invalidProofMsgToken := validProofMessage
	//vprLeaf = validProofMessage.Leaf.LegacyFromProto().(*RelayProof)
	vprLeaf.Token.RequestorSignature = ""
	invalidProofMsgToken.ClaimLeaf = vprLeaf
	// invalid blockchain
	invalidProofMsgBlkchn := validProofMessage
	//vprLeaf = validProofMessage.Leaf.LegacyFromProto().(*RelayProof)
	vprLeaf.Blockchain = ""
	invalidProofMsgBlkchn.ClaimLeaf = vprLeaf
	// invalid signature
	invalidProofMsgSignature := validProofMessage
	//vprLeaf = validProofMessage.Leaf.LegacyFromProto().(*RelayProof)
	vprLeaf.Signature = hex.EncodeToString([]byte("foobar"))
	invalidProofMsgSignature.ClaimLeaf = vprLeaf
	tests := []struct {
		name     string
		msg      MsgProof
		hasError bool
	}{
		{
			name:     "Invalid Proof Message, signature",
			msg:      invalidProofMsgSignature,
			hasError: true,
		},
		{
			name:     "Invalid Proof Message, session block height",
			msg:      invalidProofMsgSessionBlkHeight,
			hasError: true,
		},
		{
			name:     "Invalid Proof Message, hashsum",
			msg:      invalidProofMsgHashes,
			hasError: true,
		},
		{
			name:     "Invalid Proof Message, leafservicer index",
			msg:      invalidProofMsgIndex,
			hasError: true,
		},
		{
			name:     "Invalid Proof Message, token",
			msg:      invalidProofMsgToken,
			hasError: true,
		},
		{
			name:     "Invalid Proof Message, blockchain",
			msg:      invalidProofMsgBlkchn,
			hasError: true,
		},
		{
			name:     "Valid Proof Message",
			msg:      validProofMessage,
			hasError: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			assert.Equal(t, tt.hasError, err != nil, err)
		})
	}
}

func TestMsgProof_GetSignBytes(t *testing.T) {
	assert.NotPanics(t, func() {
		MsgProof{}.GetSignBytes()
	})
}

func TestMsgSubmitReportCard_Route(t *testing.T) {
	assert.Equal(t, MsgSubmitQoSReport{}.Route(), RouterKey)
}

func TestMsgSubmitReportCard_Type(t *testing.T) {
	assert.Equal(t, MsgSubmitQoSReport{}.Type(), MsgClaimName)
}

func TestMsgSubmitReportCard_GetSigners(t *testing.T) {
	addr := getRandomValidatorAddress()
	faddr := getRandomValidatorAddress()
	signers := MsgSubmitQoSReport{
		SessionHeader:    SessionHeader{},
		ServicerAddress:  addr,
		FishermanAddress: faddr,
		Report: ViperQoSReport{
			FirstSampleTimestamp: time.Now(),
			BlockHeight:          1,
			ServicerAddress:      addr,
			LatencyScore:         types.NewDecWithPrec(12345, 6),
			AvailabilityScore:    types.NewDecWithPrec(67890, 6),
			ReliabilityScore:     types.NewDecWithPrec(54321, 6),
			SampleRoot:           HashRange{Hash: []byte("sample_root_hash"), Range: Range{Lower: 0, Upper: 10}},
			Nonce:                int64(42),
			Signature:            "sample_signature",
		},
		EvidenceType: FishermanTestEvidence,
	}.GetSigners()
	assert.True(t, reflect.DeepEqual(signers, []types.Address{faddr}))
}

func TestMsgSubmitReportCard_ValidateBasic(t *testing.T) {
	servicerAddress := getRandomValidatorAddress()
	fishermanAddress := getRandomValidatorAddress()
	latencyScore := types.NewDecWithPrec(12345, 6)
	availabilityScore := types.NewDecWithPrec(67890, 6)
	reliabilityScore := types.NewDecWithPrec(54321, 6)
	sampleRoot := HashRange{Hash: Hash([]byte("sampleRoot")), Range: Range{Upper: 100}}

	invalidReportMissingServicerAddress := MsgSubmitQoSReport{
		SessionHeader: SessionHeader{
			RequestorPubKey:    "",
			Chain:              "ethereum",
			SessionBlockHeight: 1,
		},
		ServicerAddress:  types.Address{},
		FishermanAddress: fishermanAddress,
		Report: ViperQoSReport{
			FirstSampleTimestamp: time.Now(),
			BlockHeight:          1,
			ServicerAddress:      types.Address{},
			LatencyScore:         latencyScore,
			AvailabilityScore:    availabilityScore,
			ReliabilityScore:     reliabilityScore,
			SampleRoot:           sampleRoot,
			Nonce:                42,
			Signature:            "sample_signature",
		},
		EvidenceType: FishermanTestEvidence,
	}

	invalidReportNegativeBlockHeight := MsgSubmitQoSReport{
		SessionHeader: SessionHeader{
			RequestorPubKey:    "",
			Chain:              "ethereum",
			SessionBlockHeight: -1,
		},
		ServicerAddress:  servicerAddress,
		FishermanAddress: fishermanAddress,
		Report: ViperQoSReport{
			FirstSampleTimestamp: time.Now(),
			BlockHeight:          -1,
			ServicerAddress:      servicerAddress,
			LatencyScore:         latencyScore,
			AvailabilityScore:    availabilityScore,
			ReliabilityScore:     reliabilityScore,
			SampleRoot:           sampleRoot,
			Nonce:                42,
			Signature:            "sample_signature",
		},
		EvidenceType: FishermanTestEvidence,
	}

	invalidReportNegativeScores := MsgSubmitQoSReport{
		SessionHeader: SessionHeader{
			RequestorPubKey:    "",
			Chain:              "ethereum",
			SessionBlockHeight: 1,
		},
		ServicerAddress:  servicerAddress,
		FishermanAddress: fishermanAddress,
		Report: ViperQoSReport{
			FirstSampleTimestamp: time.Now(),
			BlockHeight:          1,
			ServicerAddress:      servicerAddress,
			LatencyScore:         types.NewDecWithPrec(-1, 6),
			AvailabilityScore:    types.NewDecWithPrec(-1, 6),
			ReliabilityScore:     types.NewDecWithPrec(-1, 6),
			SampleRoot:           sampleRoot,
			Nonce:                42,
			Signature:            "sample_signature",
		},
		EvidenceType: FishermanTestEvidence,
	}

	invalidReportInvalidSampleRoot := MsgSubmitQoSReport{
		SessionHeader: SessionHeader{
			RequestorPubKey:    "",
			Chain:              "ethereum",
			SessionBlockHeight: 1,
		},
		ServicerAddress:  servicerAddress,
		FishermanAddress: fishermanAddress,
		Report: ViperQoSReport{
			FirstSampleTimestamp: time.Now(),
			BlockHeight:          1,
			ServicerAddress:      servicerAddress,
			LatencyScore:         latencyScore,
			AvailabilityScore:    availabilityScore,
			ReliabilityScore:     reliabilityScore,
			SampleRoot:           HashRange{Hash: []byte("invalidSampleRoot")},
			Nonce:                42,
			Signature:            "sample_signature",
		},
		EvidenceType: FishermanTestEvidence,
	}

	invalidReportNegativeNonce := MsgSubmitQoSReport{
		SessionHeader: SessionHeader{
			RequestorPubKey:    "",
			Chain:              "ethereum",
			SessionBlockHeight: 1,
		},
		ServicerAddress:  servicerAddress,
		FishermanAddress: fishermanAddress,
		Report: ViperQoSReport{
			FirstSampleTimestamp: time.Now(),
			BlockHeight:          1,
			ServicerAddress:      servicerAddress,
			LatencyScore:         latencyScore,
			AvailabilityScore:    availabilityScore,
			ReliabilityScore:     reliabilityScore,
			SampleRoot:           sampleRoot,
			Nonce:                -1,
			Signature:            "sample_signature",
		},
		EvidenceType: FishermanTestEvidence,
	}

	invalidReportMissingSignature := MsgSubmitQoSReport{
		SessionHeader: SessionHeader{
			RequestorPubKey:    "",
			Chain:              "ethereum",
			SessionBlockHeight: 1,
		},
		ServicerAddress:  servicerAddress,
		FishermanAddress: fishermanAddress,
		Report: ViperQoSReport{
			FirstSampleTimestamp: time.Now(),
			BlockHeight:          1,
			ServicerAddress:      servicerAddress,
			LatencyScore:         latencyScore,
			AvailabilityScore:    availabilityScore,
			ReliabilityScore:     reliabilityScore,
			SampleRoot:           sampleRoot,
			Nonce:                42,
			Signature:            "",
		},
		EvidenceType: FishermanTestEvidence,
	}

	validReport := MsgSubmitQoSReport{
		SessionHeader: SessionHeader{
			RequestorPubKey:    "",
			Chain:              "ethereum",
			SessionBlockHeight: 1,
		},
		ServicerAddress:  servicerAddress,
		FishermanAddress: fishermanAddress,
		Report: ViperQoSReport{
			FirstSampleTimestamp: time.Now(),
			BlockHeight:          1,
			ServicerAddress:      servicerAddress,
			LatencyScore:         latencyScore,
			AvailabilityScore:    availabilityScore,
			ReliabilityScore:     reliabilityScore,
			SampleRoot:           sampleRoot,
			Nonce:                42,
			Signature:            "sample_signature",
		},
		EvidenceType: FishermanTestEvidence,
	}

	tests := []struct {
		name     string
		msg      MsgSubmitQoSReport
		hasError bool
	}{
		{
			name:     "Invalid Report, Missing Servicer Address",
			msg:      invalidReportMissingServicerAddress,
			hasError: true,
		},
		{
			name:     "Invalid Report, Negative Block Height",
			msg:      invalidReportNegativeBlockHeight,
			hasError: true,
		},
		{
			name:     "Invalid Report, Negative Scores",
			msg:      invalidReportNegativeScores,
			hasError: true,
		},
		{
			name:     "Invalid Report, Invalid Sample Root",
			msg:      invalidReportInvalidSampleRoot,
			hasError: true,
		},
		{
			name:     "Invalid Report, Negative Nonce",
			msg:      invalidReportNegativeNonce,
			hasError: true,
		},
		{
			name:     "Invalid Report, Missing Signature",
			msg:      invalidReportMissingSignature,
			hasError: true,
		},
		{
			name:     "Valid Report",
			msg:      validReport,
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.msg.ValidateBasic() != nil, tt.hasError)
		})
	}
}
