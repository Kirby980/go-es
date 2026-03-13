package builder_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/Kirby980/go-es/builder"
	"github.com/Kirby980/go-es/errors"
)

// --- DocumentBuilder.Exists tests ---

func TestDocumentExists_Found(t *testing.T) {
	mock := &mockESClient{
		doFunc: func(_ context.Context, _, _ string, _ any) ([]byte, error) {
			return []byte(`{}`), nil // 200 OK
		},
	}

	exists, err := builder.NewDocumentBuilder(mock, "test-index").ID("doc-1").Exists(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !exists {
		t.Fatal("expected exists=true, got false")
	}
}

func TestDocumentExists_NotFound(t *testing.T) {
	mock := &mockESClient{
		doFunc: func(_ context.Context, _, _ string, _ any) ([]byte, error) {
			return nil, &errors.ESError{StatusCode: 404, Type: "index_not_found_exception", Reason: "no such index"}
		},
	}

	exists, err := builder.NewDocumentBuilder(mock, "test-index").ID("doc-1").Exists(context.Background())
	if err != nil {
		t.Fatalf("expected no error for 404, got %v", err)
	}
	if exists {
		t.Fatal("expected exists=false for 404, got true")
	}
}

func TestDocumentExists_ServerError(t *testing.T) {
	serverErr := &errors.ESError{StatusCode: 500, Type: "internal_server_error", Reason: "cluster is red"}
	mock := &mockESClient{
		doFunc: func(_ context.Context, _, _ string, _ any) ([]byte, error) {
			return nil, serverErr
		},
	}

	exists, err := builder.NewDocumentBuilder(mock, "test-index").ID("doc-1").Exists(context.Background())
	if err == nil {
		t.Fatal("expected error for 500, got nil")
	}
	if exists {
		t.Fatal("expected exists=false for 500, got true")
	}
	// Verify the error is propagated, not swallowed
	if err != serverErr {
		t.Fatalf("expected the original 500 error to be propagated, got %v", err)
	}
}

func TestDocumentExists_NetworkError(t *testing.T) {
	networkErr := fmt.Errorf("connection refused: %w", fmt.Errorf("dial tcp 127.0.0.1:9200: connect"))
	mock := &mockESClient{
		doFunc: func(_ context.Context, _, _ string, _ any) ([]byte, error) {
			return nil, networkErr
		},
	}

	exists, err := builder.NewDocumentBuilder(mock, "test-index").ID("doc-1").Exists(context.Background())
	if err == nil {
		t.Fatal("expected error for network failure, got nil")
	}
	if exists {
		t.Fatal("expected exists=false for network failure, got true")
	}
	if err != networkErr {
		t.Fatalf("expected original network error, got %v", err)
	}
}

func TestDocumentExists_NoID(t *testing.T) {
	mock := &mockESClient{}

	exists, err := builder.NewDocumentBuilder(mock, "test-index").Exists(context.Background())
	if err == nil {
		t.Fatal("expected error when no ID is set, got nil")
	}
	if exists {
		t.Fatal("expected exists=false when no ID is set, got true")
	}
}

// --- IndexBuilder.Exists tests ---

func TestIndexExists_Found(t *testing.T) {
	mock := &mockESClient{
		doFunc: func(_ context.Context, _, _ string, _ any) ([]byte, error) {
			return []byte(`{}`), nil // 200 OK
		},
	}

	exists, err := builder.NewIndexBuilder(mock, "test-index").Exists(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !exists {
		t.Fatal("expected exists=true, got false")
	}
}

func TestIndexExists_NotFound(t *testing.T) {
	mock := &mockESClient{
		doFunc: func(_ context.Context, _, _ string, _ any) ([]byte, error) {
			return nil, &errors.ESError{StatusCode: 404, Type: "index_not_found_exception", Reason: "no such index"}
		},
	}

	exists, err := builder.NewIndexBuilder(mock, "test-index").Exists(context.Background())
	if err != nil {
		t.Fatalf("expected no error for 404, got %v", err)
	}
	if exists {
		t.Fatal("expected exists=false for 404, got true")
	}
}

func TestIndexExists_ServerError(t *testing.T) {
	serverErr := &errors.ESError{StatusCode: 500, Type: "internal_server_error", Reason: "cluster is red"}
	mock := &mockESClient{
		doFunc: func(_ context.Context, _, _ string, _ any) ([]byte, error) {
			return nil, serverErr
		},
	}

	exists, err := builder.NewIndexBuilder(mock, "test-index").Exists(context.Background())
	if err == nil {
		t.Fatal("expected error for 500, got nil")
	}
	if exists {
		t.Fatal("expected exists=false for 500, got true")
	}
	if err != serverErr {
		t.Fatalf("expected the original 500 error to be propagated, got %v", err)
	}
}
