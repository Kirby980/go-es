package builder_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/Kirby980/go-es/builder"
)

func TestTermvectorsBuilder_Do(t *testing.T) {
	mock := &mockESClient{
		doFunc: func(ctx context.Context, method, path string, body any) ([]byte, error) {
			if method != http.MethodPost {
				t.Fatalf("expected POST, got %s", method)
			}
			if path != "/docs/_termvectors/1" {
				t.Fatalf("unexpected path: %s", path)
			}
			return []byte(`{"_index":"docs","_id":"1","term_vectors":{}}`), nil
		},
	}

	resp, err := builder.NewTermvectorsBuilder(mock, "docs").
		ID("1").
		Body(map[string]any{"fields": []any{"title"}}).
		Debug(true).
		Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp["_id"] != "1" {
		t.Fatalf("unexpected resp: %v", resp)
	}
}
