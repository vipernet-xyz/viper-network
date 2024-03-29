package client

import (
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/types/tx"
	signingtypes "github.com/vipernet-xyz/viper-network/types/tx/signing"
	"github.com/vipernet-xyz/viper-network/x/authentication/signing"
)

type (
	// TxEncodingConfig defines an interface that contains transaction
	// encoders and decoders
	TxEncodingConfig interface {
		TxEncoder() sdk.TxEncoder1
		TxDecoder() sdk.TxDecoder1
		TxJSONEncoder() sdk.TxEncoder1
		TxJSONDecoder() sdk.TxDecoder1
		MarshalSignatureJSON([]signingtypes.SignatureV2) ([]byte, error)
		UnmarshalSignatureJSON([]byte) ([]signingtypes.SignatureV2, error)
	}

	// TxEncodingConfig defines an interface that contains transaction
	// encoders and decoders
	TxEncodingConfig1 interface {
		TxEncoder() sdk.TxEncoder1
		TxDecoder() sdk.TxDecoder1
		TxJSONEncoder() sdk.TxEncoder1
		TxJSONDecoder() sdk.TxDecoder1
		MarshalSignatureJSON([]signingtypes.SignatureV2) ([]byte, error)
		UnmarshalSignatureJSON([]byte) ([]signingtypes.SignatureV2, error)
	}

	// TxConfig defines an interface a client can utilize to generate an
	// application-defined concrete transaction type. The type returned must
	// implement TxBuilder.
	TxConfig interface {
		TxEncodingConfig1

		NewTxBuilder() TxBuilder
		WrapTxBuilder(sdk.Tx1) (TxBuilder, error)
		SignModeHandler() signing.SignModeHandler
	}

	// TxBuilder defines an interface which an application-defined concrete transaction
	// type must implement. Namely, it must be able to set messages, generate
	// signatures, and provide canonical bytes to sign over. The transaction must
	// also know how to encode itself.
	TxBuilder interface {
		GetTx() signing.Tx1

		SetMsgs(msgs ...sdk.Msg1) error
		SetSignatures(signatures ...signingtypes.SignatureV2) error
		SetMemo(memo string)
		SetFeeAmount(amount sdk.Coins)
		SetFeePayer(feePayer sdk.Address)
		SetGasLimit(limit uint64)
		SetTip(tip *tx.Tip)
		SetTimeoutHeight(height uint64)
		SetFeeGranter(feeGranter sdk.Address)
		AddAuxSignerData(tx.AuxSignerData) error
	}
)
