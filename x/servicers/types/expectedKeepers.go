package types

import (
	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	authexported "github.com/vipernet-xyz/viper-network/x/authentication/exported"
	providerexported "github.com/vipernet-xyz/viper-network/x/providers/exported"
	posexported "github.com/vipernet-xyz/viper-network/x/servicers/exported"
)

// AuthKeeper defines the expected supply Keeper (noalias)
type AuthKeeper interface {
	// get total supply of tokens
	GetSupply(ctx sdk.Ctx) authexported.SupplyI
	// set total supply of tokens
	SetSupply(ctx sdk.Ctx, supply authexported.SupplyI)
	// get the address of a module account
	GetModuleAddress(name string) sdk.Address
	// get the module account structure
	GetModuleAccount(ctx sdk.Ctx, moduleName string) authexported.ModuleAccountI
	// set module account structure
	SetModuleAccount(sdk.Ctx, authexported.ModuleAccountI)
	// send coins to/from module accounts
	SendCoinsFromModuleToModule(ctx sdk.Ctx, senderModule, recipientModule string, amt sdk.Coins) sdk.Error
	// send coins from module to validator
	SendCoinsFromModuleToAccount(ctx sdk.Ctx, senderModule string, recipientAddr sdk.Address, amt sdk.Coins) sdk.Error
	// send coins from validator to module
	SendCoinsFromAccountToModule(ctx sdk.Ctx, senderAddr sdk.Address, recipientModule string, amt sdk.Coins) sdk.Error
	// mint coins
	MintCoins(ctx sdk.Ctx, moduleName string, amt sdk.Coins) sdk.Error
	// burn coins
	BurnCoins(ctx sdk.Ctx, name string, amt sdk.Coins) sdk.Error
	// iterate accounts
	IterateAccounts(ctx sdk.Ctx, process func(authexported.Account) (stop bool))
	// get coins
	GetCoins(ctx sdk.Ctx, addr sdk.Address) sdk.Coins
	// set coins
	SetCoins(ctx sdk.Ctx, addr sdk.Address, amt sdk.Coins) sdk.Error
	// has coins
	HasCoins(ctx sdk.Ctx, addr sdk.Address, amt sdk.Coins) bool
	// send coins
	SendCoins(ctx sdk.Ctx, fromAddr sdk.Address, toAddr sdk.Address, amt sdk.Coins) sdk.Error
	// get account
	GetAccount(ctx sdk.Ctx, addr sdk.Address) authexported.Account
}

type ViperKeeper interface {
	// clear the cache of validators for sessions and relays
	ClearSessionCache()
}

// ValidatorSet expected properties for the set of all validators (noalias)
type ValidatorSet interface {
	// iterate through validators by address, execute func for each validator
	IterateAndExecuteOverVals(sdk.Ctx, func(index int64, validator posexported.ValidatorI) (stop bool))
	// iterate through staked validators by address, execute func for each validator
	IterateAndExecuteOverStakedVals(sdk.Ctx, func(index int64, validator posexported.ValidatorI) (stop bool))
	// iterate through the validator set of the prevState block by address, execute func for each validator
	IterateAndExecuteOverPrevStateVals(sdk.Ctx, func(index int64, validator posexported.ValidatorI) (stop bool))
	// get a particular validator by address
	Validator(sdk.Ctx, sdk.Address) posexported.ValidatorI
	// total staked tokens within the validator set
	TotalTokens(sdk.Ctx) sdk.BigInt
	// jail a validator
	JailValidator(sdk.Ctx, sdk.Address)
	// unjail a validator
	UnjailValidator(sdk.Ctx, sdk.Address)
	// MaxValidators returns the maximum amount of staked validators
	MaxValidators(sdk.Ctx) int64
	ServicerCountLock(ctx sdk.Ctx) (isOn bool)
}

type ProvidersKeeper interface {
	CalculateProviderRelays(ctx sdk.Ctx, provider providerexported.ProviderI) sdk.BigInt
	GetStakedTokens(ctx sdk.Ctx) sdk.BigInt
	Provider(ctx sdk.Ctx, addr sdk.Address) providerexported.ProviderI
	AllProviders(ctx sdk.Ctx) (providers []providerexported.ProviderI)
	TotalTokens(ctx sdk.Ctx) sdk.BigInt
	JailProvider(ctx sdk.Ctx, addr sdk.Address)
	ForceProviderUnstake(ctx sdk.Ctx, provider providerexported.ProviderI) sdk.Error
	LegacyForceProviderUnstake(ctx sdk.Ctx, provider providerexported.ProviderI) sdk.Error
	MinimumStake(ctx sdk.Ctx) (res int64)
	SetProvider(ctx sdk.Ctx, provider providerexported.ProviderI)
	BaselineThroughputStakeRate(ctx sdk.Ctx) (base int64)
	GetStakingKey(ctx sdk.Ctx, address sdk.Address) (string, error) // staking key
	SetStakingKey(ctx sdk.Ctx, address sdk.Address, stakingKey crypto.PublicKey)
}
