package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// ExplainBuilder Explain API 构建器 (单文档评分解释)
type ExplainBuilder struct {
	client ESClient
	index  string
	id     string
	query  map[string]any
	BoolQuery[ExplainBuilder]
	debugHelper
	baseBuilder
}

// NewExplainBuilder 创建 Explain 构建器
func NewExplainBuilder(c ESClient, index, id string) *ExplainBuilder {
	b := &ExplainBuilder{
		client:      c,
		index:       index,
		id:          id,
		debugHelper: debugHelper{logger: c.GetLogger()},
	}
	b.initBoolQuery(b)
	b.initBaseBuilder()
	return b
}

// Header 设置自定义 Header
func (b *ExplainBuilder) Header(key, value string) *ExplainBuilder {
	b.baseBuilder.Header(key, value)
	return b
}

// Debug 启用调试模式
func (b *ExplainBuilder) Debug() *ExplainBuilder {
	b.setDebug(true)
	return b
}

// Build 构建请求体
func (b *ExplainBuilder) Build() map[string]any {
	body := make(map[string]any)

	if boolQ := b.buildBoolQuery(); boolQ != nil {
		body["query"] = boolQ
	} else if b.query != nil {
		body["query"] = b.query
	}

	return body
}

// ExplainResponse Explain API 响应
type ExplainResponse struct {
	Index       string `json:"_index"`
	ID          string `json:"_id"`
	Matched     bool   `json:"matched"`
	Explanation struct {
		Value       float64 `json:"value"`
		Description string  `json:"description"`
		Details     []any   `json:"details"`
	} `json:"explanation"`
}

// Do 执行 Explain 请求
func (b *ExplainBuilder) Do(ctx context.Context) (*ExplainResponse, error) {
	if b.index == "" || b.id == "" {
		return nil, fmt.Errorf("index 和 id 不能为空")
	}

	path := fmt.Sprintf("/%s/_explain/%s", url.PathEscape(b.index), url.PathEscape(b.id))
	body := b.Build()

	if b.isDebug() {
		b.printDebug("POST", path, body)
		defer b.setDebug(false)
	}

	respBody, err := b.client.DoWithHeader(ctx, http.MethodPost, path, body, b.getHeaders())
	if err != nil {
		return nil, err
	}

	if b.isDebug() {
		b.printResponse(respBody)
	}

	var resp ExplainResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &resp, nil
}
