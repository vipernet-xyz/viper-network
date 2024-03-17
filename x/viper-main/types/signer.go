package types

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"golang.org/x/crypto/scrypt"
	"golang.org/x/crypto/sha3"
)

const (
	scryptHashLength = 32
	scryptN          = 32768
	scryptR          = 8
	scryptP          = 1
	scryptSec        = 12
	tagLength        = 16
	defaultKDF       = "scrypt"
)

var (
	// ErrInvalidPrivateKey error when private key is invalid
	ErrInvalidPrivateKey = errors.New("invalid private key")
	// ErrInvalidPPK error when PPK is invalid
	ErrInvalidPPK = errors.New("invalid ppk")

	base64Regex = regexp.MustCompile("^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?$")
	hexRegex    = regexp.MustCompile("^[a-fA-F0-9]+$")
)

// Signer struct handler
type Signer struct {
	address    string
	publicKey  string
	privateKey string
}

func getAESGCMValues(password, saltBytes []byte) ([]byte, cipher.AEAD, error) {
	scryptKey, err := scrypt.Key(password, saltBytes, scryptN, scryptR, scryptP, scryptHashLength)
	if err != nil {
		return nil, nil, err
	}

	block, err := aes.NewCipher(scryptKey)
	if err != nil {
		return nil, nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	return scryptKey[:scryptSec], gcm, nil
}

// Sign returns a signed request as encoded hex string
func (s *Signer) Sign(payload []byte) (string, error) {

	decodedKey, err := hex.DecodeString(s.privateKey)
	if err != nil {
		return "", fmt.Errorf("error decoding private key: %v", err)
	}

	if len(decodedKey) != ed25519.PrivateKeySize {
		return "", fmt.Errorf("invalid private key length")
	}

	signature := ed25519.Sign(decodedKey, payload)
	if len(signature) != ed25519.SignatureSize {
		return "", fmt.Errorf("invalid signature length")
	}

	return hex.EncodeToString(signature), nil
}

// SignBytes returns a signed request as raw bytes
func (s *Signer) SignBytes(payload []byte) ([]byte, error) {
	decodedKey, err := hex.DecodeString(s.privateKey)
	if err != nil {
		return nil, err
	}

	return ed25519.Sign(decodedKey, payload), nil
}

// GetAddress returns address value
func (s *Signer) GetAddress() string {
	return s.address
}

// GetPublicKey returns public key value
func (s *Signer) GetPublicKey() string {
	return s.publicKey
}

// GetPrivateKey returns private key value
func (s *Signer) GetPrivateKey() string {
	return s.privateKey
}

// NewRandomSigner returns a Signer with random keys
func NewRandomSigner() (*Signer, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, err
	}

	address, err := GetAddressFromDecodedPublickey(publicKey)
	if err != nil {
		return nil, err
	}

	return &Signer{
		address:    address,
		publicKey:  hex.EncodeToString(publicKey),
		privateKey: hex.EncodeToString(privateKey),
	}, nil
}

// Account holds an account's data
type Account struct {
	Address    string
	PublicKey  string
	PrivateKey string
}

// GetAccount returns Account struct holding all key values
func (s *Signer) GetAccount() *Account {
	return &Account{
		Address:    s.address,
		PublicKey:  s.publicKey,
		PrivateKey: s.privateKey,
	}
}

// PPK struct handler for Portable Private Key file
// Do not change this struct to stay compatible
type PPK struct {
	Kdf        string `json:"kdf"`
	Salt       string `json:"salt"`
	SecParam   string `json:"secparam"`
	Hint       string `json:"hint"`
	Ciphertext string `json:"ciphertext"`
}

