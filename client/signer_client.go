// Package client provides the core trading client for the Lighter API.
//
// SignerClient is the main entry point for trading operations. It provides
// high-level methods for creating and submitting orders, managing positions,
// and interacting with the Lighter exchange.
//
// Example:
//
//	// Create a client (recommended - network handles chain ID automatically)
//	httpClient := http.NewFullClientForNetwork(client.Mainnet)
//	client, err := client.NewSignerClientForNetwork(httpClient, client.Mainnet, privateKey, 0, accountIndex, nil)
//
//	// Or create with explicit URL and chain ID
//	httpClient := http.NewFullClient("https://mainnet.zklighter.elliot.ai")
//	client, err := client.NewSignerClient(httpClient, privateKey, 304, 0, accountIndex, nil)
//
//	// Create a market buy order
//	txInfo, err := client.CreateMarketOrder(0, 100000, true, nil) // Buy 0.01 ETH
//
//	// Submit the order
//	resp, err := client.SendAndSubmit(txInfo)
//
//	// Create a limit order
//	txInfo, err := client.CreateLimitOrder(0, 100000, 350000, true, expiry, nil)
//
//	// Create stop loss / take profit orders
//	txInfo, err := client.CreateStopLossOrder(0, 100000, 340000, false, expiry, nil)
//	txInfo, err := client.CreateTakeProfitOrder(0, 100000, 360000, false, expiry, nil)
package client

import (
	"fmt"
	"time"

	"github.com/0xJord4n/lighter-go/nonce"
	"github.com/0xJord4n/lighter-go/signer"
	"github.com/0xJord4n/lighter-go/types"
	"github.com/0xJord4n/lighter-go/types/api"
	"github.com/0xJord4n/lighter-go/types/txtypes"
)

// SignerClient extends TxClient with convenience methods for common trading patterns.
// It provides higher-level APIs similar to the Python SDK's SignerClient.
type SignerClient struct {
	*TxClient
	fullHTTP     FullHTTPClient
	nonceManager nonce.Manager
}

// NewSignerClient creates a SignerClient with full HTTP capabilities.
// If nonceManager is nil, a new OptimisticNonceManager will be created.
func NewSignerClient(httpClient FullHTTPClient, privateKey string, chainId uint32, apiKeyIndex uint8, accountIndex int64, nonceManager nonce.Manager) (*SignerClient, error) {
	txClient, err := createTxClient(httpClient, privateKey, chainId, apiKeyIndex, accountIndex)
	if err != nil {
		return nil, err
	}

	if nonceManager == nil {
		nonceManager = nonce.NewOptimisticManager(httpClient)
	}

	return &SignerClient{
		TxClient:     txClient,
		fullHTTP:     httpClient,
		nonceManager: nonceManager,
	}, nil
}

// NewSignerClientForNetwork creates a SignerClient using the chain ID from the specified network.
// The httpClient should be created using http.NewFullClientForNetwork(network) or http.NewFullClient(network.APIURL()).
//
// Example:
//
//	httpClient := http.NewFullClientForNetwork(client.Mainnet)
//	client, err := client.NewSignerClientForNetwork(httpClient, client.Mainnet, privateKey, 0, accountIndex, nil)
func NewSignerClientForNetwork(httpClient FullHTTPClient, network Network, privateKey string, apiKeyIndex uint8, accountIndex int64, nonceManager nonce.Manager) (*SignerClient, error) {
	return NewSignerClient(httpClient, privateKey, network.ChainID(), apiKeyIndex, accountIndex, nonceManager)
}

// createTxClient is a helper that creates a TxClient without registering it globally
func createTxClient(httpClient MinimalHTTPClient, privateKey string, chainId uint32, apiKeyIndex uint8, accountIndex int64) (*TxClient, error) {
	// Use the existing CreateClient function but we need to get the client back
	return CreateClient(httpClient, privateKey, chainId, apiKeyIndex, accountIndex)
}

// FullHTTP returns the full HTTP client for direct API access
func (c *SignerClient) FullHTTP() FullHTTPClient {
	return c.fullHTTP
}

// NonceManager returns the nonce manager
func (c *SignerClient) NonceManager() nonce.Manager {
	return c.nonceManager
}

