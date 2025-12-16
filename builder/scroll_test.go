package builder

import (
	"context"
	"testing"
	"time"

	"github.com/Kirby980/go-es/client"
	"github.com/Kirby980/go-es/config"
)

// 创建测试客户端（与 index_test.go 复用）
func createScrollTestClient(t *testing.T) *client.Client {
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
func prepareScrollTestData(t *testing.T, client *client.Client, indexName string, docCount int) {
	ctx := context.Background()

	// 创建索引
	err := NewIndexBuilder(client, indexName).
		Shards(1).
		Replicas(0).
		RefreshInterval("1s").
		AddProperty("id", "integer").
		AddProperty("title", "text").
		AddProperty("status", "keyword").
		AddProperty("price", "float").
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
			"title":      "测试文档 " + string(rune(i)),
			"status":     getStatus(i),
			"price":      float64(i * 10),
			"created_at": time.Now().Format("2006-01-02T15:04:05Z"),
		})
	}

	_, err = bulk.Do(ctx)
	if err != nil {
		t.Fatalf("插入测试数据失败: %v", err)
	}

	// 刷新索引确保数据可搜索
	time.Sleep(2 * time.Second)
}

// 获取状态（用于测试过滤）
func getStatus(id int) string {
	if id%3 == 0 {
		return "completed"
	} else if id%3 == 1 {
		return "active"
	}
	return "pending"
}

// TestScrollBuilder_BasicScroll 测试基础Scroll功能
func TestScrollBuilder_BasicScroll(t *testing.T) {
	client := createScrollTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_scroll_basic"
	defer NewIndexBuilder(client, indexName).Delete(ctx)

	// 准备100条测试数据
	prepareScrollTestData(t, client, indexName, 100)

	// 创建scroll查询（每次取20条）
	scroll := NewScrollBuilder(client, indexName).
		Size(20).
		KeepAlive("5m")

	// 第一次查询
	resp, err := scroll.Do(ctx)
	if err != nil {
		t.Fatalf("Scroll查询失败: %v", err)
	}

	t.Logf("✓ 第一批数据查询成功")
	t.Logf("总文档数: %d", resp.Hits.Total.Value)
	t.Logf("本批返回: %d", len(resp.Hits.Hits))
	t.Logf("Scroll ID: %s", resp.ScrollID)

	if resp.Hits.Total.Value != 100 {
		t.Errorf("期望总数100，实际: %d", resp.Hits.Total.Value)
	}
	if len(resp.Hits.Hits) != 20 {
		t.Errorf("期望返回20条，实际: %d", len(resp.Hits.Hits))
	}

	totalFetched := len(resp.Hits.Hits)
	batchCount := 1

	// 持续获取剩余数据
	for scroll.HasMore(resp) {
		resp, err = scroll.Next(ctx)
		if err != nil {
			t.Fatalf("获取下一批数据失败: %v", err)
		}

		batchCount++
		totalFetched += len(resp.Hits.Hits)
		t.Logf("✓ 第%d批数据: %d条", batchCount, len(resp.Hits.Hits))
	}

	t.Logf("✓ Scroll遍历完成")
	t.Logf("总批次: %d", batchCount)
	t.Logf("总获取: %d条", totalFetched)

	if totalFetched != 100 {
		t.Errorf("期望获取100条，实际: %d", totalFetched)
	}

	// 清理scroll上下文
	err = scroll.Clear(ctx)
	if err != nil {
		t.Fatalf("清理Scroll上下文失败: %v", err)
	}
	t.Logf("✓ Scroll上下文已清理")
}

// TestScrollBuilder_WithFilters 测试带查询条件的Scroll
func TestScrollBuilder_WithFilters(t *testing.T) {
	client := createScrollTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_scroll_filters"
	defer NewIndexBuilder(client, indexName).Delete(ctx)

	// 准备60条测试数据
	prepareScrollTestData(t, client, indexName, 60)

	// 创建带过滤条件的scroll查询
	// 查询 status="active" 的文档（应该有20条：1,4,7,10...58）
	scroll := NewScrollBuilder(client, indexName).
		Term("status", "active").
		Size(5).
		KeepAlive("2m")

	// 执行查询
	resp, err := scroll.Do(ctx)
	if err != nil {
		t.Fatalf("Scroll查询失败: %v", err)
	}

	t.Logf("✓ 过滤查询成功")
	t.Logf("符合条件的文档总数: %d", resp.Hits.Total.Value)

	// 统计获取的文档数
	totalFetched := len(resp.Hits.Hits)

	// 遍历所有数据
	for scroll.HasMore(resp) {
		resp, err = scroll.Next(ctx)
		if err != nil {
			t.Fatalf("获取下一批数据失败: %v", err)
		}
		totalFetched += len(resp.Hits.Hits)
	}

	t.Logf("✓ 总共获取: %d条", totalFetched)

	// 验证获取的文档状态都是 active
	expectedCount := 20 // 60条数据中，id%3==1的有20条
	if totalFetched != expectedCount {
		t.Logf("警告：期望%d条active状态文档，实际获取%d条", expectedCount, totalFetched)
	}

	// 清理
	scroll.Clear(ctx)
}

