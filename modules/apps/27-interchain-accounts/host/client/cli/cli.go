package cli

import (
	"github.com/spf13/cobra"
	"github.com/vipernet-xyz/viper-network/client"
)

// GetQueryCmd returns the query commands for the ICA host submodule
func GetQueryCmd() *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        "host",
		Short:                      "IBC interchain accounts host query subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
	}

	queryCmd.AddCommand(
		GetCmdParams(),
	)

	return queryCmd
}

// NewTxCmd creates and returns the tx command
func NewTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "host",
		Short:                      "IBC interchain accounts host transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		generatePacketDataCmd(),
	)

	return cmd
}
