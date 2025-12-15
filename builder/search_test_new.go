package builder

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Kirby980/go-es/client"
	"github.com/Kirby980/go-es/config"
)

// getTestClient 获取测试客户端
func getTestClient() *client.Client {
	esURL := os.Getenv("ES_URL")
	if esURL == "" {
		esURL = "https://localhost:9200"
	}

	esClient, err := client.New(
		config.WithAddresses(esURL),
		config.WithAuth("elastic", "elastic"),
		config.WithTransport(true),
	)
	if err != nil {
		panic(err)
	}
	return esClient
}

// ========== 测试新增的 Should 系列方法 ==========

func TestSearchBuilder_MatchShould(t *testing.T) {
	esClient := getTestClient()
	defer esClient.Close()
	ctx := context.Background()
	index := "test_search_match_should"

	// 创建测试索引
	err := NewIndexBuilder(esClient, index).
		Shards(1).
		Replicas(0).
		AddProperty("category", "keyword").
		AddProperty("brand", "keyword").
		AddProperty("price", "float").
		Do(ctx)
	if err != nil {
		t.Fatalf("创建索引失败: %v", err)
	}
	defer NewIndexBuilder(esClient, index).Delete(ctx)

	// 插入测试数据
	docs := []map[string]interface{}{
		{"category": "tech", "brand": "Apple", "price": 999},
		{"category": "programming", "brand": "Samsung", "price": 799},
		{"category": "database", "brand": "Huawei", "price": 699},
		{"category": "other", "brand": "Xiaomi", "price": 599},
	}

	bulk := NewBulkBuilder(esClient).Index(index)
	for i, doc := range docs {
		bulk.Add("", string(rune(i+1)), doc)
	}
	_, err = bulk.Do(ctx)
	if err != nil {
		t.Fatalf("批量插入失败: %v", err)
	}
	time.Sleep(2 * time.Second)

	// 测试：至少匹配一个 category（默认）
	resp, err := NewSearchBuilder(esClient, index).
		MatchShould("category", "tech").
		MatchShould("category", "programming").
		MatchShould("category", "database").
		Do(ctx)

	if err != nil {
		t.Fatalf("MatchShould 查询失败: %v", err)
	}

	if resp.Hits.Total.Value != 3 {
		t.Errorf("期望找到 3 条结果，实际找到 %d 条", resp.Hits.Total.Value)
	}
	t.Logf("✓ MatchShould 查询成功: 找到 %d 条结果", resp.Hits.Total.Value)
}

func TestSearchBuilder_TermShould(t *testing.T) {
	esClient := getTestClient()
	defer esClient.Close()
	ctx := context.Background()
	index := "test_search_term_should"

	// 创建测试索引
	err := NewIndexBuilder(esClient, index).
		Shards(1).
		Replicas(0).
		AddProperty("brand", "keyword").
		Do(ctx)
	if err != nil {
		t.Fatalf("创建索引失败: %v", err)
	}
	defer NewIndexBuilder(esClient, index).Delete(ctx)

	// 插入测试数据
	bulk := NewBulkBuilder(esClient).Index(index)
	bulk.Add("", "1", map[string]interface{}{"brand": "Apple"})
	bulk.Add("", "2", map[string]interface{}{"brand": "Samsung"})
	bulk.Add("", "3", map[string]interface{}{"brand": "Huawei"})
	bulk.Add("", "4", map[string]interface{}{"brand": "Xiaomi"})
	_, err = bulk.Do(ctx)
	if err != nil {
		t.Fatalf("批量插入失败: %v", err)
	}
	time.Sleep(2 * time.Second)

	// 测试：品牌是 Apple 或 Samsung
	resp, err := NewSearchBuilder(esClient, index).
		TermShould("brand", "Apple").
		TermShould("brand", "Samsung").
		Do(ctx)

	if err != nil {
		t.Fatalf("TermShould 查询失败: %v", err)
	}

	if resp.Hits.Total.Value != 2 {
		t.Errorf("期望找到 2 条结果，实际找到 %d 条", resp.Hits.Total.Value)
	}
	t.Logf("✓ TermShould 查询成功: 找到 %d 条结果", resp.Hits.Total.Value)
}

