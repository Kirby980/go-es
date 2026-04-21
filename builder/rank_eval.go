package builder

import (
	"context"
	"net/http"
	"net/url"
)

// RankEvalBuilder 是用于构建和执行 Elasticsearch RankEval 操作的链式 API 构建器。
type RankEvalBuilder struct {
	client ESClient
	index  string
	body   map[string]any
	debugHelper
	baseBuilder
}

// NewRankEvalBuilder 创建并返回一个 RankEval 构建器实例，用于构造和执行 Elasticsearch 的 RankEval 请求。
func NewRankEvalBuilder(c ESClient, index string) *RankEvalBuilder {
	b := &RankEvalBuilder{
		client: c,
		index:  index,
	}
	b.initBaseBuilder()
	return b
}

// Header 设置自定义的 HTTP 请求头 (例如: Header("Content-Type", "application/json"))。此方法支持链式调用。
func (b *RankEvalBuilder) Header(key, value string) *RankEvalBuilder {
	b.baseBuilder.Header(key, value)
	return b
}

// Body 设置请求的原始主体数据。
func (b *RankEvalBuilder) Body(body map[string]any) *RankEvalBuilder {
	b.body = body
	return b
}

// Debug 开启当前请求的单次调试模式。开启后，请求的完整 HTTP URL、Method、Body 及响应将被打印到日志中，便于排查问题。
func (b *RankEvalBuilder) Debug(enable bool) *RankEvalBuilder {
	b.setDebugPersistent(enable)
	return b
}

// Do 立即执行当前构建好的请求，并返回结构化的 Elasticsearch 响应结果或错误。请确保在调用此方法前已设置好所有必要参数。
func (b *RankEvalBuilder) Do(ctx context.Context) (map[string]any, error) {
	path := "/_rank_eval"
	if b.index != "" {
		path = "/" + url.PathEscape(b.index) + path
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
