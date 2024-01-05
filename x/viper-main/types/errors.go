package types

import (
	"errors"
	"fmt"
	"strconv"

	sdk "github.com/vipernet-xyz/viper-network/types"
)

const (
	CodeSessionGenerationError              = 1
	CodeHttpStatusCodeError                 = 2
	CodeInvalidTokenError                   = 4
	CodeInvalidEvidenceError                = 5
	CodePublKeyDecodeError                  = 6
	CodeEmptyChainError                     = 8
	CodeEmptyBlockIDError                   = 9
	CodeRequestorPubKeyError                = 10
	CodeEmptyProofsError                    = 11
	CodeUnsupportedBlockchainRequestorError = 13
	CodeInvalidSessionError                 = 14
	CodeInsufficientServicersError          = 17
	CodeEmptyNonNativeChainError            = 18
	CodeInvalidSessionKeyError              = 19
	CodeFilterServicersError                = 20
	CodeXORError                            = 21
	CodeInvalidHashError                    = 22
	CodeEmptyBlockHashError                 = 23
	CodeEmptyBlockchainError                = 24
	CodeEmptyPayloadDataError               = 25
	CodeUnsupportedBlockchainNodeError      = 26
	CodeNotStakedBlockchainError            = 27
	CodeHTTPExecutionError                  = 28
	CodeInvalidEntropyError                 = 29
	CodeEmptyResponseError                  = 30
	CodeResponseSignatureError              = 31
	CodeNegativeICCounterError              = 32
	CodeMaximumEntropyError                 = 33
	CodeInvalidNodePubKeyError              = 34
	CodeTicketsNotFoundError                = 35
	CodeDuplicateTicketError                = 36
	CodeDuplicateProofError                 = 37
	CodeInvalidSignatureSizeError           = 38
	CodeSigDecodeError                      = 39
	CodeMsgDecodeError                      = 40
	CodeInvalidSigError                     = 41
	CodePubKeySizeError                     = 42
	CodeEmptyKeybaseError                   = 43
	CodeSelfNotFoundError                   = 44
	CodeRequestorNotFoundError              = 45
	CodeChainNotHostedError                 = 46
	CodeInvalidHostedChainsError            = 47
	CodeNodeNotFoundError                   = 48
	CodeInvalidProofsError                  = 49
	CodeInconsistentPubKeyError             = 50
	CodeInvalidChainParamsError             = 51
	CodeNewHexDecodeError                   = 52
	CodeChainNotSupportedErr                = 53
	CodePubKeyError                         = 54
	CodeSignatureError                      = 55
	CodeInvalidChainError                   = 56
	CodeJSONMarshalError                    = 57
	CodeInvalidBlockchainHashLengthError    = 58
	CodeEmptySessionKeyError                = 59
	CodeInvalidBlockHeightError             = 60
	CodeInvalidRequestorPubKeyError         = 61
	CodeInvalidHashLengthError              = 62
	CodeInvalidLeafCousinProofsCombo        = 63
	CodeEmptyAddressError                   = 64
	CodeClaimNotFoundError                  = 65
	CodeInvalidMerkleVerifyError            = 66
	CodeEmptyMerkleTreeError                = 67
	CodeMerkleNodeNotFoundError             = 68
	CodeExpiredProofsSubmissionError        = 69
	CodeAddressError                        = 70
	CodeOverServiceError                    = 71
	CodeCousinLeafEquivalentError           = 72
	CodeInvalidRootError                    = 73
	CodeRequestHash                         = 74
	CodeOutOfSyncRequestError               = 75
	CodeUnsupportedBlockchainError          = 76
	CodeDuplicatePublicKeyError             = 77
	CodeMismatchedRequestHashError          = 78
	CodeNewMismatchedRequestorPubKeyError   = 79
	CodeMismatchedSessionHeightError        = 80
	CodeMismatchedBlockchainsError          = 81
	CodeNoMajorityResponseError             = 82
	CodeNodeNotInSessionError               = 83
	CodeNoEvidenceTypeErr                   = 84
	CodeInvalidPkFileErr                    = 85
	CodeReplayAttackError                   = 86
	CodeInvalidNetworkIDError               = 87
	CodeInvalidExpirationHeightErr          = 88
	CodeInvalidMerkleRangeError             = 89
	CodeEvidenceSealed                      = 90
	CodeGeoZoneNotHostedError               = 91
	CodeInvalidGeoZoneError                 = 92
	CodeSampleNotHostedError                = 93
	CodeInvalidSampleError                  = 94
	CodeEmptyGeoZoneError                   = 95
	CodeUnsupportedGeoZoneRequestorError    = 96
	CodeInvalidTestsError                   = 97
	CodeExpiredReportSubmissionError        = 98
	CodeGeoZoneNotSupportedErr              = 99
	CodeReportCardNotFoundError             = 100
	CodeInvalidRCMerkleVerifyError          = 101
)

