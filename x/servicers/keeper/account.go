package keeper

import (
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/authentication"
)

// GetBalance - Retrieve balance for account
func (k Keeper) GetBalance(ctx sdk.Ctx, addr sdk.Address) sdk.BigInt {
	coins := k.AccountKeeper.GetCoins(ctx, addr)
	return coins.AmountOf(k.StakeDenom(ctx))
}

// GetAccount - Retrieve account info
func (k Keeper) GetAccount(ctx sdk.Ctx, addr sdk.Address) (acc *authentication.BaseAccount) {
	a := k.AccountKeeper.GetAccount(ctx, addr)
	if a == nil {
		return &authentication.BaseAccount{
			Address: sdk.Address{},
		}
	}
	return a.(*authentication.BaseAccount)
}

// SendCoins - Deliver coins to account
func (k Keeper) SendCoins(ctx sdk.Ctx, fromAddress sdk.Address, toAddress sdk.Address, amount sdk.BigInt) sdk.Error {
	coins := sdk.NewCoins(sdk.NewCoin(k.StakeDenom(ctx), amount))
	err := k.AccountKeeper.SendCoins(ctx, fromAddress, toAddress, coins)
	if err != nil {
		return err
	}
	return nil
}
