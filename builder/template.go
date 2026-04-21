package builder

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// IndexTemplateBuilder 是用于构建和执行 Elasticsearch IndexTemplate 操作的链式 API 构建器。
type IndexTemplateBuilder struct {
	client        ESClient
	name          string
	indexPatterns []string
	composedOf    []string
	priority      *int
	version       *int
	meta          map[string]any
	template      map[string]any
	dataStream    bool
	debugHelper
	baseBuilder
}

// NewIndexTemplateBuilder 创建并返回一个 IndexTemplate 构建器实例，用于构造和执行 Elasticsearch 的 IndexTemplate 请求。
func NewIndexTemplateBuilder(c ESClient, name string) *IndexTemplateBuilder {
	b := &IndexTemplateBuilder{
		client:      c,
		name:        name,
		template:    make(map[string]any),
		debugHelper: debugHelper{logger: c.GetLogger()},
	}
	b.initBaseBuilder()
	return b
}

// Header 设置自定义的 HTTP 请求头 (例如: Header("Content-Type", "application/json"))。此方法支持链式调用。
func (b *IndexTemplateBuilder) Header(key, value string) *IndexTemplateBuilder {
	b.baseBuilder.Header(key, value)
	return b
}

// Debug 开启当前请求的单次调试模式。开启后，请求的完整 HTTP URL、Method、Body 及响应将被打印到日志中，便于排查问题。
func (b *IndexTemplateBuilder) Debug() *IndexTemplateBuilder {
	b.setDebug(true)
	return b
}

// IndexPatterns 指定模板适用的索引名称模式 (支持通配符，如 "logs-*")。
func (b *IndexTemplateBuilder) IndexPatterns(patterns ...string) *IndexTemplateBuilder {
	b.indexPatterns = patterns
	return b
}

// ComposedOf 指定当前索引模板组合使用的组件模板名称列表。
func (b *IndexTemplateBuilder) ComposedOf(names ...string) *IndexTemplateBuilder {
	b.composedOf = names
	return b
}

// Priority 指定模板或组件的优先级，优先级高的规则将覆盖优先级低的规则。
func (b *IndexTemplateBuilder) Priority(priority int) *IndexTemplateBuilder {
	b.priority = &priority
	return b
}

// Version 指定资源的版本号，常用于乐观锁并发控制或模板版本管理。
func (b *IndexTemplateBuilder) Version(version int) *IndexTemplateBuilder {
	b.version = &version
	return b
}

// Meta 设置附加的用户自定义元数据字典 (即 _meta 字段)。
func (b *IndexTemplateBuilder) Meta(meta map[string]any) *IndexTemplateBuilder {
	b.meta = meta
	return b
}

// DataStream 声明该模板用于创建数据流 (Data Stream)，而非普通索引。
func (b *IndexTemplateBuilder) DataStream(enable bool) *IndexTemplateBuilder {
	b.dataStream = enable
	return b
}

// TemplateSettings 配置模板中的 settings 部分。
func (b *IndexTemplateBuilder) TemplateSettings(settings map[string]any) *IndexTemplateBuilder {
	if settings != nil {
		b.template["settings"] = settings
	}
	return b
}

// TemplateMappings 配置模板中的 mappings 映射规则。
func (b *IndexTemplateBuilder) TemplateMappings(mappings map[string]any) *IndexTemplateBuilder {
	if mappings != nil {
		b.template["mappings"] = mappings
	}
	return b
}

// TemplateAliases 配置模板中包含的别名。
func (b *IndexTemplateBuilder) TemplateAliases(aliases map[string]any) *IndexTemplateBuilder {
	if aliases != nil {
		b.template["aliases"] = aliases
	}
	return b
}

