package keyring

import (
	"github.com/vipernet-xyz/viper-network/codec"
	"github.com/vipernet-xyz/viper-network/codec/legacy"
	"github.com/vipernet-xyz/viper-network/crypto/hd"
)

func init() {
	RegisterLegacyAminoCodec(legacy.Cdc)
}

// RegisterLegacyAminoCodec registers concrete types and interfaces on the given codec.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterInterface((*LegacyInfo)(nil), nil)
	cdc.RegisterConcrete(hd.BIP44Params{}, "crypto/keys/hd/BIP44Params", nil)
	cdc.RegisterConcrete(legacyLocalInfo{}, "crypto/keys/localInfo", nil)
	cdc.RegisterConcrete(legacyLedgerInfo{}, "crypto/keys/ledgerInfo", nil)
	cdc.RegisterConcrete(legacyOfflineInfo{}, "crypto/keys/offlineInfo", nil)
	cdc.RegisterConcrete(LegacyMultiInfo{}, "crypto/keys/multiInfo", nil)
}
