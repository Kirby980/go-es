package builder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type IngestPipelineBuilder struct {
	client      ESClient
	id          string
	description string
	processors  []map[string]any
	onFailure   []map[string]any
	meta        map[string]any
	version     *int
	debugHelper
	baseBuilder
}

func NewIngestPipelineBuilder(c ESClient, id string) *IngestPipelineBuilder {
	b := &IngestPipelineBuilder{
		client:      c,
		id:          id,
		processors:  make([]map[string]any, 0),
		onFailure:   make([]map[string]any, 0),
		debugHelper: debugHelper{logger: c.GetLogger()},
	}
	b.initBaseBuilder()
	return b
}

func (b *IngestPipelineBuilder) Header(key, value string) *IngestPipelineBuilder {
	b.baseBuilder.Header(key, value)
	return b
}

func (b *IngestPipelineBuilder) Debug() *IngestPipelineBuilder {
	b.setDebug(true)
	return b
}

func (b *IngestPipelineBuilder) Description(desc string) *IngestPipelineBuilder {
	b.description = desc
	return b
}

func (b *IngestPipelineBuilder) Version(version int) *IngestPipelineBuilder {
	b.version = &version
	return b
}

func (b *IngestPipelineBuilder) Meta(meta map[string]any) *IngestPipelineBuilder {
	b.meta = meta
	return b
}

func (b *IngestPipelineBuilder) AddProcessor(name string, config map[string]any) *IngestPipelineBuilder {
	b.processors = append(b.processors, map[string]any{
		name: config,
	})
	return b
}

func (b *IngestPipelineBuilder) AddOnFailure(name string, config map[string]any) *IngestPipelineBuilder {
	b.onFailure = append(b.onFailure, map[string]any{
		name: config,
	})
	return b
}

func (b *IngestPipelineBuilder) Pipeline(pipeline map[string]any) *IngestPipelineBuilder {
	if pipeline == nil {
		return b
	}
	if v, ok := pipeline["description"].(string); ok {
		b.description = v
	}
	if v, ok := pipeline["processors"].([]map[string]any); ok {
		b.processors = v
	}
	if v, ok := pipeline["on_failure"].([]map[string]any); ok {
		b.onFailure = v
	}
	if v, ok := pipeline["_meta"].(map[string]any); ok {
		b.meta = v
	}
	if v, ok := pipeline["version"].(int); ok {
		b.version = &v
	}
	return b
}

func (b *IngestPipelineBuilder) Build() map[string]any {
	body := make(map[string]any)
	if b.description != "" {
		body["description"] = b.description
	}
	if len(b.processors) > 0 {
		body["processors"] = b.processors
	}
	if len(b.onFailure) > 0 {
		body["on_failure"] = b.onFailure
	}
	if b.meta != nil {
		body["_meta"] = b.meta
	}
	if b.version != nil {
		body["version"] = *b.version
	}
	return body
}

type IngestAckResponse struct {
	Acknowledged bool `json:"acknowledged"`
}

func (b *IngestPipelineBuilder) Put(ctx context.Context) (*IngestAckResponse, error) {
	if b.id == "" {
		return nil, fmt.Errorf("pipeline id 不能为空")
	}
	path := fmt.Sprintf("/_ingest/pipeline/%s", url.PathEscape(b.id))
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

	var resp IngestAckResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	return &resp, nil
}

func (b *IngestPipelineBuilder) Get(ctx context.Context) (map[string]any, error) {
	if b.id == "" {
		return nil, fmt.Errorf("pipeline id 不能为空")
	}
	path := fmt.Sprintf("/_ingest/pipeline/%s", url.PathEscape(b.id))

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

func (b *IngestPipelineBuilder) Delete(ctx context.Context) (*IngestAckResponse, error) {
	if b.id == "" {
		return nil, fmt.Errorf("pipeline id 不能为空")
	}
	path := fmt.Sprintf("/_ingest/pipeline/%s", url.PathEscape(b.id))

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

	var resp IngestAckResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	return &resp, nil
}

type IngestSimulateResponse struct {
	Docs []map[string]any `json:"docs"`
}

func (b *IngestPipelineBuilder) Simulate(ctx context.Context, docs []map[string]any, verbose bool) (*IngestSimulateResponse, error) {
	path := "/_ingest/pipeline/_simulate"
	if b.id != "" {
		path = fmt.Sprintf("/_ingest/pipeline/%s/_simulate", url.PathEscape(b.id))
	}
	if verbose {
		path += "?verbose=true"
	}

	body := map[string]any{
		"docs": docs,
	}
	if b.id == "" {
		p := b.Build()
		if len(p) > 0 {
			body["pipeline"] = p
		}
	}

	if b.isDebug() {
		b.printDebug("POST", path, body)
		defer b.setDebug(false)
	}

	respBody, err := b.client.DoWithHeader(ctx, http.MethodPost, path, body, b.getHeaders())
	if err != nil {
		return nil, err
	}

	if b.isDebug() {
		b.printResponse(respBody)
	}

	var resp IngestSimulateResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &resp, nil
}

func (b *IngestPipelineBuilder) SimulateNDJSON(ctx context.Context, docs []map[string]any) ([]byte, error) {
	path := "/_ingest/pipeline/_simulate"
	if b.id != "" {
		path = fmt.Sprintf("/_ingest/pipeline/%s/_simulate", url.PathEscape(b.id))
	}

	body := map[string]any{
		"docs": docs,
	}
	if b.id == "" {
		p := b.Build()
		if len(p) > 0 {
			body["pipeline"] = p
		}
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("序列化请求体失败: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, b.client.GetAddress()+path, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range b.getHeaders() {
		for _, vv := range v {
			req.Header.Add(k, vv)
		}
	}

	return b.client.DoRequest(ctx, req)
}
