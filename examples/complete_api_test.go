package examples

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go-es/builder"
	"go-es/client"
	"go-es/config"
)

func TestCompleteAPI(t *testing.T) {
	// ========== 1. 创建客户端 ==========
	esClient, err := client.New(
		config.WithAddresses("https://localhost:9200"),
		config.WithTimeout(10*time.Second),
		config.WithAuth("elastic", "123456"),
		config.WithTransport(true), // 跳过 SSL 验证
	)
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}
	defer esClient.Close()

	ctx := context.Background()

	// ========== 2. 索引管理 ==========
	t.Run("IndexManagement", func(t *testing.T) {
		indexName := "products"

		// 创建索引
		err := builder.NewIndexBuilder(esClient, indexName).
			Shards(1).
			Replicas(0).
			RefreshInterval("1s").
			AddProperty("name", "text", builder.WithAnalyzer("standard")).
			AddProperty("description", "text", builder.WithAnalyzer("standard")).
			AddProperty("price", "float").
			AddProperty("quantity", "integer").
			AddProperty("category", "keyword").
			AddProperty("tags", "keyword").
			AddProperty("rating", "float").
			AddProperty("created_at", "date", builder.WithFormat("yyyy-MM-dd HH:mm:ss")).
			AddProperty("updated_at", "date", builder.WithFormat("yyyy-MM-dd HH:mm:ss")).
			AddProperty("location", "geo_point").
			AddAlias("products-alias", nil).
			Do(ctx)

		if err != nil {
			t.Logf("创建索引: %v (可能已存在)", err)
		} else {
			t.Log("✓ 索引创建成功")
		}

		// 检查索引是否存在
		exists, _ := builder.NewIndexBuilder(esClient, indexName).Exists(ctx)
		t.Logf("索引是否存在: %v", exists)

		// 获取索引信息
		info, err := builder.NewIndexBuilder(esClient, indexName).Get(ctx)
		if err != nil {
			t.Logf("获取索引信息失败: %v", err)
		} else {
			t.Logf("索引信息: %s", info.PrettyJSON())
		}
	})

	// ========== 3. 文档操作 ==========
	t.Run("DocumentOperations", func(t *testing.T) {
		indexName := "products"

		// 创建文档
		resp, err := builder.NewDocumentBuilder(esClient, indexName).
			ID("1").
			Set("name", "iPhone 15 Pro").
			Set("description", "最新款 iPhone，性能强劲").
			Set("price", 999.99).
			Set("quantity", 100).
			Set("category", "electronics").
			Set("tags", []string{"phone", "apple", "5g"}).
			Set("rating", 4.8).
			Set("created_at", time.Now().Format("2006-01-02 15:04:05")).
			Set("location", map[string]float64{"lat": 37.7749, "lon": -122.4194}).
			Do(ctx)

		if err != nil {
			t.Logf("创建文档失败: %v", err)
		} else {
			t.Logf("✓ 文档创建成功: ID=%s, Result=%s, Version=%d", resp.ID, resp.Result, resp.Version)
		}

		// 获取文档
		getResp, err := builder.NewDocumentBuilder(esClient, indexName).
			ID("1").
			Get(ctx)

		if err != nil {
			t.Logf("获取文档失败: %v", err)
		} else if getResp.Found {
			t.Logf("✓ 文档获取成功: %v", getResp.Source)
		}

		// 更新文档
		updateResp, err := builder.NewDocumentBuilder(esClient, indexName).
			ID("1").
			Set("price", 899.99).
			Set("quantity", 95).
			Update(ctx)

		if err != nil {
			t.Logf("更新文档失败: %v", err)
		} else {
			t.Logf("✓ 文档更新成功: Version=%d", updateResp.Version)
		}

		// 使用脚本更新
		scriptResp, err := builder.NewDocumentBuilder(esClient, indexName).
			ID("1").
			Script("ctx._source.quantity -= params.count", map[string]interface{}{"count": 5}).
			Update(ctx)

		if err != nil {
			t.Logf("脚本更新失败: %v", err)
		} else {
			t.Logf("✓ 脚本更新成功: Version=%d", scriptResp.Version)
		}

		// Upsert 操作
		upsertResp, err := builder.NewDocumentBuilder(esClient, indexName).
			ID("2").
			Set("name", "MacBook Pro").
			Set("price", 1999.99).
			Set("category", "electronics").
			Upsert(ctx)

		if err != nil {
			t.Logf("Upsert 失败: %v", err)
		} else {
			t.Logf("✓ Upsert 成功: Result=%s", upsertResp.Result)
		}
	})

	// ========== 4. 批量操作 ==========
	t.Run("BulkOperations", func(t *testing.T) {
		indexName := "products"

		bulkResp, err := builder.NewBulkBuilder(esClient).
			Index(indexName).
			Add("", "3", map[string]interface{}{
				"name":     "iPad Air",
				"price":    599.99,
				"category": "electronics",
			}).
			Add("", "4", map[string]interface{}{
				"name":     "Apple Watch",
				"price":    399.99,
				"category": "wearables",
			}).
			Add("", "5", map[string]interface{}{
				"name":     "AirPods Pro",
				"price":    249.99,
				"category": "audio",
			}).
			Update("", "1", map[string]interface{}{
				"rating": 4.9,
			}).
			Do(ctx)

		if err != nil {
			t.Logf("批量操作失败: %v", err)
		} else {
			t.Logf("✓ 批量操作完成: 成功=%d, 失败=%d, 总耗时=%dms",
				bulkResp.SuccessCount(),
				len(bulkResp.FailedItems()),
				bulkResp.Took)

			if bulkResp.HasErrors() {
				for _, item := range bulkResp.FailedItems() {
					t.Logf("  失败项: ID=%s, 错误=%s", item.ID, item.Error.Reason)
				}
			}
		}
	})

	// ========== 5. 搜索操作 ==========
	t.Run("SearchOperations", func(t *testing.T) {
		indexName := "products"

		// 等待索引刷新
		time.Sleep(1 * time.Second)

		// 基础搜索
		searchResp, err := builder.NewSearchBuilder(esClient, indexName).
			Match("name", "iPhone").
			Term("category", "electronics").
			Range("price", 500, 1500).
			Sort("price", "desc").
			From(0).
			Size(10).
			Highlight("name", "description").
			Source("name", "price", "category", "rating").
			Do(ctx)

		if err != nil {
			t.Logf("搜索失败: %v", err)
		} else {
			t.Logf("✓ 搜索成功: 找到 %d 条结果 (耗时 %dms)",
				searchResp.Hits.Total.Value, searchResp.Took)
			for _, hit := range searchResp.Hits.Hits {
				t.Logf("  - [%s] %v (score: %.2f)", hit.ID, hit.Source, hit.Score)
			}
		}

		// 多字段搜索
		multiResp, err := builder.NewSearchBuilder(esClient, indexName).
			MultiMatch("Apple phone", "name", "description").
			Size(5).
			Do(ctx)

		if err != nil {
			t.Logf("多字段搜索失败: %v", err)
		} else {
			t.Logf("✓ 多字段搜索: %d 条结果", multiResp.Hits.Total.Value)
		}

		// 模糊搜索
		fuzzyResp, err := builder.NewSearchBuilder(esClient, indexName).
			Fuzzy("name", "iPhon", "AUTO").
			Do(ctx)

		if err != nil {
			t.Logf("模糊搜索失败: %v", err)
		} else {
			t.Logf("✓ 模糊搜索: %d 条结果", fuzzyResp.Hits.Total.Value)
		}

		// 前缀搜索
		prefixResp, err := builder.NewSearchBuilder(esClient, indexName).
			Prefix("name", "iPa").
			Do(ctx)

		if err != nil {
			t.Logf("前缀搜索失败: %v", err)
		} else {
			t.Logf("✓ 前缀搜索: %d 条结果", prefixResp.Hits.Total.Value)
		}

		// 复杂组合查询
		complexResp, err := builder.NewSearchBuilder(esClient, indexName).
			Should(
				func(b *builder.SearchBuilder) {
					b.Match("name", "iPhone")
				},
				func(b *builder.SearchBuilder) {
					b.Match("name", "iPad")
				},
			).
			Range("price", nil, 1000).
			MustNot("category", "refurbished").
			Sort("rating", "desc").
			Sort("price", "asc").
			Size(20).
			Do(ctx)

		if err != nil {
			t.Logf("复杂查询失败: %v", err)
		} else {
			t.Logf("✓ 复杂查询: %d 条结果", complexResp.Hits.Total.Value)
		}
	})

	// ========== 6. 聚合分析 ==========
	t.Run("AggregationOperations", func(t *testing.T) {
		indexName := "products"

		// 统计聚合
		aggResp, err := builder.NewAggregationBuilder(esClient, indexName).
			Avg("avg_price", "price").
			Sum("total_quantity", "quantity").
			Min("min_price", "price").
			Max("max_price", "price").
			Stats("price_stats", "price").
			Cardinality("unique_categories", "category").
			Do(ctx)

		if err != nil {
			t.Logf("聚合查询失败: %v", err)
		} else {
			t.Logf("✓ 聚合查询成功:")
			t.Logf("  聚合结果: %s", aggResp.PrettyJSON())
		}

		// 分组聚合
		termsResp, err := builder.NewAggregationBuilder(esClient, indexName).
			Terms("categories", "category", 10).
			Do(ctx)

		if err != nil {
			t.Logf("分组聚合失败: %v", err)
		} else {
			t.Logf("✓ 分组聚合成功: %s", termsResp.PrettyJSON())
		}

		// 范围聚合
		rangeResp, err := builder.NewAggregationBuilder(esClient, indexName).
			Range("price_ranges", "price", []map[string]interface{}{
				{"key": "cheap", "to": 300},
				{"key": "medium", "from": 300, "to": 1000},
				{"key": "expensive", "from": 1000},
			}).
			Do(ctx)

		if err != nil {
			t.Logf("范围聚合失败: %v", err)
		} else {
			t.Logf("✓ 范围聚合成功: %s", rangeResp.PrettyJSON())
		}

		// 直方图聚合
		_, err = builder.NewAggregationBuilder(esClient, indexName).
			Histogram("price_histogram", "price", 200).
			Do(ctx)

		if err != nil {
			t.Logf("直方图聚合失败: %v", err)
		} else {
			t.Logf("✓ 直方图聚合成功")
		}
	})

	// ========== 7. 集群管理 ==========
	t.Run("ClusterManagement", func(t *testing.T) {
		clusterBuilder := builder.NewClusterBuilder(esClient)

		// 集群健康
		health, err := clusterBuilder.Health(ctx)
		if err != nil {
			t.Logf("获取集群健康失败: %v", err)
		} else {
			t.Logf("✓ 集群健康:")
			t.Logf("  集群名称: %s", health.ClusterName)
			t.Logf("  状态: %s", health.Status)
			t.Logf("  节点数: %d", health.NumberOfNodes)
			t.Logf("  数据节点数: %d", health.NumberOfDataNodes)
			t.Logf("  活跃分片: %d", health.ActiveShards)
			t.Logf("  未分配分片: %d", health.UnassignedShards)
		}

		// 集群统计
		stats, err := clusterBuilder.Stats(ctx)
		if err != nil {
			t.Logf("获取集群统计失败: %v", err)
		} else {
			t.Logf("✓ 集群统计:")
			t.Logf("  索引数量: %d", stats.Indices.Count)
			t.Logf("  状态: %s", stats.Status)
		}

		// 节点信息
		nodes, err := clusterBuilder.NodesInfo(ctx)
		if err != nil {
			t.Logf("获取节点信息失败: %v", err)
		} else {
			t.Logf("✓ 节点信息: 共 %d 个节点", len(nodes.Nodes))
		}
	})

	t.Log("\n========== 完整 API 测试完成 ==========")
}

