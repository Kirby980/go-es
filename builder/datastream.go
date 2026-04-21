package builder

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// DataStreamBuilder 是用于构建和执行 Elasticsearch DataStream 操作的链式 API 构建器。
type DataStreamBuilder struct {
	client ESClient
	name   string
	debugHelper
	baseBuilder
}

// NewDataStreamBuilder 创建并返回一个 DataStream 构建器实例，用于构造和执行 Elasticsearch 的 DataStream 请求。
func NewDataStreamBuilder(c ESClient, name string) *DataStreamBuilder {
	b := &DataStreamBuilder{
		client:      c,
		name:        name,
		debugHelper: debugHelper{logger: c.GetLogger()},
	}
	b.initBaseBuilder()
	return b
}

// Header 设置自定义的 HTTP 请求头 (例如: Header("Content-Type", "application/json"))。此方法支持链式调用。
func (b *DataStreamBuilder) Header(key, value string) *DataStreamBuilder {
	b.baseBuilder.Header(key, value)
	return b
}

// Debug 开启当前请求的单次调试模式。开启后，请求的完整 HTTP URL、Method、Body 及响应将被打印到日志中，便于排查问题。
func (b *DataStreamBuilder) Debug() *DataStreamBuilder {
	b.setDebug(true)
	return b
}

// DataStreamAckResponse 定义了 Elasticsearch DataStreamAck 操作返回的结构化响应数据，包含元数据及核心结果。
type DataStreamAckResponse struct {
	Acknowledged bool `json:"acknowledged"`
}

// Create 执行创建操作。若资源已存在，可能会返回冲突错误。
func (b *DataStreamBuilder) Create(ctx context.Context) (*DataStreamAckResponse, error) {
	if b.name == "" {
		return nil, fmt.Errorf("data stream name 不能为空")
	}
	path := fmt.Sprintf("/_data_stream/%s", url.PathEscape(b.name))

	if b.isDebug() {
		b.printDebug("PUT", path, nil)
		defer b.autoResetDebug()
	}

	var resp DataStreamAckResponse
	if err := b.client.DoWithHeaderAndDecode(ctx, http.MethodPut, path, nil, b.getHeaders(), &resp); err != nil {
		return nil, err
	}
	if b.isDebug() {
		b.printResponseObj(resp)
	}
	return &resp, nil
}

// Get 发起 GET 请求以获取目标资源的详细信息。
func (b *DataStreamBuilder) Get(ctx context.Context) (map[string]any, error) {
	path := "/_data_stream"
	if b.name != "" {
		path = fmt.Sprintf("/_data_stream/%s", url.PathEscape(b.name))
	}

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
func (b *DataStreamBuilder) Delete(ctx context.Context) (*DataStreamAckResponse, error) {
	if b.name == "" {
		return nil, fmt.Errorf("data stream name 不能为空")
	}
	path := fmt.Sprintf("/_data_stream/%s", url.PathEscape(b.name))

	if b.isDebug() {
		b.printDebug("DELETE", path, nil)
		defer b.autoResetDebug()
	}

	var resp DataStreamAckResponse
	if err := b.client.DoWithHeaderAndDecode(ctx, http.MethodDelete, path, nil, b.getHeaders(), &resp); err != nil {
		return nil, err
	}
	if b.isDebug() {
		b.printResponseObj(resp)
	}
	return &resp, nil
}
