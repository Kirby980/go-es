package builder

import (
	"context"
	"net/http"
	"net/url"
)

// TermvectorsBuilder ...
type TermvectorsBuilder struct {
	client ESClient
	index  string
	id     string
	body   map[string]any
	debugHelper
	baseBuilder
}

// NewTermvectorsBuilder ...
func NewTermvectorsBuilder(c ESClient, index string) *TermvectorsBuilder {
	b := &TermvectorsBuilder{
		client: c,
		index:  index,
	}
	b.initBaseBuilder()
	return b
}

// Header ...
func (b *TermvectorsBuilder) Header(key, value string) *TermvectorsBuilder {
	b.baseBuilder.Header(key, value)
	return b
}

// ID ...
func (b *TermvectorsBuilder) ID(id string) *TermvectorsBuilder {
	b.id = id
	return b
}

// Body ...
func (b *TermvectorsBuilder) Body(body map[string]any) *TermvectorsBuilder {
	b.body = body
	return b
}

// Debug ...
func (b *TermvectorsBuilder) Debug(enable bool) *TermvectorsBuilder {
	b.setDebugPersistent(enable)
	return b
}

// Do ...
func (b *TermvectorsBuilder) Do(ctx context.Context) (map[string]any, error) {
	path := "/" + url.PathEscape(b.index) + "/_termvectors"
	if b.id != "" {
		path += "/" + url.PathEscape(b.id)
	}

	if b.isDebug() {
		b.printDebug(http.MethodPost, path, b.body)
		defer b.autoResetDebug()
	}

	var resp map[string]any
	if err := b.client.DoWithHeaderAndDecode(ctx, http.MethodPost, path, b.body, b.getHeaders(), &resp); err != nil {
		return nil, err
	}

	if b.isDebug() {
		b.printResponseObj(resp)
	}

	return resp, nil
}
