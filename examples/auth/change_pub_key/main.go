// Example: Changing the public key associated with an API key slot
package main

import (
	"fmt"
	"log"

	"github.com/0xJord4n/lighter-go/client"
	"github.com/0xJord4n/lighter-go/client/http"
	"github.com/0xJord4n/lighter-go/examples"
	"github.com/0xJord4n/lighter-go/signer"
	"github.com/0xJord4n/lighter-go/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func main() {
	privateKey := examples.GetPrivateKey()
	if privateKey == "" {
		log.Fatal("LIGHTER_PRIVATE_KEY environment variable not set")
	}

	// Ethereum private key for L1 signature (required for ChangePubKey)
	ethPrivateKey := examples.GetEthPrivateKey()
	if ethPrivateKey == "" {
		log.Fatal("LIGHTER_ETH_PRIVATE_KEY environment variable not set")
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

	// Create L1 signer for Ethereum signature
	l1Signer, err := signer.NewL1Signer(ethPrivateKey)
	if err != nil {
		log.Fatalf("Failed to create L1 signer: %v", err)
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

	// Sign with Ethereum key (L1 signature)
	l1Sig, err := l1Signer.Sign(txInfo.GetL1SignatureBody())
	if err != nil {
		log.Fatalf("Failed to sign L1 message: %v", err)
	}
	txInfo.SetL1Sig(l1Sig)

	fmt.Println("Change Pub Key Transaction Created!")
	fmt.Printf("  TX Hash: %s\n", txInfo.GetTxHash())
	fmt.Printf("  New Public Key: %s\n", newPublicKey)
	fmt.Printf("  From: %s\n", l1Signer.Address())

	// Submit to API
	resp, err := signerClient.SendAndSubmit(txInfo)
	if err != nil {
		log.Fatalf("Failed to submit transaction: %v", err)
	}

	fmt.Printf("  Submitted! TX Hash: %s\n", resp.TxHash)
}
