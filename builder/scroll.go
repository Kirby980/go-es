package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ScrollBuilder Scroll深度分页构建器
type ScrollBuilder struct {
	client    ESClient
	index     string
	filters   []map[string]any
	must      []map[string]any
	should    []map[string]any
	mustNot   []map[string]any
	size      int
	keepAlive string
	scrollID  string
	DebugHelper
}

// NewScrollBuilder 创建Scroll构建器
func NewScrollBuilder(c ESClient, index string) *ScrollBuilder {
	return &ScrollBuilder{
		client:    c,
		index:     index,
		filters:   make([]map[string]any, 0),
		must:      make([]map[string]any, 0),
		should:    make([]map[string]any, 0),
		mustNot:   make([]map[string]any, 0),
		size:      1000,
		keepAlive: "5m",
	}
}

// Match 添加 match 查询条件
func (b *ScrollBuilder) Match(field string, value any) *ScrollBuilder {
	b.must = append(b.must, map[string]any{
		"match": map[string]any{
			field: value,
		},
	})
	return b
}

// Term 添加 term 查询条件
func (b *ScrollBuilder) Term(field string, value any) *ScrollBuilder {
	b.filters = append(b.filters, map[string]any{
		"term": map[string]any{
			field: value,
		},
	})
	return b
}

// Range 添加范围查询条件
func (b *ScrollBuilder) Range(field string, gte, lte any) *ScrollBuilder {
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
	b.SetDebug(true)
	return b
}

// Build 构建查询体
func (b *ScrollBuilder) Build() map[string]any {
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
		body["query"] = map[string]any{
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
			Index     string              `json:"_index"`
			ID        string              `json:"_id"`
			Score     float64             `json:"_score"`
			Source    map[string]any      `json:"_source"`
			Highlight map[string][]string `json:"highlight,omitempty"`
		} `json:"hits"`
	} `json:"hits"`
}

// Do 执行第一次scroll查询
func (b *ScrollBuilder) Do(ctx context.Context) (*ScrollResponse, error) {
	path := fmt.Sprintf("/%s/_search?scroll=%s", b.index, b.keepAlive)
	body := b.Build()

	// 如果启用调试模式，打印请求信息
	if b.debug {
		b.PrintDebug("POST", path, body)
		defer b.SetDebug(false)
	}

	respBody, err := b.client.Do(ctx, http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}

	// 如果启用调试模式，打印响应信息
	if b.debug {
		b.PrintResponse(respBody)
		defer b.SetDebug(false)
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
	body := map[string]any{
		"scroll":    b.keepAlive,
		"scroll_id": b.scrollID,
	}

	// 如果启用调试模式，打印请求信息
	if b.debug {
		b.PrintDebug("POST", path, body)
		defer b.SetDebug(false)
	}

	respBody, err := b.client.Do(ctx, http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}

	// 如果启用调试模式，打印响应信息
	if b.debug {
		b.PrintResponse(respBody)
		defer b.SetDebug(false)
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
	body := map[string]any{
		"scroll_id": b.scrollID,
	}

	// 如果启用调试模式，打印请求信息
	if b.debug {
		b.PrintDebug("DELETE", path, body)
		defer b.SetDebug(false)
	}

	respBody, err := b.client.Do(ctx, http.MethodDelete, path, body)
	if err != nil {
		return err
	}

	// 如果启用调试模式，打印响应信息
	if b.debug {
		b.PrintResponse(respBody)
		defer b.SetDebug(false)
	}

	b.scrollID = ""
	return nil
}

// HasMore 判断是否还有更多数据
func (b *ScrollBuilder) HasMore(resp *ScrollResponse) bool {
	return len(resp.Hits.Hits) > 0
}
