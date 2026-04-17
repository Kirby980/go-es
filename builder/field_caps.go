package builder

import (
	"context"
	"net/http"
	"net/url"
	"strings"
)

type FieldCapsBuilder struct {
	client         ESClient
	index          string
	fields         []string
	includeUnmapped *bool
	debugHelper
	baseBuilder
}

func NewFieldCapsBuilder(c ESClient, index string) *FieldCapsBuilder {
	b := &FieldCapsBuilder{
		client: c,
		index:  index,
	}
	b.initBaseBuilder()
	return b
}

func (b *FieldCapsBuilder) Header(key, value string) *FieldCapsBuilder {
	b.baseBuilder.Header(key, value)
	return b
}

func (b *FieldCapsBuilder) Fields(fields ...string) *FieldCapsBuilder {
	b.fields = append(b.fields, fields...)
	return b
}

func (b *FieldCapsBuilder) IncludeUnmapped(enable bool) *FieldCapsBuilder {
	b.includeUnmapped = &enable
	return b
}

func (b *FieldCapsBuilder) Debug(enable bool) *FieldCapsBuilder {
	b.setDebugPersistent(enable)
	return b
}

func (b *FieldCapsBuilder) Do(ctx context.Context) (map[string]any, error) {
	path := "/_field_caps"
	if b.index != "" {
		path = "/" + url.PathEscape(b.index) + path
	}

	values := url.Values{}
	if len(b.fields) > 0 {
		values.Set("fields", strings.Join(b.fields, ","))
	}
	if b.includeUnmapped != nil {
		if *b.includeUnmapped {
			values.Set("include_unmapped", "true")
		} else {
			values.Set("include_unmapped", "false")
		}
	}
	qs := values.Encode()
	if qs != "" {
		path += "?" + qs
	}

	if b.isDebug() {
		b.printDebug(http.MethodGet, path, nil)
		defer b.autoResetDebug()
	}

	var resp map[string]any
	if err := b.client.DoWithHeaderAndDecode(ctx, http.MethodGet, path, nil, b.getHeaders(), &resp); err != nil {
		return nil, err
	}

	if b.isDebug() {
		b.printResponseObj(resp)
	}

	return resp, nil
}
