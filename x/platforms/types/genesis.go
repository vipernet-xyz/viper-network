package types

// GenesisState - all staking state that must be provided at genesis
type GenesisState struct {
	Params    Params    `json:"params" yaml:"params"`
	Platforms Platforms `json:"platforms" yaml:"platforms"`
	Exported  bool      `json:"exported" yaml:"exported"`
}

// get raw genesis raw message for testing
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params:    DefaultParams(),
		Platforms: make(Platforms, 0),
	}
}
