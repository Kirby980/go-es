package builder

import (
	"context"
	"net/http"
	"net/url"
)

// RolloverBuilder ...
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

// RolloverResponse ...
type RolloverResponse struct {
	Acknowledged       bool           `json:"acknowledged"`
	ShardsAcknowledged bool           `json:"shards_acknowledged"`
	OldIndex           string         `json:"old_index"`
	NewIndex           string         `json:"new_index"`
	RolledOver         bool           `json:"rolled_over"`
	DryRun             bool           `json:"dry_run"`
	Conditions         map[string]any `json:"conditions"`
}

// NewRolloverBuilder ...
func NewRolloverBuilder(c ESClient, alias string) *RolloverBuilder {
	b := &RolloverBuilder{
		client: c,
		alias:  alias,
	}
	b.initBaseBuilder()
	return b
}

// Header ...
func (b *RolloverBuilder) Header(key, value string) *RolloverBuilder {
	b.baseBuilder.Header(key, value)
	return b
}

// NewIndex ...
func (b *RolloverBuilder) NewIndex(name string) *RolloverBuilder {
	b.newIndex = name
	return b
}

// Condition ...
func (b *RolloverBuilder) Condition(name string, value any) *RolloverBuilder {
	if b.conditions == nil {
		b.conditions = make(map[string]any)
	}
	b.conditions[name] = value
	return b
}

// Settings ...
func (b *RolloverBuilder) Settings(settings map[string]any) *RolloverBuilder {
	b.settings = settings
	return b
}

// Mappings ...
func (b *RolloverBuilder) Mappings(mappings map[string]any) *RolloverBuilder {
	b.mappings = mappings
	return b
}

// Aliases ...
func (b *RolloverBuilder) Aliases(aliases map[string]any) *RolloverBuilder {
	b.aliases = aliases
	return b
}

// DryRun ...
func (b *RolloverBuilder) DryRun(enable bool) *RolloverBuilder {
	b.dryRun = &enable
	return b
}

// Debug ...
func (b *RolloverBuilder) Debug(enable bool) *RolloverBuilder {
	b.setDebugPersistent(enable)
	return b
}

// Do ...
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
