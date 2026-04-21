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

// BeforeRequest 在 HTTP 请求实际发送到 Elasticsearch 之前被调用，可用于请求修改、打点或上下文注入。
func (h *LogHook) BeforeRequest(ctx context.Context, req *http.Request) context.Context {
	// 在此处可以记录开始请求的日志，但通常在 AfterRequest/OnError 记录即可
	return ctx
}

// AfterRequest 在 Elasticsearch 成功返回 HTTP 响应后被调用，可用于记录耗时、解析响应头或指标统计。
func (h *LogHook) AfterRequest(ctx context.Context, req *http.Request, resp *http.Response, duration time.Duration) {
	h.logger.Info("Elasticsearch Request Success",
		"method", req.Method,
		"url", req.URL.String(),
		"status", resp.StatusCode,
		"duration", duration.String(),
	)
}

// OnError 在发起 HTTP 请求遇到网络错误或连接失败时被调用。
func (h *LogHook) OnError(ctx context.Context, req *http.Request, err error, duration time.Duration) {
	h.logger.Error("Elasticsearch Request Failed",
		"method", req.Method,
		"url", req.URL.String(),
		"error", err.Error(),
		"duration", duration.String(),
	)
}
