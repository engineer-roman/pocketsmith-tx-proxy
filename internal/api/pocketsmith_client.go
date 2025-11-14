package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/pocketsmith-proxy/internal/domain"
	spinhttp "github.com/spinframework/spin-go-sdk/v2/http"
)

// PocketSmithClient defines the interface for interacting with PocketSmith API
type PocketSmithClient interface {
	// CreateTransaction creates a new transaction in the specified account
	CreateTransaction(accountID string, transaction *domain.PocketSmithTransaction) error
}

// HTTPPocketSmithClient implements PocketSmithClient using HTTP
type HTTPPocketSmithClient struct {
	apiKey  string
	baseURL string
}

// NewHTTPPocketSmithClient creates a new HTTP-based PocketSmith client
func NewHTTPPocketSmithClient(apiKey string) PocketSmithClient {
	return &HTTPPocketSmithClient{
		apiKey:  apiKey,
		baseURL: "https://api.pocketsmith.com/v2",
	}
}

// CreateTransaction implements PocketSmithClient.CreateTransaction
func (c *HTTPPocketSmithClient) CreateTransaction(accountID string, transaction *domain.PocketSmithTransaction) error {
	// Marshal request body
	requestBody, err := json.Marshal(transaction)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/transaction_accounts/%s/transactions", c.baseURL, accountID)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("accept", "application/json")
	httpReq.Header.Set("content-type", "application/json")
	httpReq.Header.Set("X-Developer-Key", c.apiKey)

	// Send request to PocketSmith API
	resp, err := spinhttp.Send(httpReq)
	if err != nil {
		return fmt.Errorf("send request to PocketSmith: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response from PocketSmith: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("PocketSmith request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	return nil
}
