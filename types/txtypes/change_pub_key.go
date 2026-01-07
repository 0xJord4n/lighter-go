package txtypes

import (
	"encoding/hex"
	"fmt"

	"github.com/bytedance/sonic"
	g "github.com/elliottech/poseidon_crypto/field/goldilocks"
	p2 "github.com/elliottech/poseidon_crypto/hash/poseidon2_goldilocks"
	"github.com/ethereum/go-ethereum/common"
)

var _ TxInfo = (*L2ChangePubKeyTxInfo)(nil)

type L2ChangePubKeyTxInfo struct {
	AccountIndex int64
	ApiKeyIndex  uint8

	PubKey []byte
	L1Sig  string

	ExpiredAt  int64
	Nonce      int64
	Sig        []byte
	SignedHash string `json:"-"`
}

// l2ChangePubKeyTxInfoJSON is the JSON representation with hex-encoded byte fields
type l2ChangePubKeyTxInfoJSON struct {
	AccountIndex int64  `json:"AccountIndex"`
	ApiKeyIndex  uint8  `json:"ApiKeyIndex"`
	PubKey       string `json:"PubKey"`
	L1Sig        string `json:"L1Sig"`
	ExpiredAt    int64  `json:"ExpiredAt"`
	Nonce        int64  `json:"Nonce"`
	Sig          string `json:"Sig"`
}

// MarshalJSON implements custom JSON marshaling to encode PubKey and Sig as hex strings
func (txInfo *L2ChangePubKeyTxInfo) MarshalJSON() ([]byte, error) {
	return sonic.Marshal(&l2ChangePubKeyTxInfoJSON{
		AccountIndex: txInfo.AccountIndex,
		ApiKeyIndex:  txInfo.ApiKeyIndex,
		PubKey:       hex.EncodeToString(txInfo.PubKey),
		L1Sig:        txInfo.L1Sig,
		ExpiredAt:    txInfo.ExpiredAt,
		Nonce:        txInfo.Nonce,
		Sig:          hex.EncodeToString(txInfo.Sig),
	})
}

func (txInfo *L2ChangePubKeyTxInfo) GetTxType() uint8 {
	return TxTypeL2ChangePubKey
}

func (txInfo *L2ChangePubKeyTxInfo) GetTxInfo() (string, error) {
	return getTxInfo(txInfo)
}

func (txInfo *L2ChangePubKeyTxInfo) GetTxHash() string {
	return txInfo.SignedHash
}

func (txInfo *L2ChangePubKeyTxInfo) Validate() error {
	// AccountIndex
	if txInfo.AccountIndex < MinAccountIndex {
		return ErrFromAccountIndexTooLow
	}
	if txInfo.AccountIndex > MaxAccountIndex {
		return ErrFromAccountIndexTooHigh
	}

	// ApiKeyIndex
	if txInfo.ApiKeyIndex < MinApiKeyIndex {
		return ErrApiKeyIndexTooLow
	}

	if txInfo.ApiKeyIndex > MaxApiKeyIndex {
		return ErrApiKeyIndexTooHigh
	}

	if txInfo.Nonce < MinNonce {
		return ErrNonceTooLow
	}

	if txInfo.ExpiredAt < 0 || txInfo.ExpiredAt > MaxTimestamp {
		return ErrExpiredAtInvalid
	}

	if !IsValidPubKeyLength(txInfo.PubKey) {
		return ErrPubKeyInvalid
	}

	return nil
}

func (txInfo *L2ChangePubKeyTxInfo) GetL1SignatureBody() string {
	signatureBody := fmt.Sprintf(
		TemplateChangePubKey,
		common.Bytes2Hex(txInfo.PubKey),
		getHex10FromUint64(uint64(txInfo.Nonce)),
		getHex10FromUint64(uint64(txInfo.AccountIndex)),
		getHex10FromUint64(uint64(txInfo.ApiKeyIndex)),
	)
	return signatureBody
}

func (txInfo *L2ChangePubKeyTxInfo) GetL1AddressBySignature() common.Address {
	return calculateL1AddressBySignature(txInfo.GetL1SignatureBody(), txInfo.L1Sig)
}

// SetL1Sig sets the L1 (Ethereum) signature on the transaction
func (txInfo *L2ChangePubKeyTxInfo) SetL1Sig(sig string) {
	txInfo.L1Sig = sig
}

func (txInfo *L2ChangePubKeyTxInfo) Hash(lighterChainId uint32, extra ...g.Element) (msgHash []byte, err error) {
	elems := make([]g.Element, 0, 11)

	elems = append(elems, g.FromUint32(lighterChainId))
	elems = append(elems, g.FromUint32(TxTypeL2ChangePubKey))
	elems = append(elems, g.FromInt64(txInfo.Nonce))
	elems = append(elems, g.FromInt64(txInfo.ExpiredAt))
	elems = append(elems, g.FromInt64(txInfo.AccountIndex))
	elems = append(elems, g.FromUint32(uint32(txInfo.ApiKeyIndex)))

	pubKeyFieldElems, err := g.ArrayFromCanonicalLittleEndianBytes(txInfo.PubKey)
	if err != nil {
		return nil, fmt.Errorf("failed to convert bytes to field element. bytes: %v, error: %w", txInfo.PubKey, err)
	}
	elems = append(elems, pubKeyFieldElems...)

	return p2.HashToQuinticExtension(elems).ToLittleEndianBytes(), nil
}
