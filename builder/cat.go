package builder

import (
	"context"
	"net/http"
)

// CatBuilder 提供对 Elasticsearch _cat API 的支持
type CatBuilder struct {
	client ESClient
	debugHelper
	baseBuilder
}

// NewCatBuilder 创建并返回一个 Cat 构建器实例，用于构造和执行 Elasticsearch 的 Cat 请求。
func NewCatBuilder(client ESClient) *CatBuilder {
	return &CatBuilder{
		client: client,
	}
}

// Debug 开启当前请求的单次调试模式。开启后，请求的完整 HTTP URL、Method、Body 及响应将被打印到日志中，便于排查问题。
func (b *CatBuilder) Debug(enable bool) *CatBuilder {
	b.setDebug(enable)
	return b
}

// Do 立即执行当前构建好的请求，并返回结构化的 Elasticsearch 响应结果或错误。请确保在调用此方法前已设置好所有必要参数。
func (b *CatBuilder) Do(ctx context.Context, api string, params map[string]string) (string, error) {
	path := "/_cat/" + api
	if len(params) > 0 {
		path += "?"
		first := true
		for k, v := range params {
			if !first {
				path += "&"
			}
			path += k + "=" + v
			first = false
		}
	}

	if b.isDebug() {
		b.printDebug("GET", path, nil)
		defer b.autoResetDebug()
	}

	resp, err := b.client.DoWithHeader(ctx, http.MethodGet, path, nil, b.getHeaders())
	if err != nil {
		return "", err
	}

	return string(resp), nil
}
