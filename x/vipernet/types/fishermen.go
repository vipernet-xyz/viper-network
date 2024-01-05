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

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/servicers/exported"
	servicerTypes "github.com/vipernet-xyz/viper-network/x/servicers/types"
)

func SendSampleRelay(Blockchain string, trigger FishermenTrigger, servicer exported.ValidatorI, fishermanValidator exported.ValidatorI) (*Output, error) {

	// First, we will ensure SampleRelayPools is loaded
	err := LoadSampleRelayPool()
	if err != nil {
		return nil, fmt.Errorf("Failed to load SampleRelayPools: %v", err)
	}

	start := time.Now()
	//Get the appropriate relay pool for the blockchain
	relayPool, exists := SampleRelayPools[Blockchain]
	if !exists {
		return nil, fmt.Errorf("no relay pool found for blockchain: %s", Blockchain)
	}

	//Select a random relay payload from the pool
	randIndex := rand.Intn(len(relayPool.Payloads))
	samplePayload := relayPool.Payloads[randIndex]

	//Create a RelayMeta and RelayProof
	relayMeta := &RelayMeta{
		BlockHeight: trigger.Proof.SessionBlockHeight,
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

	rpcURL := fishermanValidator.GetServiceURL()
	sender := NewSender(rpcURL)
	signer, err := NewSigner(fishermanValidator)
	if err != nil {
		return nil, err
	}
	relayer := NewRelayer(*signer, *sender)
	signedProofBytes, err := relayer.getSignedProofBytes(&RelayProof{
		RequestHash:        reqHash,
		Entropy:            entropy.Int64(),
		SessionBlockHeight: relayMeta.BlockHeight,
		ServicerPubKey:     servicer.GetAddress().String(),
		Blockchain:         Blockchain,
		Token:              trigger.Proof.Token,
		GeoZone:            trigger.Proof.GeoZone,
		NumServicers:       trigger.Proof.NumServicers,
	})
	if err != nil {
		return nil, err
	}

	//Prepare a RelayInput using the generated details
	relay := &RelayInput{
		Payload: samplePayload,
		Meta:    relayMeta,
		Proof: &RelayProof{
			RequestHash:        reqHash,
			Entropy:            entropy.Int64(),
			SessionBlockHeight: relayMeta.BlockHeight,
			ServicerPubKey:     servicer.GetAddress().String(),
			Blockchain:         Blockchain,
			Token:              trigger.Proof.Token,
			Signature:          signedProofBytes,
			GeoZone:            trigger.Proof.GeoZone,
			NumServicers:       trigger.Proof.NumServicers,
		},
	}
	// Send the relay to the servicer and measure its latency
	relayOutput, err := sender.Relay(servicer.GetServiceURL(), relay)
	servicerLatency := time.Since(start)

	if err != nil {
		return nil, err
	}

	servicerReliability := true
	var localResp *RelayOutput

	// Execute the local relay only if the request method is GET
	if samplePayload.Method == "GET" {
		c := make(chan struct{}, 1)
		go func() {
			localResp, err = sender.localRelay(fishermanValidator.GetServiceURL(), relay)
			c <- struct{}{}
		}()

		select {
		case <-c:
			if err != nil {
				return nil, fmt.Errorf("Failed to execute relay internally within fisherman: %s", err.Error())
			}

			if relayOutput.Response != localResp.Response {
				// This is a discrepancy. Handle accordingly.
				servicerReliability = false // set reliability to false if there's a mismatch
			}
		case <-time.After(servicerLatency): // time allotted for local relay is same as the servicer's latency
			// If the local relay takes more than the servicer's latency, continue with the process
			servicerReliability = true // assuming the relay was reliable as it didn't return an error
		}
	} else {
		// If the request method is not GET, just check the response for errors
		if err != nil {
			servicerReliability = false
		}
	}

	return &Output{
		RelayOutput: relayOutput,
		LocalResp:   localResp,
		Proof:       relay.Proof,
		Latency:     servicerLatency,
		Reliability: servicerReliability,
	}, nil
}

type FishermenTrigger struct {
	Proof RelayProof
}

type V1RPCRoute string

const (
	ClientRelayRoute V1RPCRoute = "/v1/client/relay"
	LocalRelayRoute  V1RPCRoute = "/v1/client/localrelay"
)

// "SessionHeader" - Returns the session header corresponding with the proof
func (ft FishermenTrigger) SessionHeader() SessionHeader {
	return SessionHeader{
		ProviderPubKey:     ft.Proof.Token.ProviderPublicKey,
		Chain:              ft.Proof.Blockchain,
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
	RelayOutput *RelayOutput
	LocalResp   *RelayOutput
	Proof       *RelayProof
	Latency     time.Duration
	Reliability bool
}

type RelayOutput struct {
	Response  string `json:"response"`
	Signature string `json:"signature"`
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

func CalculateQoSForServicer(result *ServicerResults, blockHeight int64, latencyScore sdk.BigDec) (*ViperQoSReport, error) {
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
		BlockHeight:          blockHeight,
		ServicerAddress:      result.ServicerAddress,
		AvailabilityScore:    scaledAvailabilityScore,
		ReliabilityScore:     reliabilityScore,
		LatencyScore:         latencyScore, // Set the latency score
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
