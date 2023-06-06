package types

import (
	sdk "github.com/vipernet-xyz/viper-network/types"
)

const (
	DAOAccountName              = "dao"
	DAOTransferString           = "dao_transfer"
	DAOBurnString               = "dao_burn"
	DAOTransfer       DAOAction = iota + 1
	DAOBurn
	DappNumber      = 01
	NaasToolsNumber = 02
)

type DAOAction int

func (da DAOAction) String() string {
	switch da {
	case DAOTransfer:
		return DAOTransferString
	case DAOBurn:
		return DAOBurnString
	}
	return ""
}

func DAOActionFromString(s string) (DAOAction, sdk.Error) {
	switch s {
	case DAOTransferString:
		return DAOTransfer, nil
	case DAOBurnString:
		return DAOBurn, nil
	default:
		return 0, ErrUnrecognizedDAOAction(ModuleName, s)
	}
}

func ClientTypeFromNumber(s sdk.Int64) (int64, sdk.Error) {
	switch s {
	case DappNumber:
		return 01, nil
	case NaasToolsNumber:
		return 02, nil
	default:
		return 0, ErrUnrecognizedClientType(ModuleName, s)
	}
}
