package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
)

// SearchBuilder 搜索构建器
type SearchBuilder struct {
	client    ESClient
	index     string
	query     map[string]any
	from      int
	size      int
	sort      []map[string]any
	aggs      map[string]any
	source    []string
	highlight map[string]any
	minScore  *float64 // 最小评分
	BoolQuery[SearchBuilder]
	DebugHelper
}

// NewSearchBuilder 创建搜索构建器
func NewSearchBuilder(c ESClient, index string) *SearchBuilder {
	b := &SearchBuilder{
		client:      c,
		index:       index,
		size:        10,
		aggs:        make(map[string]any),
		DebugHelper: DebugHelper{logger: c.GetLogger()},
	}
	b.initBoolQuery(b)
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
		if len(temp.must) > 0 {
			b.should = append(b.should, temp.must...)
		}
		if len(temp.filters) > 0 {
			b.should = append(b.should, temp.filters...)
		}
	}
	return b
}

// ========== 分页和排序 ==========

func (b *SearchBuilder) From(from int) *SearchBuilder {
	b.from = from
	return b
}

// Size 设置返回结果数量
func (b *SearchBuilder) Size(size int) *SearchBuilder {
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

// SearchResponse 搜索响应
type SearchResponse struct {
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
		MaxScore float64 `json:"max_score"`
		Hits     []struct {
			Index     string              `json:"_index"`
			ID        string              `json:"_id"`
			Score     float64             `json:"_score"`
			Source    map[string]any      `json:"_source"`
			Highlight map[string][]string `json:"highlight,omitempty"`
		} `json:"hits"`
	} `json:"hits"`
	Aggregations map[string]any `json:"aggregations,omitempty"`
}

// Scan 将搜索结果扫描到结构体切片中
// dest 必须是指向切片的指针，如 *[]Article
func (r *SearchResponse) Scan(dest any) error {
	// 获取目标类型
	destVal := reflect.ValueOf(dest)
	if destVal.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be a pointer to slice")
	}
	sliceVal := destVal.Elem()
	if sliceVal.Kind() != reflect.Slice {
		return fmt.Errorf("dest must be a pointer to slice")
	}

	elemType := sliceVal.Type().Elem()

	// 遍历搜索结果
	for _, hit := range r.Hits.Hits {
		if hit.Source == nil {
			continue
		}
		// 将 Source 转为 JSON 再解析到结构体
		jsonData, err := json.Marshal(hit.Source)
		if err != nil {
			return err
		}

		// 创建新元素
		var elem reflect.Value
		if elemType.Kind() == reflect.Ptr {
			elem = reflect.New(elemType.Elem())
		} else {
			elem = reflect.New(elemType)
		}

		if err := json.Unmarshal(jsonData, elem.Interface()); err != nil {
			return err
		}

		// 添加到切片
		if elemType.Kind() == reflect.Ptr {
			sliceVal.Set(reflect.Append(sliceVal, elem))
		} else {
			sliceVal.Set(reflect.Append(sliceVal, elem.Elem()))
		}
	}

	return nil
}

// Build 构建查询 DSL
func (b *SearchBuilder) Build() map[string]any {
	body := make(map[string]any)

	// 构建 bool 查询
	if boolQ := b.BuildBoolQuery(); boolQ != nil {
		body["query"] = boolQ
	}

	// 最小评分
	if b.minScore != nil {
		body["min_score"] = *b.minScore
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
	b.SetDebug(true)
	return b
}

// Do 执行搜索
func (b *SearchBuilder) Do(ctx context.Context) (*SearchResponse, error) {
	path := fmt.Sprintf("/%s/_search", url.PathEscape(b.index))
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

	var resp SearchResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &resp, nil
}

// CountResponse 计数响应
type CountResponse struct {
	Count  int `json:"count"`
	Shards struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Skipped    int `json:"skipped"`
		Failed     int `json:"failed"`
	} `json:"_shards"`
}

// Count 执行计数查询（只返回匹配文档数量，不返回文档内容）
func (b *SearchBuilder) Count(ctx context.Context) (int64, error) {
	path := fmt.Sprintf("/%s/_count", url.PathEscape(b.index))

	// 构建查询条件（不需要分页、排序等）
	body := make(map[string]any)
	if boolQ := b.BuildBoolQuery(); boolQ != nil {
		body["query"] = boolQ
	}

	// 如果启用调试模式，打印请求信息
	if b.IsDebug() {
		b.PrintDebug("POST", path, body)
		defer b.SetDebug(false)
	}

	respBody, err := b.client.Do(ctx, http.MethodPost, path, body)
	if err != nil {
		return 0, err
	}

	// 如果启用调试模式，打印响应信息
	if b.IsDebug() {
		b.PrintResponse(respBody)
	}

	var resp CountResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return 0, fmt.Errorf("解析响应失败: %w", err)
	}

	return int64(resp.Count), nil
}
