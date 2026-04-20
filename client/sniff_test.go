package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Kirby980/go-es/config"
	"github.com/Kirby980/go-es/logger"
)

func TestSniffOnce_UpdateAddresses(t *testing.T) {
	var c *Client

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/_nodes/http" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"nodes": {
				"n1": {"http": {"publish_address": "127.0.0.1:9200"}},
				"n2": {"http": {"publish_address": "127.0.0.1:9201"}}
			}
		}`))
	}))
	defer srv.Close()

	c = &Client{
		config: &config.Config{
			EnableSniff:   true,
			SniffInterval: 10 * time.Minute,
		},
		httpClient: srv.Client(),
		addresses:  []string{srv.URL},
		logger:     logger.NopLogger{},
	}

	c.sniffOnce(context.Background())

	c.addressesMu.RLock()
	defer c.addressesMu.RUnlock()
	if len(c.addresses) != 2 {
		t.Fatalf("expected 2 addresses, got %d: %#v", len(c.addresses), c.addresses)
	}
	if c.addresses[0] != "http://127.0.0.1:9200" && c.addresses[1] != "http://127.0.0.1:9200" {
		t.Fatalf("expected http://127.0.0.1:9200 in addresses, got %#v", c.addresses)
	}
}

func TestSniffOnce_PublishAddressWithSlash(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"nodes": {
				"n1": {"http": {"publish_address": "127.0.0.1/127.0.0.1:9202"}}
			}
		}`))
	}))
	defer srv.Close()

	c := &Client{
		config: &config.Config{
			EnableSniff: true,
		},
		httpClient: srv.Client(),
		addresses:  []string{srv.URL},
		logger:     logger.NopLogger{},
	}

	c.sniffOnce(context.Background())

	c.addressesMu.RLock()
	defer c.addressesMu.RUnlock()
	if len(c.addresses) != 1 || c.addresses[0] != "http://127.0.0.1:9202" {
		t.Fatalf("unexpected addresses: %#v", c.addresses)
	}
}
