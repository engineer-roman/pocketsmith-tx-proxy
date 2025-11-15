package repository

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/fermyon/spin/sdk/go/v2/redis"
	"github.com/pocketsmith-proxy/internal/domain"
)

const (
	// Cache TTL in seconds (86400 = 24 hours)
	cacheTTL = 86400
)

// CacheRepository defines the interface for cache operations
type CacheRepository interface {
	// User ID operations
	GetUserID() (int, error)
	SetUserID(userID int) error

	// Transaction accounts operations
	GetTransactionAccounts(userID int) ([]domain.TransactionAccount, error)
	SetTransactionAccounts(userID int, accounts []domain.TransactionAccount) error

	// Categories operations
	GetCategories(userID int) ([]domain.Category, error)
	SetCategories(userID int, categories []domain.Category) error
}

// RedisCacheRepository implements CacheRepository using Redis
type RedisCacheRepository struct {
	client *redis.Client
}

// NewRedisCacheRepository creates a new Redis-based cache repository
func NewRedisCacheRepository(redisAddress string) CacheRepository {
	return &RedisCacheRepository{
		client: redis.NewClient(redisAddress),
	}
}

// GetUserID retrieves the cached user ID
func (r *RedisCacheRepository) GetUserID() (int, error) {
	data, err := r.client.Get("user:id")
	if err != nil {
		return 0, fmt.Errorf("redis get user:id: %w", err)
	}

	userID, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, fmt.Errorf("parse user ID: %w", err)
	}

	log.Printf("Cache hit: user:id = %d", userID)
	return userID, nil
}

// SetUserID stores the user ID in cache with TTL
func (r *RedisCacheRepository) SetUserID(userID int) error {
	// Set the user ID
	err := r.client.Set("user:id", []byte(strconv.Itoa(userID)))
	if err != nil {
		return fmt.Errorf("redis set user:id: %w", err)
	}

	// Set expiration
	_, err = r.client.Execute("EXPIRE", "user:id", cacheTTL)
	if err != nil {
		return fmt.Errorf("redis expire user:id: %w", err)
	}

	log.Printf("Cache set: user:id = %d (TTL: %d seconds)", userID, cacheTTL)
	return nil
}

// GetTransactionAccounts retrieves cached transaction accounts for a user
func (r *RedisCacheRepository) GetTransactionAccounts(userID int) ([]domain.TransactionAccount, error) {
	key := fmt.Sprintf("user:%d:accounts", userID)

	// Get all fields from the hash
	results, err := r.client.Execute("HGETALL", key)
	if err != nil {
		return nil, fmt.Errorf("redis hgetall %s: %w", key, err)
	}

	// HGETALL returns alternating field/value pairs
	if len(results) == 0 {
		return nil, fmt.Errorf("cache miss: %s", key)
	}

	// Parse the hash into a map
	accountsMap := make(map[string]string)
	for i := 0; i < len(results); i += 2 {
		if i+1 >= len(results) {
			break
		}
		field := string(results[i].Val.([]byte))
		value := string(results[i+1].Val.([]byte))
		accountsMap[field] = value
	}

	// Extract the JSON data (stored under "data" field)
	jsonData, ok := accountsMap["data"]
	if !ok {
		return nil, fmt.Errorf("cache miss: %s (no data field)", key)
	}

	var accounts []domain.TransactionAccount
	if err := json.Unmarshal([]byte(jsonData), &accounts); err != nil {
		return nil, fmt.Errorf("unmarshal accounts: %w", err)
	}

	log.Printf("Cache hit: %s (%d accounts)", key, len(accounts))
	return accounts, nil
}

// SetTransactionAccounts stores transaction accounts in cache with TTL
func (r *RedisCacheRepository) SetTransactionAccounts(userID int, accounts []domain.TransactionAccount) error {
	key := fmt.Sprintf("user:%d:accounts", userID)

	// Marshal accounts to JSON
	data, err := json.Marshal(accounts)
	if err != nil {
		return fmt.Errorf("marshal accounts: %w", err)
	}

	// Store in hash
	_, err = r.client.Execute("HSET", key, "data", string(data))
	if err != nil {
		return fmt.Errorf("redis hset %s: %w", key, err)
	}

	// Set expiration on the key
	_, err = r.client.Execute("EXPIRE", key, cacheTTL)
	if err != nil {
		return fmt.Errorf("redis expire %s: %w", key, err)
	}

	log.Printf("Cache set: %s (%d accounts, TTL: %d seconds)", key, len(accounts), cacheTTL)
	return nil
}

// GetCategories retrieves cached categories for a user
func (r *RedisCacheRepository) GetCategories(userID int) ([]domain.Category, error) {
	key := fmt.Sprintf("user:%d:categories", userID)

	// Get all fields from the hash
	results, err := r.client.Execute("HGETALL", key)
	if err != nil {
		return nil, fmt.Errorf("redis hgetall %s: %w", key, err)
	}

	// HGETALL returns alternating field/value pairs
	if len(results) == 0 {
		return nil, fmt.Errorf("cache miss: %s", key)
	}

	// Parse the hash into a map
	categoriesMap := make(map[string]string)
	for i := 0; i < len(results); i += 2 {
		if i+1 >= len(results) {
			break
		}
		field := string(results[i].Val.([]byte))
		value := string(results[i+1].Val.([]byte))
		categoriesMap[field] = value
	}

	// Extract the JSON data (stored under "data" field)
	jsonData, ok := categoriesMap["data"]
	if !ok {
		return nil, fmt.Errorf("cache miss: %s (no data field)", key)
	}

	var categories []domain.Category
	if err := json.Unmarshal([]byte(jsonData), &categories); err != nil {
		return nil, fmt.Errorf("unmarshal categories: %w", err)
	}

	log.Printf("Cache hit: %s (%d categories)", key, len(categories))
	return categories, nil
}

// SetCategories stores categories in cache with TTL
func (r *RedisCacheRepository) SetCategories(userID int, categories []domain.Category) error {
	key := fmt.Sprintf("user:%d:categories", userID)

	// Marshal categories to JSON
	data, err := json.Marshal(categories)
	if err != nil {
		return fmt.Errorf("marshal categories: %w", err)
	}

	// Store in hash
	_, err = r.client.Execute("HSET", key, "data", string(data))
	if err != nil {
		return fmt.Errorf("redis hset %s: %w", key, err)
	}

	// Set expiration on the key
	_, err = r.client.Execute("EXPIRE", key, cacheTTL)
	if err != nil {
		return fmt.Errorf("redis expire %s: %w", key, err)
	}

	log.Printf("Cache set: %s (%d categories, TTL: %d seconds)", key, len(categories), cacheTTL)
	return nil
}
