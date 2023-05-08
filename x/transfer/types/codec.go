package types

import (
	"bytes"

	"github.com/cosmos/gogoproto/proto"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/vipernet-xyz/viper-network/codec"
	"github.com/vipernet-xyz/viper-network/codec/types"

	//crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/types/msgservice"
)

// RegisterCodec registers concrete types on the codec
func RegisterCodec(cdc *codec.Codec) {

}

// RegisterLegacyAminoCodec registers the necessary x/ibc transfer interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgTransfer{}, "vipernet-xyz/MsgTransfer", nil)
}

// RegisterInterfaces register the ibc transfer module interfaces to protobuf
// Any.
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgTransfer{})

	registry.RegisterImplementations(
		(Authorization)(nil),
		&TransferAuthorization{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

// module wide codec
var ModuleCdc *codec.Codec
var amino = codec.NewLegacyAmino()
var AminoCdc = codec.NewAminoCodec(amino)

func init() {
	RegisterLegacyAminoCodec(amino)
	amino.Seal()
}

// mustProtoMarshalJSON provides an auxiliary function to return Proto3 JSON encoded
// bytes of a message.
// NOTE: Copied from https://github.com/cosmos/cosmos-sdk/blob/971c542453e0972ef1dfc5a80159ad5049c7211c/codec/json.go
// and modified in order to allow `EmitDefaults` to be set to false for ics20 packet marshalling.
// This allows for the introduction of the memo field to be backwards compatible.
func mustProtoMarshalJSON(msg proto.Message) []byte {
	anyResolver := types.NewInterfaceRegistry()

	// EmitDefaults is set to false to prevent marshalling of unpopulated fields (memo)
	// OrigName and the anyResovler match the fields the original SDK function would expect
	// in order to minimize changes.

	// OrigName is true since there is no particular reason to use camel case
	// The any resolver is empty, but provided anyways.
	jm := &jsonpb.Marshaler{OrigName: true, EmitDefaults: false, AnyResolver: anyResolver}

	err := types.UnpackInterfaces(msg, types.ProtoJSONPacker{JSONPBMarshaler: jm})
	if err != nil {
		panic(err)
	}

	buf := new(bytes.Buffer)
	if err := jm.Marshal(buf, msg); err != nil {
		panic(err)
	}

	return buf.Bytes()
}

// Authorization represents the interface of various Authorization types implemented
// by other modules.
type Authorization interface {
	proto.Message

	// MsgTypeURL returns the fully-qualified Msg service method URL (as described in ADR 031),
	// which will process and accept or reject a request.
	MsgTypeURL() string

	// Accept determines whether this grant permits the provided sdk.Msg to be performed,
	// and if so provides an upgraded authorization instance.
	Accept(ctx sdk.Context, msg sdk.Msg) (AcceptResponse, error)

	// ValidateBasic does a simple validation check that
	// doesn't require access to any other information.
	ValidateBasic() error
}

// AcceptResponse instruments the controller of an authz message if the request is accepted
// and if it should be updated or deleted.
type AcceptResponse struct {
	// If Accept=true, the controller can accept and authorization and handle the update.
	Accept bool
	// If Delete=true, the controller must delete the authorization object and release
	// storage resources.
	Delete bool
	// Controller, who is calling Authorization.Accept must check if `Updated != nil`. If yes,
	// it must use the updated version and handle the update on the storage level.
	Updated Authorization
}
