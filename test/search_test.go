package builder_test

import (
	"context"
	"testing"
	"time"

	"github.com/Kirby980/go-es/builder"
	"github.com/Kirby980/go-es/client"
)

// 准备搜索测试数据
func prepareSearchTestData(t *testing.T, esClient *client.Client, indexName string) {
	ctx := context.Background()

	// 删除并创建索引
	_ = builder.NewIndexBuilder(esClient, indexName).Delete(ctx)
	_ = builder.NewIndexBuilder(esClient, indexName).
		Shards(1).
		Replicas(0).
		AddProperty("title", "text", builder.WithAnalyzer("ik_smart")).
		AddProperty("content", "text", builder.WithAnalyzer("ik_smart")).
		AddProperty("category", "keyword").
		AddProperty("tags", "keyword").
		AddProperty("price", "float").
		AddProperty("views", "integer").
		AddProperty("rating", "float").
		AddProperty("published", "boolean").
		AddProperty("created_at", "date").
		AddProperty("location", "geo_point").
		Do(ctx)

	time.Sleep(500 * time.Millisecond)

	// 插入测试数据
	documents := []map[string]any{
		{
			"title":     "iPhone 15 Pro Max",
			"content":   "最新款苹果手机，性能强劲",
			"category":  "electronics",
			"tags":      []string{"phone", "apple", "5g"},
			"price":     1299.99,
			"views":     1000,
			"rating":    4.8,
			"published": true,
			"location":  map[string]float64{"lat": 37.7749, "lon": -122.4194},
		},
		{
			"title":     "Samsung Galaxy S24",
			"content":   "三星旗舰手机",
			"category":  "electronics",
			"tags":      []string{"phone", "samsung", "5g"},
			"price":     999.99,
			"views":     800,
			"rating":    4.6,
			"published": true,
			"location":  map[string]float64{"lat": 37.7749, "lon": -122.4194},
		},
		{
			"title":     "iPad Air",
			"content":   "轻薄的平板电脑",
			"category":  "tablets",
			"tags":      []string{"tablet", "apple"},
			"price":     599.99,
			"views":     600,
			"rating":    4.7,
			"published": true,
		},
		{
			"title":     "MacBook Pro",
			"content":   "专业级笔记本电脑",
			"category":  "computers",
			"tags":      []string{"laptop", "apple", "m3"},
			"price":     1999.99,
			"views":     500,
			"rating":    4.9,
			"published": true,
		},
		{
			"title":     "Apple Watch Series 9",
			"content":   "智能手表",
			"category":  "wearables",
			"tags":      []string{"watch", "apple"},
			"price":     399.99,
			"views":     400,
			"rating":    4.5,
			"published": false,
		},
	}

	for i, doc := range documents {
		_, _ = builder.NewDocumentBuilder(esClient, indexName).
			ID(string(rune('1' + i))).
			SetMap(doc).
			Do(ctx)
	}

	time.Sleep(2 * time.Second) // 等待索引刷新
}

// TestBuilderSearchBuilder_Match 测试 Match 查询
func TestBuilderSearchBuilder_Match(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_match"
	prepareSearchTestData(t, client, indexName)
	// defer func() {
	// 	_ = builder.NewIndexBuilder(client, indexName).Delete(ctx)
	// }()

	t.Log(builder.NewSearchBuilder(client, indexName).MatchPhrase("content", "手机").Debug())
	resp, err := builder.NewSearchBuilder(client, indexName).
		MatchPhrase("content", "手机").
		Do(ctx)

	if err != nil {
		t.Fatalf("Match 查询失败: %v", err)
	}

	if resp.Hits.Total.Value == 0 {
		t.Error("应该找到匹配的文档")
	}

	t.Logf("✓ Match 查询成功: 找到 %d 条结果", resp.Hits.Total.Value)
}

// TestBuilderSearchBuilder_MatchPhrase 测试短语匹配
func TestBuilderSearchBuilder_MatchPhrase(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_matchphrase"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = builder.NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	resp, err := builder.NewSearchBuilder(client, indexName).
		MatchPhrase("title", "iPhone 15").
		Do(ctx)

	if err != nil {
		t.Fatalf("MatchPhrase 查询失败: %v", err)
	}

	t.Logf("✓ MatchPhrase 查询成功: 找到 %d 条结果", resp.Hits.Total.Value)
}

// TestBuilderSearchBuilder_Term 测试精确查询
func TestBuilderSearchBuilder_Term(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_term"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = builder.NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	resp, err := builder.NewSearchBuilder(client, indexName).
		Term("category", "electronics").
		Do(ctx)

	if err != nil {
		t.Fatalf("Term 查询失败: %v", err)
	}

	if resp.Hits.Total.Value == 0 {
		t.Error("应该找到匹配的文档")
	}

	t.Logf("✓ Term 查询成功: 找到 %d 条结果", resp.Hits.Total.Value)
}

