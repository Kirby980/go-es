package hooks

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Kirby980/go-es/builder"
	"github.com/Kirby980/go-es/client"
	"github.com/Kirby980/go-es/config"
)

func TestHookIntegration(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"_index":"test","_id":"1","_version":1,"result":"created"}`))
	}))
	defer ts.Close()

	metrics := NewMetricsHook()

	esClient, err := client.New(
		config.WithAddresses(ts.URL),
		config.WithHooks(metrics),
	)
	if err != nil {
		t.Fatal(err)
	}

	docBuilder := builder.NewDocumentBuilder(esClient, "test_idx")
	_, err = docBuilder.Set("msg", "hello").Do(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	reqs, errs, _ := metrics.GetMetrics()
	if reqs != 1 {
		t.Errorf("Expected 1 request, got %d", reqs)
	}
	if errs != 0 {
		t.Errorf("Expected 0 errors, got %d", errs)
	}
}
