package types

import (
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/authentication/exported"
	providersExported "github.com/vipernet-xyz/viper-network/x/providers/exported"
)

// AuthKeeper defines the expected supply Keeper (noalias)
type AuthKeeper interface {
	// get total supply of tokens
	GetSupply(ctx sdk.Ctx) exported.SupplyI
	// get the address of a module account
	GetModuleAddress(name string) sdk.Address
	// get the module account structure
	GetModuleAccount(ctx sdk.Ctx, moduleName string) exported.ModuleAccountI
	// set module account structure
	SetModuleAccount(sdk.Ctx, exported.ModuleAccountI)
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
}

type PosKeeper interface {
	RewardForRelays(ctx sdk.Ctx, relays sdk.BigInt, address sdk.Address, appAddress sdk.Address)
	GetStakedTokens(ctx sdk.Ctx) sdk.BigInt
	Validator(ctx sdk.Ctx, addr sdk.Address) providersExported.ValidatorI
	TotalTokens(ctx sdk.Ctx) sdk.BigInt
	BurnForChallenge(ctx sdk.Ctx, challenges sdk.BigInt, address sdk.Address)
	JailValidator(ctx sdk.Ctx, addr sdk.Address)
	AllValidators(ctx sdk.Ctx) (validators []providersExported.ValidatorI)
	GetStakedValidators(ctx sdk.Ctx) (validators []providersExported.ValidatorI)
	SessionBlockFrequency(ctx sdk.Ctx) (res int64)
	StakeDenom(ctx sdk.Ctx) (res string)
}
