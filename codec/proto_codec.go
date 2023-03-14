package codec

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"github.com/vipernet-xyz/viper-network/codec/types"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
)

// ProtoCodec defines a codec that utilizes Protobuf for both binary and JSON
// encoding.
type ProtoCodec struct {
	anyUnpacker types.AnyUnpacker
}

var _ Marshaler = &ProtoCodec{}

func NewProtoCodec(anyUnpacker types.AnyUnpacker) *ProtoCodec {
	return &ProtoCodec{anyUnpacker: anyUnpacker}
}

func (pc *ProtoCodec) Register(protoName string, iface interface{}, impls ...proto.Message) {
	res, ok := pc.anyUnpacker.(types.InterfaceRegistry)
	if !ok {
		panic("unable to convert protocodec.anyUnpacker into types.InterfaceRegistry")
	}
	res.RegisterInterface(protoName, iface, impls...)
}

func (pc *ProtoCodec) RegisterImplementation(iface interface{}, impls ...proto.Message) {
	res, ok := pc.anyUnpacker.(types.InterfaceRegistry)
	if !ok {
		panic("unable to convert protocodec.anyUnpacker into types.InterfaceRegistry")
	}
	res.RegisterImplementations(iface, impls...)
}

func (pc *ProtoCodec) MarshalBinaryBare(o ProtoMarshaler) ([]byte, error) {
	return o.Marshal()
}

func (pc *ProtoCodec) MustMarshalBinaryBare(o ProtoMarshaler) []byte {
	bz, err := pc.MarshalBinaryBare(o)
	if err != nil {
		panic(err)
	}

	return bz
}

func (pc *ProtoCodec) MarshalBinaryLengthPrefixed(o ProtoMarshaler) ([]byte, error) {
	bz, err := pc.MarshalBinaryBare(o)
	if err != nil {
		return nil, err
	}

	var sizeBuf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(sizeBuf[:], uint64(o.Size()))
	return append(sizeBuf[:n], bz...), nil
}

func (pc *ProtoCodec) MustMarshalBinaryLengthPrefixed(o ProtoMarshaler) []byte {
	bz, err := pc.MarshalBinaryLengthPrefixed(o)
	if err != nil {
		panic(err)
	}

	return bz
}

func (pc *ProtoCodec) UnmarshalBinaryBare(bz []byte, ptr ProtoMarshaler) error {
	err := ptr.Unmarshal(bz)
	if err != nil {
		return err
	}
	err = types.UnpackInterfaces(ptr, pc.anyUnpacker)
	if err != nil {
		return err
	}
	return nil
}

func (pc *ProtoCodec) MustUnmarshalBinaryBare(bz []byte, ptr ProtoMarshaler) {
	if err := pc.UnmarshalBinaryBare(bz, ptr); err != nil {
		panic(err)
	}
}

func (pc *ProtoCodec) UnmarshalBinaryLengthPrefixed(bz []byte, ptr ProtoMarshaler) error {
	size, n := binary.Uvarint(bz)
	if n < 0 {
		return fmt.Errorf("invalid number of bytes read from length-prefixed encoding: %d", n)
	}

	if size > uint64(len(bz)-n) {
		return fmt.Errorf("not enough bytes to read; want: %v, got: %v", size, len(bz)-n)
	} else if size < uint64(len(bz)-n) {
		return fmt.Errorf("too many bytes to read; want: %v, got: %v", size, len(bz)-n)
	}

	bz = bz[n:]
	return pc.UnmarshalBinaryBare(bz, ptr)
}

func (pc *ProtoCodec) MustUnmarshalBinaryLengthPrefixed(bz []byte, ptr ProtoMarshaler) {
	if err := pc.UnmarshalBinaryLengthPrefixed(bz, ptr); err != nil {
		panic(err)
	}
}

func (pc *ProtoCodec) MarshalJSON(o interface{}) ([]byte, error) {
	m, ok := o.(ProtoMarshaler)
	if !ok {
		return nil, fmt.Errorf("cannot protobuf JSON encode unsupported type: %T", o)
	}

	return ProtoMarshalJSON(m)
}

func (pc *ProtoCodec) MustMarshalJSON(o interface{}) []byte {
	bz, err := pc.MarshalJSON(o)
	if err != nil {
		panic(err)
	}

	return bz
}

func (pc *ProtoCodec) UnmarshalJSON(bz []byte, ptr interface{}) error {
	m, ok := ptr.(ProtoMarshaler)
	if !ok {
		return fmt.Errorf("cannot protobuf JSON decode unsupported type: %T", ptr)
	}

	err := jsonpb.Unmarshal(strings.NewReader(string(bz)), m)
	if err != nil {
		return err
	}

	return types.UnpackInterfaces(ptr, pc.anyUnpacker)
}

func (pc *ProtoCodec) MustUnmarshalJSON(bz []byte, ptr interface{}) {
	if err := pc.UnmarshalJSON(bz, ptr); err != nil {
		panic(err)
	}
}

func (pc *ProtoCodec) UnpackAny(any *types.Any, iface interface{}) error {
	return pc.anyUnpacker.UnpackAny(any, iface)
}

// ProtoCodec defines a codec that utilizes Protobuf for both binary and JSON
// encoding.
type ProtoCodec1 struct {
	interfaceRegistry types.InterfaceRegistry
}

