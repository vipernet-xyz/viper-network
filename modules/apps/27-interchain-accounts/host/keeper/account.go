package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/vipernet-xyz/viper-network/types"
	authtypes "github.com/vipernet-xyz/viper-network/x/authentication/types"

	icatypes "github.com/vipernet-xyz/viper-network/modules/apps/27-interchain-accounts/types"
)

// createInterchainAccount creates a new interchain account. An address is generated using the host connectionID, the controller portID,
// and block dependent information. An error is returned if an account already exists for the generated account.
// An interchain account type is set in the account keeper and the interchain account address mapping is updated.
func (k Keeper) createInterchainAccount(ctx sdk.Ctx, connectionID, controllerPortID string) (sdk.Address, error) {
	Address := icatypes.GenerateAddress(ctx, connectionID, controllerPortID)

	if acc := k.accountKeeper.GetAccount(ctx, Address); acc != nil {
		return nil, errorsmod.Wrapf(icatypes.ErrAccountAlreadyExist, "existing account for newly generated interchain account address %s", Address)
	}
	baseAccount := authtypes.NewBaseAccountWithAddress(Address)
	interchainAccount := icatypes.NewInterchainAccount(&baseAccount, controllerPortID)

	k.accountKeeper.NewAccount(ctx, interchainAccount)
	k.accountKeeper.SetAccount(ctx, interchainAccount)

	k.SetInterchainAccountAddress(ctx, connectionID, controllerPortID, string(interchainAccount.Address))

	return Address, nil
}
