# Codebase Concerns

**Analysis Date:** 2026-03-13

## Tech Debt

### Request Body Exhaustion on Retry
- Issue: In `client/client.go` (lines 134-171 in `Do()` method and lines 76-86 in `DoRequest()`), the request body from `bytes.NewReader()` is consumed on first HTTP request. Subsequent retry attempts send empty request body because the Reader position is at EOF.
- Files: `client/client.go`
- Impact: Any retry logic fails silently - subsequent retries transmit no body data to Elasticsearch, causing silent failures or incorrect behavior. This is critical for reliability.
- Fix approach: Reset Reader position using `io.Seeker` before retry, or recreate request for each attempt. The code already has the fix pattern in `DoWithHeader()` at lines 205-207 showing `seeker.Seek(0, io.SeekStart)` - apply same pattern to both methods.

### Error Swallowing in BulkBuilder Set Methods
- Issue: `BulkBuilder.Set()`, `SetFromStruct()`, `SetObject()`, `SetArray()`, `SetObjectArray()` methods in `builder/bulk.go` (lines 313-383) return error formatted strings via `fmt.Errorf()` but store them in `b.err` field without preventing further operations. Caller may not notice the error until `Do()` is called.
- Files: `builder/bulk.go`
- Impact: Silent failures in bulk operation construction. Complex operations can partially build before error, making debugging difficult. Production code can unknowingly send incomplete/invalid bulk requests.
- Fix approach: Consistent error handling - store errors to `b.err` field as already done, but document this behavior clearly and ensure callers check `Do()` error return. Consider returning builder instance itself with embedded error.

### Exists() Masks Real Errors
- Issue: `IndexBuilder.Exists()` in `builder/index.go:545-552` and `DocumentBuilder.Exists()` in `builder/document.go:392-403` treat all errors as "resource not found" by returning `false, nil`. Network failures, authentication errors, timeouts, and actual 404 errors are indistinguishable.
- Files: `builder/index.go`, `builder/document.go`
- Impact: Network/auth failures silently treated as "resource doesn't exist", leading to incorrect logic. Application may create duplicate resources or fail operations silently. Difficult to diagnose connectivity issues.
- Fix approach: Use error type assertion to check if error is `*errors.ESError` with `IsNotFound()` true. Return actual error for non-404 cases.

### Snake Case Conversion Breaks on Acronyms
- Issue: `toSnakeCase()` function in `builder/index.go:312-321` doesn't handle consecutive uppercase letters correctly. "HTTPServer" converts to "h_t_t_p_server" instead of "http_server".
- Files: `builder/index.go`
- Impact: AutoMigrate generates incorrect index names for structs with acronyms. Index names won't match expected patterns, causing document routing failures.
- Fix approach: Implement smarter logic checking if previous char is lowercase OR next char is lowercase to determine where underscores belong.

### AutoMigrate Doesn't Handle Slice[Struct] with nested Type
- Issue: `AutoMigrate()` in `builder/index.go:279-289` reflects struct fields but doesn't handle case where field type is `[]SomeStruct` marked as `es:"type:nested"`. Code checks `nestedType.Kind() == reflect.Struct` but for Slice fields, `Elem()` returns the Slice element type which might be a Ptr or Struct.
- Files: `builder/index.go`
- Impact: Nested field mappings fail when using slice types. Index creation with nested arrays fails or uses incorrect field types. AutoMigrate feature breaks for common patterns.
- Fix approach: Add explicit Slice handling - if `nestedType.Kind() == reflect.Slice`, call `nestedType = nestedType.Elem()` before Struct check.

### BulkBuilder AutoFlush Uses context.TODO()
- Issue: `BulkBuilder.isFlush()` method in `builder/bulk.go:103-110` uses `context.TODO()` internally when auto-flushing, ignoring any context passed by user. Errors from flush are silently dropped.
- Files: `builder/bulk.go`
- Impact: No way to control timeout for auto-flush operations. Errors from flush completely ignored - bulk insert can fail without caller knowing. Timeout/cancellation signals from parent context ignored.
- Fix approach: Thread context parameter through `Add()` and other methods to `isFlush()`. Capture flush errors to `b.err` field for return in `Do()`.

