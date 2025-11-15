package main

import (
	"log"
	"net/http"

	"github.com/fermyon/spin/sdk/go/v2/variables"
	"github.com/pocketsmith-proxy/internal/api"
	"github.com/pocketsmith-proxy/internal/handler"
	"github.com/pocketsmith-proxy/internal/repository"
	"github.com/pocketsmith-proxy/internal/service"
	spinhttp "github.com/spinframework/spin-go-sdk/v2/http"
)

func init() {
	spinhttp.Handle(handleRequest)
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	// Get configuration from environment variables
	clientAuthKey, err := variables.Get("client_auth_key")
	if err != nil {
		log.Printf("Failed to get client_auth_key: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	pocketsmithAPIKey, err := variables.Get("pocketsmith_api_key")
	if err != nil {
		log.Printf("Failed to get pocketsmith_api_key: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	redisAddress, err := variables.Get("redis_address")
	if err != nil {
		log.Printf("Failed to get redis_address: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Initialize layers (Cache -> API -> Service -> Handler)
	// Layer 0: Cache Repository
	cacheRepo := repository.NewRedisCacheRepository(redisAddress)

	// Layer 1: API Client
	apiClient := api.NewHTTPPocketSmithClient(pocketsmithAPIKey, cacheRepo)

	// Layer 2: Service
	transactionService := service.NewTransactionService(apiClient)

	// Layer 3: Handler (Facade)
	httpHandler := handler.NewHTTPHandler(transactionService, clientAuthKey)

	// Delegate to handler
	httpHandler.Handle(w, r)
}

func main() {}
