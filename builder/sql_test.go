package builder_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/Kirby980/go-es/builder"
)

func TestSQLBuilder(t *testing.T) {
	mock := &mockESClient{
		doFunc: func(ctx context.Context, method, path string, body any) ([]byte, error) {
			if method != http.MethodPost {
				t.Errorf("Expected POST, got %s", method)
			}
			if path != "/_sql?format=json" {
				t.Errorf("Expected /_sql?format=json, got %s", path)
			}

			bodyMap, ok := body.(map[string]any)
			if !ok {
				t.Errorf("Expected map[string]any body")
			}

			if bodyMap["query"] != "SELECT * FROM library" {
				t.Errorf("Expected query 'SELECT * FROM library', got %v", bodyMap["query"])
			}

			if bodyMap["fetch_size"] != 10 {
				t.Errorf("Expected fetch_size 10, got %v", bodyMap["fetch_size"])
			}

			resp := `{
				"columns": [{"name": "author", "type": "text"}],
				"rows": [["Dan Simmons"]]
			}`
			return []byte(resp), nil
		},
	}

	b := builder.NewSQLBuilder(mock).
		Query("SELECT * FROM library").
		FetchSize(10).
		Debug(true)

	resp, err := b.Do(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(resp.Columns) != 1 || resp.Columns[0].Name != "author" {
		t.Errorf("Unexpected columns: %v", resp.Columns)
	}

	if len(resp.Rows) != 1 || resp.Rows[0][0] != "Dan Simmons" {
		t.Errorf("Unexpected rows: %v", resp.Rows)
	}
}
