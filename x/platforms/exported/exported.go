package exported

import (
	"github.com/vipernet-xyz/viper-network/crypto"
	sdk "github.com/vipernet-xyz/viper-network/types"
)

// PlatformI expected platform functions
type PlatformI interface {
	IsJailed() bool                 // whether the platform is jailed
	GetStatus() sdk.StakeStatus     // status of the platform
	IsStaked() bool                 // check if has a staked status
	IsUnstaked() bool               // check if has status unstaked
	IsUnstaking() bool              // check if has status unstaking
	GetChains() []string            // retrieve the staked chains
	GetAddress() sdk.Address        // operator address to receive/return platforms coins
	GetPublicKey() crypto.PublicKey // validation consensus pubkey
	GetTokens() sdk.BigInt          // validation tokens
	GetMaxRelays() sdk.BigInt       // maximum relays
}
