package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Kirby980/go-es/client"
)

// IndexAPI 索引操作 API
type IndexAPI struct {
	client *client.Client
}

// NewIndexAPI 创建索引 API
func NewIndexAPI(c *client.Client) *IndexAPI {
	return &IndexAPI{client: c}
}

// CreateRequest 创建索引请求
type CreateRequest struct {
	Settings map[string]interface{} `json:"settings,omitempty"`
	Mappings map[string]interface{} `json:"mappings,omitempty"`
	Aliases  map[string]interface{} `json:"aliases,omitempty"`
}

// Create 创建索引
func (a *IndexAPI) Create(ctx context.Context, indexName string, req *CreateRequest) error {
	path := fmt.Sprintf("/%s", indexName)
	_, err := a.client.Do(ctx, http.MethodPut, path, req)
	return err
}

// Delete 删除索引
func (a *IndexAPI) Delete(ctx context.Context, indexName string) error {
	path := fmt.Sprintf("/%s", indexName)
	_, err := a.client.Do(ctx, http.MethodDelete, path, nil)
	return err
}

// Exists 检查索引是否存在
func (a *IndexAPI) Exists(ctx context.Context, indexName string) (bool, error) {
	path := fmt.Sprintf("/%s", indexName)
	_, err := a.client.Do(ctx, http.MethodHead, path, nil)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// IndexInfo 索引信息
type IndexInfo struct {
	Aliases  map[string]interface{} `json:"aliases"`
	Mappings map[string]interface{} `json:"mappings"`
	Settings map[string]interface{} `json:"settings"`
}

// Get 获取索引信息
func (a *IndexAPI) Get(ctx context.Context, indexName string) (*IndexInfo, error) {
	path := fmt.Sprintf("/%s", indexName)
	respBody, err := a.client.Do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var result map[string]*IndexInfo
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if info, ok := result[indexName]; ok {
		return info, nil
	}

	return nil, fmt.Errorf("索引 %s 不存在", indexName)
}

// UpdateMappingsRequest 更新映射请求
type UpdateMappingsRequest struct {
	Properties map[string]interface{} `json:"properties"`
}

// UpdateMappings 更新索引映射
func (a *IndexAPI) UpdateMappings(ctx context.Context, indexName string, req *UpdateMappingsRequest) error {
	path := fmt.Sprintf("/%s/_mapping", indexName)
	_, err := a.client.Do(ctx, http.MethodPut, path, req)
	return err
}

// Refresh 刷新索引
func (a *IndexAPI) Refresh(ctx context.Context, indexName string) error {
	path := fmt.Sprintf("/%s/_refresh", indexName)
	_, err := a.client.Do(ctx, http.MethodPost, path, nil)
	return err
}
