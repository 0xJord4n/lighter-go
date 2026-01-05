// Example: Generating a new API key pair
package main

import (
	"fmt"
	"log"

	"github.com/0xJord4n/lighter-go/client"
)

func main() {
	// Generate a new API key pair
	// This creates a random private/public key pair for API authentication
	privateKey, publicKey, err := client.GenerateAPIKey()
	if err != nil {
		log.Fatalf("Failed to generate API key: %v", err)
	}

	fmt.Println("New API Key Generated!")
	fmt.Printf("  Private Key: %s\n", privateKey)
	fmt.Printf("  Public Key:  %s\n", publicKey)
	fmt.Println()
	fmt.Println("IMPORTANT: Save your private key securely!")
	fmt.Println("You'll need to register the public key on Lighter before using it.")
}
