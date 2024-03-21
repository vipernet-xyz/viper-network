package types

import (
	"time"

	"github.com/vipernet-xyz/viper-network/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	authexported "github.com/vipernet-xyz/viper-network/x/authentication/exported"
	requestorexported "github.com/vipernet-xyz/viper-network/x/requestors/exported"
	requestorsTypes "github.com/vipernet-xyz/viper-network/x/requestors/types"
	servicersexported "github.com/vipernet-xyz/viper-network/x/servicers/exported"
	servicersTypes "github.com/vipernet-xyz/viper-network/x/servicers/types"
)

type PosKeeper interface {
	RewardForRelays(ctx sdk.Ctx, reportCard MsgSubmitQoSReport, relays sdk.BigInt, requestor requestorsTypes.Requestor) sdk.BigInt
	GetStakedTokens(ctx sdk.Ctx) sdk.BigInt
	Validator(ctx sdk.Ctx, addr sdk.Address) servicersexported.ValidatorI
	TotalTokens(ctx sdk.Ctx) sdk.BigInt
	BurnForChallenge(ctx sdk.Ctx, challenges sdk.BigInt, address sdk.Address)
	JailValidator(ctx sdk.Ctx, addr sdk.Address)
	AllValidators(ctx sdk.Ctx) (validators []servicersexported.ValidatorI)
	GetStakedValidators(ctx sdk.Ctx) (validators []servicersexported.ValidatorI)
	BlocksPerSession(ctx sdk.Ctx) (res int64)
	StakeDenom(ctx sdk.Ctx) (res string)
	GetValidator(ctx sdk.Ctx, addr sdk.Address) (validator servicersTypes.Validator, found bool)
	GetValidatorsByChain(ctx sdk.Ctx, networkID string) (validators []sdk.Address, total int)
	GetValidatorsByGeoZone(ctx sdk.Ctx, geoZone string) (validators []sdk.Address, count int)
	GetStakedValidatorsLimit(ctx sdk.Ctx, maxRetrieve int64) (validators []servicersexported.ValidatorI)
	MaxFishermen(ctx sdk.Ctx) (res int64)
	FishermenCount(ctx sdk.Ctx) (res int64)
	PauseNode(ctx sdk.Ctx, addr sdk.Address) sdk.Error
	BurnforNoActivity(ctx sdk.Ctx, height int64, addr sdk.Address)
	GetHistoricalInfo(ctx sdk.Ctx, height int64) (servicersTypes.HistoricalInfo, bool)
	UnbondingTime(ctx sdk.Ctx) time.Duration
	UpdateValidatorReportCard(ctx sdk.Ctx, addr sdk.Address, sessionReport ViperQoSReport) servicersTypes.ReportCard
	SlashFisherman(ctx sdk.Ctx, height int64, address sdk.Address)
	GetValidatorSigningInfo(ctx sdk.Ctx, addr sdk.Address) (info servicersTypes.ValidatorSigningInfo, found bool)
	SetReportCardMissedAt(ctx sdk.Ctx, addr sdk.Address, index int64, missed bool)
}

type RequestorsKeeper interface {
	GetStakedTokens(ctx sdk.Ctx) sdk.BigInt
	Requestor(ctx sdk.Ctx, addr sdk.Address) requestorexported.RequestorI
	AllRequestors(ctx sdk.Ctx) (requestors []requestorexported.RequestorI)
	TotalTokens(ctx sdk.Ctx) sdk.BigInt
	JailRequestor(ctx sdk.Ctx, addr sdk.Address)
}

type ViperKeeper interface {
	Codec() *codec.Codec
}

type AuthKeeper interface {
	GetFee(ctx sdk.Ctx, msg sdk.Msg) sdk.BigInt
	GetAccount(ctx sdk.Ctx, addr sdk.Address) authexported.Account
}