// CreateMarketOrder creates a market order with minimal parameters
func (c *SignerClient) CreateMarketOrder(marketIndex int16, size int64, isBuy bool, opts *types.TransactOpts) (*txtypes.L2CreateOrderTxInfo, error) {
	isAsk := uint8(0)
	if !isBuy {
		isAsk = 1
	}

	// For market orders, use max/min price depending on side
	var price uint32
	if isBuy {
		price = txtypes.MaxOrderPrice
	} else {
		price = txtypes.MinOrderPrice
	}

	req := &types.CreateOrderTxReq{
		MarketIndex:      marketIndex,
		ClientOrderIndex: 0, // Auto-generate
		BaseAmount:       size,
		Price:            price,
		IsAsk:            isAsk,
		Type:             txtypes.MarketOrder,
		TimeInForce:      txtypes.ImmediateOrCancel,
		ReduceOnly:       0,
		TriggerPrice:     0,
		OrderExpiry:      0,
	}

	return c.GetCreateOrderTransaction(req, opts)
}

// CreateMarketOrderWithSlippage creates a market order with slippage protection
func (c *SignerClient) CreateMarketOrderWithSlippage(marketIndex int16, size int64, isBuy bool, slippageBps int, opts *types.TransactOpts) (*txtypes.L2CreateOrderTxInfo, error) {
	// Fetch current market price
	orderBooks, err := c.fullHTTP.Order().GetOrderBooks(&marketIndex, api.MarketFilterAll)
	if err != nil {
		return nil, fmt.Errorf("failed to get order book: %w", err)
	}

	if len(orderBooks.OrderBooks) == 0 {
		return nil, fmt.Errorf("no order book data for market %d", marketIndex)
	}

	ob := orderBooks.OrderBooks[0]
	var referencePrice string
	if isBuy && len(ob.Asks) > 0 {
		referencePrice = ob.Asks[0].Price
	} else if !isBuy && len(ob.Bids) > 0 {
		referencePrice = ob.Bids[0].Price
	} else {
		return nil, fmt.Errorf("no liquidity in order book")
	}

	price, err := calculateSlippagePrice(referencePrice, slippageBps, isBuy)
	if err != nil {
		return nil, err
	}

	isAsk := uint8(0)
	if !isBuy {
		isAsk = 1
	}

	req := &types.CreateOrderTxReq{
		MarketIndex:      marketIndex,
		ClientOrderIndex: 0,
		BaseAmount:       size,
		Price:            price,
		IsAsk:            isAsk,
		Type:             txtypes.LimitOrder, // Use limit with IOC for slippage protection
		TimeInForce:      txtypes.ImmediateOrCancel,
		ReduceOnly:       0,
		TriggerPrice:     0,
		OrderExpiry:      0,
	}

	return c.GetCreateOrderTransaction(req, opts)
}

// CreateLimitOrder creates a limit order
func (c *SignerClient) CreateLimitOrder(marketIndex int16, size int64, price uint32, isBuy bool, expiry int64, opts *types.TransactOpts) (*txtypes.L2CreateOrderTxInfo, error) {
	isAsk := uint8(0)
	if !isBuy {
		isAsk = 1
	}

	req := &types.CreateOrderTxReq{
		MarketIndex:      marketIndex,
		ClientOrderIndex: 0,
		BaseAmount:       size,
		Price:            price,
		IsAsk:            isAsk,
		Type:             txtypes.LimitOrder,
		TimeInForce:      txtypes.GoodTillTime,
		ReduceOnly:       0,
		TriggerPrice:     0,
		OrderExpiry:      expiry,
	}

	return c.GetCreateOrderTransaction(req, opts)
}

// CreateTakeProfitOrder creates a take-profit market order
func (c *SignerClient) CreateTakeProfitOrder(marketIndex int16, size int64, triggerPrice uint32, isBuy bool, expiry int64, opts *types.TransactOpts) (*txtypes.L2CreateOrderTxInfo, error) {
	isAsk := uint8(0)
	if !isBuy {
		isAsk = 1
	}

	// For TP market orders, use limit price that will fill immediately once triggered
	var price uint32
	if isBuy {
		price = txtypes.MaxOrderPrice
	} else {
		price = txtypes.MinOrderPrice
	}

	req := &types.CreateOrderTxReq{
		MarketIndex:      marketIndex,
		ClientOrderIndex: 0,
		BaseAmount:       size,
		Price:            price,
		IsAsk:            isAsk,
		Type:             txtypes.TakeProfitOrder,
		TimeInForce:      txtypes.ImmediateOrCancel,
		ReduceOnly:       1,
		TriggerPrice:     triggerPrice,
		OrderExpiry:      expiry,
	}

	return c.GetCreateOrderTransaction(req, opts)
}

