package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Kirby980/go-es/config"
	"github.com/Kirby980/go-es/logger"
)

func TestCreateIndexRaw_SendsRawJSON(t *testing.T) {
	raw := `{"mappings":{"properties":{"id":{"type":"long"}}}}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/my_index" {
			t.Fatalf("expected path /my_index, got %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)

		var got, want bytes.Buffer
		if err := json.Compact(&got, body); err != nil {
			t.Fatalf("invalid json received: %v, body=%s", err, string(body))
		}
		if err := json.Compact(&want, []byte(raw)); err != nil {
			t.Fatalf("invalid raw json: %v", err)
		}
		if got.String() != want.String() {
			t.Fatalf("unexpected body: got=%s want=%s", got.String(), want.String())
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"acknowledged":true}`))
	}))
	defer srv.Close()

	c := &Client{
		config: &config.Config{
			MaxRetries:   0,
			RetryBackoff: time.Millisecond,
		},
		httpClient: srv.Client(),
		addresses:  []string{srv.URL},
		logger:     logger.NopLogger{},
	}

	if err := c.CreateIndexRaw(context.Background(), "my_index", raw); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
