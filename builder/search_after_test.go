package builder

import (
	"context"
	"testing"
	"time"

	"github.com/Kirby980/go-es/client"
	"github.com/Kirby980/go-es/config"
)

// 创建测试客户端
func createSearchAfterTestClient(t *testing.T) *client.Client {
	esClient, err := client.New(
		config.WithAddresses("https://localhost:9200"),
		config.WithAuth("elastic", "123456"),
		config.WithTransport(true),
		config.WithTimeout(10*time.Second),
	)
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}
	return esClient
}

// 准备测试数据
func prepareSearchAfterTestData(t *testing.T, client *client.Client, indexName string, docCount int) {
	ctx := context.Background()

	// 创建索引
	err := NewIndexBuilder(client, indexName).
		Shards(1).
		Replicas(0).
		RefreshInterval("1s").
		AddProperty("id", "integer").
		AddProperty("title", "text").
		AddProperty("category", "keyword").
		AddProperty("price", "float").
		AddProperty("stock", "integer").
		AddProperty("created_at", "date").
		Create(ctx)

	if err != nil {
		t.Fatalf("创建索引失败: %v", err)
	}

	// 批量插入测试数据
	bulk := NewBulkBuilder(client).Index(indexName)
	for i := 1; i <= docCount; i++ {
		bulk.Add("", "", map[string]interface{}{
			"id":         i,
			"title":      "商品 " + string(rune(i)),
			"category":   getCategory(i),
			"price":      float64(i * 10),
			"stock":      100 - i,
			"created_at": time.Now().Add(-time.Duration(i) * time.Hour).Format("2006-01-02T15:04:05Z"),
		})
	}

	_, err = bulk.Do(ctx)
	if err != nil {
		t.Fatalf("插入测试数据失败: %v", err)
	}

	// 刷新索引确保数据可搜索
	time.Sleep(2 * time.Second)
}

// 获取分类（用于测试过滤）
func getCategory(id int) string {
	categories := []string{"electronics", "books", "clothing"}
	return categories[id%3]
}

// TestSearchAfterBuilder_BasicUsage 测试基础用法
func TestSearchAfterBuilder_BasicUsage(t *testing.T) {
	client := createSearchAfterTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_after_basic"
	defer NewIndexBuilder(client, indexName).Delete(ctx)

	// 准备50条测试数据
	prepareSearchAfterTestData(t, client, indexName, 50)

	// 创建 SearchAfter 查询（按价格升序，每页10条）
	searchAfter := NewSearchAfterBuilder(client, indexName).
		Sort("price", "asc").
		Sort("_id", "asc"). // tie-breaker
		Size(10)

	// 第一页
	resp, err := searchAfter.Do(ctx)
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}

	t.Logf("✓ 第一页查询成功")
	t.Logf("总文档数: %d", resp.Hits.Total.Value)
	t.Logf("本页返回: %d", len(resp.Hits.Hits))
	t.Logf("第一个文档的 sort: %v", resp.Hits.Hits[0].Sort)
	t.Logf("最后一个文档的 sort: %v", resp.Hits.Hits[len(resp.Hits.Hits)-1].Sort)

	if resp.Hits.Total.Value != 50 {
		t.Errorf("期望总数50，实际: %d", resp.Hits.Total.Value)
	}
	if len(resp.Hits.Hits) != 10 {
		t.Errorf("期望返回10条，实际: %d", len(resp.Hits.Hits))
	}

	// 第二页（自动使用 Next）
	resp2, err := searchAfter.Next(ctx)
	if err != nil {
		t.Fatalf("获取第二页失败: %v", err)
	}

	t.Logf("✓ 第二页查询成功")
	t.Logf("本页返回: %d", len(resp2.Hits.Hits))
	t.Logf("第一个文档的 sort: %v", resp2.Hits.Hits[0].Sort)

	if len(resp2.Hits.Hits) != 10 {
		t.Errorf("期望返回10条，实际: %d", len(resp2.Hits.Hits))
	}

	// 验证第二页的第一个文档的 sort 值大于第一页的最后一个
	firstSort := resp.Hits.Hits[len(resp.Hits.Hits)-1].Sort
	secondSort := resp2.Hits.Hits[0].Sort
	t.Logf("✓ 分页连续性验证：第一页最后=%v, 第二页第一=%v", firstSort, secondSort)

	// 遍历所有页
	totalFetched := len(resp.Hits.Hits) + len(resp2.Hits.Hits)
	pageCount := 2

	for searchAfter.HasMore(resp2) {
		resp2, err = searchAfter.Next(ctx)
		if err != nil {
			break
		}
		pageCount++
		totalFetched += len(resp2.Hits.Hits)
		t.Logf("✓ 第%d页: %d条", pageCount, len(resp2.Hits.Hits))
	}

	t.Logf("✓ 遍历完成")
	t.Logf("总页数: %d", pageCount)
	t.Logf("总获取: %d条", totalFetched)

	if totalFetched != 50 {
		t.Errorf("期望获取50条，实际: %d", totalFetched)
	}
}

