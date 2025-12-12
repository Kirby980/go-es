package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Kirby980/go-es/client"
)

// UpdateByQueryBuilder 按查询更新构建器
type UpdateByQueryBuilder struct {
	client  *client.Client
	index   string
	filters []map[string]interface{}
	must    []map[string]interface{}
	should  []map[string]interface{}
	mustNot []map[string]interface{}
	script  map[string]interface{}
	debug   bool
}

// NewUpdateByQueryBuilder 创建按查询更新构建器
func NewUpdateByQueryBuilder(c *client.Client, index string) *UpdateByQueryBuilder {
	return &UpdateByQueryBuilder{
		client:  c,
		index:   index,
		filters: make([]map[string]interface{}, 0),
		must:    make([]map[string]interface{}, 0),
		should:  make([]map[string]interface{}, 0),
		mustNot: make([]map[string]interface{}, 0),
	}
}

// Match 添加 match 查询条件
func (b *UpdateByQueryBuilder) Match(field string, value interface{}) *UpdateByQueryBuilder {
	b.must = append(b.must, map[string]interface{}{
		"match": map[string]interface{}{
			field: value,
		},
	})
	return b
}

// Term 添加 term 查询条件
func (b *UpdateByQueryBuilder) Term(field string, value interface{}) *UpdateByQueryBuilder {
	b.filters = append(b.filters, map[string]interface{}{
		"term": map[string]interface{}{
			field: value,
		},
	})
	return b
}

// Terms 添加 terms 查询条件
func (b *UpdateByQueryBuilder) Terms(field string, values ...interface{}) *UpdateByQueryBuilder {
	b.filters = append(b.filters, map[string]interface{}{
		"terms": map[string]interface{}{
			field: values,
		},
	})
	return b
}

// Range 添加范围查询条件
func (b *UpdateByQueryBuilder) Range(field string, gte, lte interface{}) *UpdateByQueryBuilder {
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

// Script 设置更新脚本
func (b *UpdateByQueryBuilder) Script(source string, params map[string]interface{}) *UpdateByQueryBuilder {
	b.script = map[string]interface{}{
		"source": source,
		"lang":   "painless",
	}
	if params != nil {
		b.script["params"] = params
	}
	return b
}

// Set 设置字段值（简化的脚本更新）
func (b *UpdateByQueryBuilder) Set(field string, value interface{}) *UpdateByQueryBuilder {
	// 如果已有脚本，追加
	if b.script != nil {
		existingSource := b.script["source"].(string)
		b.script["source"] = existingSource + fmt.Sprintf("; ctx._source.%s = params.%s", field, field)
	} else {
		b.script = map[string]interface{}{
			"source": fmt.Sprintf("ctx._source.%s = params.%s", field, field),
			"lang":   "painless",
		}
	}

	// 添加参数
	if b.script["params"] == nil {
		b.script["params"] = make(map[string]interface{})
	}
	b.script["params"].(map[string]interface{})[field] = value

	return b
}

// Debug 启用调试模式
func (b *UpdateByQueryBuilder) Debug() *UpdateByQueryBuilder {
	b.debug = true
	return b
}

// printDebug 打印请求调试信息
func (b *UpdateByQueryBuilder) printDebug(method, path string, body interface{}) {
	fmt.Printf("\n[ES Debug] %s %s\n", method, path)
	if body != nil {
		data, _ := json.MarshalIndent(body, "", "  ")
		fmt.Printf("Request Body:\n%s\n", string(data))
	}
}

// printResponse 打印响应调试信息
func (b *UpdateByQueryBuilder) printResponse(respBody []byte) {
	var pretty interface{}
	json.Unmarshal(respBody, &pretty)
	data, _ := json.MarshalIndent(pretty, "", "  ")
	fmt.Printf("Response:\n%s\n\n", string(data))
}

// Build 构建请求体
func (b *UpdateByQueryBuilder) Build() map[string]interface{} {
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

	// 添加脚本
	if b.script != nil {
		body["script"] = b.script
	}

	return body
}

// UpdateByQueryResponse 更新响应
type UpdateByQueryResponse struct {
	Took             int    `json:"took"`
	TimedOut         bool   `json:"timed_out"`
	Total            int    `json:"total"`
	Updated          int    `json:"updated"`
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

// Do 执行更新
func (b *UpdateByQueryBuilder) Do(ctx context.Context) (*UpdateByQueryResponse, error) {
	if b.script == nil {
		return nil, fmt.Errorf("必须设置更新脚本")
	}

	path := fmt.Sprintf("/%s/_update_by_query", b.index)
	body := b.Build()

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

	var resp UpdateByQueryResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &resp, nil
}
