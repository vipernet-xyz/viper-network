package types

import (
	"fmt"

	sdk "github.com/vipernet-xyz/viper-network/types"
)

type StakingKeyStore struct {
	key sdk.StoreKey
}

func NewStakingKeyStore(key sdk.StoreKey) *StakingKeyStore {
	return &StakingKeyStore{
		key: key,
	}
}

func (sks StakingKeyStore) SetStakingKey(ctx sdk.Ctx, address sdk.Address, stakingKey string) {
	store := ctx.KVStore(sks.key)
	store.Set([]byte(address.String()), []byte(stakingKey))
}

func (sks StakingKeyStore) GetStakingKey(ctx sdk.Ctx, address sdk.Address) (string, error) {
	store := ctx.KVStore(sks.key)
	key := []byte(address.String())
	res, _ := store.Has(key)
	if !res {
		return "", fmt.Errorf("staking key not found for address: %s", address.String())
	}
	value, _ := store.Get(key)
	return string(value), nil
}
