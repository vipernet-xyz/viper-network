package cli

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/vipernet-xyz/viper-network/app"
	"github.com/vipernet-xyz/viper-network/crypto/keys/mintkey"
	"github.com/vipernet-xyz/viper-network/types"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(clientCmd)
	clientCmd.AddCommand(clientStakeCmd)
	clientCmd.AddCommand(clientUnstakeCmd)
	clientCmd.AddCommand(createAATCmd)
}

var clientCmd = &cobra.Command{
	Use:   "clients",
	Short: "client management",
	Long: `The clients namespace handles all client related interactions,
from staking and unstaking; to generating AATs.`,
}

func init() {
	clientStakeCmd.Flags().StringVar(&pwd, "pwd", "", "passphrase used by the cmd, non empty usage bypass interactive prompt")
	clientUnstakeCmd.Flags().StringVar(&pwd, "pwd", "", "passphrase used by the cmd, non empty usage bypass interactive prompt")
	createAATCmd.Flags().StringVar(&pwd, "pwd", "", "passphrase used by the cmd, non empty usage bypass interactive prompt")
}

var clientStakeCmd = &cobra.Command{
	Use:   "stake <fromAddr> <amount> <relayChainIDs> <networkID> <fee> ",
	Short: "Stake a client into the network",
	Long: `Stake the client into the network, giving it network throughput for the selected chains.
Will prompt the user for the <fromAddr> account passphrase. If the client is already staked, this transaction acts as an *update* transaction.
A client can updated relayChainIDs, and raise the stake/max_relays amount with this transaction.
If the client is currently staked at X and you submit an update with new stake Y. Only Y-X will be subtracted from an account
If no changes are desired for the parameter, just enter the current param value just as before`,
	Args: cobra.ExactArgs(5),
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		fromAddr := args[0]
		amount, err := strconv.Atoi(args[1])
		if err != nil {
			fmt.Println(err)
			return
		}
		reg, err := regexp.Compile("[^,a-fA-F0-9]+")
		if err != nil {
			log.Fatal(err)
		}
		fee, err := strconv.Atoi(args[4])
		if err != nil {
			fmt.Println(err)
			return
		}
		rawChains := reg.ReplaceAllString(args[2], "")
		chains := strings.Split(rawChains, ",")
		fmt.Println("Enter passphrase: ")
		res, err := StakeClient(chains, fromAddr, app.Credentials(pwd), args[3], types.NewInt(int64(amount)), int64(fee), false)
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

var clientUnstakeCmd = &cobra.Command{
	Use:   "unstake <fromAddr> <networkID> <fee>",
	Short: "Unstake a client from the network",
	Long: `Unstake a client from the network, changing it's status to Unstaking.
Prompts the user for the <fromAddr> account passphrase.`,
	Args: cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		fee, err := strconv.Atoi(args[2])
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Enter Password: ")
		res, err := UnstakeClient(args[0], app.Credentials(pwd), args[1], int64(fee), false)
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

var createAATCmd = &cobra.Command{
	Use:   "create-aat <clientAddr> <clientPubKey>",
	Short: "Creates an application authentication token",
	Long: `Creates a signed application authentication token (version 0.0.1 of the AAT spec), that can be embedded into application software for Relay servicing.
Will prompt the user for the <clientAddr> account passphrase.`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		kb := app.MustGetKeybase()
		if kb == nil {
			fmt.Println(app.UninitializedKeybaseError)
			return
		}
		addr, err := types.AddressFromHex(args[0])
		if err != nil {
			fmt.Printf("Address Error %s", err)
			return
		}
		kp, err := kb.Get(addr)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Enter passphrase: ")
		cred := app.Credentials(pwd)
		privkey, err := mintkey.UnarmorDecryptPrivKey(kp.PrivKeyArmor, cred)
		if err != nil {
			return
		}
		aat, err := app.GenerateAAT(hex.EncodeToString(kp.PublicKey.RawBytes()), args[1], privkey)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(string(aat))
	},
}