---

## Known Bugs

### Ping() Ignores Configured Addresses
- Symptoms: When multiple Elasticsearch addresses configured, `Ping()` always uses `addresses[0]`. If that node is down but others are up, Ping fails even though cluster is available.
- Files: `client/client.go:136-157`
- Trigger: Create client with multiple addresses like `["http://dead-node:9200", "http://healthy-node:9200"]`, call `Ping()` when first node is down.
- Workaround: Implement manual retry loop against all addresses in application code.

---

## Security Considerations

### Unescaped URL Path Components
- Risk: Index names, document IDs, and other path components are directly interpolated into URLs without URL encoding. Special characters like `/`, `#`, `?`, spaces will break URL structure or create unintended paths.
- Files: `builder/document.go`, `builder/index.go`, `builder/search.go`, `builder/scroll.go`, `builder/search_after.go` (numerous path constructions)
- Current mitigation: None - code concatenates strings directly to URLs.
- Recommendations: Use `url.PathEscape()` consistently on all user-provided path components before interpolating into URL strings. Example: `fmt.Sprintf("/%s/_doc/%s", url.PathEscape(b.index), url.PathEscape(b.id))`

### Request Body Not Reset on Retry
- Risk: When retry logic encounters error, subsequent requests send empty/truncated bodies. If request contains sensitive data, first attempt captures it; retry with same sensitive data might expose it to wrong endpoint if retry URL is different.
- Files: `client/client.go`
- Current mitigation: Code has no special handling for sensitive data.
- Recommendations: Ensure body reset/recreation on every retry attempt to prevent duplicate exposure of request bodies.

---

## Performance Bottlenecks

### Double Serialization in SearchResponse.Scan()
- Problem: Method converts each search hit to JSON then back to user struct - `json.Marshal(hit.Source)` then `json.Unmarshal(jsonData, elem)`. With thousands of results, this is expensive overhead.
- Files: `builder/response_ext_gen.go:41-86` (Scan method implementation)
- Cause: Current approach marshals each individual hit.Source map separately instead of batch processing. `hit.Source` is already `map[string]any` type.
- Improvement path: Use library like `mapstructure` for direct map-to-struct conversion, or collect all sources and unmarshal in single operation.

### BulkBuilder.Build() Causes Multiple Buffer Allocations
- Problem: Buffer starts with zero capacity and grows incrementally as JSON is written. Large bulk operations cause multiple allocations and copies.
- Files: `builder/bulk.go:458-490`
- Cause: No pre-allocation hint for bytes.Buffer.
- Improvement path: Call `buf.Grow(len(b.operations) * 200)` at start to pre-allocate based on estimated 200 bytes per operation. Buffer already does this optimally at line 462 - verify it's sized correctly for operations count.

### DocumentBuilder.Model() Double Serialization
- Problem: Converting struct to map via `json.Marshal()` then `json.Unmarshal()` is unnecessary indirection.
- Files: `builder/document.go:49-73`
- Cause: Converting struct→JSON→map instead of directly providing struct to request.
- Improvement path: If request body can accept struct directly, skip the map conversion. Otherwise, consider reflection-based conversion or use third-party struct-to-map library.

### structToMap() Implementation
- Problem: Function in `builder/interface.go:35-54` always uses JSON marshaling round-trip to convert structs to maps. This is convenient but slower than direct reflection.
- Files: `builder/interface.go`
- Cause: JSON marshaling is simplest but has serialization overhead.
- Improvement path: Implement direct reflection-based struct traversal, only falling back to JSON for complex nested types.

---

## Fragile Areas

