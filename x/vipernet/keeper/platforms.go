package keeper

import (
	"github.com/vipernet-xyz/viper-network/crypto"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/platforms/exported"
)

// "GetPlatform" - Retrieves an platformlication from the platform store, using the platformKeeper (a link to the platforms module)
func (k Keeper) GetPlatform(ctx sdk.Ctx, address sdk.Address) (a exported.PlatformI, found bool) {
	a = k.platformKeeper.Platform(ctx, address)
	if a == nil {
		return a, false
	}
	return a, true
}

// "GetPlatformFromPublicKey" - Retrieves an platformlication from the platform store, using the platformKeeper (a link to the platforms module)
// using a hex string public key
func (k Keeper) GetPlatformFromPublicKey(ctx sdk.Ctx, pubKey string) (platform exported.PlatformI, found bool) {
	pk, err := crypto.NewPublicKey(pubKey)
	if err != nil {
		return nil, false
	}
	return k.GetPlatform(ctx, sdk.Address(pk.Address()))
}
