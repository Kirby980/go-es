package builder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// BulkBuilder 批量操作构建器
type BulkBuilder struct {
	client     ESClient
	index      string
	operations []bulkOperation
	currentOp  *bulkOperation // 当前正在构建的操作（用于链式调用）
	debugHelper
	baseBuilder
	autoFlushSize int                 // 自动刷新大小
	onFlush       func(*BulkResponse) // 分批回调
	err           error
	ctx           context.Context // 用户提供的自动刷新上下文
}

// bulkOperation 批量操作项
type bulkOperation struct {
	action string
	meta   map[string]any
	doc    map[string]any
}

// NewBulkBuilder 创建批量操作构建器
func NewBulkBuilder(c ESClient) *BulkBuilder {
	b := &BulkBuilder{
		client:      c,
		operations:  make([]bulkOperation, 0),
		debugHelper: debugHelper{logger: c.GetLogger()},
	}
	b.initBaseBuilder()
	return b
}

// Header 设置自定义 Header (链式调用)
func (b *BulkBuilder) Header(key, value string) *BulkBuilder {
	b.baseBuilder.Header(key, value)
	return b
}

// AutoFlushSize 设置自动刷新大小
func (b *BulkBuilder) AutoFlushSize(size int) *BulkBuilder {
	b.autoFlushSize = size
	return b
}

// OnFlush 设置分批回调(每批完成后回调)
func (b *BulkBuilder) OnFlush(callback func(*BulkResponse)) *BulkBuilder {
	b.onFlush = callback
	return b
}

// AutoFlushContext 设置自动刷新使用的上下文（链式调用）
// 若不调用此方法，自动刷新默认使用 context.Background()
func (b *BulkBuilder) AutoFlushContext(ctx context.Context) *BulkBuilder {
	b.ctx = ctx
	return b
}

// Index 设置默认索引
func (b *BulkBuilder) Index(index string) *BulkBuilder {
	b.index = index
	return b
}

// commitCurrent 提交当前正在构建的操作到操作列表
func (b *BulkBuilder) commitCurrent() {
	if b.currentOp != nil {
		// 如果是 update 操作，需要包装成 {"doc": {...}}
		if b.currentOp.action == "update" && b.currentOp.doc != nil {
			b.currentOp.doc = map[string]any{"doc": b.currentOp.doc}
		}
		b.operations = append(b.operations, *b.currentOp)
		b.currentOp = nil
	}
}

