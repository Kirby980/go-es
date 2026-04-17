package builder

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// UpdateByQueryBuilder 按查询更新构建器
type UpdateByQueryBuilder struct {
	client ESClient
	index  string
	script map[string]any
	BoolQuery[UpdateByQueryBuilder]
	debugHelper
	baseBuilder
}

// NewUpdateByQueryBuilder 创建按查询更新构建器
func NewUpdateByQueryBuilder(c ESClient, index string) *UpdateByQueryBuilder {
	b := &UpdateByQueryBuilder{
		client:      c,
		index:       index,
		debugHelper: debugHelper{logger: c.GetLogger()},
	}
	b.initBoolQuery(b)
	b.initBaseBuilder()
	return b
}

// Header 设置自定义 Header (链式调用)
func (b *UpdateByQueryBuilder) Header(key, value string) *UpdateByQueryBuilder {
	b.baseBuilder.Header(key, value)
	return b
}

// Script 设置更新脚本
func (b *UpdateByQueryBuilder) Script(source string, params map[string]any) *UpdateByQueryBuilder {
	b.script = map[string]any{
		"source": source,
		"lang":   "painless",
	}
	if params != nil {
		b.script["params"] = params
	}
	return b
}

// Set 设置字段值（简化的脚本更新）
func (b *UpdateByQueryBuilder) Set(field string, value any) *UpdateByQueryBuilder {
	// 如果已有脚本，追加
	if b.script != nil {
		existingSource := b.script["source"].(string)
		b.script["source"] = existingSource + fmt.Sprintf("; ctx._source.%s = params.%s", field, field)
	} else {
		b.script = map[string]any{
			"source": fmt.Sprintf("ctx._source.%s = params.%s", field, field),
			"lang":   "painless",
		}
	}

	// 添加参数
	if b.script["params"] == nil {
		b.script["params"] = make(map[string]any)
	}
	b.script["params"].(map[string]any)[field] = value

	return b
}

// Debug 启用调试模式
func (b *UpdateByQueryBuilder) Debug() *UpdateByQueryBuilder {
	b.setDebug(true)
	return b
}

// Build 构建请求体
func (b *UpdateByQueryBuilder) Build() map[string]any {
	body := make(map[string]any)

	// 构建查询条件
	if boolQ := b.buildBoolQuery(); boolQ != nil {
		body["query"] = boolQ
	}

	// 添加脚本
	if b.script != nil {
		body["script"] = b.script
	}

	return body
}

// UpdateByQueryResponse 更新响应
type UpdateByQueryResponse struct {
	Took             int  `json:"took"`
	TimedOut         bool `json:"timed_out"`
	Total            int  `json:"total"`
	Updated          int  `json:"updated"`
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

// Do 执行更新
func (b *UpdateByQueryBuilder) Do(ctx context.Context) (*UpdateByQueryResponse, error) {
	if b.script == nil {
		return nil, fmt.Errorf("必须设置更新脚本")
	}

	path := fmt.Sprintf("/%s/_update_by_query", url.PathEscape(b.index))
	body := b.Build()

	// 如果启用调试模式，打印请求信息
	if b.isDebug() {
		b.printDebug("POST", path, body)
		defer b.autoResetDebug()
	}

	var resp UpdateByQueryResponse
	if err := b.client.DoWithHeaderAndDecode(ctx, http.MethodPost, path, body, b.getHeaders(), &resp); err != nil {
		return nil, err
	}

	if b.isDebug() {
		b.printResponseObj(resp)
	}

	return &resp, nil
}
