package builder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// IngestPipelineBuilder 是用于构建和执行 Elasticsearch IngestPipeline 操作的链式 API 构建器。
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

// NewIngestPipelineBuilder 创建并返回一个 IngestPipeline 构建器实例，用于构造和执行 Elasticsearch 的 IngestPipeline 请求。
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

// Header 设置自定义的 HTTP 请求头 (例如: Header("Content-Type", "application/json"))。此方法支持链式调用。
func (b *IngestPipelineBuilder) Header(key, value string) *IngestPipelineBuilder {
	b.baseBuilder.Header(key, value)
	return b
}

// Debug 开启当前请求的单次调试模式。开启后，请求的完整 HTTP URL、Method、Body 及响应将被打印到日志中，便于排查问题。
func (b *IngestPipelineBuilder) Debug() *IngestPipelineBuilder {
	b.setDebug(true)
	return b
}

// Description 为管道或资源添加易于阅读的描述信息。
func (b *IngestPipelineBuilder) Description(desc string) *IngestPipelineBuilder {
	b.description = desc
	return b
}

// Version 指定资源的版本号，常用于乐观锁并发控制或模板版本管理。
func (b *IngestPipelineBuilder) Version(version int) *IngestPipelineBuilder {
	b.version = &version
	return b
}

// Meta 设置附加的用户自定义元数据字典 (即 _meta 字段)。
func (b *IngestPipelineBuilder) Meta(meta map[string]any) *IngestPipelineBuilder {
	b.meta = meta
	return b
}

// AddProcessor 向摄取管道中添加一个数据处理器 (Processor，如 set, remove, grok 等)。
func (b *IngestPipelineBuilder) AddProcessor(name string, config map[string]any) *IngestPipelineBuilder {
	b.processors = append(b.processors, map[string]any{
		name: config,
	})
	return b
}

// AddOnFailure 向摄取管道中添加一个错误处理处理器，当主处理器失败时将触发此逻辑。
func (b *IngestPipelineBuilder) AddOnFailure(name string, config map[string]any) *IngestPipelineBuilder {
	b.onFailure = append(b.onFailure, map[string]any{
		name: config,
	})
	return b
}

// Pipeline 指定操作使用的摄取管道的名称。
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

// Build 根据当前构建器的状态组装请求体参数 (通常为 map[string]any 格式)，用于最终发送到 Elasticsearch。
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

// IngestAckResponse 定义了 Elasticsearch IngestAck 操作返回的结构化响应数据，包含元数据及核心结果。
type IngestAckResponse struct {
	Acknowledged bool `json:"acknowledged"`
}

// Put 发起 PUT 请求以创建或更新目标资源。
func (b *IngestPipelineBuilder) Put(ctx context.Context) (*IngestAckResponse, error) {
	if b.id == "" {
		return nil, fmt.Errorf("pipeline id 不能为空")
	}
	path := fmt.Sprintf("/_ingest/pipeline/%s", url.PathEscape(b.id))
	body := b.Build()

	if b.isDebug() {
		b.printDebug("PUT", path, body)
		defer b.autoResetDebug()
	}

	var resp IngestAckResponse
	if err := b.client.DoWithHeaderAndDecode(ctx, http.MethodPut, path, body, b.getHeaders(), &resp); err != nil {
		return nil, err
	}
	if b.isDebug() {
		b.printResponseObj(resp)
	}
	return &resp, nil
}

// Get 发起 GET 请求以获取目标资源的详细信息。
func (b *IngestPipelineBuilder) Get(ctx context.Context) (map[string]any, error) {
	if b.id == "" {
		return nil, fmt.Errorf("pipeline id 不能为空")
	}
	path := fmt.Sprintf("/_ingest/pipeline/%s", url.PathEscape(b.id))

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
func (b *IngestPipelineBuilder) Delete(ctx context.Context) (*IngestAckResponse, error) {
	if b.id == "" {
		return nil, fmt.Errorf("pipeline id 不能为空")
	}
	path := fmt.Sprintf("/_ingest/pipeline/%s", url.PathEscape(b.id))

	if b.isDebug() {
		b.printDebug("DELETE", path, nil)
		defer b.autoResetDebug()
	}

	var resp IngestAckResponse
	if err := b.client.DoWithHeaderAndDecode(ctx, http.MethodDelete, path, nil, b.getHeaders(), &resp); err != nil {
		return nil, err
	}
	if b.isDebug() {
		b.printResponseObj(resp)
	}
	return &resp, nil
}

// IngestSimulateResponse 定义了 Elasticsearch IngestSimulate 操作返回的结构化响应数据，包含元数据及核心结果。
type IngestSimulateResponse struct {
	Docs []map[string]any `json:"docs"`
}

// Simulate 对摄取管道 (Ingest Pipeline) 执行模拟运行，不会实际写入数据。
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
		defer b.autoResetDebug()
	}

	var resp IngestSimulateResponse
	if err := b.client.DoWithHeaderAndDecode(ctx, http.MethodPost, path, body, b.getHeaders(), &resp); err != nil {
		return nil, err
	}
	if b.isDebug() {
		b.printResponseObj(resp)
	}

	return &resp, nil
}

// SimulateNDJSON 以 NDJSON 格式对摄取管道执行批量模拟运行。
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
