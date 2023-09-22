package types

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	fp "path/filepath"

	sdk "github.com/vipernet-xyz/viper-network/types"
)

type RelayPool struct {
	Blockchain string
	Payloads   []*RelayPayload
}

var GlobalConfig sdk.Config
var FS = string(fp.Separator)
var SampleRelayPools map[string]*RelayPool
var samplePoolPath = GlobalConfig.ViperConfig.DataDir + FS + sdk.ConfigDirName + FS + "samplepool.json"

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

func init() {
	err := LoadSampleRelayPool()
	if err != nil {
		log.Fatalf("Failed to initialize SampleRelayPools: %v", err)
	}
}
