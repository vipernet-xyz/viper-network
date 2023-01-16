package pos

import (
	"fmt"
	"log"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/platforms/keeper"
	"github.com/vipernet-xyz/viper-network/x/platforms/types"
)

// InitGenesis sets up the module based on the genesis state
// First TM block is at height 1, so state updates platformlied from
// genesis.json are in block 0.
func InitGenesis(ctx sdk.Ctx, keeper keeper.Keeper, supplyKeeper types.AuthKeeper, posKeeper types.PosKeeper, data types.GenesisState) {
	stakedTokens := sdk.ZeroInt()
	ctx = ctx.WithBlockHeight(1 - sdk.ValidatorUpdateDelay)
	// set the parameters from the data
	keeper.SetParams(ctx, data.Params)
	for _, platform := range data.Platforms {
		if platform.IsUnstaked() || platform.IsUnstaking() {
			fmt.Println(fmt.Errorf("%v the platforms must be staked at genesis", platform))
			continue
		}
		// calculate relays
		platform.MaxRelays = keeper.CalculatePlatformRelays(ctx, platform)
		// set the platforms from the data
		keeper.SetPlatform(ctx, platform)
		if platform.IsStaked() {
			stakedTokens = stakedTokens.Add(platform.GetTokens())
		}
	}
	stakedCoins := sdk.NewCoins(sdk.NewCoin(posKeeper.StakeDenom(ctx), stakedTokens))
	// check if the staked pool accounts exists
	stakedPool := keeper.GetStakedPool(ctx)
	if stakedPool == nil {
		log.Fatal(fmt.Sprintf("%s module account has not been set", types.StakedPoolName))
	}
	// add coins if not provided on genesis
	if stakedPool.GetCoins().IsZero() {
		if err := stakedPool.SetCoins(stakedCoins); err != nil {
			log.Fatalf(fmt.Sprintf("error setting the coins for module account: %s module account", types.StakedPoolName))
		}
		supplyKeeper.SetModuleAccount(ctx, stakedPool)
	} else {
		if !stakedPool.GetCoins().IsEqual(stakedCoins) {
			log.Fatal(fmt.Errorf("%s module account total does not equal the amount in each platform account", types.StakedPoolName))
		}
	}
	// add coins to the total supply
	keeper.AccountKeeper.SetSupply(ctx, keeper.AccountKeeper.GetSupply(ctx).Inflate(stakedCoins))
	// set the params set in the keeper
	keeper.Paramstore.SetParamSet(ctx, &data.Params)
}

// ExportGenesis returns a GenesisState for a given context and keeper
func ExportGenesis(ctx sdk.Ctx, keeper keeper.Keeper) types.GenesisState {
	params := keeper.GetParams(ctx)
	platforms := keeper.GetAllPlatforms(ctx)
	return types.GenesisState{
		Params:    params,
		Platforms: platforms,
		Exported:  true,
	}
}

// ValidateGenesis validates the provided staking genesis state to ensure the
// expected invariants holds. (i.e. params in correct bounds, no duplicate platforms)
func ValidateGenesis(data types.GenesisState) error {
	err := validateGenesisStatePlatforms(data.Platforms, sdk.NewInt(data.Params.MinPlatformStake))
	if err != nil {
		return err
	}
	err = data.Params.Validate()
	if err != nil {
		return err
	}
	return nil
}

func validateGenesisStatePlatforms(platforms []types.Platform, minimumStake sdk.BigInt) (err error) {
	addrMap := make(map[string]bool, len(platforms))
	for i := 0; i < len(platforms); i++ {
		platform := platforms[i]
		strKey := platform.PublicKey.RawString()
		if _, ok := addrMap[strKey]; ok {
			return fmt.Errorf("duplicate platform in genesis state: address %v", platform.GetAddress())
		}
		if platform.Jailed && platform.IsStaked() {
			return fmt.Errorf("platform is staked and jailed in genesis state: address %v", platform.GetAddress())
		}
		if platform.StakedTokens.IsZero() && !platform.IsUnstaked() {
			return fmt.Errorf("staked/unstaked genesis platform cannot have zero stake, platform: %v", platform)
		}
		addrMap[strKey] = true
		if !platform.IsUnstaked() && platform.StakedTokens.LTE(minimumStake) {
			return fmt.Errorf("platform has less than minimum stake: %v", platform)
		}
		for _, chain := range platform.Chains {
			err := types.ValidateNetworkIdentifier(chain)
			if err != nil {
				return err
			}
		}
	}
	return
}
