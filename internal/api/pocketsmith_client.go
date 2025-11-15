package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/pocketsmith-proxy/internal/domain"
	"github.com/pocketsmith-proxy/internal/repository"
	spinhttp "github.com/spinframework/spin-go-sdk/v2/http"
)

// PocketSmithClient defines the interface for interacting with PocketSmith API
type PocketSmithClient interface {
	// GetMe gets the authenticated user's information
	GetMe() (*domain.User, error)
	// GetTransactionAccounts gets all transaction accounts for a user
	GetTransactionAccounts(userID int) ([]domain.TransactionAccount, error)
	// GetCategories gets all categories for a user
	GetCategories(userID int) ([]domain.Category, error)
	// CreateTransaction creates a new transaction in the specified account
	CreateTransaction(accountID int, transaction *domain.PocketSmithTransaction) error
}

// HTTPPocketSmithClient implements PocketSmithClient using HTTP
type HTTPPocketSmithClient struct {
	apiKey  string
	baseURL string
	cache   repository.CacheRepository
}

// NewHTTPPocketSmithClient creates a new HTTP-based PocketSmith client
func NewHTTPPocketSmithClient(apiKey string, cache repository.CacheRepository) PocketSmithClient {
	return &HTTPPocketSmithClient{
		apiKey:  apiKey,
		baseURL: "https://api.pocketsmith.com/v2",
		cache:   cache,
	}
}

// GetMe implements PocketSmithClient.GetMe
func (c *HTTPPocketSmithClient) GetMe() (*domain.User, error) {
	// Try to get from cache first
	userID, err := c.cache.GetUserID()
	if err == nil {
		// Cache hit
		return &domain.User{ID: userID}, nil
	}

	// Cache miss - fetch from API
	log.Printf("Cache miss for user ID, fetching from PocketSmith API")

	// Create HTTP request
	url := fmt.Sprintf("%s/me", c.baseURL)
	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("accept", "application/json")
	httpReq.Header.Set("X-Developer-Key", c.apiKey)

	// Send request to PocketSmith API
	resp, err := spinhttp.Send(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request to PocketSmith: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response from PocketSmith: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		log.Printf("ERROR: Failed to fetch user from PocketSmith API (status %d): %s", resp.StatusCode, string(responseBody))
		return nil, fmt.Errorf("PocketSmith request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	// Unmarshal response
	var user domain.User
	if err := json.Unmarshal(responseBody, &user); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	// Store in cache
	if err := c.cache.SetUserID(user.ID); err != nil {
		log.Printf("Warning: Failed to cache user ID: %v", err)
	}

	return &user, nil
}

// GetTransactionAccounts implements PocketSmithClient.GetTransactionAccounts
func (c *HTTPPocketSmithClient) GetTransactionAccounts(userID int) ([]domain.TransactionAccount, error) {
	// Try to get from cache first
	accounts, err := c.cache.GetTransactionAccounts(userID)
	if err == nil {
		// Cache hit
		return accounts, nil
	}

	// Cache miss - fetch from API
	log.Printf("Cache miss for transaction accounts (user %d), fetching from PocketSmith API", userID)

	// Create HTTP request
	url := fmt.Sprintf("%s/users/%d/transaction_accounts", c.baseURL, userID)
	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("accept", "application/json")
	httpReq.Header.Set("X-Developer-Key", c.apiKey)

	// Send request to PocketSmith API
	resp, err := spinhttp.Send(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request to PocketSmith: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response from PocketSmith: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		log.Printf("ERROR: Failed to fetch transaction accounts for user %d from PocketSmith API (status %d): %s", userID, resp.StatusCode, string(responseBody))
		return nil, fmt.Errorf("PocketSmith request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	// Unmarshal response
	if err := json.Unmarshal(responseBody, &accounts); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	// Store in cache
	if err := c.cache.SetTransactionAccounts(userID, accounts); err != nil {
		log.Printf("Warning: Failed to cache transaction accounts: %v", err)
	}

	return accounts, nil
}

// GetCategories implements PocketSmithClient.GetCategories
func (c *HTTPPocketSmithClient) GetCategories(userID int) ([]domain.Category, error) {
	// Try to get from cache first
	categories, err := c.cache.GetCategories(userID)
	if err == nil {
		// Cache hit
		return categories, nil
	}

	// Cache miss - fetch from API
	log.Printf("Cache miss for categories (user %d), fetching from PocketSmith API", userID)

	// Create HTTP request
	url := fmt.Sprintf("%s/users/%d/categories", c.baseURL, userID)
	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("accept", "application/json")
	httpReq.Header.Set("X-Developer-Key", c.apiKey)

	// Send request to PocketSmith API
	resp, err := spinhttp.Send(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send request to PocketSmith: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response from PocketSmith: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		log.Printf("ERROR: Failed to fetch categories for user %d from PocketSmith API (status %d): %s", userID, resp.StatusCode, string(responseBody))
		return nil, fmt.Errorf("PocketSmith request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	// Unmarshal response
	if err := json.Unmarshal(responseBody, &categories); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	// Store in cache
	if err := c.cache.SetCategories(userID, categories); err != nil {
		log.Printf("Warning: Failed to cache categories: %v", err)
	}

	return categories, nil
}

// CreateTransaction implements PocketSmithClient.CreateTransaction
func (c *HTTPPocketSmithClient) CreateTransaction(accountID int, transaction *domain.PocketSmithTransaction) error {
	// Marshal request body
	requestBody, err := json.Marshal(transaction)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/transaction_accounts/%d/transactions", c.baseURL, accountID)
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
