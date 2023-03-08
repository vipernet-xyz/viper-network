package codec

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/vipernet-xyz/viper-network/codec/types"
	"google.golang.org/grpc/encoding"

	"github.com/gogo/protobuf/proto"
	tmTypes "github.com/tendermint/tendermint/types"
)

type Codec struct {
	protoCdc        *ProtoCodec
	legacyCdc       *LegacyAmino
	upgradeOverride int
}

func NewCodec(anyUnpacker types.AnyUnpacker) *Codec {
	return &Codec{
		protoCdc:        NewProtoCodec(anyUnpacker),
		legacyCdc:       NewLegacyAminoCodec(),
		upgradeOverride: -1,
	}
}

var (
	UpgradeFeatureMap                      = make(map[string]int64)
	UpgradeHeight                    int64 = math.MaxInt64
	OldUpgradeHeight                 int64 = 0
	NotProtoCompatibleInterfaceError       = errors.New("the interface passed for encoding does not implement proto marshaller")
	TestMode                         int64 = 0
)

const (
	UpgradeCodecHeight      = int64(30024)
	CodecChainHaltHeight    = int64(30334)
	ValidatorSplitHeight    = int64(45353)
	UpgradeCodecUpdateKey   = "CODEC"
	ValidatorSplitUpdateKey = "SPLIT"
	NonCustodialUpdateKey   = "NCUST"
	TxCacheEnhancementKey   = "REDUP"
	MaxRelayProtKey         = "MREL"
	ReplayBurnKey           = "REPBR"
	BlockSizeModifyKey      = "BLOCK"
	RSCALKey                = "RSCAL"
	VEDITKey                = "VEDIT"
)

func GetCodecUpgradeHeight() int64 {
	if UpgradeHeight >= UpgradeCodecHeight {
		return UpgradeCodecHeight
	} else {
		if OldUpgradeHeight != 0 && OldUpgradeHeight < UpgradeHeight {
			return OldUpgradeHeight
		} else {
			return UpgradeHeight
		}
	}
}

func (cdc *Codec) RegisterStructure(o interface{}, name string) {
	cdc.legacyCdc.RegisterConcrete(o, name, nil)
}

func (cdc *Codec) SetUpgradeOverride(b bool) {
	if b {
		cdc.upgradeOverride = 1
	} else {
		cdc.upgradeOverride = 0
	}
}

func (cdc *Codec) DisableUpgradeOverride() {
	cdc.upgradeOverride = -1
}

func (cdc *Codec) RegisterInterface(name string, iface interface{}, impls ...proto.Message) {
	res, ok := cdc.protoCdc.anyUnpacker.(types.InterfaceRegistry)
	if !ok {
		panic("unable to convert protocodec.anyUnpacker into types.InterfaceRegistry")
	}
	res.RegisterInterface(name, iface, impls...)
	cdc.legacyCdc.Amino.RegisterInterface(iface, nil)
}

func (cdc *Codec) RegisterImplementation(iface interface{}, impls ...proto.Message) {
	res, ok := cdc.protoCdc.anyUnpacker.(types.InterfaceRegistry)
	if !ok {
		panic("unable to convert protocodec.anyUnpacker into types.InterfaceRegistry")
	}
	res.RegisterImplementations(iface, impls...)
}

func (cdc *Codec) MarshalBinaryBare(o interface{}, height int64) ([]byte, error) { // TODO take height as parameter, move upgrade height to this package, switch based on height not upgrade mod
	p, ok := o.(ProtoMarshaler)
	if !ok {
		if cdc.IsAfterCodecUpgrade(height) {
			return nil, NotProtoCompatibleInterfaceError
		}
		return cdc.legacyCdc.MarshalBinaryBare(o)
	} else {
		if cdc.IsAfterCodecUpgrade(height) {
			return cdc.protoCdc.MarshalBinaryBare(p)
		}
		return cdc.legacyCdc.MarshalBinaryBare(p)
	}
}