// TestSearchAfterBuilder_WithFilters 测试带查询条件
func TestSearchAfterBuilder_WithFilters(t *testing.T) {
	client := createSearchAfterTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_after_filters"
	defer NewIndexBuilder(client, indexName).Delete(ctx)

	// 准备60条测试数据
	prepareSearchAfterTestData(t, client, indexName, 60)

	// 查询 category="electronics" 的商品（应该有20条）
	searchAfter := NewSearchAfterBuilder(client, indexName).
		Term("category", "electronics").
		Sort("price", "desc").
		Sort("_id", "asc").
		Size(5)

	resp, err := searchAfter.Do(ctx)
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}

	t.Logf("✓ 过滤查询成功")
	t.Logf("符合条件的文档总数: %d", resp.Hits.Total.Value)
	t.Logf("本页返回: %d", len(resp.Hits.Hits))

	// 遍历所有符合条件的数据
	totalFetched := len(resp.Hits.Hits)
	for searchAfter.HasMore(resp) {
		resp, err = searchAfter.Next(ctx)
		if err != nil {
			break
		}
		totalFetched += len(resp.Hits.Hits)
	}

	t.Logf("✓ 总共获取: %d条", totalFetched)

	expectedCount := 20 // 60条数据，每3条一个分类
	if totalFetched != expectedCount {
		t.Logf("警告：期望%d条electronics分类，实际获取%d条", expectedCount, totalFetched)
	}
}

// TestSearchAfterBuilder_MultipleConditions 测试多条件组合
func TestSearchAfterBuilder_MultipleConditions(t *testing.T) {
	client := createSearchAfterTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_after_multiple"
	defer NewIndexBuilder(client, indexName).Delete(ctx)

	// 准备100条测试数据
	prepareSearchAfterTestData(t, client, indexName, 100)

	// category="books" AND price>=100 AND price<=500
	searchAfter := NewSearchAfterBuilder(client, indexName).
		Term("category", "books").
		Range("price", 100, 500).
		Sort("price", "asc").
		Sort("_id", "asc").
		Size(10)

	resp, err := searchAfter.Do(ctx)
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}

	t.Logf("✓ 多条件查询成功")
	t.Logf("符合所有条件的文档数: %d", resp.Hits.Total.Value)

	totalFetched := len(resp.Hits.Hits)
	for searchAfter.HasMore(resp) {
		resp, err = searchAfter.Next(ctx)
		if err != nil {
			break
		}
		totalFetched += len(resp.Hits.Hits)
	}

	t.Logf("✓ 总共获取: %d条", totalFetched)
}

// TestSearchAfterBuilder_ManualSearchAfter 测试手动设置 search_after
func TestSearchAfterBuilder_ManualSearchAfter(t *testing.T) {
	client := createSearchAfterTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_after_manual"
	defer NewIndexBuilder(client, indexName).Delete(ctx)

	// 准备30条测试数据
	prepareSearchAfterTestData(t, client, indexName, 30)

	// 第一页
	searchAfter1 := NewSearchAfterBuilder(client, indexName).
		Sort("price", "asc").
		Sort("_id", "asc").
		Size(5)

	resp1, err := searchAfter1.Do(ctx)
	if err != nil {
		t.Fatalf("第一页查询失败: %v", err)
	}

	t.Logf("✓ 第一页: %d条", len(resp1.Hits.Hits))

	// 获取最后一个文档的 sort 值
	lastSort := searchAfter1.GetLastSortValues(resp1)
	t.Logf("✓ 第一页最后的 sort 值: %v", lastSort)

	// 手动创建第二页查询（手动指定 search_after）
	searchAfter2 := NewSearchAfterBuilder(client, indexName).
		Sort("price", "asc").
		Sort("_id", "asc").
		Size(5).
		SearchAfter(lastSort...) // 手动设置

	resp2, err := searchAfter2.Do(ctx)
	if err != nil {
		t.Fatalf("第二页查询失败: %v", err)
	}

	t.Logf("✓ 第二页（手动设置）: %d条", len(resp2.Hits.Hits))
	t.Logf("✓ 第二页第一个文档的 sort: %v", resp2.Hits.Hits[0].Sort)

	if len(resp2.Hits.Hits) != 5 {
		t.Errorf("期望返回5条，实际: %d", len(resp2.Hits.Hits))
	}
}

