package builder_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/Kirby980/go-es/builder"
)

func TestCatBuilder_Do(t *testing.T) {
	mock := &mockESClient{
		doFunc: func(ctx context.Context, method, path string, body any) ([]byte, error) {
			if method != http.MethodGet {
				t.Errorf("Expected GET, got %s", method)
			}
			expectedPath := "/_cat/indices?v=true"
			if path != expectedPath && path != "/_cat/indices?v=true" {
				t.Errorf("Expected path %s, got %s", expectedPath, path)
			}

			resp := `health status index uuid pri rep docs.count docs.deleted store.size pri.store.size`
			return []byte(resp), nil
		},
	}

	b := builder.NewCatBuilder(mock).Debug(true)
	resp, err := b.Do(context.Background(), "indices", map[string]string{"v": "true"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if resp == "" {
		t.Errorf("Expected non-empty response")
	}
}
