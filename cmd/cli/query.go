package cli

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/vipernet-xyz/viper-network/rpc"
	"github.com/vipernet-xyz/viper-network/types"
	types2 "github.com/vipernet-xyz/viper-network/x/requestors/types"

	"github.com/vipernet-xyz/viper-network/app"
	servicerTypes "github.com/vipernet-xyz/viper-network/x/servicers/types"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(queryCmd)
	queryCmd.AddCommand(queryBlock)
	queryCmd.AddCommand(queryHeight)
	queryCmd.AddCommand(queryTx)
	queryCmd.AddCommand(queryAccountTxs)
	queryCmd.AddCommand(queryBlockTxs)
	queryCmd.AddCommand(queryServicers)
	queryCmd.AddCommand(queryBalance)
	queryCmd.AddCommand(queryAccount)
	queryCmd.AddCommand(queryServicer)
	queryCmd.AddCommand(queryClients)
	queryCmd.AddCommand(queryClient)
	queryCmd.AddCommand(queryServicerParams)
	queryCmd.AddCommand(queryClientParams)
	queryCmd.AddCommand(queryServicerClaims)
	queryCmd.AddCommand(queryServicerClaim)
	queryCmd.AddCommand(queryViperParams)
	queryCmd.AddCommand(queryViperSupportedChains)
	queryCmd.AddCommand(querySupply)
	queryCmd.AddCommand(queryUpgrade)
	queryCmd.AddCommand(queryACL)
	queryCmd.AddCommand(queryAllParams)
	queryCmd.AddCommand(queryParam)
	queryCmd.AddCommand(queryDAOOwner)
	queryCmd.AddCommand(querySigningInfo)
}

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "query the blockchain",
	Long: `The query namespace handles all queryable interactions,
From getting Blocks, transactions, height; to getting params`,
}

var queryBlock = &cobra.Command{
	Use:   "block [<height>]",
	Short: "Get block at height",
	Long:  `Retrieves the block structure at the specified height.`,
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		var height int64
		if len(args) == 0 {
			height = 0
		} else {
			var err error
			parsed, err := strconv.Atoi(args[0])
			if err != nil {
				fmt.Println(err)
				return
			}
			convert := int64(parsed)
			height = convert
		}
		params := rpc.HeightParams{Height: height}
		j, err := json.Marshal(params)
		if err != nil {
			fmt.Println(err)
			return
		}
		res, err := QueryRPC(GetBlockPath, j)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(res)
	},
}

var prove bool

func init() {
	queryTx.LocalFlags().BoolVar(&prove, "proveTx", false, "would you like a proof of the transaction")
}

var queryTx = &cobra.Command{
	Use:   "tx <hash>",
	Short: "Get the transaction by the hash",
	Long:  `Retrieves the transaction by the hash`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		params := rpc.HashAndProveParams{Hash: args[0], Prove: prove}
		j, err := json.Marshal(params)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(GetTxPath)
		res, err := QueryRPC(GetTxPath, j)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(res)
	},
}

var queryAccountTxs = &cobra.Command{
	Use:   "account-txs <address> <page> <per_page> <prove (true | false)> <received (true | false)> <order (asc | desc)>",
	Short: "Get the transactions sent by the address, paginated by page and per_page",
	Long:  `Retrieves the transactions sent by the address`,
	Args:  cobra.RangeArgs(1, 6),
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		page := 0
		perPage := 0
		prove := false
		received := false
		order := "desc"
		if len(args) >= 2 {
			parsedPage, err := strconv.Atoi(args[1])
			if err == nil {
				page = parsedPage
			}
		}
		if len(args) >= 3 {
			parsedPerPage, err := strconv.Atoi(args[2])
			if err == nil {
				perPage = parsedPerPage
			}
		}
		if len(args) >= 4 {
			parsedProve, err := strconv.ParseBool(args[3])
			if err == nil {
				prove = parsedProve
			}
		}
		if len(args) >= 5 {
			parsedReceived, err := strconv.ParseBool(args[4])
			if err == nil {
				received = parsedReceived
			}
		}
		if len(args) >= 6 {
			parsedOrder := args[5]
			switch parsedOrder {
			case "asc":
				order = "asc"
			default:
				order = "desc"
			}
		}
		var err error
		params := rpc.PaginateAddrParams{
			Address:  args[0],
			Page:     page,
			PerPage:  perPage,
			Received: received,
			Prove:    prove,
			Sort:     order,
		}
		j, err := json.Marshal(params)
		if err != nil {
			fmt.Println(err)
			return
		}
		res, err := QueryRPC(GetAccountTxsPath, j)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(res)
	},
}

