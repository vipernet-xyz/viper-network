package types

// GenesisState - all staking state that must be provided at genesis
type GenesisState struct {
	Params     Params     `json:"params" yaml:"params"`
	Requestors Requestors `json:"requestors" yaml:"requestors"`
	Exported   bool       `json:"exported" yaml:"exported"`
}

// get raw genesis raw message for testing
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params:     DefaultParams(),
		Requestors: make(Requestors, 0),
	}
}
