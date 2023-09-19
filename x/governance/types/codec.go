package types

import (
	"github.com/vipernet-xyz/viper-network/codec"
	"github.com/vipernet-xyz/viper-network/codec/types"
	sdk "github.com/vipernet-xyz/viper-network/types"
)

// module codec
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.NewCodec(types.NewInterfaceRegistry())
	RegisterCodec(ModuleCdc)
	ModuleCdc.AminoCodec().Seal()
}

// RegisterCodec registers all necessary param module types with a given codec.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterStructure(MsgChangeParam{}, "governance/msg_change_param")
	cdc.RegisterStructure(MsgDAOTransfer{}, "governance/msg_dao_transfer")
	cdc.RegisterStructure(MsgUpgrade{}, "governance/msg_upgrade")
	cdc.RegisterStructure(MsgGenerateDiscountKey{}, "governance/MsgGenerateDiscountKey")
	cdc.RegisterInterface("x.interface.nil", (*interface{})(nil))
	cdc.RegisterStructure(ACL{}, "governance/non_map_acl")
	cdc.RegisterStructure(Upgrade{}, "governance/upgrade")
	cdc.RegisterImplementation((*sdk.ProtoMsg)(nil), &MsgChangeParam{}, &MsgDAOTransfer{}, &MsgUpgrade{}, &MsgGenerateDiscountKey{})
	cdc.RegisterImplementation((*sdk.Msg)(nil), &MsgChangeParam{}, &MsgDAOTransfer{}, &MsgUpgrade{}, &MsgGenerateDiscountKey{})
	ModuleCdc = cdc
}

// Content defines an interface that a proposal must implement. It contains
// information such as the title and description along with the type and routing
// information for the appropriate handler to process the proposal. Content can
// have additional fields, which will handled by a proposal's Handler.
type Content interface {
	GetTitle() string
	GetDescription() string
	ProposalRoute() string
	ProposalType() string
	ValidateBasic() error
	String() string
}
