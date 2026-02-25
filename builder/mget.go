package builder

import (
	"context"
	"encoding/json"
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
}

// NewMGetBuilder 创建批量获取构建器
func NewMGetBuilder(c ESClient, index string) *MGetBuilder {
	return &MGetBuilder{
		client:      c,
		index:       index,
		ids:         make([]string, 0),
		debugHelper: debugHelper{logger: c.GetLogger()},
	}
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
		defer b.setDebug(false)
	}
	respBody, err := b.client.Do(ctx, http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}
	if b.isDebug() {
		b.printResponse(respBody)
	}
	var resp MGetResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &resp, nil
}
