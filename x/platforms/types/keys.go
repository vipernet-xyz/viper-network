package types

import (
	"encoding/binary"
	"time"

	sdk "github.com/vipernet-xyz/viper-network/types"
)

const (
	ModuleName   = "platform"                // name of module
	StoreKey     = ModuleName                // StoreKey is the string store representation
	TStoreKey    = "transient_" + ModuleName // TStoreKey is the string transient store representation
	QuerierRoute = ModuleName                // QuerierRoute is the querier route for the staking module
	RouterKey    = ModuleName                // RouterKey is the msg router key for the staking module
)

var (
	AllPlatformsKey       = []byte{0x01} // prefix for each key to a platform
	StakedPlatformsKey    = []byte{0x02} // prefix for each key to a staked platform index, sorted by power
	UnstakingPlatformsKey = []byte{0x03} // prefix for unstaking platform
	BurnPlatformKey       = []byte{0x04} // prefix for awarding platforms
)

// Removes the prefix bytes from a key to expose true address
func AddressFromKey(key []byte) []byte {
	return key[1:] // remove prefix bytes
}

// generates the key for the platform with address
func KeyForPlatformByAllPlatforms(addr sdk.Address) []byte {
	return append(AllPlatformsKey, addr.Bytes()...)
}

// generates the key for unstaking platforms by the unstakingtime
func KeyForUnstakingPlatforms(unstakingTime time.Time) []byte {
	bz := sdk.FormatTimeBytes(unstakingTime)
	return append(UnstakingPlatformsKey, bz...) // use the unstaking time as part of the key
}

// generates the key for a platform in the staking set
func KeyForPlatformInStakingSet(platform Platform) []byte {
	// NOTE the address doesn't need to be stored because counter bytes must always be different
	return getStakedValPowerRankKey(platform)
}

func KeyForPlatformBurn(address sdk.Address) []byte {
	return append(BurnPlatformKey, address...)
}

// get the power ranking key of a platform
// NOTE the larger values are of higher value
func getStakedValPowerRankKey(platform Platform) []byte {
	// get the consensus power
	consensusPower := sdk.TokensToConsensusPower(platform.StakedTokens)
	consensusPowerBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(consensusPowerBytes, uint64(consensusPower))

	powerBytes := consensusPowerBytes
	powerBytesLen := len(powerBytes) // 8

	// key is of format prefix || powerbytes || addrBytes
	key := make([]byte, 1+powerBytesLen+sdk.AddrLen)

	// generate the key for this platform by deriving it from the main key
	key[0] = StakedPlatformsKey[0]
	copy(key[1:powerBytesLen+1], powerBytes)
	operAddrInvr := sdk.CopyBytes(platform.Address)
	for i, b := range operAddrInvr {
		operAddrInvr[i] = ^b
	}
	copy(key[powerBytesLen+1:], operAddrInvr)

	return key
}
