package builder_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/Kirby980/go-es/builder"
)

func TestAsyncSearchBuilder_Do(t *testing.T) {
	mock := &mockESClient{
		doFunc: func(ctx context.Context, method, path string, body any) ([]byte, error) {
			if method != http.MethodPost {
				t.Errorf("Expected POST, got %s", method)
			}
			expectedPath := "/index1/_async_search?wait_for_completion_timeout=1s&keep_alive=5m"
			if path != expectedPath {
				t.Errorf("Expected path %s, got %s", expectedPath, path)
			}

			resp := `{
				"id": "FmRldE8zREVEUzA2ZVpUeGs2ejJFUFEaMkZ5QTVrSTZSaVN3WlI2MTZ0aXAL1",
				"is_partial": true,
				"is_running": true,
				"start_time_in_millis": 1583945890986
			}`
			return []byte(resp), nil
		},
	}

	b := builder.NewAsyncSearchBuilder(mock).
		Index("index1").
		Query(map[string]any{"match_all": map[string]any{}}).
		WaitForCompletionTimeout("1s").
		KeepAlive("5m").
		Debug(true)

	resp, err := b.Do(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if resp.ID != "FmRldE8zREVEUzA2ZVpUeGs2ejJFUFEaMkZ5QTVrSTZSaVN3WlI2MTZ0aXAL1" {
		t.Errorf("Unexpected ID: %s", resp.ID)
	}
	if !resp.IsRunning {
		t.Errorf("Expected IsRunning to be true")
	}
}

func TestAsyncSearchBuilder_Get(t *testing.T) {
	mock := &mockESClient{
		doFunc: func(ctx context.Context, method, path string, body any) ([]byte, error) {
			if method != http.MethodGet {
				t.Errorf("Expected GET, got %s", method)
			}
			expectedPath := "/_async_search/my_id"
			if path != expectedPath {
				t.Errorf("Expected path %s, got %s", expectedPath, path)
			}

			resp := `{
				"id": "my_id",
				"is_partial": false,
				"is_running": false
			}`
			return []byte(resp), nil
		},
	}

	b := builder.NewAsyncSearchBuilder(mock).Debug(true)
	resp, err := b.Get(context.Background(), "my_id")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if resp.IsRunning {
		t.Errorf("Expected IsRunning to be false")
	}
}
