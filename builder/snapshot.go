package builder

import (
	"context"
	"net/http"
	"net/url"
)

// SnapshotBuilder 是用于构建和执行 Elasticsearch Snapshot 操作的链式 API 构建器。
type SnapshotBuilder struct {
	client     ESClient
	repository string
	snapshot   string
	indices    []string
	settings   map[string]any
	debugHelper
	baseBuilder
}

// NewSnapshotBuilder 创建并返回一个 Snapshot 构建器实例，用于构造和执行 Elasticsearch 的 Snapshot 请求。
func NewSnapshotBuilder(client ESClient) *SnapshotBuilder {
	return &SnapshotBuilder{
		client: client,
	}
}

// Repository 指定快照仓库的名称。
func (b *SnapshotBuilder) Repository(repo string) *SnapshotBuilder {
	b.repository = repo
	return b
}

// Snapshot 指定快照的名称。
func (b *SnapshotBuilder) Snapshot(snapshot string) *SnapshotBuilder {
	b.snapshot = snapshot
	return b
}

// Indices 指定操作涉及的多个索引名称列表。
func (b *SnapshotBuilder) Indices(indices ...string) *SnapshotBuilder {
	b.indices = indices
	return b
}

// Settings 配置索引或集群的设置项 (如 number_of_shards, number_of_replicas 等)。
func (b *SnapshotBuilder) Settings(settings map[string]any) *SnapshotBuilder {
	b.settings = settings
	return b
}

// Debug 开启当前请求的单次调试模式。开启后，请求的完整 HTTP URL、Method、Body 及响应将被打印到日志中，便于排查问题。
func (b *SnapshotBuilder) Debug(enable bool) *SnapshotBuilder {
	b.setDebug(enable)
	return b
}

// CreateRepository 在 Elasticsearch 集群中注册并创建一个新的快照仓库。
func (b *SnapshotBuilder) CreateRepository(ctx context.Context, repoType string) (*AcknowledgedResponse, error) {
	path := "/_snapshot/" + url.PathEscape(b.repository)
	body := map[string]any{
		"type":     repoType,
		"settings": b.settings,
	}

	if b.isDebug() {
		b.printDebug("PUT", path, body)
		defer b.autoResetDebug()
	}

	var resp AcknowledgedResponse
	err := b.client.DoWithHeaderAndDecode(ctx, http.MethodPut, path, body, b.getHeaders(), &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// Create 执行创建操作。若资源已存在，可能会返回冲突错误。
func (b *SnapshotBuilder) Create(ctx context.Context) (*AcknowledgedResponse, error) {
	path := "/_snapshot/" + url.PathEscape(b.repository) + "/" + url.PathEscape(b.snapshot)

	body := map[string]any{}
	if len(b.indices) > 0 {
		body["indices"] = b.indices
	}

	if b.isDebug() {
		b.printDebug("PUT", path, body)
		defer b.autoResetDebug()
	}

	var resp AcknowledgedResponse
	err := b.client.DoWithHeaderAndDecode(ctx, http.MethodPut, path, body, b.getHeaders(), &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// Restore 从快照中恢复指定的索引或集群状态。
func (b *SnapshotBuilder) Restore(ctx context.Context) (*AcknowledgedResponse, error) {
	path := "/_snapshot/" + url.PathEscape(b.repository) + "/" + url.PathEscape(b.snapshot) + "/_restore"

	body := map[string]any{}
	if len(b.indices) > 0 {
		body["indices"] = b.indices
	}

	if b.isDebug() {
		b.printDebug("POST", path, body)
		defer b.autoResetDebug()
	}

	var resp AcknowledgedResponse
	err := b.client.DoWithHeaderAndDecode(ctx, http.MethodPost, path, body, b.getHeaders(), &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
