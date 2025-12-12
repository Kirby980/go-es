package builder

import (
	"context"
	"testing"
	"time"

	"github.com/Kirby980/go-es/client"
)

// 准备搜索测试数据
func prepareSearchTestData(t *testing.T, esClient *client.Client, indexName string) {
	ctx := context.Background()

	// 删除并创建索引
	_ = NewIndexBuilder(esClient, indexName).Delete(ctx)
	_ = NewIndexBuilder(esClient, indexName).
		Shards(1).
		Replicas(0).
		AddProperty("title", "text", WithAnalyzer("ik_smart")).
		AddProperty("content", "text", WithAnalyzer("ik_smart")).
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
	documents := []map[string]interface{}{
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
		_, _ = NewDocumentBuilder(esClient, indexName).
			ID(string(rune('1' + i))).
			SetMap(doc).
			Do(ctx)
	}

	time.Sleep(2 * time.Second) // 等待索引刷新
}

// TestSearchBuilder_Match 测试 Match 查询
func TestSearchBuilder_Match(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_match"
	prepareSearchTestData(t, client, indexName)
	// defer func() {
	// 	_ = NewIndexBuilder(client, indexName).Delete(ctx)
	// }()

	t.Log(NewSearchBuilder(client, indexName).MatchPhrase("content", "手机").Debug())
	resp, err := NewSearchBuilder(client, indexName).
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

// TestSearchBuilder_MatchPhrase 测试短语匹配
func TestSearchBuilder_MatchPhrase(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_matchphrase"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	resp, err := NewSearchBuilder(client, indexName).
		MatchPhrase("title", "iPhone 15").
		Do(ctx)

	if err != nil {
		t.Fatalf("MatchPhrase 查询失败: %v", err)
	}

	t.Logf("✓ MatchPhrase 查询成功: 找到 %d 条结果", resp.Hits.Total.Value)
}

// TestSearchBuilder_Term 测试精确查询
func TestSearchBuilder_Term(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_term"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	resp, err := NewSearchBuilder(client, indexName).
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

// TestSearchBuilder_Terms 测试多值精确查询
func TestSearchBuilder_Terms(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_terms"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	resp, err := NewSearchBuilder(client, indexName).
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

// TestSearchBuilder_Range 测试范围查询
func TestSearchBuilder_Range(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_range"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	resp, err := NewSearchBuilder(client, indexName).
		Range("price", 500, 1500).
		Do(ctx)

	if err != nil {
		t.Fatalf("Range 查询失败: %v", err)
	}

	t.Logf("✓ Range 查询成功: 找到 %d 条结果", resp.Hits.Total.Value)
}

// TestSearchBuilder_Exists 测试字段存在查询
func TestSearchBuilder_Exists(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_exists"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	resp, err := NewSearchBuilder(client, indexName).
		Exists("location").
		Do(ctx)

	if err != nil {
		t.Fatalf("Exists 查询失败: %v", err)
	}

	t.Logf("✓ Exists 查询成功: 找到 %d 条有 location 字段的文档", resp.Hits.Total.Value)
}

// TestSearchBuilder_Wildcard 测试通配符查询
func TestSearchBuilder_Wildcard(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_wildcard"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	resp, err := NewSearchBuilder(client, indexName).
		Wildcard("category", "electro*").
		Do(ctx)

	if err != nil {
		t.Fatalf("Wildcard 查询失败: %v", err)
	}

	t.Logf("✓ Wildcard 查询成功: 找到 %d 条结果", resp.Hits.Total.Value)
}

// TestSearchBuilder_Prefix 测试前缀查询
func TestSearchBuilder_Prefix(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_prefix"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	resp, err := NewSearchBuilder(client, indexName).
		Prefix("category", "elec").
		Do(ctx)

	if err != nil {
		t.Fatalf("Prefix 查询失败: %v", err)
	}

	t.Logf("✓ Prefix 查询成功: 找到 %d 条结果", resp.Hits.Total.Value)
}

// TestSearchBuilder_Fuzzy 测试模糊查询
func TestSearchBuilder_Fuzzy(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_fuzzy"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	resp, err := NewSearchBuilder(client, indexName).
		Fuzzy("category", "electroncs", "AUTO"). // 拼写错误
		Do(ctx)

	if err != nil {
		t.Fatalf("Fuzzy 查询失败: %v", err)
	}

	t.Logf("✓ Fuzzy 查询成功: 找到 %d 条结果", resp.Hits.Total.Value)
}

// TestSearchBuilder_MultiMatch 测试多字段匹配
func TestSearchBuilder_MultiMatch(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_multimatch"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	resp, err := NewSearchBuilder(client, indexName).
		MultiMatch("apple", "title", "content").
		Do(ctx)

	if err != nil {
		t.Fatalf("MultiMatch 查询失败: %v", err)
	}

	t.Logf("✓ MultiMatch 查询成功: 找到 %d 条结果", resp.Hits.Total.Value)
}

// TestSearchBuilder_QueryString 测试查询字符串
func TestSearchBuilder_QueryString(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_querystring"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	resp, err := NewSearchBuilder(client, indexName).
		QueryString("(iPhone OR Samsung) AND electronics", "title", "category").
		Do(ctx)

	if err != nil {
		t.Fatalf("QueryString 查询失败: %v", err)
	}

	t.Logf("✓ QueryString 查询成功: 找到 %d 条结果", resp.Hits.Total.Value)
}

// TestSearchBuilder_BoolQuery 测试布尔查询
func TestSearchBuilder_BoolQuery(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_bool"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	resp, err := NewSearchBuilder(client, indexName).
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

// TestSearchBuilder_Should 测试 Should 查询
func TestSearchBuilder_Should(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_should"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	resp, err := NewSearchBuilder(client, indexName).
		Should(
			func(b *SearchBuilder) {
				b.Match("title", "iPhone")
			},
			func(b *SearchBuilder) {
				b.Match("title", "Samsung")
			},
		).
		Do(ctx)

	if err != nil {
		t.Fatalf("Should 查询失败: %v", err)
	}

	t.Logf("✓ Should 查询成功: 找到 %d 条结果", resp.Hits.Total.Value)
}

// TestSearchBuilder_Sort 测试排序
func TestSearchBuilder_Sort(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_sort"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	resp, err := NewSearchBuilder(client, indexName).
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

// TestSearchBuilder_Pagination 测试分页
func TestSearchBuilder_Pagination(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_pagination"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	// 第一页
	resp1, err := NewSearchBuilder(client, indexName).
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
	resp2, err := NewSearchBuilder(client, indexName).
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

// TestSearchBuilder_Source 测试字段过滤
func TestSearchBuilder_Source(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_source"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	resp, err := NewSearchBuilder(client, indexName).
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

// TestSearchBuilder_Highlight 测试高亮
func TestSearchBuilder_Highlight(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_highlight"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	resp, err := NewSearchBuilder(client, indexName).
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

// TestSearchBuilder_GeoDistance 测试地理距离查询
func TestSearchBuilder_GeoDistance(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_geodistance"
	prepareSearchTestData(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	resp, err := NewSearchBuilder(client, indexName).
		GeoDistance("location", 37.7749, -122.4194, "50km").
		Do(ctx)

	if err != nil {
		t.Fatalf("GeoDistance 查询失败: %v", err)
	}

	t.Logf("✓ GeoDistance 查询成功: 找到 %d 条结果", resp.Hits.Total.Value)
}

// TestSearchBuilder_Build 测试构建查询 DSL
func TestSearchBuilder_Build(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()

	builder := NewSearchBuilder(client, "test").
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
