package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go-es/client"
)

// SearchBuilder 搜索构建器
type SearchBuilder struct {
	client    *client.Client
	index     string
	query     map[string]interface{}
	filters   []map[string]interface{}
	must      []map[string]interface{}
	should    []map[string]interface{}
	mustNot   []map[string]interface{}
	from      int
	size      int
	sort      []map[string]interface{}
	aggs      map[string]interface{}
	source    []string
	highlight map[string]interface{}
}

// NewSearchBuilder 创建搜索构建器
func NewSearchBuilder(c *client.Client, index string) *SearchBuilder {
	return &SearchBuilder{
		client:  c,
		index:   index,
		filters: make([]map[string]interface{}, 0),
		must:    make([]map[string]interface{}, 0),
		should:  make([]map[string]interface{}, 0),
		mustNot: make([]map[string]interface{}, 0),
		size:    10,
		aggs:    make(map[string]interface{}),
	}
}

// Match 添加 match 查询
func (b *SearchBuilder) Match(field string, value interface{}) *SearchBuilder {
	b.must = append(b.must, map[string]interface{}{
		"match": map[string]interface{}{
			field: value,
		},
	})
	return b
}

// MatchPhrase 添加 match_phrase 查询
func (b *SearchBuilder) MatchPhrase(field string, value interface{}) *SearchBuilder {
	b.must = append(b.must, map[string]interface{}{
		"match_phrase": map[string]interface{}{
			field: value,
		},
	})
	return b
}

// Term 添加 term 查询
func (b *SearchBuilder) Term(field string, value interface{}) *SearchBuilder {
	b.filters = append(b.filters, map[string]interface{}{
		"term": map[string]interface{}{
			field: value,
		},
	})
	return b
}

// Terms 添加 terms 查询
func (b *SearchBuilder) Terms(field string, values ...interface{}) *SearchBuilder {
	b.filters = append(b.filters, map[string]interface{}{
		"terms": map[string]interface{}{
			field: values,
		},
	})
	return b
}

// Range 添加范围查询
func (b *SearchBuilder) Range(field string, gte, lte interface{}) *SearchBuilder {
	rangeQuery := make(map[string]interface{})
	if gte != nil {
		rangeQuery["gte"] = gte
	}
	if lte != nil {
		rangeQuery["lte"] = lte
	}
	b.filters = append(b.filters, map[string]interface{}{
		"range": map[string]interface{}{
			field: rangeQuery,
		},
	})
	return b
}

// Exists 添加字段存在查询
func (b *SearchBuilder) Exists(field string) *SearchBuilder {
	b.filters = append(b.filters, map[string]interface{}{
		"exists": map[string]interface{}{
			"field": field,
		},
	})
	return b
}

// Wildcard 添加通配符查询
func (b *SearchBuilder) Wildcard(field string, value string) *SearchBuilder {
	b.must = append(b.must, map[string]interface{}{
		"wildcard": map[string]interface{}{
			field: value,
		},
	})
	return b
}

// Prefix 添加前缀查询
func (b *SearchBuilder) Prefix(field string, value string) *SearchBuilder {
	b.must = append(b.must, map[string]interface{}{
		"prefix": map[string]interface{}{
			field: value,
		},
	})
	return b
}

// Regexp 添加正则表达式查询
func (b *SearchBuilder) Regexp(field string, value string) *SearchBuilder {
	b.must = append(b.must, map[string]interface{}{
		"regexp": map[string]interface{}{
			field: value,
		},
	})
	return b
}

// Fuzzy 添加模糊查询
func (b *SearchBuilder) Fuzzy(field string, value string, fuzziness interface{}) *SearchBuilder {
	fuzzyQuery := map[string]interface{}{
		"value": value,
	}
	if fuzziness != nil {
		fuzzyQuery["fuzziness"] = fuzziness
	}
	b.must = append(b.must, map[string]interface{}{
		"fuzzy": map[string]interface{}{
			field: fuzzyQuery,
		},
	})
	return b
}

// MatchAll 匹配所有文档
func (b *SearchBuilder) MatchAll() *SearchBuilder {
	b.query = map[string]interface{}{
		"match_all": map[string]interface{}{},
	}
	return b
}

// MultiMatch 多字段匹配
func (b *SearchBuilder) MultiMatch(query string, fields ...string) *SearchBuilder {
	b.must = append(b.must, map[string]interface{}{
		"multi_match": map[string]interface{}{
			"query":  query,
			"fields": fields,
		},
	})
	return b
}

// QueryString 查询字符串
func (b *SearchBuilder) QueryString(query string, fields ...string) *SearchBuilder {
	qs := map[string]interface{}{
		"query": query,
	}
	if len(fields) > 0 {
		qs["fields"] = fields
	}
	b.must = append(b.must, map[string]interface{}{
		"query_string": qs,
	})
	return b
}

