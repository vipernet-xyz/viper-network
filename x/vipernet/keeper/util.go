package keeper

import (
	"github.com/vipernet-xyz/viper-network/crypto"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/vipernet/types"
)

// "GetPKFromFile" - Returns the private key object from a file
func (k Keeper) GetPKFromFile(ctx sdk.Ctx) (crypto.PrivateKey, error) {
	// get the Private validator key from the file
	pvKey, err := types.GetPVKeyFile()
	if err != nil {
		return nil, err
	}
	// convert the privKey to a private key object (compatible interface)
	pk, er := crypto.PrivKeyToPrivateKey(pvKey.PrivKey)
	if er != nil {
		return nil, er
	}
	return pk, nil
}
