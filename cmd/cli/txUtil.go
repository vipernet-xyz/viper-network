package cli

import (
	rand1 "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/vipernet-xyz/viper-network/app"
	"github.com/vipernet-xyz/viper-network/codec"
	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	"github.com/vipernet-xyz/viper-network/crypto/keys"
	"github.com/vipernet-xyz/viper-network/rpc"
	requestorsType "github.com/vipernet-xyz/viper-network/x/requestors/types"
	servicerTypes "github.com/vipernet-xyz/viper-network/x/servicers/types"
	viperTypes "github.com/vipernet-xyz/viper-network/x/viper-main/types"

	"github.com/tendermint/tendermint/libs/rand"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/authentication"
	authTypes "github.com/vipernet-xyz/viper-network/x/authentication/types"
	governanceTypes "github.com/vipernet-xyz/viper-network/x/governance/types"
)

// SendTransaction - Deliver Transaction to servicer
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
	msg := servicerTypes.MsgSend{
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

// LegacyStakeNode - Deliver Stake message to servicer
func LegacyStakeNode(chains []string, serviceURL, fromAddr, passphrase, chainID string, geoZone []string, amount sdk.BigInt, fees int64) (*rpc.SendRawTxParams, error) {
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
	err = servicerTypes.ValidateServiceURL(serviceURL)
	if err != nil {
		return nil, err
	}
	err = viperTypes.GeoZoneIdentifierVerification(geoZone[0])
	if err != nil {
		return nil, err
	}
	var msg sdk.ProtoMsg
	msg = &servicerTypes.MsgStake{
		PublicKey:  kp.PublicKey,
		Chains:     chains,
		Value:      amount,
		ServiceUrl: serviceURL,
		GeoZone:    geoZone,
		Output:     fa,
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

// StakeNode - Deliver Stake message to servicer
func StakeNode(chains []string, serviceURL, operatorPubKey, output, passphrase, chainID string, geoZone []string, amount sdk.BigInt, fees int64) (*rpc.SendRawTxParams, error) {
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
	err = servicerTypes.ValidateServiceURL(serviceURL)
	if err != nil {
		return nil, err
	}
	err = viperTypes.GeoZoneIdentifierVerification(geoZone[0])
	if err != nil {
		return nil, err
	}
	var msg sdk.ProtoMsg
	msg = &servicerTypes.MsgStake{
		PublicKey:  operatorPublicKey,
		Chains:     chains,
		Value:      amount,
		ServiceUrl: serviceURL,
		GeoZone:    geoZone,
		Output:     outputAddress,
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

// UnstakeNode - start unstaking message to servicer
func UnstakeNode(operatorAddr, fromAddr, passphrase, chainID string, fees int64) (*rpc.SendRawTxParams, error) {
	fa, err := sdk.AddressFromHex(fromAddr)
	if err != nil {
		return nil, err
	}
	oa, err := sdk.AddressFromHex(operatorAddr)
	if err != nil {
		return nil, err
	}
	var msg sdk.ProtoMsg
	msg = &servicerTypes.MsgBeginUnstake{
		Address: oa,
		Signer:  fa,
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

// UnjailNode - Remove servicer from jail
func UnjailNode(operatorAddr, fromAddr, passphrase, chainID string, fees int64) (*rpc.SendRawTxParams, error) {
	fa, err := sdk.AddressFromHex(fromAddr)
	if err != nil {
		return nil, err
	}
	oa, err := sdk.AddressFromHex(operatorAddr)
	if err != nil {
		return nil, err
	}
	var msg sdk.ProtoMsg
	msg = &servicerTypes.MsgUnjail{
		ValidatorAddr: oa,
		Signer:        fa}
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

func StakeClient(chains []string, fromAddr, passphrase, chainID string, amount sdk.BigInt, geoZones []string, numServicers int32, fees int64, legacyCodec bool) (*rpc.SendRawTxParams, error) {
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
	for _, geoZone := range geoZones {
		err := viperTypes.GeoZoneIdentifierVerification(geoZone)
		if err != nil {
			return nil, err
		}
	}
	if amount.LTE(sdk.NewInt(0)) {
		return nil, sdk.ErrInternal("must stake above zero")
	}
	msg := requestorsType.MsgStake{
		PubKey:       kp.PublicKey,
		Chains:       chains,
		Value:        amount,
		GeoZones:     geoZones,
		NumServicers: numServicers,
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

func UnstakeClient(fromAddr, passphrase, chainID string, fees int64, legacyCodec bool) (*rpc.SendRawTxParams, error) {
	fa, err := sdk.AddressFromHex(fromAddr)
	if err != nil {
		return nil, err
	}
	kb, err := app.GetKeybase()
	if err != nil {
		return nil, err
	}
	msg := requestorsType.MsgBeginUnstake{
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
	msg := governanceTypes.MsgDAOTransfer{
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
	msg := governanceTypes.MsgChangeParam{
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

func Upgrade(fromAddr string, upgrade governanceTypes.Upgrade, passphrase, chainID string, fees int64, legacyCodec bool) (*rpc.SendRawTxParams, error) {
	fa, err := sdk.AddressFromHex(fromAddr)
	if err != nil {
		return nil, err
	}
	kb, err := app.GetKeybase()
	if err != nil {
		return nil, err
	}
	msg := governanceTypes.MsgUpgrade{
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

	signBytes, err := authentication.StdSignBytes(chainID, entropy, fees, msg, memo)
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
		return authentication.DefaultTxEncoder(cdc)(tx, 0)
	}
	return authentication.DefaultTxEncoder(cdc)(tx, -1)
}

// PauseNode - Pause servicer
func PauseNode(operatorAddr, fromAddr, passphrase, chainID string, fees int64) (*rpc.SendRawTxParams, error) {
	fa, err := sdk.AddressFromHex(fromAddr)
	if err != nil {
		return nil, err
	}
	oa, err := sdk.AddressFromHex(operatorAddr)
	if err != nil {
		return nil, err
	}
	var msg sdk.ProtoMsg
	msg = &servicerTypes.MsgPause{
		ValidatorAddr: oa,
		Signer:        fa,
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

func UnpauseNode(operatorAddr, fromAddr, passphrase, chainID string, fees int64) (*rpc.SendRawTxParams, error) {
	fa, err := sdk.AddressFromHex(fromAddr)
	if err != nil {
		return nil, err
	}
	oa, err := sdk.AddressFromHex(operatorAddr)
	if err != nil {
		return nil, err
	}
	var msg sdk.ProtoMsg
	msg = &servicerTypes.MsgUnpause{
		ValidatorAddr: oa,
		Signer:        fa,
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

func GenerateAndSendDiscountKey(fromAddr, toAddr, passphrase, chainID string, fees int64, legacyCodec bool) (*rpc.SendRawTxParams, error) {
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

	// Generating a unique DiscountKey
	discountKey := generateUniqueDiscountKey(toAddr)

	msg := governanceTypes.MsgGenerateDiscountKey{
		FromAddress: fa,
		ToAddress:   ta,
		DiscountKey: discountKey,
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

// This function would generate a unique key for discounts.
func generateUniqueDiscountKey(requestorAddr string) string {
	// Concatenate requestor address and current time
	baseString := fmt.Sprintf("%s-%d", requestorAddr, time.Now().UnixNano())

	// Add a nonce for additional randomness
	nonce := generateNonce(16) // 16 bytes random data

	// Concatenate the nonce to the base string
	finalString := baseString + nonce

	// Hash the final string
	hashed := sha256.Sum256([]byte(finalString))

	// Return the hex representation of the hash
	return hex.EncodeToString(hashed[:])
}

// generateNonce creates a random string of the given length.
func generateNonce(length int) string {

	nonce := make([]byte, length)
	_, err := rand1.Read(nonce)
	if err != nil {

		// Handle the error. Here we're just returning a fixed string, but
		// you might want to handle it differently in production code.
		return "ErrorGeneratingNonce"
	}
	return hex.EncodeToString(nonce)
}
