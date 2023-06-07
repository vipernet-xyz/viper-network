package cli

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/vipernet-xyz/viper-network/app"
	"github.com/vipernet-xyz/viper-network/rpc"
	"github.com/vipernet-xyz/viper-network/types"

	"github.com/spf13/cobra"
)

func init() {
	servicersCmd.AddCommand(servicerStakeCmd)
	servicerStakeCmd.AddCommand(custodialStakeCmd)
	servicerStakeCmd.AddCommand(nonCustodialstakeCmd)

	custodialStakeCmd.Flags().StringVar(&pwd, "pwd", "", "passphrase used by the cmd, non empty usage bypass interactive prompt")
	nonCustodialstakeCmd.Flags().StringVar(&pwd, "pwd", "", "passphrase used by the cmd, non empty usage bypass interactive prompt")

}

var servicerStakeCmd = &cobra.Command{
	Use:   "stake",
	Short: "Stake a servicer in the network",
	Long:  "Stake the servicer into the network, making it available for service.",
}

var custodialStakeCmd = &cobra.Command{
	Use:   "custodial <fromAddr> <amount> <RelayChainIDs> <serviceURI> <networkID> <geoZone> <fee>",
	Short: "Stake a servicer in the network. Custodial stake uses the same address as operator/output for rewards/return of staked funds.",
	Long: `Stake the servicer into the network, making it available for service.
Will prompt the user for the <fromAddr> account passphrase. If the servicer is already staked, this transaction acts as an *update* transaction.
A servicer can updated relayChainIDs, serviceURI, and raise the stake amount with this transaction.
If the servicer is currently staked at X and you submit an update with new stake Y. Only Y-X will be subtracted from an account
If no changes are desired for the parameter, just enter the current param value just as before`,
	Args: cobra.ExactArgs(7),
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		fromAddr := args[0]
		amount, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Println(err)
			return
		}
		am := types.NewInt(int64(amount))
		if am.LT(types.NewInt(15100000000)) {
			fmt.Println("The amount you are staking for is below the recommendation of 15100 VIPR, would you still like to continue? y|n")
			if !app.Confirmation(pwd) {
				return
			}
		}
		reg, err := regexp.Compile("[^,a-fA-F0-9]+")
		if err != nil {
			log.Fatal(err)
		}
		rawChains := reg.ReplaceAllString(args[2], "")
		chains := strings.Split(rawChains, ",")
		serviceURI := args[3]
		fee, err := strconv.Atoi(args[6])
		if err != nil {
			fmt.Println(err)
			return
		}
		geozone, err := strconv.Atoi(args[5])
		if err != nil {
			fmt.Println(err)
			return
		}
		params := rpc.HeightAndKeyParams{
			Height: 0,
			Key:    "ServicerCountLock",
		}
		j, _ := json.Marshal(params)
		res, _ := QueryRPC(GetParamPath, j)
		if res == "true" {
			fmt.Println("Node Staking is Locked; 'ServicerCountLock' is activated to control inflated node count")
			return
		}
		fmt.Println("Enter Passphrase: ")
		res1, err := LegacyStakeNode(chains, serviceURI, fromAddr, app.Credentials(pwd), args[4], int64(geozone), types.NewInt(int64(amount)), int64(fee))
		if err != nil {
			fmt.Println(err)
			return
		}
		j1, err := json.Marshal(res1)
		if err != nil {
			fmt.Println(err)
			return
		}
		resp, err := QueryRPC(SendRawTxPath, j1)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(resp)
	},
}

var nonCustodialstakeCmd = &cobra.Command{
	Use:   "non-custodial <operatorPublicKey> <outputAddress> <amount> <RelayChainIDs> <serviceURI> <networkID> <geoZone> <fee>",
	Short: "Stake a servicer in the network, non-custodial stake allows a different output address for rewards/return of staked funds. The signer may be the operator or the output address. The signer must specify the public key of the operator",
	Long: `Stake the servicer into the network, making it available for service.
Will prompt the user for the signer account passphrase, fund and fees are collected from signer account. If both accounts are present signer priority is first output then operator. If the servicer is already staked, this transaction acts as an *update* transaction.
A servicer can updated relayChainIDs, serviceURI, and raise the stake amount with this transaction.
If the servicer is currently staked at X and you submit an update with new stake Y. Only Y-X will be subtracted from an account
If no changes are desired for the parameter, just enter the current param value just as before.
The signer may be the operator or the output address.`,
	Args: cobra.ExactArgs(8),
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		operatorPubKey := args[0]
		output := args[1]
		amount, err := strconv.Atoi(args[2])
		if err != nil {
			fmt.Println(err)
			return
		}
		am := types.NewInt(int64(amount))
		if am.LT(types.NewInt(15100000000)) {
			fmt.Println("The amount you are staking for is below the recommendation of 15100 VIPR, would you still like to continue? y|n")
			if !app.Confirmation("") {
				return
			}
		}
		reg, err := regexp.Compile("[^,a-fA-F0-9]+")
		if err != nil {
			log.Fatal(err)
		}
		rawChains := reg.ReplaceAllString(args[3], "")
		chains := strings.Split(rawChains, ",")
		serviceURI := args[4]
		fee, err := strconv.Atoi(args[7])
		if err != nil {
			fmt.Println(err)
			return
		}
		geozone, err := strconv.Atoi(args[6])
		if err != nil {
			fmt.Println(err)
			return
		}
		params := rpc.HeightAndKeyParams{
			Height: 0,
			Key:    "ServicerCountLock",
		}
		j, _ := json.Marshal(params)
		res, _ := QueryRPC(GetParamPath, j)
		if res == "true" {
			fmt.Println("Node Staking is Locked; 'ServicerCountLock' is activated to control inflated node count")
			return
		}
		fmt.Println("Enter Passphrase: ")
		res1, err := StakeNode(chains, serviceURI, operatorPubKey, output, app.Credentials(pwd), args[5], int64(geozone), types.NewInt(int64(amount)), int64(fee))
		if err != nil {
			fmt.Println(err)
			return
		}
		j1, err := json.Marshal(res1)
		if err != nil {
			fmt.Println(err)
			return
		}
		resp, err := QueryRPC(SendRawTxPath, j1)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(resp)
	},
}
