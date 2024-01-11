package types

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	fp "path/filepath"
	"sync"

	sdk "github.com/vipernet-xyz/viper-network/types"
)

// SamplePool - An object that represents a sample pool for a blockchain
type SamplePool struct {
	Blockchain string         `json:"blockchain"`
	Payloads   []RelayPayload `json:"payloads"`
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
)

func LoadSampleRelayPool() (map[string]*RelayPool, error) {
	// create the sample pool path
	home, _ := os.UserHomeDir()
	var samplePoolPath = home + FS + sdk.DefaultDDName + FS + sdk.ConfigDirName + FS + "samplepool.json"
	// if file exists, open; else, create and open
	var jsonFile *os.File
	var bz []byte

	// reopen the file to read into the variable
	jsonFile, err := os.OpenFile(samplePoolPath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("Error opening samplepool.json: %v", err)
	}

	bz, err = ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, fmt.Errorf("Error reading samplepool.json: %v", err)
	}

	// close the file
	err = jsonFile.Close()
	if err != nil {
		return nil, fmt.Errorf("Error closing samplepool.json: %v", err)
	}

	// Unmarshal directly into the expected map structure
	var resultMap map[string]*RelayPool
	err = json.Unmarshal(bz, &resultMap)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshaling samplepool.json: %v", err)
	}

	return resultMap, nil
}
