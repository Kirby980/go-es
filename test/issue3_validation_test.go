package builder_test

import (
	"context"
	"testing"
	"time"

	"github.com/Kirby980/go-es/builder"
)

func TestValidation_SearchAndIndex(t *testing.T) {
	client := createTestClientLocal(t)
	defer client.Close()

	// 1. 测试 SearchBuilder 验证
	// 这里通过 Debug 模式手动验证 Log 输出比较麻烦，主要看是否能正常 Build
	sb := builder.NewSearchBuilder(client, "test").From(-1).Size(-5)
	dsl := sb.Build()
	if dsl["from"].(int) != -1 || dsl["size"].(int) != -5 {
		t.Error("Values should be set even if they trigger a warning")
	}

	// 2. 测试 IndexBuilder 验证
	ib := builder.NewIndexBuilder(client, "test").Shards(0).Replicas(-1)
	body := ib.Build()
	settings := body["settings"].(map[string]any)
	if settings["number_of_shards"].(int) != 0 || settings["number_of_replicas"].(int) != -1 {
		t.Error("Settings should be set even if they trigger a warning")
	}
}

func TestSearchResponse_Helpers(t *testing.T) {
	resp := &builder.SearchResponse{}
	resp.Hits.Total.Value = 100
	resp.Hits.Total.Relation = "eq"

	if resp.Total() != 100 {
		t.Errorf("Expected 100, got %d", resp.Total())
	}
	if !resp.TotalIsExact() {
		t.Error("Expected exact relation to be true")
	}

	resp.Hits.Total.Relation = "gte"
	if resp.TotalIsExact() {
		t.Error("Expected exact relation to be false for gte")
	}
}

func TestContextPropagation_Example(t *testing.T) {
	client := createTestClientLocal(t)
	defer client.Close()

	// 使用带超时的 Context 示例
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := builder.NewSearchBuilder(client, "nonexistent").Do(ctx)
	if err == nil {
		// 如果索引不存在通常会报错，如果超时也会报错
	}
}
