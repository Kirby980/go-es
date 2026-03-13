# Testing Patterns

**Analysis Date:** 2026-03-13

## Test Framework

**Runner:**
- Built-in `testing` package (Go standard)
- Go version 1.24.0 (specified in `go.mod`)
- No external test runner (pytest, vitest, jest, etc.)

**Test Invocation:**
```bash
go test ./...           # Run all tests
go test -v ./...        # Verbose output
go test -run TestName   # Run specific test
go test -count=1        # Disable test caching
```

## Test File Organization

**Location:**
- Tests are separated in dedicated `/data/workspace/go-es/test/` directory
- Not co-located with source code (separate from `builder/`, `client/`, `config/` packages)
- Rationale: Tests in package `builder_test` to avoid circular imports with main code

**Naming Convention:**
- Test files use `_test.go` suffix: `search_test.go`, `document_test.go`, `bulk_test.go`, `index_test.go`
- Generated test files follow pattern: test files with functional grouping by domain

**Test Files by Domain:**
| File | Size | Purpose |
|------|------|---------|
| `index_test.go` | 1042 lines | Index creation, management, aliases, analysis |
| `document_test.go` | 994 lines | Document CRUD, batching, nested objects |
| `bulk_test.go` | 854 lines | Bulk operations, auto-flush, callbacks |
| `search_test.go` | 619 lines | Search queries, filters, aggregations, pagination |
| `search_after_test.go` | 744 lines | Pagination with search_after |
| `scroll_test.go` | 544 lines | Scroll-based pagination |
| `search_new_test.go` | 480 lines | Advanced search features |
| `aggregation_test.go` | 333 lines | Metric aggregations, terms, histogram |
| `cluster_test.go` | 262 lines | Cluster health, state, settings |
| `gorm_style_test.go` | 261 lines | GORM-like API testing |
| `performance_test.go` | 160 lines | Performance benchmarks |
| `debug_test.go` | Short | Debug mode verification |

## Test Structure

**Package Declaration:**
```go
package builder_test

import (
	"context"
	"testing"
	"time"

	"github.com/Kirby980/go-es/builder"
	"github.com/Kirby980/go-es/client"
)
```

**Standard Test Function Signature:**
```go
func TestBuilderSearchBuilder_Match(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()
	...
}
```

**Naming Pattern:**
- `Test<ComponentName>_<FeatureName>` format
- Examples: `TestBuilderSearchBuilder_Match`, `TestDocumentBuilder_Create`, `TestIndexBuilder_CreateIndex`
- Reflects the builder type and specific functionality being tested

## Test Lifecycle

**Setup Phase:**

1. **Create Test Client** (from `test/index_test.go:48-63`)
```go
func createTestClient(t *testing.T) *client.Client {
	esClient, err := client.New(
		config.WithAddresses("https://localhost:9200"),
		config.WithAuth("elastic", "123456"),
		config.WithInsecureSkipVerify(true),
		config.WithTimeout(10*time.Second),
		config.WithMaxConnsPerHost(100),
		config.WithMaxIdleConns(200),
		config.WithMaxIdleConnsPerHost(50),
		config.WithIdleConnTimeout(90*time.Second),
	)
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}
	return esClient
}
```

2. **Prepare Test Index** (from `test/search_test.go:13-100`)
```go
func prepareSearchTestData(t *testing.T, esClient *client.Client, indexName string) {
	ctx := context.Background()

	// 删除并创建索引
	_ = builder.NewIndexBuilder(esClient, indexName).Delete(ctx)
	_ = builder.NewIndexBuilder(esClient, indexName).
		Shards(1).
		Replicas(0).
		AddProperty("title", "text", builder.WithAnalyzer("ik_smart")).
		AddProperty("content", "text", builder.WithAnalyzer("ik_smart")).
		...
		Do(ctx)

	time.Sleep(500 * time.Millisecond)  // Wait for index to be ready

	// 插入测试数据
	documents := []map[string]any{
		{"title": "iPhone 15 Pro Max", "price": 1299.99, ...},
		...
	}
	...
}
```

