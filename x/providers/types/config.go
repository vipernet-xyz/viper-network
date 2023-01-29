package types

var ProviderCacheSize int64 = 5

func InitConfig(providerCacheSize int64) {
	ProviderCacheSize = providerCacheSize
}
