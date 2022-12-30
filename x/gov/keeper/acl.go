package keeper

import (
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/gov/types"
)

func (k Keeper) VerifyACL(ctx sdk.Ctx, paramName string, owner sdk.Address) sdk.Error {
	acl := k.GetACL(ctx)
	o := acl.GetOwner(paramName)
	if !o.Equals(owner) {
		return types.ErrUnauthorizedParamChange(types.ModuleName, owner, paramName)
	}
	return nil
}
