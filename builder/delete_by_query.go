package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// DeleteByQueryBuilder 按查询删除构建器
type DeleteByQueryBuilder struct {
	client ESClient
	index  string
	BoolQuery[DeleteByQueryBuilder]
	DebugHelper
}

// NewDeleteByQueryBuilder 创建按查询删除构建器
func NewDeleteByQueryBuilder(c ESClient, index string) *DeleteByQueryBuilder {
	b := &DeleteByQueryBuilder{
		client:      c,
		index:       index,
		DebugHelper: DebugHelper{logger: c.GetLogger()},
	}
	b.initBoolQuery(b)
	return b
}

// Debug 启用调试模式
func (b *DeleteByQueryBuilder) Debug() *DeleteByQueryBuilder {
	b.SetDebug(true)
	return b
}

// Build 构建请求体
func (b *DeleteByQueryBuilder) Build() map[string]any {
	body := make(map[string]any)

	// 构建查询条件
	if boolQ := b.BuildBoolQuery(); boolQ != nil {
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
	if b.IsDebug() {
		b.PrintDebug("POST", path, body)
		defer b.SetDebug(false)
	}

	respBody, err := b.client.Do(ctx, http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}

	// 如果启用调试模式，打印响应信息
	if b.IsDebug() {
		b.PrintResponse(respBody)
	}

	var resp DeleteByQueryResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &resp, nil
}