func randBytes(numBytes int) ([]byte, error) {
	b := make([]byte, numBytes)

	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// NewPPK returns an instance of PPK for given input
func NewPPK(privateKey, password, hint string) (*PPK, error) {
	saltBytes, err := randBytes(tagLength)
	if err != nil {
		return nil, err
	}

	nuance, gcm, err := getAESGCMValues([]byte(password), saltBytes)
	if err != nil {
		return nil, err
	}

	return &PPK{
		Kdf:        defaultKDF,
		Salt:       hex.EncodeToString(saltBytes),
		SecParam:   strconv.Itoa(scryptSec),
		Hint:       hint,
		Ciphertext: base64.StdEncoding.EncodeToString(gcm.Seal(nil, nuance, []byte(privateKey), nil)),
	}, nil
}

// Validate returns bool representing if PPK instance is valid or not
func (ppk *PPK) Validate() bool {
	secParam, err := strconv.Atoi(ppk.SecParam)
	if err != nil {
		return false
	}

	return ppk.Kdf == defaultKDF &&
		ppk.Salt != "" &&
		hexRegex.MatchString(ppk.Salt) &&
		secParam >= 0 &&
		ppk.Ciphertext != "" &&
		base64Regex.MatchString(ppk.Ciphertext)
}

func NewSignerFromPrivateKey(privateKey string) (*Signer, error) {
	if !ValidatePrivateKey(privateKey) {
		return nil, ErrInvalidPrivateKey
	}

	publicKey := PublicKeyFromPrivate(privateKey)

	address, err := GetAddressFromPublickey(publicKey)

	if err != nil {
		return nil, err
	}

	return &Signer{
		address:    address,
		publicKey:  publicKey,
		privateKey: privateKey,
	}, nil
}

func GetAddressFromDecodedPublickey(decodedKey []byte) (string, error) {
	hasher := sha256.New()

	_, err := hasher.Write(decodedKey)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil))[0:40], nil
}

const (
	privateKeyLength = 128
)

// ValidatePrivateKey returns bool identifying if private key is valid or not
func ValidatePrivateKey(privateKey string) bool {
	return len(privateKey) == privateKeyLength && hexRegex.MatchString(privateKey)
}

func PublicKeyFromPrivate(privateKey string) string {
	return privateKey[64:]
}

// GetAddressFromPublickey converts an Requestor's Public key into an address
// publicKey parameter is the requestor's public key
// returns the requestor's address
func GetAddressFromPublickey(publicKey string) (string, error) {
	decodedKey, err := hex.DecodeString(publicKey)
	if err != nil {
		return "", err
	}

	hasher := sha256.New()

	_, err = hasher.Write(decodedKey)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil))[0:40], nil
}

type reportForSignature struct {
	FirstSampleTimestamp time.Time   `json:"first_sample_timestamp"`
	ServicerAddress      sdk.Address `json:"servicer_addr"`
	LatencyScore         sdk.BigDec  `json:"latency_score"`
	AvailabilityScore    sdk.BigDec  `json:"availability_score"`
	ReliabilityScore     sdk.BigDec  `json:"reliability_score"`
	SampleRoot           HashRange   `json:"sample_root"`
	Nonce                int64       `json:"nonce"`
	Signature            string      `json:"signature"`
}

func (s Signer) GetSignedReportBytes(report *ViperQoSReport) (string, error) {

	reportBytes, err := GenerateReportBytes(report)
	if err != nil {
		return "", err
	}

	return s.Sign(reportBytes)
}

// GenerateProofBytes returns relay proof as encoded bytes
func GenerateReportBytes(report *ViperQoSReport) ([]byte, error) {

	reportMap := &reportForSignature{
		FirstSampleTimestamp: report.FirstSampleTimestamp,
		ServicerAddress:      report.ServicerAddress,
		LatencyScore:         report.LatencyScore,
		AvailabilityScore:    report.AvailabilityScore,
		ReliabilityScore:     report.ReliabilityScore,
		SampleRoot:           report.SampleRoot,
		Nonce:                report.Nonce,
		Signature:            "",
	}
	marshaledReport, err := json.Marshal(reportMap)
	if err != nil {
		return nil, err
	}

	hasher := sha3.New256()

	_, err = hasher.Write(marshaledReport)
	if err != nil {
		return nil, err
	}

	return hasher.Sum(nil), nil
}
