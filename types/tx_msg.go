package types

import (
	gp "github.com/cosmos/gogoproto/proto"
	"github.com/golang/protobuf/proto" // nolint
)

type Msg interface {
	// Return the message type.
	// Must be alphanumeric or empty.
	Route() string

	// Returns a human-readable string for the message, intended for utilization
	// within tags
	Type() string

	// ValidateBasic does a simple validation check that
	// doesn't require access to any other information.
	ValidateBasic() Error

	// Get the canonical byte representation of the ProtoMsg.
	GetSignBytes() []byte

	// Signers returns the addrs of signers that must sign.
	// CONTRACT: All signatures must be present to be valid.
	// CONTRACT: Returns addrs in some deterministic order.
	GetSigners() []Address

	// Returns the recipient of the tx, if no recipient returns nil
	GetRecipient() Address

	// Returns an BigInt for the ProtoMsg
	GetFee() BigInt
}

// Msg defines the interface a transaction message must fulfill.
type Msg1 interface {
	proto.Message

	// ValidateBasic does a simple validation check that
	// doesn't require access to any other information.
	ValidateBasic() error

	// GetSigners returns the addrs of signers that must sign.
	// CONTRACT: All signatures must be present to be valid.
	// CONTRACT: Returns addrs in some deterministic order.
	GetSigners() []AccAddress
}

var _ Msg = ProtoMsg(nil)

// Transactions messages must fulfill the ProtoMsg
type ProtoMsg interface {
	proto.Message
	// Return the message type.
	// Must be alphanumeric or empty.
	Route() string

	// Returns a human-readable string for the message, intended for utilization
	// within tags
	Type() string

	// ValidateBasic does a simple validation check that
	// doesn't require access to any other information.
	ValidateBasic() Error

	// Get the canonical byte representation of the ProtoMsg.
	GetSignBytes() []byte

	// Signers returns the addrs of signers that must sign.
	// CONTRACT: All signatures must be present to be valid.
	// CONTRACT: Returns addrs in some deterministic order.
	GetSigners() []Address

	// Returns the recipient of the tx, if no recipient returns nil
	GetRecipient() Address

	// Returns an BigInt for the ProtoMsg
	GetFee() BigInt
}

//__________________________________________________________

// Transactions objects must fulfill the Tx
type Tx interface {
	// Gets the all the transaction's messages.
	GetMsg() Msg

	// ValidateBasic does a simple and lightweight validation check that doesn't
	// require access to any other information.
	ValidateBasic() Error
}

//__________________________________________________________

// TxDecoder unmarshals transaction bytes
type TxDecoder func(txBytes []byte, blockHeight int64) (Tx, Error)

// TxDecoder unmarshals transaction bytes
type TxDecoder1 func(txBytes []byte, blockHeight int64) (Tx1, Error)

// TxDecoder unmarshals transaction bytes
type TxDecoder2 func(txBytes []byte) (Tx1, error)

// TxEncoder marshals transaction to bytes
type TxEncoder func(tx Tx, blockHeight int64) ([]byte, error)

// TxEncoder marshals transaction to bytes
type TxEncoder2 func(tx Tx) ([]byte, error)

// TxEncoder marshals transaction to bytes
type TxEncoder1 func(tx Tx1, blockHeight int64) ([]byte, error)

// MsgTypeURL returns the TypeURL of a `sdk.Msg`.
func MsgTypeURL(msg Msg1) string {
	return "/" + gp.MessageName(msg)
}

// TxWithMemo must have GetMemo() method to use ValidateMemoDecorator
type TxWithMemo interface {
	Tx
	GetMemo() string
}

// FeeTx defines the interface to be implemented by Tx to use the FeeDecorators
type FeeTx interface {
	Tx
	GetGas() uint64
	GetFee() Coins
	FeePayer() AccAddress
	FeeGranter() AccAddress
}

// TxWithTimeoutHeight extends the Tx interface by allowing a transaction to
// set a height timeout.
type TxWithTimeoutHeight interface {
	Tx

	GetTimeoutHeight() uint64
}

type TxWithTimeoutHeight1 interface {
	Tx1

	GetTimeoutHeight() uint64
}

// Tx defines the interface a transaction must fulfill.
type Tx1 interface {
	// GetMsgs gets the all the transaction's messages.
	GetMsgs() []Msg1

	// ValidateBasic does a simple and lightweight validation check that doesn't
	// require access to any other information.
	ValidateBasic() error
}

// TxWithMemo must have GetMemo() method to use ValidateMemoDecorator
type TxWithMemo1 interface {
	Tx1
	GetMemo() string
}

// FeeTx defines the interface to be implemented by Tx to use the FeeDecorators
type FeeTx1 interface {
	Tx1
	GetGas() uint64
	GetFee() Coins
	FeePayer() AccAddress
	FeeGranter() AccAddress
}