var (
	MissingTokenVersionError            = errors.New("the requestor authentication token version is missing")
	UnsupportedTokenVersionError        = errors.New("the requestor authentication token version is not supported")
	MissingRequestorPublicKeyError      = errors.New("the requestorlicaiton public key included in the AAT is not valid")
	MissingClientPublicKeyError         = errors.New("the client public key included in the AAT is not valid")
	InvalidTokenSignatureErorr          = errors.New("the requestor signature on the AAT is not valid")
	NegativeICCounterError              = errors.New("the IC counter is less than 0")
	MaximumEntropyError                 = errors.New("the entropy exceeds the maximum allowed relays")
	NodeNotInSessionError               = errors.New("the servicer is not within the session")
	InvalidNodePubKeyError              = errors.New("the servicer public key in the service Proof does not match this servicers public key")
	InvalidTokenError                   = errors.New("the requestor authentication token is invalid")
	EmptyProofsError                    = errors.New("the service proofs object is empty")
	DuplicateProofError                 = errors.New("the Proof with specific merkleHash already found, check entropy")
	InvalidEntropyError                 = errors.New("the entropy included in the relay request is invalid")
	EmptyResponseError                  = errors.New("the relay response payload is empty")
	ResponseSignatureError              = errors.New("response signing errored out: ")
	EmptyBlockchainError                = errors.New("the blockchain included in the relay request is empty")
	EmptyPayloadDataError               = errors.New("the payload data of the relay request is empty")
	UnsupportedBlockchainError          = errors.New("the blockchain in this request is not supported")
	UnsupportedBlockchainRequestorError = errors.New("the blockchain in the relay request is not supported for this requestor")
	UnsupportedGeoZoneRequestorError    = errors.New("the geo-zone in the relay request is not supported for this requestor")
	UnsupportedBlockchainNodeError      = errors.New("the blockchain in the relay request is not supported on this servicer")
	HttpStatusCodeError                 = errors.New("HTTP status code returned not okay: ")
	InvalidSessionError                 = errors.New("this servicer (self) is not responsible for this session provided by the client")
	ServiceSessionGenerationError       = errors.New("unable to generate a session for the seed data: ")
	NotStakedBlockchainError            = errors.New("the blockchain is not staked for this requestor")
	EmptyRequestorPubKeyError           = errors.New("the public key of the requestor is of Length 0")
	EmptyNonNativeChainError            = errors.New("the non-native chain is of Length 0")
	EmptyBlockIDError                   = errors.New("the block addr is of Length 0")
	InsufficientServicersError          = errors.New("there are less than the minimum session servicers found")
	EmptySessionKeyError                = errors.New("the session key passed is of Length 0")
	MismatchedByteArraysError           = errors.New("the byte arrays are not of the same Length")
	FilterServicersError                = errors.New("unable to filter servicers: ")
	XORError                            = errors.New("error XORing the keys: ")
	PubKeyDecodeError                   = errors.New("error decoding the string into hex bytes")
	InvalidHashError                    = errors.New("the hash is invalid: ")
	HTTPExecutionError                  = errors.New("error executing the http request: ")
	TicketsNotFoundError                = errors.New("the tickets requested could not be found")
	DuplicateTicketError                = errors.New("the ticket is a duplicate")
	InvalidSignatureSizeError           = errors.New("the signature Length is invalid")
	MessageDecodeError                  = errors.New("the message could not be hex decoded")
	SigDecodeError                      = errors.New("the signature could not be message decoded")
	InvalidSignatureError               = errors.New("the signature could not be verified with the message and pub key")
	PubKeySizeError                     = errors.New("the public key is not the correct cap")
	KeybaseError                        = errors.New("the keybase is invalid: ")
	SelfNotFoundError                   = errors.New("the self servicer is not within the world state")
	RequestorNotFoundError              = errors.New("the requestor could not be found in the world state")
	RequestHashError                    = errors.New("the request hash does not match the payload hash")
	InvalidHostedChainError             = errors.New("invalid hosted chain error")
	InvalidGeoZoneError                 = errors.New("invalid geozone error")
	InvalidSampleError                  = errors.New("invalid sample error")
	ChainNotHostedError                 = errors.New("the blockchain requested is not hosted")
	GeoZoneNotHostedError               = errors.New("the GeoZone requested is not hosted")
	SampleNotHostedError                = errors.New("the Sample requested is not hosted")
	NodeNotFoundErr                     = errors.New("the servicer is not found in world state")
	InvalidProofsError                  = errors.New("the proofs provided are invalid or less than the minimum requirement")
	InconsistentPubKeyError             = errors.New("the public keys in the proofs are inconsistent")
	InvalidChainParamsError             = errors.New("the required params for a nonNative blockchain are invalid")
	HexDecodeError                      = errors.New("the hex string could not be decoded: ")
	ChainNotSupportedErr                = errors.New("the chain is not viper supported")
	PubKeyError                         = errors.New("could not convert hex string to pub key: ")
	SignatureError                      = errors.New("there was a problem signing the message: ")
	InvalidChainError                   = errors.New("the non native chain passed was invalid: ")
	JSONMarshalError                    = errors.New("unable to marshal object into json: ")
	InvalidNetworkIDLengthError         = errors.New("the netid Length is invalid")
	InvalidBlockHeightError             = errors.New("the block height passed is invalid")
	InvalidRequestorPubKeyError         = errors.New("the requestor public key is invalid")
	InvalidHashLengthError              = errors.New("the merkleHash Length is not valid")
	InvalidLeafCousinProofsCombo        = errors.New("the merkle relayProof combo for the cousin and leaf is invalid")
	EmptyAddressError                   = errors.New("the address provided is empty")
	ClaimNotFoundError                  = errors.New("the claim was not found for the key given")
	InvalidMerkleVerifyError            = errors.New("claim resulted in an invalid merkle Proof")
	EmptyMerkleTreeError                = errors.New("the merkle tree is empty")
	NodeNotFoundError                   = errors.New("the servicer of the merkle tree requested is not found")
	ExpiredProofsSubmissionError        = errors.New("the opportunity of window to submit the Proof has closed because the secret has been revealed")
	AddressError                        = errors.New("the address is invalid")
	OverServiceError                    = errors.New("the max number of relays serviced for this servicer is exceeded")
	UninitializedKeybaseError           = errors.New("the keybase is nil")
	CousinLeafEquivalentError           = errors.New("the cousin and leaf cannot be equal")
	InvalidRootError                    = errors.New("the merkle root passed is invalid")
	MerkleNodeNotFoundError             = errors.New("the merkle servicer cannot be found")
	OutOfSyncRequestError               = errors.New("the request block height is out of sync with the current block height")
	DuplicatePublicKeyError             = errors.New("the public key is duplicated in the proof")
	MismatchedRequestHashError          = errors.New("the request hashes included in the proof do not match")
	MismatchedRequestorPubKeyError      = errors.New("the requestor public keys included in the proofs do not match")
	MismatchedSessionHeightError        = errors.New("the session block heights included in the proofs do not match")
	MismatchedBlockchainsError          = errors.New("the non-native blockchains provided in the proofs do not match")
	NoMajorityResponseError             = errors.New("no majority can be established between all of the responses")
	NoEvidenceTypeErr                   = errors.New("the GOBEvidence type is not supplied in the claim message")
	InvalidPkFileErr                    = errors.New("the PK File is not found")
	InvalidEvidenceErr                  = errors.New("the GOBEvidence type passed is not valid")
	ReplayAttackError                   = errors.New("the merkle proof is flagged as a replay attack")
	InvalidExpirationHeightErr          = errors.New("the expiration height included in the claim message is invalid (should not be set)")
	InvalidMerkleRangeError             = errors.New("the merkle hash range is invalid")
	SealedEvidenceError                 = errors.New("the evidence is sealed, either max relays reached or claim already submitted")
	InvalidNumServicersError            = errors.New("the NumServicers included in the relay request is invalid")
	ZeroTimeError                       = errors.New("Timestamp cannot be zero")
	NegativeLatency                     = errors.New("Latency should be a positive duration")
	EmptyGeoZoneError                   = errors.New("the geozone is of Length 0")
	InvalidTestsError                   = errors.New("the tests provided are invalid or less than the minimum requirement")
	ExpiredReportSubmissionError        = errors.New("the opportunity of window to submit the Report has closed because the secret has been revealed")
	GeoZoneNotSupportedErr              = errors.New("the geozone is not viper supported")
	ReportCardNotFoundError             = errors.New("the report card for the servicer could not be found")
	InvalidRCMerkleVerifyError          = errors.New("report card resulted in an invalid merkle Proof")
)

