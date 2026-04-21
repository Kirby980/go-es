package builder

import (
	"context"
	"net/http"
	"net/url"
)

// SnapshotBuilder ...
type SnapshotBuilder struct {
	client     ESClient
	repository string
	snapshot   string
	indices    []string
	settings   map[string]any
	debugHelper
	baseBuilder
}

// NewSnapshotBuilder ...
func NewSnapshotBuilder(client ESClient) *SnapshotBuilder {
	return &SnapshotBuilder{
		client: client,
	}
}

// Repository ...
func (b *SnapshotBuilder) Repository(repo string) *SnapshotBuilder {
	b.repository = repo
	return b
}

// Snapshot ...
func (b *SnapshotBuilder) Snapshot(snapshot string) *SnapshotBuilder {
	b.snapshot = snapshot
	return b
}

// Indices ...
func (b *SnapshotBuilder) Indices(indices ...string) *SnapshotBuilder {
	b.indices = indices
	return b
}

// Settings ...
func (b *SnapshotBuilder) Settings(settings map[string]any) *SnapshotBuilder {
	b.settings = settings
	return b
}

// Debug ...
func (b *SnapshotBuilder) Debug(enable bool) *SnapshotBuilder {
	b.setDebug(enable)
	return b
}

// CreateRepository ...
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

// Create ...
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

// Restore ...
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
