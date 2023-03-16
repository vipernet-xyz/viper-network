package client

import (
	rpcclient "github.com/tendermint/tendermint/rpc/client"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
)

// CometRPC defines the interface of a CometBFT RPC client needed for
// queries and transaction handling.
type TendermintRPC interface {
	rpcclient.ABCIClient

	Validators(height *int64, page, perPage int) (*coretypes.ResultValidators, error)
	Status() (*coretypes.ResultStatus, error)
	Block(height *int64) (*coretypes.ResultBlock, error)
	//BlockByHash(ctx context.Context, hash []byte) (*coretypes.ResultBlock, error)
	BlockchainInfo(minHeight, maxHeight int64) (*coretypes.ResultBlockchainInfo, error)
	Commit(height *int64) (*coretypes.ResultCommit, error)
	Tx(hash []byte, prove bool) (*coretypes.ResultTx, error)
	TxSearch(
		query string,
		prove bool,
		page, perPage int,
		orderBy string,
	) (*coretypes.ResultTxSearch, error)
	/*BlockSearch(
		ctx context.Context,
		query string,
		page, perPage *int,
		orderBy string,
	) (*coretypes.ResultBlock, error)*/
}
