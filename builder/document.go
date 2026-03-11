package builder

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/Kirby980/go-es/errors"
)

// DocumentBuilder 文档构建器
type DocumentBuilder struct {
	client        ESClient
	index         string
	id            string
	doc           map[string]any
	script        map[string]any
	refresh       string // refresh 参数: true, false, wait_for
	sourceInclude []string
	sourceExclude []string
	err           error
	debugHelper
}

// NewDocumentBuilder 创建文档构建器
func NewDocumentBuilder(c ESClient, index string) *DocumentBuilder {
	return &DocumentBuilder{
		client:      c,
		index:       index,
		doc:         make(map[string]any),
		debugHelper: debugHelper{logger: c.GetLogger()},
	}
}

// Model 从结构体设置文档数据，并自动推断索引名
// 如果结构体实现了 IndexName 接口，使用该方法返回的索引名
// 否则使用结构体名称的 snake_case + s 作为索引名
func (b *DocumentBuilder) Model(model any) *DocumentBuilder {
	// 自动推断索引名
	if b.index == "" {
		if namer, ok := model.(IndexName); ok {
			b.index = namer.IndexName()
		} else {
			t := reflect.TypeOf(model)
			if t.Kind() == reflect.Ptr {
				t = t.Elem()
			}
			b.index = toSnakeCase(t.Name() + "s")
		}
	}
	// 设置文档数据
	jsonData, err := json.Marshal(model)
	if err != nil {
		b.err = fmt.Errorf("Error marshalling model: %w", err)
		return b
	}
	err = json.Unmarshal(jsonData, &b.doc)
	if err != nil {
		b.err = fmt.Errorf("Error unmarshalling model: %w", err)
		return b
	}
	return b
}

// ID 设置文档 ID
func (b *DocumentBuilder) ID(id string) *DocumentBuilder {
	b.id = id
	return b
}

// Set 设置字段值
func (b *DocumentBuilder) Set(field string, value any) *DocumentBuilder {
	b.doc[field] = value
	return b
}

// SetMap 批量设置字段
func (b *DocumentBuilder) SetMap(data map[string]any) *DocumentBuilder {
	for k, v := range data {
		b.doc[k] = v
	}
	return b
}

// SetStruct 从结构体设置
func (b *DocumentBuilder) SetStruct(data any) *DocumentBuilder {
	jsonData, err := json.Marshal(data)
	if err != nil {
		b.err = fmt.Errorf("Error marshalling model: %w", err)
		return b
	}
	err = json.Unmarshal(jsonData, &b.doc)
	if err != nil {
		b.err = fmt.Errorf("Error unmarshalling model: %w", err)
		return b
	}
	return b
}

// SetObject 设置嵌套对象
func (b *DocumentBuilder) SetObject(key string, builder func(*NestedObject)) *DocumentBuilder {
	nested := &NestedObject{data: make(map[string]any)}
	builder(nested)
	if nested.err != nil {
		b.err = fmt.Errorf("Error building nested object: %w", nested.err)
		return b
	}
	b.doc[key] = nested.data
	return b
}

// SetArray 设置数组字段
func (b *DocumentBuilder) SetArray(key string, values ...any) *DocumentBuilder {
	b.doc[key] = values
	return b
}

// SetObjectArray 设置对象数组
func (b *DocumentBuilder) SetObjectArray(key string, builders ...func(*NestedObject)) *DocumentBuilder {
	arr := make([]map[string]any, len(builders))
	for i, builder := range builders {
		nested := &NestedObject{data: make(map[string]any)}
		builder(nested)
		if nested.err != nil {
			b.err = fmt.Errorf("Error building nested object: %w", nested.err)
			return b
		}
		arr[i] = nested.data
	}
	b.doc[key] = arr
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
	u, _ := url.Parse(basePath)
	q := u.Query()

	if b.refresh != "" {
		q.Set("refresh", b.refresh)
	}

	if len(b.sourceInclude) > 0 {
		q.Set("_source_includes", strings.Join(b.sourceInclude, ","))
	}

	if len(b.sourceExclude) > 0 {
		q.Set("_source_excludes", strings.Join(b.sourceExclude, ","))
	}

	u.RawQuery = q.Encode()
	return u.String()
}

