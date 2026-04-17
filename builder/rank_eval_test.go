package builder_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/Kirby980/go-es/builder"
)

func TestRankEvalBuilder_Do(t *testing.T) {
	mock := &mockESClient{
		doFunc: func(ctx context.Context, method, path string, body any) ([]byte, error) {
			if method != http.MethodPost {
				t.Fatalf("expected POST, got %s", method)
			}
			if path != "/products/_rank_eval" {
				t.Fatalf("unexpected path: %s", path)
			}
			return []byte(`{"metric_score":0.5}`), nil
		},
	}

	resp, err := builder.NewRankEvalBuilder(mock, "products").
		Body(map[string]any{"requests": []any{}, "metric": map[string]any{"precision": map[string]any{}}}).
		Debug(true).
		Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp["metric_score"] != 0.5 {
		t.Fatalf("unexpected resp: %v", resp)
	}
}

