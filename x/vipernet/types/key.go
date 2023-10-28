package types

import (
	sdk "github.com/vipernet-xyz/viper-network/types"
)

const (
	ModuleName  = "vipernet"              // name of the module
	StoreKey    = ModuleName              // key for state store
	TStoreKey   = "transient_" + StoreKey // transient key for state store
	MemStoreKey = "memory_" + ModuleName
)

var (
	ClaimLen      = len(ClaimKey)
	ClaimKey      = []byte{0x02} // key for pending claims
	ReportCardKey = []byte{0x03}
)

// "KeyForClaim" - Generates the key for the claim object for the state store
func KeyForClaim(ctx sdk.Ctx, addr sdk.Address, header SessionHeader, evidenceType EvidenceType) ([]byte, error) {
	// validate the header
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

// "KeyForReportCard" - Generates the key for the ViperQoSReport object for the state store
func KeyForReportCard(ctx sdk.Ctx, servicerAddress sdk.Address, fishermanAddress sdk.Address, header SessionHeader) ([]byte, error) {
	// Validate the servicer's address
	if err := AddressVerification(servicerAddress.String()); err != nil {
		return nil, err
	}

	// Validate the fisherman's address
	if err := AddressVerification(fishermanAddress.String()); err != nil {
		return nil, err
	}

	// Validate the header
	if err := header.ValidateHeader(); err != nil {
		return nil, err
	}

	// Construct the key by appending servicer's address, fisherman's address, and header's hash.
	return append(append(append(ReportCardKey, servicerAddress.Bytes()...), fishermanAddress.Bytes()...), header.Hash()...), nil
}

func KeyForReportCards(servicerAddress sdk.Address, fishermanAddress sdk.Address) ([]byte, error) {
	// verify the address
	if err := AddressVerification(servicerAddress.String()); err != nil {
		return nil, err
	}
	if err := AddressVerification(fishermanAddress.String()); err != nil {
		return nil, err
	}
	// return the key bz
	return append(append(ReportCardKey, servicerAddress.Bytes()...), fishermanAddress.Bytes()...), nil
}
