package app

import (
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/authentication/util"
)

// SendRawTx - Deliver tx bytes to node
func (app ViperCoreApp) SendRawTx(fromAddr string, txBytes []byte) (sdk.TxResponse, error) {
	fa, err := sdk.AddressFromHex(fromAddr)
	if err != nil {
		return sdk.TxResponse{}, err
	}
	tmClient := getTMClient()
	defer func() { _ = tmClient.Stop() }()
	cliCtx := util.CLIContext{
		Codec:       cdc,
		Client:      tmClient,
		FromAddress: fa,
	}
	cliCtx.BroadcastMode = util.BroadcastSync
	return cliCtx.BroadcastTx(txBytes)
}
