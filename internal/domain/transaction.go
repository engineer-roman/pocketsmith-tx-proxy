package domain

// Transaction represents a financial transaction
type Transaction struct {
	Currency string
	Merchant string
	Amount   string
	Date     string
}

// PocketSmithTransaction represents a transaction in PocketSmith API format
type PocketSmithTransaction struct {
	Payee      string `json:"payee"`
	Amount     string `json:"amount"`
	Date       string `json:"date"`
	IsTransfer bool   `json:"is_transfer"`
}

// RPCRequest represents a JSON-RPC request
type RPCRequest struct {
	Method string         `json:"method"`
	Params map[string]any `json:"params"`
}

// TransactionParams represents the parameters for adding a transaction
type TransactionParams struct {
	Currency string `json:"currency"`
	Merchant string `json:"merchant"`
	Value    string `json:"value"`
	Date     string `json:"date"`
}
