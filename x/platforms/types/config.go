package types

var PlatformCacheSize int64 = 5

func InitConfig(platformCacheSize int64) {
	PlatformCacheSize = platformCacheSize
}
