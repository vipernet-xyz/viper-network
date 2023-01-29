package types

import (
	"encoding/binary"
	"time"

	sdk "github.com/vipernet-xyz/viper-network/types"
)

const (
	ModuleName   = "provider"                // name of module
	StoreKey     = ModuleName                // StoreKey is the string store representation
	TStoreKey    = "transient_" + ModuleName // TStoreKey is the string transient store representation
	QuerierRoute = ModuleName                // QuerierRoute is the querier route for the staking module
	RouterKey    = ModuleName                // RouterKey is the msg router key for the staking module
)

var (
	AllProvidersKey       = []byte{0x01} // prefix for each key to a provider
	StakedProvidersKey    = []byte{0x02} // prefix for each key to a staked provider index, sorted by power
	UnstakingProvidersKey = []byte{0x03} // prefix for unstaking provider
	BurnProviderKey       = []byte{0x04} // prefix for awarding providers
)

// Removes the prefix bytes from a key to expose true address
func AddressFromKey(key []byte) []byte {
	return key[1:] // remove prefix bytes
}

// generates the key for the provider with address
func KeyForProviderByAllProviders(addr sdk.Address) []byte {
	return append(AllProvidersKey, addr.Bytes()...)
}

// generates the key for unstaking providers by the unstakingtime
func KeyForUnstakingProviders(unstakingTime time.Time) []byte {
	bz := sdk.FormatTimeBytes(unstakingTime)
	return append(UnstakingProvidersKey, bz...) // use the unstaking time as part of the key
}

// generates the key for a provider in the staking set
func KeyForProviderInStakingSet(provider Provider) []byte {
	// NOTE the address doesn't need to be stored because counter bytes must always be different
	return getStakedValPowerRankKey(provider)
}

func KeyForProviderBurn(address sdk.Address) []byte {
	return append(BurnProviderKey, address...)
}

// get the power ranking key of a provider
// NOTE the larger values are of higher value
func getStakedValPowerRankKey(provider Provider) []byte {
	// get the consensus power
	consensusPower := sdk.TokensToConsensusPower(provider.StakedTokens)
	consensusPowerBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(consensusPowerBytes, uint64(consensusPower))

	powerBytes := consensusPowerBytes
	powerBytesLen := len(powerBytes) // 8

	// key is of format prefix || powerbytes || addrBytes
	key := make([]byte, 1+powerBytesLen+sdk.AddrLen)

	// generate the key for this provider by deriving it from the main key
	key[0] = StakedProvidersKey[0]
	copy(key[1:powerBytesLen+1], powerBytes)
	operAddrInvr := sdk.CopyBytes(provider.Address)
	for i, b := range operAddrInvr {
		operAddrInvr[i] = ^b
	}
	copy(key[powerBytesLen+1:], operAddrInvr)

	return key
}
