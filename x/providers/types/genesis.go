package types

// GenesisState - all staking state that must be provided at genesis
type GenesisState struct {
	Params    Params    `json:"params" yaml:"params"`
	Providers Providers `json:"providers" yaml:"providers"`
	Exported  bool      `json:"exported" yaml:"exported"`
}

// get raw genesis raw message for testing
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params:    DefaultParams(),
		Providers: make(Providers, 0),
	}
}
