package hooks

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/Kirby980/go-es/logger"
)

func TestMetricsHook(t *testing.T) {
	hook := NewMetricsHook()

	req, _ := http.NewRequest("GET", "http://localhost:9200", nil)
	ctx := hook.BeforeRequest(context.Background(), req)

	hook.AfterRequest(ctx, req, &http.Response{StatusCode: 200}, 10*time.Millisecond)
	hook.AfterRequest(ctx, req, &http.Response{StatusCode: 500}, 20*time.Millisecond)
	hook.OnError(ctx, req, context.DeadlineExceeded, 30*time.Millisecond)

	reqs, errs, avg := hook.GetMetrics()

	if reqs != 3 {
		t.Errorf("expected 3 requests, got %d", reqs)
	}
	if errs != 2 {
		t.Errorf("expected 2 errors, got %d", errs)
	}
	if avg != 20*time.Millisecond {
		t.Errorf("expected avg 20ms, got %v", avg)
	}
}

func TestLogHook(t *testing.T) {
	log, _ := logger.NewDevelopmentLogger()
	hook := NewLogHook(log)

	req, _ := http.NewRequest("GET", "http://localhost:9200", nil)
	ctx := hook.BeforeRequest(context.Background(), req)

	hook.AfterRequest(ctx, req, &http.Response{StatusCode: 200}, 10*time.Millisecond)
	hook.OnError(ctx, req, context.DeadlineExceeded, 30*time.Millisecond)
}
