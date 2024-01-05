package cli

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/vipernet-xyz/viper-network/app"
	"github.com/vipernet-xyz/viper-network/types"
	governanceTypes "github.com/vipernet-xyz/viper-network/x/governance/types"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(governanceCmd)
	governanceCmd.AddCommand(governanceDAOTransfer)
	governanceCmd.AddCommand(governanceDAOBurn)
	governanceCmd.AddCommand(governanceChangeParam)
	governanceCmd.AddCommand(governanceUpgrade)
	governanceCmd.AddCommand(governanceFeatureEnable)
	governanceCmd.AddCommand(governanceGenDiscountKey)
}

var governanceCmd = &cobra.Command{
	Use:   "governance",
	Short: "governance management",
	Long: `The governance namespace handles all governance related interactions,
from DAOTransfer, change parameters; to performing protocol Upgrades. `,
}

func init() {
	governanceDAOTransfer.Flags().StringVar(&pwd, "pwd", "", "defines the passphrase used by the cmd non empty usage bypass interactive prompt ")
	governanceDAOBurn.Flags().StringVar(&pwd, "pwd", "", "defines the passphrase used by the cmd non empty usage bypass interactive prompt ")
	governanceChangeParam.Flags().StringVar(&pwd, "pwd", "", "defines the passphrase used by the cmd non empty usage bypass interactive prompt ")
	governanceUpgrade.Flags().StringVar(&pwd, "pwd", "", "defines the passphrase used by the cmd non empty usage bypass interactive prompt ")
}

var governanceDAOTransfer = &cobra.Command{
	Use:   "transfer <amount> <fromAddr> <toAddr> <networkID> <fees>",
	Short: "Transfer from DAO",
	Long: `If authorized, move funds from the DAO.
Actions: [burn, transfer]`,
	Args: cobra.ExactArgs(5),
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		toAddr := args[2]
		fromAddr := args[1]
		amount, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println(err)
			return
		}
		fees, err := strconv.Atoi(args[4])
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Enter Password: ")
		pass := app.Credentials(pwd)
		res, err := DAOTx(fromAddr, toAddr, pass, types.NewInt(int64(amount)), "dao_transfer", args[3], int64(fees), false)
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

var governanceDAOBurn = &cobra.Command{
	Use:   "burn <amount> <fromAddr> <toAddr> <networkID> <fees>",
	Short: "Burn from DAO",
	Long: `If authorized, burn funds from the DAO.
Actions: [burn, transfer]`,
	Args: cobra.ExactArgs(5),
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		var toAddr string
		if len(args) == 4 {
			toAddr = args[2]
		}
		fromAddr := args[1]
		amount, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println(err)
			return
		}
		fees, err := strconv.Atoi(args[4])
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Enter Password: ")
		pass := app.Credentials(pwd)
		res, err := DAOTx(fromAddr, toAddr, pass, types.NewInt(int64(amount)), "dao_burn", args[3], int64(fees), false)
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
var governanceChangeParam = &cobra.Command{
	Use:   "change_param <fromAddr> <networkID> <paramKey module/param> <paramValue (jsonObj)> <fees>",
	Short: "Edit a param in the network",
	Long: `If authorized, submit a tx to change any param from any module.
Will prompt the user for the <fromAddr> account passphrase.`,
	Args: cobra.ExactArgs(5),
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		fmt.Println("Enter Password: ")
		fees, err := strconv.Atoi(args[4])
		if err != nil {
			fmt.Println(err)
			return
		}

		res, err := ChangeParam(args[0], args[2], []byte(args[3]), app.Credentials(pwd), args[1], int64(fees), false)
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

var governanceUpgrade = &cobra.Command{
	Use:   "upgrade <fromAddr> <atHeight> <version> <networkID> <fees>",
	Short: "Upgrade the protocol",
	Long: `If authorized, upgrade the protocol.
Will prompt the user for the <fromAddr> account passphrase.`,
	Args: cobra.ExactArgs(5),
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		i, err := strconv.Atoi(args[1])
		if err != nil {
			log.Fatal(err)
		}
		u := governanceTypes.Upgrade{
			Height:  int64(i),
			Version: dropTag(args[2]),
		}
		fees, err := strconv.Atoi(args[4])
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("Enter Password: ")
		res, err := Upgrade(args[0], u, app.Credentials(pwd), args[3], int64(fees), false)
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

func dropTag(version string) string {
	if !strings.Contains(version, "-") {
		return version
	}
	s := strings.Split(version, "-")
	return s[1]
}

const FeatureUpgradeKey = "FEATURE"
const FeatureUpgradeHeight = int64(1)

var governanceFeatureEnable = &cobra.Command{
	Use:   "enable <fromAddr> <atHeight> <key> <networkID> <fees>",
	Short: "enable a protocol feature",
	Long: `If authorized, enable the protocol feature with the key.
Will prompt the user for the <fromAddr> account passphrase.`,
	Args: cobra.ExactArgs(5),
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		height := args[1]
		key := args[2]
		fstring := fmt.Sprintf("%s:%s", key, height)

		u := governanceTypes.Upgrade{
			Height:   FeatureUpgradeHeight,
			Version:  FeatureUpgradeKey,
			Features: []string{fstring},
		}
		fees, err := strconv.Atoi(args[4])
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("Enter Password: ")
		res, err := Upgrade(args[0], u, app.Credentials(pwd), args[3], int64(fees), false)
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

var governanceGenDiscountKey = &cobra.Command{
	Use:   "gendiscountkey <fromAddr> <toAddr> <chainID> <fees>",
	Short: "DAO generates a discount key for a requestor",
	Long:  `Only the DAO owner can use this to generate a unique discount key for a specific requestor for network usage discounts.`,
	Args:  cobra.ExactArgs(4),
	RunE: func(cmd *cobra.Command, args []string) error {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)

		fromAddr := args[0]
		toAddr := args[1]
		chainID := args[2]
		fees, err := strconv.Atoi(args[3])
		if err != nil {
			return err
		}

		fmt.Println("Enter DAO Owner's Password: ")
		passphrase := app.Credentials(pwd)

		// Generate and broadcast the discount key message
		res, err := GenerateAndSendDiscountKey(fromAddr, toAddr, passphrase, chainID, int64(fees), false)
		if err != nil {
			return err
		}

		j, err := json.Marshal(res)
		if err != nil {
			return err
		}
		resp, err := QueryRPC(SendRawTxPath, j)
		if err != nil {
			return err
		}
		fmt.Println(resp)
		return nil
	},
}
