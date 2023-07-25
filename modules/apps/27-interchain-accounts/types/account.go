package types

import (
	"encoding/json"
	"regexp"
	"strings"

	errorsmod "cosmossdk.io/errors"
	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	sdkaddress "github.com/vipernet-xyz/viper-network/types/address"
	authexported "github.com/vipernet-xyz/viper-network/x/authentication/exported"
	authtypes "github.com/vipernet-xyz/viper-network/x/authentication/types"
	yaml "gopkg.in/yaml.v2"
)

var (
	_ authtypes.GenesisAccount = (*InterchainAccount)(nil)
	_ InterchainAccountI       = (*InterchainAccount)(nil)
)

// DefaultMaxAddrLength defines the default maximum character length used in validation of addresses
var DefaultMaxAddrLength = 128

// isValidAddr defines a regular expression to check if the provided string consists of
// strictly alphanumeric characters and is non empty.
var isValidAddr = regexp.MustCompile("^[a-zA-Z0-9]+$").MatchString

type InterchainAccountI interface {
	authexported.Account
}

// interchainAccountPretty defines an unexported struct used for encoding the InterchainAccount details
type interchainAccountPretty struct {
	Address       sdk.Address `json:"address" yaml:"address"`
	PubKey        string      `json:"public_key" yaml:"public_key"`
	AccountNumber uint64      `json:"account_number" yaml:"account_number"`
	Sequence      uint64      `json:"sequence" yaml:"sequence"`
	AccountOwner  string      `json:"account_owner" yaml:"account_owner"`
}

// GenerateAddress returns an sdk.AccAddress derived using a host module account address, host connection ID, the controller portID,
// the current block app hash, and the current block data hash. The sdk.AccAddress returned is a sub-address of the host module account.
func GenerateAddress(ctx sdk.Ctx, connectionID, portID string) sdk.Address {
	hostModuleAcc := sdkaddress.Module(ModuleName, []byte(hostAccountsKey))
	header := ctx.BlockHeader()

	buf := []byte(connectionID + portID)
	buf = append(buf, header.AppHash...)
	buf = append(buf, header.DataHash...)

	return sdkaddress.Derive(hostModuleAcc, buf)
}

// ValidateAccountAddress performs basic validation of interchain account addresses, enforcing constraints
// on address length and character set
func ValidateAccountAddress(addr string) error {
	if !isValidAddr(addr) || len(addr) > DefaultMaxAddrLength {
		return errorsmod.Wrapf(
			ErrInvalidAccountAddress,
			"address must contain strictly alphanumeric characters, not exceeding %d characters in length",
			DefaultMaxAddrLength,
		)
	}

	return nil
}

// NewInterchainAccount creates and returns a new InterchainAccount type
func NewInterchainAccount(ba *authtypes.BaseAccount, accountOwner string) *InterchainAccount {
	return &InterchainAccount{
		BaseAccount:  ba,
		AccountOwner: accountOwner,
	}
}

func (ia InterchainAccount) SetPubKey(pubKey crypto.PublicKey) error {
	return errorsmod.Wrap(ErrUnsupported, "cannot set public key for interchain account")
}

func (ia InterchainAccount) SetSequence(seq uint64) error {
	return errorsmod.Wrap(ErrUnsupported, "cannot set sequence number for interchain account")
}

// Validate implements basic validation of the InterchainAccount
func (ia InterchainAccount) Validate() error {
	if strings.TrimSpace(ia.AccountOwner) == "" {
		return errorsmod.Wrap(ErrInvalidAccountAddress, "AccountOwner cannot be empty")
	}

	return ia.BaseAccount.Validate()
}

// String returns a string representation of the InterchainAccount
func (ia InterchainAccount) String() string {
	out, _ := ia.MarshalYAML()
	return string(out)
}

// MarshalYAML returns the YAML representation of the InterchainAccount
func (ia InterchainAccount) MarshalYAML() ([]byte, error) {
	accAddr, err := sdk.AccAddressFromBech32(string(ia.Address))
	if err != nil {
		return nil, err
	}

	bz, err := yaml.Marshal(interchainAccountPretty{
		Address:      accAddr,
		PubKey:       "",
		AccountOwner: ia.AccountOwner,
	})
	if err != nil {
		return nil, err
	}

	return bz, nil
}

// MarshalJSON returns the JSON representation of the InterchainAccount
func (ia InterchainAccount) MarshalJSON() ([]byte, error) {
	Addr, err := sdk.AccAddressFromBech32(string(ia.Address))
	if err != nil {
		return nil, err
	}

	bz, err := json.Marshal(interchainAccountPretty{
		Address:      Addr,
		PubKey:       "",
		AccountOwner: ia.AccountOwner,
	})
	if err != nil {
		return nil, err
	}

	return bz, nil
}

func (ia *InterchainAccount) UnmarshalJSON(bz []byte) error {
	var alias interchainAccountPretty
	if err := json.Unmarshal(bz, &alias); err != nil {
		return err
	}

	baseAccount := authtypes.NewBaseAccountWithAddress(alias.Address)
	ia.BaseAccount = &baseAccount // Take the address of the value
	ia.AccountOwner = alias.AccountOwner

	return nil
}
