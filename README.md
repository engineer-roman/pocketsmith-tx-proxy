# PocketSmith Proxy

A WebAssembly-based JSON-RPC proxy for PocketSmith API, built with Spin Framework and clean architecture principles.

## Overview

This service provides a JSON-RPC interface for adding transactions to PocketSmith, designed to be called from iOS Shortcuts or other automation tools. It implements a clean architecture with clear separation of concerns across three layers.

## Features

- **Clean Architecture**: Separated layers for cache, API client, business logic, and HTTP handling
- **Redis Caching**: 24-hour TTL cache for user data, accounts, and categories
- **JSON-RPC endpoint** for adding transactions
- **Bearer token authentication** for API security
- **Dynamic account and category lookup** - automatically finds accounts by name and categories by title
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
│  - Dynamic account lookup by name       │
│  - Dynamic category lookup by title     │
│  - Data transformation                  │
│  - Business rules                       │
└──────────────┬──────────────────────────┘
               │ uses interface
┌──────────────▼──────────────────────────┐
│         API Client Layer                │
│  - Cache-first data retrieval           │
│  - PocketSmith API communication        │
│  - HTTP requests to external API        │
└──────────────┬──────────────────────────┘
               │ uses interface
┌──────────────▼──────────────────────────┐
│       Cache Repository Layer            │
│  - Redis operations (Get/Set/HSET)      │
│  - 24-hour TTL management               │
│  - User, accounts, categories caching   │
└─────────────────────────────────────────┘
```

### Project Structure

```
.
├── main.go                           # Entry point & dependency injection
├── internal/
│   ├── domain/
│   │   └── transaction.go           # Domain models
│   ├── repository/
│   │   └── cache_repository.go      # Redis cache operations (interface + impl)
│   ├── api/
│   │   └── pocketsmith_client.go    # PocketSmith API client (interface + impl)
│   ├── service/
│   │   └── transaction_service.go   # Business logic (interface + impl)
│   └── handler/
│       └── http_handler.go          # HTTP request handling
├── spin.toml                         # Spin configuration
├── go.mod                            # Go module definition
├── .env.local.example                # Example environment variables
├── Taskfile.yml                      # Task automation (build, dev, deploy, etc.)
└── README.md                         # This file
```

## Prerequisites

- [Spin](https://developer.fermyon.com/spin/v2/install) v2.x
- [TinyGo](https://tinygo.org/getting-started/install/) for building WASM modules
- [Task](https://taskfile.dev/) (optional, for using Taskfile automation)
- [Redis](https://redis.io/) instance (local or remote)
- PocketSmith API key (get from https://my.pocketsmith.com/settings/api)

## Configuration

The application requires **two** environment variables and has one optional configuration:

### Required Variables

1. **`client_auth_key`** - Bearer token for authenticating incoming requests from your client (e.g., iOS Shortcuts)
2. **`pocketsmith_api_key`** - Your PocketSmith API developer key

### Optional Variables

3. **`redis_address`** - Redis connection string (defaults to `redis://localhost:6379`)

### Redis Caching

The application uses Redis to cache PocketSmith API data for 24 hours:
- **User ID**: Simple key-value with TTL
- **Transaction Accounts**: Hash set with TTL (keyed by user ID)
- **Categories**: Hash set with TTL (keyed by user ID)

This significantly reduces API calls and improves response times. Make sure you have a Redis instance running locally or provide a custom `redis_address`.

## Authentication Practice

This proxy uses a two-tier authentication approach to minimize API calls:

1. **Client → Proxy**: Shared secret (Bearer token) configured via `client_auth_key`
   - Set the same key on both the client (iOS Shortcuts) and server (via env var)
   - Client sends this as `Authorization: Bearer <client_auth_key>` header

2. **Proxy → PocketSmith**: API key authentication via `pocketsmith_api_key`
   - The proxy authenticates to PocketSmith on behalf of the client
   - Client never needs to know the PocketSmith API key

**Security Notes:**
- Keep `client_auth_key` and `pocketsmith_api_key` different
- Always use HTTPS in production
- Rotate keys periodically
- Never commit credentials to version control

## Building

### Using Task (Recommended)

```bash
# Build the WASM binary
task build

# Build and run with live reload
task dev

# Clean build artifacts
task clear

# Sync Go dependencies
task sync

# Deploy to Fermyon Cloud
task deploy

# Show all available tasks
task
```

### Using Spin Directly

```bash
spin build
```

## Running Locally

### Prerequisites
Make sure Redis is running:
```bash
redis-server
# or use Docker:
docker run -d -p 6379:6379 redis:latest
```

### Option 1: Using .env.local (Recommended)