// Add 添加索引操作
func (b *BulkBuilder) Add(index, id string, doc map[string]any) *BulkBuilder {
	b.commitCurrent() // 先提交链式构建的操作（如果有）
	b.isFlush()

	if index == "" {
		index = b.index
	}

	meta := map[string]any{
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

func (b *BulkBuilder) isFlush() {
	if b.autoFlushSize > 0 && len(b.operations) >= b.autoFlushSize {
		// 使用用户提供的上下文，若未设置则使用 context.Background()
		ctx := b.ctx
		if ctx == nil {
			ctx = context.Background()
		}
		resp, err := b.flush(ctx)
		if err != nil {
			b.err = fmt.Errorf("auto-flush failed: %w", err)
			return
		}
		if b.onFlush != nil {
			b.onFlush(resp)
		}
	}
}

func (b *BulkBuilder) flush(ctx context.Context) (*BulkResponse, error) {
	if len(b.operations) == 0 {
		return &BulkResponse{}, nil
	}
	flushCount := b.autoFlushSize
	if b.autoFlushSize == 0 || len(b.operations) <= b.autoFlushSize {
		flushCount = len(b.operations)
	}

	toFlush := b.operations[:flushCount]
	remaining := b.operations[flushCount:]

	b.operations = toFlush
	resp, err := b.Do(ctx)
	b.operations = remaining

	return resp, err
}

func (b *BulkBuilder) Flush(ctx context.Context) (*BulkResponse, error) {
	return b.flush(ctx)
}

// Create 添加创建操作（文档已存在则失败）
func (b *BulkBuilder) Create(index, id string, doc map[string]any) *BulkBuilder {
	b.commitCurrent() // 先提交链式构建的操作（如果有）

	if index == "" {
		index = b.index
	}

	meta := map[string]any{
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
func (b *BulkBuilder) Update(index, id string, doc map[string]any) *BulkBuilder {
	b.commitCurrent() // 先提交链式构建的操作（如果有）

	if index == "" {
		index = b.index
	}

	meta := map[string]any{
		"_index": index,
		"_id":    id,
	}

	b.operations = append(b.operations, bulkOperation{
		action: "update",
		meta:   meta,
		doc:    map[string]any{"doc": doc},
	})
	return b
}

// Delete 添加删除操作
func (b *BulkBuilder) Delete(index, id string) *BulkBuilder {
	b.commitCurrent() // 先提交链式构建的操作（如果有）

	if index == "" {
		index = b.index
	}

	meta := map[string]any{
		"_index": index,
		"_id":    id,
	}

	b.operations = append(b.operations, bulkOperation{
		action: "delete",
		meta:   meta,
	})
	return b
}

// ========== 链式调用 API ==========

// AddDoc 开始添加索引操作（链式调用）
func (b *BulkBuilder) AddDoc(id string) *BulkBuilder {
	return b.AddDocWithIndex("", id)
}

// AddDocWithIndex 开始添加索引操作并指定索引（链式调用）
func (b *BulkBuilder) AddDocWithIndex(index, id string) *BulkBuilder {
	b.commitCurrent()
	b.isFlush()
	if index == "" {
		index = b.index
	}

	meta := map[string]any{
		"_index": index,
	}
	if id != "" {
		meta["_id"] = id
	}

	b.currentOp = &bulkOperation{
		action: "index",
		meta:   meta,
		doc:    make(map[string]any),
	}

	return b
}

// CreateDoc 开始添加创建操作（链式调用）
func (b *BulkBuilder) CreateDoc(id string) *BulkBuilder {
	return b.CreateDocWithIndex("", id)
}

// CreateDocWithIndex 开始添加创建操作并指定索引（链式调用）
func (b *BulkBuilder) CreateDocWithIndex(index, id string) *BulkBuilder {
	b.commitCurrent()
	b.isFlush()

	if index == "" {
		index = b.index
	}

	meta := map[string]any{
		"_index": index,
		"_id":    id,
	}

	b.currentOp = &bulkOperation{
		action: "create",
		meta:   meta,
		doc:    make(map[string]any),
	}

	return b
}

// UpdateDoc 开始添加更新操作（链式调用）
func (b *BulkBuilder) UpdateDoc(id string) *BulkBuilder {
	return b.UpdateDocWithIndex("", id)
}

// UpdateDocWithIndex 开始添加更新操作并指定索引（链式调用）
func (b *BulkBuilder) UpdateDocWithIndex(index, id string) *BulkBuilder {
	b.commitCurrent()
	b.isFlush()

	if index == "" {
		index = b.index
	}

	meta := map[string]any{
		"_index": index,
		"_id":    id,
	}

	b.currentOp = &bulkOperation{
		action: "update",
		meta:   meta,
		doc:    make(map[string]any),
	}

	return b
}

// DeleteDoc 添加删除操作（链式调用）
func (b *BulkBuilder) DeleteDoc(id string) *BulkBuilder {
	return b.DeleteDocWithIndex("", id)
}

// DeleteDocWithIndex 添加删除操作并指定索引（链式调用）
func (b *BulkBuilder) DeleteDocWithIndex(index, id string) *BulkBuilder {
	b.commitCurrent()
	b.isFlush()

	if index == "" {
		index = b.index
	}

	meta := map[string]any{
		"_index": index,
		"_id":    id,
	}

	// Delete 不需要 doc，直接添加到 operations
	b.operations = append(b.operations, bulkOperation{
		action: "delete",
		meta:   meta,
	})

	return b
}

// Set 设置字段值（链式调用，需要先调用 AddDoc/CreateDoc/UpdateDoc）
func (b *BulkBuilder) Set(key string, value any) *BulkBuilder {
	if b.currentOp == nil {
		// 如果没有当前操作，报错
		b.err = fmt.Errorf("Set() must be called after AddDoc/CreateDoc/UpdateDoc")
		return b
	}
	b.currentOp.doc[key] = value
	return b
}

// SetFromStruct 从结构体设置字段（链式调用）
func (b *BulkBuilder) SetFromStruct(data any) *BulkBuilder {
	if b.currentOp == nil {
		b.err = fmt.Errorf("SetFromStruct() must be called after AddDoc/CreateDoc/UpdateDoc")
		return b
	}
	m, err := structToMap(data)
	if err != nil {
		b.err = fmt.Errorf("SetFromStruct() error: %v", err)
		return b
	}
	for k, v := range m {
		b.currentOp.doc[k] = v
	}
	return b
}

// SetObject 设置嵌套对象（链式调用）
func (b *BulkBuilder) SetObject(key string, builder func(*NestedObject)) *BulkBuilder {
	if b.currentOp == nil {
		b.err = fmt.Errorf("SetObject () must be called after AddDoc/CreateDoc/UpdateDoc")
		return b
	}
	nested := &NestedObject{data: make(map[string]any)}
	builder(nested)
	if nested.err != nil {
		b.err = fmt.Errorf("SetObject() error: %v", nested.err)
		return b
	}
	b.currentOp.doc[key] = nested.data
	return b
}

// SetArray 设置数组字段（链式调用）
func (b *BulkBuilder) SetArray(key string, values ...any) *BulkBuilder {
	if b.currentOp == nil {
		b.err = fmt.Errorf("SetArray() must be called after AddDoc/CreateDoc/UpdateDoc")
		return b
	}
	b.currentOp.doc[key] = values
	return b
}

// SetObjectArray 设置对象数组（链式调用）
func (b *BulkBuilder) SetObjectArray(key string, builders ...func(*NestedObject)) *BulkBuilder {
	if b.currentOp == nil {
		b.err = fmt.Errorf("SetObjectArray() must be called after AddDoc/CreateDoc/UpdateDoc")
		return b
	}
	arr := make([]map[string]any, len(builders))
	for i, builder := range builders {
		nested := &NestedObject{data: make(map[string]any)}
		builder(nested)
		if nested.err != nil {
			b.err = fmt.Errorf("SetObjectArray() error: %v", nested.err)
			return b
		}
		arr[i] = nested.data
	}
	b.currentOp.doc[key] = arr
	return b
}

// AddFromStruct 从结构体添加索引操作
func (b *BulkBuilder) AddFromStruct(index, id string, data any) *BulkBuilder {
	doc, err := structToMap(data)
	if err != nil {
		b.err = fmt.Errorf("AddFromStruct error: %v", err)
		return b
	}
	return b.Add(index, id, doc)
}

// UpdateFromStruct 从结构体添加更新操作
func (b *BulkBuilder) UpdateFromStruct(index, id string, data any) *BulkBuilder {
	doc, err := structToMap(data)
	if err != nil {
		b.err = fmt.Errorf("UpdateFromStruct error: %v", err)
		return b
	}
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
	b.commitCurrent() // 提交链式构建的操作（如果有）

	buf := getBuffer()
	defer putBuffer(buf)

	if buf.Cap() < len(b.operations)*200 {
		buf.Grow(len(b.operations) * 200)
	}

	encoder := json.NewEncoder(buf)

	for _, op := range b.operations {
		// 写入操作行
		action := map[string]any{
			op.action: op.meta,
		}
		if err := encoder.Encode(action); err != nil {
			b.err = fmt.Errorf("序列化操作行失败: %w", err)
			return nil
		}

		// 写入文档行（delete 操作不需要）
		if op.action != "delete" && op.doc != nil {
			if err := encoder.Encode(op.doc); err != nil {
				b.err = fmt.Errorf("序列化文档行失败: %w", err)
				return nil
			}
		}
	}

	// copy bytes out to return since buffer will be recycled
	res := make([]byte, buf.Len())
	copy(res, buf.Bytes())
	return res
}

// Debug 启用调试模式（链式调用）
func (b *BulkBuilder) Debug() *BulkBuilder {
	b.setDebug(true)
	return b
}

// Do 执行批量操作
func (b *BulkBuilder) Do(ctx context.Context) (*BulkResponse, error) {
	b.commitCurrent() // 提交链式构建的操作（如果有）

	if b.err != nil {
		return nil, b.err
	}
	if len(b.operations) == 0 {
		return nil, fmt.Errorf("没有待执行的批量操作")
	}

	path := "/_bulk"
	body := b.Build()

	// 如果启用调试模式，打印请求信息
	if b.isDebug() {
		b.printDebug("POST", path, body)
		defer b.autoResetDebug()
	}

	// 创建请求
	url := b.client.GetAddress() + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置正确的 Content-Type 和自定义 Header
	req.Header.Set("Content-Type", "application/x-ndjson")
	for k, v := range b.getHeaders() {
		for _, vv := range v {
			req.Header.Add(k, vv)
		}
	}

	// 执行请求
	respBody, err := b.client.DoRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	// 如果启用调试模式，打印响应信息
	if b.isDebug() {
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
	count := len(b.operations)
	if b.currentOp != nil {
		count++
	}
	return count
}
