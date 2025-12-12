package builder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Kirby980/go-es/client"
)

// BulkBuilder 批量操作构建器
type BulkBuilder struct {
	client     *client.Client
	index      string
	operations []bulkOperation
	debug      bool // 调试模式标志
}

// bulkOperation 批量操作项
type bulkOperation struct {
	action string
	meta   map[string]interface{}
	doc    map[string]interface{}
}

// NewBulkBuilder 创建批量操作构建器
func NewBulkBuilder(c *client.Client) *BulkBuilder {
	return &BulkBuilder{
		client:     c,
		operations: make([]bulkOperation, 0),
	}
}

// Index 设置默认索引
func (b *BulkBuilder) Index(index string) *BulkBuilder {
	b.index = index
	return b
}

// Add 添加索引操作
func (b *BulkBuilder) Add(index, id string, doc map[string]interface{}) *BulkBuilder {
	if index == "" {
		index = b.index
	}

	meta := map[string]interface{}{
		"_index": index,
	}
	if id != "" {
		meta["_id"] = id
	}

	b.operations = append(b.operations, bulkOperation{
		action: "index",
		meta:   meta,
		doc:    doc,
	})
	return b
}

// Create 添加创建操作（文档已存在则失败）
func (b *BulkBuilder) Create(index, id string, doc map[string]interface{}) *BulkBuilder {
	if index == "" {
		index = b.index
	}

	meta := map[string]interface{}{
		"_index": index,
		"_id":    id,
	}

	b.operations = append(b.operations, bulkOperation{
		action: "create",
		meta:   meta,
		doc:    doc,
	})
	return b
}

// Update 添加更新操作
func (b *BulkBuilder) Update(index, id string, doc map[string]interface{}) *BulkBuilder {
	if index == "" {
		index = b.index
	}

	meta := map[string]interface{}{
		"_index": index,
		"_id":    id,
	}

	b.operations = append(b.operations, bulkOperation{
		action: "update",
		meta:   meta,
		doc:    map[string]interface{}{"doc": doc},
	})
	return b
}

// Delete 添加删除操作
func (b *BulkBuilder) Delete(index, id string) *BulkBuilder {
	if index == "" {
		index = b.index
	}

	meta := map[string]interface{}{
		"_index": index,
		"_id":    id,
	}

	b.operations = append(b.operations, bulkOperation{
		action: "delete",
		meta:   meta,
	})
	return b
}

// AddFromStruct 从结构体添加索引操作
func (b *BulkBuilder) AddFromStruct(index, id string, data interface{}) *BulkBuilder {
	doc := make(map[string]interface{})
	jsonData, _ := json.Marshal(data)
	json.Unmarshal(jsonData, &doc)
	return b.Add(index, id, doc)
}

// UpdateFromStruct 从结构体添加更新操作
func (b *BulkBuilder) UpdateFromStruct(index, id string, data interface{}) *BulkBuilder {
	doc := make(map[string]interface{})
	jsonData, _ := json.Marshal(data)
	json.Unmarshal(jsonData, &doc)
	return b.Update(index, id, doc)
}

// BulkResponse 批量操作响应
type BulkResponse struct {
	Took   int                           `json:"took"`
	Errors bool                          `json:"errors"`
	Items  []map[string]BulkItemResponse `json:"items"`
}

// BulkItemResponse 批量操作单项响应
type BulkItemResponse struct {
	Index   string `json:"_index"`
	ID      string `json:"_id"`
	Version int    `json:"_version"`
	Result  string `json:"result"`
	Status  int    `json:"status"`
	Error   *struct {
		Type   string `json:"type"`
		Reason string `json:"reason"`
	} `json:"error,omitempty"`
}

// HasErrors 是否有错误
func (r *BulkResponse) HasErrors() bool {
	return r.Errors
}

// FailedItems 返回失败的操作
func (r *BulkResponse) FailedItems() []BulkItemResponse {
	failed := make([]BulkItemResponse, 0)
	for _, item := range r.Items {
		for _, resp := range item {
			if resp.Error != nil {
				failed = append(failed, resp)
			}
		}
	}
	return failed
}

// SuccessCount 成功的操作数量
func (r *BulkResponse) SuccessCount() int {
	count := 0
	for _, item := range r.Items {
		for _, resp := range item {
			if resp.Error == nil {
				count++
			}
		}
	}
	return count
}

// Build 构建批量操作请求体
func (b *BulkBuilder) Build() []byte {
	var buf bytes.Buffer

	for _, op := range b.operations {
		// 写入操作行
		action := map[string]interface{}{
			op.action: op.meta,
		}
		actionLine, _ := json.Marshal(action)
		buf.Write(actionLine)
		buf.WriteByte('\n')

		// 写入文档行（delete 操作不需要）
		if op.action != "delete" && op.doc != nil {
			docLine, _ := json.Marshal(op.doc)
			buf.Write(docLine)
			buf.WriteByte('\n')
		}
	}

	return buf.Bytes()
}

// Debug 启用调试模式（链式调用）
func (b *BulkBuilder) Debug() *BulkBuilder {
	b.debug = true
	return b
}

// printDebug 打印请求调试信息
func (b *BulkBuilder) printDebug(method, path string, body []byte) {
	fmt.Printf("\n[ES Debug] %s %s\n", method, path)
	if body != nil {
		// Bulk API 使用 NDJSON 格式，直接打印
		fmt.Printf("Request Body (NDJSON):\n%s\n", string(body))
	}
}

// printResponse 打印响应调试信息
func (b *BulkBuilder) printResponse(respBody []byte) {
	var pretty interface{}
	json.Unmarshal(respBody, &pretty)
	data, _ := json.MarshalIndent(pretty, "", "  ")
	fmt.Printf("Response:\n%s\n\n", string(data))
}

// Do 执行批量操作
func (b *BulkBuilder) Do(ctx context.Context) (*BulkResponse, error) {
	if len(b.operations) == 0 {
		return nil, fmt.Errorf("没有待执行的批量操作")
	}

	path := "/_bulk"
	body := b.Build()

	// 如果启用调试模式，打印请求信息
	if b.debug {
		b.printDebug("POST", path, body)
	}

	// 创建请求
	url := b.client.GetAddress() + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置正确的 Content-Type
	req.Header.Set("Content-Type", "application/x-ndjson")

	// 执行请求
	respBody, err := b.client.DoRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	// 如果启用调试模式，打印响应信息
	if b.debug {
		b.printResponse(respBody)
	}

	var resp BulkResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &resp, nil
}

// Clear 清空操作列表
func (b *BulkBuilder) Clear() *BulkBuilder {
	b.operations = make([]bulkOperation, 0)
	return b
}

// Count 返回操作数量
func (b *BulkBuilder) Count() int {
	return len(b.operations)
}
