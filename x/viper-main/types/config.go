package types

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/tendermint/tendermint/config"
	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	"github.com/vipernet-xyz/viper-network/types"

	"github.com/tendermint/tendermint/libs/log"
)

const (
	DefaultRPCTimeout = 3000
	MaxRPCTimeout     = 1000000
	MinRPCTimeout     = 1
)

var (
	globalRPCTimeout       time.Duration
	GlobalViperConfig      types.ViperConfig
	GlobalTenderMintConfig config.Config
)

// "InitConfig" - Initializes the cache for sessions, test and evidence
func InitConfig(chains *HostedBlockchains, geozone *HostedGeoZones, logger log.Logger, c types.Config) {
	ConfigOnce.Do(func() {
		InitGlobalServiceMetric(chains, logger, c.ViperConfig.PrometheusAddr, c.ViperConfig.PrometheusMaxOpenfiles)
	})
	InitViperNodeCaches(c, logger)
	GlobalViperConfig = c.ViperConfig
	GlobalTenderMintConfig = c.TendermintConfig
	if GlobalViperConfig.LeanViper {
		GlobalTenderMintConfig.PrivValidatorState = types.DefaultPVSNameLean
		GlobalTenderMintConfig.PrivValidatorKey = types.DefaultPVKNameLean
		GlobalTenderMintConfig.NodeKey = types.DefaultPVSNameLean
	}
	SetRPCTimeout(c.ViperConfig.RPCTimeout)
}

func ConvertEvidenceToProto(config types.Config) error {
	node := AddViperNode(crypto.GenerateEd25519PrivKey().GenPrivateKey(), log.NewNopLogger())
	InitConfig(nil, nil, log.NewNopLogger(), config)

	gec := node.EvidenceStore
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

func FlushSessionCache() {
	for _, k := range GlobalViperNodes {
		if k.SessionStore != nil {
			err := k.SessionStore.FlushToDB()
			if err != nil {
				fmt.Printf("unable to flush sessions to the database before shutdown!! %s\n", err.Error())
			}
		}
		if k.EvidenceStore != nil {
			err := k.EvidenceStore.FlushToDB()
			if err != nil {
				fmt.Printf("unable to flush GOBEvidence to the database before shutdown!! %s\n", err.Error())
			}
		}
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