func TestSearchBuilder_RangeShould(t *testing.T) {
	esClient := getTestClient()
	defer esClient.Close()
	ctx := context.Background()
	index := "test_search_range_should"

	// 创建测试索引
	err := NewIndexBuilder(esClient, index).
		Shards(1).
		Replicas(0).
		AddProperty("price", "float").
		Do(ctx)
	if err != nil {
		t.Fatalf("创建索引失败: %v", err)
	}
	defer NewIndexBuilder(esClient, index).Delete(ctx)

	// 插入测试数据
	bulk := NewBulkBuilder(esClient).Index(index)
	bulk.Add("", "1", map[string]interface{}{"price": 100})
	bulk.Add("", "2", map[string]interface{}{"price": 500})
	bulk.Add("", "3", map[string]interface{}{"price": 1500})
	bulk.Add("", "4", map[string]interface{}{"price": 2500})
	_, err = bulk.Do(ctx)
	if err != nil {
		t.Fatalf("批量插入失败: %v", err)
	}
	time.Sleep(2 * time.Second)

	// 测试：价格在 0-300 或 1000-2000 范围内
	resp, err := NewSearchBuilder(esClient, index).
		RangeShould("price", 0, 300).
		RangeShould("price", 1000, 2000).
		Do(ctx)

	if err != nil {
		t.Fatalf("RangeShould 查询失败: %v", err)
	}

	if resp.Hits.Total.Value != 2 {
		t.Errorf("期望找到 2 条结果，实际找到 %d 条", resp.Hits.Total.Value)
	}
	t.Logf("✓ RangeShould 查询成功: 找到 %d 条结果", resp.Hits.Total.Value)
}

func TestSearchBuilder_MinimumShouldMatch(t *testing.T) {
	esClient := getTestClient()
	defer esClient.Close()
	ctx := context.Background()
	index := "test_search_minimum_should_match"

	// 创建测试索引
	err := NewIndexBuilder(esClient, index).
		Shards(1).
		Replicas(0).
		AddProperty("tag", "keyword").
		Do(ctx)
	if err != nil {
		t.Fatalf("创建索引失败: %v", err)
	}
	defer NewIndexBuilder(esClient, index).Delete(ctx)

	// 插入测试数据
	bulk := NewBulkBuilder(esClient).Index(index)
	bulk.Add("", "1", map[string]interface{}{"tag": "new"})
	bulk.Add("", "2", map[string]interface{}{"tag": "popular"})
	bulk.Add("", "3", map[string]interface{}{"tag": "recommended"})
	bulk.Add("", "4", map[string]interface{}{"tag": "trending"})
	_, err = bulk.Do(ctx)
	if err != nil {
		t.Fatalf("批量插入失败: %v", err)
	}
	time.Sleep(2 * time.Second)

	// 测试：4个 should 条件，至少匹配1个（所有都能匹配）
	resp1, err := NewSearchBuilder(esClient, index).
		TermShould("tag", "new").
		TermShould("tag", "popular").
		TermShould("tag", "recommended").
		TermShould("tag", "trending").
		MinimumShouldMatch(1).
		Do(ctx)

	if err != nil {
		t.Fatalf("MinimumShouldMatch(1) 查询失败: %v", err)
	}

	if resp1.Hits.Total.Value != 4 {
		t.Errorf("MinimumShouldMatch(1): 期望找到 4 条结果，实际找到 %d 条", resp1.Hits.Total.Value)
	}
	t.Logf("✓ MinimumShouldMatch(1) 查询成功: 找到 %d 条结果", resp1.Hits.Total.Value)

	// 测试：至少匹配2个（应该找不到，因为每个文档只有1个标签）
	resp2, err := NewSearchBuilder(esClient, index).
		TermShould("tag", "new").
		TermShould("tag", "popular").
		TermShould("tag", "recommended").
		TermShould("tag", "trending").
		MinimumShouldMatch(2).
		Do(ctx)

	if err != nil {
		t.Fatalf("MinimumShouldMatch(2) 查询失败: %v", err)
	}

	if resp2.Hits.Total.Value != 0 {
		t.Errorf("MinimumShouldMatch(2): 期望找到 0 条结果，实际找到 %d 条", resp2.Hits.Total.Value)
	}
	t.Logf("✓ MinimumShouldMatch(2) 查询成功: 找到 %d 条结果（符合预期）", resp2.Hits.Total.Value)
}

