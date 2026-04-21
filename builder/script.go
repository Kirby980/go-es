package builder

import (
	"context"
	"net/http"
	"net/url"
)

// ScriptBuilder 是用于构建和执行 Elasticsearch Script 操作的链式 API 构建器。
type ScriptBuilder struct {
	client ESClient
	id     string
	script map[string]any
	debugHelper
	baseBuilder
}

// NewScriptBuilder 创建并返回一个 Script 构建器实例，用于构造和执行 Elasticsearch 的 Script 请求。
func NewScriptBuilder(c ESClient) *ScriptBuilder {
	b := &ScriptBuilder{
		client: c,
	}
	b.initBaseBuilder()
	return b
}

// Header 设置自定义的 HTTP 请求头 (例如: Header("Content-Type", "application/json"))。此方法支持链式调用。
func (b *ScriptBuilder) Header(key, value string) *ScriptBuilder {
	b.baseBuilder.Header(key, value)
	return b
}

// ID 指定操作目标文档或资源的唯一标识符 (ID)。
func (b *ScriptBuilder) ID(id string) *ScriptBuilder {
	b.id = id
	return b
}

// Script 指定要执行的脚本内容 (如 Painless 脚本) 及相关参数，用于动态更新或处理数据。
func (b *ScriptBuilder) Script(lang string, source string, params map[string]any) *ScriptBuilder {
	s := map[string]any{
		"lang":   lang,
		"source": source,
	}
	if params != nil {
		s["params"] = params
	}
	b.script = s
	return b
}

// Debug 开启当前请求的单次调试模式。开启后，请求的完整 HTTP URL、Method、Body 及响应将被打印到日志中，便于排查问题。
func (b *ScriptBuilder) Debug(enable bool) *ScriptBuilder {
	b.setDebugPersistent(enable)
	return b
}

// Put 发起 PUT 请求以创建或更新目标资源。
func (b *ScriptBuilder) Put(ctx context.Context) (*AcknowledgedResponse, error) {
	path := "/_scripts/" + url.PathEscape(b.id)
	body := map[string]any{
		"script": b.script,
	}

	if b.isDebug() {
		b.printDebug(http.MethodPut, path, body)
		defer b.autoResetDebug()
	}

	var resp AcknowledgedResponse
	if err := b.client.DoWithHeaderAndDecode(ctx, http.MethodPut, path, body, b.getHeaders(), &resp); err != nil {
		return nil, err
	}

	if b.isDebug() {
		b.printResponseObj(resp)
	}

	return &resp, nil
}

// Get 发起 GET 请求以获取目标资源的详细信息。
func (b *ScriptBuilder) Get(ctx context.Context) (map[string]any, error) {
	path := "/_scripts/" + url.PathEscape(b.id)

	if b.isDebug() {
		b.printDebug(http.MethodGet, path, nil)
		defer b.autoResetDebug()
	}

	var resp map[string]any
	if err := b.client.DoWithHeaderAndDecode(ctx, http.MethodGet, path, nil, b.getHeaders(), &resp); err != nil {
		return nil, err
	}

	if b.isDebug() {
		b.printResponseObj(resp)
	}

	return resp, nil
}

// Delete 发起 DELETE 请求以删除目标资源。
func (b *ScriptBuilder) Delete(ctx context.Context) (*AcknowledgedResponse, error) {
	path := "/_scripts/" + url.PathEscape(b.id)

	if b.isDebug() {
		b.printDebug(http.MethodDelete, path, nil)
		defer b.autoResetDebug()
	}

	var resp AcknowledgedResponse
	if err := b.client.DoWithHeaderAndDecode(ctx, http.MethodDelete, path, nil, b.getHeaders(), &resp); err != nil {
		return nil, err
	}

	if b.isDebug() {
		b.printResponseObj(resp)
	}

	return &resp, nil
}
