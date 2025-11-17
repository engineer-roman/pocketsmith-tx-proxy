package domain

// Transaction represents a financial transaction
type Transaction struct {
	Currency string
	Category string
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
	CategoryID *int   `json:"category_id,omitempty"`
}

// RPCRequest represents a JSON-RPC request
type RPCRequest struct {
	Method string         `json:"method"`
	Params map[string]any `json:"params"`
}

// TransactionParams represents the parameters for adding a transaction
type TransactionParams struct {
	Currency string `json:"currency"`
	Category string `json:"category"`
	Merchant string `json:"merchant"`
	Value    string `json:"value"`
	Date     string `json:"date"`
}

// User represents a PocketSmith user
type User struct {
	ID int `json:"id"`
}

// TransactionAccount represents a PocketSmith transaction account
type TransactionAccount struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	CurrencyCode string `json:"currency_code"`
}

// Category represents a PocketSmith category
type Category struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	ParentID *int   `json:"parent_id"`
}

// AccountInfo represents account information for the client
type AccountInfo struct {
	Name     string `json:"name"`
	Currency string `json:"currency"`
}

// ShortcutEntities represents combined accounts and categories data
type ShortcutEntities struct {
	Accounts   []AccountInfo `json:"accounts"`
	Categories []string      `json:"categories"`
}
