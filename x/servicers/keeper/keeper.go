package keeper

import (
	"fmt"
	log2 "log"

	"github.com/vipernet-xyz/viper-network/codec"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/servicers/types"

	"github.com/tendermint/tendermint/libs/log"
)

// Implements ValidatorSet interface
var _ types.ValidatorSet = Keeper{}

// Keeper of the staking store
type Keeper struct {
	storeKey       sdk.StoreKey
	Cdc            *codec.Codec
	Bcdc           codec.BinaryCodec
	AccountKeeper  types.AuthKeeper
	ViperKeeper    types.ViperKeeper // todo combine all modules
	Paramstore     sdk.Subspace
	providerKeeper types.ProvidersKeeper
	// codespace
	codespace sdk.CodespaceType
	// Cache
	validatorCache *sdk.Cache
}

// NewKeeper creates a new staking Keeper instance
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, accountKeeper types.AuthKeeper,
	paramstore sdk.Subspace, codespace sdk.CodespaceType) Keeper {
	// ensure staked module accounts are set
	if addr := accountKeeper.GetModuleAddress(types.StakedPoolName); addr == nil {
		log2.Fatal(fmt.Errorf("%s module account has not been set", types.StakedPoolName))
	}
	cache := sdk.NewCache(int(types.ValidatorCacheSize))

	return Keeper{
		storeKey:       key,
		AccountKeeper:  accountKeeper,
		Paramstore:     paramstore.WithKeyTable(ParamKeyTable()),
		codespace:      codespace,
		validatorCache: cache,
		Cdc:            cdc,
	}
}

// Logger - returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Ctx) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// Codespace - Retrieve the codespace
func (k Keeper) Codespace() sdk.CodespaceType {
	return k.codespace
}

func (k Keeper) GetMsgStakeOutputSigner(ctx sdk.Ctx, msg sdk.Msg) sdk.Address {
	stakeMsg, ok := msg.(*types.MsgStake)
	if !ok {
		return nil
	}
	operatorAddr := sdk.Address(stakeMsg.PublicKey.Address())
	outputAddr, found := k.GetValidatorOutputAddress(ctx, operatorAddr)
	if !found {
		return nil
	}
	return outputAddr
}
