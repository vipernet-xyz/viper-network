package types

var RequestorCacheSize int64 = 5

func InitConfig(requestorCacheSize int64) {
	RequestorCacheSize = requestorCacheSize
}