func (cdc *Codec) MarshalBinaryLengthPrefixed(o interface{}, height int64) ([]byte, error) {
	p, ok := o.(ProtoMarshaler)
	if !ok {
		if cdc.IsAfterCodecUpgrade(height) {
			return nil, NotProtoCompatibleInterfaceError
		}
		return cdc.legacyCdc.MarshalBinaryLengthPrefixed(o)
	} else {
		if cdc.IsAfterCodecUpgrade(height) {
			return cdc.protoCdc.MarshalBinaryLengthPrefixed(p)
		}
		return cdc.legacyCdc.MarshalBinaryLengthPrefixed(p)
	}
}

func (cdc *Codec) UnmarshalBinaryBare(bz []byte, ptr interface{}, height int64) error {
	p, ok := ptr.(ProtoMarshaler)
	if !ok {
		if cdc.IsAfterCodecUpgrade(height) {
			return NotProtoCompatibleInterfaceError
		}
		return cdc.legacyCdc.UnmarshalBinaryBare(bz, ptr)
	}
	if cdc.IsAfterCodecUpgrade(height) {
		if height == UpgradeCodecHeight {
			e := cdc.legacyCdc.UnmarshalBinaryBare(bz, ptr)
			if e != nil {
				return cdc.protoCdc.UnmarshalBinaryBare(bz, p)
			}
			return e
		}
		return cdc.protoCdc.UnmarshalBinaryBare(bz, p)
	}
	e := cdc.legacyCdc.UnmarshalBinaryBare(bz, ptr)
	if e != nil {
		return cdc.protoCdc.UnmarshalBinaryBare(bz, p)
	}
	return e
}

func (cdc *Codec) UnmarshalBinaryLengthPrefixed(bz []byte, ptr interface{}, height int64) error {
	p, ok := ptr.(ProtoMarshaler)
	if !ok {
		if cdc.IsAfterCodecUpgrade(height) {
			return NotProtoCompatibleInterfaceError
		}
		return cdc.legacyCdc.UnmarshalBinaryLengthPrefixed(bz, ptr)
	}
	if cdc.IsAfterCodecUpgrade(height) {
		if height == UpgradeCodecHeight {
			e := cdc.legacyCdc.UnmarshalBinaryLengthPrefixed(bz, ptr)
			if e != nil {
				return cdc.protoCdc.UnmarshalBinaryLengthPrefixed(bz, p)
			}
			return e
		}
		return cdc.protoCdc.UnmarshalBinaryLengthPrefixed(bz, p)
	}
	e := cdc.legacyCdc.UnmarshalBinaryLengthPrefixed(bz, ptr)
	if e != nil {
		return cdc.protoCdc.UnmarshalBinaryLengthPrefixed(bz, p)
	}
	return e
}

func (cdc *Codec) ProtoMarshalBinaryBare(o ProtoMarshaler) ([]byte, error) {
	return cdc.protoCdc.MarshalBinaryBare(o)
}

func (cdc *Codec) LegacyMarshalBinaryBare(o interface{}) ([]byte, error) {
	return cdc.legacyCdc.MarshalBinaryBare(o)
}

func (cdc *Codec) ProtoUnmarshalBinaryBare(bz []byte, ptr ProtoMarshaler) error {
	return cdc.protoCdc.UnmarshalBinaryBare(bz, ptr)
}

func (cdc *Codec) LegacyUnmarshalBinaryBare(bz []byte, ptr interface{}) error {
	return cdc.legacyCdc.UnmarshalBinaryBare(bz, ptr)
}

func (cdc *Codec) ProtoMarshalBinaryLengthPrefixed(o ProtoMarshaler) ([]byte, error) {
	return cdc.protoCdc.MarshalBinaryLengthPrefixed(o)
}

func (cdc *Codec) LegacyMarshalBinaryLengthPrefixed(o interface{}) ([]byte, error) {
	return cdc.legacyCdc.MarshalBinaryLengthPrefixed(o)
}

func (cdc *Codec) ProtoUnmarshalBinaryLengthPrefixed(bz []byte, ptr ProtoMarshaler) error {
	return cdc.protoCdc.UnmarshalBinaryLengthPrefixed(bz, ptr)
}

func (cdc *Codec) LegacyUnmarshalBinaryLengthPrefixed(bz []byte, ptr interface{}) error {
	return cdc.legacyCdc.UnmarshalBinaryLengthPrefixed(bz, ptr)
}

