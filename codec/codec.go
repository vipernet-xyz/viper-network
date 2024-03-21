package codec

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/vipernet-xyz/viper-network/codec/types"

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
	TxCacheEnhancementKey      = "REDUP"
	MaxRelayProtKey            = "MREL"
	ReplayBurnKey              = "REPBR"
	BlockSizeModifyKey         = "BLOCK"
	VEDITKey                   = "VEDIT"
	ClearUnjailedValSessionKey = "CRVAL"
)

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
	res, ok := cdc.protoCdc.AnyUnpacker.(types.InterfaceRegistry)
	if !ok {
		panic("unable to convert protocodec.anyUnpacker into types.InterfaceRegistry")
	}
	res.RegisterInterface(name, iface, impls...)
	cdc.legacyCdc.Amino.RegisterInterface(iface, nil)
}

func (cdc *Codec) RegisterImplementation(iface interface{}, impls ...proto.Message) {
	res, ok := cdc.protoCdc.AnyUnpacker.(types.InterfaceRegistry)
	if !ok {
		panic("unable to convert protocodec.anyUnpacker into types.InterfaceRegistry")
	}
	res.RegisterImplementations(iface, impls...)
}

func (cdc *Codec) MarshalBinaryBare(o interface{}) ([]byte, error) {
	p, ok := o.(ProtoMarshaler)
	if !ok {
		return cdc.legacyCdc.MarshalBinaryBare(o)
	}
	res, err := cdc.protoCdc.MarshalBinaryBare(p)
	if err == nil {
		return res, err
	}
	return cdc.legacyCdc.MarshalBinaryBare(p)

}

func (cdc *Codec) MarshalBinaryLengthPrefixed(o interface{}) ([]byte, error) {
	p, ok := o.(ProtoMarshaler)
	if !ok {
		return cdc.legacyCdc.MarshalBinaryLengthPrefixed(o)
	}
	res, err := cdc.protoCdc.MarshalBinaryLengthPrefixed(p)
	if err == nil {
		return res, err
	}
	return cdc.legacyCdc.MarshalBinaryLengthPrefixed(p)
}

func (cdc *Codec) UnmarshalBinaryBare(bz []byte, ptr interface{}) error {
	p, ok := ptr.(ProtoMarshaler)
	if !ok {
		return cdc.legacyCdc.UnmarshalBinaryBare(bz, ptr)
	}
	err := cdc.protoCdc.UnmarshalBinaryBare(bz, p)
	if err != nil {
		return cdc.legacyCdc.UnmarshalBinaryBare(bz, ptr)
	}
	return nil
}

func (cdc *Codec) UnmarshalBinaryLengthPrefixed(bz []byte, ptr interface{}) error {
	p, ok := ptr.(ProtoMarshaler)
	if !ok {
		return cdc.legacyCdc.UnmarshalBinaryLengthPrefixed(bz, ptr)
	}
	err := cdc.protoCdc.UnmarshalBinaryLengthPrefixed(bz, p)
	if err != nil {
		return cdc.legacyCdc.UnmarshalBinaryLengthPrefixed(bz, ptr)
	}
	return nil
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

// Marshal implements BinaryMarshaler.Marshal method.
func (cdc *Codec) Marshal(o ProtoMarshaler) ([]byte, error) {
	return cdc.legacyCdc.Marshal(o)
}

func (cdc *Codec) MustMarshal(o ProtoMarshaler) []byte {
	return cdc.legacyCdc.MustMarshal(o)
}

func (cdc *Codec) Unmarshal(bz []byte, ptr ProtoMarshaler) error {
	return cdc.legacyCdc.Unmarshal(bz, ptr)
}

func (cdc *Codec) MustUnmarshal(bz []byte, ptr ProtoMarshaler) {
	cdc.legacyCdc.MustUnmarshal(bz, ptr)
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

// IsAfterNamedFeatureActivationHeight Note: includes the actual upgrade height
func (cdc *Codec) IsAfterNamedFeatureActivationHeight(height int64, key string) bool {
	return UpgradeFeatureMap[key] != 0 && height >= UpgradeFeatureMap[key]
}

// IsOnNamedFeatureActivationHeight Note: includes the actual upgrade height
func (cdc *Codec) IsOnNamedFeatureActivationHeight(height int64, key string) bool {
	return UpgradeFeatureMap[key] != 0 && height == UpgradeFeatureMap[key]
}

// IsOnNamedFeatureActivationHeightWithTolerance is used to enable certain
// business logic within some tolerance (i.e. only a few blocks) of feature
// activation to have more confidence in the feature's release and avoid
// non-deterministic or hard-to-predict behaviour.
func (cdc *Codec) IsOnNamedFeatureActivationHeightWithTolerance(
	height int64,
	featureKey string,
	tolerance int64,
) bool {
	upgradeHeight := UpgradeFeatureMap[featureKey]
	if upgradeHeight == 0 {
		return false
	}
	minHeight := upgradeHeight - tolerance
	maxHeight := upgradeHeight + tolerance
	return height >= minHeight && height <= maxHeight
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

func (cdc *Codec) MarshalInterface(i proto.Message) ([]byte, error) {
	if err := assertNotNil(i); err != nil {
		return nil, err
	}
	return cdc.legacyCdc.Marshal(i)
}

func (cdc *Codec) UnmarshalInterface(bz []byte, ptr interface{}) error {
	return cdc.legacyCdc.Unmarshal(bz, ptr)
}

func (cdc *Codec) UnmarshalInterfaceJSON(bz []byte, ptr interface{}) error {
	return cdc.legacyCdc.UnmarshalJSON(bz, ptr)
}
