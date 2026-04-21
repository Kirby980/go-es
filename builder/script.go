package builder

import (
	"context"
	"net/http"
	"net/url"
)

// ScriptBuilder ...
type ScriptBuilder struct {
	client ESClient
	id     string
	script map[string]any
	debugHelper
	baseBuilder
}

// NewScriptBuilder ...
func NewScriptBuilder(c ESClient) *ScriptBuilder {
	b := &ScriptBuilder{
		client: c,
	}
	b.initBaseBuilder()
	return b
}

// Header ...
func (b *ScriptBuilder) Header(key, value string) *ScriptBuilder {
	b.baseBuilder.Header(key, value)
	return b
}

// ID ...
func (b *ScriptBuilder) ID(id string) *ScriptBuilder {
	b.id = id
	return b
}

// Script ...
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

// Debug ...
func (b *ScriptBuilder) Debug(enable bool) *ScriptBuilder {
	b.setDebugPersistent(enable)
	return b
}

// Put ...
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

// Get ...
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

// Delete ...
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
