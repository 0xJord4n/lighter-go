// Example: Transferring funds to another account
package main

import (
	"fmt"
	"log"

	"github.com/0xJord4n/lighter-go/client"
	"github.com/0xJord4n/lighter-go/client/http"
	"github.com/0xJord4n/lighter-go/examples"
	"github.com/0xJord4n/lighter-go/signer"
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

	// Create L1 signer for Ethereum signature
	l1Signer, err := signer.NewL1Signer(ethPrivateKey)
	if err != nil {
		log.Fatalf("Failed to create L1 signer: %v", err)
	}

	// Create a transfer to another account
	var memo [32]byte
	copy(memo[:], []byte("Payment for services"))

	req := &types.TransferTxReq{
		ToAccountIndex: 456,         // Recipient account index
		AssetIndex:     0,           // 0 = USDC
		FromRouteType:  0,           // Default route
		ToRouteType:    0,           // Default route
		Amount:         100_000000,  // 100 USDC (6 decimals)
		USDCFee:        0,           // Fee in USDC
		Memo:           memo,        // 32-byte memo
	}

	txInfo, err := signerClient.GetTransferTransaction(req, nil)
	if err != nil {
		log.Fatalf("Failed to create transfer transaction: %v", err)
	}

	// Sign with Ethereum key (L1 signature)
	l1Sig, err := l1Signer.Sign(txInfo.GetL1SignatureBody(chainId))
	if err != nil {
		log.Fatalf("Failed to sign L1 message: %v", err)
	}
	txInfo.SetL1Sig(l1Sig)

	fmt.Println("Transfer Transaction Created!")
	fmt.Printf("  TX Hash: %s\n", txInfo.GetTxHash())
	fmt.Printf("  From: %s\n", l1Signer.Address())
	fmt.Printf("  To Account: %d\n", req.ToAccountIndex)
	fmt.Printf("  Amount: %.2f USDC\n", float64(req.Amount)/1_000000)

	// Submit to API
	resp, err := signerClient.SendAndSubmit(txInfo)
	if err != nil {
		log.Fatalf("Failed to submit transaction: %v", err)
	}

	fmt.Printf("  Submitted! TX Hash: %s\n", resp.TxHash)
}
