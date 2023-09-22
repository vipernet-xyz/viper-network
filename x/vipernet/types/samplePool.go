package types

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	fp "path/filepath"
	"sync"

	sdk "github.com/vipernet-xyz/viper-network/types"
)

// SamplePool - An object that represents a sample pool for a blockchain
type SamplePool struct {
	Blockchain string         `json:"blockchain"`
	Payloads   []RelayPayload `json:"payloads"` // Assuming RelayPayload is your payload structure
}

// SamplePools - An object that represents the sample pools hosted
type SamplePools struct {
	M map[string]SamplePool
	L sync.Mutex
}

// Contains - Checks if the sample pool exists within the HostedSamplePools object
func (sp *SamplePools) Contains(blockchain string) bool {
	sp.L.Lock()
	defer sp.L.Unlock()
	// Quick map check
	_, found := sp.M[blockchain]
	return found
}

// GetSamplePool - Returns the sample pool or an error using the blockchain identifier
func (sp *SamplePools) GetSamplePool(blockchain string) (pool SamplePool, err sdk.Error) {
	sp.L.Lock()
	defer sp.L.Unlock()
	// Map check
	res, found := sp.M[blockchain]
	if !found {
		return SamplePool{}, NewSampleNotHostedError(ModuleName)
	}
	return res, nil
}

// Validate - Validates the sample pool objects
func (sp *SamplePools) Validate() error {
	sp.L.Lock()
	defer sp.L.Unlock()
	// Loop through all the sample pools
	for _, pool := range sp.M {
		// Validate not empty
		if pool.Blockchain == "" {
			return NewInvalidSampleError(ModuleName)
		}
	}
	return nil
}

type RelayPool struct {
	Blockchain string
	Payloads   []*RelayPayload
}

var (
	GlobalConfig     sdk.Config
	FS               = string(fp.Separator)
	SampleRelayPools map[string]*RelayPool
	samplePoolPath   = GlobalConfig.ViperConfig.DataDir + FS + sdk.ConfigDirName + FS + GlobalConfig.ViperConfig.SamplePoolName
)

func LoadSampleRelayPool() error {
	// Initialize the SampleRelayPools map
	SampleRelayPools = make(map[string]*RelayPool)

	// Read the content of the samplepool.json
	fileContent, err := ioutil.ReadFile(samplePoolPath)
	if err != nil {
		return fmt.Errorf("Error reading samplepool.json: %v", err)
	}

	// Unmarshal the file content to the SampleRelayPools
	err = json.Unmarshal(fileContent, &SampleRelayPools)
	if err != nil {
		return fmt.Errorf("Error unmarshaling samplepool.json into SampleRelayPools: %v", err)
	}

	return nil
}
