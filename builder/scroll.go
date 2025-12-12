package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Kirby980/go-es/client"
)

// ScrollBuilder Scroll深度分页构建器
type ScrollBuilder struct {
	client    *client.Client
	index     string
	filters   []map[string]interface{}
	must      []map[string]interface{}
	should    []map[string]interface{}
	mustNot   []map[string]interface{}
	size      int
	keepAlive string
	scrollID  string
	debug     bool
}

// NewScrollBuilder 创建Scroll构建器
func NewScrollBuilder(c *client.Client, index string) *ScrollBuilder {
	return &ScrollBuilder{
		client:    c,
		index:     index,
		filters:   make([]map[string]interface{}, 0),
		must:      make([]map[string]interface{}, 0),
		should:    make([]map[string]interface{}, 0),
		mustNot:   make([]map[string]interface{}, 0),
		size:      1000,
		keepAlive: "5m",
	}
}

// Match 添加 match 查询条件
func (b *ScrollBuilder) Match(field string, value interface{}) *ScrollBuilder {
	b.must = append(b.must, map[string]interface{}{
		"match": map[string]interface{}{
			field: value,
		},
	})
	return b
}

// Term 添加 term 查询条件
func (b *ScrollBuilder) Term(field string, value interface{}) *ScrollBuilder {
	b.filters = append(b.filters, map[string]interface{}{
		"term": map[string]interface{}{
			field: value,
		},
	})
	return b
}

// Range 添加范围查询条件
func (b *ScrollBuilder) Range(field string, gte, lte interface{}) *ScrollBuilder {
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

// Size 设置每批返回的文档数量
func (b *ScrollBuilder) Size(size int) *ScrollBuilder {
	b.size = size
	return b
}

// KeepAlive 设置scroll上下文保持时间（如"5m"、"1h"）
func (b *ScrollBuilder) KeepAlive(keepAlive string) *ScrollBuilder {
	b.keepAlive = keepAlive
	return b
}

// Debug 启用调试模式
func (b *ScrollBuilder) Debug() *ScrollBuilder {
	b.debug = true
	return b
}

// printDebug 打印请求调试信息
func (b *ScrollBuilder) printDebug(method, path string, body interface{}) {
	fmt.Printf("\n[ES Debug] %s %s\n", method, path)
	if body != nil {
		data, _ := json.MarshalIndent(body, "", "  ")
		fmt.Printf("Request Body:\n%s\n", string(data))
	}
}

// printResponse 打印响应调试信息
func (b *ScrollBuilder) printResponse(respBody []byte) {
	var pretty interface{}
	json.Unmarshal(respBody, &pretty)
	data, _ := json.MarshalIndent(pretty, "", "  ")
	fmt.Printf("Response:\n%s\n\n", string(data))
}

// Build 构建查询体
func (b *ScrollBuilder) Build() map[string]interface{} {
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

	body["size"] = b.size

	return body
}

// ScrollResponse Scroll响应
type ScrollResponse struct {
	ScrollID string `json:"_scroll_id"`
	Took     int    `json:"took"`
	TimedOut bool   `json:"timed_out"`
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
			Index     string                 `json:"_index"`
			ID        string                 `json:"_id"`
			Score     float64                `json:"_score"`
			Source    map[string]interface{} `json:"_source"`
			Highlight map[string][]string    `json:"highlight,omitempty"`
		} `json:"hits"`
	} `json:"hits"`
}

// Do 执行第一次scroll查询
func (b *ScrollBuilder) Do(ctx context.Context) (*ScrollResponse, error) {
	path := fmt.Sprintf("/%s/_search?scroll=%s", b.index, b.keepAlive)
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

	var resp ScrollResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 保存scroll ID供下次使用
	b.scrollID = resp.ScrollID

	return &resp, nil
}

// Next 获取下一批数据
func (b *ScrollBuilder) Next(ctx context.Context) (*ScrollResponse, error) {
	if b.scrollID == "" {
		return nil, fmt.Errorf("请先调用Do()方法初始化scroll")
	}

	path := "/_search/scroll"
	body := map[string]interface{}{
		"scroll":    b.keepAlive,
		"scroll_id": b.scrollID,
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

	var resp ScrollResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 更新scroll ID
	b.scrollID = resp.ScrollID

	return &resp, nil
}

// Clear 清除scroll上下文
func (b *ScrollBuilder) Clear(ctx context.Context) error {
	if b.scrollID == "" {
		return nil
	}

	path := "/_search/scroll"
	body := map[string]interface{}{
		"scroll_id": b.scrollID,
	}

	// 如果启用调试模式，打印请求信息
	if b.debug {
		b.printDebug("DELETE", path, body)
	}

	respBody, err := b.client.Do(ctx, http.MethodDelete, path, body)
	if err != nil {
		return err
	}

	// 如果启用调试模式，打印响应信息
	if b.debug {
		b.printResponse(respBody)
	}

	b.scrollID = ""
	return nil
}

// HasMore 判断是否还有更多数据
func (b *ScrollBuilder) HasMore(resp *ScrollResponse) bool {
	return len(resp.Hits.Hits) > 0
}
