package service

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/pocketsmith-proxy/internal/api"
	"github.com/pocketsmith-proxy/internal/domain"
)

// TransactionService defines the interface for transaction business logic
type TransactionService interface {
	// AddTransaction adds a transaction to the appropriate account
	AddTransaction(tx *domain.Transaction) error
	// GetCategories returns all category names sorted ascending
	GetCategories() ([]string, error)
	// GetAccounts returns all accounts with name and currency
	GetAccounts() ([]domain.AccountInfo, error)
	// GetShortcutEntities returns both accounts and categories for quick access
	GetShortcutEntities() (*domain.ShortcutEntities, error)
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
	accountLower := strings.ToLower(tx.Account)
	categoryLower := strings.ToLower(tx.Category)

	// Get user ID
	user, err := s.client.GetMe()
	if err != nil {
		return fmt.Errorf("failed to get user info: %w", err)
	}

	// Fetch transaction accounts
	accounts, err := s.client.GetTransactionAccounts(user.ID)
	if err != nil {
		return fmt.Errorf("failed to get transaction accounts: %w", err)
	}

	// Fetch categories
	categories, err := s.client.GetCategories(user.ID)
	if err != nil {
		return fmt.Errorf("failed to get categories: %w", err)
	}

	// Find transaction account by name
	var accountID *int
	for _, account := range accounts {
		if strings.ToLower(account.Name) == accountLower {
			accountID = &account.ID
			break
		}
	}
	if accountID == nil {
		log.Printf("ERROR: No transaction account found in PocketSmith API with name: '%s' (searched among %d accounts)", tx.Account, len(accounts))
		return &lookupError{message: fmt.Sprintf("no transaction account found with name: %s", tx.Account)}
	}

	// Find category by title
	var categoryID *int
	categoryID = s.findCategoryByTitle(categories, categoryLower)
	if categoryID == nil {
		log.Printf("ERROR: No category found in PocketSmith API with title: '%s' (searched among %d categories)", tx.Category, len(categories))
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

// GetCategories implements TransactionService.GetCategories
func (s *TransactionServiceImpl) GetCategories() ([]string, error) {
	// Get user ID
	user, err := s.client.GetMe()
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Fetch categories from cache or API
	categories, err := s.client.GetCategories(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	// Extract category names
	categoryNames := make([]string, 0, len(categories))
	for _, category := range categories {
		categoryNames = append(categoryNames, category.Title)
	}

	// Sort ascending
	sort.Strings(categoryNames)

	return categoryNames, nil
}

// GetAccounts implements TransactionService.GetAccounts
func (s *TransactionServiceImpl) GetAccounts() ([]domain.AccountInfo, error) {
	// Get user ID
	user, err := s.client.GetMe()
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Fetch accounts from cache or API
	accounts, err := s.client.GetTransactionAccounts(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction accounts: %w", err)
	}

	// Transform to AccountInfo
	accountInfos := make([]domain.AccountInfo, 0, len(accounts))
	for _, account := range accounts {
		accountInfos = append(accountInfos, domain.AccountInfo{
			Name:     account.Name,
			Currency: account.CurrencyCode,
		})
	}

	return accountInfos, nil
}

// GetShortcutEntities implements TransactionService.GetShortcutEntities
func (s *TransactionServiceImpl) GetShortcutEntities() (*domain.ShortcutEntities, error) {
	// Get user ID
	user, err := s.client.GetMe()
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Fetch accounts
	accounts, err := s.client.GetTransactionAccounts(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction accounts: %w", err)
	}

	// Fetch categories
	categories, err := s.client.GetCategories(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	// Transform accounts
	accountInfos := make([]domain.AccountInfo, 0, len(accounts))
	for _, account := range accounts {
		accountInfos = append(accountInfos, domain.AccountInfo{
			Name:     account.Name,
			Currency: account.CurrencyCode,
		})
	}

	// Extract category names
	categoryNames := make([]string, 0, len(categories))
	for _, category := range categories {
		categoryNames = append(categoryNames, category.Title)
	}

	// Sort categories
	sort.Strings(categoryNames)

	return &domain.ShortcutEntities{
		Accounts:   accountInfos,
		Categories: categoryNames,
	}, nil
}
