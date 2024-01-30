package types

import (
	"sync"

	sdk "github.com/vipernet-xyz/viper-network/types"
)

// HostedBlockchain" - An object that represents a local hosted non-native blockchain
type HostedBlockchain struct {
	ID           string    `json:"id"`            // network identifier of the hosted blockchain
	HTTPURL      string    `json:"url"`           // url of the hosted blockchain
	WebSocketURL string    `json:"websocket_url"` // websocket URL for subscribing to events on the
	BasicAuth    BasicAuth `json:"basic_auth"`    // basic http authentication optinal
}

type BasicAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// HostedBlockchains" - An object that represents the local hosted non-native blockchains
type HostedBlockchains struct {
	M map[string]HostedBlockchain // M[addr] -> addr, url
	L sync.Mutex
}

// "Contains" - Checks to see if the hosted chain is within the HostedBlockchains object
func (c *HostedBlockchains) Contains(id string) bool {
	c.L.Lock()
	defer c.L.Unlock()
	// quick map check
	_, found := c.M[id]
	return found
}

func (c *HostedBlockchains) GetChain(id string) (chain HostedBlockchain, err sdk.Error) {
	c.L.Lock()
	defer c.L.Unlock()
	// map check
	res, found := c.M[id]
	if !found {
		return HostedBlockchain{}, NewErrorChainNotHostedError(ModuleName)
	}
	return res, nil
}

// GetChainURL returns the URL (HTTP or WebSocket) of the hosted blockchain using the hex network identifier
func (c *HostedBlockchains) GetChainHTTPURL(id string) (url string, err sdk.Error) {
	chain, err := c.GetChain(id)
	if err != nil {
		return "", err
	}
	return chain.HTTPURL, nil
}

// GetChainURL returns the URL (HTTP or WebSocket) of the hosted blockchain using the hex network identifier
func (c *HostedBlockchains) GetChainWebsocketURL(id string) (url string, err sdk.Error) {
	chain, err := c.GetChain(id)
	if err != nil {
		return "", err
	}
	return chain.WebSocketURL, nil
}

func (c *HostedBlockchains) Validate() error {
	c.L.Lock()
	defer c.L.Unlock()

	// Loop through all of the chains
	for _, chain := range c.M {
		// Validate not empty
		if chain.ID == "" || (chain.HTTPURL == "" && chain.WebSocketURL == "") {
			return NewInvalidHostedChainError(ModuleName)
		}

		// Validate the network identifier
		if err := NetworkIdentifierVerification(chain.ID); err != nil {
			return err
		}
	}
	return nil
}