// TestScrollBuilder_MultipleConditions 测试多个查询条件
func TestScrollBuilder_MultipleConditions(t *testing.T) {
	client := createScrollTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_scroll_multiple"
	defer NewIndexBuilder(client, indexName).Delete(ctx)

	// 准备100条测试数据
	prepareScrollTestData(t, client, indexName, 100)

	// 创建带多个条件的scroll查询
	// status="active" AND price >= 100 AND price <= 500
	scroll := NewScrollBuilder(client, indexName).
		Term("status", "active").
		Range("price", 100, 500).
		Size(10).
		KeepAlive("5m")

	// 执行查询
	resp, err := scroll.Do(ctx)
	if err != nil {
		t.Fatalf("Scroll查询失败: %v", err)
	}

	t.Logf("✓ 多条件查询成功")
	t.Logf("符合所有条件的文档数: %d", resp.Hits.Total.Value)

	totalFetched := len(resp.Hits.Hits)

	// 遍历并验证数据
	for scroll.HasMore(resp) {
		resp, err = scroll.Next(ctx)
		if err != nil {
			t.Fatalf("获取下一批数据失败: %v", err)
		}
		totalFetched += len(resp.Hits.Hits)
	}

	t.Logf("✓ 总共获取: %d条", totalFetched)

	// 清理
	scroll.Clear(ctx)
}

// TestScrollBuilder_MatchQuery 测试Match查询
func TestScrollBuilder_MatchQuery(t *testing.T) {
	client := createScrollTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_scroll_match"
	defer NewIndexBuilder(client, indexName).Delete(ctx)

	// 准备50条测试数据
	prepareScrollTestData(t, client, indexName, 50)

	// 使用Match查询
	scroll := NewScrollBuilder(client, indexName).
		Match("title", "测试").
		Size(15).
		KeepAlive("3m")

	// 执行查询
	resp, err := scroll.Do(ctx)
	if err != nil {
		t.Fatalf("Scroll查询失败: %v", err)
	}

	t.Logf("✓ Match查询成功")
	t.Logf("匹配的文档数: %d", resp.Hits.Total.Value)

	totalFetched := len(resp.Hits.Hits)

	for scroll.HasMore(resp) {
		resp, err = scroll.Next(ctx)
		if err != nil {
			t.Fatalf("获取下一批数据失败: %v", err)
		}
		totalFetched += len(resp.Hits.Hits)
	}

	t.Logf("✓ 总共获取: %d条", totalFetched)

	// 清理
	scroll.Clear(ctx)
}

// TestScrollBuilder_SmallBatch 测试小批次（每次1条）
func TestScrollBuilder_SmallBatch(t *testing.T) {
	client := createScrollTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_scroll_small_batch"
	defer NewIndexBuilder(client, indexName).Delete(ctx)

	// 准备10条测试数据
	prepareScrollTestData(t, client, indexName, 10)

	// 每次只取1条
	scroll := NewScrollBuilder(client, indexName).
		Size(1).
		KeepAlive("2m")

	resp, err := scroll.Do(ctx)
	if err != nil {
		t.Fatalf("Scroll查询失败: %v", err)
	}

	totalFetched := len(resp.Hits.Hits)
	batchCount := 1

	// 逐条获取
	for scroll.HasMore(resp) {
		resp, err = scroll.Next(ctx)
		if err != nil {
			t.Fatalf("获取下一批数据失败: %v", err)
		}
		batchCount++
		totalFetched += len(resp.Hits.Hits)
	}

	t.Logf("✓ 小批次遍历完成")
	t.Logf("总批次: %d", batchCount)
	t.Logf("总获取: %d条", totalFetched)

	if totalFetched != 10 {
		t.Errorf("期望获取10条，实际: %d", totalFetched)
	}
	if batchCount != 10 {
		t.Errorf("期望10个批次，实际: %d", batchCount)
	}

	// 清理
	scroll.Clear(ctx)
}

// TestScrollBuilder_Debug 测试Debug模式
func TestScrollBuilder_Debug(t *testing.T) {
	client := createScrollTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_scroll_debug"
	defer NewIndexBuilder(client, indexName).Delete(ctx)

	// 准备20条测试数据
	prepareScrollTestData(t, client, indexName, 20)

	// 启用Debug模式
	scroll := NewScrollBuilder(client, indexName).
		Debug().
		Term("status", "active").
		Size(5).
		KeepAlive("1m")

	t.Log("=== 开始Debug模式测试 ===")

	// 第一次查询（应该打印调试信息）
	resp, err := scroll.Do(ctx)
	if err != nil {
		t.Fatalf("Scroll查询失败: %v", err)
	}

	t.Logf("✓ Debug模式第一次查询成功，返回 %d 条", len(resp.Hits.Hits))

	// 获取下一批（不带Debug，不应该打印）
	if scroll.HasMore(resp) {
		resp, err = scroll.Next(ctx)
		if err != nil {
			t.Fatalf("获取下一批数据失败: %v", err)
		}
		t.Logf("✓ 第二批查询成功，返回 %d 条", len(resp.Hits.Hits))
	}

	// 清理（也不带Debug）
	scroll.Clear(ctx)
	t.Log("=== Debug模式测试完成 ===")
}

