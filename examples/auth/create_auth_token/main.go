// Example: Creating an auth token for private API endpoints
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/0xJord4n/lighter-go/examples"
)

func main() {
	privateKey := examples.GetPrivateKey()
	if privateKey == "" {
		log.Fatal("LIGHTER_PRIVATE_KEY environment variable not set")
	}

	// Create signer client (uses LIGHTER_NETWORK env var, defaults to mainnet)
	apiKeyIndex := uint8(0)
	accountIndex := int64(1) // Your account index

	signerClient, err := examples.CreateSignerClient(privateKey, apiKeyIndex, accountIndex)
	if err != nil {
		log.Fatalf("Failed to create signer client: %v", err)
	}

	network := examples.GetNetwork()
	fmt.Printf("Connected to %s (chain ID: %d)\n", network.String(), network.ChainID())

	// Create auth token with 7 hour expiry
	deadline := time.Now().Add(7 * time.Hour)
	authToken, err := signerClient.GetAuthToken(deadline)
	if err != nil {
		log.Fatalf("Failed to create auth token: %v", err)
	}

	fmt.Println("Auth Token Created!")
	fmt.Printf("  Token: %s\n", authToken)
	fmt.Printf("  Expires: %s\n", deadline.Format(time.RFC3339))
	fmt.Println()
	fmt.Println("Use this token for:")
	fmt.Println("  - WebSocket private channel subscriptions")
	fmt.Println("  - HTTP private endpoints (e.g., GetActiveOrders)")
}