**Teardown Phase:**
```go
// Cleanup using defer
defer func() {
	_ = builder.NewIndexBuilder(client, indexName).Delete(ctx)
}()

// Alternative: inline cleanup
defer client.Close()
```

## Test Patterns

**Basic Test Structure (from `test/search_test.go:102-127`):**
```go
func TestBuilderSearchBuilder_Match(t *testing.T) {
	// 1. Setup
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_match"
	prepareSearchTestData(t, client, indexName)

	// 2. Execute
	resp, err := builder.NewSearchBuilder(client, indexName).
		MatchPhrase("content", "手机").
		Do(ctx)

	// 3. Assert & Log
	if err != nil {
		t.Fatalf("Match 查询失败: %v", err)
	}

	if resp.Hits.Total.Value == 0 {
		t.Error("应该找到匹配的文档")
	}

	t.Logf("✓ Match 查询成功: 找到 %d 条结果", resp.Hits.Total.Value)
}
```

**Error Handling Pattern:**
```go
// Fatal error (test stops)
if err != nil {
	t.Fatalf("创建客户端失败: %v", err)
}

// Non-fatal assertion
if resp.Hits.Total.Value == 0 {
	t.Error("应该找到匹配的文档")
}

// Log messages
t.Logf("✓ Match 查询成功: 找到 %d 条结果", resp.Hits.Total.Value)
```

**Builder Chain Pattern in Tests:**
Tests showcase real usage of builder methods:
```go
resp, err := builder.NewAggregationBuilder(client, indexName).
	Avg("avg_price", "price").
	Sum("total_views", "views").
	Min("min_price", "price").
	Max("max_price", "price").
	Stats("price_stats", "price").
	Cardinality("unique_categories", "category").
	Count("product_count", "title").
	Do(ctx)
```

## Test Data Fixtures

**Test Data Pattern (from `test/search_test.go:36-80`):**
```go
documents := []map[string]any{
	{
		"title":     "iPhone 15 Pro Max",
		"content":   "最新款苹果手机，性能强劲",
		"category":  "electronics",
		"tags":      []string{"phone", "apple", "5g"},
		"price":     1299.99,
		"views":     1000,
		"rating":    4.8,
		"published": true,
		"location":  map[string]float64{"lat": 37.7749, "lon": -122.4194},
	},
	...
}

// Insert via bulk
for _, doc := range documents {
	builder.NewDocumentBuilder(client, indexName).
		SetMap(doc).
		Do(ctx)
}
```

**Fixture Location:**
- Fixtures are created inline in test helper functions like `prepareSearchTestData`, `prepareTestIndex`
- Located in `test/` directory
- Reused across multiple tests by calling setup functions

**Fixture Strategy:**
- Create/recreate indices for each test (isolation)
- Insert sample data (products, documents, etc.) with realistic fields
- Use small sample sizes (5-10 documents) for quick execution
- Include different data types (text, numbers, dates, arrays, nested objects)

## Assertion Patterns

**Response Validation (from `test/aggregation_test.go:37-39`):**
```go
if resp.Aggregations == nil {
	t.Error("聚合结果不应该为空")
}
```

**Result Count Checking (from `test/search_test.go:122-124`):**
```go
if resp.Hits.Total.Value == 0 {
	t.Error("应该找到匹配的文档")
}

if resp.Hits.Total.Value < 2 {
	t.Error("应该找到至少 2 个文档")
}
```

**Debug Output (from `test/aggregation_test.go:42`):**
```go
t.Logf("聚合结果: %s", resp.PrettyJSON())
```

**Success Messages with emojis (from `test/search_test.go:126`):**
```go
t.Logf("✓ Match 查询成功: 找到 %d 条结果", resp.Hits.Total.Value)
```

## Coverage & Coverage Verification

**Coverage Approach:**
- No coverage requirements enforced
- No coverage configuration files (*.coverprofile, .codecov.yml, etc.)
- Tests serve as integration tests against live Elasticsearch instance

