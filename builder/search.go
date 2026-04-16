package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// SearchBuilder 搜索构建器
type SearchBuilder struct {
	client         ESClient
	index          string
	query          map[string]any
	from           int
	size           int
	sort           []map[string]any
	aggs           map[string]any
	source         []string
	highlight      map[string]any
	minScore       *float64 // 最小评分
	knn            map[string]any // KNN 向量查询 (ES 8+)
	pit            map[string]any // Point In Time (ES 7.10+)
	trackTotalHits any // 是否跟踪总命中数 (bool 或 int)
	collapse       map[string]any // Field Collapsing 字段折叠
	suggest        map[string]any // Suggest 建议查询
	BoolQuery[SearchBuilder]
	debugHelper
	baseBuilder
}

// NewSearchBuilder 创建搜索构建器
func NewSearchBuilder(c ESClient, index string) *SearchBuilder {
	b := &SearchBuilder{
		client:      c,
		index:       index,
		size:        10,
		aggs:        make(map[string]any),
		debugHelper: debugHelper{logger: c.GetLogger()},
	}
	b.initBoolQuery(b)
	b.initBaseBuilder()
	return b
}

// Header 设置自定义 Header (链式调用)
func (b *SearchBuilder) Header(key, value string) *SearchBuilder {
	b.baseBuilder.Header(key, value)
	return b
}

// Wildcard 添加通配符查询
func (b *SearchBuilder) Wildcard(field string, value string) *SearchBuilder {
	b.must = append(b.must, map[string]any{
		"wildcard": map[string]any{
			field: value,
		},
	})
	return b
}

// Prefix 添加前缀查询
func (b *SearchBuilder) Prefix(field string, value string) *SearchBuilder {
	b.must = append(b.must, map[string]any{
		"prefix": map[string]any{
			field: value,
		},
	})
	return b
}

// Regexp 添加正则表达式查询
func (b *SearchBuilder) Regexp(field string, value string) *SearchBuilder {
	b.must = append(b.must, map[string]any{
		"regexp": map[string]any{
			field: value,
		},
	})
	return b
}

// Fuzzy 添加模糊查询
func (b *SearchBuilder) Fuzzy(field string, value string, fuzziness any) *SearchBuilder {
	fuzzyQuery := map[string]any{
		"value": value,
	}
	if fuzziness != nil {
		fuzzyQuery["fuzziness"] = fuzziness
	}
	b.must = append(b.must, map[string]any{
		"fuzzy": map[string]any{
			field: fuzzyQuery,
		},
	})
	return b
}

// MatchAll 匹配所有文档
func (b *SearchBuilder) MatchAll() *SearchBuilder {
	b.query = map[string]any{
		"match_all": map[string]any{},
	}
	return b
}

// MultiMatch 多字段匹配
func (b *SearchBuilder) MultiMatch(query string, fields ...string) *SearchBuilder {
	b.must = append(b.must, map[string]any{
		"multi_match": map[string]any{
			"query":  query,
			"fields": fields,
		},
	})
	return b
}

// QueryString 查询字符串
func (b *SearchBuilder) QueryString(query string, fields ...string) *SearchBuilder {
	qs := map[string]any{
		"query": query,
	}
	if len(fields) > 0 {
		qs["fields"] = fields
	}
	b.must = append(b.must, map[string]any{
		"query_string": qs,
	})
	return b
}

// IDs 按 ID 查询
func (b *SearchBuilder) IDs(ids ...string) *SearchBuilder {
	b.filters = append(b.filters, map[string]any{
		"ids": map[string]any{
			"values": ids,
		},
	})
	return b
}

// GeoDistance 地理距离查询
func (b *SearchBuilder) GeoDistance(field string, lat, lon float64, distance string) *SearchBuilder {
	b.filters = append(b.filters, map[string]any{
		"geo_distance": map[string]any{
			"distance": distance,
			field: map[string]any{
				"lat": lat,
				"lon": lon,
			},
		},
	})
	return b
}

// GeoBoundingBox 地理边界框查询
func (b *SearchBuilder) GeoBoundingBox(field string, topLat, topLon, bottomLat, bottomLon float64) *SearchBuilder {
	b.filters = append(b.filters, map[string]any{
		"geo_bounding_box": map[string]any{
			field: map[string]any{
				"top_left": map[string]any{
					"lat": topLat,
					"lon": topLon,
				},
				"bottom_right": map[string]any{
					"lat": bottomLat,
					"lon": bottomLon,
				},
			},
		},
	})
	return b
}

