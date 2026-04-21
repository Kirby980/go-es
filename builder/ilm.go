package builder

import (
	"context"
	"net/http"
	"net/url"
)

// ILMBuilder 是用于构建和执行 Elasticsearch ILM 操作的链式 API 构建器。
type ILMBuilder struct {
	client ESClient
	name   string
	policy map[string]any
	debugHelper
	baseBuilder
}

// NewILMBuilder 创建并返回一个 ILM 构建器实例，用于构造和执行 Elasticsearch 的 ILM 请求。
func NewILMBuilder(client ESClient) *ILMBuilder {
	return &ILMBuilder{
		client: client,
	}
}

// Name 指定策略或组件的名称。
func (b *ILMBuilder) Name(name string) *ILMBuilder {
	b.name = name
	return b
}

// Policy 设置索引生命周期管理 (ILM) 等策略的具体规则。
func (b *ILMBuilder) Policy(policy map[string]any) *ILMBuilder {
	b.policy = policy
	return b
}

// Debug 开启当前请求的单次调试模式。开启后，请求的完整 HTTP URL、Method、Body 及响应将被打印到日志中，便于排查问题。
func (b *ILMBuilder) Debug(enable bool) *ILMBuilder {
	b.setDebug(enable)
	return b
}

// Put 发起 PUT 请求以创建或更新目标资源。
func (b *ILMBuilder) Put(ctx context.Context) (*AcknowledgedResponse, error) {
	path := "/_ilm/policy/" + url.PathEscape(b.name)
	body := map[string]any{
		"policy": b.policy,
	}

	if b.isDebug() {
		b.printDebug("PUT", path, body)
		defer b.autoResetDebug()
	}

	var resp AcknowledgedResponse
	err := b.client.DoWithHeaderAndDecode(ctx, http.MethodPut, path, body, b.getHeaders(), &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// Get 发起 GET 请求以获取目标资源的详细信息。
func (b *ILMBuilder) Get(ctx context.Context) (map[string]any, error) {
	path := "/_ilm/policy"
	if b.name != "" {
		path += "/" + url.PathEscape(b.name)
	}

	if b.isDebug() {
		b.printDebug("GET", path, nil)
		defer b.autoResetDebug()
	}

	var resp map[string]any
	err := b.client.DoWithHeaderAndDecode(ctx, http.MethodGet, path, nil, b.getHeaders(), &resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Delete 发起 DELETE 请求以删除目标资源。
func (b *ILMBuilder) Delete(ctx context.Context) (*AcknowledgedResponse, error) {
	path := "/_ilm/policy/" + url.PathEscape(b.name)

	if b.isDebug() {
		b.printDebug("DELETE", path, nil)
		defer b.autoResetDebug()
	}

	var resp AcknowledgedResponse
	err := b.client.DoWithHeaderAndDecode(ctx, http.MethodDelete, path, nil, b.getHeaders(), &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
