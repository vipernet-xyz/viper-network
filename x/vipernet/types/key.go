package types

import (
	sdk "github.com/vipernet-xyz/viper-network/types"
)

const (
	ModuleName = "vipernet"              // name of the module
	StoreKey   = ModuleName              // key for state store
	TStoreKey  = "transient_" + StoreKey // transient key for state store
)

var (
	ClaimLen = len(ClaimKey)
	ClaimKey = []byte{0x02} // key for pending claims
)

// "KeyForClaim" - Generates the key for the claim object for the state store
func KeyForClaim(ctx sdk.Ctx, addr sdk.Address, header SessionHeader, evidenceType EvidenceType) ([]byte, error) {
	// validat the header
	if err := header.ValidateHeader(); err != nil {
		return nil, err
	}
	// validate the address
	if err := AddressVerification(addr.String()); err != nil {
		return nil, err
	}
	// validate the GOBEvidence type
	if evidenceType != RelayEvidence && evidenceType != ChallengeEvidence {
		return nil, NewInvalidEvidenceErr(ModuleName)
	}
	et, err := evidenceType.Byte()
	if err != nil {
		return nil, err
	}
	// return the key bz
	return append(append(append(ClaimKey, addr.Bytes()...), header.Hash()...), et), nil
}

// "KeyForClaims" - Generates the key for the claims object
func KeyForClaims(addr sdk.Address) ([]byte, error) {
	// verify the address
	if err := AddressVerification(addr.String()); err != nil {
		return nil, err
	}
	// return the key bz
	return append(ClaimKey, addr.Bytes()...), nil
}

// "KeyForEvidence" - Generates the key for GOBEvidence
func KeyForEvidence(header SessionHeader, evidenceType EvidenceType) ([]byte, error) {
	// validate the GOBEvidence type
	if evidenceType != RelayEvidence && evidenceType != ChallengeEvidence {
		return nil, NewInvalidEvidenceErr(ModuleName)
	}
	et, err := evidenceType.Byte()
	if err != nil {
		return nil, err
	}
	return append(header.Hash(), et), nil
}

// "KeyForTestEvidence" - Generates the key for GOBEvidence
func KeyForTestResult(header SessionHeader, evidenceType EvidenceType, servicerAddr sdk.Address) ([]byte, error) {
	if evidenceType != FishermanTestEvidence {
		return nil, NewInvalidEvidenceErr(ModuleName)
	}
	et, err := evidenceType.Byte()
	if err != nil {
		return nil, err
	}
	combined := append(header.Hash(), servicerAddr.Bytes()...)
	return append(combined, et), nil
}
