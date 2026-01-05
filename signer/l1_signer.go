package signer

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

// L1Signer handles Ethereum L1 signatures for transactions that require them
// (ChangePubKey and Transfer)
type L1Signer struct {
	privateKey *ecdsa.PrivateKey
}

// NewL1Signer creates a new L1Signer from a hex-encoded Ethereum private key
func NewL1Signer(privateKeyHex string) (*L1Signer, error) {
	// Remove 0x prefix if present
	if len(privateKeyHex) >= 2 && privateKeyHex[:2] == "0x" {
		privateKeyHex = privateKeyHex[2:]
	}

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid ethereum private key: %w", err)
	}

	return &L1Signer{privateKey: privateKey}, nil
}

// Sign signs a message using EIP-191 personal sign and returns the hex-encoded signature
func (s *L1Signer) Sign(message string) (string, error) {
	// Hash the message using EIP-191 personal sign
	hash := accounts.TextHash([]byte(message))

	// Sign the hash
	signature, err := crypto.Sign(hash, s.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign message: %w", err)
	}

	// Transform V from 0/1 to 27/28 (Ethereum yellow paper format)
	if signature[64] < 27 {
		signature[64] += 27
	}

	return hexutil.Encode(signature), nil
}

// Address returns the Ethereum address derived from the private key
func (s *L1Signer) Address() string {
	return crypto.PubkeyToAddress(s.privateKey.PublicKey).Hex()
}