// Unmarshal implements BinaryMarshaler.Unmarshal method.
// NOTE: this function must be used with a concrete type which
// implements proto.Message. For interface please use the codec.UnmarshalInterface
func (pc *ProtoCodec1) Unmarshal(bz []byte, ptr ProtoMarshaler) error {
	err := ptr.Unmarshal(bz)
	if err != nil {
		return err
	}
	err = types.UnpackInterfaces(ptr, pc.interfaceRegistry)
	if err != nil {
		return err
	}
	return nil
}
func NewProtoCodec1(interfaceRegistry types.InterfaceRegistry) *ProtoCodec1 {
	return &ProtoCodec1{interfaceRegistry: interfaceRegistry}
}

// Marshal implements BinaryMarshaler.Marshal method.
// NOTE: this function must be used with a concrete type which
// implements proto.Message. For interface please use the codec.MarshalInterface
func (pc *ProtoCodec1) Marshal(o ProtoMarshaler) ([]byte, error) {
	// Size() check can catch the typed nil value.
	if o == nil || o.Size() == 0 {
		// return empty bytes instead of nil, because nil has special meaning in places like store.Set
		return []byte{}, nil
	}
	return o.Marshal()
}

// MarshalInterface is a convenience function for proto marshalling interfaces. It packs
// the provided value, which must be an interface, in an Any and then marshals it to bytes.
// NOTE: to marshal a concrete type, you should use Marshal instead
func (pc *ProtoCodec1) MarshalInterface(i proto.Message) ([]byte, error) {
	if err := assertNotNil(i); err != nil {
		return nil, err
	}
	any, err := types.NewAnyWithValue(i)
	if err != nil {
		return nil, err
	}
	return pc.Marshal(any)
}

func assertNotNil(i interface{}) error {
	if i == nil {
		return errors.New("can't marshal <nil> value")
	}
	return nil
}

// MarshalLengthPrefixed implements BinaryMarshaler.MarshalLengthPrefixed method.
func (pc *ProtoCodec1) MarshalLengthPrefixed(o ProtoMarshaler) ([]byte, error) {
	bz, err := pc.Marshal(o)
	if err != nil {
		return nil, err
	}

	var sizeBuf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(sizeBuf[:], uint64(o.Size()))
	return append(sizeBuf[:n], bz...), nil
}

// MustMarshal implements BinaryMarshaler.MustMarshal method.
// NOTE: this function must be used with a concrete type which
// implements proto.Message. For interface please use the codec.MarshalInterface
func (pc *ProtoCodec1) MustMarshal(o ProtoMarshaler) []byte {
	bz, err := pc.Marshal(o)
	if err != nil {
		panic(err)
	}

	return bz
}

// MustMarshalLengthPrefixed implements BinaryMarshaler.MustMarshalLengthPrefixed method.
func (pc *ProtoCodec1) MustMarshalLengthPrefixed(o ProtoMarshaler) []byte {
	bz, err := pc.MarshalLengthPrefixed(o)
	if err != nil {
		panic(err)
	}

	return bz
}

// MustUnmarshal implements BinaryMarshaler.MustUnmarshal method.
// NOTE: this function must be used with a concrete type which
// implements proto.Message. For interface please use the codec.UnmarshalInterface
func (pc *ProtoCodec1) MustUnmarshal(bz []byte, ptr ProtoMarshaler) {
	if err := pc.Unmarshal(bz, ptr); err != nil {
		panic(err)
	}
}

// UnmarshalLengthPrefixed implements BinaryMarshaler.UnmarshalLengthPrefixed method.
func (pc *ProtoCodec1) UnmarshalLengthPrefixed(bz []byte, ptr ProtoMarshaler) error {
	size, n := binary.Uvarint(bz)
	if n < 0 {
		return fmt.Errorf("invalid number of bytes read from length-prefixed encoding: %d", n)
	}

	if size > uint64(len(bz)-n) {
		return fmt.Errorf("not enough bytes to read; want: %v, got: %v", size, len(bz)-n)
	} else if size < uint64(len(bz)-n) {
		return fmt.Errorf("too many bytes to read; want: %v, got: %v", size, len(bz)-n)
	}

	bz = bz[n:]
	return pc.Unmarshal(bz, ptr)
}

// MustUnmarshalLengthPrefixed implements BinaryMarshaler.MustUnmarshalLengthPrefixed method.
func (pc *ProtoCodec1) MustUnmarshalLengthPrefixed(bz []byte, ptr ProtoMarshaler) {
	if err := pc.UnmarshalLengthPrefixed(bz, ptr); err != nil {
		panic(err)
	}
}

// UnpackAny implements AnyUnpacker.UnpackAny method,
// it unpacks the value in any to the interface pointer passed in as
// iface.
func (pc *ProtoCodec1) UnpackAny(any *types.Any, iface interface{}) error {
	return pc.interfaceRegistry.UnpackAny(any, iface)
}

// UnmarshalInterface is a convenience function for proto unmarshaling interfaces. It
// unmarshals an Any from bz bytes and then unpacks it to the `ptr`, which must
// be a pointer to a non empty interface with registered implementations.
// NOTE: to unmarshal a concrete type, you should use Unmarshal instead
//
// Example:
//
//	var x MyInterface
//	err := cdc.UnmarshalInterface(bz, &x)
func (pc *ProtoCodec1) UnmarshalInterface(bz []byte, ptr interface{}) error {
	any := &types.Any{}
	err := pc.Unmarshal(bz, any)
	if err != nil {
		return err
	}

	return pc.UnpackAny(any, ptr)
}
