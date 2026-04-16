package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

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

func (b *IndexTemplateBuilder) Header(key, value string) *IndexTemplateBuilder {
	b.baseBuilder.Header(key, value)
	return b
}

func (b *IndexTemplateBuilder) Debug() *IndexTemplateBuilder {
	b.setDebug(true)
	return b
}

func (b *IndexTemplateBuilder) IndexPatterns(patterns ...string) *IndexTemplateBuilder {
	b.indexPatterns = patterns
	return b
}

func (b *IndexTemplateBuilder) ComposedOf(names ...string) *IndexTemplateBuilder {
	b.composedOf = names
	return b
}

func (b *IndexTemplateBuilder) Priority(priority int) *IndexTemplateBuilder {
	b.priority = &priority
	return b
}

func (b *IndexTemplateBuilder) Version(version int) *IndexTemplateBuilder {
	b.version = &version
	return b
}

func (b *IndexTemplateBuilder) Meta(meta map[string]any) *IndexTemplateBuilder {
	b.meta = meta
	return b
}

func (b *IndexTemplateBuilder) DataStream(enable bool) *IndexTemplateBuilder {
	b.dataStream = enable
	return b
}

func (b *IndexTemplateBuilder) TemplateSettings(settings map[string]any) *IndexTemplateBuilder {
	if settings != nil {
		b.template["settings"] = settings
	}
	return b
}

func (b *IndexTemplateBuilder) TemplateMappings(mappings map[string]any) *IndexTemplateBuilder {
	if mappings != nil {
		b.template["mappings"] = mappings
	}
	return b
}

func (b *IndexTemplateBuilder) TemplateAliases(aliases map[string]any) *IndexTemplateBuilder {
	if aliases != nil {
		b.template["aliases"] = aliases
	}
	return b
}

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

type PutTemplateResponse struct {
	Acknowledged bool `json:"acknowledged"`
}

func (b *IndexTemplateBuilder) Put(ctx context.Context) (*PutTemplateResponse, error) {
	if b.name == "" {
		return nil, fmt.Errorf("template name 不能为空")
	}
	path := fmt.Sprintf("/_index_template/%s", url.PathEscape(b.name))
	body := b.Build()

	if b.isDebug() {
		b.printDebug("PUT", path, body)
		defer b.setDebug(false)
	}

	respBody, err := b.client.DoWithHeader(ctx, http.MethodPut, path, body, b.getHeaders())
	if err != nil {
		return nil, err
	}

	if b.isDebug() {
		b.printResponse(respBody)
	}

	var resp PutTemplateResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	return &resp, nil
}

func (b *IndexTemplateBuilder) Get(ctx context.Context) (map[string]any, error) {
	if b.name == "" {
		return nil, fmt.Errorf("template name 不能为空")
	}
	path := fmt.Sprintf("/_index_template/%s", url.PathEscape(b.name))

	if b.isDebug() {
		b.printDebug("GET", path, nil)
		defer b.setDebug(false)
	}

	respBody, err := b.client.DoWithHeader(ctx, http.MethodGet, path, nil, b.getHeaders())
	if err != nil {
		return nil, err
	}
	if b.isDebug() {
		b.printResponse(respBody)
	}

	var resp map[string]any
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	return resp, nil
}

func (b *IndexTemplateBuilder) Delete(ctx context.Context) (*PutTemplateResponse, error) {
	if b.name == "" {
		return nil, fmt.Errorf("template name 不能为空")
	}
	path := fmt.Sprintf("/_index_template/%s", url.PathEscape(b.name))

	if b.isDebug() {
		b.printDebug("DELETE", path, nil)
		defer b.setDebug(false)
	}

	respBody, err := b.client.DoWithHeader(ctx, http.MethodDelete, path, nil, b.getHeaders())
	if err != nil {
		return nil, err
	}
	if b.isDebug() {
		b.printResponse(respBody)
	}

	var resp PutTemplateResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	return &resp, nil
}

type ComponentTemplateBuilder struct {
	client   ESClient
	name     string
	version  *int
	meta     map[string]any
	template map[string]any
	debugHelper
	baseBuilder
}

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

func (b *ComponentTemplateBuilder) Header(key, value string) *ComponentTemplateBuilder {
	b.baseBuilder.Header(key, value)
	return b
}

func (b *ComponentTemplateBuilder) Debug() *ComponentTemplateBuilder {
	b.setDebug(true)
	return b
}

func (b *ComponentTemplateBuilder) Version(version int) *ComponentTemplateBuilder {
	b.version = &version
	return b
}

func (b *ComponentTemplateBuilder) Meta(meta map[string]any) *ComponentTemplateBuilder {
	b.meta = meta
	return b
}

func (b *ComponentTemplateBuilder) TemplateSettings(settings map[string]any) *ComponentTemplateBuilder {
	if settings != nil {
		b.template["settings"] = settings
	}
	return b
}

func (b *ComponentTemplateBuilder) TemplateMappings(mappings map[string]any) *ComponentTemplateBuilder {
	if mappings != nil {
		b.template["mappings"] = mappings
	}
	return b
}

func (b *ComponentTemplateBuilder) TemplateAliases(aliases map[string]any) *ComponentTemplateBuilder {
	if aliases != nil {
		b.template["aliases"] = aliases
	}
	return b
}

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

func (b *ComponentTemplateBuilder) Put(ctx context.Context) (*PutTemplateResponse, error) {
	if b.name == "" {
		return nil, fmt.Errorf("template name 不能为空")
	}
	path := fmt.Sprintf("/_component_template/%s", url.PathEscape(b.name))
	body := b.Build()

	if b.isDebug() {
		b.printDebug("PUT", path, body)
		defer b.setDebug(false)
	}

	respBody, err := b.client.DoWithHeader(ctx, http.MethodPut, path, body, b.getHeaders())
	if err != nil {
		return nil, err
	}
	if b.isDebug() {
		b.printResponse(respBody)
	}

	var resp PutTemplateResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	return &resp, nil
}

func (b *ComponentTemplateBuilder) Get(ctx context.Context) (map[string]any, error) {
	if b.name == "" {
		return nil, fmt.Errorf("template name 不能为空")
	}
	path := fmt.Sprintf("/_component_template/%s", url.PathEscape(b.name))

	if b.isDebug() {
		b.printDebug("GET", path, nil)
		defer b.setDebug(false)
	}

	respBody, err := b.client.DoWithHeader(ctx, http.MethodGet, path, nil, b.getHeaders())
	if err != nil {
		return nil, err
	}
	if b.isDebug() {
		b.printResponse(respBody)
	}

	var resp map[string]any
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	return resp, nil
}

func (b *ComponentTemplateBuilder) Delete(ctx context.Context) (*PutTemplateResponse, error) {
	if b.name == "" {
		return nil, fmt.Errorf("template name 不能为空")
	}
	path := fmt.Sprintf("/_component_template/%s", url.PathEscape(b.name))

	if b.isDebug() {
		b.printDebug("DELETE", path, nil)
		defer b.setDebug(false)
	}

	respBody, err := b.client.DoWithHeader(ctx, http.MethodDelete, path, nil, b.getHeaders())
	if err != nil {
		return nil, err
	}
	if b.isDebug() {
		b.printResponse(respBody)
	}

	var resp PutTemplateResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	return &resp, nil
}
