# Technology Stack

**Analysis Date:** 2026-03-13

## Languages

**Primary:**
- Go 1.24.0 - Core application language, used across all packages

**Secondary:**
- None

## Runtime

**Environment:**
- Go 1.24.0 (specified in `go.mod`)

**Package Manager:**
- Go Modules (go.mod/go.sum)
- Lockfile: Present and up to date

## Frameworks

**Core:**
- No web framework - Raw HTTP client library
- `net/http` (Go stdlib) - HTTP transport layer for Elasticsearch client

**Testing:**
- Go Testing (stdlib `testing` package)
- Test files in `test/` directory with naming pattern `*_test.go`

**Build/Dev:**
- None configured (uses `go build`, `go test`)

## Key Dependencies

**Critical:**
- `go.uber.org/zap` v1.27.1 - Structured logging framework (production-grade JSON logging)
  - Used for default logger in `logger/logger.go`
  - Supports development and production log formats
  - Required: Essential for client logging and debugging

**Infrastructure:**
- `go.uber.org/multierr` v1.10.0 - Error aggregation utility
  - Indirect dependency pulled by zap
  - Provides multi-error handling capabilities

## No External Service Dependencies

This is a client library that does NOT bundle SDKs for:
- Elasticsearch (client makes raw HTTP requests)
- Cloud providers (AWS, GCP, Azure)
- Auth services
- Monitoring tools

Library users supply Elasticsearch endpoints and credentials at client instantiation.

## Configuration

**Environment:**
Configuration is primarily code-based using option functions in `config/config.go`:
- Server addresses: Via `WithAddresses()` option
- Authentication: Via `WithAuth()` (username/password only, no OAuth/SAML)
- Timeouts: Via `WithTimeout()`
- Retry: Via `WithRetry()`
- Connection pool: Via `WithMaxIdleConns()`, `WithMaxIdleConnsPerHost()`, `WithMaxConnsPerHost()`, `WithIdleConnTimeout()`
- Custom logger: Via `WithLogger()`
- Debug mode: Via `WithDebug()`
- TLS: Via `WithInsecureSkipVerify()`

No `.env` file support or environment variable bindings in core client.

**Build:**
- No build configuration files (Makefile, docker-compose, etc.)
- Direct `go build` and `go test` execution

## Platform Requirements

**Development:**
- Go 1.24.0 or compatible
- Standard Go toolchain

**Production:**
- No external dependencies beyond Go stdlib
- Target: Any platform supporting Go (Linux, macOS, Windows)
- Deployment: Compiled binary

## HTTP Transport Configuration

**Defaults (from `config/config.go`):**
- `Timeout`: 30 seconds (connection + request timeout)
- `MaxRetries`: 3
- `RetryBackoff`: 1 second
- `MaxIdleConns`: 100
- `MaxIdleConnsPerHost`: 10
- `MaxConnsPerHost`: 0 (unlimited)
- `IdleConnTimeout`: 90 seconds
- `InsecureSkipVerify`: false (respects TLS by default)

**Retry Strategy:**
- Linear retry with fixed 1-second backoff
- Applies to both `Do()` and `DoWithHeader()` methods in `client/client.go`
- Body reset attempt for seekable request bodies (lines 205-207 in `client/client.go`)

## JSON Handling

- Uses Go stdlib `encoding/json` for all serialization/deserialization
- No third-party JSON library

## Concurrency Primitives

- `sync/atomic.Uint32` for round-robin address index in `client.Client` (thread-safe address selection)
- Client is thread-safe, Builder objects are NOT (documented in README)

---

*Stack analysis: 2026-03-13*