// ========== 测试新增的 MustNot 系列方法 ==========

func TestSearchBuilder_MatchMustNot(t *testing.T) {
	esClient := getTestClient()
	defer esClient.Close()
	ctx := context.Background()
	index := "test_search_match_must_not"

	// 创建测试索引
	err := NewIndexBuilder(esClient, index).
		Shards(1).
		Replicas(0).
		AddProperty("title", "text").
		Do(ctx)
	if err != nil {
		t.Fatalf("创建索引失败: %v", err)
	}
	defer NewIndexBuilder(esClient, index).Delete(ctx)

	// 插入测试数据
	bulk := NewBulkBuilder(esClient).Index(index)
	bulk.Add("", "1", map[string]interface{}{"title": "iPhone 15 Pro"})
	bulk.Add("", "2", map[string]interface{}{"title": "Samsung Galaxy S24"})
	bulk.Add("", "3", map[string]interface{}{"title": "refurbished iPhone 14"})
	bulk.Add("", "4", map[string]interface{}{"title": "Huawei Mate 60"})
	_, err = bulk.Do(ctx)
	if err != nil {
		t.Fatalf("批量插入失败: %v", err)
	}
	time.Sleep(2 * time.Second)

	// 测试：标题不能包含 "refurbished"
	resp, err := NewSearchBuilder(esClient, index).
		MatchAll().
		MatchMustNot("title", "refurbished").
		Do(ctx)

	if err != nil {
		t.Fatalf("MatchMustNot 查询失败: %v", err)
	}

	if resp.Hits.Total.Value != 3 {
		t.Errorf("期望找到 3 条结果，实际找到 %d 条", resp.Hits.Total.Value)
	}
	t.Logf("✓ MatchMustNot 查询成功: 找到 %d 条结果", resp.Hits.Total.Value)
}

func TestSearchBuilder_TermMustNot(t *testing.T) {
	esClient := getTestClient()
	defer esClient.Close()
	ctx := context.Background()
	index := "test_search_term_must_not"

	// 创建测试索引
	err := NewIndexBuilder(esClient, index).
		Shards(1).
		Replicas(0).
		AddProperty("status", "keyword").
		Do(ctx)
	if err != nil {
		t.Fatalf("创建索引失败: %v", err)
	}
	defer NewIndexBuilder(esClient, index).Delete(ctx)

	// 插入测试数据
	bulk := NewBulkBuilder(esClient).Index(index)
	bulk.Add("", "1", map[string]interface{}{"status": "active"})
	bulk.Add("", "2", map[string]interface{}{"status": "pending"})
	bulk.Add("", "3", map[string]interface{}{"status": "deleted"})
	bulk.Add("", "4", map[string]interface{}{"status": "active"})
	_, err = bulk.Do(ctx)
	if err != nil {
		t.Fatalf("批量插入失败: %v", err)
	}
	time.Sleep(2 * time.Second)

	// 测试：状态不是 deleted
	resp, err := NewSearchBuilder(esClient, index).
		MatchAll().
		TermMustNot("status", "deleted").
		Do(ctx)

	if err != nil {
		t.Fatalf("TermMustNot 查询失败: %v", err)
	}

	if resp.Hits.Total.Value != 3 {
		t.Errorf("期望找到 3 条结果，实际找到 %d 条", resp.Hits.Total.Value)
	}
	t.Logf("✓ TermMustNot 查询成功: 找到 %d 条结果", resp.Hits.Total.Value)
}

