package builder

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// IndexTemplateBuilder ...
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

// NewIndexTemplateBuilder ...
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

// Header ...
func (b *IndexTemplateBuilder) Header(key, value string) *IndexTemplateBuilder {
	b.baseBuilder.Header(key, value)
	return b
}

// Debug ...
func (b *IndexTemplateBuilder) Debug() *IndexTemplateBuilder {
	b.setDebug(true)
	return b
}

// IndexPatterns ...
func (b *IndexTemplateBuilder) IndexPatterns(patterns ...string) *IndexTemplateBuilder {
	b.indexPatterns = patterns
	return b
}

// ComposedOf ...
func (b *IndexTemplateBuilder) ComposedOf(names ...string) *IndexTemplateBuilder {
	b.composedOf = names
	return b
}

// Priority ...
func (b *IndexTemplateBuilder) Priority(priority int) *IndexTemplateBuilder {
	b.priority = &priority
	return b
}

// Version ...
func (b *IndexTemplateBuilder) Version(version int) *IndexTemplateBuilder {
	b.version = &version
	return b
}

// Meta ...
func (b *IndexTemplateBuilder) Meta(meta map[string]any) *IndexTemplateBuilder {
	b.meta = meta
	return b
}

// DataStream ...
func (b *IndexTemplateBuilder) DataStream(enable bool) *IndexTemplateBuilder {
	b.dataStream = enable
	return b
}

// TemplateSettings ...
func (b *IndexTemplateBuilder) TemplateSettings(settings map[string]any) *IndexTemplateBuilder {
	if settings != nil {
		b.template["settings"] = settings
	}
	return b
}

// TemplateMappings ...
func (b *IndexTemplateBuilder) TemplateMappings(mappings map[string]any) *IndexTemplateBuilder {
	if mappings != nil {
		b.template["mappings"] = mappings
	}
	return b
}

// TemplateAliases ...
func (b *IndexTemplateBuilder) TemplateAliases(aliases map[string]any) *IndexTemplateBuilder {
	if aliases != nil {
		b.template["aliases"] = aliases
	}
	return b
}

// Build ...
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

// PutTemplateResponse ...
type PutTemplateResponse struct {
	Acknowledged bool `json:"acknowledged"`
}

// Put ...
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

// Get ...
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

// Delete ...
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

// ComponentTemplateBuilder ...
type ComponentTemplateBuilder struct {
	client   ESClient
	name     string
	version  *int
	meta     map[string]any
	template map[string]any
	debugHelper
	baseBuilder
}

// NewComponentTemplateBuilder ...
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

// Header ...
func (b *ComponentTemplateBuilder) Header(key, value string) *ComponentTemplateBuilder {
	b.baseBuilder.Header(key, value)
	return b
}

// Debug ...
func (b *ComponentTemplateBuilder) Debug() *ComponentTemplateBuilder {
	b.setDebug(true)
	return b
}

// Version ...
func (b *ComponentTemplateBuilder) Version(version int) *ComponentTemplateBuilder {
	b.version = &version
	return b
}

// Meta ...
func (b *ComponentTemplateBuilder) Meta(meta map[string]any) *ComponentTemplateBuilder {
	b.meta = meta
	return b
}

// TemplateSettings ...
func (b *ComponentTemplateBuilder) TemplateSettings(settings map[string]any) *ComponentTemplateBuilder {
	if settings != nil {
		b.template["settings"] = settings
	}
	return b
}

// TemplateMappings ...
func (b *ComponentTemplateBuilder) TemplateMappings(mappings map[string]any) *ComponentTemplateBuilder {
	if mappings != nil {
		b.template["mappings"] = mappings
	}
	return b
}

// TemplateAliases ...
func (b *ComponentTemplateBuilder) TemplateAliases(aliases map[string]any) *ComponentTemplateBuilder {
	if aliases != nil {
		b.template["aliases"] = aliases
	}
	return b
}

// Build ...
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

// Put ...
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

// Get ...
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

// Delete ...
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
