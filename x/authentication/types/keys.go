package types

import (
	sdk "github.com/vipernet-xyz/viper-network/types"
)

const (
	// module name
	ModuleName = "authentication"
	// storeKey is string representation of the store key for authentication
	StoreKey = ModuleName
	// FeeCollectorName the root string for the fee collector account address
	FeeCollectorName = "fee_collector"
	// QuerierRoute is the querier route for authentication
	QuerierRoute = StoreKey
	// default codespace
	DefaultCodespace = ModuleName
)

var (
	// AddressStoreKeyPrefix prefix for account-by-address store
	SupplyKeyPrefix       = []byte{0x00}
	AddressStoreKeyPrefix = []byte{0x01}
	// SendEnabledPrefix is the prefix for the SendDisabled flags for a Denom.
	SendEnabledPrefix = []byte{0x04}
)

// AddressStoreKey turn an address to key used to get it from the account store
func AddressStoreKey(addr sdk.Address) []byte {
	return append(AddressStoreKeyPrefix, addr.Bytes()...)
}

// CreateSendEnabledKey creates the key of the SendDisabled flag for a denom.
func CreateSendEnabledKey(denom string) []byte {
	key := make([]byte, len(SendEnabledPrefix)+len(denom))
	copy(key, SendEnabledPrefix)
	copy(key[len(SendEnabledPrefix):], denom)
	return key
}
