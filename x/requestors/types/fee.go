package types

const (
	StakeFee   = 10000
	UnstakeFee = 10000
	UnjailFee  = 10000
)

var (
	RequestorFeeMap = map[string]int64{
		MsgRequestorStakeName:   StakeFee,
		MsgRequestorUnstakeName: UnstakeFee,
		MsgRequestorUnjailName:  UnjailFee,
	}
)
