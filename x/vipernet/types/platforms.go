package types

import (
	"github.com/vipernet-xyz/viper-network/crypto"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/platforms/exported"
)

// "GetPlatformFromPublicKey" - Retrieves an platformlication from the platform store, using the platformKeeper (a link to the platforms module)
// using a hex string public key
func GetPlatformFromPublicKey(ctx sdk.Ctx, platformsKeeper PlatformsKeeper, pubKey string) (platform exported.PlatformI, found bool) {
	pk, err := crypto.NewPublicKey(pubKey)
	if err != nil {
		return nil, false
	}
	return GetPlatform(ctx, platformsKeeper, pk.Address().Bytes())
}

// "GetPlatform" - Retrieves an platformlication from the platform store, using the platformKeeper (a link to the platforms module)
func GetPlatform(ctx sdk.Ctx, platformsKeeper PlatformsKeeper, address sdk.Address) (a exported.PlatformI, found bool) {
	a = platformsKeeper.Platform(ctx, address)
	if a == nil {
		return a, false
	}
	return a, true
}
