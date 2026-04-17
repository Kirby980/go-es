package builder

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// PitBuilder Point In Time (PIT) 构建器 (ES 7.10+ / 8.x / 9.x)
// 现代 Elasticsearch 分页推荐使用 PIT 替代 Scroll
type PitBuilder struct {
	client    ESClient
	index     string
	keepAlive string
	debugHelper
	baseBuilder
}

// NewPitBuilder 创建 PIT 构建器
func NewPitBuilder(c ESClient, index string) *PitBuilder {
	b := &PitBuilder{
		client:      c,
		index:       index,
		keepAlive:   "1m", // 默认保留时间 1 分钟
		debugHelper: debugHelper{logger: c.GetLogger()},
	}
	b.initBaseBuilder()
	return b
}

// Header 设置自定义 Header
func (b *PitBuilder) Header(key, value string) *PitBuilder {
	b.baseBuilder.Header(key, value)
	return b
}

// KeepAlive 设置保留时间 (如 "1m", "5m")
func (b *PitBuilder) KeepAlive(keepAlive string) *PitBuilder {
	b.keepAlive = keepAlive
	return b
}

// Debug 启用调试模式
func (b *PitBuilder) Debug() *PitBuilder {
	b.setDebug(true)
	return b
}

// CreatePitResponse 创建 PIT 的响应
type CreatePitResponse struct {
	ID string `json:"id"`
}

// Create 执行创建 PIT 请求
func (b *PitBuilder) Create(ctx context.Context) (*CreatePitResponse, error) {
	path := fmt.Sprintf("/%s/_pit?keep_alive=%s", url.PathEscape(b.index), b.keepAlive)

	if b.isDebug() {
		b.printDebug("POST", path, nil)
		defer b.autoResetDebug()
	}

	var resp CreatePitResponse
	if err := b.client.DoWithHeaderAndDecode(ctx, http.MethodPost, path, nil, b.getHeaders(), &resp); err != nil {
		return nil, err
	}
	if b.isDebug() {
		b.printResponseObj(resp)
	}

	return &resp, nil
}

// DeletePitResponse 删除 PIT 的响应
type DeletePitResponse struct {
	Succeeded bool  `json:"succeeded"`
	NumFreed  int   `json:"num_freed"`
	Error     error `json:"error,omitempty"`
}

// Delete 执行删除 PIT 请求 (全局操作，不需要 index)
func (b *PitBuilder) Delete(ctx context.Context, pitID string) (*DeletePitResponse, error) {
	path := "/_pit"
	body := map[string]any{
		"id": pitID,
	}

	if b.isDebug() {
		b.printDebug("DELETE", path, body)
		defer b.autoResetDebug()
	}

	var resp DeletePitResponse
	if err := b.client.DoWithHeaderAndDecode(ctx, http.MethodDelete, path, body, b.getHeaders(), &resp); err != nil {
		return nil, err
	}
	if b.isDebug() {
		b.printResponseObj(resp)
	}

	return &resp, nil
}