func NewSealedEvidenceError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeEvidenceSealed, SealedEvidenceError.Error())
}

func NewUnsupportedBlockchainError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeUnsupportedBlockchainError, UnsupportedBlockchainError.Error())
}
func NewNodeNotInSessionError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeNodeNotInSessionError, NodeNotInSessionError.Error())
}

func NewOverServiceError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeOverServiceError, OverServiceError.Error())
}

func NewAddressInvalidLengthError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeAddressError, AddressError.Error())
}

func NewExpiredProofsSubmissionError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeExpiredProofsSubmissionError, ExpiredProofsSubmissionError.Error())
}

func NewExpiredReportSubmissionError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeExpiredReportSubmissionError, ExpiredReportSubmissionError.Error())
}

func NewMerkleNodeNotFoundError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeMerkleNodeNotFoundError, MerkleNodeNotFoundError.Error())
}

func NewEmptyMerkleTreeError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeEmptyMerkleTreeError, EmptyMerkleTreeError.Error())
}

func NewInvalidMerkleVerifyError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidMerkleVerifyError, InvalidMerkleVerifyError.Error())
}

func NewReplayAttackError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeReplayAttackError, ReplayAttackError.Error())
}

func NewClaimNotFoundError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeClaimNotFoundError, ClaimNotFoundError.Error())
}

