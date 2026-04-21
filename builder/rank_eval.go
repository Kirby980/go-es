package builder

import (
	"context"
	"net/http"
	"net/url"
)

// RankEvalBuilder ...
type RankEvalBuilder struct {
	client ESClient
	index  string
	body   map[string]any
	debugHelper
	baseBuilder
}

// NewRankEvalBuilder ...
func NewRankEvalBuilder(c ESClient, index string) *RankEvalBuilder {
	b := &RankEvalBuilder{
		client: c,
		index:  index,
	}
	b.initBaseBuilder()
	return b
}

// Header ...
func (b *RankEvalBuilder) Header(key, value string) *RankEvalBuilder {
	b.baseBuilder.Header(key, value)
	return b
}

// Body ...
func (b *RankEvalBuilder) Body(body map[string]any) *RankEvalBuilder {
	b.body = body
	return b
}

// Debug ...
func (b *RankEvalBuilder) Debug(enable bool) *RankEvalBuilder {
	b.setDebugPersistent(enable)
	return b
}

// Do ...
func (b *RankEvalBuilder) Do(ctx context.Context) (map[string]any, error) {
	path := "/_rank_eval"
	if b.index != "" {
		path = "/" + url.PathEscape(b.index) + path
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
