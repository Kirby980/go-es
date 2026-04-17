package hooks

import (
	"context"
	"net/http"
	"time"

	"github.com/Kirby980/go-es/config"
	"github.com/Kirby980/go-es/logger"
)

// LogHook 是一个记录请求日志的 Hook
type LogHook struct {
	logger logger.Logger
}

// NewLogHook 创建一个新的日志 Hook
func NewLogHook(l logger.Logger) config.Hook {
	return &LogHook{logger: l}
}

func (h *LogHook) BeforeRequest(ctx context.Context, req *http.Request) context.Context {
	// 在此处可以记录开始请求的日志，但通常在 AfterRequest/OnError 记录即可
	return ctx
}

func (h *LogHook) AfterRequest(ctx context.Context, req *http.Request, resp *http.Response, duration time.Duration) {
	h.logger.Info("Elasticsearch Request Success",
		"method", req.Method,
		"url", req.URL.String(),
		"status", resp.StatusCode,
		"duration", duration.String(),
	)
}

func (h *LogHook) OnError(ctx context.Context, req *http.Request, err error, duration time.Duration) {
	h.logger.Error("Elasticsearch Request Failed",
		"method", req.Method,
		"url", req.URL.String(),
		"error", err.Error(),
		"duration", duration.String(),
	)
}