var queryBlockTxs = &cobra.Command{
	Use:   "block-txs <height> <page> <per_page> <prove (true | false)> <order (asc | desc)>",
	Short: "Get the transactions at a certain block height, paginated by page and per_page",
	Long:  `Retrieves the transactions in the block height`,
	Args:  cobra.RangeArgs(1, 5),
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		page := 0
		perPage := 0
		prove := false
		order := "desc"
		if len(args) >= 2 {
			parsedPage, err := strconv.Atoi(args[1])
			if err == nil {
				page = parsedPage
			}
		}
		if len(args) >= 3 {
			parsedPerPage, err := strconv.Atoi(args[2])
			if err == nil {
				perPage = parsedPerPage
			}
		}
		if len(args) >= 4 {
			parsedProve, err := strconv.ParseBool(args[3])
			if err == nil {
				prove = parsedProve
			}
		}
		if len(args) >= 5 {
			parsedOrder := args[4]
			switch parsedOrder {
			case "asc":
				order = "asc"
			default:
				order = "desc"
			}
		}
		height, parsingErr := strconv.ParseInt(args[0], 10, 64)
		if parsingErr != nil {
			fmt.Println(parsingErr)
			return
		}
		params := rpc.PaginatedHeightParams{
			Height:  height,
			Page:    page,
			PerPage: perPage,
			Prove:   prove,
			Sort:    order,
		}
		fmt.Println(params.Height, params.Page, params.PerPage, params.Sort)
		j, err := json.Marshal(params)
		if err != nil {
			fmt.Println(err)
			return
		}
		res, err := QueryRPC(GetBlockTxsPath, j)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(res)
	},
}

var queryHeight = &cobra.Command{
	Use:   "height",
	Short: "Get current height",
	Long:  `Retrieves the current height.`,
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		res, err := QueryRPC(GetHeightPath, []byte{})
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(res)
	},
}

var queryBalance = &cobra.Command{
	Use:   "balance <address> [<height>]",
	Short: "Gets account balance",
	Long:  `Retrieves the balance of the specified <accAddr> at the specified <height>.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		var height int
		if len(args) == 1 {
			height = 0 // latest
		} else {
			var err error
			height, err = strconv.Atoi(args[1])
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		params := rpc.HeightAndAddrParams{
			Height:  int64(height),
			Address: args[0],
		}
		j, err := json.Marshal(params)
		if err != nil {
			fmt.Println(err)
			return
		}
		res, err := QueryRPC(GetBalancePath, j)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(res)
	},
}

var queryAccount = &cobra.Command{
	Use:   "account <address> [<height>]",
	Short: "Gets an account",
	Long:  `Retrieves the account structure for a specific address.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		var height int
		if len(args) == 1 {
			height = 0 // latest
		} else {
			var err error
			height, err = strconv.Atoi(args[1])
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		params := rpc.HeightAndAddrParams{
			Height:  int64(height),
			Address: args[0],
		}
		j, err := json.Marshal(params)
		if err != nil {
			fmt.Println(err)
			return
		}
		res, err := QueryRPC(GetAccountPath, j)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(res)
	},
}

var servicerStakingStatus string
var servicerJailedStatus string
var blockchain string
var servicerPage int
var servicerLimit int

func init() {
	queryServicers.Flags().StringVar(&servicerStakingStatus, "staking-status", "", "the staking status of the servicer")
	queryServicers.Flags().StringVar(&servicerJailedStatus, "jailed-status", "", "the jailed status of the servicer")
	queryServicers.Flags().StringVar(&blockchain, "blockchain", "", "the relay chain identifiers these servicers support")
	queryServicers.Flags().IntVar(&servicerPage, "servicerPage", 1, "mark the servicerPage you want")
	queryServicers.Flags().IntVar(&servicerLimit, "servicerLimit", 10000, "reduce the amount of results")
}

// NOTE: flag "blockchain" is defined but not implemented at this time 2020/10/03

