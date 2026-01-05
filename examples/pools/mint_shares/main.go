// Example: Minting shares in a public pool
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

	// Mint shares in a public pool (deposit)
	req := &types.MintSharesTxReq{
		PublicPoolIndex: 1,          // Pool to mint shares in
		ShareAmount:     100000,     // Number of shares to mint
	}

	txInfo, err := signerClient.GetMintSharesTransaction(req, nil)
	if err != nil {
		log.Fatalf("Failed to create mint shares transaction: %v", err)
	}

	fmt.Println("Mint Shares Transaction Created!")
	fmt.Printf("  TX Hash: %s\n", txInfo.GetTxHash())
	fmt.Printf("  Pool Index: %d\n", req.PublicPoolIndex)
	fmt.Printf("  Shares to Mint: %d\n", req.ShareAmount)
	fmt.Println()
	fmt.Println("Minting shares deposits funds into the pool and")
	fmt.Println("gives you proportional ownership of pool assets.")

	// Submit to API
	resp, err := signerClient.SendAndSubmit(txInfo)
	if err != nil {
		log.Fatalf("Failed to submit transaction: %v", err)
	}

	fmt.Printf("  Submitted! TX Hash: %s\n", resp.TxHash)
}
