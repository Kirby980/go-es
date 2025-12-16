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

// AddCustomAnalyzer 添加自定义分析器（简化版，用于快速创建基于某个 tokenizer 的分析器）
// 例如：AddCustomAnalyzer("ik_case_sensitive", "ik_smart")
func (b *IndexBuilder) AddCustomAnalyzer(name string, tokenizer string, filters ...string) *IndexBuilder {
	if b.settings["analysis"] == nil {
		b.settings["analysis"] = make(map[string]interface{})
	}
	analysis := b.settings["analysis"].(map[string]interface{})

	if analysis["analyzer"] == nil {
		analysis["analyzer"] = make(map[string]interface{})
	}
	analyzers := analysis["analyzer"].(map[string]interface{})

	analyzer := map[string]interface{}{
		"type":      "custom",
		"tokenizer": tokenizer,
	}

	if len(filters) > 0 {
		analyzer["filter"] = filters
	} else {
		analyzer["filter"] = []string{}
	}

	analyzers[name] = analyzer
	return b
}

// AnalyzerOption 分析器选项
type AnalyzerOption func(map[string]interface{})

// AddAnalyzer 添加分析器（完整版，支持所有选项）
func (b *IndexBuilder) AddAnalyzer(name string, options ...AnalyzerOption) *IndexBuilder {
	if b.settings["analysis"] == nil {
		b.settings["analysis"] = make(map[string]interface{})
	}
	analysis := b.settings["analysis"].(map[string]interface{})

	if analysis["analyzer"] == nil {
		analysis["analyzer"] = make(map[string]interface{})
	}
	analyzers := analysis["analyzer"].(map[string]interface{})

	analyzer := make(map[string]interface{})

	// 应用选项
	for _, opt := range options {
		opt(analyzer)
	}

	analyzers[name] = analyzer
	return b
}

// WithAnalyzerType 设置分析器类型
func WithAnalyzerType(analyzerType string) AnalyzerOption {
	return func(analyzer map[string]interface{}) {
		analyzer["type"] = analyzerType
	}
}

// WithTokenizer 设置分词器
func WithTokenizer(tokenizer string) AnalyzerOption {
	return func(analyzer map[string]interface{}) {
		analyzer["tokenizer"] = tokenizer
	}
}

// WithTokenFilters 设置 token filters
func WithTokenFilters(filters ...string) AnalyzerOption {
	return func(analyzer map[string]interface{}) {
		analyzer["filter"] = filters
	}
}

// WithCharFilters 设置 char filters
func WithCharFilters(filters ...string) AnalyzerOption {
	return func(analyzer map[string]interface{}) {
		analyzer["char_filter"] = filters
	}
}

// TokenizerOption 分词器选项
type TokenizerOption func(map[string]interface{})

// AddTokenizer 添加自定义分词器（用于需要配置 tokenizer 本身的场景）
// 例如：AddTokenizer("ik_smart_case_sensitive", WithTokenizerType("ik_smart"), WithEnableLowercase(false))
func (b *IndexBuilder) AddTokenizer(name string, options ...TokenizerOption) *IndexBuilder {
	if b.settings["analysis"] == nil {
		b.settings["analysis"] = make(map[string]interface{})
	}
	analysis := b.settings["analysis"].(map[string]interface{})

	if analysis["tokenizer"] == nil {
		analysis["tokenizer"] = make(map[string]interface{})
	}
	tokenizers := analysis["tokenizer"].(map[string]interface{})

	tokenizer := make(map[string]interface{})

	// 应用选项
	for _, opt := range options {
		opt(tokenizer)
	}

	tokenizers[name] = tokenizer
	return b
}

// WithTokenizerType 设置分词器类型
func WithTokenizerType(tokenizerType string) TokenizerOption {
	return func(tokenizer map[string]interface{}) {
		tokenizer["type"] = tokenizerType
	}
}

// WithEnableLowercase 设置是否启用小写转换（IK 分词器专用）
func WithEnableLowercase(enable bool) TokenizerOption {
	return func(tokenizer map[string]interface{}) {
		tokenizer["enable_lowercase"] = enable
	}
}

// WithMaxTokenLength 设置最大 token 长度
func WithMaxTokenLength(length int) TokenizerOption {
	return func(tokenizer map[string]interface{}) {
		tokenizer["max_token_length"] = length
	}
}

// WithTokenizerMinGram 设置 n-gram 最小长度
func WithTokenizerMinGram(minGram int) TokenizerOption {
	return func(tokenizer map[string]interface{}) {
		tokenizer["min_gram"] = minGram
	}
}

// WithTokenizerMaxGram 设置 n-gram 最大长度
func WithTokenizerMaxGram(maxGram int) TokenizerOption {
	return func(tokenizer map[string]interface{}) {
		tokenizer["max_gram"] = maxGram
	}
}

// WithTokenChars 设置 token 字符类型
func WithTokenChars(chars ...string) TokenizerOption {
	return func(tokenizer map[string]interface{}) {
		tokenizer["token_chars"] = chars
	}
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

// Create 创建索引
func (b *IndexBuilder) Create(ctx context.Context) error {
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

// Do 执行创建索引（Create 的别名，保持向后兼容）
func (b *IndexBuilder) Do(ctx context.Context) error {
	return b.Create(ctx)
}

// UpdateSettings 更新索引设置
func (b *IndexBuilder) UpdateSettings(ctx context.Context) error {
	path := fmt.Sprintf("/%s/_settings", b.index)
	body := map[string]interface{}{
		"settings": b.settings,
	}

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

// PutMapping 更新索引映射（添加新字段或更新已有字段映射）
func (b *IndexBuilder) PutMapping(ctx context.Context) error {
	path := fmt.Sprintf("/%s/_mapping", b.index)

	// 如果启用调试模式，打印请求信息
	if b.debug {
		b.printDebug("PUT", path, b.mappings)
		defer b.resetDebug()
	}

	respBody, err := b.client.Do(ctx, http.MethodPut, path, b.mappings)
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
