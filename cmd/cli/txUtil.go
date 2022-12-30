package cli

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/vipernet-xyz/viper-network/app"
	"github.com/vipernet-xyz/viper-network/cmd/rpc"
	"github.com/vipernet-xyz/viper-network/codec"
	"github.com/vipernet-xyz/viper-network/crypto"
	"github.com/vipernet-xyz/viper-network/crypto/keys"
	appsType "github.com/vipernet-xyz/viper-network/x/apps/types"
	nodeTypes "github.com/vipernet-xyz/viper-network/x/nodes/types"
	viperTypes "github.com/vipernet-xyz/viper-network/x/vipercore/types"

	"github.com/tendermint/tendermint/libs/rand"

	//"github.com/vipernet-xyz/viper-network/crypto/keys/mintkey"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/auth"
	authTypes "github.com/vipernet-xyz/viper-network/x/auth/types"
	govTypes "github.com/vipernet-xyz/viper-network/x/gov/types"
)

// SendTransaction - Deliver Transaction to node
func SendTransaction(fromAddr, toAddr, passphrase, chainID string, amount sdk.BigInt, fees int64, memo string, legacyCodec bool) (*rpc.SendRawTxParams, error) {
	fa, err := sdk.AddressFromHex(fromAddr)
	if err != nil {
		return nil, err
	}
	ta, err := sdk.AddressFromHex(toAddr)
	if err != nil {
		return nil, err
	}
	if amount.LTE(sdk.ZeroInt()) {
		return nil, sdk.ErrInternal("must send above 0")
	}
	kb, err := app.GetKeybase()
	if err != nil {
		return nil, err
	}
	msg := nodeTypes.MsgSend{
		FromAddress: fa,
		ToAddress:   ta,
		Amount:      amount,
	}
	err = msg.ValidateBasic()
	if err != nil {
		return nil, err
	}
	txBz, err := newTxBz(app.Codec(), &msg, fa, chainID, kb, passphrase, fees, memo, legacyCodec)
	if err != nil {
		return nil, err
	}
	return &rpc.SendRawTxParams{
		Addr:        fromAddr,
		RawHexBytes: hex.EncodeToString(txBz),
	}, nil
}

// LegacyStakeNode - Deliver Stake message to node
func LegacyStakeNode(chains []string, serviceURL, fromAddr, passphrase, chainID string, amount sdk.BigInt, fees int64, isBefore8 bool) (*rpc.SendRawTxParams, error) {
	fa, err := sdk.AddressFromHex(fromAddr)
	if err != nil {
		return nil, err
	}
	kb, err := app.GetKeybase()
	if err != nil {
		return nil, err
	}
	kp, err := kb.Get(fa)
	if err != nil {
		return nil, err
	}
	m := make(map[string]struct{})
	for _, chain := range chains {
		if _, found := m[chain]; found {
			return nil, sdk.ErrInternal("cannot stake duplicate relayChainIDs: " + chain)
		}
		if len(chain) != viperTypes.NetworkIdentifierLength {
			return nil, sdk.ErrInternal("invalid relayChainID " + chain)
		}
		err := viperTypes.NetworkIdentifierVerification(chain)
		if err != nil {
			return nil, err
		}
	}
	if amount.LTE(sdk.NewInt(0)) {
		return nil, sdk.ErrInternal("must stake above zero")
	}
	err = nodeTypes.ValidateServiceURL(serviceURL)
	if err != nil {
		return nil, err
	}
	var msg sdk.ProtoMsg
	if isBefore8 {
		msg = &nodeTypes.LegacyMsgStake{
			PublicKey:  kp.PublicKey,
			Chains:     chains,
			Value:      amount,
			ServiceUrl: serviceURL,
		}
	} else {
		msg = &nodeTypes.MsgStake{
			PublicKey:  kp.PublicKey,
			Chains:     chains,
			Value:      amount,
			ServiceUrl: serviceURL,
			Output:     fa,
		}
	}
	err = msg.ValidateBasic()
	if err != nil {
		return nil, err
	}
	txBz, err := newTxBz(app.Codec(), msg, fa, chainID, kb, passphrase, fees, "", false)
	if err != nil {
		return nil, err
	}
	return &rpc.SendRawTxParams{
		Addr:        fromAddr,
		RawHexBytes: hex.EncodeToString(txBz),
	}, nil
}

