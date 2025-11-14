# PocketSmith Proxy

A WebAssembly-based JSON-RPC proxy for PocketSmith API, built with Spin Framework and clean architecture principles.

## Overview

This service provides a JSON-RPC interface for adding transactions to PocketSmith, designed to be called from iOS Shortcuts or other automation tools. It implements a clean architecture with clear separation of concerns across three layers.

## Features

- **Clean Architecture**: Separated layers for API client, business logic, and HTTP handling
- **JSON-RPC endpoint** for adding transactions
- **Bearer token authentication** for API security
- **Automatic currency-based account routing** (USD/ARS)
- **Amount normalization** (handles comma/dot decimal separators)
- **Environment-based configuration** (no hardcoded credentials)
- **Built as a WASM module** using Spin Framework

## Architecture

The application follows clean architecture principles with three distinct layers:

```
┌─────────────────────────────────────────┐
│         Handler Layer (Facade)          │
│  - HTTP request/response handling       │
│  - Authentication & validation          │
│  - JSON-RPC parsing                     │
└──────────────┬──────────────────────────┘
               │ uses interface
┌──────────────▼──────────────────────────┐
│         Service Layer (Business)        │
│  - Currency routing logic               │
│  - Data transformation                  │
│  - Business rules                       │
└──────────────┬──────────────────────────┘
               │ uses interface
┌──────────────▼──────────────────────────┐
│         API Client Layer                │
│  - PocketSmith API communication        │
│  - HTTP requests to external API        │
└─────────────────────────────────────────┘
```

### Project Structure

```
.
├── main.go                           # Entry point & dependency injection
├── internal/
│   ├── domain/
│   │   └── transaction.go           # Domain models
│   ├── api/
│   │   └── pocketsmith_client.go    # PocketSmith API client (interface + impl)
│   ├── service/
│   │   └── transaction_service.go   # Business logic (interface + impl)
│   └── handler/
│       └── http_handler.go          # HTTP request handling
├── spin.toml                         # Spin configuration
├── go.mod                            # Go module definition
├── local.toml.example                # Example local configuration
└── README.md                         # This file
```

## Prerequisites