// CreateTakeProfitLimitOrder creates a take-profit limit order
func (c *SignerClient) CreateTakeProfitLimitOrder(marketIndex int16, size int64, price uint32, triggerPrice uint32, isBuy bool, expiry int64, opts *types.TransactOpts) (*txtypes.L2CreateOrderTxInfo, error) {
	isAsk := uint8(0)
	if !isBuy {
		isAsk = 1
	}

	req := &types.CreateOrderTxReq{
		MarketIndex:      marketIndex,
		ClientOrderIndex: 0,
		BaseAmount:       size,
		Price:            price,
		IsAsk:            isAsk,
		Type:             txtypes.TakeProfitLimitOrder,
		TimeInForce:      txtypes.GoodTillTime,
		ReduceOnly:       1,
		TriggerPrice:     triggerPrice,
		OrderExpiry:      expiry,
	}

	return c.GetCreateOrderTransaction(req, opts)
}

// CreateStopLossOrder creates a stop-loss market order
func (c *SignerClient) CreateStopLossOrder(marketIndex int16, size int64, triggerPrice uint32, isBuy bool, expiry int64, opts *types.TransactOpts) (*txtypes.L2CreateOrderTxInfo, error) {
	isAsk := uint8(0)
	if !isBuy {
		isAsk = 1
	}

	var price uint32
	if isBuy {
		price = txtypes.MaxOrderPrice
	} else {
		price = txtypes.MinOrderPrice
	}

	req := &types.CreateOrderTxReq{
		MarketIndex:      marketIndex,
		ClientOrderIndex: 0,
		BaseAmount:       size,
		Price:            price,
		IsAsk:            isAsk,
		Type:             txtypes.StopLossOrder,
		TimeInForce:      txtypes.ImmediateOrCancel,
		ReduceOnly:       1,
		TriggerPrice:     triggerPrice,
		OrderExpiry:      expiry,
	}

	return c.GetCreateOrderTransaction(req, opts)
}

// CreateStopLossLimitOrder creates a stop-loss limit order
func (c *SignerClient) CreateStopLossLimitOrder(marketIndex int16, size int64, price uint32, triggerPrice uint32, isBuy bool, expiry int64, opts *types.TransactOpts) (*txtypes.L2CreateOrderTxInfo, error) {
	isAsk := uint8(0)
	if !isBuy {
		isAsk = 1
	}

	req := &types.CreateOrderTxReq{
		MarketIndex:      marketIndex,
		ClientOrderIndex: 0,
		BaseAmount:       size,
		Price:            price,
		IsAsk:            isAsk,
		Type:             txtypes.StopLossLimitOrder,
		TimeInForce:      txtypes.GoodTillTime,
		ReduceOnly:       1,
		TriggerPrice:     triggerPrice,
		OrderExpiry:      expiry,
	}

	return c.GetCreateOrderTransaction(req, opts)
}

// SendAndSubmit signs a transaction and submits it to the API
func (c *SignerClient) SendAndSubmit(txInfo txtypes.TxInfo) (*api.RespSendTx, error) {
	txInfoJSON, err := txInfo.GetTxInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize tx info: %w", err)
	}

	// Use SendTxWithIndices to pass account_index and api_key_index (matching TS SDK)
	accountIndex := c.GetAccountIndex()
	apiKeyIndex := c.GetApiKeyIndex()
	resp, err := c.fullHTTP.Transaction().SendTxWithIndices(
		txInfo.GetTxType(),
		txInfoJSON,
		nil,
		&accountIndex,
		&apiKeyIndex,
		"",
	)
	if err != nil {
		// Acknowledge failure for nonce recovery
		if optManager, ok := c.nonceManager.(*nonce.OptimisticManager); ok {
			// Extract nonce from tx info if possible
			optManager.AcknowledgeFailure(c.GetAccountIndex(), c.GetApiKeyIndex(), -1)
		}
		return nil, err
	}

	return resp, nil
}

// SendTxBatch submits multiple transactions
func (c *SignerClient) SendTxBatch(txInfos []txtypes.TxInfo) (*api.RespSendTxBatch, error) {
	txTypes := make([]uint8, len(txInfos))
	txInfoJSONs := make([]string, len(txInfos))

	for i, tx := range txInfos {
		txTypes[i] = tx.GetTxType()
		jsonStr, err := tx.GetTxInfo()
		if err != nil {
			return nil, fmt.Errorf("failed to serialize tx %d: %w", i, err)
		}
		txInfoJSONs[i] = jsonStr
	}

	return c.fullHTTP.Transaction().SendTxBatch(txTypes, txInfoJSONs)
}

