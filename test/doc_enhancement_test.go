package builder_test

import (
	"context"
	"testing"

	"github.com/Kirby980/go-es/builder"
)

func TestDocumentBuilder_SourceFiltering(t *testing.T) {
	client := createTestClientLocal(t)
	defer client.Close()
	ctx := context.Background()

	// 1. 准备数据
	indexName := "test_doc_source_filter"
	id := "1"
	_ = builder.NewIndexBuilder(client, indexName).Delete(ctx)
	_ = builder.NewIndexBuilder(client, indexName).
		AddProperty("title", "text").
		AddProperty("content", "text").
		Do(ctx)

	_, err := builder.NewDocumentBuilder(client, indexName).
		ID(id).
		Set("title", "Hello World").
		Set("content", "This is a test content").
		Do(ctx)
	if err != nil {
		t.Fatalf("Failed to create doc: %v", err)
	}

	// 2. 测试 SourceInclude
	resp, err := builder.NewDocumentBuilder(client, indexName).
		ID(id).
		SourceInclude("title").
		Get(ctx)

	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if _, ok := resp.Source["title"]; !ok {
		t.Error("Should include 'title'")
	}
	if _, ok := resp.Source["content"]; ok {
		t.Error("Should NOT include 'content'")
	}
}

func TestDocumentBuilder_ErrorPropagation(t *testing.T) {
	client := createTestClientLocal(t)
	defer client.Close()
	ctx := context.Background()

	// 故意构造一个会导致错误的 Model 调用 (非 struct)
	b := builder.NewDocumentBuilder(client, "test").Model(123)
	
	_, err := b.Do(ctx)
	if err == nil {
		t.Error("Should return error for invalid model")
	}
}