// Debug 启用调试模式（链式调用）
func (b *DocumentBuilder) Debug() *DocumentBuilder {
	b.setDebug(true)
	return b
}

// Script 设置脚本更新
func (b *DocumentBuilder) Script(source string, params map[string]any) *DocumentBuilder {
	b.script = map[string]any{
		"source": source,
		"lang":   "painless",
	}
	if params != nil {
		b.script["params"] = params
	}
	return b
}

// ScriptUpsert 脚本更新时如果不存在则插入
func (b *DocumentBuilder) ScriptUpsert(upsertDoc any) *DocumentBuilder {
	if b.script == nil {
		b.err = fmt.Errorf("ScriptUpsert requires a script to be set first")
		return b
	}
	b.script["upsert"] = upsertDoc
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

// Do 索引文档（创建或更新）
func (b *DocumentBuilder) Do(ctx context.Context) (*DocumentResponse, error) {
	if b.err != nil {
		return nil, b.err
	}
	var path string
	var method string

	if b.id != "" {
		path = fmt.Sprintf("/%s/_doc/%s", url.PathEscape(b.index), url.PathEscape(b.id))
		method = http.MethodPut
	} else {
		path = fmt.Sprintf("/%s/_doc", url.PathEscape(b.index))
		method = http.MethodPost
	}

	// 添加查询参数
	path = b.buildPath(path)

	// 如果启用调试模式，打印请求信息
	if b.isDebug() {
		b.printDebug(method, path, b.doc)
		defer b.setDebug(false)
	}

	respBody, err := b.client.Do(ctx, method, path, b.doc)
	if err != nil {
		return nil, err
	}

	// 如果启用调试模式，打印响应信息
	if b.isDebug() {
		b.printResponse(respBody)
	}

	var resp DocumentResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &resp, nil
}

// Create 创建文档（如果已存在则失败）
func (b *DocumentBuilder) Create(ctx context.Context) (*DocumentResponse, error) {
	if b.err != nil {
		return nil, b.err
	}
	if b.id == "" {
		return nil, fmt.Errorf("创建文档需要指定 ID")
	}
// ... (rest remains same but effectively prepended b.err check)

	path := fmt.Sprintf("/%s/_create/%s", url.PathEscape(b.index), url.PathEscape(b.id))
	path = b.buildPath(path)

	// 如果启用调试模式，打印请求信息
	if b.isDebug() {
		b.printDebug("PUT", path, b.doc)
		defer b.setDebug(false)
	}

	respBody, err := b.client.Do(ctx, http.MethodPut, path, b.doc)
	if err != nil {
		return nil, err
	}

	// 如果启用调试模式，打印响应信息
	if b.isDebug() {
		b.printResponse(respBody)
	}

	var resp DocumentResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &resp, nil
}

// Update 更新文档（部分更新）
func (b *DocumentBuilder) Update(ctx context.Context) (*DocumentResponse, error) {
	if b.err != nil {
		return nil, b.err
	}
	if b.id == "" {
		return nil, fmt.Errorf("更新文档需要指定 ID")
	}

	path := fmt.Sprintf("/%s/_update/%s", url.PathEscape(b.index), url.PathEscape(b.id))
	path = b.buildPath(path)

	updateBody := make(map[string]any)
	if b.script != nil {
		updateBody["script"] = b.script
	} else {
		updateBody["doc"] = b.doc
	}

	// 如果启用调试模式，打印请求信息
	if b.isDebug() {
		b.printDebug("POST", path, updateBody)
		defer b.setDebug(false)
	}

	respBody, err := b.client.Do(ctx, http.MethodPost, path, updateBody)
	if err != nil {
		return nil, err
	}

	// 如果启用调试模式，打印响应信息
	if b.isDebug() {
		b.printResponse(respBody)
	}

	var resp DocumentResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &resp, nil
}

// Upsert 更新或插入
func (b *DocumentBuilder) Upsert(ctx context.Context) (*DocumentResponse, error) {
	if b.err != nil {
		return nil, b.err
	}
	if b.id == "" {
		return nil, fmt.Errorf("upsert 需要指定 ID")
	}

	path := fmt.Sprintf("/%s/_update/%s", url.PathEscape(b.index), url.PathEscape(b.id))
	path = b.buildPath(path)

	updateBody := map[string]any{
		"doc":           b.doc,
		"doc_as_upsert": true,
	}

	// 如果启用调试模式，打印请求信息
	if b.isDebug() {
		b.printDebug("POST", path, updateBody)
		defer b.setDebug(false)
	}

	respBody, err := b.client.Do(ctx, http.MethodPost, path, updateBody)
	if err != nil {
		return nil, err
	}

	// 如果启用调试模式，打印响应信息
	if b.isDebug() {
		b.printResponse(respBody)
	}

	var resp DocumentResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &resp, nil
}

// Get 获取文档
func (b *DocumentBuilder) Get(ctx context.Context) (*GetResponse, error) {
	if b.err != nil {
		return nil, b.err
	}
	if b.id == "" {
		return nil, fmt.Errorf("获取文档需要指定 ID")
	}

	path := fmt.Sprintf("/%s/_doc/%s", url.PathEscape(b.index), url.PathEscape(b.id))

	// 如果启用调试模式，打印请求信息
	if b.isDebug() {
		b.printDebug("GET", path, nil)
		defer b.setDebug(false)
	}

	respBody, err := b.client.Do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	// 如果启用调试模式，打印响应信息
	if b.isDebug() {
		b.printResponse(respBody)
	}

	var resp GetResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &resp, nil
}

// Delete 删除文档
func (b *DocumentBuilder) Delete(ctx context.Context) (*DocumentResponse, error) {
	if b.err != nil {
		return nil, b.err
	}
	if b.id == "" {
		return nil, fmt.Errorf("删除文档需要指定 ID")
	}

	path := fmt.Sprintf("/%s/_doc/%s", url.PathEscape(b.index), url.PathEscape(b.id))
	path = b.buildPath(path)

	// 如果启用调试模式，打印请求信息
	if b.isDebug() {
		b.printDebug("DELETE", path, nil)
		defer b.setDebug(false)
	}

	respBody, err := b.client.Do(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return nil, err
	}

	// 如果启用调试模式，打印响应信息
	if b.isDebug() {
		b.printResponse(respBody)
	}

	var resp DocumentResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &resp, nil
}

// Exists 检查文档是否存在
func (b *DocumentBuilder) Exists(ctx context.Context) (bool, error) {
	if b.err != nil {
		return false, b.err
	}
	if b.id == "" {
		return false, fmt.Errorf("检查文档需要指定 ID")
	}

	path := fmt.Sprintf("/%s/_doc/%s", url.PathEscape(b.index), url.PathEscape(b.id))
	_, err := b.client.Do(ctx, http.MethodHead, path, nil)
	if err != nil {
		var esErr *errors.ESError
		if stderrors.As(err, &esErr) && esErr.IsNotFound() {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// SourceInclude 设置返回字段（包含）
func (b *DocumentBuilder) SourceInclude(fields ...string) *DocumentBuilder {
	b.sourceInclude = fields
	return b
}

// SourceExclude 设置返回字段（排除）
func (b *DocumentBuilder) SourceExclude(fields ...string) *DocumentBuilder {
	b.sourceExclude = fields
	return b
}

// GetResponse 获取文档响应
type GetResponse struct {
	Index   string         `json:"_index"`
	ID      string         `json:"_id"`
	Version int            `json:"_version"`
	Found   bool           `json:"found"`
	Source  map[string]any `json:"_source"`
}
