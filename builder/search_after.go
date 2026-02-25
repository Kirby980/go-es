package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// SearchAfterBuilder Search After深度分页构建器
// Search After 是 Elasticsearch 提供的高效分页方案，相比 Scroll 更轻量，无需维护上下文
type SearchAfterBuilder struct {
	client       ESClient
	index        string
	size         int
	sort         []map[string]any
	searchAfter  []any // 上一页最后一个文档的 sort 值
	source       []string
	highlight    map[string]any
	minScore     *float64
	lastResponse *SearchAfterResponse // 保存上次响应用于自动获取下一页
	BoolQuery[SearchAfterBuilder]
	debugHelper
}

// NewSearchAfterBuilder 创建SearchAfter构建器
func NewSearchAfterBuilder(c ESClient, index string) *SearchAfterBuilder {
	b := &SearchAfterBuilder{
		client:      c,
		index:       index,
		size:        10,
		sort:        make([]map[string]any, 0),
		highlight:   make(map[string]any),
		debugHelper: debugHelper{logger: c.GetLogger()},
	}
	b.initBoolQuery(b)
	return b
}

// Size 设置每页返回的文档数量
func (b *SearchAfterBuilder) Size(size int) *SearchAfterBuilder {
	b.size = size
	return b
}

// Sort 添加排序字段
// 注意：Search After 必须至少有一个排序字段，建议最后加上 _id 作为 tie-breaker
func (b *SearchAfterBuilder) Sort(field string, order string) *SearchAfterBuilder {
	b.sort = append(b.sort, map[string]any{
		field: order,
	})
	return b
}

// SortBy 使用复杂排序选项
func (b *SearchAfterBuilder) SortBy(field string, options map[string]any) *SearchAfterBuilder {
	b.sort = append(b.sort, map[string]any{
		field: options,
	})
	return b
}

// SearchAfter 手动设置 search_after 值（上一页最后一个文档的 sort 值）
func (b *SearchAfterBuilder) SearchAfter(values ...any) *SearchAfterBuilder {
	b.searchAfter = values
	return b
}

// Source 指定返回的字段
func (b *SearchAfterBuilder) Source(fields ...string) *SearchAfterBuilder {
	b.source = fields
	return b
}

// Highlight 添加高亮字段
func (b *SearchAfterBuilder) Highlight(fields ...string) *SearchAfterBuilder {
	highlightFields := make(map[string]any)
	for _, field := range fields {
		highlightFields[field] = map[string]any{}
	}
	b.highlight = map[string]any{
		"fields": highlightFields,
	}
	return b
}

// MinScore 设置最小评分
func (b *SearchAfterBuilder) MinScore(score float64) *SearchAfterBuilder {
	b.minScore = &score
	return b
}

// Debug 启用调试模式
func (b *SearchAfterBuilder) Debug() *SearchAfterBuilder {
	b.setDebug(true)
	return b
}

// Build 构建查询体
func (b *SearchAfterBuilder) Build() map[string]any {
	body := make(map[string]any)

	// 构建查询条件
	if boolQ := b.buildBoolQuery(); boolQ != nil {
		body["query"] = boolQ
	}

	// 添加 size
	body["size"] = b.size

	// 添加排序（Search After 必须有排序）
	if len(b.sort) > 0 {
		body["sort"] = b.sort
	} else {
		// 默认按 _id 排序（如果用户没指定）
		body["sort"] = []map[string]any{
			{"_id": "asc"},
		}
	}

	// 添加 search_after
	if len(b.searchAfter) > 0 {
		body["search_after"] = b.searchAfter
	}

	// 添加 _source
	if len(b.source) > 0 {
		body["_source"] = b.source
	}

	// 添加 highlight
	if len(b.highlight) > 0 {
		body["highlight"] = b.highlight
	}

	// 添加 min_score
	if b.minScore != nil {
		body["min_score"] = *b.minScore
	}

	return body
}

// SearchAfterResponse Search After响应
type SearchAfterResponse struct {
	Took     int  `json:"took"`
	TimedOut bool `json:"timed_out"`
	Shards   struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Skipped    int `json:"skipped"`
		Failed     int `json:"failed"`
	} `json:"_shards"`
	Hits struct {
		Total struct {
			Value    int    `json:"value"`
			Relation string `json:"relation"`
		} `json:"total"`
		MaxScore *float64 `json:"max_score"`
		Hits     []struct {
			Index     string              `json:"_index"`
			ID        string              `json:"_id"`
			Score     float64             `json:"_score"`
			Source    map[string]any      `json:"_source"`
			Sort      []any               `json:"sort"` // Search After 的关键：每个文档的排序值
			Highlight map[string][]string `json:"highlight,omitempty"`
		} `json:"hits"`
	} `json:"hits"`
}

// Do 执行查询
func (b *SearchAfterBuilder) Do(ctx context.Context) (*SearchAfterResponse, error) {
	path := fmt.Sprintf("/%s/_search", url.PathEscape(b.index))
	body := b.Build()

	// 如果启用调试模式，打印请求信息
	if b.isDebug() {
		b.printDebug("POST", path, body)
		defer b.setDebug(false)
	}

	respBody, err := b.client.Do(ctx, http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}

	// 如果启用调试模式，打印响应信息
	if b.isDebug() {
		b.printResponse(respBody)
	}

	var resp SearchAfterResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 保存响应供 Next() 使用
	b.lastResponse = &resp

	return &resp, nil
}

// Next 获取下一页数据（自动使用上一次响应的最后一个文档的 sort 值）
func (b *SearchAfterBuilder) Next(ctx context.Context) (*SearchAfterResponse, error) {
	if b.lastResponse == nil {
		return nil, fmt.Errorf("请先调用 Do() 方法初始化查询")
	}

	// 检查是否还有数据
	if len(b.lastResponse.Hits.Hits) == 0 {
		return nil, fmt.Errorf("已经没有更多数据")
	}

	// 获取最后一个文档的 sort 值
	lastHit := b.lastResponse.Hits.Hits[len(b.lastResponse.Hits.Hits)-1]
	if len(lastHit.Sort) == 0 {
		return nil, fmt.Errorf("响应中没有 sort 字段，请确保查询包含排序")
	}

	// 设置 search_after
	b.searchAfter = lastHit.Sort

	// 执行下一页查询
	return b.Do(ctx)
}

// HasMore 判断是否还有更多数据
func (b *SearchAfterBuilder) HasMore(resp *SearchAfterResponse) bool {
	return len(resp.Hits.Hits) > 0
}

// GetLastSortValues 获取最后一个文档的 sort 值（用于手动分页）
func (b *SearchAfterBuilder) GetLastSortValues(resp *SearchAfterResponse) []any {
	if len(resp.Hits.Hits) == 0 {
		return nil
	}
	lastHit := resp.Hits.Hits[len(resp.Hits.Hits)-1]
	return lastHit.Sort
}
