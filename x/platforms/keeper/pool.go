package keeper

import (
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/authentication"
	"github.com/vipernet-xyz/viper-network/x/authentication/exported"
	"github.com/vipernet-xyz/viper-network/x/platforms/types"
)

// StakedRatio - Retrieve the fraction of the staking tokens which are currently staked
func (k Keeper) StakedRatio(ctx sdk.Ctx) sdk.BigDec {
	stakedPool := k.GetStakedPool(ctx)

	stakeSupply := k.TotalTokens(ctx)
	if stakeSupply.IsPositive() {
		return stakedPool.GetCoins().AmountOf(k.StakeDenom(ctx)).ToDec().QuoInt(stakeSupply)
	}
	return sdk.ZeroDec()
}

// GetStakedTokens - Retrieve total staking tokens supply which is staked
func (k Keeper) GetStakedTokens(ctx sdk.Ctx) sdk.BigInt {
	stakedPool := k.GetStakedPool(ctx)
	return stakedPool.GetCoins().AmountOf(k.StakeDenom(ctx))
}

// TotalTokens - Retrieve total staking tokens from the total supply
func (k Keeper) TotalTokens(ctx sdk.Ctx) sdk.BigInt {
	return k.AccountKeeper.GetSupply(ctx).GetTotal().AmountOf(k.StakeDenom(ctx))
}

// GetStakedPool - Retrieve the staked tokens pool's module account
func (k Keeper) GetStakedPool(ctx sdk.Ctx) (stakedPool exported.ModuleAccountI) {
	return k.AccountKeeper.GetModuleAccount(ctx, types.StakedPoolName)
}

// coinsFromStakedToUnstkaed - Transfer coins from the module account to the platform -> used in unstaking
func (k Keeper) coinsFromStakedToUnstaked(ctx sdk.Ctx, platform types.Platform) sdk.Error {
	coins := sdk.NewCoins(sdk.NewCoin(k.StakeDenom(ctx), platform.StakedTokens))
	err := k.AccountKeeper.SendCoinsFromModuleToAccount(ctx, types.StakedPoolName, platform.Address, coins)
	if err != nil {
		return err
	}
	return nil
}

// coinsFromUnstakedToStaked - Transfer coins from the module account to platform -> used in staking
func (k Keeper) coinsFromUnstakedToStaked(ctx sdk.Ctx, platform types.Platform, amount sdk.BigInt) sdk.Error {
	if amount.LT(sdk.ZeroInt()) {
		return sdk.ErrInternal("cannot stake a negative amount of coins")
	}
	coins := sdk.NewCoins(sdk.NewCoin(k.StakeDenom(ctx), amount))
	err := k.AccountKeeper.SendCoinsFromAccountToModule(ctx, sdk.Address(platform.Address), types.StakedPoolName, coins)
	if err != nil {
		return err
	}
	return nil
}

// burnStakedTokens - Remove coins from the staked pool module account
func (k Keeper) burnStakedTokens(ctx sdk.Ctx, amt sdk.BigInt) sdk.Error {
	if !amt.IsPositive() {
		return nil
	}
	coins := sdk.NewCoins(sdk.NewCoin(k.StakeDenom(ctx), amt))
	return k.AccountKeeper.BurnCoins(ctx, types.StakedPoolName, coins)
}

// getFeePool - Retrieve fee pool
func (k Keeper) getFeePool(ctx sdk.Ctx) (feePool exported.ModuleAccountI) {
	return k.AccountKeeper.GetModuleAccount(ctx, authentication.FeeCollectorName)
}
