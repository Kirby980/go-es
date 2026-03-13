# Architecture

**Analysis Date:** 2026-03-13

## Pattern Overview

**Overall:** Fluent Builder Pattern with Interface-based Client Abstraction

**Key Characteristics:**
- **Builder Pattern Chain**: Every major operation (search, document, index, bulk) uses fluent builders (inspired by GORM)
- **Generic Type Parameters**: BoolQuery uses Go generics (`BoolQuery[T]`) to share query logic across multiple builder types
- **Interface-based Inversion of Control**: Builders depend on `ESClient` interface, not concrete `client.Client`, enabling test doubles and flexibility
- **Composition over Inheritance**: Common functionality (debug, headers, base building) injected via embedded struct composition
- **Request/Response Type Safety**: Dedicated response structs for each operation type rather than generic maps

## Layers

**Client Layer:**
- Purpose: HTTP communication, connection pooling, retry logic, authentication
- Location: `client/client.go`
- Contains: HTTP client initialization, request execution, address round-robin, error parsing
- Depends on: `config` (configuration), `logger`, `errors` (error types)
- Used by: All builders via the `ESClient` interface

**Builder Layer:**
- Purpose: Fluent API for composing Elasticsearch queries and operations
- Location: `builder/` directory (17 Go files)
- Contains: SearchBuilder, DocumentBuilder, IndexBuilder, BulkBuilder, and specialized builders
- Depends on: `ESClient` interface, `config`, `logger`, `errors`
- Used by: Application code directly or via Sugar layer

**Configuration Layer:**
- Purpose: Centralized configuration management using functional options pattern
- Location: `config/config.go`
- Contains: Config struct, DefaultConfig, Option functions for WithAddresses, WithAuth, WithTimeout, etc.
- Depends on: `logger` interface
- Used by: Client initialization

**Sugar Layer:**
- Purpose: High-level convenience API (GORM-style) for common operations
- Location: `sugar/sugar.go`
- Contains: Simplified methods like Create, Update, Delete, Find wrapping builders
- Depends on: Builder layer, Client
- Used by: Applications preferring simpler API over builder chains

**Logger Layer:**
- Purpose: Pluggable logging interface for all debug/diagnostic output
- Location: `logger/logger.go`
- Contains: Logger interface, default zap-based implementation, NopLogger for silent mode
- Depends on: go.uber.org/zap
- Used by: Client, all builders, debug helpers

**Error Layer:**
- Purpose: Custom Elasticsearch error type with status-code aware helpers
- Location: `errors/errors.go`
- Contains: ESError struct with IsNotFound(), IsConflict(), IsBadRequest(), IsTimeout() methods
- Depends on: Standard library only
- Used by: Client (error parsing), all builders (error handling)

**Constants Layer:**
- Purpose: Pre-defined field types, analyzer names, and ES constants
- Location: `const/` directory
- Contains: FieldTypeText, FieldTypeKeyword, AnalyzerIKSmart, etc.
- Depends on: Nothing
- Used by: IndexBuilder for field type definitions

## Data Flow

**Search Query Flow:**

1. User creates SearchBuilder: `builder.NewSearchBuilder(client, "index")`
2. Adds filters: `.Match("field", "value").Term("status", "active")`
3. BoolQuery methods append to internal `must`, `filters`, `should`, `mustNot` slices
4. `Do(ctx)` invoked → calls buildBoolQuery() to create bool query structure
5. Query is marshaled to JSON and sent via client.Do(method, path, body)
6. Response unmarshaled into SearchResponse struct
7. User calls resp.Scan(&results) to unmarshal hits into target struct slice

**Document CRUD Flow:**

1. User creates DocumentBuilder: `builder.NewDocumentBuilder(client, "index")`
2. Sets document data: `.ID("1").Set("field", value)` or `.Model(struct)`
3. Document fields accumulated in internal `doc` map
4. `Do(ctx)` for create/index, `Update(ctx)` for partial update, `Upsert(ctx)` for upsert
5. Appropriate HTTP method (PUT/POST) and path constructed
6. Request sent via client with JSON body
7. Response unmarshaled into DocumentResponse

**Bulk Operation Flow:**

1. User creates BulkBuilder: `builder.NewBulkBuilder(client)`
2. Chains operations: `.Add(index, id, doc).Update(...).Delete(...)`
3. Each operation appended to internal `operations` slice with metadata and body
4. Optional `AutoFlushSize(n)` causes automatic batch sends when reaching threshold
5. `Do(ctx)` builds NDJSON-formatted bulk request (metadata\ndocument per line)
6. Sent to `/_bulk` endpoint with custom HTTP request (not JSON marshaling)
7. Response parsed for per-item status (BulkResponse.Items)

**Index Migration Flow (AutoMigrate):**

1. User defines struct with `es:` tags: `Name string \`es:"type:text;analyzer:ik_smart"\``
2. Sugar.AutoMigrate(model) extracts struct fields and their tags
3. Generates mapping from field types and analyzer tags
4. Checks if index exists via IndexBuilder.Exists()
5. If exists: calls PutMapping() to update mappings (additive only)
6. If not: calls Create() to initialize new index with full settings/mappings

**State Management:**

