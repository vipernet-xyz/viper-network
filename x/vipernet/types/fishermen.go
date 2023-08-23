package types

import (
	rand1 "crypto/rand"
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
	relayer := NewRelayer(nil, *sender)
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

// "Store" - Handles the test result object by adding it to the cache
func (tr TestResult) Store(sessionHeader SessionHeader, testStore *CacheStorage) {
	// add the result to the global (in memory) collection of results
	SetTestResult(sessionHeader, FishermanTestEvidence, tr, testStore)
}

func SetTestResult(header SessionHeader, evidenceType EvidenceType, tr TestResult, testStore *CacheStorage) {
	test, err := GetTestResult(header, evidenceType, testStore)
	if err != nil {
		log.Fatalf("could not set test result object: %s", err.Error())
	}
	test.AddTestResult(tr)
	SetResult(test, testStore)
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