func (cdc *Codec) MarshalJSONIndent(o interface{}, prefix string, indent string) ([]byte, error) {
	return cdc.legacyCdc.MarshalJSONIndent(o, prefix, indent)
}

func (cdc *Codec) MarshalJSON(o interface{}) ([]byte, error) {
	return cdc.legacyCdc.MarshalJSON(o)
}

func (cdc *Codec) UnmarshalJSON(bz []byte, o interface{}) error {
	return cdc.legacyCdc.UnmarshalJSON(bz, o)
}

func (cdc *Codec) MustMarshalJSON(o interface{}) []byte {
	bz, err := cdc.MarshalJSON(o)
	if err != nil {
		panic(err)
	}
	return bz
}

func (cdc *Codec) MustUnmarshalJSON(bz []byte, ptr interface{}) {
	err := cdc.UnmarshalJSON(bz, ptr)
	if err != nil {
		panic(err)
	}
}

func RegisterEvidences(legacy *LegacyAmino, _ *ProtoCodec) {
	tmTypes.RegisterEvidences(legacy.Amino)
}

func (cdc *Codec) AminoCodec() *LegacyAmino {
	return cdc.legacyCdc
}

func (cdc *Codec) ProtoCodec() *ProtoCodec {
	return cdc.protoCdc
}

// IsAfterCodecUpgrade Note: includes the actual upgrade height
func (cdc *Codec) IsAfterCodecUpgrade(height int64) bool {
	if cdc.upgradeOverride != -1 {
		return cdc.upgradeOverride == 1
	}
	return (GetCodecUpgradeHeight() <= height || height == -1) || TestMode <= -1
}

// IsAfterValidatorSplitUpgrade Note: includes the actual upgrade height
func (cdc *Codec) IsAfterValidatorSplitUpgrade(height int64) bool {
	return height >= ValidatorSplitHeight || (height >= UpgradeHeight && UpgradeHeight > GetCodecUpgradeHeight()) || TestMode <= -2
}

// IsAfterNonCustodialUpgrade Note: includes the actual upgrade height
func (cdc *Codec) IsAfterNonCustodialUpgrade(height int64) bool {
	return (UpgradeFeatureMap[NonCustodialUpdateKey] != 0 && height >= UpgradeFeatureMap[NonCustodialUpdateKey]) || TestMode <= -3
}

// IsOnNonCustodialUpgrade Note: includes the actual upgrade height
func (cdc *Codec) IsOnNonCustodialUpgrade(height int64) bool {
	return (UpgradeFeatureMap[NonCustodialUpdateKey] != 0 && height == UpgradeFeatureMap[NonCustodialUpdateKey]) || TestMode <= -3
}

// IsAfterNamedFeatureActivationHeight Note: includes the actual upgrade height
func (cdc *Codec) IsAfterNamedFeatureActivationHeight(height int64, key string) bool {
	return UpgradeFeatureMap[key] != 0 && height >= UpgradeFeatureMap[key]
}

// IsOnNamedFeatureActivationHeight Note: includes the actual upgrade height
func (cdc *Codec) IsOnNamedFeatureActivationHeight(height int64, key string) bool {
	return UpgradeFeatureMap[key] != 0 && height == UpgradeFeatureMap[key]
}

// Upgrade Utils for feature map

// SliceToExistingMap merge slice to existing map
func SliceToExistingMap(arr []string, m map[string]int64) map[string]int64 {
	var fmap = make(map[string]int64)
	for k, v := range m {
		fmap[k] = v
	}
	for _, v := range arr {
		kv := strings.Split(v, ":")
		i, _ := strconv.ParseInt(kv[1], 10, 64)
		fmap[kv[0]] = i
	}
	return fmap
}

// SliceToMap converts slice to map
func SliceToMap(arr []string) map[string]int64 {
	var fmap = make(map[string]int64)
	for _, v := range arr {
		kv := strings.Split(v, ":")
		i, _ := strconv.ParseInt(kv[1], 10, 64)
		fmap[kv[0]] = i
	}
	return fmap
}

