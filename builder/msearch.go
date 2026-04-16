package builder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// MultiSearchBuilder 批量搜索构建器 (msearch)
type MultiSearchBuilder struct {
	client   ESClient
	searches []multiSearchItem
	debugHelper
	baseBuilder
}

type multiSearchItem struct {
	Header map[string]any // 请求头，如 index
	Body   map[string]any // 搜索请求体
}

// NewMultiSearchBuilder 创建批量搜索构建器
func NewMultiSearchBuilder(c ESClient) *MultiSearchBuilder {
	b := &MultiSearchBuilder{
		client:      c,
		searches:    make([]multiSearchItem, 0),
		debugHelper: debugHelper{logger: c.GetLogger()},
	}
	b.initBaseBuilder()
	return b
}

// Add 添加一个搜索请求
func (b *MultiSearchBuilder) Add(index string, search *SearchBuilder) *MultiSearchBuilder {
	header := make(map[string]any)
	if index != "" {
		header["index"] = index
	} else if search.index != "" {
		header["index"] = search.index
	}

	b.searches = append(b.searches, multiSearchItem{
		Header: header,
		Body:   search.Build(),
	})
	return b
}

// Header 设置自定义 Header
func (b *MultiSearchBuilder) Header(key, value string) *MultiSearchBuilder {
	b.baseBuilder.Header(key, value)
	return b
}

// Debug 启用调试模式
func (b *MultiSearchBuilder) Debug() *MultiSearchBuilder {
	b.setDebug(true)
	return b
}

// MultiSearchResponse 批量搜索响应
type MultiSearchResponse struct {
	Took      int `json:"took"`
	Responses []struct {
		SearchResponse
		Status int `json:"status,omitempty"`
		Error  *struct {
			Type   string `json:"type"`
			Reason string `json:"reason"`
		} `json:"error,omitempty"`
	} `json:"responses"`
}

// Build 构建 ndjson 请求体
func (b *MultiSearchBuilder) Build() ([]byte, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)

	for _, item := range b.searches {
		if err := encoder.Encode(item.Header); err != nil {
			return nil, fmt.Errorf("序列化 header 失败: %w", err)
		}
		if err := encoder.Encode(item.Body); err != nil {
			return nil, fmt.Errorf("序列化 body 失败: %w", err)
		}
	}
	return buf.Bytes(), nil
}

// Do 执行批量搜索
func (b *MultiSearchBuilder) Do(ctx context.Context) (*MultiSearchResponse, error) {
	if len(b.searches) == 0 {
		return nil, fmt.Errorf("没有待执行的搜索请求")
	}

	body, err := b.Build()
	if err != nil {
		return nil, err
	}

	path := "/_msearch"

	if b.isDebug() {
		b.printDebug("POST", path, body)
		defer b.setDebug(false)
	}

	// 创建请求
	url := b.client.GetAddress() + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置 NDJSON Content-Type
	req.Header.Set("Content-Type", "application/x-ndjson")
	for k, v := range b.getHeaders() {
		for _, vv := range v {
			req.Header.Add(k, vv)
		}
	}

	respBody, err := b.client.DoRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	if b.isDebug() {
		b.printResponse(respBody)
	}

	var resp MultiSearchResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &resp, nil
}
