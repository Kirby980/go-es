package builder

import (
	"context"
	"net/http"
	"net/url"
)

type SearchTemplateBuilder struct {
	client  ESClient
	index   string
	id      string
	source  any
	params  map[string]any
	explain *bool
	profile *bool
	debugHelper
	baseBuilder
}

func NewSearchTemplateBuilder(c ESClient) *SearchTemplateBuilder {
	b := &SearchTemplateBuilder{
		client: c,
	}
	b.initBaseBuilder()
	return b
}

func (b *SearchTemplateBuilder) Header(key, value string) *SearchTemplateBuilder {
	b.baseBuilder.Header(key, value)
	return b
}

func (b *SearchTemplateBuilder) Index(index string) *SearchTemplateBuilder {
	b.index = index
	return b
}

func (b *SearchTemplateBuilder) ID(id string) *SearchTemplateBuilder {
	b.id = id
	return b
}

func (b *SearchTemplateBuilder) Source(source any) *SearchTemplateBuilder {
	b.source = source
	return b
}

func (b *SearchTemplateBuilder) Params(params map[string]any) *SearchTemplateBuilder {
	b.params = params
	return b
}

func (b *SearchTemplateBuilder) Explain(enable bool) *SearchTemplateBuilder {
	b.explain = &enable
	return b
}

func (b *SearchTemplateBuilder) Profile(enable bool) *SearchTemplateBuilder {
	b.profile = &enable
	return b
}

func (b *SearchTemplateBuilder) Debug(enable bool) *SearchTemplateBuilder {
	b.setDebugPersistent(enable)
	return b
}

func (b *SearchTemplateBuilder) Do(ctx context.Context) (*SearchResponse, error) {
	path := "/_search/template"
	if b.index != "" {
		path = "/" + url.PathEscape(b.index) + path
	}

	body := make(map[string]any)
	if b.id != "" {
		body["id"] = b.id
	}
	if b.source != nil {
		body["source"] = b.source
	}
	if b.params != nil {
		body["params"] = b.params
	}
	if b.explain != nil {
		body["explain"] = *b.explain
	}
	if b.profile != nil {
		body["profile"] = *b.profile
	}

	if b.isDebug() {
		b.printDebug(http.MethodPost, path, body)
		defer b.autoResetDebug()
	}

	var resp SearchResponse
	if err := b.client.DoWithHeaderAndDecode(ctx, http.MethodPost, path, body, b.getHeaders(), &resp); err != nil {
		return nil, err
	}

	if b.isDebug() {
		b.printResponseObj(resp)
	}

	return &resp, nil
}