// StakeNode - Deliver Stake message to node
func StakeNode(chains []string, serviceURL, operatorPubKey, output, passphrase, chainID string, amount sdk.BigInt, fees int64, isBefore8 bool) (*rpc.SendRawTxParams, error) {
	var operatorPublicKey crypto.PublicKey
	var operatorAddress sdk.Address
	var fromAddress sdk.Address
	kb, err := app.GetKeybase()
	if err != nil {
		return nil, err
	}
	bz, err := hex.DecodeString(operatorPubKey)
	if err != nil {
		return nil, err
	}

	pbkey, err := crypto.NewPublicKeyBz(bz)
	if err != nil {
		return nil, err
	}
	operatorPublicKey = pbkey

	outputAddress, err := sdk.AddressFromHex(output)
	if err != nil {
		return nil, err
	}
	kp, err := kb.Get(outputAddress)
	if err != nil {
		operatorAddress = sdk.Address(operatorPublicKey.Address())
		kp, err = kb.Get(operatorAddress)
		if err != nil {
			return nil, errors.New("Neither the Output Address nor the Operator Address is able to be retrieved from the keybase" + err.Error())
		}
		fromAddress = kp.GetAddress()
	} else {
		fromAddress = outputAddress
	}
	m := make(map[string]struct{})
	for _, chain := range chains {
		if _, found := m[chain]; found {
			return nil, sdk.ErrInternal("cannot stake duplicate relayChainIDs: " + chain)
		}
		if len(chain) != viperTypes.NetworkIdentifierLength {
			return nil, sdk.ErrInternal("invalid relayChainID " + chain)
		}
		err := viperTypes.NetworkIdentifierVerification(chain)
		if err != nil {
			return nil, err
		}
	}
	if amount.LTE(sdk.NewInt(0)) {
		return nil, sdk.ErrInternal("must stake above zero")
	}
	err = nodeTypes.ValidateServiceURL(serviceURL)
	if err != nil {
		return nil, err
	}
	var msg sdk.ProtoMsg
	if isBefore8 {
		msg = &nodeTypes.LegacyMsgStake{
			PublicKey:  operatorPublicKey,
			Chains:     chains,
			Value:      amount,
			ServiceUrl: serviceURL,
		}
	} else {
		msg = &nodeTypes.MsgStake{
			PublicKey:  operatorPublicKey,
			Chains:     chains,
			Value:      amount,
			ServiceUrl: serviceURL,
			Output:     outputAddress,
		}
	}
	err = msg.ValidateBasic()
	if err != nil {
		return nil, err
	}
	txBz, err := newTxBz(app.Codec(), msg, fromAddress, chainID, kb, passphrase, fees, "", false)
	if err != nil {
		return nil, err
	}
	return &rpc.SendRawTxParams{
		Addr:        operatorAddress.String(),
		RawHexBytes: hex.EncodeToString(txBz),
	}, nil
}

// UnstakeNode - start unstaking message to node
func UnstakeNode(operatorAddr, fromAddr, passphrase, chainID string, fees int64, isBefore8 bool) (*rpc.SendRawTxParams, error) {
	fa, err := sdk.AddressFromHex(fromAddr)
	if err != nil {
		return nil, err
	}
	oa, err := sdk.AddressFromHex(operatorAddr)
	if err != nil {
		return nil, err
	}
	var msg sdk.ProtoMsg
	if isBefore8 {
		msg = &nodeTypes.LegacyMsgBeginUnstake{
			Address: oa,
		}
	} else {
		msg = &nodeTypes.MsgBeginUnstake{
			Address: oa,
			Signer:  fa,
		}
	}
	kb, err := app.GetKeybase()
	if err != nil {
		return nil, err
	}
	err = msg.ValidateBasic()
	if err != nil {
		return nil, err
	}
	txBz, err := newTxBz(app.Codec(), msg, fa, chainID, kb, passphrase, fees, "", false)
	if err != nil {
		return nil, err
	}
	return &rpc.SendRawTxParams{
		Addr:        fromAddr,
		RawHexBytes: hex.EncodeToString(txBz),
	}, nil
}

