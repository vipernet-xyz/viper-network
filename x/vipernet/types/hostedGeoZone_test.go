package types

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHostedGeoZones_GetGeoZone(t *testing.T) {
	zoneID := "zone123"
	testGeoZone := GeoZone{
		ID: zoneID,
	}
	hg := HostedGeoZones{
		M: map[string]GeoZone{testGeoZone.ID: testGeoZone},
		L: sync.Mutex{},
	}
	zone, err := hg.GetGeoZone(zoneID)
	assert.Nil(t, err)
	assert.Equal(t, zone.ID, zoneID)
}

func TestHostedGeoZones_Contains(t *testing.T) {
	zoneID := "zone123"
	otherZoneID := "zone456"
	testGeoZone := GeoZone{
		ID: zoneID,
	}
	hg := HostedGeoZones{
		M: map[string]GeoZone{testGeoZone.ID: testGeoZone},
		L: sync.Mutex{},
	}
	assert.True(t, hg.Contains(zoneID))
	assert.False(t, hg.Contains(otherZoneID))
}

func TestHostedGeoZones_Validate(t *testing.T) {
	zoneID := "zone123"
	testGeoZone := GeoZone{
		ID: zoneID,
	}
	emptyGeoZone := GeoZone{
		ID: "",
	}
	tests := []struct {
		name     string
		hg       *HostedGeoZones
		hasError bool
	}{
		{
			name:     "Valid GeoZone",
			hg:       &HostedGeoZones{M: map[string]GeoZone{testGeoZone.ID: testGeoZone}, L: sync.Mutex{}},
			hasError: false,
		},
		{
			name:     "Invalid GeoZone, Empty ID",
			hg:       &HostedGeoZones{M: map[string]GeoZone{emptyGeoZone.ID: emptyGeoZone}, L: sync.Mutex{}},
			hasError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.hg.Validate() != nil, tt.hasError)
		})
	}
}
