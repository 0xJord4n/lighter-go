// Example: Creating a sub-account
package main

import (
	"fmt"
	"log"

	"github.com/0xJord4n/lighter-go/client"
	"github.com/0xJord4n/lighter-go/client/http"
	"github.com/0xJord4n/lighter-go/examples"
)

func main() {
	privateKey := examples.GetPrivateKey()
	if privateKey == "" {
		log.Fatal("LIGHTER_PRIVATE_KEY environment variable not set")
	}

	apiURL := examples.GetAPIURL()
	httpClient := http.NewFullClient(apiURL)

	chainId := uint32(1)
	apiKeyIndex := uint8(0)
	accountIndex := int64(1)

	signerClient, err := client.NewSignerClient(httpClient, privateKey, chainId, apiKeyIndex, accountIndex, nil)
	if err != nil {
		log.Fatalf("Failed to create signer client: %v", err)
	}

	// Create a new sub-account
	// Sub-accounts allow you to isolate trading activity
	txInfo, err := signerClient.GetCreateSubAccountTransaction(nil)
	if err != nil {
		log.Fatalf("Failed to create sub-account transaction: %v", err)
	}

	fmt.Println("Create Sub-Account Transaction Created!")
	fmt.Printf("  TX Hash: %s\n", txInfo.GetTxHash())
	fmt.Println()
	fmt.Println("Sub-accounts allow you to:")
	fmt.Println("  - Isolate trading strategies")
	fmt.Println("  - Separate margin requirements")
	fmt.Println("  - Track P&L independently")

	// Submit to API
	resp, err := signerClient.SendAndSubmit(txInfo)
	if err != nil {
		log.Fatalf("Failed to submit transaction: %v", err)
	}

	fmt.Printf("  Submitted! TX Hash: %s\n", resp.TxHash)
}
