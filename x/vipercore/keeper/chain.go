package keeper

import (
	vc "github.com/vipernet-xyz/viper-network/x/vipercore/types"
)

// "GetHostedBlockchains" returns the non native chains hosted locally on this node
func (k Keeper) GetHostedBlockchains() *vc.HostedBlockchains {
	return k.hostedBlockchains
}

func (k Keeper) SetHostedBlockchains(m map[string]vc.HostedBlockchain) *vc.HostedBlockchains {
	k.hostedBlockchains.L.Lock()
	k.hostedBlockchains.M = m
	k.hostedBlockchains.L.Unlock()
	return k.hostedBlockchains
}
