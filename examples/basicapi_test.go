package examples

import (
	"context"
	"testing"
	"time"

	"go-es/api"
	"go-es/client"
	"go-es/config"
)

func TestBasicAPI(t *testing.T) {
	// 创建 ES 客户端
	esClient, err := client.New(
		config.WithAddresses("http://localhost:9200"),
		config.WithTimeout(10*time.Second),
		config.WithDebug(true),
		// 如果需要认证，可以添加：
		// config.WithAuth("username", "password"),
	)
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}
	defer esClient.Close()

	ctx := context.Background()

	// 测试连接
	if err := esClient.Ping(ctx); err != nil {
		t.Fatalf("连接 ES 失败: %v", err)
	}
	t.Log("✓ 成功连接到 Elasticsearch")

	// 索引操作示例
	indexAPI := api.NewIndexAPI(esClient)

	// 创建索引
	indexName := "test-index"
	createReq := &api.CreateRequest{
		Settings: map[string]interface{}{
			"number_of_shards":   1,
			"number_of_replicas": 0,
		},
		Mappings: map[string]interface{}{
			"properties": map[string]interface{}{
				"title": map[string]interface{}{
					"type": "text",
				},
				"content": map[string]interface{}{
					"type": "text",
				},
				"created_at": map[string]interface{}{
					"type": "date",
				},
			},
		},
	}

	if err := indexAPI.Create(ctx, indexName, createReq); err != nil {
		t.Logf("创建索引失败: %v", err)
	} else {
		t.Logf("✓ 成功创建索引: %s", indexName)
	}

	// 文档操作示例
	docAPI := api.NewDocumentAPI(esClient)

	// 索引文档
	doc := map[string]interface{}{
		"title":      "Elasticsearch 入门教程",
		"content":    "这是一个关于 Elasticsearch 的入门教程",
		"created_at": time.Now().Format(time.RFC3339),
	}

	indexResp, err := docAPI.Index(ctx, indexName, "1", doc)
	if err != nil {
		t.Logf("索引文档失败: %v", err)
	} else {
		t.Logf("✓ 成功索引文档，ID: %s, 结果: %s", indexResp.ID, indexResp.Result)
	}

	// 获取文档
	getResp, err := docAPI.Get(ctx, indexName, "1")
	if err != nil {
		t.Logf("获取文档失败: %v", err)
	} else if getResp.Found {
		t.Logf("✓ 成功获取文档: %v", getResp.Source)
	}

	// 更新文档
	updateDoc := map[string]interface{}{
		"content": "这是更新后的内容",
	}

	updateResp, err := docAPI.Update(ctx, indexName, "1", updateDoc)
	if err != nil {
		t.Logf("更新文档失败: %v", err)
	} else {
		t.Logf("✓ 成功更新文档，版本: %d", updateResp.Version)
	}

	// 删除文档（可选）
	// deleteResp, err := docAPI.Delete(ctx, indexName, "1")
	// if err != nil {
	// 	t.Logf("删除文档失败: %v", err)
	// } else {
	// 	t.Logf("✓ 成功删除文档: %s", deleteResp.Result)
	// }

	// 删除索引（清理）
	// if err := indexAPI.Delete(ctx, indexName); err != nil {
	// 	t.Logf("删除索引失败: %v", err)
	// } else {
	// 	t.Logf("✓ 成功删除索引: %s", indexName)
	// }
}
