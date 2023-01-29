package cli

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vipernet-xyz/viper-network/app"
	"github.com/vipernet-xyz/viper-network/rpc"

	"github.com/spf13/cobra"
)

var (
	datadir         string
	tmNode          string
	remoteCLIURL    string
	persistentPeers string
	seeds           string
	simulateRelay   bool
	keybase         bool
	mainnet         bool
	allBlockTxs     bool
	testnet         bool
	profileApp      bool
	useCache        bool
)

var CLIVersion = app.AppVersion

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "viper",
	Short: "Viper is a RPC relay protocol for web3.",
	Long: `               // // // // // // // // // // // // // // // // // 
	                   V I P E R  N E T W O R K
	        // // // // // // // // // // // // // // // // //
`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&datadir, "datadir", "", "data directory (default is $HOME/.github.com/vipernet-xyz/viper-network/")
	rootCmd.PersistentFlags().StringVar(&tmNode, "servicer", "", "takes a remote endpoint in the form <protocol>://<host>:<port>")
	rootCmd.PersistentFlags().StringVar(&remoteCLIURL, "remoteCLIURL", "", "takes a remote endpoint in the form of <protocol>://<host> (uses RPC Port)")
	rootCmd.PersistentFlags().StringVar(&persistentPeers, "persistent_peers", "", "a comma separated list of PeerURLs: '<ID>@<IP>:<PORT>,<ID2>@<IP2>:<PORT>...<IDn>@<IPn>:<PORT>'")
	rootCmd.PersistentFlags().StringVar(&seeds, "seeds", "", "a comma separated list of PeerURLs: '<ID>@<IP>:<PORT>,<ID2>@<IP2>:<PORT>...<IDn>@<IPn>:<PORT>'")
	startCmd.Flags().BoolVar(&simulateRelay, "simulateRelay", false, "would you like to be able to test your relays")
	startCmd.Flags().BoolVar(&keybase, "keybase", true, "run with keybase, if disabled allows you to stake for the current validator only. providing a keybase is still neccesary for staking for providers & sending transactions")
	startCmd.Flags().BoolVar(&mainnet, "mainnet", false, "run with mainnet genesis")
	startCmd.Flags().BoolVar(&allBlockTxs, "allblocktxs", false, "run with the allblocktxs endpoint (not recommended)")
	startCmd.Flags().BoolVar(&testnet, "testnet", false, "run with testnet genesis")
	startCmd.Flags().BoolVar(&profileApp, "profileApp", false, "expose cpu & memory profiling")
	startCmd.Flags().BoolVar(&useCache, "useCache", false, "use cache")
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(resetCmd)
	rootCmd.AddCommand(version)
	rootCmd.AddCommand(stopCmd)
}

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start [--keybase=(true | false)]",
	Short: "starts viper-core daemon",
	Long:  `Starts the Viper servicer, picks up the config from the assigned <datadir>`,
	Run: func(cmd *cobra.Command, args []string) {
		t := time.Unix(1625263200, 0) // Friday, July 2, 2021 6:00:00 PM GMT-04:00
		sleepDuration := time.Until(t)
		if time.Now().Before(t) {
			fmt.Println("Sleeping for ", sleepDuration)
			time.Sleep(sleepDuration)
		}
		start(cmd, args)
	},
}

func start(cmd *cobra.Command, args []string) {
	var genesisType app.GenesisType
	if mainnet && testnet {
		fmt.Println("cannot run with mainnet and testnet genesis simultaneously, please choose one")
		return
	}
	if mainnet {
		genesisType = app.MainnetGenesisType
	}
	if testnet {
		genesisType = app.TestnetGenesisType
	}
	tmNode := app.InitApp(datadir, tmNode, persistentPeers, seeds, remoteCLIURL, keybase, genesisType, useCache)
	go rpc.StartRPC(app.GlobalConfig.ViperConfig.RPCPort, app.GlobalConfig.ViperConfig.RPCTimeout, simulateRelay, profileApp, allBlockTxs, app.GlobalConfig.ViperConfig.ChainsHotReload)
	// trap kill signals (2,3,15,9)
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
		os.Kill, //nolint
		os.Interrupt)

	defer func() {
		sig := <-signalChannel
		app.ShutdownViperCore()
		err := tmNode.Stop()
		if err != nil {
			fmt.Println(err)
			return
		}
		message := fmt.Sprintf("Exit signal %s received\n", sig)
		fmt.Println(message)
		os.Exit(0)
	}()
}

// resetCmd represents the reset command
var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset viper-core",
	Long:  `Reset the Viper servicer daemon`,
	Run:   app.ResetWorldState,
}

var version = &cobra.Command{
	Use:   "version",
	Short: "Get current version",
	Long:  `Retrieves the version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("AppVersion: %s\n", CLIVersion)
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop viper-core",
	Long:  `Stop viper-core`,
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		res, err := QuerySecuredRPC(GetStopPath, []byte{}, app.GetAuthTokenFromFile())
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(res)
	},
}