// TestRealWorldScenario 真实场景示例
func TestRealWorldScenario(t *testing.T) {
	t.Skip("真实场景示例，需要时手动运行")

	esClient, _ := client.New(
		config.WithAddresses("https://localhost:9200"),
		config.WithAuth("elastic", "123456"),
		config.WithTransport(true),
	)
	defer esClient.Close()

	ctx := context.Background()

	// 场景：电商平台的商品搜索和分析

	// 1. 搜索 "iPhone" 相关商品，价格在 500-1500 之间，按评分排序
	searchResp, _ := builder.NewSearchBuilder(esClient, "products").
		Match("name", "iPhone").
		Range("price", 500, 1500).
		Sort("rating", "desc").
		Size(10).
		Highlight("name", "description").
		Do(ctx)

	fmt.Printf("找到 %d 个商品\n", searchResp.Hits.Total.Value)

	// 2. 分析各类别商品的平均价格和数量
	aggResp, _ := builder.NewAggregationBuilder(esClient, "products").
		Terms("by_category", "category", 10).
		Do(ctx)

	fmt.Printf("聚合结果: %s\n", aggResp.PrettyJSON())

	// 3. 批量更新库存
	bulkResp, _ := builder.NewBulkBuilder(esClient).
		Index("products").
		Update("", "1", map[string]interface{}{"quantity": 80}).
		Update("", "2", map[string]interface{}{"quantity": 50}).
		Update("", "3", map[string]interface{}{"quantity": 120}).
		Do(ctx)

	fmt.Printf("批量更新: 成功 %d 个\n", bulkResp.SuccessCount())
}
