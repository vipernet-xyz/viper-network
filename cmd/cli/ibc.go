package cli

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/vipernet-xyz/viper-network/app"
	"github.com/vipernet-xyz/viper-network/client"
	"github.com/vipernet-xyz/viper-network/client/tx"
	clienttypes "github.com/vipernet-xyz/viper-network/modules/core/02-client/types"
	channelutils "github.com/vipernet-xyz/viper-network/modules/core/04-channel/client/cli/utils"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/transfer/types"
)

func init() {
	rootCmd.AddCommand(ibcCmd)
	ibcCmd.AddCommand(GetCmdQueryDenomTrace)
	ibcCmd.AddCommand(GetCmdQueryEscrowAddress)
	ibcCmd.AddCommand(NewTransferTxCmd)
}

// accountsCmd represents the accounts namespace command
var ibcCmd = &cobra.Command{
	Use:   "ibc-transfer",
	Short: "IBC-Transfer",
	Long:  `IBC fungible token transfer query subcommands`,
}

// GetCmdQueryDenomTrace defines the command to query a a denomination trace from a given trace hash or ibc denom.
var GetCmdQueryDenomTrace = &cobra.Command{
	Use:     "denom-trace [hash/denom]",
	Short:   "Query the denom trace info from a given trace hash or ibc denom",
	Long:    "Query the denom trace info from a given trace hash or ibc denom",
	Example: fmt.Sprintf("%s query ibc-transfer denom-trace 27A6394C3F9FF9C9DCF5DFFADF9BB5FE9A37C7E92B006199894CF1824DF9AC7C", version.Version),
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		clientCtx, err := client.GetClientQueryContext(cmd)
		if err != nil {
			return err
		}
		queryClient := types.NewQueryClient(clientCtx)

		req := &types.QueryDenomTraceRequest{
			Hash: args[0],
		}

		res, err := queryClient.DenomTrace(cmd.Context(), req)
		if err != nil {
			return err
		}

		return clientCtx.PrintProto(res)
	},
}

// GetCmdParams returns the command handler for ibc-transfer parameter querying.
var GetCmdQueryEscrowAddress = &cobra.Command{
	Use:     "escrow-address",
	Short:   "Get the escrow address for a channel",
	Long:    "Get the escrow address for a channel",
	Args:    cobra.ExactArgs(2),
	Example: fmt.Sprintf("%s query ibc-transfer escrow-address [port] [channel-id]", version.Version),
	RunE: func(cmd *cobra.Command, args []string) error {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		clientCtx, err := client.GetClientQueryContext(cmd)
		if err != nil {
			return err
		}
		port := args[0]
		channel := args[1]
		addr := types.GetEscrowAddress(port, channel)
		return clientCtx.PrintString(fmt.Sprintf("%s\n", addr.String()))
	},
}

const (
	flagPacketTimeoutHeight    = "packet-timeout-height"
	flagPacketTimeoutTimestamp = "packet-timeout-timestamp"
	flagAbsoluteTimeouts       = "absolute-timeouts"
	flagMemo                   = "memo"
)

// NewTransferTxCmd returns the command to create a NewMsgTransfer transaction
var NewTransferTxCmd = &cobra.Command{
	Use:     "transfer [src-port] [src-channel] [receiver] [amount]",
	Short:   "Transfer a fungible token through IBC",
	Long:    strings.TrimSpace(`Transfer a fungible token through IBC`),
	Example: fmt.Sprintf("%s tx ibc-transfer transfer [src-port] [src-channel] [receiver] [amount]", version.Version),
	Args:    cobra.ExactArgs(4),
	RunE: func(cmd *cobra.Command, args []string) error {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		clientCtx, err := client.GetClientTxContext(cmd)
		if err != nil {
			return err
		}
		sender := clientCtx.GetFromAddress().String()
		srcPort := args[0]
		srcChannel := args[1]
		receiver := args[2]

		coin, err := sdk.ParseCoinNormalized(args[3])
		if err != nil {
			return err
		}

		if !strings.HasPrefix(coin.Denom, "ibc/") {
			denomTrace := types.ParseDenomTrace(coin.Denom)
			coin.Denom = denomTrace.IBCDenom()
		}

		timeoutHeightStr, err := cmd.Flags().GetString(flagPacketTimeoutHeight)
		if err != nil {
			return err
		}
		timeoutHeight, err := clienttypes.ParseHeight(timeoutHeightStr)
		if err != nil {
			return err
		}

		timeoutTimestamp, err := cmd.Flags().GetUint64(flagPacketTimeoutTimestamp)
		if err != nil {
			return err
		}

		absoluteTimeouts, err := cmd.Flags().GetBool(flagAbsoluteTimeouts)
		if err != nil {
			return err
		}

		memo, err := cmd.Flags().GetString(flagMemo)
		if err != nil {
			return err
		}

		// if the timeouts are not absolute, retrieve latest block height and block timestamp
		// for the consensus state connected to the destination port/channel
		if !absoluteTimeouts {
			consensusState, height, _, err := channelutils.QueryLatestConsensusState(clientCtx, srcPort, srcChannel)
			if err != nil {
				return err
			}

			if !timeoutHeight.IsZero() {
				absoluteHeight := height
				absoluteHeight.RevisionNumber += timeoutHeight.RevisionNumber
				absoluteHeight.RevisionHeight += timeoutHeight.RevisionHeight
				timeoutHeight = absoluteHeight
			}

			if timeoutTimestamp != 0 {
				// use local clock time as reference time if it is later than the
				// consensus state timestamp of the counter party chain, otherwise
				// still use consensus state timestamp as reference
				now := time.Now().UnixNano()
				consensusStateTimestamp := consensusState.GetTimestamp()
				if now > 0 {
					now := uint64(now)
					if now > consensusStateTimestamp {
						timeoutTimestamp = now + timeoutTimestamp
					} else {
						timeoutTimestamp = consensusStateTimestamp + timeoutTimestamp
					}
				} else {
					return errors.New("local clock time is not greater than Jan 1st, 1970 12:00 AM")
				}
			}
		}

		msg := types.NewMsgTransfer(
			srcPort, srcChannel, coin, sender, receiver, timeoutHeight, timeoutTimestamp, memo,
		)
		return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
	},
}

func init() {
	NewTransferTxCmd.Flags().String(flagPacketTimeoutHeight, types.DefaultRelativePacketTimeoutHeight, "Packet timeout block height. The timeout is disabled when set to 0-0.")
	NewTransferTxCmd.Flags().Uint64(flagPacketTimeoutTimestamp, types.DefaultRelativePacketTimeoutTimestamp, "Packet timeout timestamp in nanoseconds from now. Default is 10 minutes. The timeout is disabled when set to 0.")
	NewTransferTxCmd.Flags().Bool(flagAbsoluteTimeouts, false, "Timeout flags are used as absolute timeouts.")
	NewTransferTxCmd.Flags().String(flagMemo, "", "Memo to be sent along with the packet.")

}
