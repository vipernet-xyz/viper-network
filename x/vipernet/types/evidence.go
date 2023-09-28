package types

import (
	"fmt"
	"strings"

	"github.com/vipernet-xyz/viper-network/codec"
	"github.com/vipernet-xyz/viper-network/types"
	sdk "github.com/vipernet-xyz/viper-network/types"

	"github.com/willf/bloom"
)

// "Evidence" - A proof of work/burn for servicers.
type Evidence struct {
	Bloom         bloom.BloomFilter        `json:"bloom_filter"` // used to check if proof contains
	SessionHeader `json:"evidence_header"` // the session h serves as an identifier for the evidence
	NumOfProofs   int64                    `json:"num_of_proofs"` // the total number of proofs in the evidence
	Proofs        Proofs                   `json:"proofs"`        // a slice of Proof objects (Proof per relay or challenge)
	EvidenceType  EvidenceType             `json:"evidence_type"`
}

type Result struct {
	SessionHeader    `json:"evidence_header"`
	ServicerAddr     sdk.Address  `json:"servicer_addr"`
	NumOfTestResults int64        `json:"num_of_test_results"`
	TestResults      Tests        `json:"tests"`
	EvidenceType     EvidenceType `json:"evidence_type"`
}

func (e Evidence) IsSealable() bool {
	return true
}

func (r Result) IsSealable() bool {
	return true
}

// "GenerateMerkleRoot" - Generates the merkle root for an GOBEvidence object
func (e *Evidence) GenerateMerkleRoot(height int64, maxRelays int64, storage *CacheStorage) (root HashRange) {
	// seal the evidence in cache/db
	ev, ok := SealEvidence(*e, storage)
	if !ok {
		return HashRange{}
	}
	if int64(len(ev.Proofs)) > maxRelays {
		ev.Proofs = ev.Proofs[:maxRelays]
		ev.NumOfProofs = maxRelays
	}
	// generate the root object
	root, _ = GenerateRoot(height, ev.Proofs)
	return
}

// "GenerateMerkleRoot" - Generates the merkle root for an GOBEvidence object
func (r *Result) GenerateSampleMerkleRoot(height int64, storage *CacheStorage) (root HashRange) {
	// seal the evidence in cache/db
	ev, ok := SealResult(*r, storage)
	if !ok {
		return HashRange{}
	}
	// generate the root object
	root, _ = GenerateSampleRoot(height, ev.TestResults)
	return
}

// "AddProof" - Adds a proof obj to the GOBEvidence field
func (e *Evidence) AddProof(p Proof) {
	// add proof to GOBEvidence
	e.Proofs = append(e.Proofs, p)
	// increment total proof count
	e.NumOfProofs = e.NumOfProofs + 1
	// add proof to bloom filter
	e.Bloom.Add(p.Hash())
}

func (r *Result) AddTestResult(t Test) {
	// add proof to GOBEvidence
	r.TestResults = append(r.TestResults, t)
	// increment total proof count
	r.NumOfTestResults = r.NumOfTestResults + 1
}

// "GenerateMerkleProof" - Generates the merkle Proof for an GOBEvidence
func (e *Evidence) GenerateMerkleProof(height int64, index int, maxRelays int64) (proof MerkleProof, leaf Proof) {
	if int64(len(e.Proofs)) > maxRelays {
		e.Proofs = e.Proofs[:maxRelays]
		e.NumOfProofs = maxRelays
	}
	// generate the merkle proof
	proof, leaf = GenerateProofs(height, e.Proofs, index)
	// set the evidence in memory
	return
}

// "Evidence" - A proof of work/burn for servicers.
type evidence struct {
	BloomBytes    []byte                   `json:"bloom_bytes"`
	SessionHeader `json:"evidence_header"` // the session h serves as an identifier for the evidence
	NumOfProofs   int64                    `json:"num_of_proofs"` // the total number of proofs in the evidence
	Proofs        []Proof                  `json:"proofs"`        // a slice of Proof objects (Proof per relay or challenge)
	EvidenceType  EvidenceType             `json:"evidence_type"`
}

