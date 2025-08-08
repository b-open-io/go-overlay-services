# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Overview

go-overlay-services is a standalone HTTP server that provides a customizable interface for interacting with Overlay Services built on top of the Bitcoin SV blockchain. It implements the BSV Overlay Services architecture with support for multiple topic managers, lookup services, and GASP (Generic Application Sync Protocol) synchronization.

## Commands

### Running Tests
```bash
# Run all tests with fail-fast and vet checks
go test ./... -failfast -vet=all -count=1

# Run a specific test
go test ./pkg/core/engine/tests -run TestEngineSubmit

# Alternative using task runner
task execute-unit-tests
```

### Code Generation
```bash
# Generate OpenAPI server code and models
go generate ./...

# Alternative using task runner
task oapi-codegen
```

### Linting
```bash
# Run all linters
golangci-lint run --config=./.golangci-lint.yml
golangci-lint run --config=./.golangci-style.yml --fix

# Alternative using task runner
task execute-linters
```

### Running the Server
```bash
# Run with default config
go run examples/srv/main.go

# Run with custom config
go run examples/srv/main.go -config ./app-config.example.yaml
```

### API Documentation
```bash
# Bundle OpenAPI spec
task swagger-doc-gen

# Start Swagger UI
task swagger-ui-up

# Stop Swagger UI
task swagger-ui-down
```

## Architecture

### Core Components

1. **Engine (`pkg/core/engine/`)** - Central orchestrator that:
   - Manages topic managers and lookup services
   - Handles transaction submission and validation
   - Performs SPV verification
   - Manages UTXO state and history
   - Coordinates GASP synchronization

2. **Storage Interface (`pkg/core/engine/storage.go`)** - Abstract storage layer for:
   - UTXOs and outputs
   - Applied transactions
   - Merkle proofs
   - GASP interaction scores

3. **GASP (`pkg/core/gasp/`)** - Generic Application Sync Protocol implementation:
   - Peer-to-peer synchronization
   - Initial sync requests and responses  
   - Node-based data fetching
   - Unidirectional and bidirectional sync modes

4. **HTTP Server (`pkg/server/`)** - Fiber-based HTTP server with:
   - OpenAPI-generated handlers
   - Bearer token authentication
   - Request tracing and logging
   - CORS support
   - Health checks

### Request Flow

1. **Transaction Submission** (`/api/v1/submit`):
   - Receives TaggedBEEF (transaction + topics)
   - SPV verification via ChainTracker
   - Topic managers identify admissible outputs
   - Storage updates (mark spent, insert new)
   - Lookup service notifications
   - Optional broadcasting

2. **Lookup Queries** (`/api/v1/lookup`):
   - Routes to appropriate lookup service
   - Hydrates outputs with BEEF data
   - Supports history traversal
   - Returns OutputList or Freeform answers

3. **GASP Sync** (`/api/v1/admin/startGASPSync`):
   - Discovers peers via SHIP advertisements
   - Initiates sync for configured topics
   - Tracks interaction scores
   - Supports concurrent syncing

### Key Interfaces

- **TopicManager**: Identifies admissible outputs for a topic
- **LookupService**: Handles lookup queries and output events
- **Storage**: Persistent storage for outputs and transactions
- **ChainTracker**: SPV verification and merkle proof handling
- **Advertiser**: Creates/revokes SHIP/SLAP advertisements

## Configuration

Server configuration via `Config` struct:
- `AdminBearerToken` - Auth token for admin endpoints
- `ARCCallbackToken` - Token for ARC merkle proof callbacks
- `Port`/`Addr` - Server binding configuration
- `OctetStreamLimit` - Max request body size

Engine configuration:
- `Managers` - Map of topic managers
- `LookupServices` - Map of lookup services
- `SyncConfiguration` - Per-topic sync settings
- `SHIPTrackers`/`SLAPTrackers` - Discovery endpoints

## Testing Patterns

Tests use table-driven patterns with:
- Mock providers in `testabilities/`
- Test fixtures for HTTP handlers
- Context-based cancellation
- Assertion via `stretchr/testify`

## Important Notes

- All OpenAPI code is generated - do not edit `*.gen.go` files directly
- Linting has specific exclusions for test files and internal packages
- Two linting configs: `.golangci-lint.yml` (functional) and `.golangci-style.yml` (style)
- Task runner (`task`) provides consistent command interface
- Bearer tokens are auto-generated UUIDs if not configured