package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type DataStreamBuilder struct {
	client ESClient
	name   string
	debugHelper
	baseBuilder
}

func NewDataStreamBuilder(c ESClient, name string) *DataStreamBuilder {
	b := &DataStreamBuilder{
		client:      c,
		name:        name,
		debugHelper: debugHelper{logger: c.GetLogger()},
	}
	b.initBaseBuilder()
	return b
}

func (b *DataStreamBuilder) Header(key, value string) *DataStreamBuilder {
	b.baseBuilder.Header(key, value)
	return b
}

func (b *DataStreamBuilder) Debug() *DataStreamBuilder {
	b.setDebug(true)
	return b
}

type DataStreamAckResponse struct {
	Acknowledged bool `json:"acknowledged"`
}

func (b *DataStreamBuilder) Create(ctx context.Context) (*DataStreamAckResponse, error) {
	if b.name == "" {
		return nil, fmt.Errorf("data stream name 不能为空")
	}
	path := fmt.Sprintf("/_data_stream/%s", url.PathEscape(b.name))

	if b.isDebug() {
		b.printDebug("PUT", path, nil)
		defer b.setDebug(false)
	}

	respBody, err := b.client.DoWithHeader(ctx, http.MethodPut, path, nil, b.getHeaders())
	if err != nil {
		return nil, err
	}
	if b.isDebug() {
		b.printResponse(respBody)
	}

	var resp DataStreamAckResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	return &resp, nil
}

func (b *DataStreamBuilder) Get(ctx context.Context) (map[string]any, error) {
	path := "/_data_stream"
	if b.name != "" {
		path = fmt.Sprintf("/_data_stream/%s", url.PathEscape(b.name))
	}

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

func (b *DataStreamBuilder) Delete(ctx context.Context) (*DataStreamAckResponse, error) {
	if b.name == "" {
		return nil, fmt.Errorf("data stream name 不能为空")
	}
	path := fmt.Sprintf("/_data_stream/%s", url.PathEscape(b.name))

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

	var resp DataStreamAckResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	return &resp, nil
}