var queryServicers = &cobra.Command{
	Use:   "servicers [--staking-status (staked | unstaking)] [--jailed-status (jailed | unjailed)] [--blockchain <relayChainID>] [--servicerPage=<servicerPage>] [--servicerLimit=<servicerLimit>] [<height>]",
	Short: "Gets servicers",
	Long:  `Retrieves the list of all servicers known at the specified <height>.`,
	// Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		var height int
		if len(args) == 0 {
			height = 0 // latest
		} else {
			var err error
			height, err = strconv.Atoi(args[0])
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		var err error
		opts := servicerTypes.QueryValidatorsParams{
			Blockchain: blockchain,
			Page:       servicerPage,
			Limit:      servicerLimit,
		}
		if servicerStakingStatus != "" {
			switch strings.ToLower(servicerStakingStatus) {
			case "staked":
				opts.StakingStatus = types.Staked
			case "unstaking":
				opts.StakingStatus = types.Unstaking
			default:
				fmt.Println(fmt.Errorf("unkown staking status <staked or unstaking>"))
			}
		}
		if servicerJailedStatus != "" {
			switch strings.ToLower(servicerJailedStatus) {
			case "jailed":
				opts.JailedStatus = 1
			case "unjailed":
				opts.JailedStatus = 2
			default:
				fmt.Println(fmt.Errorf("unkown jailed status <jailed or unjailed>"))
			}
		}
		params := rpc.HeightAndValidatorOptsParams{
			Height: int64(height),
			Opts:   opts,
		}
		j, err := json.Marshal(params)
		if err != nil {
			fmt.Println(err)
			return
		}
		res, err := QueryRPC(GetServicersPath, j)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(res)
	},
}

var queryServicer = &cobra.Command{
	Use:   "servicer <address> [<height>]",
	Short: "Gets servicer from address",
	Long:  `Retrieves the servicer at the specified <height>.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		var height int
		if len(args) == 1 {
			height = 0 // latest
		} else {
			var err error
			height, err = strconv.Atoi(args[1])
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		params := rpc.HeightAndAddrParams{
			Height:  int64(height),
			Address: args[0],
		}
		j, err := json.Marshal(params)
		if err != nil {
			fmt.Println(err)
			return
		}
		res, err := QueryRPC(GetNodePath, j)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(res)
	},
}

var queryServicerParams = &cobra.Command{
	Use:   "servicer-params <height>",
	Short: "Gets servicer parameters",
	Long:  `Retrieves the servicer parameters at the specified <height>.`,
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		var height int
		if len(args) == 0 {
			height = 0 // latest
		} else {
			var err error
			height, err = strconv.Atoi(args[0])
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		params := rpc.HeightParams{
			Height: int64(height),
		}
		j, err := json.Marshal(params)
		if err != nil {
			fmt.Println(err)
			return
		}
		res, err := QueryRPC(GetNodeParamsPath, j)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(res)
	},
}

var clientStakingStatus string
var clientPage, clientLimit int

func init() {
	queryClients.Flags().StringVar(&servicerStakingStatus, "staking-status", "", "the staking status of the servicer")
	queryClients.Flags().IntVar(&clientPage, "clientPage", 1, "mark the page you want")
	queryClients.Flags().IntVar(&clientLimit, "clientLimit", 10000, "reduce the amount of results")
}

var queryClients = &cobra.Command{
	Use:   "requestors [--staking-status=<servicerStakingStatus>] [--clientPage=<clientPage>] [--servicerLimit=<servicerLimit>] [<height>]",
	Short: "Gets requestors",
	Long:  `Retrieves the list of all clients known at the specified <height>`,
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		var height int
		if len(args) == 0 {
			height = 0 // latest
		} else {
			var err error
			height, err = strconv.Atoi(args[0])
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		opts := types2.QueryRequestorsWithOpts{
			Blockchain: blockchain,
			Page:       clientPage,
			Limit:      clientLimit,
		}
		if clientStakingStatus != "" {
			switch strings.ToLower(servicerStakingStatus) {
			case "staked":
				opts.StakingStatus = types.Staked
			case "unstaking":
				opts.StakingStatus = types.Unstaking
			default:
				fmt.Println(fmt.Errorf("unkown staking status <staked or unstaking>"))
			}
		}
		params := rpc.HeightAndRequestorOptsParams{
			Height: int64(height),
			Opts:   opts,
		}
		j, err := json.Marshal(params)
		if err != nil {
			fmt.Println(err)
			return
		}
		res, err := QueryRPC(GetRequestorsPath, j)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(res)
	},
}

var queryClient = &cobra.Command{
	Use:   "client <address> [<height>]",
	Short: "Gets client from address",
	Long:  `Retrieves the client at the specified <height>.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		var height int
		if len(args) == 1 {
			height = 0 // latest
		} else {
			var err error
			height, err = strconv.Atoi(args[1])
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		params := rpc.HeightAndAddrParams{
			Height:  int64(height),
			Address: args[0],
		}
		j, err := json.Marshal(params)
		if err != nil {
			fmt.Println(err)
			return
		}
		res, err := QueryRPC(GetRequestorPath, j)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(res)
	},
}

