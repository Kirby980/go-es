package builder

import (
	"context"
	"fmt"
	"net/http"
)

// AliasesBuilder 是用于构建和执行 Elasticsearch Aliases 操作的链式 API 构建器。
type AliasesBuilder struct {
	client  ESClient
	actions []map[string]any
	debugHelper
	baseBuilder
}

// NewAliasesBuilder 创建并返回一个 Aliases 构建器实例，用于构造和执行 Elasticsearch 的 Aliases 请求。
func NewAliasesBuilder(c ESClient) *AliasesBuilder {
	b := &AliasesBuilder{
		client:      c,
		actions:     make([]map[string]any, 0),
		debugHelper: debugHelper{logger: c.GetLogger()},
	}
	b.initBaseBuilder()
	return b
}

// Header 设置自定义的 HTTP 请求头 (例如: Header("Content-Type", "application/json"))。此方法支持链式调用。
func (b *AliasesBuilder) Header(key, value string) *AliasesBuilder {
	b.baseBuilder.Header(key, value)
	return b
}

// Debug 开启当前请求的单次调试模式。开启后，请求的完整 HTTP URL、Method、Body 及响应将被打印到日志中，便于排查问题。
func (b *AliasesBuilder) Debug() *AliasesBuilder {
	b.setDebug(true)
	return b
}

// Add 向当前构建器中追加一个新项或规则。
func (b *AliasesBuilder) Add(index string, alias string, opts ...func(map[string]any)) *AliasesBuilder {
	params := map[string]any{
		"index": index,
		"alias": alias,
	}
	for _, opt := range opts {
		opt(params)
	}
	b.actions = append(b.actions, map[string]any{
		"add": params,
	})
	return b
}

// Remove 从别名或集合中移除指定项。
func (b *AliasesBuilder) Remove(index string, alias string, opts ...func(map[string]any)) *AliasesBuilder {
	params := map[string]any{
		"index": index,
		"alias": alias,
	}
	for _, opt := range opts {
		opt(params)
	}
	b.actions = append(b.actions, map[string]any{
		"remove": params,
	})
	return b
}

// RemoveIndex 移除指定索引的关联或操作。
func (b *AliasesBuilder) RemoveIndex(index string) *AliasesBuilder {
	b.actions = append(b.actions, map[string]any{
		"remove_index": map[string]any{
			"index": index,
		},
	})
	return b
}

// Replace 执行替换操作。
func (b *AliasesBuilder) Replace(alias string, fromIndex string, toIndex string, opts ...func(map[string]any)) *AliasesBuilder {
	b.Remove(fromIndex, alias)
	b.Add(toIndex, alias, opts...)
	return b
}

// WithIsWriteIndex 配置别名是否作为写入索引 (is_write_index)。
func WithIsWriteIndex(isWrite bool) func(map[string]any) {
	return func(m map[string]any) {
		m["is_write_index"] = isWrite
	}
}

// WithAliasFilter 为别名配置额外的过滤条件 (filter)，实现基于别名的视图隔离。
func WithAliasFilter(filter map[string]any) func(map[string]any) {
	return func(m map[string]any) {
		if filter != nil {
			m["filter"] = filter
		}
	}
}

// WithRouting 为别名配置自定义的路由键 (routing)，以控制数据分片分配。
func WithRouting(routing string) func(map[string]any) {
	return func(m map[string]any) {
		if routing != "" {
			m["routing"] = routing
		}
	}
}

// AliasesResponse 定义了 Elasticsearch Aliases 操作返回的结构化响应数据，包含元数据及核心结果。
type AliasesResponse struct {
	Acknowledged bool `json:"acknowledged"`
}

// Build 根据当前构建器的状态组装请求体参数 (通常为 map[string]any 格式)，用于最终发送到 Elasticsearch。
func (b *AliasesBuilder) Build() map[string]any {
	return map[string]any{
		"actions": b.actions,
	}
}

// Do 立即执行当前构建好的请求，并返回结构化的 Elasticsearch 响应结果或错误。请确保在调用此方法前已设置好所有必要参数。
func (b *AliasesBuilder) Do(ctx context.Context) (*AliasesResponse, error) {
	if len(b.actions) == 0 {
		return nil, fmt.Errorf("没有待执行的 alias 操作")
	}

	path := "/_aliases"
	body := b.Build()

	if b.isDebug() {
		b.printDebug("POST", path, body)
		defer b.autoResetDebug()
	}

	var resp AliasesResponse
	if err := b.client.DoWithHeaderAndDecode(ctx, http.MethodPost, path, body, b.getHeaders(), &resp); err != nil {
		return nil, err
	}

	if b.isDebug() {
		b.printResponseObj(resp)
	}

	return &resp, nil
}
