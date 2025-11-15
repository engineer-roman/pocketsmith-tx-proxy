package service

import (
	"fmt"
	"strings"
	"sync"

	"github.com/pocketsmith-proxy/internal/api"
	"github.com/pocketsmith-proxy/internal/domain"
)

// TransactionService defines the interface for transaction business logic
type TransactionService interface {
	// AddTransaction adds a transaction to the appropriate account
	AddTransaction(tx *domain.Transaction) error
}

// TransactionServiceImpl implements TransactionService
type TransactionServiceImpl struct {
	client api.PocketSmithClient
}

// NewTransactionService creates a new transaction service
func NewTransactionService(client api.PocketSmithClient) TransactionService {
	return &TransactionServiceImpl{
		client: client,
	}
}

// lookupError represents an error that should return 400 Bad Request
type lookupError struct {
	message string
}

func (e *lookupError) Error() string {
	return e.message
}

// IsLookupError checks if an error is a lookup error (should return 400)
func IsLookupError(err error) bool {
	_, ok := err.(*lookupError)
	return ok
}

// AddTransaction implements TransactionService.AddTransaction
func (s *TransactionServiceImpl) AddTransaction(tx *domain.Transaction) error {
	// Normalize inputs
	currencyLower := strings.ToLower(tx.Currency)
	categoryLower := strings.ToLower(tx.Category)

	// Get user ID
	user, err := s.client.GetMe()
	if err != nil {
		return fmt.Errorf("failed to get user info: %w", err)
	}

	// Fetch transaction accounts and categories concurrently
	var (
		accounts      []domain.TransactionAccount
		categories    []domain.Category
		wg            sync.WaitGroup
		errAccounts   error
		errCategories error
	)

	wg.Add(2)

	// Fetch transaction accounts
	go func() {
		defer wg.Done()
		accounts, errAccounts = s.client.GetTransactionAccounts(user.ID)
	}()

	// Fetch categories
	go func() {
		defer wg.Done()
		categories, errCategories = s.client.GetCategories(user.ID)
	}()

	wg.Wait()

	// Check for errors in concurrent fetches
	if errAccounts != nil {
		return fmt.Errorf("failed to get transaction accounts: %w", errAccounts)
	}
	if errCategories != nil {
		return fmt.Errorf("failed to get categories: %w", errCategories)
	}

	// Find transaction account by currency code
	var accountID *int
	for _, account := range accounts {
		if strings.ToLower(account.CurrencyCode) == currencyLower {
			accountID = &account.ID
			break
		}
	}
	if accountID == nil {
		return &lookupError{message: fmt.Sprintf("no transaction account found for currency: %s", tx.Currency)}
	}

	// Find category by title
	var categoryID *int
	categoryID = s.findCategoryByTitle(categories, categoryLower)
	if categoryID == nil {
		return &lookupError{message: fmt.Sprintf("no category found with title: %s", tx.Category)}
	}

	// Transform domain transaction to PocketSmith format
	psTx := &domain.PocketSmithTransaction{
		Payee:      tx.Merchant,
		Amount:     tx.Amount,
		Date:       tx.Date,
		IsTransfer: false,
		CategoryID: categoryID,
	}

	// Create transaction via API client
	return s.client.CreateTransaction(*accountID, psTx)
}

// findCategoryByTitle recursively searches for a category by title (case-insensitive)
// Categories can be nested, so we need to search the entire tree
func (s *TransactionServiceImpl) findCategoryByTitle(categories []domain.Category, titleLower string) *int {
	for _, category := range categories {
		if strings.ToLower(category.Title) == titleLower {
			return &category.ID
		}
	}
	return nil
}
