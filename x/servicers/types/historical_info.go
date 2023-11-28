package types

import (
	bytes "bytes"
	"sort"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/vipernet-xyz/viper-network/codec"
)

// NewHistoricalInfo will create a historical information struct from header and valset
// it will first sort valset before inclusion into historical info
func NewHistoricalInfo(header abci.Header, valSet Validators) HistoricalInfo {
	// Must sort in the same way that tendermint does
	sort.SliceStable(valSet, func(i, j int) bool {
		return ValidatorsByVotingPower(valSet).Less(i, j)
	})
	return HistoricalInfo{
		Header: header,
		Valset: valSet,
	}
}

// MustUnmarshalHistoricalInfo wll unmarshal historical info and panic on error
func MustUnmarshalHistoricalInfo(cdc *codec.Codec, value []byte) HistoricalInfo {
	hi, err := UnmarshalHistoricalInfo(cdc, value)
	if err != nil {
		panic(err)
	}

	return hi
}

// UnmarshalHistoricalInfo will unmarshal historical info and return any error
func UnmarshalHistoricalInfo(cdc *codec.Codec, value []byte) (hi HistoricalInfo, err error) {
	err = cdc.Unmarshal(value, &hi)
	return hi, err
}

// ValidatorsByVotingPower implements sort.Interface for []Validator based on
// the VotingPower and Address fields.
// The validators are sorted first by their voting power (descending). Secondary index - Address (ascending).
// Copied from tendermint/types/validator_set.go
type ValidatorsByVotingPower []Validator

func (valz ValidatorsByVotingPower) Len() int { return len(valz) }

func (valz ValidatorsByVotingPower) Less(i, j int) bool {
	if valz[i].ConsensusPower() == valz[j].ConsensusPower() {
		addrI := valz[i].GetAddress()
		addrJ := valz[j].GetAddress()

		return bytes.Compare(addrI, addrJ) == -1
	}
	return valz[i].ConsensusPower() > valz[j].ConsensusPower()
}

func (valz ValidatorsByVotingPower) Swap(i, j int) {
	valz[i], valz[j] = valz[j], valz[i]
}
