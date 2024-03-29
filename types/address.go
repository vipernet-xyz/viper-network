package types

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/hashicorp/golang-lru/simplelru"
	tmCrypto "github.com/tendermint/tendermint/crypto"
	"github.com/vipernet-xyz/viper-network/codec"
	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	"github.com/vipernet-xyz/viper-network/internal/conv"
	"github.com/vipernet-xyz/viper-network/types/bech32"
	"gopkg.in/yaml.v2"
)

const (
	AddrLen = 20
)

// Address is a common interface for different types of Addresses used by the SDK
type AddressI interface {
	Equals(Address) bool
	Empty() bool
	Marshal() ([]byte, error)
	MarshalJSON() ([]byte, error)
	Bytes() []byte
	String() string
	Format(s fmt.State, verb rune)
}

var (
	accAddrMu    sync.Mutex
	accAddrCache *simplelru.LRU
)

// Ensure that different address types implement the interface
var (
	_ AddressI             = Address{}
	_ yaml.Marshaler       = Address{}
	_ codec.ProtoMarshaler = &Address{}
	_ AddressI             = AccAddress{}
)

func VerifyAddressFormat(bz []byte) error {
	verifier := GetConfig().GetAddressVerifier()
	if verifier != nil {
		return verifier(bz)
	}
	if len(bz) != AddrLen {
		return errors.New("Incorrect address length")
	}
	return nil
}

// Address a wrapper around bytes meant to represent an address.
// When marshaled to a string or JSON.
type Address tmCrypto.Address

func (a *Address) Reset() {
	*a = Address{}
}

func (a Address) ProtoMessage() {
	p := a.ToProto()
	p.ProtoMessage()
}

func (a Address) MarshalTo(data []byte) (n int, err error) {
	p := a.ToProto()
	return p.MarshalTo(data)
}

func (a Address) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	p := a.ToProto()
	return p.MarshalToSizedBuffer(dAtA)
}

func (a Address) Size() int {
	p := a.ToProto()
	return p.Size()
}

func (a Address) Marshal() ([]byte, error) {
	p := a.ToProto()
	return p.Marshal()
}

func (a *Address) Unmarshal(data []byte) error {
	var pa ProtoAddress
	err := pa.Unmarshal(data)
	if err != nil {
		return err
	}
	*a = pa.FromProto()
	return nil
}

func (a Address) ToProto() ProtoAddress {
	return ProtoAddress{
		Address: a,
	}
}

func (pa ProtoAddress) FromProto() Address {
	return pa.Address
}

// Returns boolean for whether two Addresses are Equal
func (a Address) Equals(aa2 Address) bool {
	if a.Empty() && aa2.Empty() {
		return true
	}

	return bytes.Equal(a.Bytes(), aa2.Bytes())
}

// Returns boolean for whether an Address is empty
func (a Address) Empty() bool {
	if a == nil {
		return true
	}

	aa2 := Address{}
	return bytes.Equal(a.Bytes(), aa2.Bytes())
}

// MarshalJSON marshals to JSON.
func (a Address) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

// MarshalYAML marshals to YAML.
func (a Address) MarshalYAML() (interface{}, error) {
	return a.String(), nil
}

// UnmarshalJSON unmarshals from JSON.
func (a *Address) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	aa2, err := AddressFromHex(s)
	if err != nil {
		return err
	}

	*a = aa2
	return nil
}

// UnmarshalYAML unmarshals from JSON
func (a *Address) UnmarshalYAML(data []byte) error {
	var s string
	err := yaml.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	aa2, err := AddressFromHex(s)
	if err != nil {
		return err
	}

	*a = aa2
	return nil
}

// RawBytes returns the raw address bytes.
func (a Address) Bytes() []byte {
	return a
}

// String implements the Stringer interface.
func (a Address) String() string {
	if a.Empty() {
		return ""
	}

	str := hex.EncodeToString(a.Bytes())

	return str
}

// Format implements the fmt.Formatter interface.
// nolint: errcheck
func (a Address) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		_, _ = s.Write([]byte(a.String()))
	case 'p':
		_, _ = s.Write([]byte(fmt.Sprintf("%p", a)))
	default:
		_, _ = s.Write([]byte(fmt.Sprintf("%X", []byte(a))))
	}
}

// get Address from pubkey
func GetAddress(pubkey crypto.PublicKey) Address {
	return Address(pubkey.Address())
}

var _ codec.ProtoMarshaler = &Addresses{}

// Address a wrapper around bytes meant to represent an address.
// When marshaled to a string or JSON.
type Addresses []Address
type AccAddress []byte
type AccAddresses []AccAddress

func (a *Addresses) Reset() {
	*a = Addresses{}
}

func (a Addresses) String() string {
	var res string
	for _, arr := range a {
		res = res + arr.String() + "\n"
	}
	return res
}

func (a Addresses) ProtoMessage() {
	p := a.ToProto()
	p.ProtoMessage()
}

func (a Addresses) Marshal() ([]byte, error) {
	p := a.ToProto()
	return p.Marshal()
}

func (a Addresses) MarshalTo(data []byte) (n int, err error) {
	p := a.ToProto()
	return p.MarshalTo(data)
}

func (a Addresses) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	p := a.ToProto()
	return p.MarshalToSizedBuffer(dAtA)
}