**What's Tested:**
- All major builder types: Search, Document, Index, Bulk, Aggregation, Cluster, Scroll, SearchAfter
- All CRUD operations: Create, Read, Update, Delete, Exists
- Complex features: Boolean queries, filtering, aggregations, pagination, nested objects
- Error cases: Not found (404), conflicts, timeouts

**Coverage Gaps (potential improvements):**
- No mock tests - all tests require live Elasticsearch server
- No unit tests of internal helper functions (structToMap, buildBoolQuery internals)
- Performance tests are separate from functional tests (`performance_test.go`)
- No stress/load testing

## Integration Testing

**Real Elasticsearch Dependency:**
All tests are **integration tests** - they require a live Elasticsearch instance:
- Server: `https://localhost:9200`
- Credentials: `elastic` / `123456`
- SSL verification disabled in tests
- Tests are NOT isolated from state (depend on actual ES instance)

**ES Version Compatibility:**
- Tests assume Elasticsearch 8.x+ (based on API patterns)
- No version-specific test branching found

**Data Isolation:**
- Each test creates its own index: `test_search_match`, `test_doc_index`, etc.
- Index cleanup via `DeleteIndex()` in defer blocks
- Waits 500ms between index creation and data insertion for ES to be ready

## Test Utilities

**Helper Functions (defined in test files):**

1. **`createTestClient(t *testing.T) *client.Client`** (index_test.go:48)
   - Creates and configures test ES client
   - Fails test if client creation fails

2. **`prepareSearchTestData(t *testing.T, esClient, indexName string)`** (search_test.go:13)
   - Creates index with properties
   - Inserts sample product/document data
   - Returns ready-to-test index

3. **`prepareTestIndex(t *testing.T, esClient, indexName string)`** (document_test.go:12)
   - Similar to above, used in document tests
   - Configures different field types based on test domain

4. **Context Setup:**
   - `ctx := context.Background()` - standard pattern in all tests
   - No test-specific timeout contexts observed

## Performance Benchmarks

**Location:** `test/performance_test.go` (160 lines)

**Approach:**
- Separate file for performance tests
- Likely uses Go's `testing.B` benchmark harness
- Co-located with functional tests but clearly separated

## Test Organization Recommendations

**Current Strengths:**
- Clear separation of test concerns by domain (search, document, index, bulk, aggregation)
- Reusable helper functions for setup
- Consistent structure across all test functions
- Good test data with realistic documents

**Areas Worth Optimizing:**

1. **Mock/Unit Testing**: No mocks for ESClient interface - all tests hit real Elasticsearch
   - Opportunity: Create mock ESClient implementation for unit tests
   - Would allow testing without ES dependency

2. **Test Data Factory**: Fixture data inline in helper functions
   - Opportunity: Create factory functions for different document types
   - Example: `factoryProduct()`, `factoryUser()` in dedicated file

3. **Test Helpers Duplication**: Some setup code repeated across test files
   - Opportunity: Move common setup to `test/setup.go` or similar
   - Current duplication: `createTestClient` copied to multiple files

4. **Performance Testing**: Benchmarks separated but could be expanded
   - Current: `performance_test.go` with 160 lines
   - Opportunity: Add memory allocation benchmarks, concurrent operation benchmarks

5. **Table-driven Tests**: No parameterized test patterns observed
   - Opportunity: Use table-driven tests for multi-variant cases
   - Example: Different query types could share one test function

6. **Error Case Coverage**: Most tests focus on happy path
   - Opportunity: Add dedicated error case tests (e.g., `TestDocumentBuilder_InvalidInput`)
   - Current: Only implicit error checking in various tests

7. **Struct-to-Map Conversion**: Internal helper function `structToMap` untested
   - Located in `builder/interface.go:35-54`
   - Comment notes this as optimization target: "目前采用 json marshal -> unmarshal，后续可优化为反射以提升性能"
   - No dedicated unit tests found

---

*Testing analysis: 2026-03-13*
