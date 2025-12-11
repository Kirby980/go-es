package builder

import (
	"context"
	"testing"
	"time"

	"go-es/client"
	"go-es/config"
)

// 创建测试客户端
func createTestClient(t *testing.T) *client.Client {
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

// TestIndexBuilder_CreateIndex 测试创建索引
func TestIndexBuilder_CreateIndex(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_index_create"

	// 清理：先删除可能存在的索引
	_ = NewIndexBuilder(client, indexName).Delete(ctx)

	// 创建索引
	err := NewIndexBuilder(client, indexName).
		Shards(1).
		Replicas(0).
		RefreshInterval("1s").
		AddProperty("title", "text", WithAnalyzer("standard")).
		AddProperty("content", "text", WithAnalyzer("standard")).
		AddProperty("author", "keyword").
		AddProperty("views", "integer").
		AddProperty("price", "float").
		AddProperty("published", "boolean").
		AddProperty("created_at", "date", WithFormat("yyyy-MM-dd HH:mm:ss")).
		AddProperty("tags", "keyword").
		AddAlias("test_alias", nil).
		Do(ctx)

	if err != nil {
		t.Fatalf("创建索引失败: %v", err)
	}
	t.Logf("✓ 创建索引成功: %s", indexName)

	// 清理
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()
}

// TestIndexBuilder_Exists 测试索引是否存在
func TestIndexBuilder_Exists(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_index_exists"

	// 确保索引不存在
	//_ = NewIndexBuilder(client, indexName).Delete(ctx)

	// 检查不存在的索引
	exists, err := NewIndexBuilder(client, indexName).Exists(ctx)
	if err != nil {
		t.Fatalf("检查索引存在性失败: %v", err)
	}
	if exists {
		t.Error("索引不应该存在")
	}
	t.Logf("✓ 确认索引不存在")

	// 创建索引
	_ = NewIndexBuilder(client, indexName).
		Shards(1).
		Replicas(0).
		Do(ctx)

	// 检查存在的索引
	exists, err = NewIndexBuilder(client, indexName).Exists(ctx)
	if err != nil {
		t.Fatalf("检查索引存在性失败: %v", err)
	}
	if !exists {
		t.Error("索引应该存在")
	}
	t.Logf("✓ 确认索引存在")

	// 清理
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()
}

// TestIndexBuilder_GetIndexInfo 测试获取索引信息
func TestIndexBuilder_GetIndexInfo(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_index_create"

	// 创建索引
	_ = NewIndexBuilder(client, indexName).
		Shards(1).
		Replicas(0).
		AddProperty("name", "text").
		AddProperty("age", "integer").
		Do(ctx)

	// 获取索引信息
	info, err := NewIndexBuilder(client, indexName).Get(ctx)
	if err != nil {
		t.Fatalf("获取索引信息失败: %v", err)
	}

	t.Logf("✓ 获取索引信息成功")
	t.Logf("索引信息: %s", info.PrettyJSON())

	// 验证索引信息包含必要字段
	if info.Mappings == nil {
		t.Error("索引信息应该包含 mappings")
	}
	if info.Settings == nil {
		t.Error("索引信息应该包含 settings")
	}

	// 清理
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()
}

// TestIndexBuilder_DeleteIndex 测试删除索引
func TestIndexBuilder_DeleteIndex(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_index_delete"

	// 创建索引
	_ = NewIndexBuilder(client, indexName).
		Shards(1).
		Replicas(0).
		Do(ctx)

	// 确认索引存在
	exists, _ := NewIndexBuilder(client, indexName).Exists(ctx)
	if !exists {
		t.Fatal("索引应该存在")
	}

	// 删除索引
	err := NewIndexBuilder(client, indexName).Delete(ctx)
	if err != nil {
		t.Fatalf("删除索引失败: %v", err)
	}
	t.Logf("✓ 删除索引成功")

	// 确认索引已删除
	exists, _ = NewIndexBuilder(client, indexName).Exists(ctx)
	if exists {
		t.Error("索引不应该存在")
	}
}

// TestIndexBuilder_PropertyOptions 测试字段选项
func TestIndexBuilder_PropertyOptions(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_index_properties"
	_ = NewIndexBuilder(client, indexName).Delete(ctx)

	// 创建包含各种字段选项的索引
	err := NewIndexBuilder(client, indexName).
		Shards(1).
		Replicas(0).
		// 文本字段带分词器
		AddProperty("description", "text", WithAnalyzer("standard")).
		// 关键字字段
		AddProperty("keyword_field", "keyword").
		// 数值字段
		AddProperty("integer_field", "integer").
		AddProperty("float_field", "float").
		// 日期字段带格式
		AddProperty("date_field", "date", WithFormat("yyyy-MM-dd")).
		AddProperty("datetime_field", "date", WithFormat("yyyy-MM-dd HH:mm:ss")).
		// 布尔字段
		AddProperty("boolean_field", "boolean").
		// 带存储选项的字段
		AddProperty("stored_field", "text", WithStore(true)).
		// 不索引的字段
		AddProperty("not_indexed_field", "text", WithIndex(false)).
		Do(ctx)

	if err != nil {
		t.Fatalf("创建索引失败: %v", err)
	}
	t.Logf("✓ 创建包含多种字段类型的索引成功")

	// 获取索引信息验证
	info, _ := NewIndexBuilder(client, indexName).Get(ctx)
	t.Logf("索引映射: %s", info.PrettyJSON())

	// 清理
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()
}

// TestIndexBuilder_MultipleAliases 测试多个别名
func TestIndexBuilder_MultipleAliases(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_index_aliases"
	_ = NewIndexBuilder(client, indexName).Delete(ctx)

	// 创建带多个别名的索引
	err := NewIndexBuilder(client, indexName).
		Shards(1).
		Replicas(0).
		AddProperty("status", "keyword").
		AddAlias("alias1", nil).
		AddAlias("alias2", nil).
		AddAlias("filtered_alias", map[string]interface{}{
			"term": map[string]interface{}{
				"status": "active",
			},
		}).
		Do(ctx)

	if err != nil {
		t.Fatalf("创建索引失败: %v", err)
	}
	t.Logf("✓ 创建带多个别名的索引成功")

	// 获取索引信息验证别名
	info, _ := NewIndexBuilder(client, indexName).Get(ctx)
	if info.Aliases == nil {
		t.Error("索引应该包含别名")
	}
	t.Logf("索引别名: %v", info.Aliases)

	// 清理
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()
}

// TestIndexBuilder_RefreshInterval 测试刷新间隔
func TestIndexBuilder_RefreshInterval(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_index_refresh"
	_ = NewIndexBuilder(client, indexName).Delete(ctx)

	// 创建带自定义刷新间隔的索引
	err := NewIndexBuilder(client, indexName).
		Shards(1).
		Replicas(0).
		RefreshInterval("5s").
		Do(ctx)

	if err != nil {
		t.Fatalf("创建索引失败: %v", err)
	}
	t.Logf("✓ 创建带自定义刷新间隔的索引成功")

	// 清理
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()
}

// TestIndexBuilder_Build 测试构建索引定义
func TestIndexBuilder_Build(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()

	builder := NewIndexBuilder(client, "test").
		Shards(3).
		Replicas(1).
		RefreshInterval("30s").
		AddProperty("title", "text").
		AddProperty("price", "float").
		AddAlias("test_alias", nil)

	definition := builder.Build()

	// 验证定义包含所有必要部分
	if definition["settings"] == nil {
		t.Error("定义应该包含 settings")
	}
	if definition["mappings"] == nil {
		t.Error("定义应该包含 mappings")
	}
	if definition["aliases"] == nil {
		t.Error("定义应该包含 aliases")
	}

	t.Logf("✓ 索引定义构建成功")
	t.Logf("定义: %+v", definition)
}

// TestIndexBuilder_ChainCalls 测试链式调用
func TestIndexBuilder_ChainCalls(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_index_chain"
	_ = NewIndexBuilder(client, indexName).Delete(ctx)

	// 测试所有方法都返回 *IndexBuilder 以支持链式调用
	builder := NewIndexBuilder(client, indexName)
	builder = builder.Shards(1)
	builder = builder.Replicas(0)
	builder = builder.RefreshInterval("1s")
	builder = builder.AddProperty("field1", "text")
	builder = builder.AddProperty("field2", "keyword")
	builder = builder.AddAlias("alias1", nil)

	err := builder.Do(ctx)
	if err != nil {
		t.Fatalf("链式调用创建索引失败: %v", err)
	}
	t.Logf("✓ 链式调用测试成功")

	// 清理
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()
}

// TestIndexBuilder_WithFields 测试 WithSubField 多字段映射
func TestIndexBuilder_WithFields(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_index_with_fields"
	_ = NewIndexBuilder(client, indexName).Delete(ctx)

	// 使用 WithSubField 链式调用添加多字段，无需手写 map
	err := NewIndexBuilder(client, indexName).
		Shards(1).
		Replicas(0).
		AddProperty("title", "text",
			WithAnalyzer("standard"),
			WithSubField("keyword", "keyword", WithIgnoreAbove(256)),
			WithSubField("raw", "keyword"),
		).
		AddProperty("description", "text",
			WithSubField("keyword", "keyword"),
		).
		Do(ctx)

	if err != nil {
		t.Fatalf("创建带多字段映射的索引失败: %v", err)
	}
	t.Logf("✓ 创建带多字段映射的索引成功")

	// 获取索引信息验证 fields
	info, err := NewIndexBuilder(client, indexName).Get(ctx)
	if err != nil {
		t.Fatalf("获取索引信息失败: %v", err)
	}
	t.Logf("索引映射信息: %s", info.PrettyJSON())

	// 清理
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()
}

// TestIndexBuilder_MultiplePropertyOptions 测试组合使用多个 PropertyOption
func TestIndexBuilder_MultiplePropertyOptions(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_index_multiple_options"
	_ = NewIndexBuilder(client, indexName).Delete(ctx)

	// 测试在同一个字段上使用多个选项（使用链式 API）
	err := NewIndexBuilder(client, indexName).
		Shards(1).
		Replicas(0).
		// 组合使用多个选项
		AddProperty("multi_option_field", "text",
			WithAnalyzer("standard"),
			WithStore(true),
			WithSubField("keyword", "keyword"),
		).
		// 日期字段组合选项
		AddProperty("timestamp", "date",
			WithFormat("yyyy-MM-dd HH:mm:ss||epoch_millis"),
			WithStore(true),
		).
		// 文本字段组合选项
		AddProperty("content", "text",
			WithAnalyzer("standard"),
			WithIndex(true),
			WithStore(false),
		).
		Do(ctx)

	if err != nil {
		t.Fatalf("创建包含组合选项的索引失败: %v", err)
	}
	t.Logf("✓ 创建包含组合选项的索引成功")

	// 获取索引信息验证
	info, _ := NewIndexBuilder(client, indexName).Get(ctx)
	t.Logf("索引映射: %s", info.PrettyJSON())

	// 清理
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()
}

// TestIndexInfo_JSONMethods 测试 IndexInfo 的 JSON 方法
func TestIndexInfo_JSONMethods(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_index_json_methods"
	_ = NewIndexBuilder(client, indexName).Delete(ctx)

	// 创建索引
	err := NewIndexBuilder(client, indexName).
		Shards(1).
		Replicas(0).
		AddProperty("name", "text").
		AddProperty("age", "integer").
		AddAlias("json_test_alias", nil).
		Do(ctx)

	if err != nil {
		t.Fatalf("创建索引失败: %v", err)
	}

	// 获取索引信息
	info, err := NewIndexBuilder(client, indexName).Get(ctx)
	if err != nil {
		t.Fatalf("获取索引信息失败: %v", err)
	}

	// 测试 JSON() 方法（紧凑格式）
	compactJSON := info.JSON()
	if compactJSON == "" {
		t.Error("JSON() 应该返回非空字符串")
	}
	t.Logf("✓ JSON() 方法测试成功")
	t.Logf("紧凑 JSON: %s", compactJSON)

	// 测试 PrettyJSON() 方法（格式化）
	prettyJSON := info.PrettyJSON()
	if prettyJSON == "" {
		t.Error("PrettyJSON() 应该返回非空字符串")
	}
	// 格式化的 JSON 应该比紧凑的长（因为有缩进和换行）
	if len(prettyJSON) <= len(compactJSON) {
		t.Error("PrettyJSON() 应该返回格式化的 JSON（比紧凑格式长）")
	}
	t.Logf("✓ PrettyJSON() 方法测试成功")
	t.Logf("格式化 JSON:\n%s", prettyJSON)

	// 测试 String() 方法（应该等同于 PrettyJSON）
	stringOutput := info.String()
	if stringOutput != prettyJSON {
		t.Error("String() 应该返回与 PrettyJSON() 相同的结果")
	}
	t.Logf("✓ String() 方法测试成功")

	// 清理
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()
}

// TestIndexBuilder_AllPropertyOptions 测试所有 PropertyOption 函数
func TestIndexBuilder_AllPropertyOptions(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_index_all_options"
	_ = NewIndexBuilder(client, indexName).Delete(ctx)

	// 测试所有可用的 PropertyOption 函数（使用链式 API）
	err := NewIndexBuilder(client, indexName).
		Shards(1).
		Replicas(0).
		// WithAnalyzer
		AddProperty("analyzed_field", "text", WithAnalyzer("standard")).
		// WithIndex - 测试 true 和 false
		AddProperty("indexed_field", "text", WithIndex(true)).
		AddProperty("not_indexed_field", "text", WithIndex(false)).
		// WithStore - 测试 true 和 false
		AddProperty("stored_field", "text", WithStore(true)).
		AddProperty("not_stored_field", "text", WithStore(false)).
		// WithFormat - 测试不同的日期格式
		AddProperty("date_field1", "date", WithFormat("yyyy-MM-dd")).
		AddProperty("date_field2", "date", WithFormat("yyyy-MM-dd HH:mm:ss")).
		AddProperty("date_field3", "date", WithFormat("epoch_millis")).
		// WithSubField - 测试多字段映射（使用链式 API）
		AddProperty("multi_field", "text",
			WithSubField("keyword", "keyword"),
			WithSubField("english", "text", WithAnalyzer("english")),
		).
		// WithIgnoreAbove - 测试 ignore_above 参数
		AddProperty("limited_keyword", "keyword", WithIgnoreAbove(100)).
		Do(ctx)

	if err != nil {
		t.Fatalf("创建索引失败: %v", err)
	}
	t.Logf("✓ 所有 PropertyOption 函数测试成功")

	// 获取并打印索引信息
	info, _ := NewIndexBuilder(client, indexName).Get(ctx)
	t.Logf("完整索引映射:\n%s", info.PrettyJSON())

	// 清理
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()
}
