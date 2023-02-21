package types

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
)

var (
	// A list of supported token versions
	// Requires major (semantic) upgrade to update this list
	SupportedTokenVersions = []string{"0.0.1"}
)

// "VersionIsIncluded" - Returns if the version is included
func (a AAT) VersionIsIncluded() bool {
	// if version is empty return nil
	return !(a.Version == "")
}

// "VersionIsSupported" - Returns if the version of the AAT is supported by the network
func (a AAT) VersionIsSupported() bool {
	for _, v := range SupportedTokenVersions {
		if a.Version == v {
			return true
		}
	}
	return false
}

// "Validate" - Returns an error for an invalid AAT
func (a AAT) Validate() error {
	// check the version of the aat
	if err := a.ValidateVersion(); err != nil {
		return err
	}
	// check the message of the aat
	if err := a.ValidateMessage(); err != nil {
		return err
	}
	// check the provider signature of the aat
	if err := a.ValidateSignature(); err != nil {
		return err
	}
	return nil
}

// "Bytes" - Returns the bytes representation of the AAT
func (a AAT) Bytes() []byte {
	// using standard json bz
	b, err := json.Marshal(AAT{
		ProviderSignature: "",
		ProviderPublicKey: a.ProviderPublicKey,
		ClientPublicKey:   a.ClientPublicKey,
		Version:           a.Version,
	})
	if err != nil {
		log.Fatal(fmt.Sprintf("an error occured hashing the aat:\n%v", err))
	}
	return b
}

// "ID" - Returns the merkleHash of the AAT bytes
func (a AAT) Hash() []byte {
	return Hash(a.Bytes())
}

// "HashString" - Returns the string representation of the AAT merkleHash
func (a AAT) HashString() string {
	// using standard library hex
	return hex.EncodeToString(a.Hash())
}

// "ValidateVersion" - Confirms the version field of the AAT
func (a AAT) ValidateVersion() error {
	// check for valid version
	if !a.VersionIsIncluded() {
		return MissingTokenVersionError
	}
	// check if version is supported
	if !a.VersionIsSupported() {
		return UnsupportedTokenVersionError
	}
	return nil
}

// "ValidateMessage" - Confirms the message field of the AAT
func (a AAT) ValidateMessage() error {
	// check for valid provider public key
	if len(a.ProviderPublicKey) == 0 {
		return MissingProviderPublicKeyError
	}
	if err := PubKeyVerification(a.ProviderPublicKey); err != nil {
		return err
	}
	// check if client public key is valid
	if len(a.ClientPublicKey) == 0 {
		return MissingClientPublicKeyError
	}
	if err := PubKeyVerification(a.ClientPublicKey); err != nil {
		return err
	}
	return nil
}

// "ValidateSignature" - Confirms the signature field of the AAT
func (a AAT) ValidateSignature() error {
	// check for valid signature
	messageHash := a.HashString()
	// verifies the signature with the message of the AAT
	if err := SignatureVerification(a.ProviderPublicKey, messageHash, a.ProviderSignature); err != nil {
		return InvalidTokenSignatureErorr
	}
	return nil
}