var queryClientParams = &cobra.Command{
	Use:   "client-params [<height>]",
	Short: "Gets client parameters",
	Long:  `Retrieves the client parameters at the specified <height>.`,
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		var height int
		if len(args) == 0 {
			height = 0 // latest
		} else {
			var err error
			height, err = strconv.Atoi(args[0])
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		params := rpc.HeightParams{
			Height: int64(height),
		}
		j, err := json.Marshal(params)
		if err != nil {
			fmt.Println(err)
			return
		}
		res, err := QueryRPC(GetRequestorParamsPath, j)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(res)
	},
}

var queryServicerClaims = &cobra.Command{
	Use:   "servicer-claims <servicerAddr> [<height>]",
	Short: "Gets servicer pending claims for work completed",
	Long:  `Retrieves the list of all pending proof of work submitted by <servicerAddr> at <height>.`,
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		var err error
		var height int
		var address string
		switch len(args) {
		case 1:
			address = args[0]
		case 2:
			height, err = strconv.Atoi(args[1])
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		params := rpc.PaginatedHeightAndAddrParams{
			Height: int64(height),
			Addr:   address,
		}
		j, err := json.Marshal(params)
		if err != nil {
			fmt.Println(err)
			return
		}
		res, err := QueryRPC(GetNodeClaimsPath, j)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(res)
	},
}

var queryServicerClaim = &cobra.Command{
	Use:   "servicer-claim <address> <requestorPubKey> <claimType=(relay | challenge)> <relayChainID> <geoZoneID> <numOfServicers> <sessionHeight> [<height>]`",
	Short: "Gets servicer pending claim for work completed",
	Long:  `Gets servicer pending claim for verified proof of work submitted for a specific session`,
	Args:  cobra.MinimumNArgs(7),
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		var height int
		if len(args) == 7 {
			height = 0 // latest
		} else {
			var err error
			height, err = strconv.Atoi(args[7])
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		sessionheight, err := strconv.Atoi(args[6])
		if err != nil {
			fmt.Println(err)
			return
		}
		numServicers, err := strconv.Atoi(args[5])
		if err != nil {
			fmt.Println(err)
			return
		}
		params := rpc.QueryNodeReceiptParam{
			Address:         args[0],
			Blockchain:      args[3],
			GeoZone:         args[4],
			NumServicers:    int64(numServicers),
			RequestorPubkey: args[1],
			SBlockHeight:    int64(sessionheight),
			Height:          int64(height),
			ReceiptType:     args[2],
		}
		j, err := json.Marshal(params)
		if err != nil {
			fmt.Println(err)
			return
		}
		res, err := QueryRPC(GetNodeClaimPath, j)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(res)
	},
}

var queryViperParams = &cobra.Command{
	Use:   "viper-params [<height>]",
	Short: "Gets viper parameters",
	Long:  `Retrieves the viper parameters at the specified <height>.`,
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		var height int
		if len(args) == 0 {
			height = 0 // latest
		} else {
			var err error
			height, err = strconv.Atoi(args[0])
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		params := rpc.HeightParams{
			Height: int64(height),
		}
		j, err := json.Marshal(params)
		if err != nil {
			fmt.Println(err)
			return
		}
		res, err := QueryRPC(GetViperParamsPath, j)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(res)
	},
}

var queryViperSupportedChains = &cobra.Command{
	Use:   "supported-networks [<height>]",
	Short: "Gets viper supported relay chains",
	Long:  `Retrieves the list Relay Chain Identifiers supported by the network at the specified <height>`,
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		var height int
		if len(args) == 0 {
			height = 0 // latest
		} else {
			var err error
			height, err = strconv.Atoi(args[0])
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		params := rpc.HeightParams{
			Height: int64(height),
		}
		j, err := json.Marshal(params)
		if err != nil {
			fmt.Println(err)
			return
		}
		res, err := QueryRPC(GetSupportedChainsPath, j)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(res)
	},
}

