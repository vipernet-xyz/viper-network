package client

import (
	cryptotypes "github.com/vipernet-xyz/viper-network/crypto/types"
	sdk "github.com/vipernet-xyz/viper-network/types"
)

// Account defines a read-only version of the auth module's AccountI.
type Account interface {
	GetAddress() sdk.AccAddress
	GetPubKey() cryptotypes.PubKey // can return nil.
	GetAccountNumber() uint64
	GetSequence() uint64
}

// AccountRetriever defines the interfaces required by transactions to
// ensure an account exists and to be able to query for account fields necessary
// for signing.
type AccountRetriever interface {
	GetAccount(clientCtx Context, addr sdk.Address) (Account, error)
	GetAccountWithHeight(clientCtx Context, addr sdk.Address) (Account, int64, error)
	EnsureExists(clientCtx Context, addr sdk.Address) error
	GetAccountNumberSequence(clientCtx Context, addr sdk.Address) (accNum uint64, accSeq uint64, err error)
}

var _ AccountRetriever = (*MockAccountRetriever)(nil)

// MockAccountRetriever defines a no-op basic AccountRetriever that can be used
// in mocked contexts. Tests or context that need more sophisticated testing
// state should implement their own mock AccountRetriever.
type MockAccountRetriever struct{}

func (mar MockAccountRetriever) GetAccount(_ Context, _ sdk.Address) (Account, error) {
	return nil, nil
}

func (mar MockAccountRetriever) GetAccountWithHeight(_ Context, _ sdk.Address) (Account, int64, error) {
	return nil, 0, nil
}

func (mar MockAccountRetriever) EnsureExists(_ Context, _ sdk.Address) error {
	return nil
}

func (mar MockAccountRetriever) GetAccountNumberSequence(_ Context, _ sdk.Address) (uint64, uint64, error) {
	return 0, 0, nil
}
