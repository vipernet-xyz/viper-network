package types

import (
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/vipernet-xyz/viper-network/types"

	"github.com/tendermint/tendermint/libs/log"
)

const (
	DefaultRPCTimeout = 3000
	MaxRPCTimeout     = 1000000
	MinRPCTimeout     = 1
)

var (
	globalRPCTimeout  time.Duration
	GlobalViperConfig types.ViperConfig
)

// "InitConfig" - Initializes the cache for sessions and evidence
func InitConfig(chains *HostedBlockchains, logger log.Logger, c types.Config) {
	cacheOnce.Do(func() {
		globalEvidenceCache = new(CacheStorage)
		globalSessionCache = new(CacheStorage)
		globalEvidenceSealedMap = sync.Map{}
		globalEvidenceCache.Init(c.ViperConfig.DataDir, c.ViperConfig.EvidenceDBName, c.TendermintConfig.LevelDBOptions, c.ViperConfig.MaxEvidenceCacheEntires, false)
		globalSessionCache.Init(c.ViperConfig.DataDir, "", c.TendermintConfig.LevelDBOptions, c.ViperConfig.MaxSessionCacheEntries, true)
		InitGlobalServiceMetric(chains, logger, c.ViperConfig.PrometheusAddr, c.ViperConfig.PrometheusMaxOpenfiles)
	})
	GlobalViperConfig = c.ViperConfig
	SetRPCTimeout(c.ViperConfig.RPCTimeout)
}

func ConvertEvidenceToProto(config types.Config) error {
	InitConfig(nil, log.NewNopLogger(), config)
	gec := globalEvidenceCache
	it, err := gec.Iterator()
	if err != nil {
		return fmt.Errorf("error creating evidence iterator: %s", err.Error())
	}
	defer it.Close()
	for ; it.Valid(); it.Next() {
		ev, err := Evidence{}.LegacyAminoUnmarshal(it.Value())
		if err != nil {
			return fmt.Errorf("error amino unmarshalling evidence: %s", err.Error())
		}
		k, err := ev.Key()
		if err != nil {
			return fmt.Errorf("error creating key from evidence object: %s", err.Error())
		}
		gec.SetWithoutLockAndSealCheck(hex.EncodeToString(k), ev)
	}
	err = gec.FlushToDBWithoutLock()
	if err != nil {
		return fmt.Errorf("error flushing evidence objects to the database: %s", err.Error())
	}
	return nil
}

// NOTE: evidence cache is flushed every time db iterator is created (every claim/proof submission)
func FlushSessionCache() {
	err := globalSessionCache.FlushToDB()
	if err != nil {
		fmt.Printf("unable to flush sessions to the database before shutdown!! %s\n", err.Error())
	}
	err = globalEvidenceCache.FlushToDB()
	if err != nil {
		fmt.Printf("unable to flush GOBEvidence to the database before shutdown!! %s\n", err.Error())
	}
}

func GetRPCTimeout() time.Duration {
	return globalRPCTimeout
}

func SetRPCTimeout(timeout int64) {
	if timeout < MinRPCTimeout || timeout > MaxRPCTimeout {
		timeout = DefaultRPCTimeout
	}

	globalRPCTimeout = time.Duration(timeout)
}
