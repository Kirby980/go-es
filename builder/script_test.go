package builder_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/Kirby980/go-es/builder"
)

func TestScriptBuilder_Put(t *testing.T) {
	mock := &mockESClient{
		doFunc: func(ctx context.Context, method, path string, body any) ([]byte, error) {
			if method != http.MethodPut {
				t.Fatalf("expected PUT, got %s", method)
			}
			if path != "/_scripts/my_script" {
				t.Fatalf("expected path /_scripts/my_script, got %s", path)
			}
			return []byte(`{"acknowledged": true}`), nil
		},
	}

	resp, err := builder.NewScriptBuilder(mock).
		ID("my_script").
		Script("painless", "return 1", nil).
		Debug(true).
		Put(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Acknowledged {
		t.Fatalf("expected acknowledged true")
	}
}

func TestScriptBuilder_Get(t *testing.T) {
	mock := &mockESClient{
		doFunc: func(ctx context.Context, method, path string, body any) ([]byte, error) {
			if method != http.MethodGet {
				t.Fatalf("expected GET, got %s", method)
			}
			if path != "/_scripts/my_script" {
				t.Fatalf("expected path /_scripts/my_script, got %s", path)
			}
			return []byte(`{"_id":"my_script","found":true}`), nil
		},
	}

	resp, err := builder.NewScriptBuilder(mock).ID("my_script").Debug(true).Get(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp["_id"] != "my_script" {
		t.Fatalf("unexpected resp: %v", resp)
	}
}
