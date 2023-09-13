package types

import (
	"github.com/vipernet-xyz/viper-network/codec"
)

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
