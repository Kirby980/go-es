# Codebase Structure

**Analysis Date:** 2026-03-13

## Directory Layout

```
go-es/
├── builder/                 # Fluent builder implementations for all ES operations (17 files, 4296 LOC)
├── client/                  # HTTP client and connection management
├── config/                  # Configuration and functional options
├── const/                   # Pre-defined constants (field types, analyzers)
├── errors/                  # Custom error types and parsing
├── examples/                # Example usage and API demonstrations
├── logger/                  # Logging interface and zap-based implementation
├── sugar/                   # High-level GORM-style convenience API
├── test/                    # Comprehensive integration test suite (16 test files)
├── docs/                    # User-facing documentation
├── go.mod                   # Module definition (Go 1.24, uber zap logger)
├── README.md                # Main documentation
└── OPTIMIZATION.md          # Known issues and improvement suggestions
```

## Directory Purposes

**builder/**
- Purpose: Core fluent builder implementations for constructing and executing Elasticsearch operations
- Contains: Individual builder types, query logic, response structs, helper utilities
- Key files: search.go (444 LOC), document.go (486 LOC), index.go (614 LOC), bulk.go (565 LOC), aggregation.go (442 LOC)
- Builders: SearchBuilder, DocumentBuilder, IndexBuilder, BulkBuilder, AggregationBuilder, ScrollBuilder, SearchAfterBuilder, ClusterBuilder, DeleteByQueryBuilder, UpdateByQueryBuilder, MGetBuilder
- Utilities: query.go (BoolQuery generic type), debug.go (debug helper), base.go (baseBuilder), interface.go (ESClient interface, IndexName interface, helpers)

**client/**
- Purpose: HTTP client implementation handling connection pooling, retry logic, and request/response management
- Contains: Client struct, request execution, authentication, error parsing, address management
- Key files: client.go (236 LOC) - single file containing all client logic

**config/**
- Purpose: Configuration management using functional options pattern
- Contains: Config struct, DefaultConfig factory, Option functions
- Key files: config.go (143 LOC)
- Options: WithAddresses, WithAuth, WithTimeout, WithRetry, WithDebug, WithLogger, WithMaxIdleConns, WithMaxConnsPerHost, WithIdleConnTimeout, WithInsecureSkipVerify

**const/**
- Purpose: Pre-defined Elasticsearch field types and analyzer constants
- Contains: FieldType constants (text, keyword, integer, float, date, boolean, nested, object, geo_point, etc.), AnalyzerType constants (ik_smart, ik_max_word, standard, etc.)
- Key files: field_types.go, analyzer_constants.go

**errors/**
- Purpose: Custom error type for Elasticsearch responses with status-aware helpers
- Contains: ESError struct with parsing and checking methods
- Key files: errors.go (64 LOC)
- Methods: IsNotFound(), IsConflict(), IsBadRequest(), IsTimeout()

**examples/**
- Purpose: Demonstration code showing API usage patterns
- Contains: builder_test.go, complete_api_test.go, debug_example.go
- Value: Shows both builder and sugar patterns; useful for users learning the API

**logger/**
- Purpose: Logging abstraction and default implementation
- Contains: Logger interface, default zap-based logger, NopLogger for silent mode
- Key files: logger.go

**sugar/**
- Purpose: High-level GORM-style convenience API wrapping builders
- Contains: Sugar struct, AutoMigrate, Create, Update, Upsert, Delete, Get, Find methods
- Key files: sugar.go (82 LOC)
- Value: Simplifies common operations for users familiar with GORM ORM

**test/**
- Purpose: Comprehensive integration test suite (requires live ES instance)
- Contains: 16 test files covering search, document operations, bulk, aggregation, clustering, indexing, scrolling
- Test patterns: Test data preparation helpers, assertion utilities, concurrency tests
- Notable tests: performance_test.go (benchmarks), search_after_test.go (pagination), bulk_test.go (batch operations)

**docs/**
- Purpose: User-facing documentation for features and API usage
- Contains: index.md, document.md, search.md, aggregation.md, advanced.md, errors.md
- Format: Markdown with code examples

## Key File Locations

**Entry Points:**
- `client/client.go:New()`: Client factory accepting options
- `sugar/sugar.go:New()`: Sugar factory for GORM-style API
- `builder/search.go:NewSearchBuilder()`: Search builder entry
- `builder/document.go:NewDocumentBuilder()`: Document builder entry
- `builder/index.go:NewIndexBuilder()`: Index builder entry
- `builder/bulk.go:NewBulkBuilder()`: Bulk builder entry

**Configuration:**
- `config/config.go`: Config struct, DefaultConfig, all Option functions
- `const/field_types.go`: FieldType constants
- `const/analyzer_constants.go`: Analyzer constants
- `logger/logger.go`: Logger interface definition

**Core Logic:**
- `client/client.go`: HTTP Do(), DoWithHeader(), DoRequest(), retry logic, auth, address selection
- `builder/query.go`: BoolQuery[T] generic type with Must/Filter/Should/MustNot methods
- `builder/search.go`: SearchBuilder.Do() with query building, SearchBuilder.Count(), SearchBuilder.Scan()
- `builder/document.go`: DocumentBuilder.Do() (create), .Update(), .Upsert(), .Get(), .Delete()
- `builder/index.go`: IndexBuilder.Create(), .PutMapping(), .Exists(), AutoMigrate support
- `builder/bulk.go`: BulkBuilder.Add(), .Create(), .Update(), .Delete(), .Do() with NDJSON formatting
- `builder/aggregation.go`: AggregationBuilder with Avg, Sum, Min, Max, Terms, DateHistogram, etc.
- `builder/search_after.go`: SearchAfterBuilder for efficient deep pagination
- `builder/scroll.go`: ScrollBuilder for cursor-based iteration
- `builder/cluster.go`: ClusterBuilder for cluster admin operations

**Testing:**
- `test/search_test.go`: Search query patterns (match, term, range, bool combinations)
- `test/document_test.go`: CRUD operations and field updates
- `test/bulk_test.go`: Batch operations with auto-flush
- `test/index_test.go`: Index management, mapping, settings, AutoMigrate
- `test/aggregation_test.go`: Stats, terms, date histogram aggregations
- `test/performance_test.go`: Benchmarks for bulk and search operations
- `test/search_after_test.go`: Deep pagination patterns
- `test/gorm_style_test.go`: Sugar API tests
- `test/scroll_test.go`: Scroll cursor iteration

## Naming Conventions

**Files:**
- Plural nouns for directories: `builder/`, `errors/`, `const/`, `logger/`, `examples/`, `test/`
- Descriptive builder names: `search.go`, `document.go`, `bulk.go`, `aggregation.go`
- Specialized builders with operation: `scroll.go`, `search_after.go`, `delete_by_query.go`, `update_by_query.go`
- Tests named after package under test: `builder_test` in `test/` and `examples/` directories
- Constants files: `field_types.go`, `analyzer_constants.go`
- Generated files: `gen_json_methods.go` (build tool), `response_ext_gen.go` (generated output)

**Directories:**
- All lowercase, plural form for packages: `builder`, `errors`, `const`, `logger`, `examples`, `test`, `sugar`, `docs`
- No abbreviations: `client` not `cli`, `config` not `cfg`

**Struct and Interface Names:**
- Builders: `{Operation}Builder` (SearchBuilder, DocumentBuilder, IndexBuilder, BulkBuilder, AggregationBuilder, etc.)
- Responses: `{Operation}Response` (SearchResponse, DocumentResponse, BulkResponse, AggregationResponse, GetResponse, etc.)
- Config: `Config` struct, `Option` type for functional options
- Interfaces: `ESClient`, `IndexName`, `Logger`, `JSONSerializer`
- Helper types: `debugHelper`, `baseBuilder`, `BoolQuery[T]`, `NestedObject`

**Method Names:**
- Builder methods chainable: Match(), Term(), Range(), Set(), ID(), Size(), From(), Sort(), Debug()
- Action methods: Do(), Get(), Update(), Upsert(), Delete(), Create(), Exists(), Count()
- Configuration: Shards(), Replicas(), RefreshInterval(), Refresh()
- Query methods inherited from BoolQuery: MatchShould(), TermShould(), MatchMustNot(), TermMustNot(), MatchPhrase()
- Helpers: buildBoolQuery(), buildPath(), structToMap()

## Where to Add New Code

**New Search Feature:**
- Primary code: `builder/search.go` (add method to SearchBuilder, append to must/filters/should/mustNot lists)
- Use BoolQuery[SearchBuilder] methods where applicable
- Query building: modify buildBoolQuery() if new query type needed (currently in search.go)
- Tests: `test/search_test.go` or new test file if significant feature

**New Document Operation:**
- Primary code: `builder/document.go` (add method to DocumentBuilder)
- Path building: use buildPath() pattern with query parameters
- HTTP execution: call client.Do() or client.DoWithHeader() in action method
- Tests: `test/document_test.go`

**New Index Management Feature:**
- Primary code: `builder/index.go` (add method to IndexBuilder)
- Settings/Mappings: accumulate in index.settings or index.mappings maps
- Query building: construct ES request structure and call client.Do()
- Tests: `test/index_test.go`

**New Bulk Operation:**
- Primary code: `builder/bulk.go` (add operation type or builder method)
- NDJSON formatting: update Build() method to handle new operation format
- Tests: `test/bulk_test.go`

**New Builder Type (e.g., for new ES API):**
1. Create file: `builder/{operation}.go`
2. Define struct with embedded ESClient, debugHelper, baseBuilder, and operation-specific fields
3. Implement NewXyzBuilder(client ESClient, ...) constructor
4. Add method chain methods
5. Implement Do(ctx context.Context) (*ResponseType, error) for execution
6. Add response struct in same file or separate response.go file
7. Register tests in `test/{operation}_test.go`

**New Constant:**
- Field types: add to `const/field_types.go` as `const FieldTypeXyz = "xyz"`
- Analyzers: add to `const/analyzer_constants.go` as `const AnalyzerXyz = "xyz"`

**Shared Utilities:**
- Helpers used across builders: add to `builder/interface.go` or create `builder/util.go`
- Current shared: structToMap(), toSnakeCase(), jsonMarshal/unmarshal helpers

**Logger/Config Extensions:**
- New logger method: add to `logger/logger.go` interface and default implementation
- New config option: add func WithXyz() Option to `config/config.go`

## Special Directories

**builder/**
- Purpose: Main business logic
- Generated: response_ext_gen.go (auto-generated JSON methods for response types, see gen_json_methods.go)
- Committed: All files including generated response_ext_gen.go
- Note: gen_json_methods.go is build tool (marked with `//go:build ignore`), not compiled into binary

**test/**
- Purpose: Integration tests
- Generated: None
- Committed: All test files
- Note: Tests require live Elasticsearch instance; not unit tests; skip with `go test -short` if no ES available

**examples/**
- Purpose: Runnable examples
- Generated: None
- Committed: All example files
- Note: Can be executed as tests with `go test ./examples -run TestCompleteAPI`

**docs/**
- Purpose: User documentation
- Generated: None
- Committed: All markdown files
- Note: Referenced from README.md; not API docs (no godoc generated)

---

## Key File Dependencies and Layers

```
Application Code
    ↓
sugar/ (optional GORM-style layer)
    ↓
builder/ (fluent API)
    ├─→ client/ (HTTP transport)
    ├─→ config/ (configuration)
    ├─→ errors/ (error handling)
    ├─→ logger/ (logging)
    └─→ const/ (constants)
```

**Files NOT to Modify Without Review:**
- `go.mod`: Language version and dependency declarations
- `response_ext_gen.go`: Auto-generated; regenerate via `go run builder/gen_json_methods.go`
- `gen_json_methods.go`: Build tool; only modify if adding new response types

---

*Structure analysis: 2026-03-13*
