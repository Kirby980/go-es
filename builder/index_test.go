package builder

import (
	"context"
	"testing"
	"time"

	"github.com/Kirby980/go-es/client"
	"github.com/Kirby980/go-es/config"
)

// 创建测试客户端
func createTestClient(t *testing.T) *client.Client {
	esClient, err := client.New(
		config.WithAddresses("https://localhost:9200"),
		config.WithAuth("elastic", "123456"),
		config.WithTransport(true),
		config.WithTimeout(10*time.Second),
		config.WithMaxConnsPerHost(100),
		config.WithMaxIdConns(200),
		config.WithMaxIdleConnsPerHost(50),
		config.WithIdleConnTimeout(90*time.Second),
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
		AddProperty("title", "text", WithAnalyzer("ik_smart")).
		AddProperty("content", "text", WithAnalyzer("ik_smart")).
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
		AddProperty("description", "text", WithAnalyzer("ik_smart")).
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
			WithAnalyzer("ik_smart"),
			WithSubField("keyword", "keyword", WithIgnoreAbove(256)),
			WithSubField("raw", "keyword"),
		).
		// 使用 WithSubProperties 添加嵌套对象（必须是 object 或 nested 类型）
		AddProperty("author", "object",
			WithSubProperties("name", "text"),
			WithSubProperties("email", "keyword"),
			WithSubProperties("profile", "object",
				WithSubProperties("age", "integer"),
				WithSubProperties("city", "keyword"),
			),
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
	// defer func() {
	// 	_ = NewIndexBuilder(client, indexName).Delete(ctx)
	// }()
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
			WithAnalyzer("ik_smart"),
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
			WithAnalyzer("ik_smart"),
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
		AddProperty("analyzed_field", "text", WithAnalyzer("ik_smart")).
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

// TestIndexBuilder_UpdateSettings 测试更新索引设置
func TestIndexBuilder_UpdateSettings(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_index_update_settings"
	_ = NewIndexBuilder(client, indexName).Delete(ctx)

	// 1. 先创建索引（初始设置：1个副本，1秒刷新间隔）
	err := NewIndexBuilder(client, indexName).
		Shards(1).
		Replicas(1).
		RefreshInterval("1s").
		AddProperty("name", "text").
		AddProperty("price", "float").
		Create(ctx)

	if err != nil {
		t.Fatalf("创建索引失败: %v", err)
	}
	t.Logf("✓ 创建索引成功，初始设置：副本数=1，刷新间隔=1s")

	// 2. 获取初始索引信息
	info, _ := NewIndexBuilder(client, indexName).Get(ctx)
	t.Logf("初始索引设置:\n%s", info.PrettyJSON())

	// 3. 更新索引设置（修改副本数为2，刷新间隔为30s）
	err = NewIndexBuilder(client, indexName).
		Debug().
		Replicas(2).
		RefreshInterval("30s").
		UpdateSettings(ctx)

	if err != nil {
		t.Fatalf("更新索引设置失败: %v", err)
	}
	t.Logf("✓ 更新索引设置成功，新设置：副本数=2，刷新间隔=30s")

	// 4. 再次获取索引信息，验证设置已更新
	updatedInfo, err := NewIndexBuilder(client, indexName).Get(ctx)
	if err != nil {
		t.Fatalf("获取更新后的索引信息失败: %v", err)
	}
	t.Logf("更新后的索引设置:\n%s", updatedInfo.PrettyJSON())

	// 5. 测试只更新刷新间隔
	err = NewIndexBuilder(client, indexName).
		RefreshInterval("5s").
		UpdateSettings(ctx)

	if err != nil {
		t.Fatalf("更新刷新间隔失败: %v", err)
	}
	t.Logf("✓ 单独更新刷新间隔成功")

	// 清理
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()
}

// TestIndexBuilder_PutMapping 测试更新索引映射
func TestIndexBuilder_PutMapping(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_index_put_mapping"
	_ = NewIndexBuilder(client, indexName).Delete(ctx)

	// 1. 先创建索引（只有基础字段）
	err := NewIndexBuilder(client, indexName).
		Shards(1).
		Replicas(0).
		AddProperty("name", "text", WithAnalyzer("ik_smart")).
		AddProperty("price", "float").
		Create(ctx)

	if err != nil {
		t.Fatalf("创建索引失败: %v", err)
	}
	t.Logf("✓ 创建索引成功，初始字段：name, price")

	// 2. 获取初始索引映射
	info, _ := NewIndexBuilder(client, indexName).Get(ctx)
	t.Logf("初始索引映射:\n%s", info.PrettyJSON())

	// 3. 添加新字段（使用 PutMapping）
	err = NewIndexBuilder(client, indexName).
		Debug().
		AddProperty("description", "text", WithAnalyzer("ik_max_word")).
		AddProperty("stock", "integer").
		AddProperty("category", "keyword").
		AddProperty("created_at", "date", WithFormat("yyyy-MM-dd HH:mm:ss")).
		PutMapping(ctx)

	if err != nil {
		t.Fatalf("更新索引映射失败: %v", err)
	}
	t.Logf("✓ 更新索引映射成功，添加字段：description, stock, category, created_at")

	// 4. 再次获取索引信息，验证新字段已添加
	updatedInfo, err := NewIndexBuilder(client, indexName).Get(ctx)
	if err != nil {
		t.Fatalf("获取更新后的索引信息失败: %v", err)
	}
	t.Logf("更新后的索引映射:\n%s", updatedInfo.PrettyJSON())

	// 5. 测试添加嵌套字段
	err = NewIndexBuilder(client, indexName).
		AddProperty("tags", "keyword").
		AddProperty("author", "object",
			WithSubProperties("name", "text"),
			WithSubProperties("email", "keyword"),
		).
		PutMapping(ctx)

	if err != nil {
		t.Fatalf("添加嵌套字段失败: %v", err)
	}
	t.Logf("✓ 添加嵌套字段成功")

	// 6. 最终验证
	finalInfo, _ := NewIndexBuilder(client, indexName).Get(ctx)
	t.Logf("最终索引映射:\n%s", finalInfo.PrettyJSON())

	// 清理
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()
}

// TestIndexBuilder_UpdateSettingsAndPutMapping 测试同时使用 UpdateSettings 和 PutMapping
func TestIndexBuilder_UpdateSettingsAndPutMapping(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_index_update_both"
	_ = NewIndexBuilder(client, indexName).Delete(ctx)

	// 1. 创建初始索引
	err := NewIndexBuilder(client, indexName).
		Shards(1).
		Replicas(0).
		RefreshInterval("1s").
		AddProperty("id", "keyword").
		AddProperty("title", "text").
		Create(ctx)

	if err != nil {
		t.Fatalf("创建索引失败: %v", err)
	}
	t.Logf("✓ 步骤1：创建初始索引成功")

	// 2. 更新设置
	err = NewIndexBuilder(client, indexName).
		Replicas(1).
		RefreshInterval("10s").
		UpdateSettings(ctx)

	if err != nil {
		t.Fatalf("更新设置失败: %v", err)
	}
	t.Logf("✓ 步骤2：更新索引设置成功")

	// 3. 添加新字段
	err = NewIndexBuilder(client, indexName).
		AddProperty("content", "text", WithAnalyzer("ik_smart")).
		AddProperty("status", "keyword").
		AddProperty("views", "long").
		PutMapping(ctx)

	if err != nil {
		t.Fatalf("添加新字段失败: %v", err)
	}
	t.Logf("✓ 步骤3：添加新字段成功")

	// 4. 验证最终结果
	finalInfo, err := NewIndexBuilder(client, indexName).Get(ctx)
	if err != nil {
		t.Fatalf("获取最终索引信息失败: %v", err)
	}
	t.Logf("✓ 步骤4：验证完成")
	t.Logf("最终索引完整信息:\n%s", finalInfo.PrettyJSON())

	// 验证索引包含预期的字段和设置
	if finalInfo.Mappings == nil {
		t.Error("索引应该包含 mappings")
	}
	if finalInfo.Settings == nil {
		t.Error("索引应该包含 settings")
	}

	// 清理
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()
}

// TestIndexBuilder_CreateMethodAlias 测试 Create 和 Do 方法的兼容性
func TestIndexBuilder_CreateMethodAlias(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	// 1. 使用新的 Create() 方法
	indexName1 := "test_index_create_method"
	_ = NewIndexBuilder(client, indexName1).Delete(ctx)

	err := NewIndexBuilder(client, indexName1).
		Shards(1).
		Replicas(0).
		AddProperty("field1", "text").
		Create(ctx)

	if err != nil {
		t.Fatalf("使用 Create() 创建索引失败: %v", err)
	}
	t.Logf("✓ 使用 Create() 方法创建索引成功")

	// 2. 使用兼容的 Do() 方法（应该调用 Create）
	indexName2 := "test_index_do_method"
	_ = NewIndexBuilder(client, indexName2).Delete(ctx)

	err = NewIndexBuilder(client, indexName2).
		Shards(1).
		Replicas(0).
		AddProperty("field1", "text").
		Do(ctx)

	if err != nil {
		t.Fatalf("使用 Do() 创建索引失败: %v", err)
	}
	t.Logf("✓ 使用 Do() 方法创建索引成功（向后兼容）")

	// 验证两个索引都存在
	exists1, _ := NewIndexBuilder(client, indexName1).Exists(ctx)
	exists2, _ := NewIndexBuilder(client, indexName2).Exists(ctx)

	if !exists1 {
		t.Error("使用 Create() 创建的索引应该存在")
	}
	if !exists2 {
		t.Error("使用 Do() 创建的索引应该存在")
	}

	t.Logf("✓ Create() 和 Do() 方法都能正常工作，保持向后兼容")

	// 清理
	defer func() {
		_ = NewIndexBuilder(client, indexName1).Delete(ctx)
		_ = NewIndexBuilder(client, indexName2).Delete(ctx)
	}()
}

// TestIndexBuilder_AddCustomAnalyzer 测试添加自定义分析器（简化版）
func TestIndexBuilder_AddCustomAnalyzer(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_index_custom_analyzer"
	_ = NewIndexBuilder(client, indexName).Delete(ctx)

	// 创建带自定义分析器的索引（使用常量，避免拼写错误）
	err := NewIndexBuilder(client, indexName).
		Shards(1).
		Replicas(0).
		AddCustomAnalyzer("ik_case_sensitive", TokenizerIKSmart). // 使用常量
		AddProperty("title", "text", WithAnalyzer("ik_case_sensitive")).
		AddProperty("content", "text", WithAnalyzer(AnalyzerIKSmart)). // 使用内置分析器常量
		Debug().
		Create(ctx)

	if err != nil {
		t.Fatalf("创建带自定义分析器的索引失败: %v", err)
	}
	t.Logf("✓ 创建带自定义分析器的索引成功")

	// 获取索引信息验证
	info, err := NewIndexBuilder(client, indexName).Get(ctx)
	if err != nil {
		t.Fatalf("获取索引信息失败: %v", err)
	}
	t.Logf("索引配置:\n%s", info.PrettyJSON())

	// 清理
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()
}

// TestIndexBuilder_AddAnalyzer 测试添加分析器（完整版）
func TestIndexBuilder_AddAnalyzer(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_index_analyzer_full"
	_ = NewIndexBuilder(client, indexName).Delete(ctx)

	// 创建带自定义分析器的索引（完整版，使用 Option 模式）
	err := NewIndexBuilder(client, indexName).
		Shards(1).
		Replicas(0).
		// 添加自定义分析器 1：只有 tokenizer（最简单的自定义）
		AddAnalyzer("my_analyzer",
			WithAnalyzerType("custom"),
			WithTokenizer("ik_smart"),
		).
		// 添加自定义分析器 2：使用不同的 tokenizer
		AddAnalyzer("simple_ik",
			WithAnalyzerType("custom"),
			WithTokenizer("ik_max_word"),
		).
		AddProperty("html_content", "text", WithAnalyzer("my_analyzer")).
		AddProperty("description", "text", WithAnalyzer("simple_ik")).
		Debug().
		Create(ctx)

	if err != nil {
		t.Fatalf("创建带自定义分析器的索引失败: %v", err)
	}
	t.Logf("✓ 创建带完整自定义分析器的索引成功")

	// 获取索引信息验证
	info, err := NewIndexBuilder(client, indexName).Get(ctx)
	if err != nil {
		t.Fatalf("获取索引信息失败: %v", err)
	}
	t.Logf("索引配置:\n%s", info.PrettyJSON())

	// 清理
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()
}

// TestIndexBuilder_AddTokenizer 测试添加自定义分词器
func TestIndexBuilder_AddTokenizer(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_index_custom_tokenizer"
	_ = NewIndexBuilder(client, indexName).Delete(ctx)

	// 像用户的例子那样：先自定义 tokenizer，再在 analyzer 中使用
	err := NewIndexBuilder(client, indexName).
		Shards(1).
		Replicas(0).
		// 1. 先自定义一个 tokenizer（禁用小写转换）
		AddTokenizer("ik_smart_case_sensitive",
			WithTokenizerType(TokenizerIKSmart),
			WithEnableLowercase(false),
		).
		// 2. 创建 analyzer 使用自定义的 tokenizer
		AddAnalyzer("ik_case_sensitive",
			WithAnalyzerType(AnalyzerTypeCustom),
			WithTokenizer("ik_smart_case_sensitive"), // 使用自定义的 tokenizer
			WithTokenFilters(),                       // 空的 filter
		).
		// 3. 在字段中使用这个 analyzer
		AddProperty("title", "text", WithAnalyzer("ik_case_sensitive")).
		AddProperty("content", "text", WithAnalyzer("ik_case_sensitive")).
		Debug().
		Create(ctx)

	if err != nil {
		t.Fatalf("创建带自定义 tokenizer 的索引失败: %v", err)
	}
	t.Logf("✓ 创建带自定义 tokenizer 的索引成功")

	// 获取索引信息验证
	info, err := NewIndexBuilder(client, indexName).Get(ctx)
	if err != nil {
		t.Fatalf("获取索引信息失败: %v", err)
	}
	t.Logf("索引配置:\n%s", info.PrettyJSON())

	// 清理
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()
}

// TestIndexBuilder_FieldTypeConstants 测试使用字段类型常量
func TestIndexBuilder_FieldTypeConstants(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_field_type_constants"
	_ = NewIndexBuilder(client, indexName).Delete(ctx)

	// 使用常量创建索引，避免拼写错误
	err := NewIndexBuilder(client, indexName).
		Shards(1).
		Replicas(0).
		// 字符串类型
		AddProperty("title", FieldTypeText, WithAnalyzer(AnalyzerIKSmart)).
		AddProperty("sku", FieldTypeKeyword).
		// 数值类型
		AddProperty("price", FieldTypeFloat).
		AddProperty("quantity", FieldTypeInt).
		AddProperty("views", FieldTypeLong).
		// 布尔类型
		AddProperty("available", FieldTypeBoolean).
		// 日期类型
		AddProperty("created_at", FieldTypeDate, WithFormat("yyyy-MM-dd HH:mm:ss")).
		// 地理位置
		AddProperty("location", FieldTypeGeoPoint).
		// 对象类型
		AddProperty("author", FieldTypeObject,
			WithSubProperties("name", FieldTypeText),
			WithSubProperties("email", FieldTypeKeyword),
		).
		Debug().
		Create(ctx)

	if err != nil {
		t.Fatalf("创建索引失败: %v", err)
	}
	t.Logf("✓ 使用字段类型常量创建索引成功")

	// 获取索引信息验证
	info, err := NewIndexBuilder(client, indexName).Get(ctx)
	if err != nil {
		t.Fatalf("获取索引信息失败: %v", err)
	}
	t.Logf("索引映射:\n%s", info.PrettyJSON())

	// 清理
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()
}
