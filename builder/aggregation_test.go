package builder

import (
	"context"
	"testing"
)

// TestAggregationBuilder_MetricAggregations 测试指标聚合
func TestAggregationBuilder_MetricAggregations(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_agg_metrics"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	// 测试多种指标聚合
	resp, err := NewAggregationBuilder(client, indexName).
		Avg("avg_price", "price").
		Sum("total_views", "views").
		Min("min_price", "price").
		Max("max_price", "price").
		Stats("price_stats", "price").
		Cardinality("unique_categories", "category").
		Count("product_count", "title").
		Do(ctx)

	if err != nil {
		t.Fatalf("指标聚合失败: %v", err)
	}

	if resp.Aggregations == nil {
		t.Error("聚合结果不应该为空")
	}

	t.Logf("✓ 指标聚合成功")
	t.Logf("聚合结果: %s", resp.PrettyJSON())
}

// TestAggregationBuilder_TermsAggregation 测试分组聚合
func TestAggregationBuilder_TermsAggregation(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_agg_terms"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	// Terms 聚合
	resp, err := NewAggregationBuilder(client, indexName).
		Terms("by_category", "category", 10).
		Do(ctx)

	if err != nil {
		t.Fatalf("Terms 聚合失败: %v", err)
	}

	t.Logf("✓ Terms 聚合成功")
	t.Logf("聚合结果: %s", resp.PrettyJSON())
}

// TestAggregationBuilder_TermsWithOrder 测试带排序的分组聚合
func TestAggregationBuilder_TermsWithOrder(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_agg_terms_order"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	// 带排序的 Terms 聚合
	resp, err := NewAggregationBuilder(client, indexName).
		TermsWithOrder("top_categories", "category", 5, "_count", "desc").
		Do(ctx)

	if err != nil {
		t.Fatalf("带排序的 Terms 聚合失败: %v", err)
	}

	t.Logf("✓ 带排序的 Terms 聚合成功")
	_ = resp // 使用变量
}

// TestAggregationBuilder_Histogram 测试直方图聚合
func TestAggregationBuilder_Histogram(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_agg_histogram"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	// 直方图聚合
	_, err := NewAggregationBuilder(client, indexName).
		Histogram("price_distribution", "price", 500).
		Do(ctx)

	if err != nil {
		t.Fatalf("直方图聚合失败: %v", err)
	}

	t.Logf("✓ 直方图聚合成功")
}

// TestAggregationBuilder_Range 测试范围聚合
func TestAggregationBuilder_Range(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_agg_range"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	// 范围聚合
	resp, err := NewAggregationBuilder(client, indexName).
		Range("price_ranges", "price", []map[string]interface{}{
			{"key": "cheap", "to": 500},
			{"key": "medium", "from": 500, "to": 1000},
			{"key": "expensive", "from": 1000},
		}).
		Do(ctx)

	if err != nil {
		t.Fatalf("范围聚合失败: %v", err)
	}

	t.Logf("✓ 范围聚合成功")
	t.Logf("范围结果: %s", resp.PrettyJSON())
}

// TestAggregationBuilder_Percentiles 测试百分位聚合
func TestAggregationBuilder_Percentiles(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_agg_percentiles"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	// 百分位聚合
	_, err := NewAggregationBuilder(client, indexName).
		Percentiles("price_percentiles", "price", 25, 50, 75, 95, 99).
		Do(ctx)

	if err != nil {
		t.Fatalf("百分位聚合失败: %v", err)
	}

	t.Logf("✓ 百分位聚合成功")
}

// TestAggregationBuilder_ExtendedStats 测试扩展统计聚合
func TestAggregationBuilder_ExtendedStats(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_agg_extended_stats"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	// 扩展统计聚合
	_, err := NewAggregationBuilder(client, indexName).
		ExtendedStats("price_extended_stats", "price").
		Do(ctx)

	if err != nil {
		t.Fatalf("扩展统计聚合失败: %v", err)
	}

	t.Logf("✓ 扩展统计聚合成功")
}

// TestAggregationBuilder_Filter 测试过滤器聚合
func TestAggregationBuilder_Filter(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_agg_filter"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	// 过滤器聚合
	_, err := NewAggregationBuilder(client, indexName).
		Filter("electronics_only", map[string]interface{}{
			"term": map[string]interface{}{
				"category": "electronics",
			},
		}).
		Do(ctx)

	if err != nil {
		t.Fatalf("过滤器聚合失败: %v", err)
	}

	t.Logf("✓ 过滤器聚合成功")
}

// TestAggregationBuilder_Missing 测试缺失值聚合
func TestAggregationBuilder_Missing(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_agg_missing"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	// 缺失值聚合
	_, err := NewAggregationBuilder(client, indexName).
		Missing("missing_location", "location").
		Do(ctx)

	if err != nil {
		t.Fatalf("缺失值聚合失败: %v", err)
	}

	t.Logf("✓ 缺失值聚合成功")
}

// TestAggregationBuilder_WithQuery 测试带查询条件的聚合
func TestAggregationBuilder_WithQuery(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_agg_query"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	// 带查询条件的聚合
	_, err := NewAggregationBuilder(client, indexName).
		Query(map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": []map[string]interface{}{
					{
						"term": map[string]interface{}{
							"category": "electronics",
						},
					},
				},
			},
		}).
		Avg("avg_price", "price").
		Terms("by_tags", "tags", 10).
		Do(ctx)

	if err != nil {
		t.Fatalf("带查询条件的聚合失败: %v", err)
	}

	t.Logf("✓ 带查询条件的聚合成功")
}

// TestAggregationBuilder_SizeControl 测试结果数量控制
func TestAggregationBuilder_SizeControl(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_agg_size"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	// 设置返回0个文档，只返回聚合结果
	resp, err := NewAggregationBuilder(client, indexName).
		Size(0).
		Avg("avg_price", "price").
		Terms("categories", "category", 5).
		Do(ctx)

	if err != nil {
		t.Fatalf("聚合失败: %v", err)
	}

	t.Logf("✓ 聚合结果数量控制成功")
	t.Logf("聚合结果: %s", resp.PrettyJSON())
}

// TestAggregationBuilder_Build 测试构建聚合请求
func TestAggregationBuilder_Build(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()

	builder := NewAggregationBuilder(client, "test").
		Avg("avg_price", "price").
		Terms("by_category", "category", 10).
		Size(0)

	body := builder.Build()

	// 验证构建的请求
	if body["aggs"] == nil {
		t.Error("请求应该包含 aggs")
	}
	if body["size"].(int) != 0 {
		t.Error("size 应该为 0")
	}

	t.Logf("✓ 聚合请求构建成功")
	t.Logf("请求体: %+v", body)
}
