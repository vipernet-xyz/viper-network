package tx

import (
	sdk "github.com/vipernet-xyz/viper-network/types"
)

// TipTx defines the interface to be implemented by Txs that handle Tips.
type TipTx interface {
	sdk.FeeTx
	GetTip() *Tip
}