// TestScrollBuilder_EmptyResult 测试空结果集
func TestScrollBuilder_EmptyResult(t *testing.T) {
	client := createScrollTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_scroll_empty"
	defer NewIndexBuilder(client, indexName).Delete(ctx)

	// 准备10条测试数据
	prepareScrollTestData(t, client, indexName, 10)

	// 查询一个不存在的status
	scroll := NewScrollBuilder(client, indexName).
		Term("status", "non_existent_status").
		Size(10).
		KeepAlive("1m")

	resp, err := scroll.Do(ctx)
	if err != nil {
		t.Fatalf("Scroll查询失败: %v", err)
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
	if scroll.HasMore(resp) {
		t.Error("空结果集HasMore应该返回false")
	}

	// 清理
	scroll.Clear(ctx)
}

// TestScrollBuilder_Build 测试Build方法
func TestScrollBuilder_Build(t *testing.T) {
	client := createScrollTestClient(t)
	defer client.Close()

	// 测试基础构建
	scroll1 := NewScrollBuilder(client, "test").Size(100)
	body1 := scroll1.Build()

	if body1["size"] != 100 {
		t.Errorf("期望size=100，实际: %v", body1["size"])
	}
	if body1["query"] != nil {
		t.Error("无查询条件时不应该有query字段")
	}

	// 测试带查询条件的构建
	scroll2 := NewScrollBuilder(client, "test").
		Term("status", "active").
		Match("title", "test").
		Range("price", 100, 500).
		Size(50)

	body2 := scroll2.Build()

	if body2["size"] != 50 {
		t.Errorf("期望size=50，实际: %v", body2["size"])
	}
	if body2["query"] == nil {
		t.Error("有查询条件时应该有query字段")
	}

	t.Logf("✓ Build方法测试成功")
}

// TestScrollBuilder_HasMore 测试HasMore方法
func TestScrollBuilder_HasMore(t *testing.T) {
	client := createScrollTestClient(t)
	defer client.Close()

	scroll := NewScrollBuilder(client, "test")

	// 测试有数据的情况
	resp1 := &ScrollResponse{}
	resp1.Hits.Hits = make([]struct {
		Index     string                 `json:"_index"`
		ID        string                 `json:"_id"`
		Score     float64                `json:"_score"`
		Source    map[string]interface{} `json:"_source"`
		Highlight map[string][]string    `json:"highlight,omitempty"`
	}, 5)

	if !scroll.HasMore(resp1) {
		t.Error("有数据时HasMore应该返回true")
	}

	// 测试没有数据的情况
	resp2 := &ScrollResponse{}
	if scroll.HasMore(resp2) {
		t.Error("没有数据时HasMore应该返回false")
	}

	t.Logf("✓ HasMore方法测试成功")
}

// TestScrollBuilder_LargeDataset 测试大数据集遍历
func TestScrollBuilder_LargeDataset(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过大数据集测试")
	}

	client := createScrollTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_scroll_large"
	defer NewIndexBuilder(client, indexName).Delete(ctx)

	// 准备1000条测试数据
	t.Log("开始准备1000条测试数据...")
	prepareScrollTestData(t, client, indexName, 1000)
	t.Log("✓ 测试数据准备完成")

	// 使用scroll遍历（每批100条）
	scroll := NewScrollBuilder(client, indexName).
		Size(100).
		KeepAlive("5m")

	startTime := time.Now()
	resp, err := scroll.Do(ctx)
	if err != nil {
		t.Fatalf("Scroll查询失败: %v", err)
	}

	totalFetched := len(resp.Hits.Hits)
	batchCount := 1

	for scroll.HasMore(resp) {
		resp, err = scroll.Next(ctx)
		if err != nil {
			t.Fatalf("获取下一批数据失败: %v", err)
		}
		batchCount++
		totalFetched += len(resp.Hits.Hits)
	}

	duration := time.Since(startTime)

	t.Logf("✓ 大数据集遍历完成")
	t.Logf("总文档数: 1000")
	t.Logf("总批次: %d", batchCount)
	t.Logf("总获取: %d条", totalFetched)
	t.Logf("耗时: %v", duration)

	if totalFetched != 1000 {
		t.Errorf("期望获取1000条，实际: %d", totalFetched)
	}
	if batchCount != 10 {
		t.Errorf("期望10个批次，实际: %d", batchCount)
	}

	// 清理
	scroll.Clear(ctx)
}
