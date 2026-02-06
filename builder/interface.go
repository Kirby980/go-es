package builder

import (
	"context"
	"net/http"
)

// ESClient builder 包需要的客户端接口
// 不依赖 client 包，避免循环引用
type ESClient interface {
	// Do 执行 HTTP 请求（JSON 格式）
	Do(ctx context.Context, method, path string, body interface{}) ([]byte, error)
	// GetAddress 获取 ES 地址（用于 bulk 等需要自定义 Content-Type 的场景）
	GetAddress() string
	// DoRequest 执行自定义 HTTP 请求（用于 bulk 等需要 ndjson 格式的场景）
	DoRequest(ctx context.Context, req *http.Request) ([]byte, error)
}

// IndexName 用于自定义索引名（类似 GORM 的 Tabler 接口）
type IndexName interface {
	IndexName() string
}
