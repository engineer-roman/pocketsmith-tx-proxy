package service

import (
	"fmt"
	"strings"

	"github.com/pocketsmith-proxy/internal/api"
	"github.com/pocketsmith-proxy/internal/domain"
)

// TransactionService defines the interface for transaction business logic
type TransactionService interface {
	// AddTransaction adds a transaction to the appropriate account
	AddTransaction(tx *domain.Transaction) error
}

// Config holds the configuration for the transaction service
type Config struct {
	USDAccountID string
	ARSAccountID string
}

// TransactionServiceImpl implements TransactionService
type TransactionServiceImpl struct {
	client api.PocketSmithClient
	config *Config
}

// NewTransactionService creates a new transaction service
func NewTransactionService(client api.PocketSmithClient, config *Config) TransactionService {
	return &TransactionServiceImpl{
		client: client,
		config: config,
	}
}

// AddTransaction implements TransactionService.AddTransaction
func (s *TransactionServiceImpl) AddTransaction(tx *domain.Transaction) error {
	// Determine account ID based on currency
	accountID, err := s.getAccountID(tx.Currency)
	if err != nil {
		return err
	}

	// Transform domain transaction to PocketSmith format
	psTx := &domain.PocketSmithTransaction{
		Payee:      tx.Merchant,
		Amount:     tx.Amount,
		Date:       tx.Date,
		IsTransfer: false,
	}

	// Create transaction via API client
	return s.client.CreateTransaction(accountID, psTx)
}

// getAccountID returns the account ID for the given currency
func (s *TransactionServiceImpl) getAccountID(currency string) (string, error) {
	switch strings.ToLower(currency) {
	case "usd":
		if s.config.USDAccountID == "" {
			return "", fmt.Errorf("USD account ID not configured")
		}
		return s.config.USDAccountID, nil
	case "ars":
		if s.config.ARSAccountID == "" {
			return "", fmt.Errorf("ARS account ID not configured")
		}
		return s.config.ARSAccountID, nil
	default:
		return "", fmt.Errorf("unsupported currency: %s", currency)
	}
}
