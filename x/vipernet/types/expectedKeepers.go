package types

import (
	"github.com/vipernet-xyz/viper-network/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	appexported "github.com/vipernet-xyz/viper-network/x/apps/exported"
	authexported "github.com/vipernet-xyz/viper-network/x/auth/exported"
	nodesexported "github.com/vipernet-xyz/viper-network/x/nodes/exported"
)

type PosKeeper interface {
	RewardForRelays(ctx sdk.Ctx, relays sdk.BigInt, address sdk.Address, appAddress sdk.Address) sdk.BigInt
	GetStakedTokens(ctx sdk.Ctx) sdk.BigInt
	Validator(ctx sdk.Ctx, addr sdk.Address) nodesexported.ValidatorI
	TotalTokens(ctx sdk.Ctx) sdk.BigInt
	BurnForChallenge(ctx sdk.Ctx, challenges sdk.BigInt, address sdk.Address)
	JailValidator(ctx sdk.Ctx, addr sdk.Address)
	AllValidators(ctx sdk.Ctx) (validators []nodesexported.ValidatorI)
	GetStakedValidators(ctx sdk.Ctx) (validators []nodesexported.ValidatorI)
	BlocksPerSession(ctx sdk.Ctx) (res int64)
	StakeDenom(ctx sdk.Ctx) (res string)
	GetValidatorsByChain(ctx sdk.Ctx, networkID string) (validators []sdk.Address, total int)
}

type AppsKeeper interface {
	GetStakedTokens(ctx sdk.Ctx) sdk.BigInt
	Application(ctx sdk.Ctx, addr sdk.Address) appexported.ApplicationI
	AllApplications(ctx sdk.Ctx) (applications []appexported.ApplicationI)
	TotalTokens(ctx sdk.Ctx) sdk.BigInt
	JailApplication(ctx sdk.Ctx, addr sdk.Address)
}

type ViperKeeper interface {
	SessionNodeCount(ctx sdk.Ctx) (res int64)
	Codec() *codec.Codec
}

type AuthKeeper interface {
	GetFee(ctx sdk.Ctx, msg sdk.Msg) sdk.BigInt
	GetAccount(ctx sdk.Ctx, addr sdk.Address) authexported.Account
}
