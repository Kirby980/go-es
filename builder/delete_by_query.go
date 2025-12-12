package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Kirby980/go-es/client"
)

// DeleteByQueryBuilder 按查询删除构建器
type DeleteByQueryBuilder struct {
	client  *client.Client
	index   string
	filters []map[string]interface{}
	must    []map[string]interface{}
	should  []map[string]interface{}
	mustNot []map[string]interface{}
	debug   bool
}

// NewDeleteByQueryBuilder 创建按查询删除构建器
func NewDeleteByQueryBuilder(c *client.Client, index string) *DeleteByQueryBuilder {
	return &DeleteByQueryBuilder{
		client:  c,
		index:   index,
		filters: make([]map[string]interface{}, 0),
		must:    make([]map[string]interface{}, 0),
		should:  make([]map[string]interface{}, 0),
		mustNot: make([]map[string]interface{}, 0),
	}
}

// Match 添加 match 查询条件
func (b *DeleteByQueryBuilder) Match(field string, value interface{}) *DeleteByQueryBuilder {
	b.must = append(b.must, map[string]interface{}{
		"match": map[string]interface{}{
			field: value,
		},
	})
	return b
}

// Term 添加 term 查询条件
func (b *DeleteByQueryBuilder) Term(field string, value interface{}) *DeleteByQueryBuilder {
	b.filters = append(b.filters, map[string]interface{}{
		"term": map[string]interface{}{
			field: value,
		},
	})
	return b
}

// Terms 添加 terms 查询条件
func (b *DeleteByQueryBuilder) Terms(field string, values ...interface{}) *DeleteByQueryBuilder {
	b.filters = append(b.filters, map[string]interface{}{
		"terms": map[string]interface{}{
			field: values,
		},
	})
	return b
}

// Range 添加范围查询条件
func (b *DeleteByQueryBuilder) Range(field string, gte, lte interface{}) *DeleteByQueryBuilder {
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
func (b *DeleteByQueryBuilder) Exists(field string) *DeleteByQueryBuilder {
	b.filters = append(b.filters, map[string]interface{}{
		"exists": map[string]interface{}{
			"field": field,
		},
	})
	return b
}

// Debug 启用调试模式
func (b *DeleteByQueryBuilder) Debug() *DeleteByQueryBuilder {
	b.debug = true
	return b
}

// printDebug 打印请求调试信息
func (b *DeleteByQueryBuilder) printDebug(method, path string, body interface{}) {
	fmt.Printf("\n[ES Debug] %s %s\n", method, path)
	if body != nil {
		data, _ := json.MarshalIndent(body, "", "  ")
		fmt.Printf("Request Body:\n%s\n", string(data))
	}
}

// printResponse 打印响应调试信息
func (b *DeleteByQueryBuilder) printResponse(respBody []byte) {
	var pretty interface{}
	json.Unmarshal(respBody, &pretty)
	data, _ := json.MarshalIndent(pretty, "", "  ")
	fmt.Printf("Response:\n%s\n\n", string(data))
}

// Build 构建请求体
func (b *DeleteByQueryBuilder) Build() map[string]interface{} {
	body := make(map[string]interface{})

	// 构建查询条件
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

	return body
}

// DeleteByQueryResponse 删除响应
type DeleteByQueryResponse struct {
	Took             int    `json:"took"`
	TimedOut         bool   `json:"timed_out"`
	Total            int    `json:"total"`
	Deleted          int    `json:"deleted"`
	Batches          int    `json:"batches"`
	VersionConflicts int    `json:"version_conflicts"`
	Noops            int    `json:"noops"`
	Retries          struct {
		Bulk   int `json:"bulk"`
		Search int `json:"search"`
	} `json:"retries"`
	ThrottledMillis      int                      `json:"throttled_millis"`
	RequestsPerSecond    float64                  `json:"requests_per_second"`
	ThrottledUntilMillis int                      `json:"throttled_until_millis"`
	Failures             []map[string]interface{} `json:"failures"`
}

// Do 执行删除
func (b *DeleteByQueryBuilder) Do(ctx context.Context) (*DeleteByQueryResponse, error) {
	path := fmt.Sprintf("/%s/_delete_by_query", b.index)
	body := b.Build()

	// 检查是否有查询条件
	if len(body) == 0 {
		return nil, fmt.Errorf("必须设置查询条件，避免误删除所有数据")
	}

	// 如果启用调试模式，打印请求信息
	if b.debug {
		b.printDebug("POST", path, body)
	}

	respBody, err := b.client.Do(ctx, http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}

	// 如果启用调试模式，打印响应信息
	if b.debug {
		b.printResponse(respBody)
	}

	var resp DeleteByQueryResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &resp, nil
}
