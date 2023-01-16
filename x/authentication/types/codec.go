package types

import (
	"github.com/vipernet-xyz/viper-network/codec"
	"github.com/vipernet-xyz/viper-network/codec/types"
	"github.com/vipernet-xyz/viper-network/crypto"
	sdk "github.com/vipernet-xyz/viper-network/types"
	"github.com/vipernet-xyz/viper-network/x/authentication/exported"
)

// RegisterCodec registers concrete types on the codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface("x.authentication.ModuleAccount", (*exported.ModuleAccountI)(nil), &ModuleAccount{})
	cdc.RegisterInterface("x.authentication.Account", (*exported.Account)(nil), &BaseAccount{}, &ModuleAccount{})
	cdc.RegisterInterface("x.authentication.Supply", (*exported.SupplyI)(nil), &Supply{})
	cdc.RegisterStructure(&BaseAccount{}, "posmint/Account")
	cdc.RegisterStructure(StdTx{}, "posmint/StdTx")
	cdc.RegisterStructure(&Supply{}, "posmint/Supply")
	cdc.RegisterStructure(&ModuleAccount{}, "posmint/ModuleAccount")
	cdc.RegisterImplementation((*sdk.Tx)(nil), &StdTx{})
	ModuleCdc = cdc
}

// module wide codec
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.NewCodec(types.NewInterfaceRegistry())
	RegisterCodec(ModuleCdc)
	crypto.RegisterAmino(ModuleCdc.AminoCodec().Amino)
}
