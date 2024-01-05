package types

import (
	"encoding/binary"
	"time"

	sdk "github.com/vipernet-xyz/viper-network/types"
)

const (
	ModuleName   = "requestor"               // name of module
	StoreKey     = ModuleName                // StoreKey is the string store representation
	TStoreKey    = "transient_" + ModuleName // TStoreKey is the string transient store representation
	QuerierRoute = ModuleName                // QuerierRoute is the querier route for the staking module
	RouterKey    = ModuleName                // RouterKey is the msg router key for the staking module
	MemStoreKey  = "memory_" + ModuleName
)

var (
	AllRequestorsKey       = []byte{0x01} // prefix for each key to a requestor
	StakedRequestorsKey    = []byte{0x02} // prefix for each key to a staked requestor index, sorted by power
	UnstakingRequestorsKey = []byte{0x03} // prefix for unstaking requestor
	BurnRequestorKey       = []byte{0x04} // prefix for awarding requestors
)

// Removes the prefix bytes from a key to expose true address
func AddressFromKey(key []byte) []byte {
	return key[1:] // remove prefix bytes
}

// generates the key for the requestor with address
func KeyForRequestorByAllRequestors(addr sdk.Address) []byte {
	return append(AllRequestorsKey, addr.Bytes()...)
}

// generates the key for unstaking requestors by the unstakingtime
func KeyForUnstakingRequestors(unstakingTime time.Time) []byte {
	bz := sdk.FormatTimeBytes(unstakingTime)
	return append(UnstakingRequestorsKey, bz...) // use the unstaking time as part of the key
}

// generates the key for a requestor in the staking set
func KeyForRequestorInStakingSet(requestor Requestor) []byte {
	// NOTE the address doesn't need to be stored because counter bytes must always be different
	return getStakedValPowerRankKey(requestor)
}

func KeyForRequestorBurn(address sdk.Address) []byte {
	return append(BurnRequestorKey, address...)
}

// get the power ranking key of a requestor
// NOTE the larger values are of higher value
func getStakedValPowerRankKey(requestor Requestor) []byte {
	// get the consensus power
	consensusPower := sdk.TokensToConsensusPower(requestor.StakedTokens)
	consensusPowerBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(consensusPowerBytes, uint64(consensusPower))

	powerBytes := consensusPowerBytes
	powerBytesLen := len(powerBytes) // 8

	// key is of format prefix || powerbytes || addrBytes
	key := make([]byte, 1+powerBytesLen+sdk.AddrLen)

	// generate the key for this requestor by deriving it from the main key
	key[0] = StakedRequestorsKey[0]
	copy(key[1:powerBytesLen+1], powerBytes)
	operAddrInvr := sdk.CopyBytes(requestor.Address)
	for i, b := range operAddrInvr {
		operAddrInvr[i] = ^b
	}
	copy(key[powerBytesLen+1:], operAddrInvr)

	return key
}
