package builder

import (
	"context"
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
			ID(string(rune('0' + i))).
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
