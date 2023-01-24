package keeper

// DONTCOVER

import (
	"github.com/vipernet-xyz/viper-network/codec"
	cdcTypes "github.com/vipernet-xyz/viper-network/codec/types"
	"github.com/vipernet-xyz/viper-network/crypto"
	"github.com/vipernet-xyz/viper-network/store"
	sdk "github.com/vipernet-xyz/viper-network/types"
	authTypes "github.com/vipernet-xyz/viper-network/x/authentication/types"
	governanceKeeper "github.com/vipernet-xyz/viper-network/x/governance/keeper"
	governanceTypes "github.com/vipernet-xyz/viper-network/x/governance/types"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

type testInput struct {
	cdc    *codec.Codec
	ctx    sdk.Context
	Keeper Keeper
}

func setupTestInput() testInput {
	db := dbm.NewMemDB()

	cdc := codec.NewCodec(cdcTypes.NewInterfaceRegistry())
	authTypes.RegisterCodec(cdc)
	crypto.RegisterAmino(cdc.AminoCodec().Amino)

	authCapKey := sdk.NewKVStoreKey("authentication")
	keyParams := sdk.ParamsKey
	tkeyParams := sdk.ParamsTKey

	ms := store.NewCommitMultiStore(db, false, 5000000)
	ms.MountStoreWithDB(authCapKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	_ = ms.LoadLatestVersion()
	akSubspace := sdk.NewSubspace(authTypes.DefaultCodespace)
	ak := NewKeeper(
		cdc, authCapKey, akSubspace, nil,
	)
	governanceKeeper.NewKeeper(cdc, sdk.ParamsKey, sdk.ParamsTKey, governanceTypes.DefaultCodespace, ak, akSubspace)
	ctx := sdk.NewContext(ms, abci.Header{ChainID: "test-chain-id"}, false, log.NewNopLogger())
	ak.SetParams(ctx, authTypes.DefaultParams())
	return testInput{Keeper: ak, cdc: cdc, ctx: ctx}
}
