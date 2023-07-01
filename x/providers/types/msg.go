package types

import (
	"fmt"

	"github.com/vipernet-xyz/viper-network/codec"
	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
)

// ensure ProtoMsg interface compliance at compile time
var (
	_ sdk.ProtoMsg         = &MsgStake{}
	_ codec.ProtoMarshaler = &MsgStake{}
	_ sdk.ProtoMsg         = &MsgBeginUnstake{}
	_ sdk.ProtoMsg         = &MsgUnjail{}
	_ sdk.ProtoMsg         = &MsgStakingKey{}
)

const (
	MsgProviderStakeName   = "provider_stake"
	MsgProviderUnstakeName = "provider_begin_unstake"
	MsgProviderUnjailName  = "provider_unjail"
	MsgStakingKeyName      = "stake_key"
)

type MsgStake struct {
	PubKey  crypto.PublicKey `json:"pubkey" yaml:"pubkey"`
	Chains  []string         `json:"chains" yaml:"chains"`
	Value   sdk.BigInt       `json:"value" yaml:"value"`
	GeoZone []string         `json:"geo_zone" yaml:"geo_zone"`
}

// GetSigners return address(es) that must sign over msg.GetSignBytes()
func (msg MsgStake) GetSigners() []sdk.Address {
	return []sdk.Address{sdk.Address(msg.PubKey.Address())}
}

func (msg MsgStake) GetRecipient() sdk.Address {
	return nil
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgStake) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic quick validity check for staking an provider
func (msg MsgStake) ValidateBasic() sdk.Error {
	if msg.PubKey == nil || msg.PubKey.RawString() == "" {
		return ErrNilProviderAddr(DefaultCodespace)
	}
	if msg.Value.LTE(sdk.ZeroInt()) {
		return ErrBadStakeAmount(DefaultCodespace)
	}
	if len(msg.Chains) == 0 {
		return ErrNoChains(DefaultCodespace)
	}
	for _, chain := range msg.Chains {
		if err := ValidateNetworkIdentifier(chain); err != nil {
			return err
		}
	}
	return nil
}

// Route provides router key for msg
func (msg MsgStake) Route() string { return RouterKey }

// Type provides msg name
func (msg MsgStake) Type() string { return MsgProviderStakeName }

// GetFee get fee for msg
func (msg MsgStake) GetFee() sdk.BigInt {
	return sdk.NewInt(ProviderFeeMap[msg.Type()])
}

func (msg *MsgStake) Marshal() ([]byte, error) {
	m := msg.ToProto()
	return m.Marshal()
}

func (msg *MsgStake) MarshalTo(data []byte) (n int, err error) {
	m := msg.ToProto()
	return m.MarshalTo(data)
}

func (msg *MsgStake) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	m := msg.ToProto()
	return m.MarshalToSizedBuffer(dAtA)
}

func (msg *MsgStake) Size() int {
	m := msg.ToProto()
	return m.Size()
}

func (msg *MsgStake) XXX_MessageName() string {
	p := msg.ToProto()
	return p.XXX_MessageName()
}

func (msg *MsgStake) Unmarshal(data []byte) error {
	var m MsgProtoStake
	err := m.Unmarshal(data)
	if err != nil {
		return err
	}
	pk, err := crypto.NewPublicKeyBz(m.PubKey)
	if err != nil {
		return err
	}
	*msg = MsgStake{
		PubKey: pk,
		Chains: m.Chains,
		Value:  m.Value,
	}
	return nil
}

func (msg *MsgStake) Reset() {
	*msg = MsgStake{}
}

func (msg MsgStake) String() string {
	return fmt.Sprintf("Public Key: %s\nChains: %s\nValue: %s\n", msg.PubKey.RawString(), msg.Chains, msg.Value.String())
}

func (msg MsgStake) ProtoMessage() {
	m := msg.ToProto()
	m.ProtoMessage()
}

