package builder_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/Kirby980/go-es/builder"
	"github.com/Kirby980/go-es/logger"
)

type dslOnlyClient struct{}

func (c *dslOnlyClient) Do(ctx context.Context, method, path string, body any) ([]byte, error) {
	return []byte(`{}`), nil
}

func (c *dslOnlyClient) DoWithHeader(ctx context.Context, method, path string, body any, header http.Header) ([]byte, error) {
	return []byte(`{}`), nil
}

func (c *dslOnlyClient) DoWithHeaderAndDecode(ctx context.Context, method, path string, body any, header http.Header, target any) error {
	return nil
}

func (c *dslOnlyClient) GetAddress() string {
	return "http://localhost:9200"
}

func (c *dslOnlyClient) DoRequest(ctx context.Context, req *http.Request) ([]byte, error) {
	return []byte(`{}`), nil
}

func (c *dslOnlyClient) GetLogger() logger.Logger {
	return logger.NopLogger{}
}

func TestSearchBuilder_Should_Nesting(t *testing.T) {
	client := &dslOnlyClient{}

	// 模拟一个复杂的 Should 查询
	b := builder.NewSearchBuilder(client, "test_index").
		Should(
			func(sb *builder.SearchBuilder) {
				sb.Match("title", "iPhone").Term("category", "phone")
			},
			func(sb *builder.SearchBuilder) {
				sb.Match("title", "Samsung").Term("category", "phone")
			},
		)

	dsl := b.Build()

	query, ok := dsl["query"].(map[string]any)
	if !ok {
		t.Fatal("DSL should have 'query' field")
	}

	boolQuery, ok := query["bool"].(map[string]any)
	if !ok {
		t.Fatal("query should have 'bool' field")
	}

	should, ok := boolQuery["should"].([]map[string]any)
	if !ok {
		t.Fatal("bool should have 'should' field as slice")
	}

	if len(should) != 2 {
		t.Errorf("expected 2 should conditions, got %d", len(should))
	}

	// 验证每一项是否都是一个完整的 bool 查询（即实现了嵌套）
	for i, item := range should {
		if _, ok := item["bool"]; !ok {
			// 打印详细 DSL 以便调试
			data, _ := json.MarshalIndent(dsl, "", "  ")
			t.Errorf("Condition %d is not nested in a 'bool' query. DSL:\n%s", i, string(data))
		}
	}
}