// Nested 嵌套查询
func (b *SearchBuilder) Nested(path string, query map[string]any) *SearchBuilder {
	b.must = append(b.must, map[string]any{
		"nested": map[string]any{
			"path":  path,
			"query": query,
		},
	})
	return b
}

// MinScore 设置最小评分
func (b *SearchBuilder) MinScore(score float64) *SearchBuilder {
	b.minScore = &score
	return b
}

// Should 添加 should 条件（至少匹配一个）
func (b *SearchBuilder) Should(conditions ...func(*SearchBuilder)) *SearchBuilder {
	for _, condition := range conditions {
		temp := &SearchBuilder{}
		temp.initBoolQuery(temp)
		condition(temp)

		// 构建子条件的 bool 查询
		if subBool := temp.buildBoolQuery(); subBool != nil {
			b.should = append(b.should, subBool)
		}
	}
	return b
}

// ========== 分页和排序 ==========

// From 设置返回结果的起始偏移量
func (b *SearchBuilder) From(from int) *SearchBuilder {
	if from < 0 {
		b.debugHelper.logger.Error("SearchBuilder.From: value must be >= 0")
	}
	b.from = from
	return b
}

// Size 设置返回结果数量
func (b *SearchBuilder) Size(size int) *SearchBuilder {
	if size < 0 {
		b.debugHelper.logger.Error("SearchBuilder.Size: value must be >= 0")
	}
	b.size = size
	return b
}

// Sort 添加排序
func (b *SearchBuilder) Sort(field string, order string) *SearchBuilder {
	b.sort = append(b.sort, map[string]any{
		field: map[string]any{
			"order": order,
		},
	})
	return b
}

// Source 设置返回字段
func (b *SearchBuilder) Source(fields ...string) *SearchBuilder {
	b.source = fields
	return b
}

// Agg 添加聚合
func (b *SearchBuilder) Agg(name string, aggType string, field string) *SearchBuilder {
	b.aggs[name] = map[string]any{
		aggType: map[string]any{
			"field": field,
		},
	}
	return b
}

// Highlight 添加高亮
func (b *SearchBuilder) Highlight(fields ...string) *SearchBuilder {
	highlightFields := make(map[string]any)
	for _, field := range fields {
		highlightFields[field] = map[string]any{}
	}
	b.highlight = map[string]any{
		"fields": highlightFields,
	}
	return b
}

// KNN 添加 K-Nearest Neighbor (向量) 查询 (ES 8.x / 9.x)
func (b *SearchBuilder) KNN(field string, queryVector []float32, k int, numCandidates int) *SearchBuilder {
	b.knn = map[string]any{
		"field":          field,
		"query_vector":   queryVector,
		"k":              k,
		"num_candidates": numCandidates,
	}
	return b
}

// PointInTime 添加 PIT 查询 (ES 7.10+ / 8.x / 9.x)
func (b *SearchBuilder) PointInTime(id string, keepAlive string) *SearchBuilder {
	b.pit = map[string]any{
		"id":         id,
		"keep_alive": keepAlive,
	}
	return b
}

// TrackTotalHits 设置是否跟踪总命中数 (用于突破默认 10000 条限制，可传 bool 或 int)
func (b *SearchBuilder) TrackTotalHits(track any) *SearchBuilder {
	b.trackTotalHits = track
	return b
}

// Collapse 添加字段折叠 (Field Collapsing)
func (b *SearchBuilder) Collapse(field string, innerHits ...map[string]any) *SearchBuilder {
	collapseParams := map[string]any{
		"field": field,
	}
	if len(innerHits) > 0 {
		collapseParams["inner_hits"] = innerHits[0]
	}
	b.collapse = collapseParams
	return b
}

// Suggest 添加搜索建议
func (b *SearchBuilder) Suggest(name string, text string, suggester map[string]any) *SearchBuilder {
	if b.suggest == nil {
		b.suggest = make(map[string]any)
	}
	b.suggest[name] = map[string]any{
		"text": text,
	}
	for k, v := range suggester {
		b.suggest[name].(map[string]any)[k] = v
	}
	return b
}

// SearchResponse 搜索响应
type SearchResponse struct {
	Took     int        `json:"took"`
	TimedOut bool       `json:"timed_out"`
	Shards   ShardsInfo `json:"_shards"`
	Hits     struct {
		Total    HitsTotal `json:"total"`
		MaxScore float64   `json:"max_score"`
		Hits     []HitItem `json:"hits"`
	} `json:"hits"`
	Aggregations map[string]any `json:"aggregations,omitempty"`
	Suggest      map[string]any `json:"suggest,omitempty"`
	PitID        string         `json:"pit_id,omitempty"` // PIT 标识
}