func (msg MsgStake) ToProto() MsgProtoStake {
	var pkbz []byte
	if msg.PubKey != nil {
		pkbz = msg.PubKey.RawBytes()
	}
	return MsgProtoStake{
		PubKey: pkbz,
		Chains: msg.Chains,
		Value:  msg.Value,
	}
}

//----------------------------------------------------------------------------------------------------------------------

// GetSigners address(es) that must sign over msg.GetSignBytes()
func (msg MsgBeginUnstake) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Address}
}

func (msg MsgBeginUnstake) GetRecipient() sdk.Address {
	return nil
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgBeginUnstake) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic quick validity check for staking an provider
func (msg MsgBeginUnstake) ValidateBasic() sdk.Error {
	if msg.Address.Empty() {
		return ErrNilProviderAddr(DefaultCodespace)
	}
	return nil
}

// Route provides router key for msg
func (msg MsgBeginUnstake) Route() string { return RouterKey }

// Type provides msg name
func (msg MsgBeginUnstake) Type() string { return MsgProviderUnstakeName }

// GetFee get fee for msg
func (msg MsgBeginUnstake) GetFee() sdk.BigInt {
	return sdk.NewInt(ProviderFeeMap[msg.Type()])
}

// ----------------------------------------------------------------------------------------------------------------------
// Route provides router key for msg
func (msg MsgUnjail) Route() string { return RouterKey }

// Type provides msg name
func (msg MsgUnjail) Type() string { return MsgProviderUnjailName }

// GetFee get fee for msg
func (msg MsgUnjail) GetFee() sdk.BigInt {
	return sdk.NewInt(ProviderFeeMap[msg.Type()])
}

// GetSigners return address(es) that must sign over msg.GetSignBytes()
func (msg MsgUnjail) GetSigners() []sdk.Address {
	return []sdk.Address{msg.ProviderAddr}
}

func (msg MsgUnjail) GetRecipient() sdk.Address {
	return nil
}

// GetSignBytes returns the message bytes to sign over.
func (msg MsgUnjail) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic quick validity check for staking an provider
func (msg MsgUnjail) ValidateBasic() sdk.Error {
	if msg.ProviderAddr.Empty() {
		return ErrBadProviderAddr(DefaultCodespace)
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
	Address    sdk.Address      `json:"address" yaml:"address"`
	StakingKey crypto.PublicKey `json:"pubkey" yaml:"pubkey"`
}

// Route provides router key for msg
func (msg MsgStakingKey) Route() string { return RouterKey }

// Type provides msg name
func (msg MsgStakingKey) Type() string { return MsgStakingKeyName }

// GetFee get fee for msg
func (msg MsgStakingKey) GetFee() sdk.BigInt {
	return sdk.NewInt(ProviderFeeMap[msg.Type()])
}

// GetSigners return address(es) that must sign over msg.GetSignBytes()
func (msg MsgStakingKey) GetSigners() []sdk.Address {
	return []sdk.Address{msg.Address}
}

// GetSigners return address(es) that must sign over msg.GetSignBytes()
func (msg MsgStakingKey) GetRecipient() sdk.Address {
	return nil
}

func (msg *MsgStakingKey) Reset() {
	*msg = MsgStakingKey{}
}

func (msg MsgStakingKey) String() string {
	return fmt.Sprintf("Public Key: %s\nAddress: %s\n", msg.StakingKey.String(), msg.Address)
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
		pkbz = []byte(msg.StakingKey.String())
	}
	return MsgProtoStakingKey{
		Address: msg.Address,
		PubKey:  pkbz,
	}
}

// ValidateBasic quick validity check
func (msg MsgStakingKey) ValidateBasic() sdk.Error {
	if msg.Address == nil {
		return sdk.ErrInvalidAddress("nil from address")
	}

	if msg.Address == nil {
		return sdk.ErrInvalidAddress("nil to address")
	}
	return nil
}