1. Copy the example configuration:
```bash
cp .env.local.example .env.local
```

2. Edit `.env.local` with your actual values:
```bash
SPIN_VARIABLE_CLIENT_AUTH_KEY=your-client-bearer-token
SPIN_VARIABLE_POCKETSMITH_API_KEY=your-pocketsmith-api-key
SPIN_VARIABLE_REDIS_ADDRESS=redis://localhost:6379
```

3. Run the service using Task (which automatically loads `.env.local`):
```bash
task dev
```

### Option 2: Using Environment Variables Directly

You can also set variables directly when running:
```bash
SPIN_VARIABLE_CLIENT_AUTH_KEY="your-token" \
SPIN_VARIABLE_POCKETSMITH_API_KEY="your-api-key" \
SPIN_VARIABLE_REDIS_ADDRESS="redis://localhost:6379" \
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
    "account": "USD General",
    "category": "Eating out",
    "merchant": "Coffee Shop",
    "value": "-5.50",
    "date": "2025-01-13"
  }
}
```

#### Parameters

- **`account`** (string, required): Account name (e.g., `USD General`, `ARS General`) - must match an account name in your PocketSmith (case-insensitive)
- **`category`** (string, required): Category title - must match a category in your PocketSmith (case-insensitive)
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
      "account": "USD General",
      "category": "Groceries",
      "merchant": "Grocery Store",
      "value": "-42.50",
      "date": "2025-01-13"
    }
  }'
```

## Deployment

### Deploy to Fermyon Cloud

```bash
task deploy
# or
spin cloud deploy
```

### Set the required variables:

```bash
spin cloud variables set client_auth_key="your-client-bearer-token"
spin cloud variables set pocketsmith_api_key="your-pocketsmith-api-key"
```

### Optional: Configure Redis

If using a remote Redis instance:
```bash
spin cloud variables set redis_address="redis://your-redis-host:6379"
```

Note: For production deployments, consider using Fermyon Cloud's built-in key-value store or a managed Redis service.

## iOS Shortcuts Integration

Example Shortcuts workflow:

1. **Get transaction details** from user input (merchant, amount, account name, category, date)
2. **Make HTTP POST request** to the proxy endpoint
3. **Include Authorization header** with your Bearer token (same as `client_auth_key`)
4. **Send JSON body** with transaction details in JSON-RPC format
5. **Display result** to user

### Example iOS Shortcut Steps

1. Ask for input: Merchant name
2. Ask for input: Amount (number)
3. Ask for input: Account (choose from: USD General, ARS General, etc.)
4. Ask for input: Category (choose from: Groceries, Eating out, etc.)
5. Get current date
6. Create dictionary with transaction data
7. Get contents of URL:
   - URL: `https://your-app.fermyon.app/api/v1/transactions/append`
   - Method: POST
   - Headers:
     - `Content-Type: application/json`
     - `Authorization: Bearer your-client-bearer-token`
   - Request Body: JSON from dictionary

## Error Handling

The service returns appropriate HTTP status codes:

- **200 OK**: Transaction created successfully
- **400 Bad Request**: Invalid request format, missing required fields, or entity not found (account or category)
- **403 Forbidden**: Invalid or missing authentication token
- **405 Method Not Allowed**: HTTP method is not POST
- **422 Unprocessable Entity**: Invalid amount format (e.g., multiple decimal separators)
- **500 Internal Server Error**: Server-side error (check logs)

When an account or category is not found, detailed error messages are logged indicating:
- The exact search string used
- How many entities were searched
- Whether the lookup was from cache or API

## Development

### Adding Support for New Accounts or Categories

No configuration changes needed! The application **automatically** discovers:
- All transaction accounts from your PocketSmith (sorted alphabetically)
- All categories from your PocketSmith (sorted alphabetically)
- Net worth accounts are automatically filtered out

Simply use the correct account name or category title in your API requests. The service will:
1. Check the cache first (24-hour TTL)
2. Fall back to PocketSmith API if not cached
3. Sort the data alphabetically before caching
4. Return a 400 error with detailed logging if the entity is not found

**Note:** Account and category lookups are case-insensitive, so "USD General", "usd general", and "USD GENERAL" will all match the same account.

### Running Tests

```bash
task test
# or
go test ./...
```

### Building for Development

```bash
task dev
# or
spin build --up
```

This builds and runs the service in one command, with automatic rebuilding on file changes.

### Cache Management

To clear the cache during development:
```bash
redis-cli FLUSHALL
```

Or to remove specific keys:
```bash
redis-cli DEL user:id
redis-cli DEL user:{USER_ID}:accounts
redis-cli DEL user:{USER_ID}:categories
```

## License

MIT