// Build 根据当前构建器的状态组装请求体参数 (通常为 map[string]any 格式)，用于最终发送到 Elasticsearch。
func (b *IndexTemplateBuilder) Build() map[string]any {
	body := make(map[string]any)
	if len(b.indexPatterns) > 0 {
		body["index_patterns"] = b.indexPatterns
	}
	if len(b.composedOf) > 0 {
		body["composed_of"] = b.composedOf
	}
	if b.priority != nil {
		body["priority"] = *b.priority
	}
	if b.version != nil {
		body["version"] = *b.version
	}
	if b.meta != nil {
		body["_meta"] = b.meta
	}
	if len(b.template) > 0 {
		body["template"] = b.template
	}
	if b.dataStream {
		body["data_stream"] = map[string]any{}
	}
	return body
}

// PutTemplateResponse 定义了 Elasticsearch PutTemplate 操作返回的结构化响应数据，包含元数据及核心结果。
type PutTemplateResponse struct {
	Acknowledged bool `json:"acknowledged"`
}

// Put 发起 PUT 请求以创建或更新目标资源。
func (b *IndexTemplateBuilder) Put(ctx context.Context) (*PutTemplateResponse, error) {
	if b.name == "" {
		return nil, fmt.Errorf("template name 不能为空")
	}
	path := fmt.Sprintf("/_index_template/%s", url.PathEscape(b.name))
	body := b.Build()

	if b.isDebug() {
		b.printDebug("PUT", path, body)
		defer b.autoResetDebug()
	}

	var resp PutTemplateResponse
	if err := b.client.DoWithHeaderAndDecode(ctx, http.MethodPut, path, body, b.getHeaders(), &resp); err != nil {
		return nil, err
	}
	if b.isDebug() {
		b.printResponseObj(resp)
	}
	return &resp, nil
}