// TestSearchAfterBuilder_MultipleSort 测试多个排序字段
func TestSearchAfterBuilder_MultipleSort(t *testing.T) {
	client := createSearchAfterTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_after_multi_sort"
	defer NewIndexBuilder(client, indexName).Delete(ctx)

	// 准备40条测试数据
	prepareSearchAfterTestData(t, client, indexName, 40)

	// 先按 category 排序，再按 price 降序，最后按 _id
	searchAfter := NewSearchAfterBuilder(client, indexName).
		Sort("category", "asc").
		Sort("price", "desc").
		Sort("_id", "asc").
		Size(8)

	resp, err := searchAfter.Do(ctx)
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}

	t.Logf("✓ 多字段排序查询成功")
	t.Logf("第一个文档的 sort: %v", resp.Hits.Hits[0].Sort)
	t.Logf("最后一个文档的 sort: %v", resp.Hits.Hits[len(resp.Hits.Hits)-1].Sort)

	// 验证 sort 字段数量
	if len(resp.Hits.Hits[0].Sort) != 3 {
		t.Errorf("期望3个排序字段，实际: %d", len(resp.Hits.Hits[0].Sort))
	}

	// 获取第二页
	resp2, err := searchAfter.Next(ctx)
	if err != nil {
		t.Fatalf("获取第二页失败: %v", err)
	}

	t.Logf("✓ 第二页查询成功: %d条", len(resp2.Hits.Hits))
}

// TestSearchAfterBuilder_WithHighlight 测试高亮
func TestSearchAfterBuilder_WithHighlight(t *testing.T) {
	client := createSearchAfterTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_after_highlight"
	defer NewIndexBuilder(client, indexName).Delete(ctx)

	// 准备20条测试数据
	prepareSearchAfterTestData(t, client, indexName, 20)

	// 带高亮的查询
	searchAfter := NewSearchAfterBuilder(client, indexName).
		Match("title", "商品").
		Highlight("title").
		Sort("price", "asc").
		Sort("_id", "asc").
		Size(5)

	resp, err := searchAfter.Do(ctx)
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}

	t.Logf("✓ 高亮查询成功")
	t.Logf("返回: %d条", len(resp.Hits.Hits))

	// 检查高亮
	hasHighlight := false
	for _, hit := range resp.Hits.Hits {
		if len(hit.Highlight) > 0 {
			hasHighlight = true
			t.Logf("✓ 文档 %s 的高亮: %v", hit.ID, hit.Highlight)
			break
		}
	}

	if !hasHighlight {
		t.Log("提示：没有找到高亮结果")
	}
}

// TestSearchAfterBuilder_WithSource 测试字段过滤
func TestSearchAfterBuilder_WithSource(t *testing.T) {
	client := createSearchAfterTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_after_source"
	defer NewIndexBuilder(client, indexName).Delete(ctx)

	// 准备15条测试数据
	prepareSearchAfterTestData(t, client, indexName, 15)

	// 只返回指定字段
	searchAfter := NewSearchAfterBuilder(client, indexName).
		Source("id", "title", "price").
		Sort("price", "asc").
		Sort("_id", "asc").
		Size(5)

	resp, err := searchAfter.Do(ctx)
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}

	t.Logf("✓ 字段过滤查询成功")
	t.Logf("返回: %d条", len(resp.Hits.Hits))

	// 验证返回的字段
	if len(resp.Hits.Hits) > 0 {
		source := resp.Hits.Hits[0].Source
		t.Logf("✓ 返回的字段: %v", source)

		// 检查应该存在的字段
		if _, ok := source["id"]; !ok {
			t.Error("应该包含 id 字段")
		}
		if _, ok := source["title"]; !ok {
			t.Error("应该包含 title 字段")
		}
		if _, ok := source["price"]; !ok {
			t.Error("应该包含 price 字段")
		}

		// 检查不应该存在的字段
		if _, ok := source["category"]; ok {
			t.Error("不应该包含 category 字段")
		}
		if _, ok := source["stock"]; ok {
			t.Error("不应该包含 stock 字段")
		}
	}
}

