package pos

import (
	"fmt"

	"github.com/vipernet-xyz/viper-network/codec"
	"github.com/vipernet-xyz/viper-network/crypto/keys"
	"github.com/vipernet-xyz/viper-network/crypto/keys/mintkey"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/authentication"
	"github.com/vipernet-xyz/viper-network/x/authentication/util"
	"github.com/vipernet-xyz/viper-network/x/platforms/types"

	"github.com/tendermint/tendermint/rpc/client"
)

func StakeTx(cdc *codec.Codec, tmNode client.Client, keybase keys.Keybase, chains []string, amount sdk.BigInt, kp keys.KeyPair, passphrase string, legacyCodec bool) (*sdk.TxResponse, error) {
	fromAddr := kp.GetAddress()
	msg := types.MsgStake{
		PubKey: kp.PublicKey,
		Value:  amount,
		Chains: chains, // non native blockchains
	}
	txBuilder, cliCtx, err := newTx(cdc, &msg, fromAddr, tmNode, keybase, passphrase)
	if err != nil {
		return nil, err
	}
	err = msg.ValidateBasic()
	if err != nil {
		return nil, err
	}
	return util.CompleteAndBroadcastTxCLI(txBuilder, cliCtx, &msg, legacyCodec)
}

func UnstakeTx(cdc *codec.Codec, tmNode client.Client, keybase keys.Keybase, address sdk.Address, passphrase string, legacyCodec bool) (*sdk.TxResponse, error) {
	msg := types.MsgBeginUnstake{Address: address}
	txBuilder, cliCtx, err := newTx(cdc, &msg, address, tmNode, keybase, passphrase)
	if err != nil {
		return nil, err
	}
	err = msg.ValidateBasic()
	if err != nil {
		return nil, err
	}
	return util.CompleteAndBroadcastTxCLI(txBuilder, cliCtx, &msg, legacyCodec)
}

func newTx(cdc *codec.Codec, msg sdk.ProtoMsg, fromAddr sdk.Address, tmNode client.Client, keybase keys.Keybase, passphrase string) (txBuilder authentication.TxBuilder, cliCtx util.CLIContext, err error) {
	genDoc, err := tmNode.Genesis()
	if err != nil {
		return
	}
	chainID := genDoc.Genesis.ChainID
	kp, err := keybase.Get(fromAddr)
	if err != nil {
		return
	}
	privkey, err := mintkey.UnarmorDecryptPrivKey(kp.PrivKeyArmor, passphrase)
	if err != nil {
		return
	}
	cliCtx = util.NewCLIContext(tmNode, fromAddr, passphrase).WithCodec(cdc)
	cliCtx.BroadcastMode = util.BroadcastSync
	cliCtx.PrivateKey = privkey
	account, err := cliCtx.GetAccount(fromAddr)
	if err != nil {
		return
	}
	fee := msg.GetFee()
	if account.GetCoins().AmountOf(sdk.DefaultStakeDenom).LT(fee) { // todo get stake denom
		_ = fmt.Errorf("insufficient funds: the fee needed is %v", fee)
		return
	}
	txBuilder = authentication.NewTxBuilder(
		authentication.DefaultTxEncoder(cdc),
		authentication.DefaultTxDecoder(cdc),
		chainID,
		"",
		sdk.NewCoins(sdk.NewCoin(sdk.DefaultStakeDenom, fee))).WithKeybase(keybase)
	return
}
