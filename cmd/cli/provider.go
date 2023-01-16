package cli

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/vipernet-xyz/viper-network/app"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(providersCmd)
	providersCmd.AddCommand(providerUnstakeCmd)
	providersCmd.AddCommand(providerUnjailCmd)
}

var providersCmd = &cobra.Command{
	Use:   "providers",
	Short: "provider management",
	Long: `The provider namespace handles all provider related interactions,
from staking and unstaking; to unjailing.`,
}

func init() {
	providerUnstakeCmd.Flags().StringVar(&pwd, "pwd", "", "passphrase used by the cmd, non empty usage bypass interactive prompt")
	providerUnjailCmd.Flags().StringVar(&pwd, "pwd", "", "passphrase used by the cmd, non empty usage bypass interactive prompt")
}

var providerUnstakeCmd = &cobra.Command{
	Use:   "unstake <operatorAddr> <fromAddr> <networkID> <fee>",
	Short: "Unstake a provider in the network",
	Long: `Unstake a provider from the network, changing it's status to Unstaking.
Will prompt the user for the <fromAddr> account passphrase.`,
	Args: cobra.ExactArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		fee, err := strconv.Atoi(args[3])
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Enter Password: ")
		res, err := UnstakeNode(args[0], args[1], app.Credentials(pwd), args[2], int64(fee))
		if err != nil {
			fmt.Println(err)
			return
		}
		j, err := json.Marshal(res)
		if err != nil {
			fmt.Println(err)
			return
		}
		resp, err := QueryRPC(SendRawTxPath, j)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(resp)
	},
}

var providerUnjailCmd = &cobra.Command{
	Use:   "unjail <operatorAddr> <fromAddr> <networkID> <fee>",
	Short: "Unjails a provider in the network",
	Long: `Unjails a provider from the network, allowing it to participate in service and consensus again.
Will prompt the user for the <fromAddr> account passphrase.`,
	Args: cobra.ExactArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		fee, err := strconv.Atoi(args[3])
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Enter Password: ")
		res, err := UnjailNode(args[0], args[1], app.Credentials(pwd), args[2], int64(fee))
		if err != nil {
			fmt.Println(err)
			return
		}
		j, err := json.Marshal(res)
		if err != nil {
			fmt.Println(err)
			return
		}
		resp, err := QueryRPC(SendRawTxPath, j)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(resp)
	},
}
