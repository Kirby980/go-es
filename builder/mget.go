package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// MGetBuilder 批量获取文档构建器
type MGetBuilder struct {
	client ESClient
	index  string
	ids    []string
	DebugHelper
}

// NewMGetBuilder 创建批量获取构建器
func NewMGetBuilder(c ESClient, index string) *MGetBuilder {
	return &MGetBuilder{
		client: c,
		index:  index,
		ids:    make([]string, 0),
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
	path := fmt.Sprintf("/%s/_mget", b.index)
	body := map[string]any{
		"ids": b.ids,
	}
	if b.IsDebug() {
		b.PrintDebug("POST", path, body)
		defer b.SetDebug(false)
	}
	respBody, err := b.client.Do(ctx, http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}
	if b.IsDebug() {
		b.PrintResponse(respBody)
	}
	var resp MGetResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &resp, nil
}
