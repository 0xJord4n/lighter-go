// Example: Transferring funds to another account
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

	// Ethereum private key for L1 signature (required for transfers)
	ethPrivateKey := examples.GetEthPrivateKey()
	if ethPrivateKey == "" {
		log.Fatal("LIGHTER_ETH_PRIVATE_KEY environment variable not set")
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

	// Create a transfer to another account
	var memo [32]byte
	copy(memo[:], []byte("Payment for services"))

	req := &types.TransferTxReq{
		ToAccountIndex: 456,        // Recipient account index
		AssetIndex:     0,          // 0 = USDC
		FromRouteType:  0,          // Default route
		ToRouteType:    0,          // Default route
		Amount:         100_000000, // 100 USDC (6 decimals)
		USDCFee:        0,          // Fee in USDC
		Memo:           memo,       // 32-byte memo
	}

	// Transfer handles L1 signing internally - just pass the eth private key
	resp, err := signerClient.Transfer(ethPrivateKey, req, nil)
	if err != nil {
		log.Fatalf("Failed to submit transfer: %v", err)
	}

	fmt.Println("Transfer Submitted!")
	fmt.Printf("  TX Hash: %s\n", resp.TxHash)
	fmt.Printf("  To Account: %d\n", req.ToAccountIndex)
	fmt.Printf("  Amount: %.2f USDC\n", float64(req.Amount)/1_000000)
}