// UnjailNode - Remove node from jail
func UnjailNode(operatorAddr, fromAddr, passphrase, chainID string, fees int64, isBefore8 bool) (*rpc.SendRawTxParams, error) {
	fa, err := sdk.AddressFromHex(fromAddr)
	if err != nil {
		return nil, err
	}
	oa, err := sdk.AddressFromHex(operatorAddr)
	if err != nil {
		return nil, err
	}
	var msg sdk.ProtoMsg
	if isBefore8 {
		msg = &nodeTypes.LegacyMsgUnjail{
			ValidatorAddr: oa,
		}
	} else {
		msg = &nodeTypes.MsgUnjail{
			ValidatorAddr: oa,
			Signer:        fa,
		}
	}
	kb, err := app.GetKeybase()
	if err != nil {
		return nil, err
	}
	err = msg.ValidateBasic()
	if err != nil {
		return nil, err
	}
	txBz, err := newTxBz(app.Codec(), msg, fa, chainID, kb, passphrase, fees, "", false)
	if err != nil {
		return nil, err
	}
	return &rpc.SendRawTxParams{
		Addr:        fromAddr,
		RawHexBytes: hex.EncodeToString(txBz),
	}, nil
}

func StakeApp(chains []string, fromAddr, passphrase, chainID string, amount sdk.BigInt, fees int64, legacyCodec bool) (*rpc.SendRawTxParams, error) {
	fa, err := sdk.AddressFromHex(fromAddr)
	if err != nil {
		return nil, err
	}
	kb, err := app.GetKeybase()
	if err != nil {
		return nil, err
	}
	kp, err := kb.Get(fa)
	if err != nil {
		return nil, err
	}
	for _, chain := range chains {
		fmt.Println(chain)
		err := viperTypes.NetworkIdentifierVerification(chain)
		if err != nil {
			return nil, err
		}
	}
	if amount.LTE(sdk.NewInt(0)) {
		return nil, sdk.ErrInternal("must stake above zero")
	}
	msg := appsType.MsgStake{
		PubKey: kp.PublicKey,
		Chains: chains,
		Value:  amount,
	}
	err = msg.ValidateBasic()
	if err != nil {
		return nil, err
	}
	txBz, err := newTxBz(app.Codec(), &msg, fa, chainID, kb, passphrase, fees, "", legacyCodec)
	if err != nil {
		return nil, err
	}
	return &rpc.SendRawTxParams{
		Addr:        fromAddr,
		RawHexBytes: hex.EncodeToString(txBz),
	}, nil
}

func UnstakeApp(fromAddr, passphrase, chainID string, fees int64, legacyCodec bool) (*rpc.SendRawTxParams, error) {
	fa, err := sdk.AddressFromHex(fromAddr)
	if err != nil {
		return nil, err
	}
	kb, err := app.GetKeybase()
	if err != nil {
		return nil, err
	}
	msg := appsType.MsgBeginUnstake{
		Address: fa,
	}
	err = msg.ValidateBasic()
	if err != nil {
		return nil, err
	}
	txBz, err := newTxBz(app.Codec(), &msg, fa, chainID, kb, passphrase, fees, "", legacyCodec)
	if err != nil {
		return nil, err
	}
	return &rpc.SendRawTxParams{
		Addr:        fromAddr,
		RawHexBytes: hex.EncodeToString(txBz),
	}, nil
}

