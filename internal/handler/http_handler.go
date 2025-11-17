package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/pocketsmith-proxy/internal/domain"
	"github.com/pocketsmith-proxy/internal/service"
)

// HTTPHandler handles HTTP requests for the transaction API
type HTTPHandler struct {
	service       service.TransactionService
	clientAuthKey string
}

// NewHTTPHandler creates a new HTTP handler
func NewHTTPHandler(svc service.TransactionService, clientAuthKey string) *HTTPHandler {
	return &HTTPHandler{
		service:       svc,
		clientAuthKey: clientAuthKey,
	}
}

// Handle processes incoming HTTP requests
func (h *HTTPHandler) Handle(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	path := r.URL.Path

	// Route based on path and method
	switch {
	case path == "/api/v1/transactions/append" && method == http.MethodPost:
		h.handleAddTransaction(w, r)
	case path == "/api/v1/categories" && method == http.MethodGet:
		h.handleGetCategories(w, r)
	case path == "/api/v1/accounts" && method == http.MethodGet:
		h.handleGetAccounts(w, r)
	case path == "/api/v1/shortcut_entities" && method == http.MethodGet:
		h.handleGetShortcutEntities(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, "Not found")
		h.logRequest(method, path, http.StatusNotFound)
	}
}

// handleAddTransaction handles POST /api/v1/transactions/append
func (h *HTTPHandler) handleAddTransaction(w http.ResponseWriter, r *http.Request) {
	var statusCode int
	method := r.Method
	path := r.URL.Path

	// Validate and parse request
	tx, statusCode, errMsg := h.validateAndParseRequest(r)
	if statusCode != http.StatusOK {
		w.WriteHeader(statusCode)
		fmt.Fprintln(w, errMsg)
		h.logRequest(method, path, statusCode)
		return
	}

	// Process transaction
	if err := h.service.AddTransaction(tx); err != nil {
		// Check if it's a lookup error (return 400) or internal error (return 500)
		if service.IsLookupError(err) {
			statusCode = http.StatusBadRequest
		} else {
			statusCode = http.StatusInternalServerError
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		errorResponse := map[string]string{
			"error": err.Error(),
		}
		json.NewEncoder(w).Encode(errorResponse)
		h.logRequest(method, path, statusCode)
		return
	}

	// Success
	statusCode = http.StatusOK
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	fmt.Fprintln(w, `{"result": "ok"}`)
	h.logRequest(method, path, statusCode)
}

// handleGetCategories handles GET /api/v1/categories
func (h *HTTPHandler) handleGetCategories(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	path := r.URL.Path
	var statusCode int

	// Validate auth
	if !h.validateAuth(r) {
		statusCode = http.StatusForbidden
		w.WriteHeader(statusCode)
		fmt.Fprintln(w, "Forbidden")
		h.logRequest(method, path, statusCode)
		return
	}

	// Get categories from service
	categories, err := h.service.GetCategories()
	if err != nil {
		statusCode = http.StatusInternalServerError
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		errorResponse := map[string]string{
			"error": err.Error(),
		}
		json.NewEncoder(w).Encode(errorResponse)
		h.logRequest(method, path, statusCode)
		return
	}

	// Success response
	statusCode = http.StatusOK
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := map[string]interface{}{
		"items": categories,
	}
	json.NewEncoder(w).Encode(response)
	h.logRequest(method, path, statusCode)
}

// handleGetAccounts handles GET /api/v1/accounts
func (h *HTTPHandler) handleGetAccounts(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	path := r.URL.Path
	var statusCode int

	// Validate auth
	if !h.validateAuth(r) {
		statusCode = http.StatusForbidden
		w.WriteHeader(statusCode)
		fmt.Fprintln(w, "Forbidden")
		h.logRequest(method, path, statusCode)
		return
	}

	// Get accounts from service
	accounts, err := h.service.GetAccounts()
	if err != nil {
		statusCode = http.StatusInternalServerError
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		errorResponse := map[string]string{
			"error": err.Error(),
		}
		json.NewEncoder(w).Encode(errorResponse)
		h.logRequest(method, path, statusCode)
		return
	}

	// Success response
	statusCode = http.StatusOK
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := map[string]interface{}{
		"items": accounts,
	}
	json.NewEncoder(w).Encode(response)
	h.logRequest(method, path, statusCode)
}

// handleGetShortcutEntities handles GET /api/v1/shortcut_entities
func (h *HTTPHandler) handleGetShortcutEntities(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	path := r.URL.Path
	var statusCode int

	// Validate auth
	if !h.validateAuth(r) {
		statusCode = http.StatusForbidden
		w.WriteHeader(statusCode)
		fmt.Fprintln(w, "Forbidden")
		h.logRequest(method, path, statusCode)
		return
	}

	// Get shortcut entities from service
	entities, err := h.service.GetShortcutEntities()
	if err != nil {
		statusCode = http.StatusInternalServerError
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		errorResponse := map[string]string{
			"error": err.Error(),
		}
		json.NewEncoder(w).Encode(errorResponse)
		h.logRequest(method, path, statusCode)
		return
	}

	// Success response
	statusCode = http.StatusOK
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := map[string]interface{}{
		"data": entities,
	}
	json.NewEncoder(w).Encode(response)
	h.logRequest(method, path, statusCode)
}

// validateAndParseRequest validates the HTTP request and parses it into a Transaction
func (h *HTTPHandler) validateAndParseRequest(r *http.Request) (*domain.Transaction, int, string) {
	// Validate HTTP method is POST
	if r.Method != http.MethodPost {
		return nil, http.StatusMethodNotAllowed, "Method not allowed"
	}

	// Validate Content-Type is application/json
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		return nil, http.StatusBadRequest, "Bad request"
	}

	// Validate Authorization header
	clientToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if h.clientAuthKey != clientToken {
		log.Println("Invalid client auth")
		return nil, http.StatusForbidden, "Forbidden"
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, http.StatusBadRequest, "Error reading request body"
	}
	defer r.Body.Close()

	// Decode JSON body into RPCRequest
	var rpcReq domain.RPCRequest
	if err := json.Unmarshal(body, &rpcReq); err != nil {
		return nil, http.StatusBadRequest, "Bad request"
	}

	// Validate method field equals 'transactions.add'
	if rpcReq.Method != "transactions.add" {
		return nil, http.StatusBadRequest, "Bad request"
	}

	// Validate params contains required fields for transactions.add
	if rpcReq.Params == nil {
		return nil, http.StatusBadRequest, "Bad request"
	}

	// Convert params map to TransactionParams to validate structure
	paramsJSON, err := json.Marshal(rpcReq.Params)
	if err != nil {
		return nil, http.StatusBadRequest, "Bad request"
	}

	var txParams domain.TransactionParams
	if err := json.Unmarshal(paramsJSON, &txParams); err != nil {
		return nil, http.StatusBadRequest, "Bad request"
	}

	// Validate all required fields are present
	if txParams.Currency == "" || txParams.Category == "" || txParams.Merchant == "" || txParams.Value == "" || txParams.Date == "" {
		return nil, http.StatusBadRequest, "Bad request"
	}

	// Validate and normalize the amount field
	amount := txParams.Value
	// Replace commas with dots
	amount = strings.ReplaceAll(amount, ",", ".")
	// Check for multiple dots
	if strings.Count(amount, ".") > 1 {
		return nil, http.StatusUnprocessableEntity, "Invalid amount format: multiple decimal separators"
	}

	// Create domain transaction
	tx := &domain.Transaction{
		Currency: txParams.Currency,
		Category: txParams.Category,
		Merchant: txParams.Merchant,
		Amount:   amount,
		Date:     txParams.Date,
	}

	return tx, http.StatusOK, ""
}

// validateAuth validates the Authorization header
func (h *HTTPHandler) validateAuth(r *http.Request) bool {
	clientToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if h.clientAuthKey != clientToken {
		log.Println("Invalid client auth")
		return false
	}
	return true
}

// logRequest logs the HTTP request details
func (h *HTTPHandler) logRequest(method, path string, statusCode int) {
	log.Printf("- %s %d %s\n", method, statusCode, path)
}
