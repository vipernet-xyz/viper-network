package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/vipernet-xyz/viper-network/x/viper-main/types"

	"github.com/vipernet-xyz/viper-network/app"

	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/state"
)

func init() {
	rootCmd.AddCommand(utilCmd)
	utilCmd.AddCommand(chainsGenCmd)
	utilCmd.AddCommand(chainsDelCmd)
	utilCmd.AddCommand(geoZonesGenCmd)
	utilCmd.AddCommand(geoZonesDelCmd)
	utilCmd.AddCommand(decodeTxCmd)
	utilCmd.AddCommand(exportGenesisForReset)
	utilCmd.AddCommand(convertViperEvidenceDB)
	utilCmd.AddCommand(completionCmd)
	utilCmd.AddCommand(updateConfigsCmd)
	utilCmd.AddCommand(printDefaultConfigCmd)
}

var utilCmd = &cobra.Command{
	Use:   "util",
	Short: "utility functions",
	Long:  `The util namespace handles all utility functions`,
}

var chainsGenCmd = &cobra.Command{
	Use:   "generate-chains",
	Short: "Generates chains file",
	Long:  `Generate the chains file for network identifiers`,
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		c := app.NewHostedChains(true)
		fmt.Println(app.GlobalConfig.ViperConfig.ChainsName + " contains: \n")
		for _, chain := range c.M {
			fmt.Printf("%s @ %s\n", chain.ID, chain.HTTPURL)
			if chain.WebSocketURL != "" {
				fmt.Printf("WebSocket: %s\n", chain.WebSocketURL)
			}
		}
		fmt.Println("If incorrect: please remove the chains.json with the " + chainsDelCmd.NameAndAliases() + " command")
	},
}

var chainsDelCmd = &cobra.Command{
	Use:   "delete-chains",
	Short: "Remove chains file.",
	Long:  `Remove the chains file for network identifiers.`,
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		app.DeleteHostedChains()
		fmt.Println("Successfully deleted " + app.GlobalConfig.ViperConfig.ChainsName + ".")
	},
}

var geoZonesGenCmd = &cobra.Command{
	Use:   "generate-geozone",
	Short: "Generates geoZone file",
	Long:  "Generate the geoZone file for hosted geoZone",
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		geoZones := app.NewHostedGeoZones(true)
		fmt.Println(app.GlobalConfig.ViperConfig.GeoZoneName + " contains:\n")
		for _, zone := range geoZones.M {
			fmt.Println(zone.ID)
		}
		fmt.Println("If incorrect, please remove the geozone.json file with the " + geoZonesDelCmd.NameAndAliases() + " command")
	},
}

var geoZonesDelCmd = &cobra.Command{
	Use:   "delete-geozones",
	Short: "Remove geoZones file",
	Long:  "Remove the geoZones file for hosted geoZones",
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		app.DeleteHostedGeoZone()
		fmt.Println("Successfully deleted " + app.GlobalConfig.ViperConfig.GeoZoneName + ".")
	},
}

var updateConfigsCmd = &cobra.Command{
	Use:   "update-configs",
	Short: "Updates config.json file.",
	Long:  `Update the config file with new params defaults for consensus/leveldbopts/p2p/cache/mempool/fastsync.`,
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		app.UpdateConfig(datadir)
		fmt.Println("Successfully updated config file.")
	},
}

var printDefaultConfigCmd = &cobra.Command{
	Use:   "print-configs",
	Short: "Prints Default config.json to console",
	Long:  `Prints Default config.json to console`,
	Run: func(cmd *cobra.Command, args []string) {
		c := app.GetDefaultConfig(datadir)
		fmt.Println(c)
	},
}

var decodeTxCmd = &cobra.Command{
	Use:   "decode-tx <tx> <legacyCodec>",
	Short: "Decodes a given transaction encoded in Amino/Proto base64 bytes",
	Long:  `Decodes a given transaction encoded in Amino/Proto base64 bytes`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		txStr := args[0]
		legacy := args[1]
		height := int64(-1)
		if legacy == "true" || legacy == "t" {
			height = 0
		}
		stdTx, err := app.UnmarshalTxStr(txStr, height)
		if err != nil {
			fmt.Println("an error occurred unmarshalling the transaction string", err.Error())
			return
		}
		fmt.Printf(
			"Type:\t\t%s\nMsg:\t\t%v\nFee:\t\t%s\nEntropy:\t%d\nMemo:\t\t%s\nSigners\t\t%v\nSig:\t\t%s\n",
			stdTx.GetMsg().Type(), stdTx.GetMsg(), stdTx.GetFee().String(), stdTx.GetEntropy(), stdTx.GetMemo(), stdTx.GetMsg().GetSigners(),
			stdTx.GetSignature().GetPublicKey())
	},
}

var exportGenesisForReset = &cobra.Command{
	Use:   "export-genesis-for-reset <height> <newChainID>",
	Short: "exports new genesis based on state",
	Long:  `In the event of a network reset, this will export a genesis file based on the previous state`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		height, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("error parsing height: ", err)
			return
		}
		db, err := app.OpenApplicationDB(app.GlobalConfig)
		if err != nil {
			fmt.Println("error loading application database: ", err)
			return
		}
		loggerFile, _ := os.Open(os.DevNull)
		a := app.NewViperCoreApp(nil, nil, nil, nil, nil, log.NewTMLogger(loggerFile), db, false, app.GlobalConfig.ViperConfig.IavlCacheSize)
		// initialize stores
		blockStore, _, _, _, err := state.BlocksAndStateFromDB(&app.GlobalConfig.TendermintConfig, state.DefaultDBProvider)
		if err != nil {
			fmt.Println("err loading blockstore: ", err.Error())
			return
		}
		a.SetBlockstore(blockStore)
		chainID := args[1]
		j, err := a.ExportState(int64(height), chainID)
		if err != nil {
			fmt.Println("could not export genesis state: ", err.Error())
			return
		}
		fmt.Println(j)
	},
}

var convertViperEvidenceDB = &cobra.Command{
	Use:   "convert-viper-evidence-db",
	Short: "convert viper evidence db to proto from amino",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		err := types.ConvertEvidenceToProto(app.GlobalConfig)
		if err != nil {
			fmt.Println("ERROR: ", err.Error())
			return
		}
		fmt.Println("Successfully converted evidence to proto")
	},
}

var (
	blocks bool
)

var completionCmd = &cobra.Command{
	Use:   "completion (bash | zsh | fish | powershell)",
	Short: "Generate completion script",
	Long: `To load completions:

Bash:

# Dependencies:
# bash 4.1+
# bash_completion from Kubectl

# add this to bash profile
$ source <(viper util completion bash)

# To load completions for each session, execute once:
Linux:
  $ viper util completion bash > /etc/bash_completion.d/viper
MacOS:
  $ viper util completion bash > /usr/local/etc/bash_completion.d/viper

Zsh:

# If shell completion is not already enabled in your environment you will need
# to enable it.  You can execute the following once:

$ echo "autoload -U compinit; compinit" >> ~/.zshrc

# To load completions for each session, execute once:
$ viper util completion zsh > "${fpath[1]}/viper"

# You will need to start a new shell for this setup to take effect.

Fish:

$ viper util completion fish | source

# To load completions for each session, execute once:
$ viper util completion fish > ~/.config/fish/completions/viper.fish
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactValidArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			_ = cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			_ = cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			_ = cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			_ = cmd.Root().GenPowerShellCompletion(os.Stdout)
		}
	},
}
