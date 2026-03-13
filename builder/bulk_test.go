package builder_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/Kirby980/go-es/builder"
)

// --- AutoFlush context tests ---

// TestBulkBuilder_AutoFlushContext verifies that auto-flush uses the
// user-provided context (via AutoFlushContext), not context.TODO().
func TestBulkBuilder_AutoFlushContext(t *testing.T) {
	type ctxKey string
	key := ctxKey("test-marker")

	var capturedCtx context.Context
	mock := &mockESClient{
		doReqFunc: func(ctx context.Context, req *http.Request) ([]byte, error) {
			capturedCtx = ctx
			// Return valid bulk response
			return []byte(`{"took":1,"errors":false,"items":[{"index":{"_index":"test","_id":"1","status":201}},{"index":{"_index":"test","_id":"2","status":201}}]}`), nil
		},
	}

	userCtx := context.WithValue(context.Background(), key, "present")

	b := builder.NewBulkBuilder(mock).
		Index("test").
		AutoFlushSize(2).
		AutoFlushContext(userCtx)

	// AutoFlush triggers when len(operations) >= autoFlushSize BEFORE adding
	// the next item. So we need 3 Add calls: after 2 items are stored, the
	// 3rd Add triggers isFlush() which flushes the first 2.
	b.Add("test", "1", map[string]any{"field": "value1"}).
		Add("test", "2", map[string]any{"field": "value2"}).
		Add("test", "3", map[string]any{"field": "value3"})

	// The auto-flush should have used our context
	if capturedCtx == nil {
		t.Fatal("expected DoRequest to be called during auto-flush, but it wasn't")
	}
	if capturedCtx.Value(key) != "present" {
		t.Error("auto-flush did not use the user-provided context")
	}
}

// TestBulkBuilder_AutoFlushError verifies that auto-flush errors are
// captured and surfaced through Do().
func TestBulkBuilder_AutoFlushError(t *testing.T) {
	mock := &mockESClient{
		doReqFunc: func(ctx context.Context, req *http.Request) ([]byte, error) {
			return nil, fmt.Errorf("network error: connection refused")
		},
	}

	b := builder.NewBulkBuilder(mock).
		Index("test").
		AutoFlushSize(2)

	// 3 Add calls: the 3rd triggers auto-flush which will fail
	b.Add("test", "1", map[string]any{"field": "value1"}).
		Add("test", "2", map[string]any{"field": "value2"}).
		Add("test", "3", map[string]any{"field": "value3"})

	_, err := b.Do(context.Background())
	if err == nil {
		t.Fatal("expected Do() to return the auto-flush error, got nil")
	}
	if err.Error() == "network error: connection refused" {
		// Direct match is fine
	} else if !containsSubstring(err.Error(), "auto-flush") && !containsSubstring(err.Error(), "network error") {
		t.Errorf("expected auto-flush error, got: %v", err)
	}
}

// TestBulkBuilder_AutoFlushNoContext verifies that auto-flush uses
// context.Background() as fallback when AutoFlushContext is not called.
func TestBulkBuilder_AutoFlushNoContext(t *testing.T) {
	var capturedCtx context.Context
	mock := &mockESClient{
		doReqFunc: func(ctx context.Context, req *http.Request) ([]byte, error) {
			capturedCtx = ctx
			return []byte(`{"took":1,"errors":false,"items":[{"index":{"_index":"test","_id":"1","status":201}},{"index":{"_index":"test","_id":"2","status":201}}]}`), nil
		},
	}

	b := builder.NewBulkBuilder(mock).
		Index("test").
		AutoFlushSize(2)
	// NOTE: AutoFlushContext is NOT called

	// 3 Add calls: the 3rd triggers flush when len(operations)=2 >= autoFlushSize
	b.Add("test", "1", map[string]any{"field": "value1"}).
		Add("test", "2", map[string]any{"field": "value2"}).
		Add("test", "3", map[string]any{"field": "value3"})

	if capturedCtx == nil {
		t.Fatal("expected DoRequest to be called during auto-flush")
	}
	// context.Background() and context.TODO() are both valid empty contexts,
	// but we specifically want to verify it's not context.TODO().
	// Use fmt.Sprint to get the string representation of the context.
	ctxStr := fmt.Sprint(capturedCtx)
	if ctxStr == "context.TODO" {
		t.Error("auto-flush should use context.Background(), not context.TODO()")
	}
}

// --- Set error propagation tests ---

// TestBulkBuilder_SetErrorPropagation verifies that calling Set() without
// a prior AddDoc() produces an error that propagates through Do().
func TestBulkBuilder_SetErrorPropagation(t *testing.T) {
	mock := &mockESClient{
		doReqFunc: func(ctx context.Context, req *http.Request) ([]byte, error) {
			return []byte(`{"took":1,"errors":false,"items":[]}`), nil
		},
	}

	b := builder.NewBulkBuilder(mock).Index("test")

	// Set without AddDoc should produce an error
	b.Set("field", "value")

	_, err := b.Do(context.Background())
	if err == nil {
		t.Fatal("expected Do() to return Set error, got nil")
	}
	if !containsSubstring(err.Error(), "Set() must be called after") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestBulkBuilder_SetFromStructError verifies that SetFromStruct with
// invalid input produces an error that propagates through Do().
func TestBulkBuilder_SetFromStructError(t *testing.T) {
	mock := &mockESClient{
		doReqFunc: func(ctx context.Context, req *http.Request) ([]byte, error) {
			return []byte(`{"took":1,"errors":false,"items":[]}`), nil
		},
	}

	b := builder.NewBulkBuilder(mock).Index("test")

	// SetFromStruct without prior AddDoc should produce error
	b.SetFromStruct(map[string]any{"key": "value"})

	_, err := b.Do(context.Background())
	if err == nil {
		t.Fatal("expected Do() to return SetFromStruct error, got nil")
	}
	if !containsSubstring(err.Error(), "SetFromStruct() must be called after") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// containsSubstring is a test helper.
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
