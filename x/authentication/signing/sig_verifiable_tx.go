package signing

import (
	"github.com/cosmos/cosmos-sdk/types"
	cryptotypes "github.com/vipernet-xyz/viper-network/crypto/types"
	"github.com/vipernet-xyz/viper-network/types/tx"
	"github.com/vipernet-xyz/viper-network/types/tx/signing"
)

// SigVerifiableTx defines a transaction interface for all signature verification
// handlers.
type SigVerifiableTx interface {
	types.Tx
	GetSigners() []types.AccAddress
	GetPubKeys() ([]cryptotypes.PubKey, error) // If signer already has pubkey in context, this list will have nil in its place
	GetSignaturesV2() ([]signing.SignatureV2, error)
}

// Tx defines a transaction interface that supports all standard message, signature
// fee, memo, tips, and auxiliary interfaces.
type Tx interface {
	SigVerifiableTx

	types.TxWithMemo
	types.FeeTx
	tx.TipTx
	types.TxWithTimeoutHeight
}
