package builder

import (
	"context"
	"net/http"
	"net/url"
)

// SearchTemplateBuilder 是用于构建和执行 Elasticsearch SearchTemplate 操作的链式 API 构建器。
type SearchTemplateBuilder struct {
	client  ESClient
	index   string
	id      string
	source  any
	params  map[string]any
	explain *bool
	profile *bool
	debugHelper
	baseBuilder
}

// NewSearchTemplateBuilder 创建并返回一个 SearchTemplate 构建器实例，用于构造和执行 Elasticsearch 的 SearchTemplate 请求。
func NewSearchTemplateBuilder(c ESClient) *SearchTemplateBuilder {
	b := &SearchTemplateBuilder{
		client: c,
	}
	b.initBaseBuilder()
	return b
}

// Header 设置自定义的 HTTP 请求头 (例如: Header("Content-Type", "application/json"))。此方法支持链式调用。
func (b *SearchTemplateBuilder) Header(key, value string) *SearchTemplateBuilder {
	b.baseBuilder.Header(key, value)
	return b
}

// Index 指定操作目标的 Elasticsearch 索引名称。
func (b *SearchTemplateBuilder) Index(index string) *SearchTemplateBuilder {
	b.index = index
	return b
}

// ID 指定操作目标文档或资源的唯一标识符 (ID)。
func (b *SearchTemplateBuilder) ID(id string) *SearchTemplateBuilder {
	b.id = id
	return b
}

// Source 设置需要返回的 _source 字段列表，用于控制搜索结果中包含的文档字段。
func (b *SearchTemplateBuilder) Source(source any) *SearchTemplateBuilder {
	b.source = source
	return b
}

// Params 传递脚本或模板执行时所需的动态参数字典。
func (b *SearchTemplateBuilder) Params(params map[string]any) *SearchTemplateBuilder {
	b.params = params
	return b
}

// Explain 开启解释模式，返回查询或执行的详细执行计划和评分来源，通常用于性能调优或调试。
func (b *SearchTemplateBuilder) Explain(enable bool) *SearchTemplateBuilder {
	b.explain = &enable
	return b
}

// Profile 开启性能分析模式，返回查询在各分片上的详细执行时间分解。
func (b *SearchTemplateBuilder) Profile(enable bool) *SearchTemplateBuilder {
	b.profile = &enable
	return b
}

// Debug 开启当前请求的单次调试模式。开启后，请求的完整 HTTP URL、Method、Body 及响应将被打印到日志中，便于排查问题。
func (b *SearchTemplateBuilder) Debug(enable bool) *SearchTemplateBuilder {
	b.setDebugPersistent(enable)
	return b
}

// Do 立即执行当前构建好的请求，并返回结构化的 Elasticsearch 响应结果或错误。请确保在调用此方法前已设置好所有必要参数。
func (b *SearchTemplateBuilder) Do(ctx context.Context) (*SearchResponse, error) {
	path := "/_search/template"
	if b.index != "" {
		path = "/" + url.PathEscape(b.index) + path
	}

	body := make(map[string]any)
	if b.id != "" {
		body["id"] = b.id
	}
	if b.source != nil {
		body["source"] = b.source
	}
	if b.params != nil {
		body["params"] = b.params
	}
	if b.explain != nil {
		body["explain"] = *b.explain
	}
	if b.profile != nil {
		body["profile"] = *b.profile
	}

	if b.isDebug() {
		b.printDebug(http.MethodPost, path, body)
		defer b.autoResetDebug()
	}

	var resp SearchResponse
	if err := b.client.DoWithHeaderAndDecode(ctx, http.MethodPost, path, body, b.getHeaders(), &resp); err != nil {
		return nil, err
	}

	if b.isDebug() {
		b.printResponseObj(resp)
	}

	return &resp, nil
}
