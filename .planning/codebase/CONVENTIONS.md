# Coding Conventions

**Analysis Date:** 2026-03-13

## Naming Patterns

**Files:**
- Snake case for Go source files: `search.go`, `bulk.go`, `document.go`
- Test files use `_test.go` suffix: `search_test.go`, `document_test.go`
- Generated files use `_gen.go` suffix: `response_ext_gen.go`, `gen_json_methods.go`

**Packages:**
- Lowercase, single-word packages: `builder`, `client`, `config`, `logger`, `errors`, `const`
- Test files use package `builder_test` to avoid circular dependencies

**Functions:**
- PascalCase for exported functions (receiver methods): `NewSearchBuilder`, `Match`, `Do`, `Debug`
- camelCase for unexported functions: `initBoolQuery`, `buildBoolQuery`, `structToMap`, `getAnalysisChild`
- Method receivers use single or two letter abbreviations: `func (b *SearchBuilder)`, `func (c *Client)`
- Constructor functions follow `New*` pattern: `NewSearchBuilder`, `NewDocumentBuilder`, `NewIndexBuilder`

**Variables & Constants:**
- camelCase for local variables: `indexName`, `esClient`, `statusCode`
- UPPERCASE for constants: `timeout`, constants defined in `const/` package use PascalCase like standard Go pattern
- Short receiver variables: `b` for builder types, `c` for client, `q` for query, `d` for debug helper, `t` for testing

**Types:**
- PascalCase for struct types: `SearchBuilder`, `DocumentBuilder`, `BulkBuilder`, `ESError`
- Interface types PascalCase: `ESClient`, `IndexName`, `Logger`
- Unexported struct types (package-private): `baseBuilder`, `debugHelper`, `bulkOperation`

**Type Aliases for Options:**
- AnalyzerOption pattern: `type AnalyzerOption func(map[string]any)` at `builder/index.go:106`
- These are used for variadic functional options to configure complex objects

## Code Style

**Formatting:**
- Standard Go formatting (gofmt compatible)
- Line length: No hard limit enforced, but files range from 24 to 1042 lines
- Indentation: Tabs (Go standard)
- No formatter config files present (relies on gofmt defaults)

**Comments:**
- Package-level documentation comments above `package` keyword
- Type comments for exported structs and interfaces: `// SearchBuilder 搜索构建器`
- Method comments for exported methods: `// Debug 启用调试模式（链式调用）`
- Chinese comments throughout (Chinese language documentation is standard in this codebase)
- Comments often include usage context and examples inline
- No JSDoc/TSDoc style comments, using Go comment conventions

**Comment Style:**
```go
// baseBuilder 基础构建器，提供通用的链式调用方法
type baseBuilder struct {
	headers http.Header
}

// Header 为单个请求设置自定义 HTTP Header
func (b *baseBuilder) Header(key, value string) {
	...
}
```

**Error Messages:**
- Chinese error messages: `"初始化默认日志失败"`, `"Error converting model to map"`
- Mix of Chinese and English error messages found
- Format: Use `fmt.Errorf` with `%w` for error wrapping: `fmt.Errorf("Error converting model to map: %w", err)`
- Prefix error messages with context: `"IndexBuilder.Shards: value must be >= 1"`

## Import Organization

**Order:**
1. Standard library imports: `context`, `encoding/json`, `fmt`, `net/http`, `net/url`, `reflect`, `time`
2. Internal imports (go-es packages): `github.com/Kirby980/go-es/builder`, `github.com/Kirby980/go-es/client`, etc.
3. Third-party imports: `go.uber.org/zap`, `go.uber.org/multierr`

**Example (from `builder/search.go` lines 3-10):**
```go
import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
)
```

**Path Aliases:**
- Standard import alias `stderrors "errors"` used in `builder/index.go:6` and `builder/document.go:6` to avoid name collision with custom `github.com/Kirby980/go-es/errors` package

**Import Grouping:**
- No blank lines between import groups (single group)
- All imports in alphabetical order within group

## Error Handling

**Pattern - Immediate Return:**
```go
// From client/client.go:42
if err != nil {
	return nil, fmt.Errorf("初始化默认日志失败: %w", err)
}
```

**Pattern - Delayed Error Check:**
Builders accumulate errors and check at execution time:
```go
// From builder/document.go:52-70
func (b *DocumentBuilder) Model(model any) *DocumentBuilder {
	...
	m, err := structToMap(model)
	if err != nil {
		b.err = fmt.Errorf("Error converting model to map: %w", err)
		return b
	}
	...
	return b
}

// Error checked at Do time
func (b *DocumentBuilder) Do(ctx context.Context) (*DocumentResponse, error) {
	if b.err != nil {
		return nil, b.err
	}
	...
}
```

**Custom Error Type:**
```go
// errors/errors.go:8-38
type ESError struct {
	StatusCode int
	Type       string
	Reason     string
	RootCause  []map[string]any
	RawBody    []byte
}

// Helper methods
func (e *ESError) IsNotFound() bool { ... }
func (e *ESError) IsConflict() bool { ... }
func (e *ESError) IsTimeout() bool { ... }
```

**Error Assertions:**
Never use panic or fatal in library code. Builders validate input and log errors via Logger interface.
```go
// From builder/index.go:50-54
func (b *IndexBuilder) Shards(shards int) *IndexBuilder {
	if shards < 1 {
		b.debugHelper.logger.Error("IndexBuilder.Shards: value must be >= 1")
	}
	b.settings["number_of_shards"] = shards
	return b
}
```

## Logging

