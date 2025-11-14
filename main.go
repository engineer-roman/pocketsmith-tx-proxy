package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/fermyon/spin/sdk/go/v2/variables"
	spinhttp "github.com/spinframework/spin-go-sdk/v2/http"
)

// Request structures
type RPCRequest struct {
	Method string         `json:"method"`
	Params map[string]any `json:"params"`
}

type TransactionParams struct {
	Currency string `json:"currency"`
	Merchant string `json:"merchant"`
	Value    string `json:"value"`
	Date     string `json:"date"`
}

type PocketSmithTransactionRequest struct {
	Payee      string `json:"payee"`
	Amount     string `json:"amount"`
	Date       string `json:"date"`
	IsTransfer bool   `json:"is_transfer"`
}

func init() {
	spinhttp.Handle(handleRequest)
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	var statusCode int
	method := r.Method
	path := r.URL.Path

	// Validate method and headers
	rpcReq, statusCode, errMsg := validateRequest(r)
	if statusCode != http.StatusOK {
		w.WriteHeader(statusCode)
		fmt.Fprintln(w, errMsg)
		logRequest(method, path, statusCode)
		return
	}

	// Handle the request
	res, err := handle(rpcReq)
	if err != nil {
		statusCode = http.StatusInternalServerError
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		errorResponse := map[string]string{
			"error": err.Error(),
		}
		json.NewEncoder(w).Encode(errorResponse)
		logRequest(method, path, statusCode)
		return
	}

	// Success
	statusCode = http.StatusOK
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	fmt.Fprintln(w, res)
	logRequest(method, path, statusCode)
}

func validateRequest(r *http.Request) (*RPCRequest, int, string) {
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
	cfgToken, err := variables.Get("client_auth_key")
	if err != nil {
		return nil, http.StatusForbidden, "Forbidden"
	}
	clientToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if cfgToken != clientToken {
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
	var rpcReq RPCRequest
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

	var txParams TransactionParams
	if err := json.Unmarshal(paramsJSON, &txParams); err != nil {
		return nil, http.StatusBadRequest, "Bad request"
	}

	// Validate all required fields are present
	if txParams.Currency == "" || txParams.Merchant == "" || txParams.Value == "" || txParams.Date == "" {
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
	// Update the params with normalized amount
	rpcReq.Params["value"] = amount

	return &rpcReq, http.StatusOK, ""
}

func handle(req *RPCRequest) (string, error) {
	// Convert params to TransactionParams
	paramsJSON, err := json.Marshal(req.Params)
	if err != nil {
		return "", fmt.Errorf("marshal params: %w", err)
	}

	var txParams TransactionParams
	if err := json.Unmarshal(paramsJSON, &txParams); err != nil {
		return "", fmt.Errorf("unmarshal params: %w", err)
	}

	currency := strings.ToLower(txParams.Currency)
	usdAccount := ""
	arsAccount := ""
	var accountID string
	if currency == "usd" {
		accountID = usdAccount
	} else if currency == "ars" {
		accountID = arsAccount
	} else {
		return "", fmt.Errorf("invalid currency")
	}

	// Create PocketSmith API request payload
	psRequest := PocketSmithTransactionRequest{
		Payee:      txParams.Merchant,
		Amount:     txParams.Value,
		Date:       txParams.Date,
		IsTransfer: false,
	}

	// Marshal request body
	requestBody, err := json.Marshal(psRequest)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	authKey, err := variables.Get("pocketsmith_api_key")
	if err != nil {
		return "", fmt.Errorf("fetch pocketsmith api key: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("https://api.pocketsmith.com/v2/transaction_accounts/%s/transactions", accountID)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("accept", "application/json")
	httpReq.Header.Set("content-type", "application/json")
	httpReq.Header.Set("X-Developer-Key", authKey)

	// Send request to PocketSmith API
	resp, err := spinhttp.Send(httpReq)
	if err != nil {
		return "", fmt.Errorf("send request to PocketSmith: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response from PocketSmith: %w", err)
	}

	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("PocketSmith request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	return "{\"result\": \"ok\"}", nil
}

func logRequest(method, path string, statusCode int) {
	log.Printf("- %s %d %s\n", method, statusCode, path)
}

func main() {}
