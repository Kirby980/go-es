package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Kirby980/go-es/config"
	"github.com/Kirby980/go-es/logger"
)

// --- DoRequest body reset tests ---

// TestDoRequestBodyReset_RetryWithBody verifies that DoRequest sends the
// full original body on every retry attempt, not an empty/EOF body.
func TestDoRequestBodyReset_RetryWithBody(t *testing.T) {
	var attempt atomic.Int32
	var bodies [][]byte

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		n := attempt.Add(1)
		bodies = append(bodies, body)
		if n <= 1 {
			// Fail first attempt to trigger retry by closing connection.
			hj, ok := w.(http.Hijacker)
			if ok {
				conn, _, _ := hj.Hijack()
				conn.Close()
				return
			}
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := &Client{
		config: &config.Config{
			MaxRetries:   3,
			RetryBackoff: 1 * time.Millisecond,
		},
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				DisableKeepAlives: true,
			},
		},
		addresses: []string{srv.URL},
		logger:    logger.NopLogger{},
	}

	payload := []byte(`{"index":"test","doc":{"field":"value"}}`)
	bodyReader := bytes.NewReader(payload)

	req, err := http.NewRequestWithContext(context.Background(), "POST", srv.URL+"/_bulk", bodyReader)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	resp, err := c.DoRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("DoRequest failed: %v", err)
	}

	if string(resp) != `{"ok":true}` {
		t.Errorf("unexpected response: %s", resp)
	}

	// The critical assertion: the server must have received the full body
	// on the retry attempt, not an empty body.
	if len(bodies) < 2 {
		t.Fatalf("expected at least 2 attempts, got %d", len(bodies))
	}
	retryBody := bodies[len(bodies)-1]
	if string(retryBody) != string(payload) {
		t.Errorf("retry sent wrong body: got %q, want %q", string(retryBody), string(payload))
	}
}

// TestDoRequestBodyReset_NilBody verifies DoRequest with nil body and
// retry does not panic.
func TestDoRequestBodyReset_NilBody(t *testing.T) {
	var attempt atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := attempt.Add(1)
		if n <= 1 {
			hj, ok := w.(http.Hijacker)
			if ok {
				conn, _, _ := hj.Hijack()
				conn.Close()
			}
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := &Client{
		config: &config.Config{
			MaxRetries:   2,
			RetryBackoff: 1 * time.Millisecond,
		},
		httpClient: srv.Client(),
		addresses:  []string{srv.URL},
		logger:     logger.NopLogger{},
	}

	req, err := http.NewRequestWithContext(context.Background(), "GET", srv.URL+"/", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	_, err = c.DoRequest(context.Background(), req)
	if err != nil {
		t.Fatalf("DoRequest with nil body should not fail: %v", err)
	}
}

// --- Ping tests ---

// TestPing_AllAddresses_FirstDown verifies that Ping succeeds when the
// first address is down but the second address is healthy.
func TestPing_AllAddresses_FirstDown(t *testing.T) {
	// Create two servers; close the first one so it's unreachable.
	srv1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	srv1.Close() // make first address unreachable

	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv2.Close()

	c := &Client{
		config: &config.Config{},
		httpClient: &http.Client{
			Timeout: 2 * time.Second,
		},
		addresses: []string{srv1.URL, srv2.URL},
		logger:    logger.NopLogger{},
	}

	err := c.Ping(context.Background())
	if err != nil {
		t.Errorf("Ping should succeed when second address is up, got: %v", err)
	}
}

// TestPing_AllAddresses_AllDown verifies that Ping returns an error when
// all servers are unreachable.
func TestPing_AllAddresses_AllDown(t *testing.T) {
	srv1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv1.Close()
	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv2.Close()

	c := &Client{
		config: &config.Config{},
		httpClient: &http.Client{
			Timeout: 2 * time.Second,
		},
		addresses: []string{srv1.URL, srv2.URL},
		logger:    logger.NopLogger{},
	}

	err := c.Ping(context.Background())
	if err == nil {
		t.Error("Ping should return error when all addresses are down")
	}
}

// TestPing_AllAddresses_FirstUp verifies that Ping returns nil immediately
// when the first address is healthy.
func TestPing_AllAddresses_FirstUp(t *testing.T) {
	srv1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv1.Close()

	// Second server should NOT be reached. We verify by checking
	// the request count below.
	var srv2Hits atomic.Int32
	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		srv2Hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv2.Close()

	c := &Client{
		config: &config.Config{},
		httpClient: &http.Client{
			Timeout: 2 * time.Second,
		},
		addresses: []string{srv1.URL, srv2.URL},
		logger:    logger.NopLogger{},
	}

	err := c.Ping(context.Background())
	if err != nil {
		t.Errorf("Ping should succeed when first address is up, got: %v", err)
	}
	if srv2Hits.Load() != 0 {
		t.Error("Ping should not check second address when first is healthy")
	}
}

// TestPing_ContextCancelled verifies that Ping respects context cancellation
// and returns the context error promptly.
func TestPing_ContextCancelled(t *testing.T) {
	// Create a slow server that blocks.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := &Client{
		config: &config.Config{},
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		addresses: []string{srv.URL},
		logger:    logger.NopLogger{},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	err := c.Ping(ctx)
	if err == nil {
		t.Error("Ping should return error when context is cancelled")
	}
}

// TestPing_EmptyAddresses verifies that Ping with no addresses returns
// an error (not a panic).
func TestPing_EmptyAddresses(t *testing.T) {
	c := &Client{
		config: &config.Config{},
		httpClient: &http.Client{
			Timeout: 2 * time.Second,
		},
		addresses: []string{},
		logger:    logger.NopLogger{},
	}

	err := c.Ping(context.Background())
	if err == nil {
		t.Error("Ping should return error when addresses list is empty")
	}
	expected := "no addresses configured"
	if err != nil && err.Error() != expected {
		// Accept any non-nil error, but prefer our specific message.
		fmt.Printf("Note: got error %q (expected %q)\n", err.Error(), expected)
	}
}
