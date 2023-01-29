package types

import (
	"github.com/vipernet-xyz/viper-network/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	authexported "github.com/vipernet-xyz/viper-network/x/authentication/exported"
	providerexported "github.com/vipernet-xyz/viper-network/x/providers/exported"
	servicersexported "github.com/vipernet-xyz/viper-network/x/servicers/exported"
)

type PosKeeper interface {
	RewardForRelays(ctx sdk.Ctx, relays sdk.BigInt, address sdk.Address, providerAddress sdk.Address) sdk.BigInt
	GetStakedTokens(ctx sdk.Ctx) sdk.BigInt
	Validator(ctx sdk.Ctx, addr sdk.Address) servicersexported.ValidatorI
	TotalTokens(ctx sdk.Ctx) sdk.BigInt
	BurnForChallenge(ctx sdk.Ctx, challenges sdk.BigInt, address sdk.Address)
	JailValidator(ctx sdk.Ctx, addr sdk.Address)
	AllValidators(ctx sdk.Ctx) (validators []servicersexported.ValidatorI)
	GetStakedValidators(ctx sdk.Ctx) (validators []servicersexported.ValidatorI)
	BlocksPerSession(ctx sdk.Ctx) (res int64)
	StakeDenom(ctx sdk.Ctx) (res string)
	GetValidatorsByChain(ctx sdk.Ctx, networkID string) (validators []sdk.Address, total int)
}

type ProvidersKeeper interface {
	GetStakedTokens(ctx sdk.Ctx) sdk.BigInt
	Provider(ctx sdk.Ctx, addr sdk.Address) providerexported.ProviderI
	AllProviders(ctx sdk.Ctx) (providerlications []providerexported.ProviderI)
	TotalTokens(ctx sdk.Ctx) sdk.BigInt
	JailProvider(ctx sdk.Ctx, addr sdk.Address)
}

type ViperKeeper interface {
	SessionNodeCount(ctx sdk.Ctx) (res int64)
	Codec() *codec.Codec
}

type AuthKeeper interface {
	GetFee(ctx sdk.Ctx, msg sdk.Msg) sdk.BigInt
	GetAccount(ctx sdk.Ctx, addr sdk.Address) authexported.Account
}