// Transfer creates, signs (L1 + L2), and submits a transfer transaction.
// This is a convenience method that handles L1 signing internally.
// ethPrivateKey is the Ethereum private key (hex-encoded) for L1 signature.
func (c *SignerClient) Transfer(ethPrivateKey string, req *types.TransferTxReq, opts *types.TransactOpts) (*api.RespSendTx, error) {
	// Create L1 signer
	l1Signer, err := signer.NewL1Signer(ethPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create L1 signer: %w", err)
	}

	// Create the transfer transaction (L2 signed)
	txInfo, err := c.GetTransferTransaction(req, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create transfer transaction: %w", err)
	}

	// Sign with Ethereum key (L1 signature)
	l1Sig, err := l1Signer.Sign(txInfo.GetL1SignatureBody(c.GetChainId()))
	if err != nil {
		return nil, fmt.Errorf("failed to sign L1 message: %w", err)
	}
	txInfo.SetL1Sig(l1Sig)

	// Submit
	return c.SendAndSubmit(txInfo)
}

// ChangePubKey creates, signs (L1 + L2), and submits a change pub key transaction.
// This is a convenience method that handles L1 signing internally.
// ethPrivateKey is the Ethereum private key (hex-encoded) for L1 signature.
func (c *SignerClient) ChangePubKey(ethPrivateKey string, req *types.ChangePubKeyReq, opts *types.TransactOpts) (*api.RespSendTx, error) {
	// Create L1 signer
	l1Signer, err := signer.NewL1Signer(ethPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create L1 signer: %w", err)
	}

	// Create the change pub key transaction (L2 signed)
	txInfo, err := c.GetChangePubKeyTransaction(req, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create change pub key transaction: %w", err)
	}

	// Sign with Ethereum key (L1 signature)
	l1Sig, err := l1Signer.Sign(txInfo.GetL1SignatureBody())
	if err != nil {
		return nil, fmt.Errorf("failed to sign L1 message: %w", err)
	}
	txInfo.SetL1Sig(l1Sig)

	// Submit
	txInfoJSON, err := txInfo.GetTxInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize tx info: %w", err)
	}
	return c.fullHTTP.Transaction().SendTx(txInfo.GetTxType(), txInfoJSON, nil)
}

// GetOpenOrders retrieves open orders for the account
func (c *SignerClient) GetOpenOrders(marketID *int16) (*api.Orders, error) {
	authToken, err := c.getAuthToken()
	if err != nil {
		return nil, err
	}
	return c.fullHTTP.Order().GetActiveOrders(c.GetAccountIndex(), marketID, authToken)
}

// CancelAllOrders cancels all open orders
func (c *SignerClient) CancelAllOrders(opts *types.TransactOpts) (*txtypes.L2CancelAllOrdersTxInfo, error) {
	req := &types.CancelAllOrdersTxReq{
		TimeInForce: txtypes.ImmediateCancelAll,
		Time:        0,
	}
	return c.GetCancelAllOrdersTransaction(req, opts)
}

// GetPositions retrieves current positions
func (c *SignerClient) GetPositions() (*api.DetailedAccounts, error) {
	return c.fullHTTP.Account().GetAccount(api.QueryByIndex, fmt.Sprintf("%d", c.GetAccountIndex()))
}

// Helper to get auth token
func (c *SignerClient) getAuthToken() (string, error) {
	deadline := time.Now().Add(8 * time.Hour)
	authInfo, err := c.GetAuthToken(deadline)
	if err != nil {
		return "", err
	}

	// The auth token is returned as a string directly
	return authInfo, nil
}

// calculateSlippagePrice calculates price with slippage
func calculateSlippagePrice(priceStr string, slippageBps int, isBuy bool) (uint32, error) {
	// Parse price as integer (prices are typically scaled integers)
	var price int64
	if _, err := fmt.Sscanf(priceStr, "%d", &price); err != nil {
		return 0, fmt.Errorf("failed to parse price: %w", err)
	}

	// Calculate slippage adjustment
	adjustment := (price * int64(slippageBps)) / 10000

	if isBuy {
		price += adjustment
		if price > int64(txtypes.MaxOrderPrice) {
			price = int64(txtypes.MaxOrderPrice)
		}
	} else {
		price -= adjustment
		if price < int64(txtypes.MinOrderPrice) {
			price = int64(txtypes.MinOrderPrice)
		}
	}

	return uint32(price), nil
}
