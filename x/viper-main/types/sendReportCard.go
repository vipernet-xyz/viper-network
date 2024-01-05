package types

import (
	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/authentication"
	"github.com/vipernet-xyz/viper-network/x/authentication/util"
)

func SendReportCardTx(pk crypto.PrivateKey, cliCtx util.CLIContext, txBuilder authentication.TxBuilder, header SessionHeader, servicerAddr sdk.Address, reportCard ViperQoSReport) (*sdk.TxResponse, error) {
	msg := MsgSubmitReportCard{
		SessionHeader:    header,
		ServicerAddress:  servicerAddr,
		FishermanAddress: sdk.Address(pk.PublicKey().Address()),
		Report:           reportCard,
	}
	err := msg.ValidateBasic()
	if err != nil {
		return nil, err
	}
	var legacyCodec bool

	legacyCodec = false
	return util.CompleteAndBroadcastTxCLI(txBuilder, cliCtx, &msg, legacyCodec)
}
