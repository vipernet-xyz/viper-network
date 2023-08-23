package types

import (
	"fmt"
	"sync"

	"github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/privval"
	crypto "github.com/vipernet-xyz/viper-network/crypto/codec"
	"github.com/vipernet-xyz/viper-network/types"
	sdk "github.com/vipernet-xyz/viper-network/types"
)

// GlobalEvidenceCache & GlobalSessionCache is used for the first viper node and acts as backwards-compatibility for pre-lean viper
var GlobalEvidenceCache *CacheStorage
var GlobalSessionCache *CacheStorage

var GlobalViperNodes = map[string]*ViperNode{}

// ViperNode represents an entity in the network that is able to handle dispatches, servicing, challenges, and submit proofs/claims.
type ViperNode struct {
	PrivateKey      crypto.PrivateKey
	EvidenceStore   *CacheStorage
	SessionStore    *CacheStorage
	TestStore       *CacheStorage
	DoCacheInitOnce sync.Once
}

func (n *ViperNode) GetAddress() sdk.Address {
	return sdk.GetAddress(n.PrivateKey.PublicKey())
}

func AddViperNode(pk crypto.PrivateKey, logger log.Logger) *ViperNode {
	key := sdk.GetAddress(pk.PublicKey()).String()
	logger.Info("Adding " + key + " to list of viper nodes")
	node, exists := GlobalViperNodes[key]
	if exists {
		return node
	}
	node = &ViperNode{
		PrivateKey: pk,
	}
	GlobalViperNodes[key] = node
	return node
}

func AddViperNodeByFilePVKey(fpvKey privval.FilePVKey, logger log.Logger) {
	key, err := crypto.PrivKeyToPrivateKey(fpvKey.PrivKey)
	if err != nil {
		return
	}
	AddViperNode(key, logger)
}

// InitViperNodeCache adds a ViperNode with its SessionStore and EvidenceStore initialized
func InitViperNodeCache(node *ViperNode, c types.Config, logger log.Logger) {
	node.DoCacheInitOnce.Do(func() {
		evidenceDbName := c.ViperConfig.EvidenceDBName
		address := node.GetAddress().String()
		// In LeanViper, we create a evidence store on disk with suffix of the node's address
		if c.ViperConfig.LeanViper {
			evidenceDbName = evidenceDbName + "_" + address
		}
		logger.Info("Initializing " + address + " session and evidence cache")
		node.EvidenceStore = &CacheStorage{}
		node.SessionStore = &CacheStorage{}
		node.EvidenceStore.Init(c.ViperConfig.DataDir, evidenceDbName, c.TendermintConfig.LevelDBOptions, c.ViperConfig.MaxEvidenceCacheEntires, false)
		node.SessionStore.Init(c.ViperConfig.DataDir, "", c.TendermintConfig.LevelDBOptions, c.ViperConfig.MaxSessionCacheEntries, true)

		// Set the GOBSession and GOBEvidence Global for backwards compatibility for pre-LeanViper
		if GlobalSessionCache == nil {
			GlobalSessionCache = node.SessionStore
			GlobalEvidenceCache = node.EvidenceStore
		}
	})
}

func InitViperNodeCaches(c types.Config, logger log.Logger) {
	for _, node := range GlobalViperNodes {
		InitViperNodeCache(node, c, logger)
	}
}

// GetViperNodeByAddress returns a ViperNode from global map GlobalViperNodes
func GetViperNodeByAddress(address *sdk.Address) (*ViperNode, error) {
	node, ok := GlobalViperNodes[address.String()]
	if !ok {
		return nil, fmt.Errorf("failed to find private key for %s", address.String())
	}
	return node, nil
}

// CleanViperNodes sets the global viper nodes and its caches back to original state as if the node is starting up again.
// Cleaning up viper nodes is used for unit and integration tests where the cache is initialized in various scenarios (relays, tx, etc).
func CleanViperNodes() {
	for _, n := range GlobalViperNodes {
		if n == nil {
			continue
		}
		cacheToClean := []*CacheStorage{n.EvidenceStore, n.SessionStore}
		for _, r := range cacheToClean {
			if r == nil {
				continue
			}
			r.Clear()
			if r.DB == nil {
				continue
			}
			r.DB.Close()
		}
		GlobalEvidenceCache = nil
		GlobalSessionCache = nil
		GlobalViperNodes = map[string]*ViperNode{}
	}
}

// GetViperNode returns a ViperNode from global map GlobalViperNodes, it does not guarantee order
func GetViperNode() *ViperNode {
	for _, r := range GlobalViperNodes {
		if r != nil {
			return r
		}
	}
	return nil
}