### Bulk Operation State Machine
- Files: `builder/bulk.go` (entire file)
- Why fragile: Complex state maintained across `currentOp`, `operations` slice, and implicit state from method call order. Calling `Set()` without prior `AddDoc()`/`CreateDoc()`/`UpdateDoc()` produces confusing error. The `commitCurrent()` pattern is called in 8+ different places.
- Safe modification: When adding new Set* methods, must remember to add currentOp nil check. Must call commitCurrent() at appropriate times. Document the state machine clearly.
- Test coverage: Gaps in error path testing - what happens if user mixes API styles (Add() then Set()), what happens with empty bulk requests, concurrent modification scenarios not tested.

### Builder Embedded Type Composition
- Files: `builder/search.go`, `builder/scroll.go`, `builder/search_after.go`, `builder/update_by_query.go`, `builder/delete_by_query.go`
- Why fragile: Multiple builders embed `BoolQuery[T]` via `BoolQuery[SearchBuilder]` etc. This creates implicit dependencies - method receivers call `b.buildBoolQuery()` which must exist on each builder type. If BoolQuery interface changes, must update all embedders.
- Safe modification: Before changing BoolQuery interface, grep for all builder types that embed it. Must ensure all have matching buildBoolQuery method signatures.
- Test coverage: BoolQuery logic tested through individual builder tests but cross-builder consistency not explicitly verified.

### Error Type Assertions
- Files: Multiple builder files using `errors.ESError`
- Why fragile: Code assumes error is `*errors.ESError` from `client.Do()`, but `Do()` wraps errors - some codepaths might return wrapped errors that fail type assertion.
- Safe modification: Use `errors.As()` for proper error chain unwrapping instead of direct type assertion.
- Test coverage: Error handling tested only for happy path in most tests - error scenarios with various error types not covered.

### Index Settings as map[string]any
- Files: `builder/index.go`
- Why fragile: Settings stored as nested maps with no validation. Typos in setting names or wrong value types only caught by Elasticsearch API, not during build.
- Safe modification: Consider helper methods like `.Shards()`, `.Replicas()` already exist - use them instead of raw map manipulation.
- Test coverage: Settings validation relies on integration tests against real Elasticsearch - unit tests don't catch invalid settings.

---

## Scaling Limits

### Single HTTP Client Per Process
- Current capacity: One `*http.Client` shared across all goroutines. Default connection pool: `MaxIdleConns=100`, `MaxConnsPerHost=0` (unlimited).
- Limit: Under extreme load (millions of concurrent requests), single client's connection pool and internal state become bottleneck. Goroutine overhead for each request increases.
- Scaling path: For mega-scale (>10k RPS), consider connection pool tuning or multiple client instances. Current defaults should handle typical use cases (1k-10k RPS).

### Memory Growth with Scroll Operations
- Current capacity: `ScrollBuilder` maintains `scrollID` string in memory. Long-running scroll operations with millions of docs work but each scroll context uses ES server memory.
- Limit: Keeping many concurrent scrolls open starves ES cluster memory. Keep-alive TTL limits maximum operation time.
- Scaling path: Implement mechanism to limit concurrent scrolls, warn when scroll context becomes stale. Document best practice: clear scroll promptly with `Clear()`.

### Bulk Operation Size
- Current capacity: `BulkBuilder` keeps all operations in memory until `Do()`. For millions of operations, memory footprint is O(n).
- Limit: 1M operations = ~200MB+ memory if each doc is ~200 bytes. Can exceed available memory.
- Scaling path: Implement streaming mode where operations are flushed after reaching size threshold. AutoFlushSize feature partially addresses this but only when explicitly configured.

---

## Dependencies at Risk

### No External Dependencies (Minimal Risk)
- The project has zero external Go dependencies in main code. Only development/optional dependencies are `go.uber.org/zap` and `go.uber.org/multierr` for logging.
- Risk: Low - self-contained, no supply chain risk.
- Impact: None known.
- Migration plan: N/A.

### Embedded JSON Unmarshaling
- Risk: Project relies on stdlib `encoding/json` which is stable but doesn't handle nested ES response structures elegantly. No schema validation on responses.
- Impact: If ES API changes response format, code silently ignores new fields. Missing fields in responses not caught.
- Migration plan: Could add `serde` validation layer but not critical for current maturity.

