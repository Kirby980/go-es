package builder

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"unicode"

	"github.com/Kirby980/go-es/errors"
)

// IndexBuilder 索引构建器
type IndexBuilder struct {
	client      ESClient
	index       string
	settings    map[string]any
	mappings    map[string]any
	aliases     map[string]any
	DebugHelper // debug模式
}

// NewIndexBuilder 创建索引构建器
func NewIndexBuilder(c ESClient, index string) *IndexBuilder {
	return &IndexBuilder{
		client:      c,
		index:       index,
		settings:    make(map[string]any),
		mappings:    make(map[string]any),
		aliases:     make(map[string]any),
		DebugHelper: DebugHelper{logger: c.GetLogger()},
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
		b.settings["analysis"] = make(map[string]any)
	}
	analysis := b.settings["analysis"].(map[string]any)

	if analysis["analyzer"] == nil {
		analysis["analyzer"] = make(map[string]any)
	}
	analyzers := analysis["analyzer"].(map[string]any)

	analyzer := map[string]any{
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
type AnalyzerOption func(map[string]any)

// AddAnalyzer 添加分析器（完整版，支持所有选项）
func (b *IndexBuilder) AddAnalyzer(name string, options ...AnalyzerOption) *IndexBuilder {
	if b.settings["analysis"] == nil {
		b.settings["analysis"] = make(map[string]any)
	}
	analysis := b.settings["analysis"].(map[string]any)

	if analysis["analyzer"] == nil {
		analysis["analyzer"] = make(map[string]any)
	}
	analyzers := analysis["analyzer"].(map[string]any)

	analyzer := make(map[string]any)

	// 应用选项
	for _, opt := range options {
		opt(analyzer)
	}

	analyzers[name] = analyzer
	return b
}

// WithAnalyzerType 设置分析器类型
func WithAnalyzerType(analyzerType string) AnalyzerOption {
	return func(analyzer map[string]any) {
		analyzer["type"] = analyzerType
	}
}

// WithTokenizer 设置分词器
func WithTokenizer(tokenizer string) AnalyzerOption {
	return func(analyzer map[string]any) {
		analyzer["tokenizer"] = tokenizer
	}
}

// WithTokenFilters 设置 token filters
func WithTokenFilters(filters ...string) AnalyzerOption {
	return func(analyzer map[string]any) {
		analyzer["filter"] = filters
	}
}

// WithCharFilters 设置 char filters
func WithCharFilters(filters ...string) AnalyzerOption {
	return func(analyzer map[string]any) {
		analyzer["char_filter"] = filters
	}
}

// TokenizerOption 分词器选项
type TokenizerOption func(map[string]any)

// AddTokenizer 添加自定义分词器（用于需要配置 tokenizer 本身的场景）
// 例如：AddTokenizer("ik_smart_case_sensitive", WithTokenizerType("ik_smart"), WithEnableLowercase(false))
func (b *IndexBuilder) AddTokenizer(name string, options ...TokenizerOption) *IndexBuilder {
	if b.settings["analysis"] == nil {
		b.settings["analysis"] = make(map[string]any)
	}
	analysis := b.settings["analysis"].(map[string]any)

	if analysis["tokenizer"] == nil {
		analysis["tokenizer"] = make(map[string]any)
	}
	tokenizers := analysis["tokenizer"].(map[string]any)

	tokenizer := make(map[string]any)

	// 应用选项
	for _, opt := range options {
		opt(tokenizer)
	}

	tokenizers[name] = tokenizer
	return b
}

// WithTokenizerType 设置分词器类型
func WithTokenizerType(tokenizerType string) TokenizerOption {
	return func(tokenizer map[string]any) {
		tokenizer["type"] = tokenizerType
	}
}

// WithEnableLowercase 设置是否启用小写转换（IK 分词器专用）
func WithEnableLowercase(enable bool) TokenizerOption {
	return func(tokenizer map[string]any) {
		tokenizer["enable_lowercase"] = enable
	}
}

// WithMaxTokenLength 设置最大 token 长度
func WithMaxTokenLength(length int) TokenizerOption {
	return func(tokenizer map[string]any) {
		tokenizer["max_token_length"] = length
	}
}

// WithTokenizerMinGram 设置 n-gram 最小长度
func WithTokenizerMinGram(minGram int) TokenizerOption {
	return func(tokenizer map[string]any) {
		tokenizer["min_gram"] = minGram
	}
}

// WithTokenizerMaxGram 设置 n-gram 最大长度
func WithTokenizerMaxGram(maxGram int) TokenizerOption {
	return func(tokenizer map[string]any) {
		tokenizer["max_gram"] = maxGram
	}
}

// WithTokenChars 设置 token 字符类型
func WithTokenChars(chars ...string) TokenizerOption {
	return func(tokenizer map[string]any) {
		tokenizer["token_chars"] = chars
	}
}

// AddProperty 添加字段映射
// 通用模板
func (b *IndexBuilder) AddProperty(name string, fieldType string, options ...PropertyOption) *IndexBuilder {
	if b.mappings["properties"] == nil {
		b.mappings["properties"] = make(map[string]any)
	}

	properties := b.mappings["properties"].(map[string]any)
	field := map[string]any{
		"type": fieldType,
	}

	// 应用选项
	for _, opt := range options {
		opt(field)
	}

	properties[name] = field
	return b
}

// AutoMigrate 自动迁移
func (b *IndexBuilder) AutoMigrate(model any) *IndexBuilder {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if b.index == "" {
		if idxName, ok := model.(IndexName); ok {
			b.index = idxName.IndexName()
		} else {
			// 默认使用复数索引名
			b.index = toSnakeCase(t.Name() + "s")
		}
	}

	// 解析字段映射并设置到 mappings
	properties := parseStructFields(t)
	if len(properties) > 0 {
		b.mappings["properties"] = properties
	}
	return b
}

// parseStructFields 解析结构体字段（支持递归）
func parseStructFields(t reflect.Type) map[string]any {
	properties := make(map[string]any)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("es")

		if tag == "" || tag == "-" {
			continue
		}
		props := parseTag(tag)
		fieldType, ok := props["type"]
		if !ok {
			continue
		}

		fieldName := toSnakeCase(field.Name)
		fieldMapping := map[string]any{
			"type": fieldType,
		}

		// 添加其他属性
		for k, v := range props {
			if k == "type" {
				continue
			}
			fieldMapping[k] = v
		}

		// 处理 object 或 nested 类型，递归解析嵌套结构体
		if fieldType == "object" || fieldType == "nested" {
			nestedType := field.Type
			if nestedType.Kind() == reflect.Ptr {
				nestedType = nestedType.Elem()
			}
			if nestedType.Kind() == reflect.Slice {
				nestedType = nestedType.Elem()
			}
			if nestedType.Kind() == reflect.Ptr {
				nestedType = nestedType.Elem()
			}
			if nestedType.Kind() == reflect.Struct {
				nestedProps := parseStructFields(nestedType)
				if len(nestedProps) > 0 {
					fieldMapping["properties"] = nestedProps
				}
			}
		}

		properties[fieldName] = fieldMapping
	}

	return properties
}

// parseTag 解析 tag
func parseTag(tag string) map[string]string {
	result := make(map[string]string)
	pairs := strings.Split(tag, ";")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, ":", 2)
		if len(kv) == 2 {
			result[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return result
}

// toSnakeCase 转换为 snake_case
func toSnakeCase(s string) string {
	var buf strings.Builder
	runes := []rune(s)
	for i, r := range runes {
		if i > 0 && unicode.IsUpper(r) {
			// 前一个字符是小写，或者下一个字符是小写（处理缩写词尾部）
			if unicode.IsLower(runes[i-1]) || (i+1 < len(runes) && unicode.IsLower(runes[i+1])) {
				buf.WriteByte('_')
			}
		}
		buf.WriteRune(unicode.ToLower(r))
	}
	return buf.String()
}

// PropertyOption 字段选项
type PropertyOption func(map[string]any)

// WithAnalyzer 设置分词器
func WithAnalyzer(analyzer string) PropertyOption {
	return func(field map[string]any) {
		field["analyzer"] = analyzer
	}
}

// WithIndex 设置是否索引
func WithIndex(index bool) PropertyOption {
	return func(field map[string]any) {
		field["index"] = index
	}
}

// WithStore 设置是否存储
func WithStore(store bool) PropertyOption {
	return func(field map[string]any) {
		field["store"] = store
	}
}

// WithFormat 设置日期格式
func WithFormat(format string) PropertyOption {
	return func(field map[string]any) {
		field["format"] = format
	}
}

// WithFields 添加多字段
func WithFields(fields map[string]any) PropertyOption {
	return func(field map[string]any) {
		field["fields"] = fields
	}
}

// WithSubField 添加子字段（多字段映射）
func WithSubField(name string, fieldType string, options ...PropertyOption) PropertyOption {
	return func(field map[string]any) {
		if field["fields"] == nil {
			field["fields"] = make(map[string]any)
		}

		subField := map[string]any{
			"type": fieldType,
		}

		// 应用子字段选项
		for _, opt := range options {
			opt(subField)
		}

		field["fields"].(map[string]any)[name] = subField
	}
}

// WithSubProperties 添加子属性（嵌套属性）
func WithSubProperties(name string, fileType string, options ...PropertyOption) PropertyOption {
	return func(m map[string]any) {
		if m["properties"] == nil {
			m["properties"] = make(map[string]any)
		}

		subField := map[string]any{
			"type": fileType,
		}

		// 应用子字段选项
		for _, opt := range options {
			opt(subField)
		}

		m["properties"].(map[string]any)[name] = subField
	}
}

// WithIgnoreAbove 设置 keyword 类型的 ignore_above 参数
func WithIgnoreAbove(limit int) PropertyOption {
	return func(field map[string]any) {
		field["ignore_above"] = limit
	}
}

// AddAlias 添加别名
func (b *IndexBuilder) AddAlias(alias string, filter map[string]any) *IndexBuilder {
	aliasConfig := make(map[string]any)
	if filter != nil {
		aliasConfig["filter"] = filter
	}
	b.aliases[alias] = aliasConfig
	return b
}

// Build 构建索引定义
func (b *IndexBuilder) Build() map[string]any {
	body := make(map[string]any)

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
	b.SetDebug(true)
	return b
}

// Create 创建索引
func (b *IndexBuilder) Create(ctx context.Context) error {
	path := fmt.Sprintf("/%s", b.index)
	body := b.Build()

	// 如果启用调试模式，打印请求信息
	if b.IsDebug() {
		b.PrintDebug("PUT", path, body)
		defer b.SetDebug(false)
	}

	respBody, err := b.client.Do(ctx, http.MethodPut, path, body)
	if err != nil {
		return err
	}

	// 如果启用调试模式，打印响应信息
	if b.IsDebug() {
		b.PrintResponse(respBody)
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
	body := map[string]any{
		"settings": b.settings,
	}

	// 如果启用调试模式，打印请求信息
	if b.IsDebug() {
		b.PrintDebug("PUT", path, body)
		defer b.SetDebug(false)
	}

	respBody, err := b.client.Do(ctx, http.MethodPut, path, body)
	if err != nil {
		return err
	}

	// 如果启用调试模式，打印响应信息
	if b.IsDebug() {
		b.PrintResponse(respBody)
	}

	return nil
}

// PutMapping 更新索引映射（添加新字段或更新已有字段映射）
func (b *IndexBuilder) PutMapping(ctx context.Context) error {
	path := fmt.Sprintf("/%s/_mapping", b.index)

	// 如果启用调试模式，打印请求信息
	if b.IsDebug() {
		b.PrintDebug("PUT", path, b.mappings)
		defer b.SetDebug(false)
	}

	respBody, err := b.client.Do(ctx, http.MethodPut, path, b.mappings)
	if err != nil {
		return err
	}

	// 如果启用调试模式，打印响应信息
	if b.IsDebug() {
		b.PrintResponse(respBody)
	}

	return nil
}

// Delete 删除索引
func (b *IndexBuilder) Delete(ctx context.Context) error {
	path := fmt.Sprintf("/%s", b.index)

	// 如果启用调试模式，打印请求信息
	if b.IsDebug() {
		b.PrintDebug("DELETE", path, nil)
		defer b.SetDebug(false)
	}

	respBody, err := b.client.Do(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	// 如果启用调试模式，打印响应信息
	if b.IsDebug() {
		b.PrintResponse(respBody)
	}

	return nil
}

// Exists 检查索引是否存在
func (b *IndexBuilder) Exists(ctx context.Context) (bool, error) {
	path := fmt.Sprintf("/%s", b.index)
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

// IndexInfo 索引信息
type IndexInfo struct {
	Aliases  map[string]any `json:"aliases"`
	Mappings map[string]any `json:"mappings"`
	Settings map[string]any `json:"settings"`
}

// Get 获取索引信息
func (b *IndexBuilder) Get(ctx context.Context) (*IndexInfo, error) {
	path := fmt.Sprintf("/%s", b.index)

	// 如果启用调试模式，打印请求信息
	if b.IsDebug() {
		b.PrintDebug("GET", path, nil)
		defer b.SetDebug(false)
	}

	respBody, err := b.client.Do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	// 如果启用调试模式，打印响应信息
	if b.IsDebug() {
		b.PrintResponse(respBody)
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
