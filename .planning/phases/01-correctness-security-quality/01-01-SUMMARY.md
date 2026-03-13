---
phase: 01-correctness-security-quality
plan: 01
subsystem: client
tags: [retry, body-reset, ping, bulk, autoflush, context, tdd]

# Dependency graph
requires: []
provides:
  - "DoRequest body reset on retry via GetBody/Seeker"
  - "Ping all-addresses iteration with context cancellation"
  - "BulkBuilder AutoFlushContext() method for user-provided context"
  - "BulkBuilder auto-flush error capture propagated through Do()"
affects: [02-resilience-observability, 03-node-pool-circuit-breaker]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "GetBody-first body reset for HTTP retry (DoRequest + DoWithHeader)"
    - "Range-loop health check over all addresses with early-exit on success"
    - "Auto-flush context threading via stored b.ctx field"
    - "TDD RED-GREEN cycle for production bug fixes"

key-files:
  created:
    - "client/client_test.go"
    - "builder/bulk_test.go"
  modified:
    - "client/client.go"
    - "builder/bulk.go"

key-decisions:
  - "Used req.GetBody() for body reset instead of io.Seeker alone, because http.NewRequestWithContext wraps readers in NopCloser which strips Seeker interface"
  - "Upgraded DoWithHeader to use same GetBody-first pattern for consistency"
  - "AutoFlush fallback context is context.Background() not context.TODO() to signal intentional default"

patterns-established:
  - "TDD with httptest: use Hijack+close for reliable transport errors, DisableKeepAlives for clean test isolation"
  - "Mock ESClient in builder_test: extend existing mock_client_test.go with doReqFunc for context capture"

requirements-completed: [RELY-01, RELY-02, RELY-04, RELY-05]

# Metrics
duration: 32min
completed: 2026-03-13
---

# Phase 1 Plan 01: Critical Bug Fixes Summary

**Fixed DoRequest retry body reset via GetBody, Ping all-addresses iteration, and BulkBuilder AutoFlush context threading with error capture**

## Performance

- **Duration:** 32 min
- **Started:** 2026-03-13T09:31:36Z
- **Completed:** 2026-03-13T10:03:37Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments
- DoRequest retry now sends the full original body on every attempt (previously sent empty body after EOF)
- Ping iterates all configured addresses and returns success if any responds, handles empty list and context cancellation
- BulkBuilder AutoFlush uses user-provided context (via new AutoFlushContext() method) with context.Background() fallback
- BulkBuilder auto-flush errors now propagate through Do() instead of being silently dropped
- DoWithHeader body reset upgraded to use GetBody-first approach for consistency

## Task Commits

Each task was committed atomically:

1. **Task 1: Fix DoRequest body reset and Ping all-addresses (RED)** - `cbc149b` (test)
2. **Task 1: Fix DoRequest body reset and Ping all-addresses (GREEN)** - `83c5e93` (fix)
3. **Task 2: Fix BulkBuilder AutoFlush context and error propagation (RED)** - `2cbdade` (test)
4. **Task 2: Fix BulkBuilder AutoFlush context and error propagation (GREEN)** - `015ed5c` (fix)

_Note: TDD tasks have separate test and fix commits (RED-GREEN cycle)_

## Files Created/Modified
- `client/client_test.go` - 7 unit tests: DoRequest body reset (2), Ping all-addresses (5)
- `client/client.go` - Body reset via GetBody in DoRequest retry loop, Ping rewritten to iterate all addresses
- `builder/bulk_test.go` - 5 unit tests: AutoFlush context (1), AutoFlush error (1), no-context fallback (1), Set error (1), SetFromStruct error (1)
- `builder/bulk.go` - Added ctx field, AutoFlushContext() method, isFlush() context threading and error capture

## Decisions Made
- Used `req.GetBody()` for body reset instead of `io.Seeker` alone, because `http.NewRequestWithContext` wraps `*bytes.Reader` in `io.NopCloser` which strips the `io.Seeker` interface. GetBody is set automatically by Go for `*bytes.Reader` and `*bytes.Buffer`.
- Upgraded `DoWithHeader` to use the same GetBody-first pattern for consistency with the DoRequest fix.
- AutoFlush fallback context is `context.Background()` (not `context.TODO()`) to signal an intentional default rather than a placeholder.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] GetBody approach instead of Seeker-only for DoRequest body reset**
- **Found during:** Task 1 (GREEN phase)
- **Issue:** Plan specified Seeker-based body reset, but `http.NewRequestWithContext` wraps readers in `io.NopCloser`, stripping the `io.Seeker` interface. The Seeker approach silently fails.
- **Fix:** Used `req.GetBody()` as primary mechanism (set by Go for `*bytes.Reader`), with Seeker as fallback for custom readers that implement it directly.
- **Files modified:** client/client.go
- **Verification:** TestDoRequestBodyReset_RetryWithBody passes -- server receives full body on retry
- **Committed in:** 83c5e93

**2. [Rule 2 - Missing Critical] DoWithHeader GetBody upgrade**
- **Found during:** Task 1 (GREEN phase)
- **Issue:** DoWithHeader used Seeker-only approach which has the same NopCloser limitation. Since `bytes.NewReader` is used for the JSON body, GetBody is available.
- **Fix:** Added GetBody-first check to DoWithHeader's retry loop for consistency.
- **Files modified:** client/client.go
- **Verification:** Existing behavior preserved, same pattern as DoRequest
- **Committed in:** 83c5e93

---

**Total deviations:** 2 auto-fixed (1 bug fix refinement, 1 consistency improvement)
**Impact on plan:** Both fixes are strictly correct improvements to the planned approach. No scope creep.

## Issues Encountered
- `test/` package integration tests fail (pre-existing) because they require a live Elasticsearch instance. Not related to our changes.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- DoRequest body reset and Ping all-addresses are prerequisites for node pool and circuit breaker work in later phases
- BulkBuilder context threading enables proper timeout and cancellation propagation in bulk operations
- All fixes are backward-compatible (AutoFlushContext is opt-in, Ping behavior is strictly improved)

---
*Phase: 01-correctness-security-quality*
*Completed: 2026-03-13*

## Self-Check: PASSED

- All 4 source files exist (client/client.go, client/client_test.go, builder/bulk.go, builder/bulk_test.go)
- SUMMARY.md created at expected path
- All 4 commits verified (cbc149b, 83c5e93, 2cbdade, 015ed5c)
- Key patterns found: GetBody (4), range c.addresses (1), AutoFlushContext (2), auto-flush failed (1)
- Test file line counts: client_test.go=266 (min 80), bulk_test.go=174 (min 60)
- All 12 unit tests pass, go vet clean
