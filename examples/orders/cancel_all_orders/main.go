// Example: Cancelling all orders
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/0xJord4n/lighter-go/client"
	"github.com/0xJord4n/lighter-go/client/http"
	"github.com/0xJord4n/lighter-go/examples"
	"github.com/0xJord4n/lighter-go/types"
	"github.com/0xJord4n/lighter-go/types/txtypes"
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

	// Cancel all GTT (Good-Till-Time) orders
	// TimeInForce options:
	//   - txtypes.GoodTillTime (1): Cancel GTT orders
	//   - txtypes.ImmediateOrCancel (0): Cancel IOC orders
	//   - txtypes.PostOnly (2): Cancel post-only orders
	req := &types.CancelAllOrdersTxReq{
		TimeInForce: txtypes.GoodTillTime,
		Time:        time.Now().UnixMilli(),
	}

	txInfo, err := signerClient.GetCancelAllOrdersTransaction(req, nil)
	if err != nil {
		log.Fatalf("Failed to create cancel all orders transaction: %v", err)
	}

	fmt.Println("Cancel All Orders Transaction Created!")
	fmt.Printf("  TX Hash: %s\n", txInfo.GetTxHash())
	fmt.Printf("  Time In Force: GoodTilTime\n")
	fmt.Printf("  Cancels all matching orders\n")

	// Submit to API
	resp, err := signerClient.SendAndSubmit(txInfo)
	if err != nil {
		log.Fatalf("Failed to submit transaction: %v", err)
	}

	fmt.Printf("  Submitted! TX Hash: %s\n", resp.TxHash)
}
