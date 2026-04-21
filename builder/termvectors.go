package builder

import (
	"context"
	"net/http"
	"net/url"
)

// TermvectorsBuilder 是用于构建和执行 Elasticsearch Termvectors 操作的链式 API 构建器。
type TermvectorsBuilder struct {
	client ESClient
	index  string
	id     string
	body   map[string]any
	debugHelper
	baseBuilder
}

// NewTermvectorsBuilder 创建并返回一个 Termvectors 构建器实例，用于构造和执行 Elasticsearch 的 Termvectors 请求。
func NewTermvectorsBuilder(c ESClient, index string) *TermvectorsBuilder {
	b := &TermvectorsBuilder{
		client: c,
		index:  index,
	}
	b.initBaseBuilder()
	return b
}

// Header 设置自定义的 HTTP 请求头 (例如: Header("Content-Type", "application/json"))。此方法支持链式调用。
func (b *TermvectorsBuilder) Header(key, value string) *TermvectorsBuilder {
	b.baseBuilder.Header(key, value)
	return b
}

// ID 指定操作目标文档或资源的唯一标识符 (ID)。
func (b *TermvectorsBuilder) ID(id string) *TermvectorsBuilder {
	b.id = id
	return b
}

// Body 设置请求的原始主体数据。
func (b *TermvectorsBuilder) Body(body map[string]any) *TermvectorsBuilder {
	b.body = body
	return b
}

// Debug 开启当前请求的单次调试模式。开启后，请求的完整 HTTP URL、Method、Body 及响应将被打印到日志中，便于排查问题。
func (b *TermvectorsBuilder) Debug(enable bool) *TermvectorsBuilder {
	b.setDebugPersistent(enable)
	return b
}

// Do 立即执行当前构建好的请求，并返回结构化的 Elasticsearch 响应结果或错误。请确保在调用此方法前已设置好所有必要参数。
func (b *TermvectorsBuilder) Do(ctx context.Context) (map[string]any, error) {
	path := "/" + url.PathEscape(b.index) + "/_termvectors"
	if b.id != "" {
		path += "/" + url.PathEscape(b.id)
	}

	if b.isDebug() {
		b.printDebug(http.MethodPost, path, b.body)
		defer b.autoResetDebug()
	}

	var resp map[string]any
	if err := b.client.DoWithHeaderAndDecode(ctx, http.MethodPost, path, b.body, b.getHeaders(), &resp); err != nil {
		return nil, err
	}

	if b.isDebug() {
		b.printResponseObj(resp)
	}

	return resp, nil
}
