package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// SearchAfterBuilder Search After深度分页构建器
// Search After 是 Elasticsearch 提供的高效分页方案，相比 Scroll 更轻量，无需维护上下文
type SearchAfterBuilder struct {
	client             ESClient
	index              string
	filters            []map[string]any
	must               []map[string]any
	should             []map[string]any
	mustNot            []map[string]any
	minimumShouldMatch any // 最少匹配 should 条件数量
	size               int
	sort               []map[string]any
	searchAfter        []any // 上一页最后一个文档的 sort 值
	source             []string
	highlight          map[string]any
	minScore           *float64
	lastResponse       *SearchAfterResponse // 保存上次响应用于自动获取下一页
	DebugHelper
}

// NewSearchAfterBuilder 创建SearchAfter构建器
func NewSearchAfterBuilder(c ESClient, index string) *SearchAfterBuilder {
	return &SearchAfterBuilder{
		client:    c,
		index:     index,
		filters:   make([]map[string]any, 0),
		must:      make([]map[string]any, 0),
		should:    make([]map[string]any, 0),
		mustNot:   make([]map[string]any, 0),
		size:      10,
		sort:      make([]map[string]any, 0),
		highlight: make(map[string]any),
	}
}

// Match 添加 match 查询条件
func (b *SearchAfterBuilder) Match(field string, value any) *SearchAfterBuilder {
	b.must = append(b.must, map[string]any{
		"match": map[string]any{
			field: value,
		},
	})
	return b
}

// MatchPhrase 添加 match_phrase 查询
func (b *SearchAfterBuilder) MatchPhrase(field string, value any) *SearchAfterBuilder {
	b.must = append(b.must, map[string]any{
		"match_phrase": map[string]any{
			field: value,
		},
	})
	return b
}

// Term 添加 term 查询条件
func (b *SearchAfterBuilder) Term(field string, value any) *SearchAfterBuilder {
	b.filters = append(b.filters, map[string]any{
		"term": map[string]any{
			field: value,
		},
	})
	return b
}

// Terms 添加 terms 查询
func (b *SearchAfterBuilder) Terms(field string, values ...any) *SearchAfterBuilder {
	b.filters = append(b.filters, map[string]any{
		"terms": map[string]any{
			field: values,
		},
	})
	return b
}

// Range 添加范围查询条件
func (b *SearchAfterBuilder) Range(field string, gte, lte any) *SearchAfterBuilder {
	rangeQuery := make(map[string]any)
	if gte != nil {
		rangeQuery["gte"] = gte
	}
	if lte != nil {
		rangeQuery["lte"] = lte
	}
	b.filters = append(b.filters, map[string]any{
		"range": map[string]any{
			field: rangeQuery,
		},
	})
	return b
}

// Exists 添加字段存在性查询
func (b *SearchAfterBuilder) Exists(field string) *SearchAfterBuilder {
	b.filters = append(b.filters, map[string]any{
		"exists": map[string]any{
			"field": field,
		},
	})
	return b
}

// MatchShould 添加 match should 条件（OR关系）
func (b *SearchAfterBuilder) MatchShould(field string, value any) *SearchAfterBuilder {
	b.should = append(b.should, map[string]any{
		"match": map[string]any{
			field: value,
		},
	})
	return b
}

// TermShould 添加 term should 条件（OR关系）
func (b *SearchAfterBuilder) TermShould(field string, value any) *SearchAfterBuilder {
	b.should = append(b.should, map[string]any{
		"term": map[string]any{
			field: value,
		},
	})
	return b
}

// MatchMustNot 添加 match must_not 条件（NOT关系）
func (b *SearchAfterBuilder) MatchMustNot(field string, value any) *SearchAfterBuilder {
	b.mustNot = append(b.mustNot, map[string]any{
		"match": map[string]any{
			field: value,
		},
	})
	return b
}

// MinimumShouldMatch 设置最少匹配的 should 条件数量
func (b *SearchAfterBuilder) MinimumShouldMatch(value any) *SearchAfterBuilder {
	b.minimumShouldMatch = value
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
	b.SetDebug(true)
	return b
}

// Build 构建查询体
func (b *SearchAfterBuilder) Build() map[string]any {
	body := make(map[string]any)

	// 构建查询条件
	if len(b.must) > 0 || len(b.filters) > 0 || len(b.should) > 0 || len(b.mustNot) > 0 {
		boolQuery := make(map[string]any)
		if len(b.must) > 0 {
			boolQuery["must"] = b.must
		}
		if len(b.filters) > 0 {
			boolQuery["filter"] = b.filters
		}
		if len(b.should) > 0 {
			boolQuery["should"] = b.should
		}
		if len(b.mustNot) > 0 {
			boolQuery["must_not"] = b.mustNot
		}
		if b.minimumShouldMatch != nil {
			boolQuery["minimum_should_match"] = b.minimumShouldMatch
		}
		body["query"] = map[string]any{
			"bool": boolQuery,
		}
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
	path := fmt.Sprintf("/%s/_search", b.index)
	body := b.Build()

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
