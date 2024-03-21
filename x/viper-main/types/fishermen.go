package types

import (
	rand1 "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	math "math"
	"math/big"
	"math/rand"
	"time"

	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/servicers/exported"
	servicerTypes "github.com/vipernet-xyz/viper-network/x/servicers/types"
)

func (r *Relayer) SendSampleRelay(blockHeight int64, Blockchain string, trigger FishermenTrigger, servicer exported.ValidatorI, fishermanValidator exported.ValidatorI, hostedBlockchains *HostedBlockchains) (*Output, error) {

	// First, we will ensure SampleRelayPools is loaded
	pools, err := LoadSampleRelayPool()

	if err != nil {
		return nil, fmt.Errorf("failed to load SampleRelayPools: %v", err)
	}

	start := time.Now()
	//Get the appropriate relay pool for the blockchain
	relayPool, exists := pools[Blockchain]
	if !exists {
		return nil, fmt.Errorf("no relay pool found for blockchain: %s", Blockchain)
	}

	//Select a random relay payload from the pool
	randIndex := rand.Intn(len(relayPool.Payloads))
	samplePayload := relayPool.Payloads[randIndex]

	//Create a RelayMeta and RelayProof
	relayMeta := &RelayMeta{
		BlockHeight: blockHeight,
	}

	//Hash the request
	reqHash, err := HashRequest(&RequestHash{
		Payload: samplePayload,
		Meta:    relayMeta,
	})
	if err != nil {
		return nil, err
	}

	//Generate entropy and signed proof bytes
	entropy, err := rand1.Int(rand1.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return nil, err
	}

	signedProofBytes, err := r.getSignedProofBytes(&RelayProof{
		RequestHash:        reqHash,
		Entropy:            entropy.Int64(),
		SessionBlockHeight: blockHeight,
		ServicerPubKey:     servicer.GetPublicKey().RawString(),
		Blockchain:         Blockchain,
		GeoZone:            trigger.Proof.GeoZone,
		NumServicers:       trigger.Proof.NumServicers,
		Token:              trigger.Proof.Token,
		Signature:          "",
	}, trigger.Account)
	if err != nil {
		return nil, err
	}

	//Prepare a RelayInput using the generated details
	relay := RelayInput{
		Payload: samplePayload,
		Meta:    relayMeta,
		Proof: &RelayProof{
			RequestHash:        reqHash,
			Entropy:            entropy.Int64(),
			SessionBlockHeight: blockHeight,
			ServicerPubKey:     servicer.GetPublicKey().RawString(),
			Blockchain:         Blockchain,
			GeoZone:            trigger.Proof.GeoZone,
			NumServicers:       trigger.Proof.NumServicers,
			Token:              trigger.Proof.Token,
			Signature:          signedProofBytes,
		},
	}

	// Send the relay to the servicer and measure its latency
	relayOutput, err := r.sender.Relay(servicer.GetServiceURL(), &relay)

	servicerLatency := time.Since(start)

	servicerReliability := false

	sevicerAvailability := false

	var localResp string

	if err != nil {
		// If no response received, mark latency as timeout latency and availability as false
		servicerLatency = 0
		return &Output{
			RelayOutput:  relayOutput,
			LocalResp:    localResp,
			Proof:        relay.Proof,
			Latency:      servicerLatency,
			Availability: sevicerAvailability,
			Reliability:  servicerReliability,
		}, err
	} else {
		// Perform local relay only if we receive a non-empty response
		addr := fishermanValidator.GetAddress()
		fishermanAddress := &addr

		localResp, err = relay.ExecuteLocal(hostedBlockchains, fishermanAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to execute relay internally within fisherman: %s", err)
		}

		sevicerAvailability = true
		// Check if the response from the servicer matches the local response
		servicerReliability = (relayOutput.Response == localResp)
	}

	return &Output{
		RelayOutput:  relayOutput,
		LocalResp:    localResp,
		Proof:        relay.Proof,
		Latency:      servicerLatency,
		Availability: sevicerAvailability,
		Reliability:  servicerReliability,
	}, nil
}

