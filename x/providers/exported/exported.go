package exported

import (
	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
)

// ProviderI expected provider functions
type ProviderI interface {
	IsJailed() bool             // whether the provider is jailed
	GetStatus() sdk.StakeStatus // status of the provider
	IsStaked() bool             // check if has a staked status
	IsUnstaked() bool           // check if has status unstaked
	IsUnstaking() bool          // check if has status unstaking
	GetChains() []string        // retrieve the staked chains
	GetGeoZones() []string
	GetAddress() sdk.Address        // operator address to receive/return providers coins
	GetPublicKey() crypto.PublicKey // validation consensus pubkey
	GetTokens() sdk.BigInt          // validation tokens
	GetMaxRelays() sdk.BigInt       // maximum relays
	GetNumServicers() int32
}
