package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Kirby980/go-es/client"
)

// MGetBuilder 批量获取文档构建器
type MGetBuilder struct {
	client *client.Client
	index  string
	ids    []string
}

// NewMGetBuilder 创建批量获取构建器
func NewMGetBuilder(c *client.Client, index string) *MGetBuilder {
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
	body := map[string]interface{}{
		"ids": b.ids,
	}

	respBody, err := b.client.Do(ctx, http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}

	var resp MGetResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &resp, nil
}
