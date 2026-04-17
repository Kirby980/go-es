package builder

import (
	"context"
	"net/http"
)

// CatBuilder 提供对 Elasticsearch _cat API 的支持
type CatBuilder struct {
	client ESClient
	debugHelper
	baseBuilder
}

func NewCatBuilder(client ESClient) *CatBuilder {
	return &CatBuilder{
		client: client,
	}
}

func (b *CatBuilder) Debug(enable bool) *CatBuilder {
	b.setDebug(enable)
	return b
}

func (b *CatBuilder) Do(ctx context.Context, api string, params map[string]string) (string, error) {
	path := "/_cat/" + api
	if len(params) > 0 {
		path += "?"
		first := true
		for k, v := range params {
			if !first {
				path += "&"
			}
			path += k + "=" + v
			first = false
		}
	}

	if b.isDebug() {
		b.printDebug("GET", path, nil)
		defer b.setDebug(false)
	}

	resp, err := b.client.DoWithHeader(ctx, http.MethodGet, path, nil, b.getHeaders())
	if err != nil {
		return "", err
	}

	return string(resp), nil
}
