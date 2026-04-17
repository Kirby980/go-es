package builder

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// DeleteByQueryBuilder 按查询删除构建器
type DeleteByQueryBuilder struct {
	client ESClient
	index  string
	BoolQuery[DeleteByQueryBuilder]
	debugHelper
	baseBuilder
}

// NewDeleteByQueryBuilder 创建按查询删除构建器
func NewDeleteByQueryBuilder(c ESClient, index string) *DeleteByQueryBuilder {
	b := &DeleteByQueryBuilder{
		client:      c,
		index:       index,
		debugHelper: debugHelper{logger: c.GetLogger()},
	}
	b.initBoolQuery(b)
	b.initBaseBuilder()
	return b
}

// Header 设置自定义 Header (链式调用)
func (b *DeleteByQueryBuilder) Header(key, value string) *DeleteByQueryBuilder {
	b.baseBuilder.Header(key, value)
	return b
}

// Debug 启用调试模式
func (b *DeleteByQueryBuilder) Debug() *DeleteByQueryBuilder {
	b.setDebug(true)
	return b
}

// Build 构建请求体
func (b *DeleteByQueryBuilder) Build() map[string]any {
	body := make(map[string]any)

	// 构建查询条件
	if boolQ := b.buildBoolQuery(); boolQ != nil {
		body["query"] = boolQ
	}

	return body
}

// DeleteByQueryResponse 删除响应
type DeleteByQueryResponse struct {
	Took             int  `json:"took"`
	TimedOut         bool `json:"timed_out"`
	Total            int  `json:"total"`
	Deleted          int  `json:"deleted"`
	Batches          int  `json:"batches"`
	VersionConflicts int  `json:"version_conflicts"`
	Noops            int  `json:"noops"`
	Retries          struct {
		Bulk   int `json:"bulk"`
		Search int `json:"search"`
	} `json:"retries"`
	ThrottledMillis      int              `json:"throttled_millis"`
	RequestsPerSecond    float64          `json:"requests_per_second"`
	ThrottledUntilMillis int              `json:"throttled_until_millis"`
	Failures             []map[string]any `json:"failures"`
}

// Do 执行删除
func (b *DeleteByQueryBuilder) Do(ctx context.Context) (*DeleteByQueryResponse, error) {
	path := fmt.Sprintf("/%s/_delete_by_query", url.PathEscape(b.index))
	body := b.Build()

	// 检查是否有查询条件
	if len(body) == 0 {
		return nil, fmt.Errorf("必须设置查询条件，避免误删除所有数据")
	}

	// 如果启用调试模式，打印请求信息
	if b.isDebug() {
		b.printDebug("POST", path, body)
		defer b.autoResetDebug()
	}

	var resp DeleteByQueryResponse
	if err := b.client.DoWithHeaderAndDecode(ctx, http.MethodPost, path, body, b.getHeaders(), &resp); err != nil {
		return nil, err
	}

	if b.isDebug() {
		b.printResponseObj(resp)
	}

	return &resp, nil
}
