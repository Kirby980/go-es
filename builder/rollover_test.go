package builder_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/Kirby980/go-es/builder"
)

func TestRolloverBuilder_Do(t *testing.T) {
	mock := &mockESClient{
		doFunc: func(ctx context.Context, method, path string, body any) ([]byte, error) {
			if method != http.MethodPost {
				t.Fatalf("expected POST, got %s", method)
			}
			if path != "/logs_write/_rollover/logs-000002?dry_run=true" {
				t.Fatalf("unexpected path: %s", path)
			}
			return []byte(`{"acknowledged":true,"shards_acknowledged":true,"old_index":"logs-000001","new_index":"logs-000002","rolled_over":false,"dry_run":true,"conditions":{"max_age":"1d"}}`), nil
		},
	}

	resp, err := builder.NewRolloverBuilder(mock, "logs_write").
		NewIndex("logs-000002").
		Condition("max_age", "1d").
		DryRun(true).
		Debug(true).
		Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Acknowledged || resp.NewIndex != "logs-000002" {
		t.Fatalf("unexpected resp: %+v", resp)
	}
}

