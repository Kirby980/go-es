package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go-es/client"
)

// DocumentAPI 文档操作 API
type DocumentAPI struct {
	client *client.Client
}

// NewDocumentAPI 创建文档 API
func NewDocumentAPI(c *client.Client) *DocumentAPI {
	return &DocumentAPI{client: c}
}

// IndexResponse 索引文档响应
type IndexResponse struct {
	Index   string `json:"_index"`
	ID      string `json:"_id"`
	Version int    `json:"_version"`
	Result  string `json:"result"`
}

// Index 索引文档（创建或更新）
func (a *DocumentAPI) Index(ctx context.Context, indexName, docID string, doc interface{}) (*IndexResponse, error) {
	var path string
	if docID != "" {
		path = fmt.Sprintf("/%s/_doc/%s", indexName, docID)
	} else {
		path = fmt.Sprintf("/%s/_doc", indexName)
	}

	respBody, err := a.client.Do(ctx, http.MethodPost, path, doc)
	if err != nil {
		return nil, err
	}

	var resp IndexResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &resp, nil
}

// GetResponse 获取文档响应
type GetResponse struct {
	Index  string                 `json:"_index"`
	ID     string                 `json:"_id"`
	Found  bool                   `json:"found"`
	Source map[string]interface{} `json:"_source"`
}

// Get 获取文档
func (a *DocumentAPI) Get(ctx context.Context, indexName, docID string) (*GetResponse, error) {
	path := fmt.Sprintf("/%s/_doc/%s", indexName, docID)
	respBody, err := a.client.Do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp GetResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &resp, nil
}

// Update 更新文档
func (a *DocumentAPI) Update(ctx context.Context, indexName, docID string, doc map[string]interface{}) (*IndexResponse, error) {
	path := fmt.Sprintf("/%s/_update/%s", indexName, docID)

	body := map[string]interface{}{
		"doc": doc,
	}

	respBody, err := a.client.Do(ctx, http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}

	var resp IndexResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &resp, nil
}

// DeleteResponse 删除文档响应
type DeleteResponse struct {
	Index   string `json:"_index"`
	ID      string `json:"_id"`
	Version int    `json:"_version"`
	Result  string `json:"result"`
}

// Delete 删除文档
func (a *DocumentAPI) Delete(ctx context.Context, indexName, docID string) (*DeleteResponse, error) {
	path := fmt.Sprintf("/%s/_doc/%s", indexName, docID)
	respBody, err := a.client.Do(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return nil, err
	}

	var resp DeleteResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &resp, nil
}

// BulkRequest 批量操作请求
type BulkRequest struct {
	Operations []map[string]interface{}
}

// BulkResponse 批量操作响应
type BulkResponse struct {
	Took   int                      `json:"took"`
	Errors bool                     `json:"errors"`
	Items  []map[string]interface{} `json:"items"`
}

// Bulk 批量操作
func (a *DocumentAPI) Bulk(ctx context.Context, indexName string, req *BulkRequest) (*BulkResponse, error) {
	path := fmt.Sprintf("/%s/_bulk", indexName)
	respBody, err := a.client.Do(ctx, http.MethodPost, path, req.Operations)
	if err != nil {
		return nil, err
	}

	var resp BulkResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &resp, nil
}
