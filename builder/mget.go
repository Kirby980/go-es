package builder

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// MGetBuilder 批量获取文档构建器
type MGetBuilder struct {
	client ESClient
	index  string
	ids    []string
	debugHelper
	baseBuilder
}

// NewMGetBuilder 创建批量获取构建器
func NewMGetBuilder(c ESClient, index string) *MGetBuilder {
	b := &MGetBuilder{
		client:      c,
		index:       index,
		ids:         make([]string, 0),
		debugHelper: debugHelper{logger: c.GetLogger()},
	}
	b.initBaseBuilder()
	return b
}

// Header 设置自定义 Header (链式调用)
func (b *MGetBuilder) Header(key, value string) *MGetBuilder {
	b.baseBuilder.Header(key, value)
	return b
}

// IDs 设置要获取的文档 ID 列表
func (b *MGetBuilder) IDs(ids ...string) *MGetBuilder {
	b.ids = append(b.ids, ids...)
	return b
}

// MGetResponse 批量获取响应
type MGetResponse struct {
	Docs []GetResponse `json:"docs"`
}

// Do 执行批量获取
func (b *MGetBuilder) Do(ctx context.Context) (*MGetResponse, error) {
	path := fmt.Sprintf("/%s/_mget", url.PathEscape(b.index))
	body := map[string]any{
		"ids": b.ids,
	}
	if b.isDebug() {
		b.printDebug("POST", path, body)
		defer b.autoResetDebug()
	}
	var resp MGetResponse
	if err := b.client.DoWithHeaderAndDecode(ctx, http.MethodPost, path, body, b.getHeaders(), &resp); err != nil {
		return nil, err
	}
	if b.isDebug() {
		b.printResponseObj(resp)
	}

	return &resp, nil
}