// TestBuilderSearchBuilder_Terms 测试多值精确查询
func TestBuilderSearchBuilder_Terms(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_terms"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = builder.NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	resp, err := builder.NewSearchBuilder(client, indexName).
		Terms("category", "electronics", "tablets").
		Do(ctx)

	if err != nil {
		t.Fatalf("Terms 查询失败: %v", err)
	}

	if resp.Hits.Total.Value < 2 {
		t.Error("应该找到至少 2 个文档")
	}

	t.Logf("✓ Terms 查询成功: 找到 %d 条结果", resp.Hits.Total.Value)
}

// TestBuilderSearchBuilder_Range 测试范围查询
func TestBuilderSearchBuilder_Range(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_range"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = builder.NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	resp, err := builder.NewSearchBuilder(client, indexName).
		Range("price", 500, 1500).
		Do(ctx)

	if err != nil {
		t.Fatalf("Range 查询失败: %v", err)
	}

	t.Logf("✓ Range 查询成功: 找到 %d 条结果", resp.Hits.Total.Value)
}

// TestBuilderSearchBuilder_Exists 测试字段存在查询
func TestBuilderSearchBuilder_Exists(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_exists"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = builder.NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	resp, err := builder.NewSearchBuilder(client, indexName).
		Exists("location").
		Do(ctx)

	if err != nil {
		t.Fatalf("Exists 查询失败: %v", err)
	}

	t.Logf("✓ Exists 查询成功: 找到 %d 条有 location 字段的文档", resp.Hits.Total.Value)
}

// TestBuilderSearchBuilder_Wildcard 测试通配符查询
func TestBuilderSearchBuilder_Wildcard(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_wildcard"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = builder.NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	resp, err := builder.NewSearchBuilder(client, indexName).
		Wildcard("category", "electro*").
		Do(ctx)

	if err != nil {
		t.Fatalf("Wildcard 查询失败: %v", err)
	}

	t.Logf("✓ Wildcard 查询成功: 找到 %d 条结果", resp.Hits.Total.Value)
}

// TestBuilderSearchBuilder_Prefix 测试前缀查询
func TestBuilderSearchBuilder_Prefix(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_prefix"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = builder.NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	resp, err := builder.NewSearchBuilder(client, indexName).
		Prefix("category", "elec").
		Do(ctx)

	if err != nil {
		t.Fatalf("Prefix 查询失败: %v", err)
	}

	t.Logf("✓ Prefix 查询成功: 找到 %d 条结果", resp.Hits.Total.Value)
}

// TestBuilderSearchBuilder_Fuzzy 测试模糊查询
func TestBuilderSearchBuilder_Fuzzy(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_fuzzy"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = builder.NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	resp, err := builder.NewSearchBuilder(client, indexName).
		Fuzzy("category", "electroncs", "AUTO"). // 拼写错误
		Do(ctx)

	if err != nil {
		t.Fatalf("Fuzzy 查询失败: %v", err)
	}

	t.Logf("✓ Fuzzy 查询成功: 找到 %d 条结果", resp.Hits.Total.Value)
}

// TestBuilderSearchBuilder_MultiMatch 测试多字段匹配
func TestBuilderSearchBuilder_MultiMatch(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_multimatch"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = builder.NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	resp, err := builder.NewSearchBuilder(client, indexName).
		MultiMatch("apple", "title", "content").
		Do(ctx)

	if err != nil {
		t.Fatalf("MultiMatch 查询失败: %v", err)
	}

	t.Logf("✓ MultiMatch 查询成功: 找到 %d 条结果", resp.Hits.Total.Value)
}

// TestBuilderSearchBuilder_QueryString 测试查询字符串
func TestBuilderSearchBuilder_QueryString(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_querystring"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = builder.NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	resp, err := builder.NewSearchBuilder(client, indexName).
		QueryString("(iPhone OR Samsung) AND electronics", "title", "category").
		Do(ctx)

	if err != nil {
		t.Fatalf("QueryString 查询失败: %v", err)
	}

	t.Logf("✓ QueryString 查询成功: 找到 %d 条结果", resp.Hits.Total.Value)
}

// TestBuilderSearchBuilder_BoolQuery 测试布尔查询
func TestBuilderSearchBuilder_BoolQuery(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_bool"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = builder.NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	resp, err := builder.NewSearchBuilder(client, indexName).
		Match("title", "phone").
		Term("category", "electronics").
		Range("price", 500, 1500).
		MustNot("published", false).
		Do(ctx)

	if err != nil {
		t.Fatalf("Bool 查询失败: %v", err)
	}

	t.Logf("✓ Bool 查询成功: 找到 %d 条结果", resp.Hits.Total.Value)
}

// TestBuilderSearchBuilder_Should 测试 Should 查询
func TestBuilderSearchBuilder_Should(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_should"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = builder.NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	resp, err := builder.NewSearchBuilder(client, indexName).
		Should(
			func(b *builder.SearchBuilder) {
				b.Match("title", "iPhone")
			},
			func(b *builder.SearchBuilder) {
				b.Match("title", "Samsung")
			},
		).
		Do(ctx)

	if err != nil {
		t.Fatalf("Should 查询失败: %v", err)
	}

	t.Logf("✓ Should 查询成功: 找到 %d 条结果", resp.Hits.Total.Value)
}

