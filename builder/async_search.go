package builder

import (
	"context"
	"net/http"
	"net/url"
)

// AsyncSearchBuilder 是用于构建和执行 Elasticsearch AsyncSearch 操作的链式 API 构建器。
type AsyncSearchBuilder struct {
	client                   ESClient
	index                    string
	query                    map[string]any
	waitForCompletionTimeout string
	keepAlive                string
	debugHelper
	baseBuilder
}

// NewAsyncSearchBuilder 创建并返回一个 AsyncSearch 构建器实例，用于构造和执行 Elasticsearch 的 AsyncSearch 请求。
func NewAsyncSearchBuilder(client ESClient) *AsyncSearchBuilder {
	return &AsyncSearchBuilder{
		client: client,
	}
}

// Index 指定操作目标的 Elasticsearch 索引名称。
func (b *AsyncSearchBuilder) Index(index string) *AsyncSearchBuilder {
	b.index = index
	return b
}

// Query 设置核心的查询 DSL 条件。
func (b *AsyncSearchBuilder) Query(query map[string]any) *AsyncSearchBuilder {
	b.query = query
	return b
}

// WaitForCompletionTimeout 设置等待异步操作完成的超时时间。如果在此时间内未完成，请求将返回一个任务 ID，以便后续轮询状态。
func (b *AsyncSearchBuilder) WaitForCompletionTimeout(timeout string) *AsyncSearchBuilder {
	b.waitForCompletionTimeout = timeout
	return b
}

// KeepAlive 设置上下文 (如 Scroll, PIT, Async Search) 在服务器端保留的时间长度 (如 "5m" 表示 5 分钟)。
func (b *AsyncSearchBuilder) KeepAlive(keepAlive string) *AsyncSearchBuilder {
	b.keepAlive = keepAlive
	return b
}

// Debug 开启当前请求的单次调试模式。开启后，请求的完整 HTTP URL、Method、Body 及响应将被打印到日志中，便于排查问题。
func (b *AsyncSearchBuilder) Debug(enable bool) *AsyncSearchBuilder {
	b.setDebug(enable)
	return b
}

// AsyncSearchResponse 定义了 Elasticsearch AsyncSearch 操作返回的结构化响应数据，包含元数据及核心结果。
type AsyncSearchResponse struct {
	ID             string         `json:"id"`
	IsPartial      bool           `json:"is_partial"`
	IsRunning      bool           `json:"is_running"`
	StartTime      int64          `json:"start_time_in_millis"`
	ExpirationTime int64          `json:"expiration_time_in_millis"`
	Response       SearchResponse `json:"response"`
}

// Do 立即执行当前构建好的请求，并返回结构化的 Elasticsearch 响应结果或错误。请确保在调用此方法前已设置好所有必要参数。
func (b *AsyncSearchBuilder) Do(ctx context.Context) (*AsyncSearchResponse, error) {
	path := "/_async_search"
	if b.index != "" {
		path = "/" + url.PathEscape(b.index) + path
	}

	if b.waitForCompletionTimeout != "" || b.keepAlive != "" {
		path += "?"
		first := true
		if b.waitForCompletionTimeout != "" {
			path += "wait_for_completion_timeout=" + b.waitForCompletionTimeout
			first = false
		}
		if b.keepAlive != "" {
			if !first {
				path += "&"
			}
			path += "keep_alive=" + b.keepAlive
		}
	}

	body := map[string]any{
		"query": b.query,
	}

	if b.isDebug() {
		b.printDebug("POST", path, body)
		defer b.autoResetDebug()
	}

	var resp AsyncSearchResponse
	err := b.client.DoWithHeaderAndDecode(ctx, http.MethodPost, path, body, b.getHeaders(), &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// Get 发起 GET 请求以获取目标资源的详细信息。
func (b *AsyncSearchBuilder) Get(ctx context.Context, id string) (*AsyncSearchResponse, error) {
	path := "/_async_search/" + url.PathEscape(id)

	if b.isDebug() {
		b.printDebug("GET", path, nil)
		defer b.autoResetDebug()
	}

	var resp AsyncSearchResponse
	err := b.client.DoWithHeaderAndDecode(ctx, http.MethodGet, path, nil, b.getHeaders(), &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
