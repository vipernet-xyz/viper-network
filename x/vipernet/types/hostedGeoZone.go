package types

import (
	"sync"

	sdk "github.com/vipernet-xyz/viper-network/types"
)

// GeoZone - An object that represents a geo zone
type GeoZone struct {
	ID        string    `json:"id"` // identifier of the geo zone
	BasicAuth BasicAuth `json:"basic_auth"`
}

// GeoZones - An object that represents the geo zones
type HostedGeoZones struct {
	M map[string]GeoZone // M[id] -> id
	L sync.Mutex
}

// Contains - Checks if the geo zone exists within the GeoZones object
func (g *HostedGeoZones) Contains(id string) bool {
	g.L.Lock()
	defer g.L.Unlock()
	// Quick map check
	_, found := g.M[id]
	return found
}

// GetGeoZone - Returns the geo zone or an error using the geo zone identifier
func (g *HostedGeoZones) GetGeoZone(id string) (zone GeoZone, err sdk.Error) {
	g.L.Lock()
	defer g.L.Unlock()
	// Map check
	res, found := g.M[id]
	if !found {
		return GeoZone{}, NewErrorGeoZoneNotHostedError(ModuleName)
	}
	return res, nil
}

// Validate - Validates the geo zone objects
func (g *HostedGeoZones) Validate() error {
	g.L.Lock()
	defer g.L.Unlock()
	// Loop through all the geo zones
	for _, zone := range g.M {
		// Validate not empty
		if zone.ID == "" {
			return NewInvalidGeoZoneError(ModuleName)
		}
	}
	return nil
}