- [Spin](https://developer.fermyon.com/spin/v2/install) v2.x
- [TinyGo](https://tinygo.org/getting-started/install/) for building WASM modules
- PocketSmith API key (get from https://my.pocketsmith.com/settings/api)
- PocketSmith account IDs for your currencies

## Configuration

The application requires **four** environment variables (no hardcoded credentials):

### Required Variables

1. **`client_key`** - Bearer token for authenticating incoming requests from your client (e.g., iOS Shortcuts)
2. **`pocketsmith_api_key`** - Your PocketSmith API developer key
3. **`usd_account_id`** - PocketSmith transaction account ID for USD transactions
4. **`ars_account_id`** - PocketSmith transaction account ID for ARS transactions

### Finding Your Account IDs

To find your PocketSmith account IDs:
1. Visit https://my.pocketsmith.com
2. Navigate to your accounts
3. Click on an account to view its details
4. The account ID is in the URL: `https://my.pocketsmith.com/accounts/{account_id}`

Or use the PocketSmith API:
```bash
curl -H "X-Developer-Key: YOUR_API_KEY" \
  https://api.pocketsmith.com/v2/transaction_accounts
```

## Authentication Practice

This proxy uses a two-tier authentication approach to minimize API calls:

1. **Client → Proxy**: Shared secret (Bearer token) configured via `client_key`
   - Set the same key on both the client (iOS Shortcuts) and server (via env var)
   - Client sends this as `Authorization: Bearer <client_key>` header

2. **Proxy → PocketSmith**: API key authentication via `pocketsmith_api_key`
   - The proxy authenticates to PocketSmith on behalf of the client
   - Client never needs to know the PocketSmith API key

**Security Notes:**
- Keep `client_key` and `pocketsmith_api_key` different
- Always use HTTPS in production
- Rotate keys periodically
- Never commit credentials to version control

## Building

```bash
spin build
```

## Running Locally

1. Copy the example configuration:
```bash
cp local.toml.example local.toml
```

2. Edit `local.toml` with your actual values:
```toml
[variables]
client_key = "your-client-bearer-token"
pocketsmith_api_key = "your-pocketsmith-api-key"
usd_account_id = "123456"
ars_account_id = "789012"
```

3. Run the service:
```bash
spin up
```

The service will be available at `http://localhost:3000/api/v1/transactions/append`

## API Usage

### Endpoint

```
POST /api/v1/transactions/append
```

### Headers

```
Content-Type: application/json
Authorization: Bearer <your-client-key>
```

### Request Format

```json
{
  "method": "transactions.add",
  "params": {
    "currency": "USD",
    "merchant": "Coffee Shop",
    "value": "-5.50",
    "date": "2025-01-13"
  }
}
```

#### Parameters

- **`currency`** (string, required): Currency code - `USD` or `ARS`
- **`merchant`** (string, required): Merchant/payee name
- **`value`** (string, required): Transaction amount (negative for expenses, positive for income)
  - Supports both comma (`,`) and dot (`.`) as decimal separator
  - Will be automatically normalized
- **`date`** (string, required): Transaction date in `YYYY-MM-DD` format

### Response

Success:
```json
{"result": "ok"}
```

Error:
```json
{"error": "error message"}
```

### Example cURL Request

```bash
curl -X POST http://localhost:3000/api/v1/transactions/append \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-client-bearer-token" \
  -d '{
    "method": "transactions.add",
    "params": {
      "currency": "USD",
      "merchant": "Grocery Store",
      "value": "-42.50",
      "date": "2025-01-13"
    }
  }'
```

## Deployment

Deploy to Fermyon Cloud:

```bash
spin cloud deploy
```

Set the required variables:

```bash
spin cloud variables set client_key="your-client-bearer-token"
spin cloud variables set pocketsmith_api_key="your-pocketsmith-api-key"
spin cloud variables set usd_account_id="123456"
spin cloud variables set ars_account_id="789012"
```

## iOS Shortcuts Integration

Example Shortcuts workflow:

1. **Get transaction details** from user input (merchant, amount, currency, date)
2. **Make HTTP POST request** to the proxy endpoint
3. **Include Authorization header** with your Bearer token (same as `client_key`)
4. **Send JSON body** with transaction details in JSON-RPC format
5. **Display result** to user

### Example iOS Shortcut Steps

1. Ask for input: Merchant name
2. Ask for input: Amount (number)
3. Ask for input: Currency (choose from: USD, ARS)
4. Get current date
5. Create dictionary with transaction data
6. Get contents of URL:
   - URL: `https://your-app.fermyon.app/api/v1/transactions/append`
   - Method: POST
   - Headers:
     - `Content-Type: application/json`
     - `Authorization: Bearer your-client-bearer-token`
   - Request Body: JSON from dictionary

## Error Handling

The service returns appropriate HTTP status codes:

- **200 OK**: Transaction created successfully
- **400 Bad Request**: Invalid request format or missing required fields
- **403 Forbidden**: Invalid or missing authentication token
- **405 Method Not Allowed**: HTTP method is not POST
- **422 Unprocessable Entity**: Invalid amount format (e.g., multiple decimal separators)
- **500 Internal Server Error**: Server-side error (check logs)

## Development

### Adding Support for New Currencies

1. Add a new account ID variable in `spin.toml`:
```toml
[variables]
eur_account_id = { required = true }
```

2. Update the component variables:
```toml
[component.pocketsmith-rpc.variables]
eur_account_id = "{{ eur_account_id }}"
```

3. Update `main.go` to read the new variable and pass it to the service config

4. Update `service/transaction_service.go` to handle the new currency in `getAccountID()`

### Running Tests

```bash
go test ./...
```

### Building for Development

```bash
spin build --up
```

This builds and runs the service in one command, with automatic rebuilding on file changes.

## License

MIT