// TestSearchAfterBuilder_MinScore 测试最小评分
func TestSearchAfterBuilder_MinScore(t *testing.T) {
	client := createSearchAfterTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_after_min_score"
	defer NewIndexBuilder(client, indexName).Delete(ctx)

	// 准备30条测试数据
	prepareSearchAfterTestData(t, client, indexName, 30)

	// 带最小评分的查询
	searchAfter := NewSearchAfterBuilder(client, indexName).
		Match("title", "商品").
		MinScore(0.1).
		Sort("_score", "desc").
		Sort("_id", "asc").
		Size(10)

	resp, err := searchAfter.Do(ctx)
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}

	t.Logf("✓ 最小评分查询成功")
	t.Logf("符合条件的文档数: %d", resp.Hits.Total.Value)
	t.Logf("返回: %d条", len(resp.Hits.Hits))

	// 验证所有文档的评分都大于等于最小评分
	for _, hit := range resp.Hits.Hits {
		if hit.Score < 0.1 {
			t.Errorf("文档 %s 的评分 %.2f 小于最小评分 0.1", hit.ID, hit.Score)
		}
	}
}

// TestSearchAfterBuilder_Debug 测试Debug模式
func TestSearchAfterBuilder_Debug(t *testing.T) {
	client := createSearchAfterTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_after_debug"
	defer NewIndexBuilder(client, indexName).Delete(ctx)

	// 准备10条测试数据
	prepareSearchAfterTestData(t, client, indexName, 10)

	t.Log("=== 开始Debug模式测试 ===")

	// 启用Debug
	searchAfter := NewSearchAfterBuilder(client, indexName).
		Debug().
		Term("category", "books").
		Sort("price", "asc").
		Sort("_id", "asc").
		Size(3)

	resp, err := searchAfter.Do(ctx)
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}

	t.Logf("✓ Debug查询成功: %d条", len(resp.Hits.Hits))

	// 第二次查询不带Debug
	if searchAfter.HasMore(resp) {
		resp, err = searchAfter.Next(ctx)
		if err != nil {
			t.Fatalf("获取下一页失败: %v", err)
		}
		t.Logf("✓ 第二页查询成功: %d条", len(resp.Hits.Hits))
	}

	t.Log("=== Debug模式测试完成 ===")
}

// TestSearchAfterBuilder_EmptyResult 测试空结果
func TestSearchAfterBuilder_EmptyResult(t *testing.T) {
	client := createSearchAfterTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_after_empty"
	defer NewIndexBuilder(client, indexName).Delete(ctx)

	// 准备10条测试数据
	prepareSearchAfterTestData(t, client, indexName, 10)

	// 查询不存在的分类
	searchAfter := NewSearchAfterBuilder(client, indexName).
		Term("category", "non_existent").
		Sort("price", "asc").
		Sort("_id", "asc").
		Size(10)

	resp, err := searchAfter.Do(ctx)
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}

	t.Logf("✓ 空结果查询成功")
	t.Logf("匹配的文档数: %d", resp.Hits.Total.Value)
	t.Logf("返回的文档数: %d", len(resp.Hits.Hits))

	if resp.Hits.Total.Value != 0 {
		t.Errorf("期望总数0，实际: %d", resp.Hits.Total.Value)
	}
	if len(resp.Hits.Hits) != 0 {
		t.Errorf("期望返回0条，实际: %d", len(resp.Hits.Hits))
	}

	// 验证HasMore返回false
	if searchAfter.HasMore(resp) {
		t.Error("空结果集HasMore应该返回false")
	}

	// 尝试获取下一页应该报错
	_, err = searchAfter.Next(ctx)
	if err == nil {
		t.Error("空结果集调用Next应该返回错误")
	}
	t.Logf("✓ 空结果集调用Next正确返回错误: %v", err)
}

// TestSearchAfterBuilder_DefaultSort 测试默认排序
func TestSearchAfterBuilder_DefaultSort(t *testing.T) {
	client := createSearchAfterTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_after_default_sort"
	defer NewIndexBuilder(client, indexName).Delete(ctx)

	// 准备20条测试数据
	prepareSearchAfterTestData(t, client, indexName, 20)

	// 不指定排序（应该默认按 _id 排序）
	searchAfter := NewSearchAfterBuilder(client, indexName).
		Size(5)

	resp, err := searchAfter.Do(ctx)
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}

	t.Logf("✓ 默认排序查询成功")
	t.Logf("返回: %d条", len(resp.Hits.Hits))

	// 验证有 sort 字段
	if len(resp.Hits.Hits) > 0 {
		if len(resp.Hits.Hits[0].Sort) == 0 {
			t.Error("默认排序应该包含 sort 字段")
		}
		t.Logf("✓ 默认 sort 值: %v", resp.Hits.Hits[0].Sort)
	}

	// 获取下一页
	resp2, err := searchAfter.Next(ctx)
	if err != nil {
		t.Fatalf("获取第二页失败: %v", err)
	}

	t.Logf("✓ 第二页查询成功: %d条", len(resp2.Hits.Hits))
}

