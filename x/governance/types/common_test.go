package types

import (
	"math/rand"

	"github.com/vipernet-xyz/viper-network/codec"
	"github.com/vipernet-xyz/viper-network/codec/types"
	"github.com/vipernet-xyz/viper-network/crypto"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/authentication"
)

// nolint: deadcode unused
// create a codec used only for testing
func makeTestCodec() *codec.Codec {
	var cdc = codec.NewCodec(types.NewInterfaceRegistry())
	authentication.RegisterCodec(cdc)
	RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	crypto.RegisterAmino(cdc.AminoCodec().Amino)
	return cdc
}

func getRandomPubKey() crypto.Ed25519PublicKey {
	var pub crypto.Ed25519PublicKey
	_, err := rand.Read(pub[:])
	if err != nil {
		_ = err
	}
	return pub
}

func getRandomValidatorAddress() sdk.Address {
	return sdk.Address(getRandomPubKey().Address())
}

var testACL ACL

func createTestACL() ACL {
	if testACL == nil {
		acl := ACL(make([]ACLPair, 0))
		acl.SetOwner("authentication/MaxMemoCharacters", getRandomValidatorAddress())
		acl.SetOwner("authentication/TxSigLimit", getRandomValidatorAddress())
		acl.SetOwner("governance/daoOwner", getRandomValidatorAddress())
		acl.SetOwner("governance/acl", getRandomValidatorAddress())
		acl.SetOwner("governance/upgrade", getRandomValidatorAddress())
		testACL = acl
	}
	return testACL
}

func createTestAdjacencyMap() map[string]bool {
	m := make(map[string]bool)
	m["authentication/MaxMemoCharacters"] = true // set
	m["authentication/TxSigLimit"] = true        // set
	m["governance/daoOwner"] = true              // set
	m["governance/acl"] = true                   // set
	m["governance/upgrade"] = true
	return m
}
