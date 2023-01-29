package types

const (
	StakeFee   = 10000
	UnstakeFee = 10000
	UnjailFee  = 10000
)

var (
	ProviderFeeMap = map[string]int64{
		MsgProviderStakeName:   StakeFee,
		MsgProviderUnstakeName: UnstakeFee,
		MsgProviderUnjailName:  UnjailFee,
	}
)
