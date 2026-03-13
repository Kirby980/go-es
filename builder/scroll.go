package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// ScrollBuilder Scroll深度分页构建器
type ScrollBuilder struct {
	client    ESClient
	index     string
	size      int
	keepAlive string
	scrollID  string
	BoolQuery[ScrollBuilder]
	debugHelper
	baseBuilder
}

// NewScrollBuilder 创建Scroll构建器
func NewScrollBuilder(c ESClient, index string) *ScrollBuilder {
	b := &ScrollBuilder{
		client:    c,
		index:     index,
		size:      1000,
		keepAlive: "5m",
		debugHelper: debugHelper{logger: c.GetLogger()},
	}
	b.initBoolQuery(b)
	b.initBaseBuilder()
	return b
}

// Header 设置自定义 Header (链式调用)
func (b *ScrollBuilder) Header(key, value string) *ScrollBuilder {
	b.baseBuilder.Header(key, value)
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
	b.setDebug(true)
	return b
}

// Build 构建查询体
func (b *ScrollBuilder) Build() map[string]any {
	body := make(map[string]any)

	// 构建查询条件
	if boolQ := b.buildBoolQuery(); boolQ != nil {
		body["query"] = boolQ
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
	path := fmt.Sprintf("/%s/_search?scroll=%s", url.PathEscape(b.index), b.keepAlive)
	body := b.Build()

	// 如果启用调试模式，打印请求信息
	if b.debug {
		b.printDebug("POST", path, body)
		defer b.setDebug(false)
	}

	respBody, err := b.client.DoWithHeader(ctx, http.MethodPost, path, body, b.getHeaders())
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
	body := map[string]any{
		"scroll":    b.keepAlive,
		"scroll_id": b.scrollID,
	}

	// 如果启用调试模式，打印请求信息
	if b.debug {
		b.printDebug("POST", path, body)
		defer b.setDebug(false)
	}

	respBody, err := b.client.DoWithHeader(ctx, http.MethodPost, path, body, b.getHeaders())
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
	body := map[string]any{
		"scroll_id": b.scrollID,
	}

	// 如果启用调试模式，打印请求信息
	if b.debug {
		b.printDebug("DELETE", path, body)
		defer b.setDebug(false)
	}

	respBody, err := b.client.DoWithHeader(ctx, http.MethodDelete, path, body, b.getHeaders())
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
