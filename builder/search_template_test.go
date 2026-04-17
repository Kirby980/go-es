package builder_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/Kirby980/go-es/builder"
)

func TestSearchTemplateBuilder_Do(t *testing.T) {
	mock := &mockESClient{
		doFunc: func(ctx context.Context, method, path string, body any) ([]byte, error) {
			if method != http.MethodPost {
				t.Fatalf("expected POST, got %s", method)
			}
			if path != "/books/_search/template" {
				t.Fatalf("expected path /books/_search/template, got %s", path)
			}
			return []byte(`{"took":1,"timed_out":false,"_shards":{"total":1,"successful":1,"skipped":0,"failed":0},"hits":{"total":{"value":0,"relation":"eq"},"max_score":0,"hits":[]}}`), nil
		},
	}

	resp, err := builder.NewSearchTemplateBuilder(mock).
		Index("books").
		ID("my_tpl").
		Params(map[string]any{"q": "foo"}).
		Debug(true).
		Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Took != 1 {
		t.Fatalf("unexpected resp: %v", resp)
	}
}

