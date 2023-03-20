package client

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/vipernet-xyz/viper-network/types"
	govtypes "github.com/vipernet-xyz/viper-network/x/governance"

	ibcerrors "github.com/vipernet-xyz/viper-network/internal/errors"
	"github.com/vipernet-xyz/viper-network/modules/core/02-client/keeper"
	"github.com/vipernet-xyz/viper-network/modules/core/02-client/types"
)

// NewClientProposalHandler defines the 02-client proposal handler
func NewClientProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.ClientUpdateProposal:
			return k.ClientUpdateProposal(ctx, c)
		case *types.UpgradeProposal:
			return k.HandleUpgradeProposal(ctx, c)

		default:
			return errorsmod.Wrapf(ibcerrors.ErrUnknownRequest, "unrecognized ibc proposal content type: %T", c)
		}
	}
}
