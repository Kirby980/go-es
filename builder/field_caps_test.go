package builder_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/Kirby980/go-es/builder"
)

func TestFieldCapsBuilder_Do(t *testing.T) {
	mock := &mockESClient{
		doFunc: func(ctx context.Context, method, path string, body any) ([]byte, error) {
			if method != http.MethodGet {
				t.Fatalf("expected GET, got %s", method)
			}
			if path != "/logs/_field_caps?fields=host%2Cmessage&include_unmapped=true" {
				t.Fatalf("unexpected path: %s", path)
			}
			return []byte(`{"fields":{"host":{"keyword":{"type":"keyword","searchable":true,"aggregatable":true}}}}`), nil
		},
	}

	resp, err := builder.NewFieldCapsBuilder(mock, "logs").
		Fields("host", "message").
		IncludeUnmapped(true).
		Debug(true).
		Do(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp["fields"] == nil {
		t.Fatalf("unexpected resp: %v", resp)
	}
}