func DAOTx(fromAddr, toAddr, passphrase string, amount sdk.BigInt, action, chainID string, fees int64, legacyCodec bool) (*rpc.SendRawTxParams, error) {
	fa, err := sdk.AddressFromHex(fromAddr)
	if err != nil {
		return nil, err
	}
	ta, err := sdk.AddressFromHex(toAddr)
	if err != nil {
		return nil, err
	}
	kb, err := app.GetKeybase()
	if err != nil {
		return nil, err
	}
	msg := govTypes.MsgDAOTransfer{
		FromAddress: fa,
		ToAddress:   ta,
		Amount:      amount,
		Action:      action,
	}
	err = msg.ValidateBasic()
	if err != nil {
		return nil, err
	}
	txBz, err := newTxBz(app.Codec(), &msg, fa, chainID, kb, passphrase, fees, "", legacyCodec)
	if err != nil {
		return nil, err
	}
	return &rpc.SendRawTxParams{
		Addr:        fromAddr,
		RawHexBytes: hex.EncodeToString(txBz),
	}, nil
}

func ChangeParam(fromAddr, paramACLKey string, paramValue json.RawMessage, passphrase, chainID string, fees int64, legacyCodec bool) (*rpc.SendRawTxParams, error) {
	fa, err := sdk.AddressFromHex(fromAddr)
	if err != nil {
		return nil, err
	}
	kb, err := app.GetKeybase()
	if err != nil {
		return nil, err
	}

	valueBytes, err := app.Codec().MarshalJSON(paramValue)
	if err != nil {
		return nil, err

	}
	msg := govTypes.MsgChangeParam{
		FromAddress: fa,
		ParamKey:    paramACLKey,
		ParamVal:    valueBytes,
	}
	err = msg.ValidateBasic()
	if err != nil {
		return nil, err
	}
	txBz, err := newTxBz(app.Codec(), &msg, fa, chainID, kb, passphrase, fees, "", legacyCodec)
	if err != nil {
		return nil, err
	}
	return &rpc.SendRawTxParams{
		Addr:        fromAddr,
		RawHexBytes: hex.EncodeToString(txBz),
	}, nil
}

func Upgrade(fromAddr string, upgrade govTypes.Upgrade, passphrase, chainID string, fees int64, legacyCodec bool) (*rpc.SendRawTxParams, error) {
	fa, err := sdk.AddressFromHex(fromAddr)
	if err != nil {
		return nil, err
	}
	kb, err := app.GetKeybase()
	if err != nil {
		return nil, err
	}
	msg := govTypes.MsgUpgrade{
		Address: fa,
		Upgrade: upgrade,
	}
	err = msg.ValidateBasic()
	if err != nil {
		return nil, err
	}
	txBz, err := newTxBz(app.Codec(), &msg, fa, chainID, kb, passphrase, fees, "", legacyCodec)
	if err != nil {
		return nil, err
	}
	return &rpc.SendRawTxParams{
		Addr:        fromAddr,
		RawHexBytes: hex.EncodeToString(txBz),
	}, nil
}

func newTxBz(cdc *codec.Codec, msg sdk.ProtoMsg, fromAddr sdk.Address, chainID string, keybase keys.Keybase, passphrase string, fee int64, memo string, legacyCodec bool) (transactionBz []byte, err error) {
	// fees
	fees := sdk.NewCoins(sdk.NewCoin(sdk.DefaultStakeDenom, sdk.NewInt(fee)))
	// entroyp
	entropy := rand.Int64()
	signBytes, err := auth.StdSignBytes(chainID, entropy, fees, msg, memo)
	if err != nil {
		return nil, err
	}
	sig, pubKey, err := keybase.Sign(fromAddr, passphrase, signBytes)
	if err != nil {
		return nil, err
	}
	s := authTypes.StdSignature{PublicKey: pubKey, Signature: sig}
	tx := authTypes.NewTx(msg, fees, s, memo, entropy)
	if legacyCodec {
		return auth.DefaultTxEncoder(cdc)(tx, 0)
	}
	return auth.DefaultTxEncoder(cdc)(tx, -1)
}
