package hooks

import (
	"context"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/Kirby980/go-es/config"
)

// MetricsHook 收集简单的指标（原子计数器）
type MetricsHook struct {
	TotalRequests uint64
	TotalErrors   uint64
	TotalDuration uint64 // 纳秒
}

// NewMetricsHook 创建一个新的指标 Hook
func NewMetricsHook() *MetricsHook {
	return &MetricsHook{}
}

// BeforeRequest ...
func (h *MetricsHook) BeforeRequest(ctx context.Context, req *http.Request) context.Context {
	return ctx
}

// AfterRequest ...
func (h *MetricsHook) AfterRequest(ctx context.Context, req *http.Request, resp *http.Response, duration time.Duration) {
	atomic.AddUint64(&h.TotalRequests, 1)
	atomic.AddUint64(&h.TotalDuration, uint64(duration.Nanoseconds()))
	if resp != nil && resp.StatusCode >= 400 {
		atomic.AddUint64(&h.TotalErrors, 1)
	}
}

// OnError ...
func (h *MetricsHook) OnError(ctx context.Context, req *http.Request, err error, duration time.Duration) {
	atomic.AddUint64(&h.TotalRequests, 1)
	atomic.AddUint64(&h.TotalErrors, 1)
	atomic.AddUint64(&h.TotalDuration, uint64(duration.Nanoseconds()))
}

// GetMetrics ...
func (h *MetricsHook) GetMetrics() (reqs, errs uint64, avgDuration time.Duration) {
	reqs = atomic.LoadUint64(&h.TotalRequests)
	errs = atomic.LoadUint64(&h.TotalErrors)
	totDur := atomic.LoadUint64(&h.TotalDuration)
	if reqs > 0 {
		avgDuration = time.Duration(totDur / reqs)
	}
	return
}

// 确保实现 Hook 接口
var _ config.Hook = (*MetricsHook)(nil)
