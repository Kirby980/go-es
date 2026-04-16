package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ReindexBuilder 重建索引构建器 (支持 ES 7.x/8.x/9.x)
type ReindexBuilder struct {
	client      ESClient
	sourceIndex string
	destIndex   string
	query       map[string]any // 过滤要重建的文档
	script      map[string]any // 脚本处理文档
	maxDocs     *int           // 最大处理文档数
	slices      *int           // 并发切片数
	waitForComp *bool          // 是否等待完成 (异步)
	debugHelper
	baseBuilder
}

// NewReindexBuilder 创建重建索引构建器
func NewReindexBuilder(c ESClient, sourceIndex, destIndex string) *ReindexBuilder {
	b := &ReindexBuilder{
		client:      c,
		sourceIndex: sourceIndex,
		destIndex:   destIndex,
		debugHelper: debugHelper{logger: c.GetLogger()},
	}
	b.initBaseBuilder()
	return b
}

// Header 设置自定义 Header
func (b *ReindexBuilder) Header(key, value string) *ReindexBuilder {
	b.baseBuilder.Header(key, value)
	return b
}

// Query 设置要重建的文档过滤条件 (DSL)
func (b *ReindexBuilder) Query(query map[string]any) *ReindexBuilder {
	b.query = query
	return b
}

// Script 设置对文档进行转换的脚本
func (b *ReindexBuilder) Script(script map[string]any) *ReindexBuilder {
	b.script = script
	return b
}

// MaxDocs 设置最大处理文档数量
func (b *ReindexBuilder) MaxDocs(max int) *ReindexBuilder {
	b.maxDocs = &max
	return b
}

// Slices 设置并行切片数量 (如 5，或者 0 代表 "auto")
func (b *ReindexBuilder) Slices(slices int) *ReindexBuilder {
	b.slices = &slices
	return b
}

// WaitForCompletion 设置是否等待请求完成 (设为 false 则返回 task_id 异步执行)
func (b *ReindexBuilder) WaitForCompletion(wait bool) *ReindexBuilder {
	b.waitForComp = &wait
	return b
}

// Debug 启用调试模式
func (b *ReindexBuilder) Debug() *ReindexBuilder {
	b.setDebug(true)
	return b
}

// ReindexResponse 重建索引响应
type ReindexResponse struct {
	Took             int  `json:"took"`
	TimedOut         bool `json:"timed_out"`
	Total            int  `json:"total"`
	Updated          int  `json:"updated"`
	Created          int  `json:"created"`
	Deleted          int  `json:"deleted"`
	Batches          int  `json:"batches"`
	VersionConflicts int  `json:"version_conflicts"`
	Noops            int  `json:"noops"`
	Retries          struct {
		Bulk   int `json:"bulk"`
		Search int `json:"search"`
	} `json:"retries"`
	Failures []any  `json:"failures"`
	Task     string `json:"task,omitempty"` // 如果 waitForCompletion=false 返回任务ID
}

// Build 构建请求体
func (b *ReindexBuilder) Build() map[string]any {
	body := map[string]any{
		"source": map[string]any{
			"index": b.sourceIndex,
		},
		"dest": map[string]any{
			"index": b.destIndex,
		},
	}

	if b.query != nil {
		body["source"].(map[string]any)["query"] = b.query
	}

	if b.script != nil {
		body["script"] = b.script
	}

	if b.maxDocs != nil {
		body["max_docs"] = *b.maxDocs
	}

	return body
}

// Do 执行重建索引操作
func (b *ReindexBuilder) Do(ctx context.Context) (*ReindexResponse, error) {
	path := "/_reindex"

	// 拼接查询参数
	queryParams := ""
	if b.waitForComp != nil {
		if *b.waitForComp {
			queryParams += "?wait_for_completion=true"
		} else {
			queryParams += "?wait_for_completion=false"
		}
	}
	if b.slices != nil {
		sep := "?"
		if queryParams != "" {
			sep = "&"
		}
		if *b.slices == 0 {
			queryParams += sep + "slices=auto"
		} else {
			queryParams += fmt.Sprintf("%sslices=%d", sep, *b.slices)
		}
	}
	path += queryParams

	body := b.Build()

	if b.isDebug() {
		b.printDebug("POST", path, body)
		defer b.setDebug(false)
	}

	respBody, err := b.client.DoWithHeader(ctx, http.MethodPost, path, body, b.getHeaders())
	if err != nil {
		return nil, err
	}

	if b.isDebug() {
		b.printResponse(respBody)
	}

	var resp ReindexResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &resp, nil
}
