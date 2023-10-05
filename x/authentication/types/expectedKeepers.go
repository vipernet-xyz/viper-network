package types

import sdk "github.com/vipernet-xyz/viper-network/types"

type PosKeeper interface {
	GetMsgStakeOutputSigner(sdk.Ctx, sdk.Msg) sdk.Address
}
