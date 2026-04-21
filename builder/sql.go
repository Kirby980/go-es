package builder

import (
	"context"
	"net/http"
)

// SQLBuilder 是用于构建和执行 Elasticsearch SQL 操作的链式 API 构建器。
type SQLBuilder struct {
	client    ESClient
	query     string
	fetchSize int
	debugHelper
	baseBuilder
}

// NewSQLBuilder 创建并返回一个 SQL 构建器实例，用于构造和执行 Elasticsearch 的 SQL 请求。
func NewSQLBuilder(client ESClient) *SQLBuilder {
	return &SQLBuilder{
		client: client,
	}
}

// Query 设置核心的查询 DSL 条件。
func (b *SQLBuilder) Query(query string) *SQLBuilder {
	b.query = query
	return b
}

// FetchSize 设置每次游标抓取返回的结果集大小。
func (b *SQLBuilder) FetchSize(size int) *SQLBuilder {
	b.fetchSize = size
	return b
}

// Debug 开启当前请求的单次调试模式。开启后，请求的完整 HTTP URL、Method、Body 及响应将被打印到日志中，便于排查问题。
func (b *SQLBuilder) Debug(enable bool) *SQLBuilder {
	b.setDebug(enable)
	return b
}

// SQLResponse 定义了 Elasticsearch SQL 操作返回的结构化响应数据，包含元数据及核心结果。
type SQLResponse struct {
	Columns []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"columns"`
	Rows   [][]any `json:"rows"`
	Cursor string  `json:"cursor,omitempty"`
}

// Do 立即执行当前构建好的请求，并返回结构化的 Elasticsearch 响应结果或错误。请确保在调用此方法前已设置好所有必要参数。
func (b *SQLBuilder) Do(ctx context.Context) (*SQLResponse, error) {
	path := "/_sql?format=json"

	body := map[string]any{
		"query": b.query,
	}
	if b.fetchSize > 0 {
		body["fetch_size"] = b.fetchSize
	}

	if b.isDebug() {
		b.printDebug("POST", path, body)
		defer b.autoResetDebug()
	}

	var resp SQLResponse
	err := b.client.DoWithHeaderAndDecode(ctx, http.MethodPost, path, body, b.getHeaders(), &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