func (a Addresses) Size() int {
	p := a.ToProto()
	return p.Size()
}

func (a *Addresses) Unmarshal(data []byte) error {
	var pa ProtoAddresses
	err := pa.Unmarshal(data)
	if err != nil {
		return err
	}
	*a = pa.FromProto()
	return nil
}

func (a Addresses) ToProto() ProtoAddresses {
	return ProtoAddresses{Arr: a}
}

func (pa ProtoAddresses) FromProto() Addresses {
	return pa.Arr
}

// AddressFromHex creates an Address from a hex string.
func AddressFromHex(address string) (addr Address, err error) {
	if len(address) == 0 {
		return Address{}, nil
	}

	bz, err := hex.DecodeString(address)
	if err != nil {
		return nil, err
	}
	err = VerifyAddressFormat(bz)
	if err != nil {
		return nil, err
	}

	return bz, nil
}

var errBech32EmptyAddress = errors.New("decoding Bech32 address failed: must provide a non empty address")

// GetFromBech32 decodes a bytestring from a Bech32 encoded string.
func GetFromBech32(bech32str, prefix string) ([]byte, error) {
	if len(bech32str) == 0 {
		return nil, errBech32EmptyAddress
	}

	hrp, bz, err := bech32.DecodeAndConvert(bech32str)
	if err != nil {
		return nil, err
	}

	if hrp != prefix {
		return nil, fmt.Errorf("invalid Bech32 prefix; expected %s, got %s", prefix, hrp)
	}

	return bz, nil
}

// AddressFromBech32 creates an AccAddress from a Bech32 string.
func AccAddressFromBech32(address string) (addr Address, err error) {
	if len(strings.TrimSpace(address)) == 0 {
		return Address{}, errors.New("empty address string is not allowed")
	}

	bech32PrefixAccAddr := GetConfig().GetBech32AccountAddrPrefix()

	bz, err := GetFromBech32(address, bech32PrefixAccAddr)
	if err != nil {
		return nil, err
	}

	err = VerifyAddressFormat(bz)
	if err != nil {
		return nil, err
	}

	return Address(bz), nil
}

// Returns boolean for whether an AccAddress is empty
func (aa AccAddress) Empty() bool {
	return len(aa) == 0
}

// String implements the Stringer interface.
func (aa AccAddress) String() string {
	if aa.Empty() {
		return ""
	}

	key := conv.UnsafeBytesToStr(aa)
	accAddrMu.Lock()
	defer accAddrMu.Unlock()
	addr, ok := accAddrCache.Get(key)
	if ok {
		return addr.(string)
	}
	return cacheBech32Addr(GetConfig().GetBech32AccountAddrPrefix(), aa, accAddrCache, key)
}

// cacheBech32Addr is not concurrency safe. Concurrent access to cache causes race condition.
func cacheBech32Addr(prefix string, addr []byte, cache *simplelru.LRU, cacheKey string) string {
	bech32Addr, err := bech32.ConvertAndEncode(prefix, addr)
	if err != nil {
		panic(err)
	}
	cache.Add(cacheKey, bech32Addr)
	return bech32Addr
}

// MustAccAddressFromBech32 calls AccAddressFromBech32 and panics on error.
func MustAddressFromBech32(address string) Address {
	addr, err := AccAddressFromBech32(address)
	if err != nil {
		panic(err)
	}

	return addr
}

// RawBytes returns the raw address bytes.
func (aa AccAddress) Bytes() []byte {
	return aa
}

// Returns boolean for whether two Addresses are Equal
func (aa AccAddress) Equals(aa2 Address) bool {
	if aa.Empty() && aa2.Empty() {
		return true
	}

	return bytes.Equal(aa.Bytes(), aa2.Bytes())
}

// Format implements the fmt.Formatter interface.
// nolint: errcheck
func (aa AccAddress) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		_, _ = s.Write([]byte(aa.String()))
	case 'p':
		_, _ = s.Write([]byte(fmt.Sprintf("%p", aa)))
	default:
		_, _ = s.Write([]byte(fmt.Sprintf("%X", []byte(aa))))
	}
}

// Marshal returns the raw address bytes. It is needed for protobuf
// compatibility.
func (aa AccAddress) Marshal() ([]byte, error) {
	return aa, nil
}

// MarshalJSON marshals to JSON using Bech32.
func (aa AccAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(aa.String())
}

// ValAddress defines a wrapper around bytes meant to present a validator's
// operator. When marshaled to a string or JSON, it uses Bech32.
type ValAddress []byte

// ValAddressFromBech32 creates a ValAddress from a Bech32 string.
func ValAddressFromBech32(address string) (addr ValAddress, err error) {
	if len(strings.TrimSpace(address)) == 0 {
		return ValAddress{}, errors.New("empty address string is not allowed")
	}

	bech32PrefixValAddr := GetConfig().GetBech32ValidatorAddrPrefix()

	bz, err := GetFromBech32(address, bech32PrefixValAddr)
	if err != nil {
		return nil, err
	}

	err = VerifyAddressFormat(bz)
	if err != nil {
		return nil, err
	}

	return ValAddress(bz), nil
}

// ConsAddress defines a wrapper around bytes meant to present a consensus node.
// When marshaled to a string or JSON, it uses Bech32.
type ConsAddress []byte
