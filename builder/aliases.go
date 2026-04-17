package builder

import (
	"context"
	"fmt"
	"net/http"
)

type AliasesBuilder struct {
	client  ESClient
	actions []map[string]any
	debugHelper
	baseBuilder
}

func NewAliasesBuilder(c ESClient) *AliasesBuilder {
	b := &AliasesBuilder{
		client:      c,
		actions:     make([]map[string]any, 0),
		debugHelper: debugHelper{logger: c.GetLogger()},
	}
	b.initBaseBuilder()
	return b
}

func (b *AliasesBuilder) Header(key, value string) *AliasesBuilder {
	b.baseBuilder.Header(key, value)
	return b
}

func (b *AliasesBuilder) Debug() *AliasesBuilder {
	b.setDebug(true)
	return b
}

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

func (b *AliasesBuilder) RemoveIndex(index string) *AliasesBuilder {
	b.actions = append(b.actions, map[string]any{
		"remove_index": map[string]any{
			"index": index,
		},
	})
	return b
}

func (b *AliasesBuilder) Replace(alias string, fromIndex string, toIndex string, opts ...func(map[string]any)) *AliasesBuilder {
	b.Remove(fromIndex, alias)
	b.Add(toIndex, alias, opts...)
	return b
}

func WithIsWriteIndex(isWrite bool) func(map[string]any) {
	return func(m map[string]any) {
		m["is_write_index"] = isWrite
	}
}

func WithAliasFilter(filter map[string]any) func(map[string]any) {
	return func(m map[string]any) {
		if filter != nil {
			m["filter"] = filter
		}
	}
}

func WithRouting(routing string) func(map[string]any) {
	return func(m map[string]any) {
		if routing != "" {
			m["routing"] = routing
		}
	}
}

type AliasesResponse struct {
	Acknowledged bool `json:"acknowledged"`
}

func (b *AliasesBuilder) Build() map[string]any {
	return map[string]any{
		"actions": b.actions,
	}
}

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
