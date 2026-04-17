package builder

import (
	"context"
	"net/http"
	"net/url"
)

type ILMBuilder struct {
	client ESClient
	name   string
	policy map[string]any
	debugHelper
	baseBuilder
}

func NewILMBuilder(client ESClient) *ILMBuilder {
	return &ILMBuilder{
		client: client,
	}
}

func (b *ILMBuilder) Name(name string) *ILMBuilder {
	b.name = name
	return b
}

func (b *ILMBuilder) Policy(policy map[string]any) *ILMBuilder {
	b.policy = policy
	return b
}

func (b *ILMBuilder) Debug(enable bool) *ILMBuilder {
	b.setDebug(enable)
	return b
}

func (b *ILMBuilder) Put(ctx context.Context) (*AcknowledgedResponse, error) {
	path := "/_ilm/policy/" + url.PathEscape(b.name)
	body := map[string]any{
		"policy": b.policy,
	}

	if b.isDebug() {
		b.printDebug("PUT", path, body)
		defer b.setDebug(false)
	}

	var resp AcknowledgedResponse
	err := b.client.DoWithHeaderAndDecode(ctx, http.MethodPut, path, body, b.getHeaders(), &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (b *ILMBuilder) Get(ctx context.Context) (map[string]any, error) {
	path := "/_ilm/policy"
	if b.name != "" {
		path += "/" + url.PathEscape(b.name)
	}

	if b.isDebug() {
		b.printDebug("GET", path, nil)
		defer b.setDebug(false)
	}

	var resp map[string]any
	err := b.client.DoWithHeaderAndDecode(ctx, http.MethodGet, path, nil, b.getHeaders(), &resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (b *ILMBuilder) Delete(ctx context.Context) (*AcknowledgedResponse, error) {
	path := "/_ilm/policy/" + url.PathEscape(b.name)

	if b.isDebug() {
		b.printDebug("DELETE", path, nil)
		defer b.setDebug(false)
	}

	var resp AcknowledgedResponse
	err := b.client.DoWithHeaderAndDecode(ctx, http.MethodDelete, path, nil, b.getHeaders(), &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
