// Example: Changing the public key associated with an API key slot
package main

import (
	"fmt"
	"log"

	"github.com/0xJord4n/lighter-go/client"
	"github.com/0xJord4n/lighter-go/client/http"
	"github.com/0xJord4n/lighter-go/examples"
	"github.com/0xJord4n/lighter-go/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
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
	accountIndex := int64(1) // Your account index

	signerClient, err := client.NewSignerClient(httpClient, privateKey, chainId, apiKeyIndex, accountIndex, nil)
	if err != nil {
		log.Fatalf("Failed to create signer client: %v", err)
	}

	// Generate a new API key to register
	_, newPublicKey, err := client.GenerateAPIKey()
	if err != nil {
		log.Fatalf("Failed to generate new API key: %v", err)
	}

	// Decode public key from hex to bytes
	pubKeyBytes, err := hexutil.Decode(newPublicKey)
	if err != nil {
		log.Fatalf("Failed to decode public key: %v", err)
	}

	if len(pubKeyBytes) != 40 {
		log.Fatalf("Invalid public key length: expected 40 bytes, got %d", len(pubKeyBytes))
	}

	var pubKey [40]byte
	copy(pubKey[:], pubKeyBytes)

	// Create change pub key transaction
	req := &types.ChangePubKeyReq{
		PubKey: pubKey,
	}

	txInfo, err := signerClient.GetChangePubKeyTransaction(req, nil)
	if err != nil {
		log.Fatalf("Failed to create change pub key transaction: %v", err)
	}

	fmt.Println("Change Pub Key Transaction Created!")
	fmt.Printf("  TX Hash: %s\n", txInfo.GetTxHash())
	fmt.Printf("  New Public Key: %s\n", newPublicKey)
	fmt.Println()
	fmt.Println("NOTE: This transaction requires L1 signature for security.")
	fmt.Printf("  Message to sign: %s\n", txInfo.GetL1SignatureBody())

	// Submit to API
	resp, err := signerClient.SendAndSubmit(txInfo)
	if err != nil {
		log.Fatalf("Failed to submit transaction: %v", err)
	}

	fmt.Printf("  Submitted! TX Hash: %s\n", resp.TxHash)
}
