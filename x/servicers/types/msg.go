package types

import (
	"fmt"

	"github.com/vipernet-xyz/viper-network/codec"
	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
)

// ensure ProtoMsg interface compliance at compile time
var (
	_ sdk.ProtoMsg = &MsgBeginUnstake{}
	_ sdk.ProtoMsg = &MsgUnjail{}
	_ sdk.ProtoMsg = &MsgSend{}
	_ sdk.ProtoMsg = &MsgStake{}
	_ sdk.ProtoMsg = &MsgPause{}
	_ sdk.ProtoMsg = &MsgUnpause{}
)

const (
	MsgStakeName   = "stake_validator"
	MsgUnstakeName = "begin_unstake_validator"
	MsgUnjailName  = "unjail_validator"
	MsgSendName    = "send"
	MsgPauseName   = "pause"
	MsgUnpauseName = "unpause"
)

// ----------------------------------------------------------------------------------------------------------------------
// GetSigners return address(es) that must sign over msg.GetSignBytes()
func (msg MsgBeginUnstake) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Signer, msg.Address}
}

func (msg MsgBeginUnstake) GetRecipient() sdk.Address {
	return nil
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgBeginUnstake) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic quick validity check, stateless
func (msg MsgBeginUnstake) ValidateBasic() sdk.Error {
	if msg.Address.Empty() {
		return ErrNilValidatorAddr(DefaultCodespace)
	}
	if msg.Signer.Empty() {
		return ErrNilSignerAddr(DefaultCodespace)
	}
	return nil
}

// Route provides router key for msg
func (msg MsgBeginUnstake) Route() string { return RouterKey }

// Type provides msg name
func (msg MsgBeginUnstake) Type() string { return MsgUnstakeName }

// GetFee get fee for msg
func (msg MsgBeginUnstake) GetFee() sdk.BigInt {
	return sdk.NewInt(NodeFeeMap[msg.Type()])
}

//----------------------------------------------------------------------------------------------------------------------

// GetSigners return address(es) that must sign over msg.GetSignBytes()
func (msg MsgUnjail) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Signer, msg.ValidatorAddr}
}

func (msg MsgUnjail) GetRecipient() sdk.Address {
	return nil
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgUnjail) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic quick validity check, stateless
func (msg MsgUnjail) ValidateBasic() sdk.Error {
	if msg.ValidatorAddr.Empty() {
		return ErrNoValidatorFound(DefaultCodespace)
	}
	if msg.Signer.Empty() {
		return ErrNilSignerAddr(DefaultCodespace)
	}
	return nil
}

// Route provides router key for msg
func (msg MsgUnjail) Route() string { return RouterKey }

// Type provides msg name
func (msg MsgUnjail) Type() string { return MsgUnjailName }

// GetFee get fee for msg
func (msg MsgUnjail) GetFee() sdk.BigInt {
	return sdk.NewInt(NodeFeeMap[msg.Type()])
}

//----------------------------------------------------------------------------------------------------------------------

// GetSigners return address(es) that must sign over msg.GetSignBytes()
func (msg MsgSend) GetSigners() []sdk.Address {
	return []sdk.Address{msg.FromAddress}
}

func (msg MsgSend) GetRecipient() sdk.Address {
	return msg.ToAddress
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgSend) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic quick validity check, stateless
func (msg MsgSend) ValidateBasic() sdk.Error {
	if msg.FromAddress.Empty() {
		return ErrNoValidatorFound(DefaultCodespace)
	}
	if msg.ToAddress.Empty() {
		return ErrNoValidatorFound(DefaultCodespace)
	}
	if msg.Amount.LTE(sdk.ZeroInt()) {
		return ErrBadSendAmount(DefaultCodespace)
	}
	return nil
}

// Route provides router key for msg
func (msg MsgSend) Route() string { return RouterKey }

// Type provides msg name
func (msg MsgSend) Type() string { return MsgSendName }

// GetFee get fee for msg
func (msg MsgSend) GetFee() sdk.BigInt {
	return sdk.NewInt(NodeFeeMap[msg.Type()])
}

// ----------------------------------------------------------------------------------------------------------------------
var _ codec.ProtoMarshaler = &MsgStake{}

// MsgStake - struct for staking transactions
type MsgStake struct {
	PublicKey  crypto.PublicKey `json:"public_key" yaml:"public_key"`
	Chains     []string         `json:"chains" yaml:"chains"`
	Value      sdk.BigInt       `json:"value" yaml:"value"`
	ServiceUrl string           `json:"service_url" yaml:"service_url"`
	GeoZone    []string         `json:"geo_zone" yaml:"geo_zone"`
	Output     sdk.Address      `json:"output_address,omitempty" yaml:"output_address"`
}

func (msg *MsgStake) Marshal() ([]byte, error) {
	p := msg.ToProto()
	return p.Marshal()
}

func (msg *MsgStake) MarshalTo(data []byte) (n int, err error) {
	p := msg.ToProto()
	return p.MarshalTo(data)
}

func (msg *MsgStake) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	p := msg.ToProto()
	return p.MarshalToSizedBuffer(dAtA)
}

func (msg *MsgStake) Size() int {
	p := msg.ToProto()
	return p.Size()
}

func (msg *MsgStake) Unmarshal(data []byte) error {
	var m MsgProtoStake
	err := m.Unmarshal(data)
	if err != nil {
		return err
	}
	publicKey, err := crypto.NewPublicKeyBz(m.Publickey)
	if err != nil {
		return err
	}
	newMsg := MsgStake{
		PublicKey:  publicKey,
		Chains:     m.Chains,
		Value:      m.Value,
		ServiceUrl: m.ServiceUrl,
		GeoZone:    m.GeoZone,
		Output:     m.OutputAddress,
	}
	*msg = newMsg
	return nil
}

