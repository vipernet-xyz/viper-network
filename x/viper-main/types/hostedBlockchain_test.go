package types

import (
	"encoding/hex"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHostedBlockchains_GetChainURL(t *testing.T) {
	url := "https://www.google.com:443"
	wurl := "wss://www.google.com/ws"
	ethereum := hex.EncodeToString([]byte{01})
	testHostedBlockchain := HostedBlockchain{
		ID:           ethereum,
		HTTPURL:      url,
		WebSocketURL: wurl,
	}
	hb := HostedBlockchains{
		M: map[string]HostedBlockchain{testHostedBlockchain.ID: testHostedBlockchain},
		L: sync.Mutex{},
	}
	u, err := hb.GetChainHTTPURL(ethereum)
	w, err := hb.GetChainWebsocketURL(ethereum)
	assert.Nil(t, err)
	assert.Equal(t, u, url)
	assert.Equal(t, w, wurl)
}

func TestHostedBlockchains_ContainsFromString(t *testing.T) {
	url := "https://www.google.com:443"
	wurl := "wss://www.google.com/ws"
	ethereum := hex.EncodeToString([]byte{01})
	bitcoin := hex.EncodeToString([]byte{02})
	testHostedBlockchain := HostedBlockchain{
		ID:           ethereum,
		HTTPURL:      url,
		WebSocketURL: wurl,
	}
	hb := HostedBlockchains{
		M: map[string]HostedBlockchain{testHostedBlockchain.ID: testHostedBlockchain},
		L: sync.Mutex{},
	}
	assert.True(t, hb.Contains(ethereum))
	assert.False(t, hb.Contains(bitcoin))
}

func TestHostedBlockchains_Validate(t *testing.T) {
	url := "https://www.google.com:443"
	wurl := "wss://www.google.com/ws"
	ethereum := hex.EncodeToString([]byte{01})
	testHostedBlockchain := HostedBlockchain{
		ID:           ethereum,
		HTTPURL:      url,
		WebSocketURL: wurl,
	}
	HCNoURL := HostedBlockchain{
		ID:           ethereum,
		HTTPURL:      "",
		WebSocketURL: "",
	}
	HCNoHash := HostedBlockchain{
		ID:           "",
		HTTPURL:      url,
		WebSocketURL: wurl,
	}
	HCInvalidHash := HostedBlockchain{
		ID:           hex.EncodeToString([]byte("badlksajfljasdfklj")),
		HTTPURL:      url,
		WebSocketURL: wurl,
	}
	tests := []struct {
		name     string
		hc       *HostedBlockchains
		hasError bool
	}{
		{
			name:     "Invalid HostedBlockchain, no URL",
			hc:       &HostedBlockchains{M: map[string]HostedBlockchain{HCNoURL.HTTPURL: HCNoURL, HCNoURL.WebSocketURL: HCNoURL}, L: sync.Mutex{}},
			hasError: true,
		},
		{
			name:     "Invalid HostedBlockchain, no URL",
			hc:       &HostedBlockchains{M: map[string]HostedBlockchain{HCNoHash.HTTPURL: HCNoHash, HCNoHash.WebSocketURL: HCNoHash}, L: sync.Mutex{}},
			hasError: true,
		},
		{
			name:     "Invalid HostedBlockchain, invalid ID",
			hc:       &HostedBlockchains{M: map[string]HostedBlockchain{HCInvalidHash.HTTPURL: HCInvalidHash, HCInvalidHash.WebSocketURL: HCInvalidHash}, L: sync.Mutex{}},
			hasError: true,
		},
		{
			name:     "Valid HostedBlockchain",
			hc:       &HostedBlockchains{M: map[string]HostedBlockchain{testHostedBlockchain.ID: testHostedBlockchain}, L: sync.Mutex{}},
			hasError: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.hc.Validate() != nil, tt.hasError)
		})
	}
}