- **SearchBuilder state**: query conditions (must/filters/should/mustNot), pagination (from/size), aggregations, sort specs accumulated into internal maps during chain
- **DocumentBuilder state**: field values accumulated in `doc` map; refresh strategy persists in `refresh` field
- **BulkBuilder state**: operations queue in `operations` slice; autoFlushSize and onFlush callback persist across Add() calls
- **Debug state**: per-builder `debug` boolean and `logger` reference; set via Debug() method, persists until explicitly disabled
- **Builder state NOT shared**: Each builder instance is independent; builders are NOT thread-safe and should not be shared across goroutines

## Key Abstractions

**ESClient Interface:**
- Purpose: Defines contract between builders and HTTP client without tight coupling
- Examples: `client/client.go` implements this; builders accept interface in constructor
- Pattern: Enables testing with mock clients; allows custom client implementations
- Methods: Do(), DoWithHeader(), DoRequest(), GetAddress(), GetLogger()

**IndexName Interface:**
- Purpose: Custom index naming for models (similar to GORM's Tabler)
- Examples: Struct implementing `func (p *Product) IndexName() string { return "my_products" }`
- Pattern: Allows model-aware index selection without explicit parameters

**BoolQuery[T] Generic Type:**
- Purpose: Shared query building logic (Match, Term, Range, etc.) used by multiple builders
- Examples: Embedded in SearchBuilder, ScrollBuilder, SearchAfterBuilder, UpdateByQueryBuilder, DeleteByQueryBuilder
- Pattern: Generic type parameter T points back to embedding type for method chaining
- Reduces duplication: Same Must/Filter/Should/MustNot methods available on all query builders

**Builder Base Types:**
- debugHelper: Embedded in all builders; provides Debug() method and internal debug logging
- baseBuilder: Provides Header() method for custom HTTP headers (recently added)
- Response types: DocumentResponse, SearchResponse, BulkResponse, AggregationResponse, etc. with JSON serialization methods

## Entry Points

**Client Initialization:**
- Location: `client/client.go:New(opts ...config.Option)`
- Triggers: Application startup
- Responsibilities: Parse options, set up HTTP transport with connection pooling, initialize logger, configure retry/timeout

**Builder Creation:**
- Locations: `builder/search.go:NewSearchBuilder()`, `builder/document.go:NewDocumentBuilder()`, etc.
- Triggers: For each ES operation (search, create, update, delete, etc.)
- Responsibilities: Initialize builder state, embed logger from client, set defaults (size=10 for search, empty doc map for documents)

**Sugar API:**
- Location: `sugar/sugar.go:New(client)`, `.Create(ctx, model)`, `.Find(index)`, etc.
- Triggers: High-level GORM-style operations
- Responsibilities: Wrap builder construction and method calls; auto-infer index names from model structs

## Error Handling

**Strategy:** Errors propagated up to caller; specific error types checked with helpers before retry decisions

**Patterns:**

1. **HTTP Status Code Errors**: Client.Do() parses ES error responses into `errors.ESError` with status code and ES error type/reason
2. **Error Checking Helpers**: ESError provides IsNotFound(), IsConflict(), IsBadRequest(), IsTimeout() for caller decisions
3. **Builder Error Accumulation**: DocumentBuilder and BulkBuilder may accumulate errors in an `err` field during chain construction (e.g., JSON marshal failures), returned on Do()
4. **Panic in Edge Cases**: BulkBuilder may panic if Set() called without prior Add/Create/Update (should be replaced with error accumulation per OPTIMIZATION.md issue #3)
5. **Silent Error Drops**: Exists() methods drop all errors and return false (per OPTIMIZATION.md issue #4, should distinguish 404 from network errors)

## Cross-Cutting Concerns

**Logging:**
- Via logger.Logger interface injected from client
- DebugHelper in builders accesses logger.GetLogger()
- All debug output goes through logger, not fmt.Printf (recently improved)
- Application can inject custom logger via config.WithLogger()

**Validation:**
- Occurs on-chain during builder construction (e.g., validation of snapshot settings in ClusterBuilder)
- Some validation deferred to request time (e.g., checking required fields only when Do() called)
- Limited structural validation; relies on ES server to reject invalid queries

**Authentication:**
- HTTP Basic Auth configured in client.Config (username/password)
- Applied in client.DoWithHeader() for all requests
- No support for token-based auth (would require Config extension)

**Request Tracing:**
- Custom HTTP headers supported via builder.Header(key, value) method (baseBuilder)
- Allows X-Request-ID, X-Opaque-Id per ES docs
- Headers accumulated in map and passed to client.DoWithHeader()

**Retry Logic:**
- In client.Do() and client.DoRequest()
- Fixed backoff interval configurable via config.WithRetry(maxRetries, backoff)
- Retries network errors only; does NOT retry 5xx status codes (per OPTIMIZATION.md issue #9, should be enhanced)
- Body reset via io.Seeker for idempotent retries

**Connection Pooling:**
- HTTP Transport configured with MaxIdleConns, MaxIdleConnsPerHost, MaxConnsPerHost, IdleConnTimeout
- Tunable via config options
- Multiple ES addresses supported with round-robin address selection via atomic.Uint32 index

---

*Architecture analysis: 2026-03-13*