// GetSigners return address(es) that could sign over msg.GetSignBytes()
func (msg MsgStake) GetSigners() []sdk.Address {
	return []sdk.Address{sdk.Address(msg.PublicKey.Address()), msg.Output}
}

func (msg MsgStake) GetRecipient() sdk.Address {
	return nil
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgStake) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic quick validity check, stateless
func (msg MsgStake) ValidateBasic() sdk.Error {
	if msg.PublicKey == nil || msg.PublicKey.RawString() == "" {
		return ErrNilValidatorAddr(DefaultCodespace)
	}
	if msg.Value.LTE(sdk.ZeroInt()) {
		return ErrBadDelegationAmount(DefaultCodespace)
	}
	if len(msg.Chains) == 0 {
		return ErrNoChains(DefaultCodespace)
	}
	for _, chain := range msg.Chains {
		err := ValidateNetworkIdentifier(chain)
		if err != nil {
			return err
		}
	}
	if len(msg.GeoZone) == 0 {
		return ErrNoGeoZone(DefaultCodespace)
	}
	for _, geozone := range msg.GeoZone {
		err := ValidateGeoZone(geozone)
		if err != nil {
			return err
		}
	}
	if err := ValidateServiceURL(msg.ServiceUrl); err != nil {
		return err
	}
	return nil
}

// Route provides router key for msg
func (msg MsgStake) Route() string { return RouterKey }

// Type provides msg name
func (msg MsgStake) Type() string { return MsgStakeName }

// GetFee get fee for msg
func (msg MsgStake) GetFee() sdk.BigInt {
	return sdk.NewInt(NodeFeeMap[msg.Type()])
}
func (msg *MsgStake) Reset() {
	*msg = MsgStake{}
}

func (msg *MsgStake) XXX_MessageName() string {
	m := msg.ToProto()
	return m.XXX_MessageName()
}

func (msg MsgStake) String() string {
	return fmt.Sprintf("Public Key: %s\nChains: %s\nValue: %s\nOutputAddress: %s\n", msg.PublicKey.RawString(), msg.Chains, msg.Value.String(), msg.Output)
}

func (msg *MsgStake) ProtoMessage() {
	m := msg.ToProto()
	m.ProtoMessage()
	return
}

// GetFee get fee for msg
func (msg MsgStake) ToProto() MsgProtoStake {
	pubKeyBz := make([]byte, 0)
	if msg.PublicKey != nil {
		pubKeyBz = msg.PublicKey.RawBytes()
	}
	return MsgProtoStake{
		Publickey:     pubKeyBz,
		Chains:        msg.Chains,
		Value:         msg.Value,
		ServiceUrl:    msg.ServiceUrl,
		GeoZone:       msg.GeoZone,
		OutputAddress: msg.Output,
	}
}

func (msg MsgStake) CheckServiceUrlLength(url string) sdk.Error {
	if len(url) > 255 {
		return ErrInvalidServiceURL(DefaultCodespace, fmt.Errorf("url too long"))
	}
	return nil
}

func (*MsgProtoStake) XXX_MessageName() string {
	return "x.servicers.MsgProtoStake"
}

func (*MsgBeginUnstake) XXX_MessageName() string {
	return "x.servicers.MsgBeginUnstake"
}

func (*MsgUnjail) XXX_MessageName() string {
	return "x.servicers.MsgUnjail"
}

func (*MsgPause) XXX_MessageName() string {
	return "x.servicers.MsgPause"
}

func (*MsgUnpause) XXX_MessageName() string {
	return "x.servicers.MsgUnpause"
}

//----------------------------------------------------------------------------------------------------------------------

// GetSigners return address(es) that must sign over msg.GetSignBytes()
func (msg MsgPause) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Signer, msg.ValidatorAddr}
}

func (msg MsgPause) GetRecipient() sdk.Address {
	return nil
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgPause) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic quick validity check, stateless
func (msg MsgPause) ValidateBasic() sdk.Error {
	if msg.ValidatorAddr.Empty() {
		return ErrNoValidatorFound(DefaultCodespace)
	}
	if msg.Signer.Empty() {
		return ErrNilSignerAddr(DefaultCodespace)
	}
	return nil
}

// Route provides router key for msg
func (msg MsgPause) Route() string { return RouterKey }

// Type provides msg name
func (msg MsgPause) Type() string { return MsgPauseName }

// GetFee get fee for msg
func (msg MsgPause) GetFee() sdk.BigInt {
	return sdk.NewInt(NodeFeeMap[msg.Type()])
}

//----------------------------------------------------------------------------------------------------------------------

// GetSigners return address(es) that must sign over msg.GetSignBytes()
func (msg MsgUnpause) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Signer, msg.ValidatorAddr}
}

func (msg MsgUnpause) GetRecipient() sdk.Address {
	return nil
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgUnpause) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic quick validity check, stateless
func (msg MsgUnpause) ValidateBasic() sdk.Error {
	if msg.ValidatorAddr.Empty() {
		return ErrNoValidatorFound(DefaultCodespace)
	}
	if msg.Signer.Empty() {
		return ErrNilSignerAddr(DefaultCodespace)
	}
	return nil
}

// Route provides router key for msg
func (msg MsgUnpause) Route() string { return RouterKey }

// Type provides msg name
func (msg MsgUnpause) Type() string { return MsgUnpauseName }

// GetFee get fee for msg
func (msg MsgUnpause) GetFee() sdk.BigInt {
	return sdk.NewInt(NodeFeeMap[msg.Type()])
}