func NewEmptyAddressError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeEmptyAddressError, EmptyAddressError.Error())
}

func NewZeroTimeError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeEmptyAddressError, ZeroTimeError.Error())
}

func NewNegativeLatency(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeEmptyAddressError, NegativeLatency.Error())
}

func NewCousinLeafEquivalentError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeCousinLeafEquivalentError, CousinLeafEquivalentError.Error())
}

func NewInvalidLeafCousinProofsComboError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidLeafCousinProofsCombo, InvalidLeafCousinProofsCombo.Error())
}

func NewInvalidRootError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidRootError, InvalidRootError.Error())
}

func NewInvalidHashLengthError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidHashLengthError, InvalidHashLengthError.Error())
}
func NewInvalidNetIDLengthError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidNetworkIDError, InvalidNetworkIDLengthError.Error())
}
func NewInvalidRequestorPubKeyError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidRequestorPubKeyError, InvalidRequestorPubKeyError.Error())
}

func NewInvalidBlockHeightError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidBlockHeightError, InvalidBlockHeightError.Error())
}

func NewJSONMarshalError(codespace sdk.CodespaceType, err error) sdk.Error {
	return sdk.NewError(codespace, CodeJSONMarshalError, JSONMarshalError.Error()+err.Error())
}

func NewSignatureError(codespace sdk.CodespaceType, err error) sdk.Error {
	return sdk.NewError(codespace, CodeSignatureError, SignatureError.Error()+err.Error())
}

