package types

import (
	rand1 "crypto/rand"
	"fmt"
	math "math"
	"math/big"
	"math/rand"
	"time"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/servicers/exported"
	servicerTypes "github.com/vipernet-xyz/viper-network/x/servicers/types"
)

func SendSampleRelay(Blockchain string, trigger FishermenTrigger, servicer exported.ValidatorI, fishermanValidator exported.ValidatorI) (*Output, error) {
	//Get the appropriate relay pool for the blockchain
	relayPool, exists := SampleRelayPools[Blockchain]
	if !exists {
		return nil, fmt.Errorf("no relay pool found for blockchain: %s", Blockchain)
	}

	//Select a random relay payload from the pool
	randIndex := rand.Intn(len(relayPool.Payloads))
	samplePayload := relayPool.Payloads[randIndex]

	//Create a RelayMeta and RelayProof, assuming you can derive the required details for these
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
	// Assuming a function like getSignedProofBytes exists in the current scope
	signedProofBytes, err := relayer.getSignedProofBytes(&RelayProof{
		RequestHash:        reqHash,
		Entropy:            entropy.Int64(),
		SessionBlockHeight: relayMeta.BlockHeight,
		ServicerPubKey:     servicer.GetAddress().String(),
		Blockchain:         Blockchain,
		Token:              trigger.Proof.Token,
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
			Signature:          signedProofBytes,
		},
	}
	//Send the relay using your Relay function
	relayOutput, err := sender.Relay(servicer.GetServiceURL(), relay)
	if err != nil {
		return nil, err
	}

	return &Output{
		RelayOutput: relayOutput,
		Proof:       relay.Proof,
	}, nil
}

type FishermenTrigger struct {
	Proof RelayProof
}

type V1RPCRoute string

const (
	ClientRelayRoute V1RPCRoute = "/v1/client/relay"
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
	Proof       *RelayProof
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

func CalculateQoSForServicer(result *ServicerResults, blockHeight int64) (*ViperQoSReport, error) {
	expectedLatency := CalculateExpectedLatency(globalRPCTimeout)

	firstSampleTimestamp := time.Time{}
	if len(result.Timestamps) > 0 {
		firstSampleTimestamp = result.Timestamps[0]
	}

	// Calculate availability score
	_, scaledAvailabilityScore := CalculateAvailabilityScore(len(result.Timestamps), countTrue(result.Availabilities))
	latencyScore := sdk.BigDec{}
	if len(result.Latencies) > 0 {
		/*
			Let's say result.Latencies has three values: 10ms, 20ms, 30ms. Thus, the total latency is 60ms and the average latency is 20ms.
			If the expected latency is 15ms, then the latency score would be 15/20 = 0.75
		*/
		totalLatency := sumDurations(result.Latencies)
		averageLatency := totalLatency / time.Duration(len(result.Latencies))
		latencyScore = sdk.MinDec(sdk.OneDec(), sdk.NewDecFromInt(sdk.NewInt(int64(expectedLatency))).Quo(sdk.NewDecFromInt(sdk.NewInt(int64(averageLatency)))))
	}

	report := &ViperQoSReport{
		FirstSampleTimestamp: firstSampleTimestamp,
		BlockHeight:          blockHeight,
		LatencyScore:         latencyScore,
		AvailabilityScore:    scaledAvailabilityScore,
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

// returns the expected latency to a threshold.
func CalculateExpectedLatency(timeoutGivenToRelay time.Duration) time.Duration {
	expectedLatency := (timeoutGivenToRelay / 2)
	return expectedLatency
}

func sumDurations(durations []time.Duration) time.Duration {
	sum := time.Duration(0)
	for _, d := range durations {
		sum += d
	}
	return sum
}
