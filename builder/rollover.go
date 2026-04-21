package builder

import (
	"context"
	"net/http"
	"net/url"
)

// RolloverBuilder 是用于构建和执行 Elasticsearch Rollover 操作的链式 API 构建器。
type RolloverBuilder struct {
	client     ESClient
	alias      string
	newIndex   string
	conditions map[string]any
	settings   map[string]any
	mappings   map[string]any
	aliases    map[string]any
	dryRun     *bool
	debugHelper
	baseBuilder
}

// RolloverResponse 定义了 Elasticsearch Rollover 操作返回的结构化响应数据，包含元数据及核心结果。
type RolloverResponse struct {
	Acknowledged       bool           `json:"acknowledged"`
	ShardsAcknowledged bool           `json:"shards_acknowledged"`
	OldIndex           string         `json:"old_index"`
	NewIndex           string         `json:"new_index"`
	RolledOver         bool           `json:"rolled_over"`
	DryRun             bool           `json:"dry_run"`
	Conditions         map[string]any `json:"conditions"`
}

// NewRolloverBuilder 创建并返回一个 Rollover 构建器实例，用于构造和执行 Elasticsearch 的 Rollover 请求。
func NewRolloverBuilder(c ESClient, alias string) *RolloverBuilder {
	b := &RolloverBuilder{
		client: c,
		alias:  alias,
	}
	b.initBaseBuilder()
	return b
}

// Header 设置自定义的 HTTP 请求头 (例如: Header("Content-Type", "application/json"))。此方法支持链式调用。
func (b *RolloverBuilder) Header(key, value string) *RolloverBuilder {
	b.baseBuilder.Header(key, value)
	return b
}

// NewIndex 指定 Rollover 或 Reindex 等操作生成的新索引名称。
func (b *RolloverBuilder) NewIndex(name string) *RolloverBuilder {
	b.newIndex = name
	return b
}

// Condition 设置触发某些操作 (如 Rollover) 的条件 (例如: max_age, max_docs, max_size)。
func (b *RolloverBuilder) Condition(name string, value any) *RolloverBuilder {
	if b.conditions == nil {
		b.conditions = make(map[string]any)
	}
	b.conditions[name] = value
	return b
}

// Settings 配置索引或集群的设置项 (如 number_of_shards, number_of_replicas 等)。
func (b *RolloverBuilder) Settings(settings map[string]any) *RolloverBuilder {
	b.settings = settings
	return b
}

// Mappings 配置索引的映射规则 (定义字段的类型、分词器等)。
func (b *RolloverBuilder) Mappings(mappings map[string]any) *RolloverBuilder {
	b.mappings = mappings
	return b
}

// Aliases 为索引配置别名。
func (b *RolloverBuilder) Aliases(aliases map[string]any) *RolloverBuilder {
	b.aliases = aliases
	return b
}

// DryRun 开启试运行模式。开启后，请求不会对 Elasticsearch 产生实际修改，仅返回预期的操作结果。
func (b *RolloverBuilder) DryRun(enable bool) *RolloverBuilder {
	b.dryRun = &enable
	return b
}

// Debug 开启当前请求的单次调试模式。开启后，请求的完整 HTTP URL、Method、Body 及响应将被打印到日志中，便于排查问题。
func (b *RolloverBuilder) Debug(enable bool) *RolloverBuilder {
	b.setDebugPersistent(enable)
	return b
}

// Do 立即执行当前构建好的请求，并返回结构化的 Elasticsearch 响应结果或错误。请确保在调用此方法前已设置好所有必要参数。
func (b *RolloverBuilder) Do(ctx context.Context) (*RolloverResponse, error) {
	path := "/" + url.PathEscape(b.alias) + "/_rollover"
	if b.newIndex != "" {
		path += "/" + url.PathEscape(b.newIndex)
	}
	if b.dryRun != nil && *b.dryRun {
		path += "?dry_run=true"
	}

	body := make(map[string]any)
	if b.conditions != nil {
		body["conditions"] = b.conditions
	}
	if b.settings != nil {
		body["settings"] = b.settings
	}
	if b.mappings != nil {
		body["mappings"] = b.mappings
	}
	if b.aliases != nil {
		body["aliases"] = b.aliases
	}

	if b.isDebug() {
		b.printDebug(http.MethodPost, path, body)
		defer b.autoResetDebug()
	}

	var resp RolloverResponse
	if err := b.client.DoWithHeaderAndDecode(ctx, http.MethodPost, path, body, b.getHeaders(), &resp); err != nil {
		return nil, err
	}

	if b.isDebug() {
		b.printResponseObj(resp)
	}

	return &resp, nil
}