func NewPubKeyError(codespace sdk.CodespaceType, err error) sdk.Error {
	return sdk.NewError(codespace, CodePubKeyError, PubKeyError.Error()+err.Error())
}

func NewMismatchedRequestHashError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeMismatchedRequestHashError, MismatchedRequestHashError.Error())
}

func NewMismatchedRequestorPubKeyError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeNewMismatchedRequestorPubKeyError, MismatchedRequestorPubKeyError.Error())
}

func NewMismatchedSessionHeightError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeMismatchedSessionHeightError, MismatchedSessionHeightError.Error())
}

func NewMismatchedBlockchainsError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeMismatchedBlockchainsError, MismatchedBlockchainsError.Error())
}

func NewNoMajorityResponseError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeNoMajorityResponseError, NoMajorityResponseError.Error())
}

func NewDuplicatePublicKeyError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeDuplicatePublicKeyError, DuplicatePublicKeyError.Error())
}

func NewChainNotSupportedErr(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeChainNotSupportedErr, ChainNotSupportedErr.Error())
}

func NewGeoZoneNotSupportedErr(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeGeoZoneNotSupportedErr, GeoZoneNotSupportedErr.Error())
}

func NewNoEvidenceTypeErr(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeNoEvidenceTypeErr, NoEvidenceTypeErr.Error())
}

func NewInvalidEvidenceErr(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidEvidenceError, InvalidEvidenceErr.Error())
}

func NewInvalidMerkleRangeError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidMerkleRangeError, InvalidMerkleRangeError.Error())
}

func NewInvalidExpirationHeightErr(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidExpirationHeightErr, InvalidExpirationHeightErr.Error())
}

func NewHexDecodeError(codespace sdk.CodespaceType, err error) sdk.Error {
	return sdk.NewError(codespace, CodeNewHexDecodeError, HexDecodeError.Error()+err.Error())
}

func NewInvalidChainParamsError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidChainParamsError, InvalidChainParamsError.Error())
}

func NewInvalidProofsError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidProofsError, InvalidProofsError.Error())
}

func NewInvalidTestsError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidTestsError, InvalidTestsError.Error())
}

func NewNodeNotFoundErr(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeNodeNotFoundError, NodeNotFoundErr.Error())
}

func NewInvalidHostedChainError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidHostedChainsError, InvalidHostedChainError.Error())
}

func NewInvalidGeoZoneError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidGeoZoneError, InvalidHostedChainError.Error())
}

func NewErrorChainNotHostedError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeChainNotHostedError, ChainNotHostedError.Error())
}

func NewErrorGeoZoneNotHostedError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeGeoZoneNotHostedError, GeoZoneNotHostedError.Error())
}

func NewRequestorNotFoundError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeRequestorNotFoundError, RequestorNotFoundError.Error())
}

func NewSelfNotFoundError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeSelfNotFoundError, SelfNotFoundError.Error())
}

func NewKeybaseError(codespace sdk.CodespaceType, err error) sdk.Error {
	return sdk.NewError(codespace, CodeEmptyKeybaseError, KeybaseError.Error()+err.Error())
}

func NewPubKeySizeError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodePubKeySizeError, PubKeySizeError.Error())
}

func NewInvalidSignatureError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidSigError, InvalidSignatureError.Error())
}

func NewMsgDecodeError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeMsgDecodeError, MessageDecodeError.Error())
}

func NewSigDecodeError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeSigDecodeError, SigDecodeError.Error())
}

func NewInvalidSignatureSizeError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidSignatureSizeError, InvalidSignatureSizeError.Error())
}

func NewRequestHashError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeRequestHash, RequestHashError.Error())
}

func NewDuplicateProofError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeDuplicateProofError, DuplicateProofError.Error())
}

