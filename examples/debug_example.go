package examples

import (
	"context"
	"fmt"
	"time"

	"github.com/Kirby980/go-es/builder"
	"github.com/Kirby980/go-es/client"
	"github.com/Kirby980/go-es/config"
)

// DebugExample 演示如何使用链式Debug功能
func DebugExample() {
	// 创建客户端
	esClient, err := client.New(
		config.WithAddresses("https://localhost:9200"),
		config.WithAuth("elastic", "password"),
		config.WithTransport(true),
		config.WithTimeout(10*time.Second),
	)
	if err != nil {
		fmt.Printf("创建客户端失败: %v\n", err)
		return
	}
	defer esClient.Close()

	ctx := context.Background()

	// ========== SearchBuilder Debug示例 ==========
	fmt.Println("========== SearchBuilder Debug示例 ==========")

	// 带Debug的搜索 - 会打印请求和响应
	_, _ = builder.NewSearchBuilder(esClient, "products").
		Debug().  // 启用调试模式
		Match("name", "iPhone").
		Range("price", 500, 1500).
		Sort("price", "desc").
		Size(5).
		Do(ctx)

	fmt.Println("\n不带Debug的搜索 - 不会打印任何东西")

	// 不带Debug的搜索 - 不会打印
	_, _ = builder.NewSearchBuilder(esClient, "products").
		Match("name", "Samsung").
		Size(5).
		Do(ctx)

	// ========== DocumentBuilder Debug示例 ==========
	fmt.Println("\n========== DocumentBuilder Debug示例 ==========")

	// 带Debug的文档操作
	_, _ = builder.NewDocumentBuilder(esClient, "products").
		Debug().  // 启用调试模式
		ID("1").
		Set("name", "iPhone 15 Pro").
		Set("price", 999.99).
		Do(ctx)

	// 不带Debug的文档操作
	_, _ = builder.NewDocumentBuilder(esClient, "products").
		ID("2").
		Set("name", "iPad Air").
		Set("price", 599.99).
		Do(ctx)

	// ========== BulkBuilder Debug示例 ==========
	fmt.Println("\n========== BulkBuilder Debug示例 ==========")

	// 带Debug的批量操作
	_, _ = builder.NewBulkBuilder(esClient).
		Debug().  // 启用调试模式
		Index("products").
		Add("", "3", map[string]interface{}{
			"name":  "MacBook Pro",
			"price": 1999.99,
		}).
		Add("", "4", map[string]interface{}{
			"name":  "Apple Watch",
			"price": 399.99,
		}).
		Do(ctx)

	// ========== IndexBuilder Debug示例 ==========
	fmt.Println("\n========== IndexBuilder Debug示例 ==========")

	// 带Debug的索引创建
	_ = builder.NewIndexBuilder(esClient, "test-index").
		Debug().  // 启用调试模式
		Shards(1).
		Replicas(0).
		AddProperty("title", "text").
		AddProperty("price", "float").
		Do(ctx)

	// 清理：删除测试索引
	_ = builder.NewIndexBuilder(esClient, "test-index").
		Debug().  // 删除时也可以看到debug信息
		Delete(ctx)

	// ========== AggregationBuilder Debug示例 ==========
	fmt.Println("\n========== AggregationBuilder Debug示例 ==========")

	// 带Debug的聚合查询
	_, _ = builder.NewAggregationBuilder(esClient, "products").
		Debug().  // 启用调试模式
		Avg("avg_price", "price").
		Terms("by_category", "category", 10).
		Do(ctx)

	// ========== ClusterBuilder Debug示例 ==========
	fmt.Println("\n========== ClusterBuilder Debug示例 ==========")

	clusterBuilder := builder.NewClusterBuilder(esClient)

	// 带Debug的集群健康查询
	_, _ = clusterBuilder.Debug().Health(ctx)

	// 不带Debug的集群统计查询
	_, _ = clusterBuilder.Stats(ctx)

	// 再次带Debug
	_, _ = clusterBuilder.Debug().Stats(ctx)
}
