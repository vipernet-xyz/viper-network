package cli

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/vipernet-xyz/viper-network/app"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(servicersCmd)
	servicersCmd.AddCommand(servicerUnstakeCmd)
	servicersCmd.AddCommand(servicerUnjailCmd)
	servicersCmd.AddCommand(servicerPauseCmd)
}

var servicersCmd = &cobra.Command{
	Use:   "servicers",
	Short: "servicer management",
	Long: `The servicer namespace handles all servicer related interactions,
from staking and unstaking; to unjailing.`,
}

func init() {
	servicerUnstakeCmd.Flags().StringVar(&pwd, "pwd", "", "passphrase used by the cmd, non empty usage bypass interactive prompt")
	servicerUnjailCmd.Flags().StringVar(&pwd, "pwd", "", "passphrase used by the cmd, non empty usage bypass interactive prompt")
}

var servicerUnstakeCmd = &cobra.Command{
	Use:   "unstake <operatorAddr> <fromAddr> <networkID> <fee>",
	Short: "Unstake a servicer in the network",
	Long: `Unstake a servicer from the network, changing it's status to Unstaking.
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

var servicerUnjailCmd = &cobra.Command{
	Use:   "unjail <operatorAddr> <fromAddr> <networkID> <fee>",
	Short: "Unjails a servicer in the network",
	Long: `Unjails a servicer from the network, allowing it to participate in service and consensus again.
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

var servicerPauseCmd = &cobra.Command{
	Use:   "pause <operatorAddr> <fromAddr> <networkID> <fee>",
	Short: "Pauses a servicer in the network",
	Long: `Pauses a servicer in the network, temporarily disabling its service.
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
		res, err := PauseNode(args[0], args[1], app.Credentials(pwd), args[2], int64(fee))
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

var servicerUnpauseCmd = &cobra.Command{
	Use:   "Unpause <operatorAddr> <fromAddr> <networkID> <fee>",
	Short: "Unpauses a servicer in the network",
	Long: `Unpauses a servicer in the network.
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
		res, err := UnpauseNode(args[0], args[1], app.Credentials(pwd), args[2], int64(fee))
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
