package pos

import (
	"fmt"
	"log"

	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/requestors/keeper"
	"github.com/vipernet-xyz/viper-network/x/requestors/types"
)

// InitGenesis sets up the module based on the genesis state
// First TM block is at height 1, so state updates requestorlied from
// genesis.json are in block 0.
func InitGenesis(ctx sdk.Ctx, keeper keeper.Keeper, supplyKeeper types.AuthKeeper, posKeeper types.PosKeeper, data types.GenesisState) {
	stakedTokens := sdk.ZeroInt()
	ctx = ctx.WithBlockHeight(1 - sdk.ValidatorUpdateDelay)
	// set the parameters from the data
	keeper.SetParams(ctx, data.Params)
	for _, requestor := range data.Requestors {
		if requestor.IsUnstaked() || requestor.IsUnstaking() {
			fmt.Println(fmt.Errorf("%v the requestors must be staked at genesis", requestor))
			continue
		}
		// calculate relays
		requestor.MaxRelays = keeper.CalculateRequestorRelays(ctx, requestor)
		// set the requestors from the data
		keeper.SetRequestor(ctx, requestor)
		if requestor.IsStaked() {
			stakedTokens = stakedTokens.Add(requestor.GetTokens())
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
			log.Fatal(fmt.Errorf("%s module account total does not equal the amount in each requestor account", types.StakedPoolName))
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
	requestors := keeper.GetAllRequestors(ctx)
	return types.GenesisState{
		Params:     params,
		Requestors: requestors,
		Exported:   true,
	}
}

// ValidateGenesis validates the provided staking genesis state to ensure the
// expected invariants holds. (i.e. params in correct bounds, no duplicate requestors)
func ValidateGenesis(data types.GenesisState) error {
	err := validateGenesisStateRequestors(data.Requestors, sdk.NewInt(data.Params.MinRequestorStake))
	if err != nil {
		return err
	}
	err = data.Params.Validate()
	if err != nil {
		return err
	}
	return nil
}

func validateGenesisStateRequestors(requestors []types.Requestor, minimumStake sdk.BigInt) (err error) {
	addrMap := make(map[string]bool, len(requestors))
	for i := 0; i < len(requestors); i++ {
		requestor := requestors[i]
		strKey := requestor.PublicKey.RawString()
		if _, ok := addrMap[strKey]; ok {
			return fmt.Errorf("duplicate requestor in genesis state: address %v", requestor.GetAddress())
		}
		if requestor.Jailed && requestor.IsStaked() {
			return fmt.Errorf("requestor is staked and jailed in genesis state: address %v", requestor.GetAddress())
		}
		if requestor.StakedTokens.IsZero() && !requestor.IsUnstaked() {
			return fmt.Errorf("staked/unstaked genesis requestor cannot have zero stake, requestor: %v", requestor)
		}
		addrMap[strKey] = true
		if !requestor.IsUnstaked() && requestor.StakedTokens.LTE(minimumStake) {
			return fmt.Errorf("requestor has less than minimum stake: %v", requestor)
		}
		for _, chain := range requestor.Chains {
			err := types.ValidateNetworkIdentifier(chain)
			if err != nil {
				return err
			}
		}
		for _, geoZone := range requestor.GeoZones {
			err := types.ValidateGeoZoneIdentifier(geoZone)
			if err != nil {
				return err
			}
		}
	}
	return
}