var querySupply = &cobra.Command{
	Use:   "supply [<height>]",
	Short: "Gets the supply at <height>",
	Long:  `Retrieves the list of servicer params specified in the <height>`,
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		var height int
		if len(args) == 0 {
			height = 0 // latest
		} else {
			var err error
			height, err = strconv.Atoi(args[0])
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		params := rpc.HeightParams{
			Height: int64(height),
		}
		j, err := json.Marshal(params)
		if err != nil {
			fmt.Println(err)
			return
		}
		res, err := QueryRPC(GetSupplyPath, j)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(res)
	},
}

var queryDAOOwner = &cobra.Command{
	Use:   "daoOwner [<height>]",
	Short: "Gets the owner of the dao",
	Long:  `Retrieves the owner of the DAO (the account that can send/burn coins from the dao)`,
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		var height int
		if len(args) == 0 {
			height = 0 // latest
		} else {
			var err error
			height, err = strconv.Atoi(args[0])
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		params := rpc.HeightParams{
			Height: int64(height),
		}
		j, err := json.Marshal(params)
		if err != nil {
			fmt.Println(err)
			return
		}
		res, err := QueryRPC(GetDAOOwnerPath, j)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(res)
	},
}

var queryACL = &cobra.Command{
	Use:   "acl [<height>]",
	Short: "Gets the governance acl",
	Long:  `Retrieves the access control list of governance params (which account can change the param)`,
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		var height int
		if len(args) == 0 {
			height = 0 // latest
		} else {
			var err error
			height, err = strconv.Atoi(args[0])
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		params := rpc.HeightParams{
			Height: int64(height),
		}
		j, err := json.Marshal(params)
		if err != nil {
			fmt.Println(err)
			return
		}
		res, err := QueryRPC(GetACLPath, j)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(res)
	},
}

var queryAllParams = &cobra.Command{
	Use:   "params [<height>]",
	Short: "Gets all parameters",
	Long:  `Retrieves the parameters at the specified <height>.`,
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		var height int
		if len(args) == 0 {
			height = 0 // latest
		} else {
			var err error
			height, err = strconv.Atoi(args[0])
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		params := rpc.HeightParams{
			Height: int64(height),
		}
		j, err := json.Marshal(params)
		if err != nil {
			fmt.Println(err)
			return
		}
		res, err := QueryRPC(GetAllParamsPath, j)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(res)
	},
}

var queryParam = &cobra.Command{
	Use:   "param <key> [<height>]",
	Short: "Get a parameter with the given key",
	Long:  `Retrieves the parameter at the specified <height>.`,
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		var height int
		if len(args) == 1 {
			height = 0 // latest
		} else {
			var err error
			height, err = strconv.Atoi(args[1])
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		params := rpc.HeightAndKeyParams{
			Height: int64(height),
			Key:    args[0],
		}
		j, err := json.Marshal(params)
		if err != nil {
			fmt.Println(err)
			return
		}
		res, err := QueryRPC(GetParamPath, j)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(res)
	},
}

var queryUpgrade = &cobra.Command{
	Use:   "upgrade [<height>]",
	Short: "Gets the latest governance upgrade",
	Long:  `Retrieves the latest protocol upgrade by governance`,
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		var height int
		if len(args) == 0 {
			height = 0 // latest
		} else {
			var err error
			height, err = strconv.Atoi(args[0])
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		params := rpc.HeightParams{
			Height: int64(height),
		}
		j, err := json.Marshal(params)
		if err != nil {
			fmt.Println(err)
			return
		}
		res, err := QueryRPC(GetUpgradePath, j)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(res)
	},
}

var querySigningInfo = &cobra.Command{
	Use:   "signing-info <address> [<height>]",
	Short: "Gets validator signing info",
	Long:  `Retrieves the validator signing info with <address> at <height>.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		app.InitConfig(datadir, tmNode, persistentPeers, seeds, remoteCLIURL)
		var err error
		var height int
		var address string
		switch len(args) {
		case 1:
			address = args[0]
		case 2:
			address = args[0]
			height, err = strconv.Atoi(args[1])
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		params := rpc.PaginatedHeightAndAddrParams{
			Height: int64(height),
			Addr:   address,
		}
		j, err := json.Marshal(params)
		if err != nil {
			fmt.Println(err)
			return
		}
		res, err := QueryRPC(GetSigningInfoPath, j)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(res)
	},
}
