package builder

import (
	"context"
	"net/http"
	"net/url"
)

type SnapshotBuilder struct {
	client     ESClient
	repository string
	snapshot   string
	indices    []string
	settings   map[string]any
	debugHelper
	baseBuilder
}

func NewSnapshotBuilder(client ESClient) *SnapshotBuilder {
	return &SnapshotBuilder{
		client: client,
	}
}

func (b *SnapshotBuilder) Repository(repo string) *SnapshotBuilder {
	b.repository = repo
	return b
}

func (b *SnapshotBuilder) Snapshot(snapshot string) *SnapshotBuilder {
	b.snapshot = snapshot
	return b
}

func (b *SnapshotBuilder) Indices(indices ...string) *SnapshotBuilder {
	b.indices = indices
	return b
}

func (b *SnapshotBuilder) Settings(settings map[string]any) *SnapshotBuilder {
	b.settings = settings
	return b
}

func (b *SnapshotBuilder) Debug(enable bool) *SnapshotBuilder {
	b.setDebug(enable)
	return b
}

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
