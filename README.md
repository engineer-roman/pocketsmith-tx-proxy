# PocketSmith Proxy

A WebAssembly-based JSON-RPC proxy for PocketSmith API, built with Spin Framework.

## Overview

This service provides a JSON-RPC interface for adding transactions to PocketSmith, designed to be called from iOS Shortcuts or other automation tools.

## Features

- JSON-RPC endpoint for adding transactions
- Bearer token authentication for API security
- Automatic currency-based account routing (USD/ARS)
- Amount normalization (handles comma/dot decimal separators)
- Built as a WASM module using Spin Framework

## Prerequisites

- [Spin](https://developer.fermyon.com/spin/v2/install) v2.x
- [TinyGo](https://tinygo.org/getting-started/install/) for building WASM modules
- PocketSmith API key (get from https://my.pocketsmith.com/settings/api)

## Configuration

The application requires two environment variables:

1. `pocketsmith_client_key` - Bearer token for authenticating incoming requests
2. `pocketsmith_api_key` - Your PocketSmith API key

### Account Configuration

The proxy currently routes transactions to hardcoded account IDs based on currency:
- USD → Account ID: `123`
- ARS → Account ID: `124`

To use your own accounts, update the account IDs in `main.go`:

```go
usdAccount := "YOUR_USD_ACCOUNT_ID"
arsAccount := "YOUR_ARS_ACCOUNT_ID"
```

## Building

```bash
spin build
```

## Running Locally

Create a `local.toml` file with your credentials:

```toml
[variables]
pocketsmith_client_key = "your-client-bearer-token"
pocketsmith_api_key = "your-pocketsmith-api-key"
```

Then run:

```bash
spin up
```

The service will be available at `http://localhost:3000/pocketsmith/api/v1/jsonrpc`

## API Usage

### Endpoint

```
POST /pocketsmith/api/v1/jsonrpc
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

### Response

Success:
```json
{"result": "ok"}
```

Error:
```json
{"error": "error message"}
```

## Deployment

Deploy to Fermyon Cloud:

```bash
spin cloud deploy
```

Set the required variables:

```bash
spin cloud variables set pocketsmith_client_key="your-client-bearer-token"
spin cloud variables set pocketsmith_api_key="your-pocketsmith-api-key"
```

## iOS Shortcuts Integration

Example Shortcuts workflow:

1. Get transaction details from user input
2. Make HTTP POST request to the proxy endpoint
3. Include Authorization header with Bearer token
4. Send JSON body with transaction details

## Development

The project structure:
```
.
├── main.go          # Main application code
├── go.mod           # Go module definition
├── spin.toml        # Spin configuration
├── .gitignore       # Git ignore rules
└── README.md        # This file
```

## License

MIT
