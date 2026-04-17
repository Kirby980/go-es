package builder_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/Kirby980/go-es/builder"
)

func TestILMBuilder_Put(t *testing.T) {
	mock := &mockESClient{
		doFunc: func(ctx context.Context, method, path string, body any) ([]byte, error) {
			if method != http.MethodPut {
				t.Errorf("Expected PUT, got %s", method)
			}
			if path != "/_ilm/policy/my_policy" {
				t.Errorf("Expected path /_ilm/policy/my_policy, got %s", path)
			}

			resp := `{"acknowledged": true}`
			return []byte(resp), nil
		},
	}

	b := builder.NewILMBuilder(mock).
		Name("my_policy").
		Policy(map[string]any{"phases": map[string]any{}}).
		Debug(true)

	resp, err := b.Put(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !resp.Acknowledged {
		t.Errorf("Expected acknowledged true")
	}
}

func TestILMBuilder_Get(t *testing.T) {
	mock := &mockESClient{
		doFunc: func(ctx context.Context, method, path string, body any) ([]byte, error) {
			if method != http.MethodGet {
				t.Errorf("Expected GET, got %s", method)
			}
			if path != "/_ilm/policy/my_policy" {
				t.Errorf("Expected path /_ilm/policy/my_policy, got %s", path)
			}

			resp := `{"my_policy": {"version": 1}}`
			return []byte(resp), nil
		},
	}

	b := builder.NewILMBuilder(mock).Name("my_policy").Debug(true)

	resp, err := b.Get(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if _, ok := resp["my_policy"]; !ok {
		t.Errorf("Expected my_policy in response")
	}
}
