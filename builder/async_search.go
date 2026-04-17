package builder

import (
	"context"
	"net/http"
	"net/url"
)

type AsyncSearchBuilder struct {
	client                   ESClient
	index                    string
	query                    map[string]any
	waitForCompletionTimeout string
	keepAlive                string
	debugHelper
	baseBuilder
}

func NewAsyncSearchBuilder(client ESClient) *AsyncSearchBuilder {
	return &AsyncSearchBuilder{
		client: client,
	}
}

func (b *AsyncSearchBuilder) Index(index string) *AsyncSearchBuilder {
	b.index = index
	return b
}

func (b *AsyncSearchBuilder) Query(query map[string]any) *AsyncSearchBuilder {
	b.query = query
	return b
}

func (b *AsyncSearchBuilder) WaitForCompletionTimeout(timeout string) *AsyncSearchBuilder {
	b.waitForCompletionTimeout = timeout
	return b
}

func (b *AsyncSearchBuilder) KeepAlive(keepAlive string) *AsyncSearchBuilder {
	b.keepAlive = keepAlive
	return b
}

func (b *AsyncSearchBuilder) Debug(enable bool) *AsyncSearchBuilder {
	b.setDebug(enable)
	return b
}

type AsyncSearchResponse struct {
	ID             string         `json:"id"`
	IsPartial      bool           `json:"is_partial"`
	IsRunning      bool           `json:"is_running"`
	StartTime      int64          `json:"start_time_in_millis"`
	ExpirationTime int64          `json:"expiration_time_in_millis"`
	Response       SearchResponse `json:"response"`
}

func (b *AsyncSearchBuilder) Do(ctx context.Context) (*AsyncSearchResponse, error) {
	path := "/_async_search"
	if b.index != "" {
		path = "/" + url.PathEscape(b.index) + path
	}

	if b.waitForCompletionTimeout != "" || b.keepAlive != "" {
		path += "?"
		first := true
		if b.waitForCompletionTimeout != "" {
			path += "wait_for_completion_timeout=" + b.waitForCompletionTimeout
			first = false
		}
		if b.keepAlive != "" {
			if !first {
				path += "&"
			}
			path += "keep_alive=" + b.keepAlive
		}
	}

	body := map[string]any{
		"query": b.query,
	}

	if b.isDebug() {
		b.printDebug("POST", path, body)
		defer b.autoResetDebug()
	}

	var resp AsyncSearchResponse
	err := b.client.DoWithHeaderAndDecode(ctx, http.MethodPost, path, body, b.getHeaders(), &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (b *AsyncSearchBuilder) Get(ctx context.Context, id string) (*AsyncSearchResponse, error) {
	path := "/_async_search/" + url.PathEscape(id)

	if b.isDebug() {
		b.printDebug("GET", path, nil)
		defer b.autoResetDebug()
	}

	var resp AsyncSearchResponse
	err := b.client.DoWithHeaderAndDecode(ctx, http.MethodGet, path, nil, b.getHeaders(), &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