// TestBuilderSearchBuilder_Sort 测试排序
func TestBuilderSearchBuilder_Sort(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_sort"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = builder.NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	resp, err := builder.NewSearchBuilder(client, indexName).
		MatchAll().
		Sort("price", "desc").
		Sort("rating", "asc").
		Do(ctx)

	if err != nil {
		t.Fatalf("Sort 查询失败: %v", err)
	}

	t.Logf("✓ Sort 查询成功: 找到 %d 条结果", resp.Hits.Total.Value)

	// 验证排序
	if len(resp.Hits.Hits) >= 2 {
		price1 := resp.Hits.Hits[0].Source["price"].(float64)
		price2 := resp.Hits.Hits[1].Source["price"].(float64)
		if price1 < price2 {
			t.Error("价格排序不正确，应该是降序")
		}
	}
}

// TestBuilderSearchBuilder_Pagination 测试分页
func TestBuilderSearchBuilder_Pagination(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_pagination"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = builder.NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	// 第一页
	resp1, err := builder.NewSearchBuilder(client, indexName).
		MatchAll().
		From(0).
		Size(2).
		Do(ctx)

	if err != nil {
		t.Fatalf("分页查询失败: %v", err)
	}

	if len(resp1.Hits.Hits) != 2 {
		t.Errorf("第一页应该返回 2 条结果, 实际=%d", len(resp1.Hits.Hits))
	}

	// 第二页
	resp2, err := builder.NewSearchBuilder(client, indexName).
		MatchAll().
		From(2).
		Size(2).
		Do(ctx)

	if err != nil {
		t.Fatalf("分页查询失败: %v", err)
	}

	t.Logf("✓ 分页查询成功: 第一页=%d 条, 第二页=%d 条",
		len(resp1.Hits.Hits), len(resp2.Hits.Hits))
}

// TestBuilderSearchBuilder_Source 测试字段过滤
func TestBuilderSearchBuilder_Source(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_source"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = builder.NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	resp, err := builder.NewSearchBuilder(client, indexName).
		MatchAll().
		Source("title", "price").
		Size(1).
		Do(ctx)

	if err != nil {
		t.Fatalf("Source 查询失败: %v", err)
	}

	if len(resp.Hits.Hits) > 0 {
		source := resp.Hits.Hits[0].Source
		if _, ok := source["title"]; !ok {
			t.Error("Source 应该包含 title 字段")
		}
		if _, ok := source["price"]; !ok {
			t.Error("Source 应该包含 price 字段")
		}
		if _, ok := source["content"]; ok {
			t.Error("Source 不应该包含 content 字段")
		}
	}

	t.Logf("✓ Source 字段过滤成功")
}

// TestBuilderSearchBuilder_Highlight 测试高亮
func TestBuilderSearchBuilder_Highlight(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_highlight"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = builder.NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	resp, err := builder.NewSearchBuilder(client, indexName).
		Match("title", "iPhone").
		Highlight("title").
		Do(ctx)

	if err != nil {
		t.Fatalf("Highlight 查询失败: %v", err)
	}

	if len(resp.Hits.Hits) > 0 {
		if resp.Hits.Hits[0].Highlight != nil {
			t.Logf("✓ Highlight 成功: %v", resp.Hits.Hits[0].Highlight)
		}
	}
}

// TestBuilderSearchBuilder_GeoDistance 测试地理距离查询
func TestBuilderSearchBuilder_GeoDistance(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_geodistance"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = builder.NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	resp, err := builder.NewSearchBuilder(client, indexName).Debug().
		GeoDistance("location", 37.7749, -122.4194, "50km").
		Do(ctx)

	if err != nil {
		t.Fatalf("GeoDistance 查询失败: %v", err)
	}

	t.Logf("✓ GeoDistance 查询成功: 找到 %d 条结果", resp.Hits.Total.Value)
}

// TestBuilderSearchBuilder_Build 测试构建查询 DSL
func TestBuilderSearchBuilder_Build(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()

	builder := builder.NewSearchBuilder(client, "test").
		Match("title", "test").
		Term("category", "electronics").
		Range("price", 100, 500).
		Sort("price", "desc").
		From(0).
		Size(10)

	dsl := builder.Build()

	// 验证 DSL 包含所有必要部分
	if dsl["query"] == nil {
		t.Error("DSL 应该包含 query")
	}
	if dsl["sort"] == nil {
		t.Error("DSL 应该包含 sort")
	}
	if dsl["from"].(int) != 0 {
		t.Error("DSL from 应该为 0")
	}
	if dsl["size"].(int) != 10 {
		t.Error("DSL size 应该为 10")
	}

	t.Logf("✓ 查询 DSL 构建成功")
	t.Logf("DSL: %+v", dsl)
}
