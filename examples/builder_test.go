package examples

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Kirby980/go-es/client"
	"github.com/Kirby980/go-es/config"

	"github.com/Kirby980/go-es/builder"
)

func TestBuilder(t *testing.T) {
	// 创建客户端
	esClient, err := client.New(
		config.WithAddresses("https://localhost:9200"),
		config.WithTimeout(10*time.Second),
		config.WithAuth("elastic", "123456"),
		config.WithTransport(true),
	)
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}
	defer esClient.Close()

	ctx := context.Background()
	b, err := builder.NewIndexBuilder(esClient, "test").Get(ctx)
	if err != nil {
		t.Logf("err:%s", err)
	}
	fmt.Println(b.PrettyJSON())
	// ========== 创建索引（链式调用）==========
	// err = builder.NewIndexBuilder(esClient, "products").
	// 	Shards(1).
	// 	Replicas(0).
	// 	AddProperty("name", "text", builder.WithAnalyzer("ik_smart")).
	// 	AddProperty("price", "float").
	// 	AddProperty("category", "keyword").
	// 	AddProperty("created_at", "date", builder.WithFormat("yyyy-MM-dd HH:mm:ss")).
	// 	AddAlias("products-alias", nil).
	// 	Do(ctx)

	// if err != nil {
	// 	t.Logf("创建索引失败: %v", err)
	// } else {
	// 	t.Log("✓ 索引创建成功")
	// }

	// // ========== 索引文档（链式调用）==========
	// resp, err := builder.NewDocumentBuilder(esClient, "products").
	// 	ID("1").
	// 	Set("name", "iPhone 15 Pro").
	// 	Set("price", 999.99).
	// 	Set("category", "electronics").
	// 	Set("created_at", time.Now().Format("2006-01-02 15:04:05")).
	// 	Do(ctx)

	// if err != nil {
	// 	t.Logf("索引文档失败: %v", err)
	// } else {
	// 	t.Logf("✓ 文档索引成功: ID=%s, Result=%s", resp.ID, resp.Result)
	// }

	// ========== 搜索（链式调用）==========
	// searchResp, err := builder.NewSearchBuilder(esClient, "products").
	// 	Match("name", "iPhone").
	// 	Term("category", "electronics").
	// 	Range("price", 500, 1500).
	// 	Sort("created_at", "desc").
	// 	From(0).
	// 	Size(10).
	// 	Highlight("name").
	// 	Source("name", "price", "category").
	// 	Agg("avg_price", "avg", "price").
	// 	Do(ctx)

	// if err != nil {
	// 	t.Logf("搜索失败: %v", err)
	// } else {
	// 	t.Logf("✓ 搜索成功: 找到 %d 条结果", searchResp.Hits.Total.Value)
	// 	for _, hit := range searchResp.Hits.Hits {
	// 		t.Logf("  - [%s] %v (score: %.2f)", hit.ID, hit.Source, hit.Score)
	// 	}
	// }

	// ========== 复杂查询示例 ==========
	// complexResp, err := builder.NewSearchBuilder(esClient, "products").
	// 	Match("name", "phone").
	// 	Terms("category", "electronics", "mobile").
	// 	Range("price", nil, 2000). // 只设置上限
	// 	Exists("created_at").
	// 	Should( // 至少匹配一个条件
	// 		func(b *builder.SearchBuilder) {
	// 			b.Match("name", "iPhone")
	// 		},
	// 		func(b *builder.SearchBuilder) {
	// 			b.Match("name", "Samsung")
	// 		},
	// 	).
	// 	MustNot("category", "refurbished").
	// 	Sort("price", "asc").
	// 	Sort("created_at", "desc").
	// 	Size(20).
	// 	Do(ctx)

	// if err != nil {
	// 	t.Logf("复杂查询失败: %v", err)
	// } else {
	// 	t.Logf("✓ 复杂查询成功: %d 条结果", complexResp.Hits.Total.Value)
	// }

	// // ========== 更新文档（链式调用）==========
	// updateResp, err := builder.NewDocumentBuilder(esClient, "products").
	// 	ID("1").
	// 	Set("price", 899.99).
	// 	Set("on_sale", true).
	// 	Update(ctx)

	// if err != nil {
	// 	t.Logf("更新失败: %v", err)
	// } else {
	// 	t.Logf("✓ 更新成功: Version=%d", updateResp.Version)
	// }

	// // ========== 获取文档（链式调用）==========
	// getResp, err := builder.NewDocumentBuilder(esClient, "products").
	// 	ID("1").
	// 	Get(ctx)

	// if err != nil {
	// 	t.Logf("获取失败: %v", err)
	// } else if getResp.Found {
	// 	t.Logf("✓ 文档获取成功: %v", getResp.Source)
	// }
}
