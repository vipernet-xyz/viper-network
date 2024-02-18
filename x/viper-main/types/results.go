package types

import (
	"fmt"

	"github.com/vipernet-xyz/viper-network/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
)

type Result struct {
	SessionHeader    `json:"evidence_header"`
	ServicerAddr     sdk.Address  `json:"servicer_addr"`
	NumOfTestResults int64        `json:"num_of_test_results"`
	TestResults      Tests        `json:"tests"`
	EvidenceType     EvidenceType `json:"evidence_type"`
}

func (r Result) IsSealable() bool {
	return true
}

// "GenerateMerkleRoot" - Generates the merkle root for an GOBEvidence object
func (r *Result) GenerateSampleMerkleRoot(height int64, storage *CacheStorage) (root HashRange) {
	// seal the evidence in cache/db
	re, ok := SealResult(*r, storage)
	if !ok {
		return HashRange{}
	}
	// generate the root object
	root, _ = GenerateSampleRoot(height, re.TestResults)
	return
}

func (r *Result) AddTestResult(t Test) {
	// add proof to GOBEvidence
	r.TestResults = append(r.TestResults, t)
	// increment total proof count
	r.NumOfTestResults = r.NumOfTestResults + 1
}

// "GenerateMerkleProof" - Generates the merkle Proof for an GOBEvidence
func (r *Result) GenerateMerkleProof(height int64, index int) (test MerkleProof, leaf Test) {
	// generate the merkle proof
	test, leaf = GenerateTRProofs(height, r.TestResults, index)
	// set the evidence in memory
	return
}

// "Evidence" - A proof of work/burn for servicers.
type result struct {
	SessionHeader    `json:"evidence_header"`
	ServicerAddr     sdk.Address  `json:"servicer_addr"`
	NumOfTestResults int64        `json:"num_of_test_results"`
	TestResults      []Test       `json:"tests"`
	EvidenceType     EvidenceType `json:"evidence_type"`
}

func (r Result) LegacyAminoMarshal() ([]byte, error) {
	re := result{
		SessionHeader:    r.SessionHeader,
		NumOfTestResults: r.NumOfTestResults,
		TestResults:      r.TestResults,
		EvidenceType:     r.EvidenceType,
	}
	return ModuleCdc.MarshalBinaryBare(re)
}

func (r Result) LegacyAminoUnmarshal(b []byte) (CacheObject, error) {
	re := result{}
	err := ModuleCdc.UnmarshalBinaryBare(b, &re)
	if err != nil {
		return Result{}, fmt.Errorf("could not unmarshal into evidence from cache, moduleCdc unmarshal binary bare: %s", err.Error())
	}
	evidence := Result{
		SessionHeader:    re.SessionHeader,
		NumOfTestResults: re.NumOfTestResults,
		TestResults:      re.TestResults,
		EvidenceType:     re.EvidenceType,
	}
	return evidence, nil
}

var (
	_ CacheObject          = Result{} // satisfies the cache object interface
	_ codec.ProtoMarshaler = &Result{}
)

func (r *Result) Reset() {
	*r = Result{}
}

func (r *Result) String() string {
	return fmt.Sprintf("SessionHeader: %v\nNumOfTestResults: %v\nTestResults: %v\nEvidenceType: %vServicerAddr: %v\n",
		r.SessionHeader, r.NumOfTestResults, r.TestResults, r.EvidenceType, r.ServicerAddr)
}

func (r *Result) ProtoMessage() {}

func (e *Result) Marshal() ([]byte, error) {
	pe, err := e.ToProto()
	if err != nil {
		return nil, err
	}
	return pe.Marshal()
}

func (r *Result) MarshalTo(data []byte) (n int, err error) {
	pr, err := r.ToProto()
	if err != nil {
		return 0, err
	}
	return pr.MarshalTo(data)
}

func (r *Result) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	pe, err := r.ToProto()
	if err != nil {
		return 0, err
	}
	return pe.MarshalToSizedBuffer(dAtA)
}

func (r *Result) Size() int {
	pe, err := r.ToProto()
	if err != nil {
		return 0
	}
	return pe.Size()
}

func (r *Result) Unmarshal(data []byte) error {
	pe := ProtoResult{}
	err := pe.Unmarshal(data)
	if err != nil {
		return err
	}
	*r, err = pe.FromProto()
	return err
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

func (pr *ProtoResult) FromProto() (Result, error) {
	return Result{
		SessionHeader:    *pr.SessionHeader,
		ServicerAddr:     pr.ServicerAddr,
		NumOfTestResults: pr.NumOfTestResults,
		TestResults:      pr.TestResults.FromTestI(),
		EvidenceType:     pr.EvidenceType}, nil
}

func (r Result) MarshalObject() ([]byte, error) {
	pr, err := r.ToProto()
	if err != nil {
		return nil, err
	}
	return ModuleCdc.ProtoMarshalBinaryBare(pr)
}

func (r Result) UnmarshalObject(b []byte) (CacheObject, error) {
	pr := ProtoResult{}
	err := ModuleCdc.ProtoUnmarshalBinaryBare(b, &pr)
	if err != nil {
		return Result{}, fmt.Errorf("could not unmarshal into ProtoResult from cache, moduleCdc unmarshal binary bare: %s", err.Error())
	}
	return pr.FromProto()
}

func (r Result) Key() ([]byte, error) {
	return KeyForTestResult(r.SessionHeader, r.EvidenceType, r.ServicerAddr)
}
