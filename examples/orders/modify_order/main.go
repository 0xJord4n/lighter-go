// Example: Modifying an existing order
package main

import (
	"fmt"
	"log"

	"github.com/0xJord4n/lighter-go/client"
	"github.com/0xJord4n/lighter-go/client/http"
	"github.com/0xJord4n/lighter-go/examples"
	"github.com/0xJord4n/lighter-go/types"
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

	// Modify an existing order
	// You need the order index from when the order was created
	req := &types.ModifyOrderTxReq{
		MarketIndex:  0,              // ETH-USD perp
		Index:        12345,          // Order index to modify
		BaseAmount:   2000000,        // New size: 0.02 ETH
		Price:        3100_000000,    // New price: $3100
		TriggerPrice: 0,              // For stop/TP orders
	}

	txInfo, err := signerClient.GetModifyOrderTransaction(req, nil)
	if err != nil {
		log.Fatalf("Failed to create modify order transaction: %v", err)
	}

	fmt.Println("Modify Order Transaction Created!")
	fmt.Printf("  TX Hash: %s\n", txInfo.GetTxHash())
	fmt.Printf("  Order Index: %d\n", req.Index)
	fmt.Printf("  New Size: %d\n", req.BaseAmount)
	fmt.Printf("  New Price: %d\n", req.Price)

	// Submit to API
	resp, err := signerClient.SendAndSubmit(txInfo)
	if err != nil {
		log.Fatalf("Failed to submit transaction: %v", err)
	}

	fmt.Printf("  Submitted! TX Hash: %s\n", resp.TxHash)
}