func TestSearchBuilder_RangeMustNot(t *testing.T) {
	esClient := getTestClient()
	defer esClient.Close()
	ctx := context.Background()
	index := "test_search_range_must_not"

	// 创建测试索引
	err := NewIndexBuilder(esClient, index).
		Shards(1).
		Replicas(0).
		AddProperty("age", "integer").
		Do(ctx)
	if err != nil {
		t.Fatalf("创建索引失败: %v", err)
	}
	defer NewIndexBuilder(esClient, index).Delete(ctx)

	// 插入测试数据
	bulk := NewBulkBuilder(esClient).Index(index)
	bulk.Add("", "1", map[string]interface{}{"age": 15})
	bulk.Add("", "2", map[string]interface{}{"age": 25})
	bulk.Add("", "3", map[string]interface{}{"age": 35})
	bulk.Add("", "4", map[string]interface{}{"age": 45})
	_, err = bulk.Do(ctx)
	if err != nil {
		t.Fatalf("批量插入失败: %v", err)
	}
	time.Sleep(2 * time.Second)

	// 测试：排除 18 岁以下（不包含 18）
	resp, err := NewSearchBuilder(esClient, index).
		MatchAll().
		RangeMustNot("age", nil, 17). // 排除 <= 17 岁的
		Do(ctx)

	if err != nil {
		t.Fatalf("RangeMustNot 查询失败: %v", err)
	}

	if resp.Hits.Total.Value != 3 {
		t.Errorf("期望找到 3 条结果，实际找到 %d 条", resp.Hits.Total.Value)
	}
	t.Logf("✓ RangeMustNot 查询成功: 找到 %d 条结果", resp.Hits.Total.Value)
}

// ========== 综合测试 ==========

func TestSearchBuilder_ComplexLogic(t *testing.T) {
	esClient := getTestClient()
	defer esClient.Close()
	ctx := context.Background()
	index := "test_search_complex_logic"

	// 创建测试索引
	err := NewIndexBuilder(esClient, index).
		Shards(1).
		Replicas(0).
		AddProperty("category", "keyword").
		AddProperty("status", "keyword").
		AddProperty("brand", "keyword").
		AddProperty("price", "float").
		AddProperty("title", "text").
		Do(ctx)
	if err != nil {
		t.Fatalf("创建索引失败: %v", err)
	}
	defer NewIndexBuilder(esClient, index).Delete(ctx)

	// 插入测试数据
	bulk := NewBulkBuilder(esClient).Index(index)
	bulk.Add("", "1", map[string]interface{}{
		"category": "electronics",
		"status":   "active",
		"brand":    "Apple",
		"price":    999,
		"title":    "iPhone 15 Pro",
	})
	bulk.Add("", "2", map[string]interface{}{
		"category": "electronics",
		"status":   "active",
		"brand":    "Samsung",
		"price":    899,
		"title":    "Samsung Galaxy S24",
	})
	bulk.Add("", "3", map[string]interface{}{
		"category": "electronics",
		"status":   "active",
		"brand":    "Huawei",
		"price":    799,
		"title":    "Huawei Mate 60",
	})
	bulk.Add("", "4", map[string]interface{}{
		"category": "electronics",
		"status":   "inactive",
		"brand":    "Apple",
		"price":    699,
		"title":    "refurbished iPhone 14",
	})
	bulk.Add("", "5", map[string]interface{}{
		"category": "clothing",
		"status":   "active",
		"brand":    "Nike",
		"price":    199,
		"title":    "Nike Air Max",
	})
	_, err = bulk.Do(ctx)
	if err != nil {
		t.Fatalf("批量插入失败: %v", err)
	}
	time.Sleep(2 * time.Second)

	// 综合测试：AND + OR + NOT
	// 需求：
	// 1. category 必须是 electronics (AND)
	// 2. status 必须是 active (AND)
	// 3. price 必须在 700-1000 之间 (AND)
	// 4. brand 是 Apple 或 Samsung（至少1个）(OR)
	// 5. title 不能包含 refurbished (NOT)
	resp, err := NewSearchBuilder(esClient, index).
		// AND 条件
		Term("category", "electronics").
		Term("status", "active").
		Range("price", 700, 1000).
		// OR 条件
		TermShould("brand", "Apple").
		TermShould("brand", "Samsung").
		MinimumShouldMatch(1).
		// NOT 条件
		MatchMustNot("title", "refurbished").
		Do(ctx)

	if err != nil {
		t.Fatalf("综合查询失败: %v", err)
	}

	// 应该只找到 2 条：iPhone 15 Pro (Apple, 999) 和 Samsung Galaxy S24 (Samsung, 899)
	if resp.Hits.Total.Value != 2 {
		t.Errorf("期望找到 2 条结果，实际找到 %d 条", resp.Hits.Total.Value)
	}

	t.Logf("✓ 综合查询成功: 找到 %d 条结果", resp.Hits.Total.Value)
	for _, hit := range resp.Hits.Hits {
		t.Logf("  - %v", hit.Source)
	}
}
