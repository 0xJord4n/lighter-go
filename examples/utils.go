// Package examples provides example code for using the lighter-go SDK.
package examples

import (
	"os"
)

// Environment variable names
const (
	EnvPrivateKey    = "LIGHTER_PRIVATE_KEY"
	EnvEthPrivateKey = "LIGHTER_ETH_PRIVATE_KEY"
	EnvAPIURL        = "LIGHTER_API_URL"
	EnvWSURL         = "LIGHTER_WS_URL"
)

// DefaultAPIURL is the default mainnet API URL
const DefaultAPIURL = "https://mainnet.zklighter.elliot.ai"

// DefaultWSURL is the default mainnet WebSocket URL
const DefaultWSURL = "wss://mainnet.zklighter.elliot.ai/stream"

// GetPrivateKey returns the private key from environment
func GetPrivateKey() string {
	return os.Getenv(EnvPrivateKey)
}

// GetEthPrivateKey returns the Ethereum private key from environment
// This is needed for L1 signatures (ChangePubKey and Transfer transactions)
func GetEthPrivateKey() string {
	return os.Getenv(EnvEthPrivateKey)
}

// GetAPIURL returns the API URL from environment or default
func GetAPIURL() string {
	if url := os.Getenv(EnvAPIURL); url != "" {
		return url
	}
	return DefaultAPIURL
}

// GetWSURL returns the WebSocket URL from environment or default
func GetWSURL() string {
	if url := os.Getenv(EnvWSURL); url != "" {
		return url
	}
	return DefaultWSURL
}
