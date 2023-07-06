package keeper

import (
	vc "github.com/vipernet-xyz/viper-network/x/vipernet/types"
)

// "GetHostedBlockchains" returns the non native chains hosted locally on this servicer
func (k Keeper) GetHostedGeoZone() *vc.HostedGeoZones {
	return k.hostedGeoZone
}

func (k Keeper) SetHostedGeoZone(m map[string]vc.GeoZone) *vc.HostedGeoZones {
	k.hostedGeoZone.L.Lock()
	k.hostedGeoZone.M = m
	k.hostedGeoZone.L.Unlock()
	return k.hostedGeoZone
}