func (e Evidence) LegacyAminoMarshal() ([]byte, error) {
	encodedBloom, err := e.Bloom.GobEncode()
	if err != nil {
		return nil, err
	}
	ep := evidence{
		BloomBytes:    encodedBloom,
		SessionHeader: e.SessionHeader,
		NumOfProofs:   e.NumOfProofs,
		Proofs:        e.Proofs,
		EvidenceType:  e.EvidenceType,
	}
	return ModuleCdc.MarshalBinaryBare(ep)
}

func (e Evidence) LegacyAminoUnmarshal(b []byte) (CacheObject, error) {
	ep := evidence{}
	err := ModuleCdc.UnmarshalBinaryBare(b, &ep)
	if err != nil {
		return Evidence{}, fmt.Errorf("could not unmarshal into evidence from cache, moduleCdc unmarshal binary bare: %s", err.Error())
	}
	bloomFilter := bloom.BloomFilter{}
	err = bloomFilter.GobDecode(ep.BloomBytes)
	if err != nil {
		return Evidence{}, fmt.Errorf("could not unmarshal into evidence from cache, bloom bytes gob decode: %s", err.Error())
	}
	evidence := Evidence{
		Bloom:         bloomFilter,
		SessionHeader: ep.SessionHeader,
		NumOfProofs:   ep.NumOfProofs,
		Proofs:        ep.Proofs,
		EvidenceType:  ep.EvidenceType,
	}
	return evidence, nil
}

var (
	_ CacheObject          = Evidence{} // satisfies the cache object interface
	_ codec.ProtoMarshaler = &Evidence{}
)

func (e *Evidence) Reset() {
	*e = Evidence{}
}

func (e *Evidence) String() string {
	return fmt.Sprintf("SessionHeader: %v\nNumOfProofs: %v\nProofs: %v\nEvidenceType: %vBloomFilter: %v\n",
		e.SessionHeader, e.NumOfProofs, e.Proofs, e.EvidenceType, e.Bloom)
}

func (e *Evidence) ProtoMessage() {}

func (e *Evidence) Marshal() ([]byte, error) {
	pe, err := e.ToProto()
	if err != nil {
		return nil, err
	}
	return pe.Marshal()
}

func (e *Evidence) MarshalTo(data []byte) (n int, err error) {
	pe, err := e.ToProto()
	if err != nil {
		return 0, err
	}
	return pe.MarshalTo(data)
}

func (r *Result) MarshalTo(data []byte) (n int, err error) {
	pr, err := r.ToProto()
	if err != nil {
		return 0, err
	}
	return pr.MarshalTo(data)
}

func (e *Evidence) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	pe, err := e.ToProto()
	if err != nil {
		return 0, err
	}
	return pe.MarshalToSizedBuffer(dAtA)
}

func (e *Evidence) Size() int {
	pe, err := e.ToProto()
	if err != nil {
		return 0
	}
	return pe.Size()
}

func (e *Evidence) Unmarshal(data []byte) error {
	pe := ProtoEvidence{}
	err := pe.Unmarshal(data)
	if err != nil {
		return err
	}
	*e, err = pe.FromProto()
	return err
}

func (e *Evidence) ToProto() (*ProtoEvidence, error) {
	encodedBloom, err := e.Bloom.GobEncode()
	if err != nil {
		return nil, err
	}
	return &ProtoEvidence{
		BloomBytes:    encodedBloom,
		SessionHeader: &e.SessionHeader,
		NumOfProofs:   e.NumOfProofs,
		Proofs:        e.Proofs.ToProofI(),
		EvidenceType:  e.EvidenceType,
	}, nil
}

