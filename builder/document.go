package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Kirby980/go-es/client"
)

// DocumentBuilder 文档构建器
type DocumentBuilder struct {
	client  *client.Client
	index   string
	id      string
	doc     map[string]interface{}
	script  map[string]interface{}
	refresh string // refresh 参数: true, false, wait_for
}

// NewDocumentBuilder 创建文档构建器
func NewDocumentBuilder(c *client.Client, index string) *DocumentBuilder {
	return &DocumentBuilder{
		client: c,
		index:  index,
		doc:    make(map[string]interface{}),
	}
}

// ID 设置文档 ID
func (b *DocumentBuilder) ID(id string) *DocumentBuilder {
	b.id = id
	return b
}

// Set 设置字段值
func (b *DocumentBuilder) Set(field string, value interface{}) *DocumentBuilder {
	b.doc[field] = value
	return b
}

// SetMap 批量设置字段
func (b *DocumentBuilder) SetMap(data map[string]interface{}) *DocumentBuilder {
	for k, v := range data {
		b.doc[k] = v
	}
	return b
}

// SetStruct 从结构体设置
func (b *DocumentBuilder) SetStruct(data interface{}) *DocumentBuilder {
	jsonData, _ := json.Marshal(data)
	json.Unmarshal(jsonData, &b.doc)
	return b
}

// Refresh 设置刷新策略
// - "true": 立即刷新，操作后文档立即可见
// - "false": 不刷新，等待自动刷新（默认）
// - "wait_for": 等待刷新完成后再返回
func (b *DocumentBuilder) Refresh(refresh string) *DocumentBuilder {
	b.refresh = refresh
	return b
}

// buildPath 构建带查询参数的路径
func (b *DocumentBuilder) buildPath(basePath string) string {
	if b.refresh != "" {
		return fmt.Sprintf("%s?refresh=%s", basePath, b.refresh)
	}
	return basePath
}

// Script 设置脚本更新
func (b *DocumentBuilder) Script(source string, params map[string]interface{}) *DocumentBuilder {
	b.script = map[string]interface{}{
		"source": source,
		"lang":   "painless",
	}
	if params != nil {
		b.script["params"] = params
	}
	return b
}

// DocumentResponse 文档操作响应
type DocumentResponse struct {
	Index   string `json:"_index"`
	ID      string `json:"_id"`
	Version int    `json:"_version"`
	Result  string `json:"result"` // created, updated, deleted, noop
	Shards  struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Failed     int `json:"failed"`
	} `json:"_shards"`
}

// GetResponse 获取文档响应
type GetResponse struct {
	Index   string                 `json:"_index"`
	ID      string                 `json:"_id"`
	Version int                    `json:"_version"`
	Found   bool                   `json:"found"`
	Source  map[string]interface{} `json:"_source"`
}

// Do 索引文档（创建或更新）
func (b *DocumentBuilder) Do(ctx context.Context) (*DocumentResponse, error) {
	var path string
	var method string

	if b.id != "" {
		path = fmt.Sprintf("/%s/_doc/%s", b.index, b.id)
		method = http.MethodPut
	} else {
		path = fmt.Sprintf("/%s/_doc", b.index)
		method = http.MethodPost
	}

	// 添加查询参数
	path = b.buildPath(path)

	respBody, err := b.client.Do(ctx, method, path, b.doc)
	if err != nil {
		return nil, err
	}

	var resp DocumentResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &resp, nil
}

// Create 创建文档（如果已存在则失败）
func (b *DocumentBuilder) Create(ctx context.Context) (*DocumentResponse, error) {
	if b.id == "" {
		return nil, fmt.Errorf("创建文档需要指定 ID")
	}

	path := fmt.Sprintf("/%s/_create/%s", b.index, b.id)
	path = b.buildPath(path)

	respBody, err := b.client.Do(ctx, http.MethodPut, path, b.doc)
	if err != nil {
		return nil, err
	}

	var resp DocumentResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &resp, nil
}

// Update 更新文档（部分更新）
func (b *DocumentBuilder) Update(ctx context.Context) (*DocumentResponse, error) {
	if b.id == "" {
		return nil, fmt.Errorf("更新文档需要指定 ID")
	}

	path := fmt.Sprintf("/%s/_update/%s", b.index, b.id)
	path = b.buildPath(path)

	updateBody := make(map[string]interface{})
	if b.script != nil {
		updateBody["script"] = b.script
	} else {
		updateBody["doc"] = b.doc
	}

	respBody, err := b.client.Do(ctx, http.MethodPost, path, updateBody)
	if err != nil {
		return nil, err
	}

	var resp DocumentResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &resp, nil
}

// Upsert 更新或插入
func (b *DocumentBuilder) Upsert(ctx context.Context) (*DocumentResponse, error) {
	if b.id == "" {
		return nil, fmt.Errorf("upsert 需要指定 ID")
	}

	path := fmt.Sprintf("/%s/_update/%s", b.index, b.id)
	path = b.buildPath(path)

	updateBody := map[string]interface{}{
		"doc":           b.doc,
		"doc_as_upsert": true,
	}

	respBody, err := b.client.Do(ctx, http.MethodPost, path, updateBody)
	if err != nil {
		return nil, err
	}

	var resp DocumentResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &resp, nil
}

// Get 获取文档
func (b *DocumentBuilder) Get(ctx context.Context) (*GetResponse, error) {
	if b.id == "" {
		return nil, fmt.Errorf("获取文档需要指定 ID")
	}

	path := fmt.Sprintf("/%s/_doc/%s", b.index, b.id)
	respBody, err := b.client.Do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp GetResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &resp, nil
}

// Delete 删除文档
func (b *DocumentBuilder) Delete(ctx context.Context) (*DocumentResponse, error) {
	if b.id == "" {
		return nil, fmt.Errorf("删除文档需要指定 ID")
	}

	path := fmt.Sprintf("/%s/_doc/%s", b.index, b.id)
	path = b.buildPath(path)

	respBody, err := b.client.Do(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return nil, err
	}

	var resp DocumentResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &resp, nil
}

// Exists 检查文档是否存在
func (b *DocumentBuilder) Exists(ctx context.Context) (bool, error) {
	if b.id == "" {
		return false, fmt.Errorf("检查文档需要指定 ID")
	}

	path := fmt.Sprintf("/%s/_doc/%s", b.index, b.id)
	_, err := b.client.Do(ctx, http.MethodHead, path, nil)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// MGet 批量获取文档
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
