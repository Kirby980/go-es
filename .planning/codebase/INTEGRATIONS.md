# External Integrations

**Analysis Date:** 2026-03-13

## APIs & External Services

**Elasticsearch:**
- Service: Elasticsearch (externally managed)
- What it's used for: Search, indexing, aggregation, cluster operations
  - SDK/Client: Raw HTTP client (Go stdlib `net/http`)
  - Auth: Basic Auth (username/password) via `req.SetBasicAuth()` in `client/client.go` lines 97-98, 142-144, 195-197
  - Configuration: Via `config.WithAddresses()` and `config.WithAuth()`
  - Protocol: HTTP/HTTPS
  - Default timeout: 30 seconds (configurable)

**No other external APIs integrated.**

This is a client library, not a service. All integrations are provided by the user (client application).

## Data Storage

**Databases:**
- Elasticsearch (externally managed by user)
  - Connection: Specified via `WithAddresses()` option
  - Client: Raw HTTP using Go stdlib `net/http`
  - No ORM or query builder dependency

**File Storage:**
- None - Library does not handle file I/O
- Document content stored in Elasticsearch

**Caching:**
- None - No caching layer in client

## Authentication & Identity

**Auth Provider:**
- Custom Basic Auth (username/password)
  - Implementation: HTTP Basic Authentication (RFC 7617)
  - Transport: TLS/SSL (configurable via `WithInsecureSkipVerify()`)
  - Location: Enforced in request phase via `req.SetBasicAuth()` in `client/client.go`

**No OAuth, SAML, or JWT support.**

User must provide username and password at client creation:
```go
config.WithAuth("elastic", "password")
```

## Monitoring & Observability

**Error Tracking:**
- None - No integration with error tracking services (Sentry, etc.)
- Errors are parsed by `errors.ParseESError()` in `errors/errors.go`
- Error type detection: `IsNotFound()`, `IsConflict()`, `IsBadRequest()`, `IsTimeout()` methods

**Logs:**
- Structured logging via `go.uber.org/zap`
- Two built-in formats:
  - Production: JSON format, stderr output, Info level and above
  - Development: Human-readable, Debug level and above
- Custom logger injection via `config.WithLogger()`
- Debug mode: `config.WithDebug()` enables verbose retry logging (lines 113-115 in `client/client.go`)
- No external log aggregation (ELK, Datadog, etc.)

**Log Behavior:**
- Default logger created if none supplied: `logger.NewDefaultLogger()` in `client/client.go` lines 40-45
- User can disable logging entirely: `config.WithLogger(logger.NopLogger{})`

## CI/CD & Deployment

**Hosting:**
- Not applicable - This is a client library
- User responsible for deployment and Elasticsearch infrastructure

**CI Pipeline:**
- None detected in repository
- Tests run via standard `go test` command
- Test directory: `test/` (integration tests)
- Example tests: `examples/complete_api_test.go`

## Environment Configuration

**No environment variables used by the client library itself.**

All configuration is code-based via option functions:
- `config.WithAddresses()`
- `config.WithAuth()`
- `config.WithTimeout()`
- `config.WithRetry()`
- `config.WithDebug()`
- Connection pool options
- Logger injection

**User applications** may read environment variables and pass them to these options:
```go
client.New(
    config.WithAddresses(os.Getenv("ES_ADDRESSES")),
    config.WithAuth(os.Getenv("ES_USERNAME"), os.Getenv("ES_PASSWORD")),
)
```

## Webhooks & Callbacks

**Incoming:**
- None - Library is a client, not a server

**Outgoing:**
- None - Library makes requests to Elasticsearch only

## Network

**Elasticsearch Communication:**
- HTTP/HTTPS only
- Configurable addresses: Can connect to multiple Elasticsearch nodes
- Round-robin address selection: Via `GetAddress()` in `client/client.go` line 89-91
  - Atomic counter ensures thread-safe address distribution
- Custom headers: Supported via `DoWithHeader()` in `client/client.go` and `Header()` method in `builder/base.go`

**Connection Pooling (HTTP Transport):**
- `MaxIdleConns`: 100 (global)
- `MaxIdleConnsPerHost`: 10 (per node)
- `MaxConnsPerHost`: 0 (unlimited per node)
- `IdleConnTimeout`: 90 seconds
- TLS verification: Enabled by default, can be disabled per request

## Request/Response Handling

**Request Format:**
- JSON (default)
- Custom headers: Supported since custom headers feature (PR #9)
- Content-Type: Always set to "application/json" in `DoWithHeader()` line 188

**Response Format:**
- Raw bytes (JSON) returned from Elasticsearch
- User must unmarshal to structs/types

**Error Handling:**
- HTTP status codes >= 400 parsed as ES errors
- Error parsing in `errors.ParseESError()` extracts error type and reason
- Retry mechanism with exponential backoff support (user can implement)

## Rate Limiting

**Elasticsearch Rate Limits:**
- No client-side rate limiting
- User responsible for respecting ES cluster limits
- Backpressure: Retry backoff can be configured via `WithRetry()` option

---

*Integration audit: 2026-03-13*