// MapToSlice converts map to slice
func MapToSlice(m map[string]int64) []string {
	var fslice = make([]string, 0)
	for k, v := range m {
		kv := fmt.Sprintf("%s:%d", k, v)
		fslice = append(fslice, kv)
	}
	return fslice
}

// CleanUpgradeFeatureSlice convert slice to map and back to remove duplicates
func CleanUpgradeFeatureSlice(arr []string) []string {
	m := SliceToMap(arr)
	s := MapToSlice(m)
	sort.Strings(s)
	return s
}

type (
	BinaryCodec interface {
		// Marshal returns binary encoding of v.
		Marshal(o ProtoMarshaler) ([]byte, error)
		// MustMarshal calls Marshal and panics if error is returned.
		MustMarshal(o ProtoMarshaler) []byte

		// MarshalLengthPrefixed returns binary encoding of v with bytes length prefix.
		MarshalLengthPrefixed(o ProtoMarshaler) ([]byte, error)
		// MustMarshalLengthPrefixed calls MarshalLengthPrefixed and panics if
		// error is returned.
		MustMarshalLengthPrefixed(o ProtoMarshaler) []byte

		// Unmarshal parses the data encoded with Marshal method and stores the result
		// in the value pointed to by v.
		Unmarshal(bz []byte, ptr ProtoMarshaler) error
		// MustUnmarshal calls Unmarshal and panics if error is returned.
		MustUnmarshal(bz []byte, ptr ProtoMarshaler)

		// Unmarshal parses the data encoded with UnmarshalLengthPrefixed method and stores
		// the result in the value pointed to by v.
		UnmarshalLengthPrefixed(bz []byte, ptr ProtoMarshaler) error
		// MustUnmarshalLengthPrefixed calls UnmarshalLengthPrefixed and panics if error
		// is returned.
		MustUnmarshalLengthPrefixed(bz []byte, ptr ProtoMarshaler)

		// MarshalInterface is a helper method which will wrap `i` into `Any` for correct
		// binary interface (de)serialization.
		MarshalInterface(i proto.Message) ([]byte, error)
		// UnmarshalInterface is a helper method which will parse binary enoded data
		// into `Any` and unpack any into the `ptr`. It fails if the target interface type
		// is not registered in codec, or is not compatible with the serialized data
		UnmarshalInterface(bz []byte, ptr interface{}) error

		types.AnyUnpacker
	}

	JSONCodec interface {
		// MarshalJSON returns JSON encoding of v.
		MarshalJSON(o proto.Message) ([]byte, error)
		// MustMarshalJSON calls MarshalJSON and panics if error is returned.
		MustMarshalJSON(o proto.Message) []byte
		// MarshalInterfaceJSON is a helper method which will wrap `i` into `Any` for correct
		// JSON interface (de)serialization.
		MarshalInterfaceJSON(i proto.Message) ([]byte, error)
		// UnmarshalInterfaceJSON is a helper method which will parse JSON enoded data
		// into `Any` and unpack any into the `ptr`. It fails if the target interface type
		// is not registered in codec, or is not compatible with the serialized data
		UnmarshalInterfaceJSON(bz []byte, ptr interface{}) error

		// UnmarshalJSON parses the data encoded with MarshalJSON method and stores the result
		// in the value pointed to by v.
		UnmarshalJSON(bz []byte, ptr proto.Message) error
		// MustUnmarshalJSON calls Unmarshal and panics if error is returned.
		MustUnmarshalJSON(bz []byte, ptr proto.Message)
	}

	// AminoMarshaler defines an interface a type must implement to serialize itself
	// for Amino codec.
	AminoMarshaler interface {
		MarshalAmino() ([]byte, error)
		UnmarshalAmino([]byte) error
		MarshalAminoJSON() ([]byte, error)
		UnmarshalAminoJSON([]byte) error
	}

	// GRPCCodecProvider is implemented by the Codec
	// implementations which return a gRPC encoding.Codec.
	// And it is used to decode requests and encode responses
	// passed through gRPC.
	GRPCCodecProvider interface {
		GRPCCodec() encoding.Codec
	}
)
