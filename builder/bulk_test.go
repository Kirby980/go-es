package builder

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"
)

// TestBulkBuilder_IndexOperations 测试批量索引操作
func TestBulkBuilder_IndexOperations(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_bulk_index"
	prepareTestIndex(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	// 批量索引
	resp, err := NewBulkBuilder(client).
		Index(indexName).
		Add("", "bulk1", map[string]interface{}{
			"title": "批量文档1",
			"views": 100,
		}).
		Add("", "bulk2", map[string]interface{}{
			"title": "批量文档2",
			"views": 200,
		}).
		Add("", "bulk3", map[string]interface{}{
			"title": "批量文档3",
			"views": 300,
		}).
		Do(ctx)

	if err != nil {
		t.Fatalf("批量索引失败: %v", err)
	}

	if resp.SuccessCount() != 3 {
		t.Errorf("期望成功 3 个, 实际=%d", resp.SuccessCount())
	}

	t.Logf("✓ 批量索引成功: 成功=%d, 失败=%d, 耗时=%dms",
		resp.SuccessCount(), len(resp.FailedItems()), resp.Took)
}

// TestBulkBuilder_MixedOperations 测试混合操作
func TestBulkBuilder_MixedOperations(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_bulk_mixed"
	prepareTestIndex(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	// 先创建一些文档
	_, _ = NewDocumentBuilder(client, indexName).
		ID("mixed1").
		Set("title", "原始文档1").
		Set("views", 100).
		Do(ctx)

	_, _ = NewDocumentBuilder(client, indexName).
		ID("mixed2").
		Set("title", "原始文档2").
		Set("views", 200).
		Do(ctx)

	time.Sleep(1 * time.Second)

	// 混合操作：索引、更新、删除
	resp, err := NewBulkBuilder(client).
		Index(indexName).
		Add("", "mixed3", map[string]interface{}{
			"title": "新文档3",
			"views": 300,
		}).
		Update("", "mixed1", map[string]interface{}{
			"views": 150,
		}).
		Delete("", "mixed2").
		Do(ctx)

	if err != nil {
		t.Fatalf("混合批量操作失败: %v", err)
	}

	t.Logf("✓ 混合批量操作成功: 成功=%d, 失败=%d",
		resp.SuccessCount(), len(resp.FailedItems()))

	if resp.HasErrors() {
		for _, item := range resp.FailedItems() {
			t.Logf("  失败项: ID=%s, 错误=%s", item.ID, item.Error.Reason)
		}
	}
}

// TestBulkBuilder_CreateOperations 测试批量创建
func TestBulkBuilder_CreateOperations(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_bulk_create"
	prepareTestIndex(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	// 批量创建
	resp, err := NewBulkBuilder(client).
		Index(indexName).
		Create("", "create1", map[string]interface{}{
			"title": "创建文档1",
		}).
		Create("", "create2", map[string]interface{}{
			"title": "创建文档2",
		}).
		Do(ctx)

	if err != nil {
		t.Fatalf("批量创建失败: %v", err)
	}

	t.Logf("✓ 批量创建成功: 成功=%d", resp.SuccessCount())

	// 再次创建相同 ID 应该部分失败
	resp2, err := NewBulkBuilder(client).
		Index(indexName).
		Create("", "create1", map[string]interface{}{
			"title": "重复文档",
		}).
		Create("", "create3", map[string]interface{}{
			"title": "新文档3",
		}).
		Do(ctx)

	if err != nil {
		t.Fatalf("批量创建失败: %v", err)
	}

	if !resp2.HasErrors() {
		t.Error("应该有错误（文档已存在）")
	}

	t.Logf("✓ 重复创建检测: 成功=%d, 失败=%d",
		resp2.SuccessCount(), len(resp2.FailedItems()))
}

// TestBulkBuilder_UpdateOperations 测试批量更新
func TestBulkBuilder_UpdateOperations(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_bulk_update"
	prepareTestIndex(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	// 先创建文档
	_, _ = NewDocumentBuilder(client, indexName).
		ID("update1").
		Set("title", "原文档1").
		Set("views", 100).
		Do(ctx)

	_, _ = NewDocumentBuilder(client, indexName).
		ID("update2").
		Set("title", "原文档2").
		Set("views", 200).
		Do(ctx)

	time.Sleep(1 * time.Second)

	// 批量更新
	resp, err := NewBulkBuilder(client).
		Index(indexName).
		Update("", "update1", map[string]interface{}{
			"views": 150,
		}).
		Update("", "update2", map[string]interface{}{
			"views": 250,
		}).
		Do(ctx)

	if err != nil {
		t.Fatalf("批量更新失败: %v", err)
	}

	t.Logf("✓ 批量更新成功: 成功=%d", resp.SuccessCount())
}

// TestBulkBuilder_DeleteOperations 测试批量删除
func TestBulkBuilder_DeleteOperations(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_bulk_delete"
	prepareTestIndex(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	// 先创建文档
	for i := 1; i <= 3; i++ {
		_, _ = NewDocumentBuilder(client, indexName).
			ID(string(rune('0'+i))).
			Set("title", "待删除文档").
			Do(ctx)
	}

	time.Sleep(1 * time.Second)

	// 批量删除
	resp, err := NewBulkBuilder(client).
		Index(indexName).
		Delete("", "1").
		Delete("", "2").
		Delete("", "3").
		Do(ctx)

	if err != nil {
		t.Fatalf("批量删除失败: %v", err)
	}

	if resp.SuccessCount() != 3 {
		t.Errorf("期望成功删除 3 个, 实际=%d", resp.SuccessCount())
	}

	t.Logf("✓ 批量删除成功: 成功=%d", resp.SuccessCount())
}

// TestBulkBuilder_FromStruct 测试从结构体批量操作
func TestBulkBuilder_FromStruct(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_bulk_struct"
	prepareTestIndex(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	type Doc struct {
		Title string `json:"title"`
		Views int    `json:"views"`
	}

	// 使用结构体批量索引
	resp, err := NewBulkBuilder(client).
		Index(indexName).
		AddFromStruct("", "struct1", Doc{Title: "结构体1", Views: 100}).
		AddFromStruct("", "struct2", Doc{Title: "结构体2", Views: 200}).
		Do(ctx)

	if err != nil {
		t.Fatalf("从结构体批量索引失败: %v", err)
	}

	t.Logf("✓ 从结构体批量索引成功: 成功=%d", resp.SuccessCount())
}

// TestBulkBuilder_Clear 测试清空操作
func TestBulkBuilder_Clear(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()

	builder := NewBulkBuilder(client).
		Index("test").
		Add("", "1", map[string]interface{}{"title": "doc1"}).
		Add("", "2", map[string]interface{}{"title": "doc2"})

	if builder.Count() != 2 {
		t.Errorf("期望 2 个操作, 实际=%d", builder.Count())
	}

	builder.Clear()

	if builder.Count() != 0 {
		t.Errorf("清空后应该为 0, 实际=%d", builder.Count())
	}

	t.Logf("✓ 清空操作测试成功")
}

// TestBulkBuilder_Count 测试操作计数
func TestBulkBuilder_Count(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()

	builder := NewBulkBuilder(client).
		Index("test").
		Add("", "1", map[string]interface{}{"title": "doc1"}).
		Update("", "2", map[string]interface{}{"title": "doc2"}).
		Delete("", "3")

	if builder.Count() != 3 {
		t.Errorf("期望 3 个操作, 实际=%d", builder.Count())
	}

	t.Logf("✓ 操作计数测试成功: %d 个操作", builder.Count())
}

// TestBulkBuilder_Build 测试构建批量请求
func TestBulkBuilder_Build(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()

	builder := NewBulkBuilder(client).
		Index("test").
		Add("", "1", map[string]interface{}{"title": "doc1"}).
		Update("", "2", map[string]interface{}{"title": "doc2"}).
		Delete("", "3")

	body := builder.Build()

	if len(body) == 0 {
		t.Error("构建的请求体不应该为空")
	}

	t.Logf("✓ 构建批量请求成功: %d 字节", len(body))
}

// TestBulkBuilder_ErrorHandling 测试错误处理
func TestBulkBuilder_ErrorHandling(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_bulk_error"
	prepareTestIndex(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	// 创建一些会成功和会失败的操作
	resp, err := NewBulkBuilder(client).
		Index(indexName).
		Add("", "success1", map[string]interface{}{
			"title": "成功文档",
			"views": 100,
		}).
		// 这个可能会失败（如果映射不匹配）
		Add("", "success2", map[string]interface{}{
			"title": "另一个成功文档",
			"views": "invalid", // 尝试用字符串赋值给整数字段
		}).
		Do(ctx)

	if err != nil {
		t.Logf("批量操作完成但有错误: %v", err)
	}

	if resp.HasErrors() {
		t.Logf("✓ 检测到错误:")
		for _, item := range resp.FailedItems() {
			t.Logf("  - ID=%s, Status=%d, Error=%s",
				item.ID, item.Status, item.Error.Reason)
		}
	}

	t.Logf("批量操作结果: 成功=%d, 失败=%d",
		resp.SuccessCount(), len(resp.FailedItems()))
}

// TestBulkBuilder_ChainedAPI 测试链式调用 API
func TestBulkBuilder_ChainedAPI(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_bulk_chained"
	prepareTestIndex(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	// 使用链式调用添加文档
	resp, err := NewBulkBuilder(client).
		Index(indexName).
		AddDoc("chain1").
		Set("title", "链式文档1").
		Set("views", 100).
		AddDoc("chain2").
		Set("title", "链式文档2").
		Set("views", 200).
		AddDoc("chain3").
		Set("title", "链式文档3").
		Set("views", 300).
		Do(ctx)

	if err != nil {
		t.Fatalf("链式调用批量索引失败: %v", err)
	}

	if resp.SuccessCount() != 3 {
		t.Errorf("期望成功 3 个, 实际=%d", resp.SuccessCount())
	}

	t.Logf("✓ 链式调用批量索引成功: 成功=%d, 失败=%d",
		resp.SuccessCount(), len(resp.FailedItems()))
}

// TestBulkBuilder_ChainedMixedOperations 测试链式调用混合操作
func TestBulkBuilder_ChainedMixedOperations(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_bulk_chained_mixed"
	prepareTestIndex(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	// 先创建一些文档
	_, _ = NewDocumentBuilder(client, indexName).
		ID("mix1").
		Set("title", "原始文档1").
		Set("views", 100).
		Do(ctx)

	_, _ = NewDocumentBuilder(client, indexName).
		ID("mix2").
		Set("title", "原始文档2").
		Set("views", 200).
		Do(ctx)

	time.Sleep(1 * time.Second)

	// 使用链式调用进行混合操作
	resp, err := NewBulkBuilder(client).
		Index(indexName).
		AddDoc("mix3").
		Set("title", "新文档3").
		Set("views", 300).
		UpdateDoc("mix1").
		Set("views", 150).
		DeleteDoc("mix2").
		Do(ctx)

	if err != nil {
		t.Fatalf("链式调用混合操作失败: %v", err)
	}

	t.Logf("✓ 链式调用混合操作成功: 成功=%d, 失败=%d",
		resp.SuccessCount(), len(resp.FailedItems()))

	if resp.HasErrors() {
		for _, item := range resp.FailedItems() {
			t.Logf("  失败项: ID=%s, 错误=%s", item.ID, item.Error.Reason)
		}
	}
}

// TestBulkBuilder_ChainedWithStruct 测试链式调用使用结构体
func TestBulkBuilder_ChainedWithStruct(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_bulk_chained_struct"
	prepareTestIndex(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	type Doc struct {
		Title string `json:"title"`
		Views int    `json:"views"`
	}

	// 使用链式调用和 SetFromStruct
	resp, err := NewBulkBuilder(client).
		Index(indexName).
		AddDoc("struct1").
		SetFromStruct(Doc{Title: "结构体1", Views: 100}).
		AddDoc("struct2").
		SetFromStruct(Doc{Title: "结构体2", Views: 200}).
		Do(ctx)

	if err != nil {
		t.Fatalf("链式调用结构体失败: %v", err)
	}

	if resp.SuccessCount() != 2 {
		t.Errorf("期望成功 2 个, 实际=%d", resp.SuccessCount())
	}

	t.Logf("✓ 链式调用结构体成功: 成功=%d", resp.SuccessCount())
}

// TestBulkBuilder_MixedAPIStyles 测试混用两种 API 风格
func TestBulkBuilder_MixedAPIStyles(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_bulk_mixed_styles"
	prepareTestIndex(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	// 混用传统 map 方式和链式调用方式
	resp, err := NewBulkBuilder(client).
		Index(indexName).
		Add("", "map1", map[string]interface{}{
			"title": "Map方式1",
			"views": 100,
		}).
		AddDoc("chain1").
		Set("title", "链式方式1").
		Set("views", 200).
		Add("", "map2", map[string]interface{}{
			"title": "Map方式2",
			"views": 300,
		}).
		AddDoc("chain2").
		Set("title", "链式方式2").
		Set("views", 400).
		Do(ctx)

	if err != nil {
		t.Fatalf("混用 API 风格失败: %v", err)
	}

	if resp.SuccessCount() != 4 {
		t.Errorf("期望成功 4 个, 实际=%d", resp.SuccessCount())
	}

	t.Logf("✓ 混用 API 风格成功: 成功=%d", resp.SuccessCount())
}

// TestBulkBuilder_ChainedWithNestedObject 测试链式调用嵌套对象
func TestBulkBuilder_ChainedWithNestedObject(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_bulk_nested"
	prepareTestIndex(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	// 使用链式调用添加嵌套对象
	resp, err := NewBulkBuilder(client).
		Index(indexName).
		AddDoc("nested1").
		Set("title", "嵌套文档1").
		SetObject("user", func(obj *NestedObject) {
			obj.Set("name", "张三").
				Set("age", 25).
				SetObject("address", func(addr *NestedObject) {
					addr.Set("city", "北京").
						Set("street", "长安街")
				})
		}).
		AddDoc("nested2").
		Set("title", "嵌套文档2").
		SetObject("user", func(obj *NestedObject) {
			obj.Set("name", "李四").
				Set("age", 30).
				SetObject("address", func(addr *NestedObject) {
					addr.Set("city", "上海").
						Set("street", "南京路")
				})
		}).
		Do(ctx)

	if err != nil {
		t.Fatalf("链式调用嵌套对象失败: %v", err)
	}

	if resp.SuccessCount() != 2 {
		t.Errorf("期望成功 2 个, 实际=%d", resp.SuccessCount())
	}

	t.Logf("✓ 链式调用嵌套对象成功: 成功=%d", resp.SuccessCount())

	// 验证嵌套对象
	time.Sleep(1 * time.Second)
	getResp, _ := NewDocumentBuilder(client, indexName).
		ID("nested1").
		Get(ctx)

	user := getResp.Source["user"].(map[string]interface{})
	if user["name"] != "张三" {
		t.Errorf("user.name 应该为 '张三', 实际=%v", user["name"])
	}

	address := user["address"].(map[string]interface{})
	if address["city"] != "北京" {
		t.Errorf("user.address.city 应该为 '北京', 实际=%v", address["city"])
	}

	t.Logf("✓ 嵌套对象验证成功")
}

// TestBulkBuilder_ChainedWithObjectArray 测试链式调用对象数组
func TestBulkBuilder_ChainedWithObjectArray(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_bulk_obj_array"
	prepareTestIndex(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	// 使用链式调用添加对象数组
	resp, err := NewBulkBuilder(client).
		Index(indexName).
		AddDoc("array1").
		Set("title", "对象数组文档").
		SetArray("tags", "Go", "ES", "测试").
		SetObjectArray("comments",
			func(obj *NestedObject) {
				obj.Set("author", "用户1").
					Set("content", "评论1").
					Set("rating", 5)
			},
			func(obj *NestedObject) {
				obj.Set("author", "用户2").
					Set("content", "评论2").
					Set("rating", 4)
			},
		).
		Do(ctx)

	if err != nil {
		t.Fatalf("链式调用对象数组失败: %v", err)
	}

	if resp.SuccessCount() != 1 {
		t.Errorf("期望成功 1 个, 实际=%d", resp.SuccessCount())
	}

	t.Logf("✓ 链式调用对象数组成功: 成功=%d", resp.SuccessCount())

	// 验证对象数组
	time.Sleep(1 * time.Second)
	getResp, _ := NewDocumentBuilder(client, indexName).
		ID("array1").
		Get(ctx)

	comments := getResp.Source["comments"].([]interface{})
	if len(comments) != 2 {
		t.Errorf("应该有 2 条评论, 实际=%d", len(comments))
	}

	tags := getResp.Source["tags"].([]interface{})
	if len(tags) != 3 {
		t.Errorf("tags 应该有 3 个元素, 实际=%d", len(tags))
	}

	t.Logf("✓ 对象数组验证成功")
}

// TestBulkBuilder_ComplexNested 测试批量操作复杂嵌套
func TestBulkBuilder_ComplexNested(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_bulk_complex"
	prepareTestIndex(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	// 批量添加复杂嵌套结构
	resp, err := NewBulkBuilder(client).
		Index(indexName).
		AddDoc("complex1").
		Set("title", "复杂文档1").
		Set("price", 99.99).
		SetArray("tags", "标签1", "标签2").
		SetObject("creator", func(creator *NestedObject) {
			creator.Set("name", "作者1").
				SetObject("profile", func(profile *NestedObject) {
					profile.Set("bio", "简介").
						Set("followers", 1000)
				}).
				SetObjectArray("projects", func(proj *NestedObject) {
					proj.Set("name", "项目A")
				}, func(proj *NestedObject) {
					proj.Set("name", "项目B")
				})
		}).
		AddDoc("complex2").
		Set("title", "复杂文档2").
		Set("price", 199.99).
		SetObject("creator", func(creator *NestedObject) {
			creator.Set("name", "作者2").
				SetObjectArray("projects", func(proj *NestedObject) {
					proj.Set("name", "项目C")
				})
		}).
		Do(ctx)

	if err != nil {
		t.Fatalf("批量复杂嵌套失败: %v", err)
	}

	if resp.SuccessCount() != 2 {
		t.Errorf("期望成功 2 个, 实际=%d", resp.SuccessCount())
	}

	t.Logf("✓ 批量复杂嵌套成功: 成功=%d", resp.SuccessCount())

	// 验证第一个文档
	time.Sleep(1 * time.Second)
	getResp, _ := NewDocumentBuilder(client, indexName).
		ID("complex1").
		Get(ctx)

	creator := getResp.Source["creator"].(map[string]interface{})
	profile := creator["profile"].(map[string]interface{})
	if profile["bio"] != "简介" {
		t.Error("creator.profile.bio 不正确")
	}

	projects := creator["projects"].([]interface{})
	if len(projects) != 2 {
		t.Errorf("projects 应该有 2 个元素, 实际=%d", len(projects))
	}

	t.Logf("✓ 复杂嵌套验证成功")
}

func TestBulkBuilder_Flush(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_bulk_flush"
	prepareTestIndex(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	// 测试参数
	totalDocs := 255
	batchSize := 50
	expectedBatches := totalDocs / batchSize // 255 / 50 = 5 批（自动提交）

	// 统计信息
	var flushCount int
	var autoFlushSuccessCount int

	// 创建 Bulk Builder 并设置自动提交
	bulk := NewBulkBuilder(client).
		Index(indexName).
		AutoFlushSize(batchSize).
		OnFlush(func(resp *BulkResponse) {
			flushCount++
			autoFlushSuccessCount += resp.SuccessCount()
			t.Logf("✓ 自动提交批次 #%d: 成功 %d 条，累计 %d 条",
				flushCount, resp.SuccessCount(), autoFlushSuccessCount)

			// 验证每批次没有错误
			if resp.HasErrors() {
				t.Errorf("批次 #%d 有错误", flushCount)
				for _, item := range resp.FailedItems() {
					t.Errorf("  失败项: ID=%s, 错误=%s", item.ID, item.Error.Reason)
				}
			}
		})
	// 添加文档（会触发自动提交）
	for i := 0; i < totalDocs; i++ {
		bulk.AddDoc(strconv.Itoa(i)).Set("price", i).Set("name", fmt.Sprintf("product_%d", i))
	}

	t.Logf("添加了 %d 个文档，触发了 %d 次自动提交", totalDocs, flushCount)

	// 手动提交剩余的文档
	remainingCount := bulk.Count()
	t.Logf("剩余 %d 个文档待提交", remainingCount)

	finalResp, err := bulk.Do(ctx)
	if err != nil {
		t.Fatalf("最终提交失败: %v", err)
	}

	finalSuccessCount := finalResp.SuccessCount()
	t.Logf("✓ 最终提交: 成功 %d 条", finalSuccessCount)

	// 验证结果
	totalSuccess := autoFlushSuccessCount + finalSuccessCount

	t.Logf("\n=== 统计结果 ===")
	t.Logf("总文档数: %d", totalDocs)
	t.Logf("批次大小: %d", batchSize)
	t.Logf("自动提交次数: %d (预期 %d)", flushCount, expectedBatches)
	t.Logf("自动提交成功: %d", autoFlushSuccessCount)
	t.Logf("最终提交成功: %d", finalSuccessCount)
	t.Logf("总成功数: %d", totalSuccess)

	// 断言：自动提交次数应该等于预期批次数
	if flushCount != expectedBatches {
		t.Errorf("自动提交次数错误: 期望 %d 次, 实际 %d 次", expectedBatches, flushCount)
	}

	// 断言：自动提交的数量应该是批次大小的整数倍
	expectedAutoFlush := flushCount * batchSize
	if autoFlushSuccessCount != expectedAutoFlush {
		t.Errorf("自动提交数量错误: 期望 %d, 实际 %d", expectedAutoFlush, autoFlushSuccessCount)
	}

	// 断言：剩余文档数应该是总数对批次大小的余数
	expectedRemaining := totalDocs % batchSize
	if remainingCount != expectedRemaining {
		t.Errorf("剩余文档数错误: 期望 %d, 实际 %d", expectedRemaining, remainingCount)
	}

	// 断言：总成功数应该等于总文档数
	if totalSuccess != totalDocs {
		t.Errorf("总成功数错误: 期望 %d, 实际 %d", totalDocs, totalSuccess)
	}

	// 验证最终提交的数量
	if finalSuccessCount != expectedRemaining {
		t.Errorf("最终提交数量错误: 期望 %d, 实际 %d", expectedRemaining, finalSuccessCount)
	}

	// 检查是否有错误
	if finalResp.HasErrors() {
		t.Error("最终提交有错误:")
		for _, item := range finalResp.FailedItems() {
			t.Errorf("  失败项: ID=%s, 错误=%s", item.ID, item.Error.Reason)
		}
	}

	t.Logf("✓ 自动分批提交测试通过")
}
