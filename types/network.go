package types

// Network represents a Lighter network (mainnet or testnet)
type Network uint8

const (
	// Mainnet is the Lighter mainnet network
	Mainnet Network = iota
	// Testnet is the Lighter testnet network
	Testnet
)

// ChainID returns the chain ID for the network
func (n Network) ChainID() uint32 {
	switch n {
	case Mainnet:
		return 304
	case Testnet:
		return 300
	default:
		return 304 // Default to mainnet
	}
}

// APIURL returns the HTTP API URL for the network
func (n Network) APIURL() string {
	switch n {
	case Mainnet:
		return "https://mainnet.zklighter.elliot.ai"
	case Testnet:
		return "https://testnet.zklighter.elliot.ai"
	default:
		return "https://mainnet.zklighter.elliot.ai"
	}
}

// WSURL returns the WebSocket URL for the network
func (n Network) WSURL() string {
	switch n {
	case Mainnet:
		return "wss://mainnet.zklighter.elliot.ai/stream"
	case Testnet:
		return "wss://testnet.zklighter.elliot.ai/stream"
	default:
		return "wss://mainnet.zklighter.elliot.ai/stream"
	}
}

// String returns the network name
func (n Network) String() string {
	switch n {
	case Mainnet:
		return "mainnet"
	case Testnet:
		return "testnet"
	default:
		return "mainnet"
	}
}
