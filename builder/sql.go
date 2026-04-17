package builder

import (
	"context"
	"net/http"
)

type SQLBuilder struct {
	client ESClient
	query  string
	fetchSize int
	debugHelper
	baseBuilder
}

func NewSQLBuilder(client ESClient) *SQLBuilder {
	return &SQLBuilder{
		client: client,
	}
}

func (b *SQLBuilder) Query(query string) *SQLBuilder {
	b.query = query
	return b
}

func (b *SQLBuilder) FetchSize(size int) *SQLBuilder {
	b.fetchSize = size
	return b
}

func (b *SQLBuilder) Debug(enable bool) *SQLBuilder {
	b.setDebug(enable)
	return b
}

type SQLResponse struct {
	Columns []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"columns"`
	Rows [][]any `json:"rows"`
	Cursor string `json:"cursor,omitempty"`
}

func (b *SQLBuilder) Do(ctx context.Context) (*SQLResponse, error) {
	path := "/_sql?format=json"
	
	body := map[string]any{
		"query": b.query,
	}
	if b.fetchSize > 0 {
		body["fetch_size"] = b.fetchSize
	}

	if b.isDebug() {
		b.printDebug("POST", path, body)
		defer b.setDebug(false)
	}

	var resp SQLResponse
	err := b.client.DoWithHeaderAndDecode(ctx, http.MethodPost, path, body, b.getHeaders(), &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