// IDs 按 ID 查询
func (b *SearchBuilder) IDs(ids ...string) *SearchBuilder {
	b.filters = append(b.filters, map[string]interface{}{
		"ids": map[string]interface{}{
			"values": ids,
		},
	})
	return b
}

// GeoDistance 地理距离查询
func (b *SearchBuilder) GeoDistance(field string, lat, lon float64, distance string) *SearchBuilder {
	b.filters = append(b.filters, map[string]interface{}{
		"geo_distance": map[string]interface{}{
			"distance": distance,
			field: map[string]interface{}{
				"lat": lat,
				"lon": lon,
			},
		},
	})
	return b
}

// GeoBoundingBox 地理边界框查询
func (b *SearchBuilder) GeoBoundingBox(field string, topLat, topLon, bottomLat, bottomLon float64) *SearchBuilder {
	b.filters = append(b.filters, map[string]interface{}{
		"geo_bounding_box": map[string]interface{}{
			field: map[string]interface{}{
				"top_left": map[string]interface{}{
					"lat": topLat,
					"lon": topLon,
				},
				"bottom_right": map[string]interface{}{
					"lat": bottomLat,
					"lon": bottomLon,
				},
			},
		},
	})
	return b
}

// Nested 嵌套查询
func (b *SearchBuilder) Nested(path string, query map[string]interface{}) *SearchBuilder {
	b.must = append(b.must, map[string]interface{}{
		"nested": map[string]interface{}{
			"path":  path,
			"query": query,
		},
	})
	return b
}

// MinScore 设置最小评分
func (b *SearchBuilder) MinScore(score float64) *SearchBuilder {
	// 将在 Build 时处理
	return b
}

// Should 添加 should 条件（至少匹配一个）
func (b *SearchBuilder) Should(conditions ...func(*SearchBuilder)) *SearchBuilder {
	for _, condition := range conditions {
		temp := &SearchBuilder{
			must:    make([]map[string]interface{}, 0),
			filters: make([]map[string]interface{}, 0),
		}
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

// MustNot 添加 must_not 条件
func (b *SearchBuilder) MustNot(field string, value interface{}) *SearchBuilder {
	b.mustNot = append(b.mustNot, map[string]interface{}{
		"term": map[string]interface{}{
			field: value,
		},
	})
	return b
}

// From 设置分页起始位置
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
	b.sort = append(b.sort, map[string]interface{}{
		field: map[string]interface{}{
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
	b.aggs[name] = map[string]interface{}{
		aggType: map[string]interface{}{
			"field": field,
		},
	}
	return b
}

// Highlight 添加高亮
func (b *SearchBuilder) Highlight(fields ...string) *SearchBuilder {
	highlightFields := make(map[string]interface{})
	for _, field := range fields {
		highlightFields[field] = map[string]interface{}{}
	}
	b.highlight = map[string]interface{}{
		"fields": highlightFields,
	}
	return b
}

// SearchResponse 搜索响应
type SearchResponse struct {
	Took     int  `json:"took"`
	TimedOut bool `json:"timed_out"`
	Hits     struct {
		Total struct {
			Value    int    `json:"value"`
			Relation string `json:"relation"`
		} `json:"total"`
		MaxScore float64 `json:"max_score"`
		Hits     []struct {
			Index     string                 `json:"_index"`
			ID        string                 `json:"_id"`
			Score     float64                `json:"_score"`
			Source    map[string]interface{} `json:"_source"`
			Highlight map[string][]string    `json:"highlight,omitempty"`
		} `json:"hits"`
	} `json:"hits"`
	Aggregations map[string]interface{} `json:"aggregations,omitempty"`
}

// JSON 返回 JSON 格式字符串（紧凑）
func (r *SearchResponse) JSON() string {
	data, _ := json.Marshal(r)
	return string(data)
}

// PrettyJSON 返回格式化的 JSON 字符串
func (r *SearchResponse) PrettyJSON() string {
	data, _ := json.MarshalIndent(r, "", "  ")
	return string(data)
}

// String 实现 Stringer 接口，默认返回格式化 JSON
func (r *SearchResponse) String() string {
	return r.PrettyJSON()
}

// Build 构建查询 DSL
func (b *SearchBuilder) Build() map[string]interface{} {
	body := make(map[string]interface{})

	// 构建 bool 查询
	if len(b.must) > 0 || len(b.filters) > 0 || len(b.should) > 0 || len(b.mustNot) > 0 {
		boolQuery := make(map[string]interface{})
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
		body["query"] = map[string]interface{}{
			"bool": boolQuery,
		}
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

// Do 执行搜索
func (b *SearchBuilder) Do(ctx context.Context) (*SearchResponse, error) {
	path := fmt.Sprintf("/%s/_search", b.index)
	body := b.Build()

	respBody, err := b.client.Do(ctx, http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}

	var resp SearchResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &resp, nil
}