func (tr *TestResult) Validate(resp Output, sessionHeader SessionHeader, node *ViperNode) error {
	// Retrieve the public key of the servicer from the relay proof
	servicerPubKey := resp.Proof.ServicerPubKey

	pk, err := crypto.NewPublicKey(servicerPubKey)
	if err != nil {
		return NewPubKeyDecodeError(ModuleName)
	}
	servicerAddr := pk.Address().Bytes()
	err = tr.ValidateLocal(servicerAddr)
	if err != nil {
		return err
	}
	resp.RelayOutput.Proof = *resp.Proof
	if err := SignatureVerification(servicerPubKey, resp.RelayOutput.HashString(), resp.RelayOutput.Signature); err != nil {
		return err
	}
	testResults, _ := GetTotalTestResults(resp.Proof.SessionHeader(), FishermanTestEvidence, tr.ServicerAddress, node.SessionStore)
	if !IsUniqueResult(tr, testResults) {
		return NewDuplicateTestResultError(ModuleName)
	}
	return nil
}

func (ro RelayOutput) HashString() string {
	return hex.EncodeToString(ro.Hash())
}

func (ro RelayOutput) Hash() []byte {
	res := ro.Bytes()
	return Hash(res)
}

func (ro RelayOutput) Bytes() []byte {
	res, err := json.Marshal(relayResponse{
		Signature: "",
		Response:  ro.Response,
		Proof:     ro.Proof.HashString(),
	})
	if err != nil {
		log.Fatal(fmt.Errorf("an error occured converting the relay RelayProof to bytes:\n%v", err).Error())
	}
	return res
}

func IsUniqueResult(t Test, result Result) bool {
	hash := t.HashString()
	for _, existingResults := range result.TestResults {
		if existingResults.HashString() == hash {
			return false
		}
	}
	return true
}

type FishermenTrigger struct {
	Proof   RelayProof
	Account Account
}

type V1RPCRoute string

const (
	ClientRelayRoute V1RPCRoute = "/v1/client/relay"
	QueryHeightRoute V1RPCRoute = "/v1/query/height"
)

// "SessionHeader" - Returns the session header corresponding with the proof
func (ft FishermenTrigger) SessionHeader() SessionHeader {
	return SessionHeader{
		RequestorPubKey:    ft.Proof.Token.RequestorPublicKey,
		Chain:              ft.Proof.Blockchain,
		GeoZone:            ft.Proof.GeoZone,
		NumServicers:       ft.Proof.NumServicers,
		SessionBlockHeight: ft.Proof.SessionBlockHeight,
	}
}

// Struct to hold results for a servicer
type ServicerResults struct {
	ServicerAddress sdk.Address
	Timestamps      []time.Time
	Latencies       []time.Duration
	Availabilities  []bool
	Reliabilities   []bool
}

type MProof_Leaf struct {
	MerkleProof MerkleProof
	Leaf        TestI
	NumOfTests  int64
}

type RelayHeaders map[string]string

type RelayPayload struct {
	Data    string       `json:"data"`
	Method  string       `json:"method"`
	Path    string       `json:"path"`
	Headers RelayHeaders `json:"headers"`
}

type Input struct {
	Blockchain string
	Data       string
	Headers    RelayHeaders
	Method     string
	Node       *servicerTypes.Validator
	Path       string
	ViperAAT   *AAT
	Session    *Session
}

// RequestHash struct holding data needed to create a request hash
type RequestHash struct {
	Payload *RelayPayload `json:"payload"`
	Meta    *RelayMeta    `json:"meta"`
}

// Output struct for data needed as output for relay request
type Output struct {
	RelayOutput  *RelayOutput
	LocalResp    string
	Proof        *RelayProof
	Latency      time.Duration
	Availability bool
	Reliability  bool
}

type RelayOutput struct {
	Signature string     `json:"signature"`
	Response  string     `json:"response"`
	Proof     RelayProof `json:"proof"`
}

// Order of fields matters for signature
type relayProofForSignature struct {
	Entropy            int64  `json:"entropy"`
	SessionBlockHeight int64  `json:"session_block_height"`
	ServicerPubKey     string `json:"servicer_pub_key"`
	Blockchain         string `json:"blockchain"`
	Signature          string `json:"signature"`
	Token              string `json:"token"`
	RequestHash        string `json:"request_hash"`
	GeoZone            string `json:"zone"`
	NumServicers       int64  `json:"num_servicers"`
}

// RelayInput represents input needed to do a Relay to Viper
type RelayInput struct {
	Payload *RelayPayload `json:"payload"`
	Meta    *RelayMeta    `json:"meta"`
	Proof   *RelayProof   `json:"proof"`
}

func Shuffle(proofs []Test, rng *rand.Rand) {
	n := len(proofs)
	for i := n - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		proofs[i], proofs[j] = proofs[j], proofs[i]
	}
}

