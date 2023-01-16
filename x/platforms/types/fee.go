package types

const (
	StakeFee   = 10000
	UnstakeFee = 10000
	UnjailFee  = 10000
)

var (
	PlatformFeeMap = map[string]int64{
		MsgPlatformStakeName:   StakeFee,
		MsgPlatformUnstakeName: UnstakeFee,
		MsgPlatformUnjailName:  UnjailFee,
	}
)
