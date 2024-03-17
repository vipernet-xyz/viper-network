package types

import (
	"encoding/hex"
	"encoding/json"
	"errors"

	"golang.org/x/crypto/sha3"
)

var (
	// ErrNoSigner error when no signer is provided
	ErrNoSigner = errors.New("no signer provided")
	// ErrNoSession error when no session is provided
	ErrNoSession = errors.New("no session provided")
	// ErrNoSessionHeader error when no session header is provided
	ErrNoSessionHeader = errors.New("no session header provided")
	// ErrNoSender error when no sender is provided
	ErrNoSender = errors.New("no sender provided")
	// ErrNoViperAAT error when no Viper AAT is provided
	ErrNoViperAAT = errors.New("no Viper AAT provided")
	// ErrSessionHasNoNodes error when provided session has no nodes
	ErrSessionHasNoNodes = errors.New("session has no nodes")
	// ErrNodeNotInSession error when given node is not in session
	ErrNodeNotInSession = errors.New("node not in session")
)

// Relayer implementation of relayer interface
type Relayer struct {
	signer Signer
	sender Sender
}

// NewRelayer returns instance of Relayer with given input
func NewRelayer(signer Signer, sender Sender) *Relayer {
	return &Relayer{
		signer: signer,
		sender: sender,
	}
}

func (r *Relayer) validateRelayRequest(input *Input) error {
	if (r.signer == Signer{}) {
		return ErrNoSigner
	}

	if (r.sender == Sender{}) {
		return ErrNoSender
	}

	if input.Session == nil {
		return ErrNoSession
	}

	if input.ViperAAT == nil {
		return ErrNoViperAAT
	}

	if len(input.Session.SessionServicers) == 0 {
		return ErrSessionHasNoNodes
	}

	if len(input.Session.SessionFishermen) == 0 {
		return ErrSessionHasNoNodes
	}

	if (input.Session.SessionHeader == SessionHeader{}) {
		return ErrNoSessionHeader
	}

	return nil
}

func (r *Relayer) getSignedProofBytes(proof *RelayProof) (string, error) {
	proofBytes, err := GenerateProofBytes(proof)
	if err != nil {
		return "", err
	}

	return r.signer.Sign(proofBytes)
}

// GenerateProofBytes returns relay proof as encoded bytes
func GenerateProofBytes(proof *RelayProof) ([]byte, error) {
	token, err := HashAAT(&proof.Token)
	if err != nil {
		return nil, err
	}

	proofMap := &relayProofForSignature{
		RequestHash:        proof.RequestHash,
		Entropy:            proof.Entropy,
		SessionBlockHeight: proof.SessionBlockHeight,
		ServicerPubKey:     proof.ServicerPubKey,
		Blockchain:         proof.Blockchain,
		Token:              token,
		Signature:          "",
		GeoZone:            proof.GeoZone,
		NumServicers:       proof.NumServicers,
	}

	marshaledProof, err := json.Marshal(proofMap)
	if err != nil {
		return nil, err
	}

	hasher := sha3.New256()

	_, err = hasher.Write(marshaledProof)
	if err != nil {
		return nil, err
	}

	return hasher.Sum(nil), nil
}

// HashAAT returns Viper AAT as hashed string
func HashAAT(aat *AAT) (string, error) {
	tokenToSend := *aat
	tokenToSend.RequestorSignature = ""

	marshaledAAT, err := json.Marshal(tokenToSend)
	if err != nil {
		return "", err
	}

	hasher := sha3.New256()

	_, err = hasher.Write(marshaledAAT)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// HashRequest creates the request hash from its structure
func HashRequest(reqHash *RequestHash) (string, error) {
	marshaledReqHash, err := json.Marshal(reqHash)
	if err != nil {
		return "", err
	}

	hasher := sha3.New256()

	_, err = hasher.Write(marshaledReqHash)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
