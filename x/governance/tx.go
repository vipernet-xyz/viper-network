package governance

import (
	"fmt"

	"github.com/vipernet-xyz/viper-network/codec"
	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	"github.com/vipernet-xyz/viper-network/crypto/keys"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/authentication"
	"github.com/vipernet-xyz/viper-network/x/authentication/util"
	"github.com/vipernet-xyz/viper-network/x/governance/types"

	"github.com/tendermint/tendermint/rpc/client"
)

func ChangeParamsTx(cdc *codec.Codec, tmNode client.Client, keybase keys.Keybase, fromAddress sdk.Address, aclKey string, paramValue interface{}, passphrase string, fee int64, legacyCodec bool) (*sdk.TxResponse, error) {
	//valueBytes, err := json.Marshal(paramValue)
	valueBytes, err := cdc.MarshalJSON(paramValue)
	if err != nil {
		return nil, err
	}
	msg := types.MsgChangeParam{
		FromAddress: fromAddress,
		ParamKey:    aclKey,
		ParamVal:    valueBytes,
	}
	txBuilder, cliCtx := newTx(cdc, &msg, fromAddress, tmNode, keybase, passphrase, fee)
	err = msg.ValidateBasic()
	if err != nil {
		return nil, err
	}
	return util.CompleteAndBroadcastTxCLI(txBuilder, cliCtx, &msg, legacyCodec)
}

func DAOTransferTx(cdc *codec.Codec, tmNode client.Client, keybase keys.Keybase, fromAddress, toAddress sdk.Address, amount sdk.BigInt, action, passphrase string, fee int64, legacyCodec bool) (*sdk.TxResponse, error) {
	msg := types.MsgDAOTransfer{
		FromAddress: fromAddress,
		ToAddress:   toAddress,
		Amount:      amount,
		Action:      action,
	}
	txBuilder, cliCtx := newTx(cdc, &msg, fromAddress, tmNode, keybase, passphrase, fee)
	err := msg.ValidateBasic()
	if err != nil {
		return nil, err
	}
	return util.CompleteAndBroadcastTxCLI(txBuilder, cliCtx, &msg, legacyCodec)
}

func UpgradeTx(cdc *codec.Codec, tmNode client.Client, keybase keys.Keybase, fromAddress sdk.Address, upgrade types.Upgrade, passphrase string, fee int64, legacyCodec bool) (*sdk.TxResponse, error) {
	msg := types.MsgUpgrade{
		Address: fromAddress,
		Upgrade: upgrade,
	}
	txBuilder, cliCtx := newTx(cdc, &msg, fromAddress, tmNode, keybase, passphrase, fee)
	err := msg.ValidateBasic()
	if err != nil {
		return nil, err
	}
	return util.CompleteAndBroadcastTxCLI(txBuilder, cliCtx, &msg, legacyCodec)
}

func newTx(cdc *codec.Codec, msg sdk.ProtoMsg, fromAddr sdk.Address, tmNode client.Client, keybase keys.Keybase, passphrase string, fee int64) (txBuilder authentication.TxBuilder, cliCtx util.CLIContext) {
	genDoc, err := tmNode.Genesis()
	if err != nil {
		return
	}
	chainID := genDoc.Genesis.ChainID
	cliCtx = util.NewCLIContext(tmNode, fromAddr, passphrase).WithCodec(cdc)
	cliCtx.BroadcastMode = util.BroadcastSync
	account, err := cliCtx.GetAccount(fromAddr)
	if err != nil {
		return
	}
	fees := sdk.NewInt(fee)
	if account.GetCoins().AmountOf(sdk.DefaultStakeDenom).LT(fees) { // todo get stake denom
		_ = fmt.Errorf("insufficient funds: the fee needed is %v", fee)
		return
	}
	txBuilder = authentication.NewTxBuilder(
		authentication.DefaultTxEncoder(cdc),
		authentication.DefaultTxDecoder(cdc),
		chainID,
		"",
		sdk.NewCoins(sdk.NewCoin(sdk.DefaultStakeDenom, fees))).WithKeybase(keybase)
	return
}

func BuildAndSignMulti(cdc *codec.Codec, address sdk.Address, publicKey crypto.PublicKeyMultiSig, msg sdk.ProtoMsg, tmNode client.Client, keybase keys.Keybase, passphrase string, fee int64, legacyCodec bool) (txBytes []byte, err error) {
	genDoc, err := tmNode.Genesis()
	if err != nil {
		return nil, err
	}
	txBuilder := authentication.NewTxBuilder(
		authentication.DefaultTxEncoder(cdc),
		authentication.DefaultTxDecoder(cdc),
		genDoc.Genesis.ChainID,
		"", nil).WithKeybase(keybase)
	return txBuilder.BuildAndSignMultisigTransaction(address, publicKey, msg, passphrase, fee, legacyCodec)
}

func SignMulti(cdc *codec.Codec, fromAddr sdk.Address, tx []byte, keys []crypto.PublicKey, tmNode client.Client, keybase keys.Keybase, passphrase string, legacyCodec bool) (txBytes []byte, err error) {
	genDoc, err := tmNode.Genesis()
	if err != nil {
		return nil, err
	}
	txBuilder := authentication.NewTxBuilder(
		authentication.DefaultTxEncoder(cdc),
		authentication.DefaultTxDecoder(cdc),
		genDoc.Genesis.ChainID,
		"", nil).WithKeybase(keybase)
	return txBuilder.SignMultisigTransaction(fromAddr, keys, passphrase, tx, legacyCodec)
}