// Total 返回总命中数
func (r *SearchResponse) Total() int {
	return r.Hits.Total.Value
}

// TotalIsExact 返回总命中数是否为准确值（而非估值）
func (r *SearchResponse) TotalIsExact() bool {
	return r.Hits.Total.Relation == "eq"
}

// Scan 将搜索结果扫描到结构体切片中
// dest 必须是指向切片的指针，如 *[]Article
func (r *SearchResponse) Scan(dest any) error {
	return scanHits(r.Hits.Hits, dest)
}

// Build 构建查询 DSL
func (b *SearchBuilder) Build() map[string]any {
	body := make(map[string]any)

	// 构建查询：优先使用 bool 查询，否则使用直接设置的 query（如 match_all）
	if boolQ := b.buildBoolQuery(); boolQ != nil {
		body["query"] = boolQ
	} else if b.query != nil {
		body["query"] = b.query
	}

	// 最小评分
	if b.minScore != nil {
		body["min_score"] = *b.minScore
	}

	// 跟踪总命中数
	if b.trackTotalHits != nil {
		body["track_total_hits"] = b.trackTotalHits
	}

	// KNN 搜索 (ES 8.x / 9.x)
	if b.knn != nil {
		body["knn"] = b.knn
	}

	// PIT (ES 7.10+ / 8.x / 9.x)
	if b.pit != nil {
		body["pit"] = b.pit
	}

	// 字段折叠
	if b.collapse != nil {
		body["collapse"] = b.collapse
	}

	// 搜索建议
	if b.suggest != nil {
		body["suggest"] = b.suggest
	}

	// 分页
	body["from"] = b.from
	body["size"] = b.size

	// 排序
	if len(b.sort) > 0 {
		body["sort"] = b.sort
	}

	// 返回字段
	if len(b.source) > 0 {
		body["_source"] = b.source
	}

	// 聚合
	if len(b.aggs) > 0 {
		body["aggs"] = b.aggs
	}

	// 高亮
	if b.highlight != nil {
		body["highlight"] = b.highlight
	}

	return body
}

// Debug 启用调试模式（链式调用）
func (b *SearchBuilder) Debug() *SearchBuilder {
	b.setDebug(true)
	return b
}

// Do 执行搜索
func (b *SearchBuilder) Do(ctx context.Context) (*SearchResponse, error) {
	// 如果使用了 PIT，路径通常是不带 index 的 `/_search`，或者 ES 会忽略 URL 上的 index
	path := fmt.Sprintf("/%s/_search", url.PathEscape(b.index))
	if b.index == "" || b.pit != nil {
		path = "/_search"
	}
	body := b.Build()

	// 如果启用调试模式，打印请求信息
	if b.isDebug() {
		b.printDebug("POST", path, body)
		defer b.setDebug(false)
	}

	respBody, err := b.client.DoWithHeader(ctx, http.MethodPost, path, body, b.getHeaders())
	if err != nil {
		return nil, err
	}

	// 如果启用调试模式，打印响应信息
	if b.isDebug() {
		b.printResponse(respBody)
	}

	var resp SearchResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &resp, nil
}

// CountResponse 计数响应
type CountResponse struct {
	Count  int        `json:"count"`
	Shards ShardsInfo `json:"_shards"`
}

// Count 执行计数查询（只返回匹配文档数量，不返回文档内容）
func (b *SearchBuilder) Count(ctx context.Context) (int64, error) {
	path := fmt.Sprintf("/%s/_count", url.PathEscape(b.index))

	// 构建查询条件（不需要分页、排序等）
	body := make(map[string]any)
	if boolQ := b.buildBoolQuery(); boolQ != nil {
		body["query"] = boolQ
	}

	// 如果启用调试模式，打印请求信息
	if b.isDebug() {
		b.printDebug("POST", path, body)
		defer b.setDebug(false)
	}

	respBody, err := b.client.Do(ctx, http.MethodPost, path, body)
	if err != nil {
		return 0, err
	}

	// 如果启用调试模式，打印响应信息
	if b.isDebug() {
		b.printResponse(respBody)
	}

	var resp CountResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return 0, fmt.Errorf("解析响应失败: %w", err)
	}

	return int64(resp.Count), nil
}
