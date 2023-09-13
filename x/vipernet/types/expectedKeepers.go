package types

import (
	"time"

	"github.com/vipernet-xyz/viper-network/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	authexported "github.com/vipernet-xyz/viper-network/x/authentication/exported"
	providerexported "github.com/vipernet-xyz/viper-network/x/providers/exported"
	providersTypes "github.com/vipernet-xyz/viper-network/x/providers/types"
	servicersexported "github.com/vipernet-xyz/viper-network/x/servicers/exported"
	servicersTypes "github.com/vipernet-xyz/viper-network/x/servicers/types"
)

type PosKeeper interface {
	RewardForRelays(ctx sdk.Ctx, relays sdk.BigInt, address sdk.Address, provider providersTypes.Provider) sdk.BigInt
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
	GetValidatorsByGeoZone(ctx sdk.Ctx, geoZone string) (validators []sdk.Address, count int)
	GetStakedValidatorsLimit(ctx sdk.Ctx, maxRetrieve int64) (validators []servicersexported.ValidatorI)
	MaxFishermen(ctx sdk.Ctx) (res int64)
	FishermenCount(ctx sdk.Ctx) (res int64)
	PauseNode(ctx sdk.Ctx, addr sdk.Address)
	BurnforNoActivity(ctx sdk.Ctx, addr sdk.Address)
	GetHistoricalInfo(ctx sdk.Ctx, height int64) (servicersTypes.HistoricalInfo, bool)
	UnbondingTime(ctx sdk.Ctx) time.Duration
}

type ProvidersKeeper interface {
	GetStakedTokens(ctx sdk.Ctx) sdk.BigInt
	Provider(ctx sdk.Ctx, addr sdk.Address) providerexported.ProviderI
	AllProviders(ctx sdk.Ctx) (providers []providerexported.ProviderI)
	TotalTokens(ctx sdk.Ctx) sdk.BigInt
	JailProvider(ctx sdk.Ctx, addr sdk.Address)
}

type ViperKeeper interface {
	Codec() *codec.Codec
}

type AuthKeeper interface {
	GetFee(ctx sdk.Ctx, msg sdk.Msg) sdk.BigInt
	GetAccount(ctx sdk.Ctx, addr sdk.Address) authexported.Account
}
