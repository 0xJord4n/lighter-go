// Package examples provides example code for using the lighter-go SDK.
package examples

import (
	"os"

	"github.com/0xJord4n/lighter-go/client"
	lighterhttp "github.com/0xJord4n/lighter-go/client/http"
	"github.com/0xJord4n/lighter-go/types"
)

// Environment variable names
const (
	EnvPrivateKey    = "LIGHTER_PRIVATE_KEY"
	EnvEthPrivateKey = "LIGHTER_ETH_PRIVATE_KEY"
	EnvNetwork       = "LIGHTER_NETWORK" // "mainnet" or "testnet"
)

// GetNetwork returns the network based on environment variable.
// Defaults to Mainnet if not set or invalid.
func GetNetwork() types.Network {
	network := os.Getenv(EnvNetwork)
	switch network {
	case "testnet":
		return types.Testnet
	default:
		return types.Mainnet
	}
}

// GetPrivateKey returns the private key from environment
func GetPrivateKey() string {
	return os.Getenv(EnvPrivateKey)
}

// GetEthPrivateKey returns the Ethereum private key from environment
// This is needed for L1 signatures (ChangePubKey and Transfer transactions)
func GetEthPrivateKey() string {
	return os.Getenv(EnvEthPrivateKey)
}

// CreateHTTPClient creates an HTTP client for the configured network
func CreateHTTPClient() client.FullHTTPClient {
	return lighterhttp.NewFullClientForNetwork(GetNetwork())
}

// CreateSignerClient creates a SignerClient for the configured network
func CreateSignerClient(privateKey string, apiKeyIndex uint8, accountIndex int64) (*client.SignerClient, error) {
	httpClient := CreateHTTPClient()
	network := GetNetwork()
	return client.NewSignerClient(httpClient, privateKey, network.ChainID(), apiKeyIndex, accountIndex, nil)
}
