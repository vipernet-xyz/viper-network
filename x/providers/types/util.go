package types

import (
	"encoding/hex"
	"fmt"

	sdk "github.com/vipernet-xyz/viper-network/types"
)

// TODO shared code among modules below

const (
	NetworkIdentifierLength = 2
	GeoZoneIdentifierLength = 2
)

func ValidateNetworkIdentifier(chain string) sdk.Error {
	// decode string into bz
	h, err := hex.DecodeString(chain)
	if err != nil {
		return ErrInvalidNetworkIdentifier(ModuleName, err)
	}
	// ensure length isn't 0
	if len(h) == 0 {
		return ErrInvalidNetworkIdentifier(ModuleName, fmt.Errorf("net id is empty"))
	}
	// ensure length
	if len(h) > NetworkIdentifierLength {
		return ErrInvalidNetworkIdentifier(ModuleName, fmt.Errorf("net id length is > %d", NetworkIdentifierLength))
	}
	return nil
}

func ValidateGeoZoneIdentifier(geoZone string) sdk.Error {
	// decode string into bz
	h, err := hex.DecodeString(geoZone)
	if err != nil {
		return ErrInvalidGeoZoneIdentifier(ModuleName, err)
	}
	// ensure length isn't 0
	if len(h) == 0 {
		return ErrInvalidGeoZoneIdentifier(ModuleName, fmt.Errorf("geo zone is empty"))
	}
	// ensure length
	if len(h) > GeoZoneIdentifierLength {
		return ErrInvalidGeoZoneIdentifier(ModuleName, fmt.Errorf("geo zone length is > %d", NetworkIdentifierLength))
	}
	return nil
}
