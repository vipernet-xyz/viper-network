package pos

import (
	"fmt"
	"log"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/providers/keeper"
	"github.com/vipernet-xyz/viper-network/x/providers/types"
)

// InitGenesis sets up the module based on the genesis state
// First TM block is at height 1, so state updates providerlied from
// genesis.json are in block 0.
func InitGenesis(ctx sdk.Ctx, keeper keeper.Keeper, supplyKeeper types.AuthKeeper, posKeeper types.PosKeeper, data types.GenesisState) {
	stakedTokens := sdk.ZeroInt()
	ctx = ctx.WithBlockHeight(1 - sdk.ValidatorUpdateDelay)
	// set the parameters from the data
	keeper.SetParams(ctx, data.Params)
	for _, provider := range data.Providers {
		if provider.IsUnstaked() || provider.IsUnstaking() {
			fmt.Println(fmt.Errorf("%v the providers must be staked at genesis", provider))
			continue
		}
		// calculate relays
		provider.MaxRelays = keeper.CalculateProviderRelays(ctx, provider)
		// set the providers from the data
		keeper.SetProvider(ctx, provider)
		if provider.IsStaked() {
			stakedTokens = stakedTokens.Add(provider.GetTokens())
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
			log.Fatal(fmt.Errorf("%s module account total does not equal the amount in each provider account", types.StakedPoolName))
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
	providers := keeper.GetAllProviders(ctx)
	return types.GenesisState{
		Params:    params,
		Providers: providers,
		Exported:  true,
	}
}

// ValidateGenesis validates the provided staking genesis state to ensure the
// expected invariants holds. (i.e. params in correct bounds, no duplicate providers)
func ValidateGenesis(data types.GenesisState) error {
	err := validateGenesisStateProviders(data.Providers, sdk.NewInt(data.Params.MinProviderStake))
	if err != nil {
		return err
	}
	err = data.Params.Validate()
	if err != nil {
		return err
	}
	return nil
}

func validateGenesisStateProviders(providers []types.Provider, minimumStake sdk.BigInt) (err error) {
	addrMap := make(map[string]bool, len(providers))
	for i := 0; i < len(providers); i++ {
		provider := providers[i]
		strKey := provider.PublicKey.RawString()
		if _, ok := addrMap[strKey]; ok {
			return fmt.Errorf("duplicate provider in genesis state: address %v", provider.GetAddress())
		}
		if provider.Jailed && provider.IsStaked() {
			return fmt.Errorf("provider is staked and jailed in genesis state: address %v", provider.GetAddress())
		}
		if provider.StakedTokens.IsZero() && !provider.IsUnstaked() {
			return fmt.Errorf("staked/unstaked genesis provider cannot have zero stake, provider: %v", provider)
		}
		addrMap[strKey] = true
		if !provider.IsUnstaked() && provider.StakedTokens.LTE(minimumStake) {
			return fmt.Errorf("provider has less than minimum stake: %v", provider)
		}
		for _, chain := range provider.Chains {
			err := types.ValidateNetworkIdentifier(chain)
			if err != nil {
				return err
			}
		}
	}
	return
}