func CalculateQoSForServicer(result *ServicerResults, latencyScore sdk.BigDec) (*ViperQoSReport, error) {
	firstSampleTimestamp := time.Time{}
	if len(result.Timestamps) > 0 {
		firstSampleTimestamp = result.Timestamps[0]
	}

	// Calculate availability score
	_, scaledAvailabilityScore := CalculateAvailabilityScore(len(result.Availabilities), countTrue(result.Availabilities))

	// Calculate reliability score
	reliabilityScore := CalculateReliabilityScore(len(result.Reliabilities), countTrue(result.Reliabilities))

	report := &ViperQoSReport{
		FirstSampleTimestamp: firstSampleTimestamp,
		AvailabilityScore:    scaledAvailabilityScore,
		ReliabilityScore:     reliabilityScore,
		LatencyScore:         latencyScore,
	}
	return report, nil
}

func countTrue(bools []bool) int {
	count := 0
	for _, b := range bools {
		if b {
			count++
		}
	}
	return count
}

func CalculateAvailabilityScore(totalRelays, answeredRelays int) (downtimePercentage sdk.BigDec, scaledAvailabilityScore sdk.BigDec) {
	/*
		If there are 100 total relays and only 80 are answered, then:
		- downtimePercentage = (100 - 80) / 100 = 0.20 (or 20%)
		- scaledAvailabilityScore = 1 - 0.20 = 0.80 (or 80%)
	*/
	downtimePercentage = sdk.NewDecWithPrec(int64(totalRelays-answeredRelays), 0).Quo(sdk.NewDecWithPrec(int64(totalRelays), 0))
	scaledAvailabilityScore = sdk.MaxDec(sdk.ZeroDec(), sdk.OneDec().Sub(downtimePercentage))
	return downtimePercentage, scaledAvailabilityScore
}

func CalculateReliabilityScore(totalSamples int, matchedSamples int) sdk.BigDec {
	if totalSamples == 0 {
		return sdk.ZeroDec()
	}
	return sdk.NewDec(int64(matchedSamples)).Quo(sdk.NewDec(int64(totalSamples)))
}

func SumDurations(durations []time.Duration) time.Duration {
	sum := time.Duration(0)
	for _, d := range durations {
		sum += d
	}
	return sum
}

// Bytes returns the bytes representation of the FishermenTrigger
func (ft FishermenTrigger) Bytes() []byte {
	// Marshal the FishermenTrigger into bytes
	res, err := json.Marshal(ft)
	if err != nil {
		log.Fatal(fmt.Errorf("cannot marshal FishermenTrigger: %s", err.Error()))
	}
	return res
}

// "Requesthash" - The cryptographic merkleHash representation of the request
func (ft FishermenTrigger) RequestHash() []byte {
	return Hash(ft.Bytes())
}

// "RequestHashString" - The hex string representation of the request merkleHash
func (ft FishermenTrigger) RequestHashString() string {
	return hex.EncodeToString(ft.RequestHash())
}

type report struct {
	FirstSampleTimestamp time.Time   `json:"first_sample_timestamp"`
	ServicerAddress      sdk.Address `json:"servicer_addr"`
	LatencyScore         sdk.BigDec  `json:"latency_score"`
	AvailabilityScore    sdk.BigDec  `json:"availability_score"`
	ReliabilityScore     sdk.BigDec  `json:"reliability_score"`
	SampleRoot           HashRange   `json:"sample_root"`
	Nonce                int64       `json:"nonce"`
	Signature            string      `json:"signature"`
}

func (vr ViperQoSReport) HashString() string {
	return hex.EncodeToString(vr.Hash())
}

func (vr ViperQoSReport) Hash() []byte {
	res := vr.Bytes()
	return Hash(res)
}

func (vr ViperQoSReport) Bytes() []byte {
	res, err := json.Marshal(report{
		FirstSampleTimestamp: vr.FirstSampleTimestamp,
		ServicerAddress:      vr.ServicerAddress,
		LatencyScore:         vr.LatencyScore,
		AvailabilityScore:    vr.AvailabilityScore,
		ReliabilityScore:     vr.ReliabilityScore,
		SampleRoot:           vr.SampleRoot,
		Nonce:                vr.Nonce,
		Signature:            "",
	})
	if err != nil {
		log.Fatal(fmt.Errorf("an error occured converting the report to bytes:\n%v", err).Error())
	}
	return res
}
