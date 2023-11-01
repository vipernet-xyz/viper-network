package keeper

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vipernet-xyz/viper-network/x/vipernet/types"
	vc "github.com/vipernet-xyz/viper-network/x/vipernet/types"
)

func TestKeeper_GetHostedGeoZone(t *testing.T) {
	// Define test geo zones
	geo1 := types.GeoZone{
		ID: "geozone1",
		BasicAuth: types.BasicAuth{
			Username: "user1",
			Password: "pass1",
		},
	}
	geo2 := types.GeoZone{
		ID: "geozone2",
		BasicAuth: types.BasicAuth{
			Username: "user2",
			Password: "pass2",
		},
	}

	// Create a test environment
	_, _, _, _, keeper, _, _ := createTestInput(t, false)

	// Set the hosted geo zones in the keeper
	geoZones := map[string]vc.GeoZone{
		geo1.ID: geo1,
		geo2.ID: geo2,
	}
	keeper.SetHostedGeoZone(geoZones)

	// Retrieve the hosted geo zones
	geoZoneKeeper := keeper.GetHostedGeoZone()
	assert.NotNil(t, geoZoneKeeper)

	// Check if geo zones exist
	assert.True(t, geoZoneKeeper.Contains(geo1.ID))
	assert.True(t, geoZoneKeeper.Contains(geo2.ID))
}