**Framework:** Uber's `go.uber.org/zap` with custom Logger interface wrapper

**Logger Interface (logger/logger.go:7-12):**
```go
type Logger interface {
	Debug(msg string, keysAndValues ...any)
	Info(msg string, keysAndValues ...any)
	Warn(msg string, keysAndValues ...any)
	Error(msg string, keysAndValues ...any)
}
```

**Implementations:**
- `ZapLogger` - wraps zap.SugaredLogger (default)
- `NopLogger` - silent no-op implementation for disabling logs
- Custom loggers can be injected via `config.WithLogger()`

**Usage Pattern:**
```go
// From builder/debug.go:31-35
func (d *debugHelper) printDebug(method, path string, body any) {
	log := d.log()
	if body != nil {
		log.Info("[ES Debug] 请求", "method", method, "path", path, "body", body)
	}
}
```

**When to Log:**
- Debug: Request/response details in debug mode
- Info: Major operations like index creation, bulk flush
- Warn: Validation warnings, non-fatal issues
- Error: Failed operations, invalid configurations

**Debug Mode:**
- Each builder has optional `.Debug()` method that enables logging for that specific request
- Debug output includes request method, path, and body

## Module Design

**Builder Pattern:**
Every major operation follows the builder pattern:
- Struct type: `*SearchBuilder`, `*DocumentBuilder`, `*IndexBuilder`, `*BulkBuilder`, `*AggregationBuilder`, etc.
- Constructor: `NewSearchBuilder(c ESClient, index string) *SearchBuilder`
- Configuration methods: Return receiver type for chaining: `func (b *SearchBuilder) Match(...) *SearchBuilder`
- Execution: `.Do(ctx context.Context)` or `.Build()` method
- Optional debug: `.Debug() *SearchBuilder` method

**Example (builder/search.go:29-41):**
```go
func NewSearchBuilder(c ESClient, index string) *SearchBuilder {
	b := &SearchBuilder{
		client:      c,
		index:       index,
		size:        10,
		aggs:        make(map[string]any),
		debugHelper: debugHelper{logger: c.GetLogger()},
	}
	b.initBoolQuery(b)
	b.initBaseBuilder()
	return b
}
```

**Embedded Traits:**
Builders embed smaller structs to share functionality:
- `debugHelper` - provides Debug() and logging capability
- `baseBuilder` - provides Header() method for custom HTTP headers
- `BoolQuery[T any]` - generic embedded struct providing query building methods

**Struct Embedding Example (builder/search.go:13-27):**
```go
type SearchBuilder struct {
	client    ESClient
	index     string
	query     map[string]any
	...
	BoolQuery[SearchBuilder]  // Embedded: provides Must, Should, Filter, MustNot
	debugHelper               // Embedded: provides Debug, logging
	baseBuilder              // Embedded: provides Header
}
```

**Interface-based Client Dependency:**
Builders depend on `ESClient` interface, not concrete `*client.Client`:
- Avoids circular imports (`builder` ↔ `client` would be circular)
- Makes mocking in tests easier
- Defined in `builder/interface.go:15-26`

## Data Structure Patterns

**Map-based Query Building:**
Elasticsearch queries are built as `map[string]any` internally, then serialized to JSON:
```go
// From builder/query.go:25-32
func (q *BoolQuery[T]) Match(field string, value any) *T {
	q.must = append(q.must, map[string]any{
		"match": map[string]any{
			field: value,
		},
	})
	return q.self
}
```

**Struct-to-Map Conversion:**
Documents and models converted via JSON marshal/unmarshal (deferred optimization noted):
```go
// From builder/interface.go:35-54
func structToMap(v any) (map[string]any, error) {
	...
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	err = json.Unmarshal(data, &m)
	return m, err
}
```
- Comment at line 34: "目前采用 json marshal -> unmarshal，后续可优化为反射以提升性能"
- This is noted as a future optimization opportunity

**JSON Tags:**
- Custom struct field tags use `es` tag: `type:text;analyzer:ik_smart;format:yyyy-MM-dd HH:mm:ss`
- Example from README: `Price float64 \`es:"type:float"\``

## Function Design

**Builder Methods - Chain Return Pattern:**
All builder configuration methods return the receiver type to enable chaining:
```go
// Returns *SearchBuilder for chaining
func (b *SearchBuilder) Wildcard(field string, value string) *SearchBuilder {
	b.must = append(b.must, map[string]any{...})
	return b
}
```

**Execution Methods:**
Final execution methods in builders:
- `.Do(ctx context.Context) (*Response, error)` - execute and get typed response
- `.Build() map[string]any` - build query without executing
- `.Debug() *Builder` - enable debug logging (returns builder for chaining)

**Generic Function Receivers:**
`BoolQuery[T any]` uses generics to provide shared query methods to multiple builder types:
```go
// From builder/query.go:3-10
type BoolQuery[T any] struct {
	self               *T  // Reference to the embedding type
	filters            []map[string]any
	must               []map[string]any
	...
}

// Methods work on any builder type that embeds BoolQuery
func (q *BoolQuery[T]) Match(field string, value any) *T {
	...
	return q.self  // Returns the embedding type
}
```

**Variadic Options Pattern:**
Used in index builder and elsewhere:
```go
// From builder/index.go:87-102
func (b *IndexBuilder) AddCustomAnalyzer(name string, tokenizer string, filters ...string) *IndexBuilder {
	...
	if len(filters) > 0 {
		analyzer["filter"] = filters
	}
	return b
}
```

---

*Convention analysis: 2026-03-13*
