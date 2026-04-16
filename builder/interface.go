package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/Kirby980/go-es/logger"
)

// ESClient builder 包需要的客户端接口
// 不依赖 client 包，避免循环引用
type ESClient interface {
	// Do 执行 HTTP 请求（JSON 格式）
	Do(ctx context.Context, method, path string, body any) ([]byte, error)
	// DoWithHeader 执行 HTTP 请求并带上自定义 Header
	DoWithHeader(ctx context.Context, method, path string, body any, header http.Header) ([]byte, error)
	// DoWithHeaderAndDecode 执行 HTTP 请求并使用流式解析直接反序列化到 target，避免一次性读入全量 []byte 造成高内存占用
	DoWithHeaderAndDecode(ctx context.Context, method, path string, body any, header http.Header, target any) error
	// GetAddress 获取 ES 地址（用于 bulk 等需要自定义 Content-Type 的场景）
	GetAddress() string
	// DoRequest 执行自定义 HTTP 请求（用于 bulk 等需要 ndjson 格式的场景）
	DoRequest(ctx context.Context, req *http.Request) ([]byte, error)
	// GetLogger 获取客户端日志实例，供 builder 内部使用
	GetLogger() logger.Logger
}

// IndexName 用于自定义索引名（类似 GORM 的 Tabler 接口）
type IndexName interface {
	IndexName() string
}

// structToMap 内部辅助方法，高效地将结构体转为 map[string]any
// 使用原生的 JSON 序列化和反序列化
func structToMap(v any) (map[string]any, error) {
	if v == nil {
		return nil, nil
	}
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct && t.Kind() != reflect.Map {
		return nil, fmt.Errorf("expected struct or map, got %v", t.Kind())
	}

	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	err = json.Unmarshal(data, &m)
	return m, err
}
