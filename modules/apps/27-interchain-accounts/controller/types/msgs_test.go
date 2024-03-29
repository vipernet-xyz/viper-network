package types_test

/*
	"testing"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
	sdk "github.com/vipernet-xyz/viper-network/types"
	banktypes "github.com/vipernet-xyz/viper-network/x/bank/types"

	"github.com/vipernet-xyz/viper-network/modules/apps/27-interchain-accounts/controller/types"
	icatypes "github.com/vipernet-xyz/viper-network/modules/apps/27-interchain-accounts/types"
	feetypes "github.com/vipernet-xyz/viper-network/modules/apps/29-fee/types"
	ibctesting "github.com/vipernet-xyz/viper-network/testing"
	"github.com/vipernet-xyz/viper-network/testing/simapp"
*/

/*
func TestMsgRegisterInterchainAccountValidateBasic(t *testing.T) {
	var msg *types.MsgRegisterInterchainAccount

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"success",
			func() {},
			true,
		},
		{
			"success: with empty channel version",
			func() {
				msg.Version = ""
			},
			true,
		},
		{
			"success: with fee enabled channel version",
			func() {
				feeMetadata := feetypes.Metadata{
					FeeVersion: feetypes.Version,
					AppVersion: icatypes.NewDefaultMetadataString(ibctesting.FirstConnectionID, ibctesting.FirstConnectionID),
				}

				bz := feetypes.ModuleCdc.MustMarshalJSON(&feeMetadata)
				msg.Version = string(bz)
			},
			true,
		},
		{
			"connection id is invalid",
			func() {
				msg.ConnectionId = ""
			},
			false,
		},
		{
			"owner address is empty",
			func() {
				msg.Owner = ""
			},
			false,
		},
	}

	for i, tc := range testCases {

		msg = types.NewMsgRegisterInterchainAccount(
			ibctesting.FirstConnectionID,
			ibctesting.TestAccAddress,
			icatypes.NewDefaultMetadataString(ibctesting.FirstConnectionID, ibctesting.FirstConnectionID),
		)

		tc.malleate()

		err := msg.ValidateBasic()
		if tc.expPass {
			require.NoError(t, err, "valid test case %d failed: %s", i, tc.name)
		} else {
			require.Error(t, err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func TestMsgRegisterInterchainAccountGetSigners(t *testing.T) {
	expSigner, err := sdk.AccAddressFromBech32(ibctesting.TestAccAddress)
	require.NoError(t, err)

	msg := types.NewMsgRegisterInterchainAccount(ibctesting.FirstConnectionID, ibctesting.TestAccAddress, "")
	require.Equal(t, []sdk.AccAddress{expSigner}, msg.GetSigners())
}

func TestMsgSendTxValidateBasic(t *testing.T) {
	var msg *types.MsgSendTx

	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{
			"success",
			func() {},
			true,
		},
		{
			"connection id is invalid",
			func() {
				msg.ConnectionId = ""
			},
			false,
		},
		{
			"owner address is empty",
			func() {
				msg.Owner = ""
			},
			false,
		},
		{
			"relative timeout is not set",
			func() {
				msg.RelativeTimeout = 0
			},
			false,
		},
		{
			"messages array is empty",
			func() {
				msg.PacketData = icatypes.InterchainAccountPacketData{}
			},
			false,
		},
	}

	for i, tc := range testCases {

		msgBankSend := &banktypes.MsgSend{
			FromAddress: ibctesting.TestAccAddress,
			ToAddress:   ibctesting.TestAccAddress,
			Amount:      ibctesting.TestCoins,
		}

		data, err := icatypes.SerializeCosmosTx(simapp.MakeTestEncodingConfig().Marshaler, []proto.Message{msgBankSend})
		require.NoError(t, err)

		packetData := icatypes.InterchainAccountPacketData{
			Type: icatypes.EXECUTE_TX,
			Data: data,
		}

		msg = types.NewMsgSendTx(
			ibctesting.TestAccAddress,
			ibctesting.FirstConnectionID,
			100000,
			packetData,
		)

		tc.malleate()

		err = msg.ValidateBasic()
		if tc.expPass {
			require.NoError(t, err, "valid test case %d failed: %s", i, tc.name)
		} else {
			require.Error(t, err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func TestMsgSendTxGetSigners(t *testing.T) {
	expSigner, err := sdk.AccAddressFromBech32(ibctesting.TestAccAddress)
	require.NoError(t, err)

	msgBankSend := &banktypes.MsgSend{
		FromAddress: ibctesting.TestAccAddress,
		ToAddress:   ibctesting.TestAccAddress,
		Amount:      ibctesting.TestCoins,
	}

	data, err := icatypes.SerializeCosmosTx(simapp.MakeTestEncodingConfig().Marshaler, []proto.Message{msgBankSend})
	require.NoError(t, err)

	packetData := icatypes.InterchainAccountPacketData{
		Type: icatypes.EXECUTE_TX,
		Data: data,
	}

	msg := types.NewMsgSendTx(
		ibctesting.TestAccAddress,
		ibctesting.FirstConnectionID,
		100000,
		packetData,
	)
	require.Equal(t, []sdk.Address{expSigner}, msg.GetSigners())
}
*/
