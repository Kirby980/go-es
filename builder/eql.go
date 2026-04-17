package builder

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// EQLBuilder EQL (Event Query Language) 构建器 (ES 7.9+)
// 常用于安全日志和时间序列分析
type EQLBuilder struct {
	client     ESClient
	index      string
	query      string // EQL 语句
	size       *int
	tiebreaker *int
	fetchSize  *int
	filter     map[string]any // 额外的前置过滤 DSL
	debugHelper
	baseBuilder
}

// NewEQLBuilder 创建 EQL 构建器
func NewEQLBuilder(c ESClient, index string) *EQLBuilder {
	b := &EQLBuilder{
		client:      c,
		index:       index,
		debugHelper: debugHelper{logger: c.GetLogger()},
	}
	b.initBaseBuilder()
	return b
}

// Header 设置自定义 Header
func (b *EQLBuilder) Header(key, value string) *EQLBuilder {
	b.baseBuilder.Header(key, value)
	return b
}

// Query 设置 EQL 语句
func (b *EQLBuilder) Query(query string) *EQLBuilder {
	b.query = query
	return b
}

// Size 设置返回的序列或事件数量
func (b *EQLBuilder) Size(size int) *EQLBuilder {
	b.size = &size
	return b
}

// Filter 设置前置的普通 ES Query DSL 过滤器，用于在 EQL 执行前缩小数据范围
func (b *EQLBuilder) Filter(filter map[string]any) *EQLBuilder {
	b.filter = filter
	return b
}

// Debug 启用调试模式
func (b *EQLBuilder) Debug() *EQLBuilder {
	b.setDebug(true)
	return b
}

// Build 构建请求体
func (b *EQLBuilder) Build() map[string]any {
	body := map[string]any{
		"query": b.query,
	}

	if b.size != nil {
		body["size"] = *b.size
	}
	if b.filter != nil {
		body["filter"] = b.filter
	}

	return body
}

// EQLResponse EQL 查询响应
type EQLResponse struct {
	IsPartial bool `json:"is_partial"`
	IsRunning bool `json:"is_running"`
	Took      int  `json:"took"`
	TimedOut  bool `json:"timed_out"`
	Hits      struct {
		Total  HitsTotal `json:"total"`
		Events []struct {
			Index  string         `json:"_index"`
			ID     string         `json:"_id"`
			Source map[string]any `json:"_source"`
		} `json:"events,omitempty"`
		Sequences []struct {
			JoinKeys []any `json:"join_keys"`
			Events   []struct {
				Index  string         `json:"_index"`
				ID     string         `json:"_id"`
				Source map[string]any `json:"_source"`
			} `json:"events"`
		} `json:"sequences,omitempty"`
	} `json:"hits"`
}

// Do 执行 EQL 查询
func (b *EQLBuilder) Do(ctx context.Context) (*EQLResponse, error) {
	if b.query == "" {
		return nil, fmt.Errorf("EQL query cannot be empty")
	}

	path := fmt.Sprintf("/%s/_eql/search", url.PathEscape(b.index))
	body := b.Build()

	if b.isDebug() {
		b.printDebug("POST", path, body)
		defer b.autoResetDebug()
	}

	var resp EQLResponse
	if err := b.client.DoWithHeaderAndDecode(ctx, http.MethodPost, path, body, b.getHeaders(), &resp); err != nil {
		return nil, err
	}

	if b.isDebug() {
		b.printResponseObj(resp)
	}

	return &resp, nil
}
