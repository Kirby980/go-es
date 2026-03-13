# Roadmap: go-es Quality Fixes & Feature Enhancements

## Overview

This roadmap transforms go-es from a functional but fragile Elasticsearch client into a production-hardened library. The work follows a strict dependency chain: correctness bugs first (so retry and failover have a solid foundation), then resilience infrastructure (node pool, circuit breaker, middleware), then performance and remaining features (cache, aliases). Every phase delivers a verifiable capability. Tests are co-located with the features they cover.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [ ] **Phase 1: Correctness, Security & Quality Fixes** - Fix all bugs that affect data integrity, security, and error handling
- [ ] **Phase 2: Node Pool & Retry Enhancement** - Replace naive round-robin with resilient node pool and exponential backoff retry
- [ ] **Phase 3: Circuit Breaker** - Add fail-fast protection when the Elasticsearch cluster is unreachable
- [ ] **Phase 4: Middleware** - Enable user-extensible request/response pipeline
- [ ] **Phase 5: Performance Optimization** - Eliminate double serialization and optimize hot paths
- [ ] **Phase 6: Cache & Index Aliases** - Add query caching and index alias management

## Phase Details

### Phase 1: Correctness, Security & Quality Fixes
**Goal**: The client correctly retries requests, escapes URLs, propagates errors, and handles edge cases -- no silent data loss or security holes
**Depends on**: Nothing (first phase)
**Requirements**: RELY-01, RELY-02, RELY-03, RELY-04, RELY-05, RELY-06, SECU-01, QUAL-01, QUAL-02, TEST-02, TEST-03, TEST-04
**Success Criteria** (what must be TRUE):
  1. A retried request sends the full original body (not an empty body) on every attempt
  2. Ping() returns success if any configured node is reachable, not just the first one
  3. Exists() returns an error (not false) when the request fails for reasons other than 404
  4. Index names and document IDs containing special characters (/, ?, #, spaces, non-ASCII) produce correctly escaped URLs
  5. errors.As() is used for all error type checks, and wrapped errors are correctly unwrapped throughout the codebase
**Plans:** 2/3 plans executed

Plans:
- [ ] 01-01-PLAN.md -- Fix confirmed bugs: DoRequest body reset, Ping all-addresses, BulkBuilder context/error
- [ ] 01-02-PLAN.md -- Verify already-fixed behaviors with tests + update docs to errors.As()
- [ ] 01-03-PLAN.md -- Comprehensive edge-case tests: error paths, URL encoding boundaries, AutoMigrate edges

### Phase 2: Node Pool & Retry Enhancement
**Goal**: The client automatically routes around failed nodes and retries with exponential backoff, preventing thundering herd and single-node dependency
**Depends on**: Phase 1 (body fix must be in place before retry improvements)
**Requirements**: POOL-01, POOL-02, POOL-03, POOL-04, TEST-01, TEST-05
**Success Criteria** (what must be TRUE):
  1. When a node fails, subsequent requests are routed to other live nodes without manual intervention
  2. A dead node is automatically re-tested after an exponentially increasing delay and restored to the live pool if healthy
  3. When all nodes are dead, the client force-resurrects the longest-dead node instead of returning an error immediately
  4. Retry delays follow exponential backoff with full jitter, and HTTP 429/502/503/504 responses trigger automatic retry
**Plans**: TBD

Plans:
- [ ] 02-01: TBD
- [ ] 02-02: TBD
- [ ] 02-03: TBD

### Phase 3: Circuit Breaker
**Goal**: The client fails fast when the Elasticsearch cluster is unreachable, preventing cascading failures in the calling application
**Depends on**: Phase 2 (circuit breaker responsibilities are defined relative to node pool)
**Requirements**: BRKR-01, BRKR-02, BRKR-03, TEST-06
**Success Criteria** (what must be TRUE):
  1. After a configurable number of consecutive failures, the circuit opens and subsequent requests fail immediately without hitting the network
  2. After a configurable timeout, the circuit enters half-open state and allows a probe request to test recovery
  3. Context cancellation errors and 4xx client errors do not count toward the circuit breaker failure threshold
  4. Circuit breaker is configured via functional options and disabled by default (backward compatible)
**Plans**: TBD

Plans:
- [ ] 03-01: TBD
- [ ] 03-02: TBD

### Phase 4: Middleware
**Goal**: Users can inject custom logic (logging, metrics, tracing) into the request/response pipeline without modifying the client
**Depends on**: Phase 3 (transport layer should be stable before wrapping it)
**Requirements**: MIDW-01, MIDW-02, MIDW-03, TEST-07
**Success Criteria** (what must be TRUE):
  1. A user-defined middleware function can inspect and modify requests before they are sent and responses after they are received
  2. Multiple middleware functions execute in a defined, documented order (outermost wraps innermost)
  3. Middleware is configured via functional options and the ESClient interface remains unchanged
**Plans**: TBD

Plans:
- [ ] 04-01: TBD
- [ ] 04-02: TBD

### Phase 5: Performance Optimization
**Goal**: Hot paths (scan, model conversion, bulk build) run without redundant serialization or allocation overhead
**Depends on**: Phase 1 (quality fixes to structToMap and AutoMigrate must be in place)
**Requirements**: PERF-01, PERF-02, PERF-03, PERF-04
**Success Criteria** (what must be TRUE):
  1. SearchResponse.Scan() converts results to target structs without an intermediate JSON marshal/unmarshal cycle
  2. DocumentBuilder.Model() converts a struct to a map without going through JSON as an intermediate format
  3. BulkBuilder.Build() pre-allocates its output buffer proportional to the number of operations, reducing GC pressure
**Plans**: TBD

Plans:
- [ ] 05-01: TBD
- [ ] 05-02: TBD

### Phase 6: Cache & Index Aliases
**Goal**: Repeated read queries are served from an in-memory cache, and index aliases can be managed through the builder API
**Depends on**: Phase 4 (cache benefits from middleware pattern being in place)
**Requirements**: CACH-01, CACH-02, CACH-03, ALIA-01, ALIA-02, TEST-08
**Success Criteria** (what must be TRUE):
  1. A repeated search query within the TTL window returns the cached result without hitting Elasticsearch
  2. Write operations (index, delete, bulk) do not return cached results and the cache respects TTL-based staleness
  3. Cache is configured via functional options (on/off, TTL, max entries) and disabled by default
  4. AutoMigrate can create and manage index aliases, and IndexBuilder provides Add/Remove/Switch alias operations
**Plans**: TBD

Plans:
- [ ] 06-01: TBD
- [ ] 06-02: TBD
- [ ] 06-03: TBD

## Progress

**Execution Order:**
Phases execute in numeric order: 1 -> 2 -> 3 -> 4 -> 5 -> 6

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Correctness, Security & Quality Fixes | 2/3 | In Progress|  |
| 2. Node Pool & Retry Enhancement | 0/3 | Not started | - |
| 3. Circuit Breaker | 0/2 | Not started | - |
| 4. Middleware | 0/2 | Not started | - |
| 5. Performance Optimization | 0/2 | Not started | - |
| 6. Cache & Index Aliases | 0/3 | Not started | - |
