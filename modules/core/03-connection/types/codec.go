package types

import (
	"github.com/vipernet-xyz/viper-network/codec"
	codectypes "github.com/vipernet-xyz/viper-network/codec/types"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/types/msgservice"

	"github.com/vipernet-xyz/viper-network/modules/core/exported"
)

// RegisterInterfaces register the ibc interfaces submodule implementations to protobuf
// Any.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterInterface(
		"ibc.core.connection.v1.ConnectionI",
		(*exported.ConnectionI)(nil),
		&ConnectionEnd{},
	)
	registry.RegisterInterface(
		"ibc.core.connection.v1.CounterpartyConnectionI",
		(*exported.CounterpartyConnectionI)(nil),
		&Counterparty{},
	)
	registry.RegisterInterface(
		"ibc.core.connection.v1.Version",
		(*exported.Version)(nil),
		&Version{},
	)
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgConnectionOpenInit{},
		&MsgConnectionOpenTry{},
		&MsgConnectionOpenAck{},
		&MsgConnectionOpenConfirm{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

// SubModuleCdc references the global x/ibc/core/03-connection module codec. Note, the codec should
// ONLY be used in certain instances of tests and for JSON encoding.
//
// The actual codec used for serialization should be provided to x/ibc/core/03-connection and
// defined at the application level.
var SubModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
