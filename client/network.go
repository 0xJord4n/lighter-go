package client

import "github.com/0xJord4n/lighter-go/types"

// Network represents a Lighter network (mainnet or testnet)
// This is re-exported from types package for convenience.
type Network = types.Network

const (
	// Mainnet is the Lighter mainnet network
	Mainnet = types.Mainnet
	// Testnet is the Lighter testnet network
	Testnet = types.Testnet
)