// Get 发起 GET 请求以获取目标资源的详细信息。
func (b *IndexTemplateBuilder) Get(ctx context.Context) (map[string]any, error) {
	if b.name == "" {
		return nil, fmt.Errorf("template name 不能为空")
	}
	path := fmt.Sprintf("/_index_template/%s", url.PathEscape(b.name))

	if b.isDebug() {
		b.printDebug("GET", path, nil)
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
func (b *IndexTemplateBuilder) Delete(ctx context.Context) (*PutTemplateResponse, error) {
	if b.name == "" {
		return nil, fmt.Errorf("template name 不能为空")
	}
	path := fmt.Sprintf("/_index_template/%s", url.PathEscape(b.name))

	if b.isDebug() {
		b.printDebug("DELETE", path, nil)
		defer b.autoResetDebug()
	}

	var resp PutTemplateResponse
	if err := b.client.DoWithHeaderAndDecode(ctx, http.MethodDelete, path, nil, b.getHeaders(), &resp); err != nil {
		return nil, err
	}
	if b.isDebug() {
		b.printResponseObj(resp)
	}
	return &resp, nil
}

// ComponentTemplateBuilder 是用于构建和执行 Elasticsearch ComponentTemplate 操作的链式 API 构建器。
type ComponentTemplateBuilder struct {
	client   ESClient
	name     string
	version  *int
	meta     map[string]any
	template map[string]any
	debugHelper
	baseBuilder
}

// NewComponentTemplateBuilder 创建并返回一个 ComponentTemplate 构建器实例，用于构造和执行 Elasticsearch 的 ComponentTemplate 请求。
func NewComponentTemplateBuilder(c ESClient, name string) *ComponentTemplateBuilder {
	b := &ComponentTemplateBuilder{
		client:      c,
		name:        name,
		template:    make(map[string]any),
		debugHelper: debugHelper{logger: c.GetLogger()},
	}
	b.initBaseBuilder()
	return b
}

// Header 设置自定义的 HTTP 请求头 (例如: Header("Content-Type", "application/json"))。此方法支持链式调用。
func (b *ComponentTemplateBuilder) Header(key, value string) *ComponentTemplateBuilder {
	b.baseBuilder.Header(key, value)
	return b
}

// Debug 开启当前请求的单次调试模式。开启后，请求的完整 HTTP URL、Method、Body 及响应将被打印到日志中，便于排查问题。
func (b *ComponentTemplateBuilder) Debug() *ComponentTemplateBuilder {
	b.setDebug(true)
	return b
}

// Version 指定资源的版本号，常用于乐观锁并发控制或模板版本管理。
func (b *ComponentTemplateBuilder) Version(version int) *ComponentTemplateBuilder {
	b.version = &version
	return b
}

// Meta 设置附加的用户自定义元数据字典 (即 _meta 字段)。
func (b *ComponentTemplateBuilder) Meta(meta map[string]any) *ComponentTemplateBuilder {
	b.meta = meta
	return b
}

// TemplateSettings 配置模板中的 settings 部分。
func (b *ComponentTemplateBuilder) TemplateSettings(settings map[string]any) *ComponentTemplateBuilder {
	if settings != nil {
		b.template["settings"] = settings
	}
	return b
}

// TemplateMappings 配置模板中的 mappings 映射规则。
func (b *ComponentTemplateBuilder) TemplateMappings(mappings map[string]any) *ComponentTemplateBuilder {
	if mappings != nil {
		b.template["mappings"] = mappings
	}
	return b
}

// TemplateAliases 配置模板中包含的别名。
func (b *ComponentTemplateBuilder) TemplateAliases(aliases map[string]any) *ComponentTemplateBuilder {
	if aliases != nil {
		b.template["aliases"] = aliases
	}
	return b
}

// Build 根据当前构建器的状态组装请求体参数 (通常为 map[string]any 格式)，用于最终发送到 Elasticsearch。
func (b *ComponentTemplateBuilder) Build() map[string]any {
	body := make(map[string]any)
	body["template"] = b.template
	if b.version != nil {
		body["version"] = *b.version
	}
	if b.meta != nil {
		body["_meta"] = b.meta
	}
	return body
}

// Put 发起 PUT 请求以创建或更新目标资源。
func (b *ComponentTemplateBuilder) Put(ctx context.Context) (*PutTemplateResponse, error) {
	if b.name == "" {
		return nil, fmt.Errorf("template name 不能为空")
	}
	path := fmt.Sprintf("/_component_template/%s", url.PathEscape(b.name))
	body := b.Build()

	if b.isDebug() {
		b.printDebug("PUT", path, body)
		defer b.autoResetDebug()
	}

	var resp PutTemplateResponse
	if err := b.client.DoWithHeaderAndDecode(ctx, http.MethodPut, path, body, b.getHeaders(), &resp); err != nil {
		return nil, err
	}
	if b.isDebug() {
		b.printResponseObj(resp)
	}
	return &resp, nil
}

// Get 发起 GET 请求以获取目标资源的详细信息。
func (b *ComponentTemplateBuilder) Get(ctx context.Context) (map[string]any, error) {
	if b.name == "" {
		return nil, fmt.Errorf("template name 不能为空")
	}
	path := fmt.Sprintf("/_component_template/%s", url.PathEscape(b.name))

	if b.isDebug() {
		b.printDebug("GET", path, nil)
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
func (b *ComponentTemplateBuilder) Delete(ctx context.Context) (*PutTemplateResponse, error) {
	if b.name == "" {
		return nil, fmt.Errorf("template name 不能为空")
	}
	path := fmt.Sprintf("/_component_template/%s", url.PathEscape(b.name))

	if b.isDebug() {
		b.printDebug("DELETE", path, nil)
		defer b.autoResetDebug()
	}

	var resp PutTemplateResponse
	if err := b.client.DoWithHeaderAndDecode(ctx, http.MethodDelete, path, nil, b.getHeaders(), &resp); err != nil {
		return nil, err
	}
	if b.isDebug() {
		b.printResponseObj(resp)
	}
	return &resp, nil
}
