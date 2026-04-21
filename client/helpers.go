package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// CreateIndexRaw 提供了一种便捷方法，允许直接传递原始的 JSON 字符串 (如 {"settings": {...}, "mappings": {...}}) 来创建 Elasticsearch 索引，内部自动处理序列化，避免二次转义。
func (c *Client) CreateIndexRaw(ctx context.Context, index string, rawJSON string) error {
	if index == "" {
		return fmt.Errorf("index 不能为空")
	}
	if rawJSON == "" {
		return fmt.Errorf("rawJSON 不能为空")
	}
	rawBytes := []byte(rawJSON)
	if !json.Valid(rawBytes) {
		return fmt.Errorf("rawJSON 不是合法的 JSON")
	}
	path := "/" + url.PathEscape(index)
	_, err := c.DoWithHeader(ctx, http.MethodPut, path, json.RawMessage(rawBytes), nil)
	return err
}