func NewInvalidNodePubKeyError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidNodePubKeyError, InvalidNodePubKeyError.Error())
}

func NewResponseSignatureError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeResponseSignatureError, ResponseSignatureError.Error())
}

func NewEmptyResponseError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeEmptyResponseError, EmptyResponseError.Error())
}

func NewOutOfSyncRequestError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeOutOfSyncRequestError, OutOfSyncRequestError.Error())
}

func NewInvalidEntropyError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidEntropyError, InvalidEntropyError.Error())
}

func NewHTTPExecutionError(codespace sdk.CodespaceType, err error) sdk.Error {
	return sdk.NewError(codespace, CodeHTTPExecutionError, HTTPExecutionError.Error()+err.Error())
}

func NewUnsupportedBlockchainNodeError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeUnsupportedBlockchainNodeError, UnsupportedBlockchainNodeError.Error())
}

func NewEmptyPayloadDataError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeEmptyPayloadDataError, EmptyPayloadDataError.Error())
}

func NewInvalidHashError(codespace sdk.CodespaceType, err error, h string) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidHashError, fmt.Sprintf("%s %s%s", h, InvalidHashError.Error(), err.Error()))
}

func NewEmptyHashError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeEmptyBlockHashError, InvalidHashError.Error())
}

func NewPubKeyDecodeError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodePublKeyDecodeError, PubKeyDecodeError.Error())
}

func NewXORError(codespace sdk.CodespaceType, err error) sdk.Error {
	return sdk.NewError(codespace, CodeXORError, XORError.Error()+err.Error())
}

func NewFilterServicersError(codespace sdk.CodespaceType, err error) sdk.Error {
	return sdk.NewError(codespace, CodeFilterServicersError, FilterServicersError.Error()+err.Error())
}

func NewInvalidSessionKeyError(codespace sdk.CodespaceType, err error) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidSessionKeyError, InvalidSessionError.Error()+err.Error())
}

func NewEmptyNonNativeChainError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeEmptyNonNativeChainError, EmptyNonNativeChainError.Error())
}

func NewEmptyGeoZoneError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeEmptyGeoZoneError, EmptyGeoZoneError.Error())
}

func NewInsufficientServicersError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInsufficientServicersError, InsufficientServicersError.Error())
}

func NewInvalidSessionError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidSessionError, InvalidSessionError.Error())
}

func NewUnsupportedBlockchainRequestorError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeUnsupportedBlockchainRequestorError, UnsupportedBlockchainRequestorError.Error())
}

func NewUnsupportedGeoZoneRequestorError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeUnsupportedGeoZoneRequestorError, UnsupportedGeoZoneRequestorError.Error())
}

func NewEmptyProofsError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeEmptyProofsError, EmptyProofsError.Error())
}

func NewEmptyBlockIDError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeEmptyBlockIDError, EmptyBlockIDError.Error())
}
func NewEmptyChainError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeEmptyChainError, EmptyNonNativeChainError.Error())
}

func NewHTTPStatusCodeError(codespace sdk.CodespaceType, statusCode int) sdk.Error {
	return sdk.NewError(codespace, CodeHttpStatusCodeError, HttpStatusCodeError.Error()+strconv.Itoa(statusCode))
}

func NewInvalidTokenError(codespace sdk.CodespaceType, err error) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidTokenError, InvalidTokenError.Error()+" : "+err.Error())
}

func NewInvalidPKError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidPkFileErr, InvalidPkFileErr.Error())
}

func NewInvalidNumServicersError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidEntropyError, InvalidEntropyError.Error())
}

func NewInvalidSampleError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidSampleError, InvalidSampleError.Error())
}

func NewSampleNotHostedError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeSampleNotHostedError, SampleNotHostedError.Error())
}

func NewReportCardNotFoundError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeReportCardNotFoundError, ReportCardNotFoundError.Error())
}

func NewInvalidRCMerkleVerifyError(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeInvalidRCMerkleVerifyError, InvalidRCMerkleVerifyError.Error())
}
