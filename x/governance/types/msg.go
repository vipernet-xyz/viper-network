package types

import (
	"fmt"

	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
)

// ensure ProtoMsg interface compliance at compile time
var (
	_ sdk.ProtoMsg = &MsgChangeParam{}
	_ sdk.ProtoMsg = &MsgDAOTransfer{}
	_ sdk.ProtoMsg = &MsgUpgrade{}
	_ sdk.ProtoMsg = &MsgStakingKey{}
)

const (
	MsgDAOTransferName = "dao_tranfer"
	MsgChangeParamName = "change_param"
	MsgUpgradeName     = "upgrade"
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

// MsgStakingKey structure for changing governance parameters
// type MsgStakingKey struct {
// 	FromAddress sdk.Address `json:"from_address"`
// 	ToAddress   sdk.Address `json:"to_address"`
// 	PubKey      crypro.PublicKey  `json:"pubKey"`
// 	ClientType  int64      `json:"client_type"`
// }

type MsgStakingKey struct {
	FromAddress sdk.Address      `json:"from_address yaml:"from_address"`
	ToAddress   sdk.Address      `json:"to_address" yaml:"to_address"`
	StakingKey  crypto.PublicKey `json:"pubkey" yaml:"pubkey"`
	ClientType  sdk.Int64        `json:"client_type" yaml:"client_type"`
}

// Route provides router key for msg
func (msg MsgStakingKey) Route() string { return RouterKey }

// Type provides msg name
func (msg MsgStakingKey) Type() string { return MsgDAOTransferName }

// GetFee get fee for msg
func (msg MsgStakingKey) GetFee() sdk.BigInt {
	return sdk.NewInt(GovFeeMap[msg.Type()])
}

// GetSigners return address(es) that must sign over msg.GetSignBytes()
func (msg MsgStakingKey) GetSigners() []sdk.Address {
	return []sdk.Address{msg.FromAddress}
}

// Define a global mapping to store the relationship between toAddress and StakingKey
var stakingKeyMap = make(map[string]crypto.PublicKey)

// Function to map the toAddress to the corresponding StakingKey
func MapToAddressToStakingKey(toAddress sdk.Address, stakingKey crypto.PublicKey) {
	stakingKeyMap[toAddress.String()] = stakingKey
}

// Function to retrieve the StakingKey based on the toAddress
func GetStakingKey(toAddress sdk.Address) (crypto.PublicKey, error) {
	stakingKey, ok := stakingKeyMap[toAddress.String()]
	if !ok {
		return nil, fmt.Errorf("no StakingKey found for the given toAddress")
	}
	return stakingKey, nil
}

// GetSigners return address(es) that must sign over msg.GetSignBytes()
func (msg MsgStakingKey) GetRecipient() sdk.Address {
	return nil
}

func (msg *MsgStakingKey) Reset() {
	*msg = MsgStakingKey{}
}

func (msg MsgStakingKey) String() string {
	return fmt.Sprintf("Public Key: %s\nAddress: %s\nClientType: %s\n", msg.StakingKey.RawString(), msg.ToAddress, msg.ClientType.String())
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgStakingKey) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg MsgStakingKey) ProtoMessage() {
	m := msg.ToProto()
	m.ProtoMessage()
}

func (msg MsgStakingKey) ToProto() MsgProtoStakingKey {
	var pkbz []byte
	if msg.StakingKey != nil {
		pkbz = msg.StakingKey.RawBytes()
	}
	return MsgProtoStakingKey{
		FromAddress: msg.FromAddress,
		ToAddress:   msg.ToAddress,
		PubKey:      pkbz,
		ClientType:  msg.ClientType,
	}
}

// ValidateBasic quick validity check
func (msg MsgStakingKey) ValidateBasic() sdk.Error {
	if msg.FromAddress == nil {
		return sdk.ErrInvalidAddress("nil from address")
	}
	clientType, err := ClientTypeFromNumber(msg.ClientType)
	if err != nil {
		return err
	}
	if clientType == 02 && msg.ToAddress == nil {
		return sdk.ErrInvalidAddress("nil to address")
	}
	return nil
}
