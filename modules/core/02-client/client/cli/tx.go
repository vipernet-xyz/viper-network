package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vipernet-xyz/viper-network/client"
	"github.com/vipernet-xyz/viper-network/client/flags"
	"github.com/vipernet-xyz/viper-network/client/tx"
	"github.com/vipernet-xyz/viper-network/codec"

	"github.com/vipernet-xyz/viper-network/modules/core/02-client/types"
	"github.com/vipernet-xyz/viper-network/modules/core/exported"
)

// NewCreateClientCmd defines the command to create a new IBC light client.
func NewCreateClientCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [path/to/client_state.json] [path/to/consensus_state.json]",
		Short: "create new IBC client",
		Long: `create a new IBC client with the specified client state and consensus state
	- ClientState JSON example: {"@type":"/ibc.lightclients.solomachine.v1.ClientState","sequence":"1","frozen_sequence":"0","consensus_state":{"public_key":{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"AtK50+5pJOoaa04qqAqrnyAqsYrwrR/INnA6UPIaYZlp"},"diversifier":"testing","timestamp":"10"},"allow_update_after_proposal":false}
	- ConsensusState JSON example: {"@type":"/ibc.lightclients.solomachine.v1.ConsensusState","public_key":{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"AtK50+5pJOoaa04qqAqrnyAqsYrwrR/INnA6UPIaYZlp"},"diversifier":"testing","timestamp":"10"}`,
		Example: fmt.Sprintf("%s tx ibc %s create [path/to/client_state.json] [path/to/consensus_state.json] --from node0 --home ../node0/<app>cli --chain-id $CID", 0.1, types.SubModuleName),
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			cdc := codec.NewProtoCodec1(clientCtx.InterfaceRegistry)

			// attempt to unmarshal client state argument
			var clientState exported.ClientState
			clientContentOrFileName := args[0]
			if err := cdc.UnmarshalInterfaceJSON([]byte(clientContentOrFileName), &clientState); err != nil {

				// check for file path if JSON input is not provided
				contents, err := os.ReadFile(clientContentOrFileName)
				if err != nil {
					return fmt.Errorf("neither JSON input nor path to .json file for client state were provided: %w", err)
				}

				if err := cdc.UnmarshalInterfaceJSON(contents, &clientState); err != nil {
					return fmt.Errorf("error unmarshalling client state file: %w", err)
				}
			}

			// attempt to unmarshal consensus state argument
			var consensusState exported.ConsensusState
			consensusContentOrFileName := args[1]
			if err := cdc.UnmarshalInterfaceJSON([]byte(consensusContentOrFileName), &consensusState); err != nil {

				// check for file path if JSON input is not provided
				contents, err := os.ReadFile(consensusContentOrFileName)
				if err != nil {
					return fmt.Errorf("neither JSON input nor path to .json file for consensus state were provided: %w", err)
				}

				if err := cdc.UnmarshalInterfaceJSON(contents, &consensusState); err != nil {
					return fmt.Errorf("error unmarshalling consensus state file: %w", err)
				}
			}

			msg, err := types.NewMsgCreateClient(clientState, consensusState, clientCtx.GetFromAddress().String())
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewUpdateClientCmd defines the command to update an IBC client.
func NewUpdateClientCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update [client-id] [path/to/client_msg.json]",
		Short:   "update existing client with a client message",
		Long:    "update existing client with a client message, for example a header, misbehaviour or batch update",
		Example: fmt.Sprintf("%s tx ibc %s update [client-id] [path/to/client_msg.json] --from node0 --home ../node0/<app>cli --chain-id $CID", 0.1, types.SubModuleName),
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			clientID := args[0]

			cdc := codec.NewProtoCodec1(clientCtx.InterfaceRegistry)

			var clientMsg exported.ClientMessage
			clientMsgContentOrFileName := args[1]
			if err := cdc.UnmarshalInterfaceJSON([]byte(clientMsgContentOrFileName), &clientMsg); err != nil {

				// check for file path if JSON input is not provided
				contents, err := os.ReadFile(clientMsgContentOrFileName)
				if err != nil {
					return fmt.Errorf("neither JSON input nor path to .json file for header were provided: %w", err)
				}

				if err := cdc.UnmarshalInterfaceJSON(contents, &clientMsg); err != nil {
					return fmt.Errorf("error unmarshalling header file: %w", err)
				}
			}

			msg, err := types.NewMsgUpdateClient(clientID, clientMsg, clientCtx.GetFromAddress().String())
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewSubmitMisbehaviourCmd defines the command to submit a misbehaviour to prevent
// future updates.
// Deprecated: NewSubmitMisbehaviourCmd is deprecated and will be removed in a future release.
// Please use NewUpdateClientCmd instead.
func NewSubmitMisbehaviourCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "misbehaviour [clientID] [path/to/misbehaviour.json]",
		Short:   "submit a client misbehaviour",
		Long:    "submit a client misbehaviour to prevent future updates",
		Example: fmt.Sprintf("%s tx ibc %s misbehaviour [clientID] [path/to/misbehaviour.json] --from node0 --home ../node0/<app>cli --chain-id $CID", 0.1, types.SubModuleName),
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			cdc := codec.NewProtoCodec1(clientCtx.InterfaceRegistry)

			var misbehaviour exported.ClientMessage
			clientID := args[0]
			misbehaviourContentOrFileName := args[1]
			if err := cdc.UnmarshalInterfaceJSON([]byte(misbehaviourContentOrFileName), &misbehaviour); err != nil {

				// check for file path if JSON input is not provided
				contents, err := os.ReadFile(misbehaviourContentOrFileName)
				if err != nil {
					return fmt.Errorf("neither JSON input nor path to .json file for misbehaviour were provided: %w", err)
				}

				if err := cdc.UnmarshalInterfaceJSON(contents, misbehaviour); err != nil {
					return fmt.Errorf("error unmarshalling misbehaviour file: %w", err)
				}
			}

			msg, err := types.NewMsgSubmitMisbehaviour(clientID, misbehaviour, clientCtx.GetFromAddress().String())
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewUpgradeClientCmd defines the command to upgrade an IBC light client.
func NewUpgradeClientCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade [client-identifier] [path/to/client_state.json] [path/to/consensus_state.json] [upgrade-client-proof] [upgrade-consensus-state-proof]",
		Short: "upgrade an IBC client",
		Long: `upgrade the IBC client associated with the provided client identifier while providing proof committed by the counterparty chain to the new client and consensus states
	- ClientState JSON example: {"@type":"/ibc.lightclients.solomachine.v1.ClientState","sequence":"1","frozen_sequence":"0","consensus_state":{"public_key":{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"AtK50+5pJOoaa04qqAqrnyAqsYrwrR/INnA6UPIaYZlp"},"diversifier":"testing","timestamp":"10"},"allow_update_after_proposal":false}
	- ConsensusState JSON example: {"@type":"/ibc.lightclients.solomachine.v1.ConsensusState","public_key":{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"AtK50+5pJOoaa04qqAqrnyAqsYrwrR/INnA6UPIaYZlp"},"diversifier":"testing","timestamp":"10"}`,
		Example: fmt.Sprintf("%s tx ibc %s upgrade [client-identifier] [path/to/client_state.json] [path/to/consensus_state.json] [client-state-proof] [consensus-state-proof] --from node0 --home ../node0/<app>cli --chain-id $CID", 0.1, types.SubModuleName),
		Args:    cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			cdc := codec.NewProtoCodec1(clientCtx.InterfaceRegistry)
			clientID := args[0]

			// attempt to unmarshal client state argument
			var clientState exported.ClientState
			clientContentOrFileName := args[1]
			if err := cdc.UnmarshalInterfaceJSON([]byte(clientContentOrFileName), &clientState); err != nil {

				// check for file path if JSON input is not provided
				contents, err := os.ReadFile(clientContentOrFileName)
				if err != nil {
					return fmt.Errorf("neither JSON input nor path to .json file for client state were provided: %w", err)
				}

				if err := cdc.UnmarshalInterfaceJSON(contents, &clientState); err != nil {
					return fmt.Errorf("error unmarshalling client state file: %w", err)
				}
			}

			// attempt to unmarshal consensus state argument
			var consensusState exported.ConsensusState
			consensusContentOrFileName := args[2]
			if err := cdc.UnmarshalInterfaceJSON([]byte(consensusContentOrFileName), &consensusState); err != nil {

				// check for file path if JSON input is not provided
				contents, err := os.ReadFile(consensusContentOrFileName)
				if err != nil {
					return fmt.Errorf("neither JSON input nor path to .json file for consensus state were provided: %w", err)
				}

				if err := cdc.UnmarshalInterfaceJSON(contents, &consensusState); err != nil {
					return fmt.Errorf("error unmarshalling consensus state file: %w", err)
				}
			}

			proofUpgradeClient := []byte(args[3])
			proofUpgradeConsensus := []byte(args[4])

			msg, err := types.NewMsgUpgradeClient(clientID, clientState, consensusState, proofUpgradeClient, proofUpgradeConsensus, clientCtx.GetFromAddress().String())
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}
