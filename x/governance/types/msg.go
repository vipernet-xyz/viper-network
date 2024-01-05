package types

import (
	sdk "github.com/vipernet-xyz/viper-network/types"
)

// ensure ProtoMsg interface compliance at compile time
var (
	_ sdk.ProtoMsg = &MsgChangeParam{}
	_ sdk.ProtoMsg = &MsgDAOTransfer{}
	_ sdk.ProtoMsg = &MsgUpgrade{}
	_ sdk.ProtoMsg = &MsgGenerateDiscountKey{}
)

const (
	MsgDAOTransferName         = "dao_tranfer"
	MsgChangeParamName         = "change_param"
	MsgUpgradeName             = "upgrade"
	MsgGenerateDiscountKeyName = "generate_discount_key"
)

//----------------------------------------------------------------------------------------------------------------------
// MsgChangeParam structure for changing governance parameters
// type MsgChangeParam struct {
// 	FromAddress sdk.Address `json:"address"`
// 	ParamKey    string      `json:"param_key"`
// 	ParamVal    []byte      `json:"param_value"`
// }

// Route provides router key for msg
func (msg MsgChangeParam) Route() string { return RouterKey }

// Type provides msg name
func (msg MsgChangeParam) Type() string { return MsgChangeParamName }

// GetFee get fee for msg
func (msg MsgChangeParam) GetFee() sdk.BigInt {
	return sdk.NewInt(GovFeeMap[msg.Type()])
}

// GetSigners return address(es) that must sign over msg.GetSignBytes()
func (msg MsgChangeParam) GetSigners() []sdk.Address {
	return []sdk.Address{msg.FromAddress}
}

// GetSigners return address(es) that must sign over msg.GetSignBytes()
func (msg MsgChangeParam) GetRecipient() sdk.Address {
	return nil
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgChangeParam) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic quick validity check
func (msg MsgChangeParam) ValidateBasic() sdk.Error {
	if msg.FromAddress == nil {
		return sdk.ErrInvalidAddress("nil address")
	}
	if msg.ParamKey == "" {
		return ErrEmptyKey(ModuleName)
	}
	if msg.ParamVal == nil {
		return ErrEmptyValue(ModuleName)
	}
	return nil
}

//----------------------------------------------------------------------------------------------------------------------

// MsgDAOTransfer structure for changing governance parameters
// type MsgDAOTransfer struct {
// 	FromAddress sdk.Address `json:"from_address"`
// 	ToAddress   sdk.Address `json:"to_address"`
// 	Amount      sdk.BigInt     `json:"amount"`
// 	Action      string      `json:"action"`
// }

// Route provides router key for msg
func (msg MsgDAOTransfer) Route() string { return RouterKey }

// Type provides msg name
func (msg MsgDAOTransfer) Type() string { return MsgDAOTransferName }

// GetFee get fee for msg
func (msg MsgDAOTransfer) GetFee() sdk.BigInt {
	return sdk.NewInt(GovFeeMap[msg.Type()])
}

// GetSigners return address(es) that must sign over msg.GetSignBytes()
func (msg MsgDAOTransfer) GetSigners() []sdk.Address {
	return []sdk.Address{msg.FromAddress}
}

// GetSigners return address(es) that must sign over msg.GetSignBytes()
func (msg MsgDAOTransfer) GetRecipient() sdk.Address {
	return nil
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgDAOTransfer) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic quick validity check
func (msg MsgDAOTransfer) ValidateBasic() sdk.Error {
	if msg.FromAddress == nil {
		return sdk.ErrInvalidAddress("nil from address")
	}
	if msg.Amount.Int64() == 0 {
		return ErrZeroValueDAOAction(ModuleName)
	}
	daoAction, err := DAOActionFromString(msg.Action)
	if err != nil {
		return err
	}
	if daoAction == DAOTransfer && msg.ToAddress == nil {
		return sdk.ErrInvalidAddress("nil to address")
	}
	return nil
}

//----------------------------------------------------------------------------------------------------------------------

// MsgGenerateDiscountKey structure for generating a discount key for a requestor
//type MsgGenerateDiscountKey struct {
//	FromAddress sdk.Address `json:"from_address"`
//	ToAddress   sdk.Address `json:"to_address"`
//	DiscountKey string      `json:"discount_key"`
//}

// Route provides router key for msg
func (msg MsgGenerateDiscountKey) Route() string { return RouterKey }

// Type provides msg name
func (msg MsgGenerateDiscountKey) Type() string { return MsgGenerateDiscountKeyName }

// GetFee get fee for msg
// This assumes you have a similar GovFeeMap mechanism for determining fees
func (msg MsgGenerateDiscountKey) GetFee() sdk.BigInt {
	return sdk.NewInt(GovFeeMap[msg.Type()])
}

// GetSigners return address(es) that must sign over msg.GetSignBytes()
func (msg MsgGenerateDiscountKey) GetSigners() []sdk.Address {
	return []sdk.Address{msg.FromAddress}
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgGenerateDiscountKey) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners return address(es) that must sign over msg.GetSignBytes()
func (msg MsgGenerateDiscountKey) GetRecipient() sdk.Address {
	return nil
}

// ValidateBasic quick validity check
func (msg MsgGenerateDiscountKey) ValidateBasic() sdk.Error {
	if msg.FromAddress == nil {
		return sdk.ErrInvalidAddress("nil from address")
	}
	if msg.ToAddress == nil {
		return sdk.ErrInvalidAddress("nil to address")
	}
	if len(msg.DiscountKey) == 0 {
		return sdk.ErrUnknownRequest("Discount Key cannot be empty")
	}
	return nil
}

//----------------------------------------------------------------------------------------------------------------------

// MsgUpgrade structure for changing governance parameters
// type MsgUpgrade struct {
// 	Address sdk.Address `json:"address"`
// 	Upgrade Upgrade     `json:"upgrade"`
// }

// Route provides router key for msg
func (msg MsgUpgrade) Route() string { return RouterKey }

// Type provides msg name
func (msg MsgUpgrade) Type() string { return MsgUpgradeName }

// GetFee get fee for msg
func (msg MsgUpgrade) GetFee() sdk.BigInt {
	return sdk.NewInt(GovFeeMap[msg.Type()])
}

// GetSigners return address(es) that must sign over msg.GetSignBytes()
func (msg MsgUpgrade) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Address}
}

// GetSigners return address(es) that must sign over msg.GetSignBytes()
func (msg MsgUpgrade) GetRecipient() sdk.Address {
	return nil
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgUpgrade) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic quick validity check
func (msg MsgUpgrade) ValidateBasic() sdk.Error {
	if msg.Address == nil {
		return sdk.ErrInvalidAddress("nil from address")
	}
	if msg.Upgrade.UpgradeHeight() == 0 {
		return ErrZeroHeightUpgrade(ModuleName)
	}
	if msg.Upgrade.UpgradeVersion() == "" {
		return ErrZeroHeightUpgrade(ModuleName)
	}
	return nil
}
