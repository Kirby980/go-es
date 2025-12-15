package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Kirby980/go-es/client"
)

// IndexBuilder 索引构建器
type IndexBuilder struct {
	client   *client.Client
	index    string
	settings map[string]interface{}
	mappings map[string]interface{}
	aliases  map[string]interface{}
	debug    bool // 调试模式标志
}

// NewIndexBuilder 创建索引构建器
func NewIndexBuilder(c *client.Client, index string) *IndexBuilder {
	return &IndexBuilder{
		client:   c,
		index:    index,
		settings: make(map[string]interface{}),
		mappings: make(map[string]interface{}),
		aliases:  make(map[string]interface{}),
	}
}

// Shards 设置分片数
func (b *IndexBuilder) Shards(shards int) *IndexBuilder {
	b.settings["number_of_shards"] = shards
	return b
}

// Replicas 设置副本数
func (b *IndexBuilder) Replicas(replicas int) *IndexBuilder {
	b.settings["number_of_replicas"] = replicas
	return b
}

// RefreshInterval 设置刷新间隔
func (b *IndexBuilder) RefreshInterval(interval string) *IndexBuilder {
	b.settings["refresh_interval"] = interval
	return b
}

// AddProperty 添加字段映射
func (b *IndexBuilder) AddProperty(name string, fieldType string, options ...PropertyOption) *IndexBuilder {
	if b.mappings["properties"] == nil {
		b.mappings["properties"] = make(map[string]interface{})
	}

	properties := b.mappings["properties"].(map[string]interface{})
	field := map[string]interface{}{
		"type": fieldType,
	}

	// 应用选项
	for _, opt := range options {
		opt(field)
	}

	properties[name] = field
	return b
}

// PropertyOption 字段选项
type PropertyOption func(map[string]interface{})

// WithAnalyzer 设置分词器
func WithAnalyzer(analyzer string) PropertyOption {
	return func(field map[string]interface{}) {
		field["analyzer"] = analyzer
	}
}

// WithIndex 设置是否索引
func WithIndex(index bool) PropertyOption {
	return func(field map[string]interface{}) {
		field["index"] = index
	}
}

// WithStore 设置是否存储
func WithStore(store bool) PropertyOption {
	return func(field map[string]interface{}) {
		field["store"] = store
	}
}

// WithFormat 设置日期格式
func WithFormat(format string) PropertyOption {
	return func(field map[string]interface{}) {
		field["format"] = format
	}
}

// WithFields 添加多字段
func WithFields(fields map[string]interface{}) PropertyOption {
	return func(field map[string]interface{}) {
		field["fields"] = fields
	}
}

// WithSubField 添加子字段（多字段映射）
func WithSubField(name string, fieldType string, options ...PropertyOption) PropertyOption {
	return func(field map[string]interface{}) {
		if field["fields"] == nil {
			field["fields"] = make(map[string]interface{})
		}

		subField := map[string]interface{}{
			"type": fieldType,
		}

		// 应用子字段选项
		for _, opt := range options {
			opt(subField)
		}

		field["fields"].(map[string]interface{})[name] = subField
	}
}

// WithSubProperties 添加子属性（嵌套属性）
func WithSubProperties(name string, fileType string, options ...PropertyOption) PropertyOption {
	return func(m map[string]interface{}) {
		if m["properties"] == nil {
			m["properties"] = make(map[string]interface{})
		}

		subField := map[string]interface{}{
			"type": fileType,
		}

		// 应用子字段选项
		for _, opt := range options {
			opt(subField)
		}

		m["properties"].(map[string]interface{})[name] = subField
	}
}

// WithIgnoreAbove 设置 keyword 类型的 ignore_above 参数
func WithIgnoreAbove(limit int) PropertyOption {
	return func(field map[string]interface{}) {
		field["ignore_above"] = limit
	}
}

// AddAlias 添加别名
func (b *IndexBuilder) AddAlias(alias string, filter map[string]interface{}) *IndexBuilder {
	aliasConfig := make(map[string]interface{})
	if filter != nil {
		aliasConfig["filter"] = filter
	}
	b.aliases[alias] = aliasConfig
	return b
}

// Build 构建索引定义
func (b *IndexBuilder) Build() map[string]interface{} {
	body := make(map[string]interface{})

	if len(b.settings) > 0 {
		body["settings"] = b.settings
	}

	if len(b.mappings) > 0 {
		body["mappings"] = b.mappings
	}

	if len(b.aliases) > 0 {
		body["aliases"] = b.aliases
	}

	return body
}

// Debug 启用调试模式（链式调用）
func (b *IndexBuilder) Debug() *IndexBuilder {
	b.debug = true
	return b
}

// printDebug 打印请求调试信息
func (b *IndexBuilder) printDebug(method, path string, body interface{}) {
	fmt.Printf("\n[ES Debug] %s %s\n", method, path)
	if body != nil {
		data, _ := json.MarshalIndent(body, "", "  ")
		fmt.Printf("Request Body:\n%s\n", string(data))
	}
}

// printResponse 打印响应调试信息
func (b *IndexBuilder) printResponse(respBody []byte) {
	if len(respBody) == 0 {
		fmt.Printf("Response: (empty)\n\n")
		return
	}
	var pretty interface{}
	json.Unmarshal(respBody, &pretty)
	data, _ := json.MarshalIndent(pretty, "", "  ")
	fmt.Printf("Response:\n%s\n\n", string(data))
}

// resetDebug 执行后重置debug标志（让每次调用可以独立控制）
func (b *IndexBuilder) resetDebug() {
	b.debug = false
}

// Do 执行创建索引
func (b *IndexBuilder) Do(ctx context.Context) error {
	path := fmt.Sprintf("/%s", b.index)
	body := b.Build()

	// 如果启用调试模式，打印请求信息
	if b.debug {
		b.printDebug("PUT", path, body)
		defer b.resetDebug()
	}

	respBody, err := b.client.Do(ctx, http.MethodPut, path, body)
	if err != nil {
		return err
	}

	// 如果启用调试模式，打印响应信息
	if b.debug {
		b.printResponse(respBody)
	}

	return nil
}

// Delete 删除索引
func (b *IndexBuilder) Delete(ctx context.Context) error {
	path := fmt.Sprintf("/%s", b.index)

	// 如果启用调试模式，打印请求信息
	if b.debug {
		b.printDebug("DELETE", path, nil)
		defer b.resetDebug()
	}

	respBody, err := b.client.Do(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	// 如果启用调试模式，打印响应信息
	if b.debug {
		b.printResponse(respBody)
	}

	return nil
}

// Exists 检查索引是否存在
func (b *IndexBuilder) Exists(ctx context.Context) (bool, error) {
	path := fmt.Sprintf("/%s", b.index)
	_, err := b.client.Do(ctx, http.MethodHead, path, nil)
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
func (b *IndexBuilder) Get(ctx context.Context) (*IndexInfo, error) {
	path := fmt.Sprintf("/%s", b.index)

	// 如果启用调试模式，打印请求信息
	if b.debug {
		b.printDebug("GET", path, nil)
		defer b.resetDebug()
	}

	respBody, err := b.client.Do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	// 如果启用调试模式，打印响应信息
	if b.debug {
		b.printResponse(respBody)
	}

	var result map[string]*IndexInfo
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	if info, ok := result[b.index]; ok {
		return info, nil
	}

	return nil, fmt.Errorf("索引 %s 不存在", b.index)
}