func (r *Result) ToProto() (*ProtoResult, error) {
	return &ProtoResult{
		SessionHeader:    &r.SessionHeader,
		ServicerAddr:     r.ServicerAddr,
		NumOfTestResults: r.NumOfTestResults,
		TestResults:      r.TestResults.ToTestI(),
		EvidenceType:     r.EvidenceType,
	}, nil
}

func (pe *ProtoEvidence) FromProto() (Evidence, error) {
	bloomFilter := bloom.BloomFilter{}
	err := bloomFilter.GobDecode(pe.BloomBytes)
	if err != nil {
		return Evidence{}, fmt.Errorf("could not unmarshal into ProtoEvidence from cache, bloom bytes gob decode: %s", err.Error())
	}
	return Evidence{
		Bloom:         bloomFilter,
		SessionHeader: *pe.SessionHeader,
		NumOfProofs:   pe.NumOfProofs,
		Proofs:        pe.Proofs.FromProofI(),
		EvidenceType:  pe.EvidenceType}, nil
}

func (pr *ProtoResult) FromProto() (Result, error) {
	return Result{
		SessionHeader:    *pr.SessionHeader,
		ServicerAddr:     pr.ServicerAddr,
		NumOfTestResults: pr.NumOfTestResults,
		TestResults:      pr.TestResults.FromTestI(),
		EvidenceType:     pr.EvidenceType}, nil
}

func (e Evidence) MarshalObject() ([]byte, error) {
	pe, err := e.ToProto()
	if err != nil {
		return nil, err
	}
	return ModuleCdc.ProtoMarshalBinaryBare(pe)
}

func (r Result) MarshalObject() ([]byte, error) {
	pr, err := r.ToProto()
	if err != nil {
		return nil, err
	}
	return ModuleCdc.ProtoMarshalBinaryBare(pr)
}

func (e Evidence) UnmarshalObject(b []byte) (CacheObject, error) {
	pe := ProtoEvidence{}
	err := ModuleCdc.ProtoUnmarshalBinaryBare(b, &pe)
	if err != nil {
		return Evidence{}, fmt.Errorf("could not unmarshal into ProtoEvidence from cache, moduleCdc unmarshal binary bare: %s", err.Error())
	}
	return pe.FromProto()
}

func (r Result) UnmarshalObject(b []byte) (CacheObject, error) {
	pr := ProtoResult{}
	err := ModuleCdc.ProtoUnmarshalBinaryBare(b, &pr)
	if err != nil {
		return Result{}, fmt.Errorf("could not unmarshal into ProtoResult from cache, moduleCdc unmarshal binary bare: %s", err.Error())
	}
	return pr.FromProto()
}

func (e Evidence) Key() ([]byte, error) {
	return KeyForEvidence(e.SessionHeader, e.EvidenceType)
}

func (r Result) Key() ([]byte, error) {
	return KeyForTestResult(r.SessionHeader, r.EvidenceType, r.ServicerAddr)
}

// "EvidenceType" type to distinguish the types of GOBEvidence (relay/challenge)
type EvidenceType int

const (
	RelayEvidence EvidenceType = iota + 1 // essentially an enum for GOBEvidence types
	ChallengeEvidence
	FishermanTestEvidence
)

// "Convert GOBEvidence type to bytes
func (et EvidenceType) Byte() (byte, error) {
	switch et {
	case RelayEvidence:
		return 0, nil
	case ChallengeEvidence:
		return 1, nil
	case FishermanTestEvidence:
		return 2, nil
	default:
		return 0, fmt.Errorf("unrecognized GOBEvidence type")
	}
}

func EvidenceTypeFromString(evidenceType string) (et EvidenceType, err types.Error) {
	switch strings.ToLower(evidenceType) {
	case "relay":
		et = RelayEvidence
	case "challenge":
		et = ChallengeEvidence
	case "test":
		et = FishermanTestEvidence
	default:
		err = types.ErrInternal("type in the receipt query is not recognized: (relay or challenge)")
	}
	return
}
