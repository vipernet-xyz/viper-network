package types

import (
	types "github.com/tendermint/tendermint/abci/types"
	"github.com/vipernet-xyz/viper-network/codec"
)

// MustUnmarshalHistoricalInfo wll unmarshal historical info and panic on error
func MustUnmarshalHistoricalInfo(cdc codec.BinaryCodec, value []byte) HistoricalInfo {
	hi, err := UnmarshalHistoricalInfo(cdc, value)
	if err != nil {
		panic(err)
	}

	return hi
}

// UnmarshalHistoricalInfo will unmarshal historical info and return any error
func UnmarshalHistoricalInfo(cdc codec.BinaryCodec, value []byte) (hi HistoricalInfo, err error) {
	err = cdc.Unmarshal(value, &hi)
	return hi, err
}

type HistoricalInfo struct {
	Header types.Header `protobuf:"bytes,1,opt,name=header,proto3" json:"header"`
	Valset []Validator  `protobuf:"bytes,2,rep,name=valset,proto3" json:"valset"`
}
