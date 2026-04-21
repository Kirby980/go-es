package builder

import (
	"context"
	"net/http"
	"net/url"
	"strings"
)

// FieldCapsBuilder 是用于构建和执行 Elasticsearch FieldCaps 操作的链式 API 构建器。
type FieldCapsBuilder struct {
	client          ESClient
	index           string
	fields          []string
	includeUnmapped *bool
	debugHelper
	baseBuilder
}

// NewFieldCapsBuilder 创建并返回一个 FieldCaps 构建器实例，用于构造和执行 Elasticsearch 的 FieldCaps 请求。
func NewFieldCapsBuilder(c ESClient, index string) *FieldCapsBuilder {
	b := &FieldCapsBuilder{
		client: c,
		index:  index,
	}
	b.initBaseBuilder()
	return b
}

// Header 设置自定义的 HTTP 请求头 (例如: Header("Content-Type", "application/json"))。此方法支持链式调用。
func (b *FieldCapsBuilder) Header(key, value string) *FieldCapsBuilder {
	b.baseBuilder.Header(key, value)
	return b
}

// Fields 指定需要查询或返回的具体字段列表。
func (b *FieldCapsBuilder) Fields(fields ...string) *FieldCapsBuilder {
	b.fields = append(b.fields, fields...)
	return b
}

// IncludeUnmapped 设置是否在结果中包含未映射 (unmapped) 的字段信息。
func (b *FieldCapsBuilder) IncludeUnmapped(enable bool) *FieldCapsBuilder {
	b.includeUnmapped = &enable
	return b
}

// Debug 开启当前请求的单次调试模式。开启后，请求的完整 HTTP URL、Method、Body 及响应将被打印到日志中，便于排查问题。
func (b *FieldCapsBuilder) Debug(enable bool) *FieldCapsBuilder {
	b.setDebugPersistent(enable)
	return b
}

// Do 立即执行当前构建好的请求，并返回结构化的 Elasticsearch 响应结果或错误。请确保在调用此方法前已设置好所有必要参数。
func (b *FieldCapsBuilder) Do(ctx context.Context) (map[string]any, error) {
	path := "/_field_caps"
	if b.index != "" {
		path = "/" + url.PathEscape(b.index) + path
	}

	values := url.Values{}
	if len(b.fields) > 0 {
		values.Set("fields", strings.Join(b.fields, ","))
	}
	if b.includeUnmapped != nil {
		if *b.includeUnmapped {
			values.Set("include_unmapped", "true")
		} else {
			values.Set("include_unmapped", "false")
		}
	}
	qs := values.Encode()
	if qs != "" {
		path += "?" + qs
	}

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