---

## Test Coverage Gaps

### Retry Logic Not Tested
- What's not tested: The request body reset behavior during retries. Manual test of actual retry scenario with network errors.
- Files: `client/client.go` (retry logic), `test/` (no retry simulation tests)
- Risk: Retry logic bug (request body exhaustion) went undetected because tests don't simulate actual retries.
- Priority: High - this is critical production path.

### Error Path Testing Incomplete
- What's not tested: How builders behave when given invalid inputs (nil values, empty strings, special characters in IDs/index names). What happens when `Do()` is called after error in builder chain.
- Files: All builder files, test files mostly test happy path.
- Risk: Users hitting edge cases get confusing errors or silent failures.
- Priority: Medium - affects user experience on error cases.

### URL Encoding Edge Cases
- What's not tested: Index names with special characters `/`, `?`, `#`, spaces. Document IDs with these characters. Field names with non-ASCII characters.
- Files: All builders constructing paths, `test/` has no such tests.
- Risk: Users with non-ASCII index names or special-character IDs get failed requests.
- Priority: Medium - mainly affects non-English users.

### Concurrent Access to Builders
- What's not tested: Whether builders are actually safe to use concurrently (they shouldn't be, but tests don't verify this).
- Files: All builder implementations, test files don't attempt concurrent access.
- Risk: Users might try to share builder across goroutines. Race detector would catch this but no explicit tests prevent it.
- Priority: Low-Medium - documented as not thread-safe but could test to be sure.

### AutoMigrate Edge Cases
- What's not tested: Nested structs with multiple nesting levels, slices of nested structs, pointers to slices, anonymous embedded structs.
- Files: `builder/index.go:213-289`, `test/` (minimal AutoMigrate tests)
- Risk: Complex struct definitions fail during migration without clear error message.
- Priority: Medium - affects advanced users defining complex types.

---

## Missing Critical Features

### No Automatic Node Failover
- Problem: If primary Elasticsearch node becomes unavailable, application doesn't fail over. Must configure multiple addresses but `Ping()` only tries first address.
- Blocks: High-availability deployments. Customers must implement their own failover logic.
- Priority: High for production use.

### No Built-in Circuit Breaker
- Problem: If Elasticsearch is down, every request waits full timeout (default 30s). No exponential backoff for cluster unavailability.
- Blocks: Graceful degradation. Application becomes unresponsive when ES down.
- Priority: High for production use.

### Limited Request/Response Middleware Support
- Problem: No way to inject custom logic before request or after response (logging, metrics, tracing, authentication enrichment).
- Blocks: Observability. Production monitoring integration difficult.
- Priority: Medium-High.

### No Built-in Caching
- Problem: Every search query hits ES. No local caching layer for repeated queries.
- Blocks: Performance optimization for read-heavy workloads.
- Priority: Low-Medium (user can implement at application level).

### AutoMigrate Doesn't Support Index Aliases
- Problem: Can create index but not set up aliases for zero-downtime reindexing pattern.
- Blocks: Advanced deployment patterns.
- Priority: Low.

---

## Issues at Risk of Regression

### Debug Mode Auto-Off Behavior
- Issue: Each `Do()` call disables debug mode via `defer b.setDebug(false)`. Users must re-enable after each operation.
- Files: Multiple builders using pattern `defer b.setDebug(false)` after printing debug info
- Risk: If future code removes defer statement, debug mode becomes sticky. Current behavior unintuitive.
- Recommendation: Make debug mode persistent or provide separate one-shot debug methods.

### Close() Implementation Incomplete
- Issue: `Client.Close()` in `client/client.go:77-82` is actually implemented correctly (lines 78-80 show `transport.CloseIdleConnections()`). Previous concern in OPTIMIZATION.md was incorrect - this is already fixed.
- Files: `client/client.go`
- Status: Non-issue - close is properly implemented.

---

*Concerns audit: 2026-03-13*
