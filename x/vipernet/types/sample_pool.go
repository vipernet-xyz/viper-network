package types

type RelayPool struct {
	Blockchain string
	Payloads   []*RelayPayload
}

var SampleRelayPools map[string]*RelayPool

func InitSampleRelayPool() {
	SampleRelayPools = make(map[string]*RelayPool)

	// Example: Adding sample relays for Ethereum
	ethRelayPool := &RelayPool{
		Blockchain: "0001", //Eth
		Payloads:   []*RelayPayload{ /* ... populate with Ethereum-specific relay payloads ... */ },
	}
	SampleRelayPools["0001"] = ethRelayPool

	// Repeat for other blockchains...
}
