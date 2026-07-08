# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

This is a Go library for interacting with the QuickBooks Online API. It is a fork of https://github.com/rwestlund/quickbooks-go that aims to improve the design, particularly leveraging Go generics where possible.

## Commands

### Build and Test
```bash
# Build the entire project
go build ./...

# Run all tests
go test -cover .

# Run a specific test
go test -run TestAccount

# Run tests with verbose output
go test -v .

# Get dependencies
go mod download

# Tidy dependencies
go mod tidy
```

## Architecture

### Core Components

1. **Client Structure** (`client.go`): Central handler for all API interactions
   - Manages HTTP client, endpoints, authentication tokens
   - Provides base methods: `get()`, `post()`, `query()` for HTTP operations
   - Handles rate limiting (500 req/s limit)

2. **Generic Query System** (`query.go`): Already uses generics
   - `Query[T any]()` - Generic query method that unmarshals results into any type
   - `QueryPaged[T any]()` - Adds pagination support to queries

3. **Entity Pattern**: Each entity follows a consistent pattern with methods like:
   - `Create{Entity}()` - Creates new entity
   - `Find{Entity}s()` - Retrieves all entities
   - `Find{Entity}ById()` - Retrieves single entity by ID
   - `Query{Entity}s()` - Custom query for entities
   - `Update{Entity}()` - Updates existing entity
   - Some entities also have `Delete{Entity}()` and specialized methods

### Current Entity Types
- Customer, Item, Account, Invoice, Estimate, CreditMemo
- Payment, Bill, Deposit, Vendor, Employee
- CompanyInfo, CustomerType, Attachable

### Improvement Opportunities

1. **Generic CRUD Operations**: The repetitive entity methods (`CreateCustomer`, `CreateItem`, etc.) could be replaced with generic methods like `Create[T Entity]()`.

2. **Response Handling**: Current pattern has duplicate unmarshal logic across entity types that could be genericized.

3. **Query Builder**: The query string construction is manual - could benefit from a type-safe builder pattern using generics.

4. **Find Methods**: The various `Find*` methods follow identical patterns and could be consolidated using generics.

## Testing Approach

- Uses `stretchr/testify` for assertions
- Test data stored in `data/testing/` as JSON files
- Tests unmarshal JSON responses to verify struct mapping
- No mocking framework currently in use

## Key Dependencies

- `golang.org/x/oauth2` - OAuth2 authentication
- `gopkg.in/guregu/null.v4` - Nullable types for JSON
- `github.com/stretchr/testify` - Test assertions

## API Version

- Uses QuickBooks API v3
- Default minor version: 75 (configurable)
- Supports both Production and Sandbox endpoints