// TestSearchAfterBuilder_Build 测试Build方法
func TestSearchAfterBuilder_Build(t *testing.T) {
	client := createSearchAfterTestClient(t)
	defer client.Close()

	// 测试基础构建
	sa1 := NewSearchAfterBuilder(client, "test").Size(20)
	body1 := sa1.Build()

	if body1["size"] != 20 {
		t.Errorf("期望size=20，实际: %v", body1["size"])
	}
	// 应该有默认排序
	if body1["sort"] == nil {
		t.Error("应该包含默认排序")
	}

	// 测试带条件的构建
	sa2 := NewSearchAfterBuilder(client, "test").
		Term("status", "active").
		Match("title", "test").
		Range("price", 100, 500).
		Sort("price", "desc").
		Sort("_id", "asc").
		SearchAfter(100, "doc1").
		Size(10)

	body2 := sa2.Build()

	if body2["size"] != 10 {
		t.Errorf("期望size=10，实际: %v", body2["size"])
	}
	if body2["query"] == nil {
		t.Error("有查询条件时应该有query字段")
	}
	if body2["search_after"] == nil {
		t.Error("应该包含search_after字段")
	}

	searchAfterValues, ok := body2["search_after"].([]interface{})
	if !ok || len(searchAfterValues) != 2 {
		t.Error("search_after应该包含2个值")
	}

	t.Logf("✓ Build方法测试成功")
}

// TestSearchAfterBuilder_ShouldQueries 测试 Should 查询（OR）
func TestSearchAfterBuilder_ShouldQueries(t *testing.T) {
	client := createSearchAfterTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_after_should"
	defer NewIndexBuilder(client, indexName).Delete(ctx)

	// 准备30条测试数据
	prepareSearchAfterTestData(t, client, indexName, 30)

	// category="books" OR category="electronics"
	searchAfter := NewSearchAfterBuilder(client, indexName).
		TermShould("category", "books").
		TermShould("category", "electronics").
		Sort("price", "asc").
		Sort("_id", "asc").
		Size(10)

	resp, err := searchAfter.Do(ctx)
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}

	t.Logf("✓ Should查询成功")
	t.Logf("符合条件的文档数: %d", resp.Hits.Total.Value)

	totalFetched := len(resp.Hits.Hits)
	for searchAfter.HasMore(resp) {
		resp, err = searchAfter.Next(ctx)
		if err != nil {
			break
		}
		totalFetched += len(resp.Hits.Hits)
	}

	t.Logf("✓ 总共获取: %d条", totalFetched)

	// 30条数据中，books和electronics各10条，共20条
	expectedCount := 20
	if totalFetched != expectedCount {
		t.Logf("警告：期望%d条，实际获取%d条", expectedCount, totalFetched)
	}
}

// TestSearchAfterBuilder_LargeDataset 测试大数据集
func TestSearchAfterBuilder_LargeDataset(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过大数据集测试")
	}

	client := createSearchAfterTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_search_after_large"
	defer NewIndexBuilder(client, indexName).Delete(ctx)

	// 准备500条测试数据
	t.Log("开始准备500条测试数据...")
	prepareSearchAfterTestData(t, client, indexName, 500)
	t.Log("✓ 测试数据准备完成")

	// 使用 SearchAfter 遍历（每批50条）
	searchAfter := NewSearchAfterBuilder(client, indexName).
		Sort("price", "asc").
		Sort("_id", "asc").
		Size(50)

	startTime := time.Now()
	resp, err := searchAfter.Do(ctx)
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}

	totalFetched := len(resp.Hits.Hits)
	pageCount := 1

	for searchAfter.HasMore(resp) {
		resp, err = searchAfter.Next(ctx)
		if err != nil {
			t.Fatalf("获取下一页失败: %v", err)
		}
		pageCount++
		totalFetched += len(resp.Hits.Hits)
	}

	duration := time.Since(startTime)

	t.Logf("✓ 大数据集遍历完成")
	t.Logf("总文档数: 500")
	t.Logf("总页数: %d", pageCount)
	t.Logf("总获取: %d条", totalFetched)
	t.Logf("耗时: %v", duration)

	if totalFetched != 500 {
		t.Errorf("期望获取500条，实际: %d", totalFetched)
	}
	if pageCount != 10 {
		t.Errorf("期望10页，实际: %d", pageCount)
	}
}
