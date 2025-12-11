package builder

import (
	"context"
	"testing"
	"time"

	"go-es/client"
)

// 准备测试索引
func prepareTestIndex(t *testing.T, esClient *client.Client, indexName string) {
	ctx := context.Background()

	// 删除可能存在的索引
	_ = NewIndexBuilder(esClient, indexName).Delete(ctx)

	// 创建测试索引
	err := NewIndexBuilder(esClient, indexName).
		Shards(1).
		Replicas(0).
		AddProperty("title", "text").
		AddProperty("content", "text").
		AddProperty("author", "keyword").
		AddProperty("views", "integer").
		AddProperty("price", "float").
		AddProperty("published", "boolean").
		AddProperty("tags", "keyword").
		AddProperty("created_at", "date").
		Do(ctx)

	if err != nil {
		t.Logf("创建测试索引: %v (可能已存在)", err)
	}

	// 等待索引就绪
	time.Sleep(500 * time.Millisecond)
}

// TestDocumentBuilder_IndexDocument 测试索引文档
func TestDocumentBuilder_IndexDocument(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_doc_index"
	prepareTestIndex(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	// 使用 Do 方法索引文档（带 ID）
	resp, err := NewDocumentBuilder(client, indexName).
		ID("1").
		Set("title", "测试文档").
		Set("content", "这是一篇测试文档的内容").
		Set("author", "张三").
		Set("views", 100).
		Set("price", 99.99).
		Set("published", true).
		Set("tags", []string{"测试", "文档"}).
		Do(ctx)

	if err != nil {
		t.Fatalf("索引文档失败: %v", err)
	}

	if resp.ID != "1" {
		t.Errorf("期望 ID=1, 实际=%s", resp.ID)
	}
	if resp.Result != "created" && resp.Result != "updated" {
		t.Errorf("期望 Result=created/updated, 实际=%s", resp.Result)
	}

	t.Logf("✓ 索引文档成功: ID=%s, Result=%s, Version=%d", resp.ID, resp.Result, resp.Version)
}

// TestDocumentBuilder_IndexWithoutID 测试无 ID 索引文档
func TestDocumentBuilder_IndexWithoutID(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_doc_no_id"
	prepareTestIndex(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	// 不指定 ID，让 ES 自动生成
	resp, err := NewDocumentBuilder(client, indexName).
		Set("title", "自动生成 ID 的文档").
		Set("content", "ES 会自动分配一个唯一 ID").
		Do(ctx)

	if err != nil {
		t.Fatalf("索引文档失败: %v", err)
	}

	if resp.ID == "" {
		t.Error("ID 不应该为空")
	}

	t.Logf("✓ 自动生成 ID 索引成功: ID=%s", resp.ID)
}

// TestDocumentBuilder_Create 测试创建文档
func TestDocumentBuilder_Create(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_doc_create"
	prepareTestIndex(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	// 第一次创建应该成功
	resp, err := NewDocumentBuilder(client, indexName).
		ID("create-1").
		Set("title", "创建文档").
		Create(ctx)

	if err != nil {
		t.Fatalf("创建文档失败: %v", err)
	}
	if resp.Result != "created" {
		t.Errorf("期望 Result=created, 实际=%s", resp.Result)
	}

	t.Logf("✓ 创建文档成功: Result=%s", resp.Result)

	// 第二次创建同一 ID 应该失败
	_, err = NewDocumentBuilder(client, indexName).
		ID("create-1").
		Set("title", "重复文档").
		Create(ctx)

	if err == nil {
		t.Error("重复创建应该失败")
	} else {
		t.Logf("✓ 重复创建正确失败: %v", err)
	}
}

// TestDocumentBuilder_Get 测试获取文档
func TestDocumentBuilder_Get(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_doc_get"
	prepareTestIndex(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	// 先创建一个文档
	_, _ = NewDocumentBuilder(client, indexName).
		ID("get-1").
		Set("title", "待获取的文档").
		Set("author", "李四").
		Set("views", 200).
		Do(ctx)

	time.Sleep(1 * time.Second) // 等待索引刷新

	// 获取文档
	getResp, err := NewDocumentBuilder(client, indexName).
		ID("get-1").
		Get(ctx)

	if err != nil {
		t.Fatalf("获取文档失败: %v", err)
	}

	if !getResp.Found {
		t.Error("文档应该被找到")
	}
	if getResp.ID != "get-1" {
		t.Errorf("期望 ID=get-1, 实际=%s", getResp.ID)
	}
	if getResp.Source["title"] != "待获取的文档" {
		t.Errorf("标题不匹配: %v", getResp.Source["title"])
	}

	t.Logf("✓ 获取文档成功: %v", getResp.Source)
}

// TestDocumentBuilder_GetNotFound 测试获取不存在的文档
func TestDocumentBuilder_GetNotFound(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_doc_get_notfound"
	prepareTestIndex(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	// 获取不存在的文档
	getResp, err := NewDocumentBuilder(client, indexName).
		ID("non-existent").
		Get(ctx)

	if err != nil {
		t.Fatalf("获取文档请求失败: %v", err)
	}

	if getResp.Found {
		t.Error("不存在的文档 Found 应该为 false")
	}

	t.Logf("✓ 正确识别文档不存在")
}

// TestDocumentBuilder_Update 测试更新文档
func TestDocumentBuilder_Update(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_doc_update"
	prepareTestIndex(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	// 先创建文档
	_, _ = NewDocumentBuilder(client, indexName).
		ID("update-1").
		Set("title", "原始标题").
		Set("views", 100).
		Set("price", 50.0).
		Do(ctx)

	time.Sleep(1 * time.Second)

	// 更新文档
	updateResp, err := NewDocumentBuilder(client, indexName).
		ID("update-1").
		Set("views", 150).
		Set("price", 45.0).
		Update(ctx)

	if err != nil {
		t.Fatalf("更新文档失败: %v", err)
	}

	if updateResp.Result != "updated" && updateResp.Result != "noop" {
		t.Errorf("期望 Result=updated/noop, 实际=%s", updateResp.Result)
	}
	if updateResp.Version < 2 {
		t.Errorf("更新后版本应该 >= 2, 实际=%d", updateResp.Version)
	}

	t.Logf("✓ 更新文档成功: Version=%d", updateResp.Version)

	// 验证更新结果
	getResp, _ := NewDocumentBuilder(client, indexName).
		ID("update-1").
		Get(ctx)

	if getResp.Source["views"].(float64) != 150 {
		t.Errorf("views 应该为 150, 实际=%v", getResp.Source["views"])
	}
}

// TestDocumentBuilder_ScriptUpdate 测试脚本更新
func TestDocumentBuilder_ScriptUpdate(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_doc_script"
	prepareTestIndex(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	// 创建文档
	_, _ = NewDocumentBuilder(client, indexName).
		ID("script-1").
		Set("title", "脚本测试").
		Set("views", 100).
		Do(ctx)

	time.Sleep(1 * time.Second)

	// 使用脚本更新
	scriptResp, err := NewDocumentBuilder(client, indexName).
		ID("script-1").
		Script("ctx._source.views += params.increment",
			map[string]interface{}{"increment": 50}).
		Update(ctx)

	if err != nil {
		t.Fatalf("脚本更新失败: %v", err)
	}

	t.Logf("✓ 脚本更新成功: Version=%d", scriptResp.Version)

	// 验证结果
	time.Sleep(1 * time.Second)
	getResp, _ := NewDocumentBuilder(client, indexName).
		ID("script-1").
		Get(ctx)

	if getResp.Source["views"].(float64) != 150 {
		t.Errorf("views 应该为 150, 实际=%v", getResp.Source["views"])
	}
}

// TestDocumentBuilder_Upsert 测试 Upsert
func TestDocumentBuilder_Upsert(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_doc_upsert"
	prepareTestIndex(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	// 第一次 Upsert（文档不存在，应该创建）
	resp1, err := NewDocumentBuilder(client, indexName).
		ID("upsert-1").
		Set("title", "Upsert 文档").
		Set("views", 100).
		Upsert(ctx)

	if err != nil {
		t.Fatalf("Upsert 失败: %v", err)
	}

	t.Logf("✓ Upsert 创建文档: Result=%s", resp1.Result)

	time.Sleep(1 * time.Second)

	// 第二次 Upsert（文档存在，应该更新）
	resp2, err := NewDocumentBuilder(client, indexName).
		ID("upsert-1").
		Set("views", 200).
		Upsert(ctx)

	if err != nil {
		t.Fatalf("Upsert 更新失败: %v", err)
	}

	t.Logf("✓ Upsert 更新文档: Result=%s, Version=%d", resp2.Result, resp2.Version)

	// 验证结果
	getResp, _ := NewDocumentBuilder(client, indexName).
		ID("upsert-1").
		Get(ctx)

	if getResp.Source["views"].(float64) != 200 {
		t.Errorf("views 应该为 200, 实际=%v", getResp.Source["views"])
	}
}

// TestDocumentBuilder_Delete 测试删除文档
func TestDocumentBuilder_Delete(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_doc_delete"
	prepareTestIndex(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	// 创建文档
	_, _ = NewDocumentBuilder(client, indexName).
		ID("delete-1").
		Set("title", "待删除文档").
		Do(ctx)

	time.Sleep(1 * time.Second)

	// 删除文档
	delResp, err := NewDocumentBuilder(client, indexName).
		ID("delete-1").
		Delete(ctx)

	if err != nil {
		t.Fatalf("删除文档失败: %v", err)
	}

	if delResp.Result != "deleted" {
		t.Errorf("期望 Result=deleted, 实际=%s", delResp.Result)
	}

	t.Logf("✓ 删除文档成功: Result=%s", delResp.Result)

	// 验证文档已删除
	time.Sleep(1 * time.Second)
	getResp, _ := NewDocumentBuilder(client, indexName).
		ID("delete-1").
		Get(ctx)

	if getResp.Found {
		t.Error("文档应该已被删除")
	}
}

// TestDocumentBuilder_Exists 测试文档存在性检查
func TestDocumentBuilder_Exists(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_doc_exists"
	prepareTestIndex(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	// 检查不存在的文档
	exists, err := NewDocumentBuilder(client, indexName).
		ID("non-existent").
		Exists(ctx)

	if err != nil {
		t.Fatalf("检查文档存在性失败: %v", err)
	}
	if exists {
		t.Error("文档不应该存在")
	}

	t.Logf("✓ 确认文档不存在")

	// 创建文档
	_, _ = NewDocumentBuilder(client, indexName).
		ID("exists-1").
		Set("title", "存在性测试").
		Do(ctx)

	time.Sleep(1 * time.Second)

	// 检查存在的文档
	exists, err = NewDocumentBuilder(client, indexName).
		ID("exists-1").
		Exists(ctx)

	if err != nil {
		t.Fatalf("检查文档存在性失败: %v", err)
	}
	if !exists {
		t.Error("文档应该存在")
	}

	t.Logf("✓ 确认文档存在")
}

// TestDocumentBuilder_SetMap 测试批量设置字段
func TestDocumentBuilder_SetMap(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_doc_setmap"
	prepareTestIndex(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	// 使用 SetMap 批量设置
	data := map[string]interface{}{
		"title":   "批量设置测试",
		"author":  "王五",
		"views":   300,
		"price":   29.99,
		"published": true,
	}

	resp, err := NewDocumentBuilder(client, indexName).
		ID("setmap-1").
		SetMap(data).
		Do(ctx)

	if err != nil {
		t.Fatalf("SetMap 失败: %v", err)
	}

	t.Logf("✓ SetMap 索引文档成功: ID=%s", resp.ID)

	// 验证
	time.Sleep(1 * time.Second)
	getResp, _ := NewDocumentBuilder(client, indexName).
		ID("setmap-1").
		Get(ctx)

	if getResp.Source["title"] != "批量设置测试" {
		t.Error("SetMap 设置的字段不正确")
	}
}

// TestDocumentBuilder_SetStruct 测试从结构体设置
func TestDocumentBuilder_SetStruct(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_doc_setstruct"
	prepareTestIndex(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	// 定义结构体
	type Article struct {
		Title   string   `json:"title"`
		Author  string   `json:"author"`
		Views   int      `json:"views"`
		Tags    []string `json:"tags"`
	}

	article := Article{
		Title:  "结构体测试",
		Author: "赵六",
		Views:  400,
		Tags:   []string{"Go", "Elasticsearch"},
	}

	resp, err := NewDocumentBuilder(client, indexName).
		ID("struct-1").
		SetStruct(article).
		Do(ctx)

	if err != nil {
		t.Fatalf("SetStruct 失败: %v", err)
	}

	t.Logf("✓ SetStruct 索引文档成功: ID=%s", resp.ID)

	// 验证
	time.Sleep(1 * time.Second)
	getResp, _ := NewDocumentBuilder(client, indexName).
		ID("struct-1").
		Get(ctx)

	if getResp.Source["title"] != "结构体测试" {
		t.Error("SetStruct 设置的字段不正确")
	}
}

// TestDocumentBuilder_MGet 测试批量获取
func TestDocumentBuilder_MGet(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_doc_mget"
	prepareTestIndex(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	// 创建多个文档
	for i := 1; i <= 5; i++ {
		_, _ = NewDocumentBuilder(client, indexName).
			ID(string(rune('0' + i))).
			Set("title", "文档"+string(rune('0'+i))).
			Do(ctx)
	}

	time.Sleep(1 * time.Second)

	// 批量获取
	mgetResp, err := NewMGetBuilder(client, indexName).
		IDs("1", "2", "3").
		Do(ctx)

	if err != nil {
		t.Fatalf("MGet 失败: %v", err)
	}

	if len(mgetResp.Docs) != 3 {
		t.Errorf("期望获取 3 个文档, 实际=%d", len(mgetResp.Docs))
	}

	t.Logf("✓ 批量获取成功: 获取了 %d 个文档", len(mgetResp.Docs))

	for _, doc := range mgetResp.Docs {
		t.Logf("  - ID=%s, Found=%v", doc.ID, doc.Found)
	}
}

// TestDocumentBuilder_Refresh 测试 Refresh 参数
func TestDocumentBuilder_Refresh(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	indexName := "test_doc_refresh"
	prepareTestIndex(t, client, indexName)
	defer func() {
		_ = NewIndexBuilder(client, indexName).Delete(ctx)
	}()

	// 测试 1: 使用 refresh=true 立即刷新
	t.Run("refresh=true", func(t *testing.T) {
		resp, err := NewDocumentBuilder(client, indexName).
			ID("refresh-test-1").
			Set("title", "立即刷新测试").
			Set("content", "使用 refresh=true").
			Refresh("true").
			Do(ctx)

		if err != nil {
			t.Fatalf("索引文档失败: %v", err)
		}
		t.Logf("✓ 文档索引成功 (refresh=true): ID=%s, Result=%s", resp.ID, resp.Result)

		// 立即搜索应该能找到（因为使用了 refresh=true）
		searchResp, err := NewSearchBuilder(client, indexName).
			Match("title", "立即刷新测试").
			Do(ctx)

		if err != nil {
			t.Fatalf("搜索失败: %v", err)
		}

		if searchResp.Hits.Total.Value < 1 {
			t.Errorf("使用 refresh=true 后立即搜索应该能找到文档")
		} else {
			t.Logf("✓ 使用 refresh=true 后立即找到了文档")
		}
	})

	// 测试 2: 使用 refresh=wait_for 等待刷新
	t.Run("refresh=wait_for", func(t *testing.T) {
		resp, err := NewDocumentBuilder(client, indexName).
			ID("refresh-test-2").
			Set("title", "等待刷新测试").
			Set("content", "使用 refresh=wait_for").
			Refresh("wait_for").
			Do(ctx)

		if err != nil {
			t.Fatalf("索引文档失败: %v", err)
		}
		t.Logf("✓ 文档索引成功 (refresh=wait_for): ID=%s, Result=%s", resp.ID, resp.Result)

		// 等待刷新完成后搜索应该能找到
		searchResp, err := NewSearchBuilder(client, indexName).
			Match("title", "等待刷新测试").
			Do(ctx)

		if err != nil {
			t.Fatalf("搜索失败: %v", err)
		}

		if searchResp.Hits.Total.Value < 1 {
			t.Errorf("使用 refresh=wait_for 后搜索应该能找到文档")
		} else {
			t.Logf("✓ 使用 refresh=wait_for 后找到了文档")
		}
	})

	// 测试 3: 不使用 refresh（默认行为）
	t.Run("no refresh", func(t *testing.T) {
		resp, err := NewDocumentBuilder(client, indexName).
			ID("refresh-test-3").
			Set("title", "默认刷新测试").
			Set("content", "不使用 refresh 参数").
			Do(ctx)

		if err != nil {
			t.Fatalf("索引文档失败: %v", err)
		}
		t.Logf("✓ 文档索引成功 (no refresh): ID=%s, Result=%s", resp.ID, resp.Result)

		// 立即搜索可能找不到（需要等待自动刷新）
		searchResp, _ := NewSearchBuilder(client, indexName).
			Match("title", "默认刷新测试").
			Do(ctx)

		if searchResp != nil && searchResp.Hits.Total.Value > 0 {
			t.Logf("  立即搜索找到了文档（可能已自动刷新）")
		} else {
			t.Logf("  立即搜索未找到文档（等待自动刷新）")
		}

		// 等待一段时间后应该能找到
		time.Sleep(1 * time.Second)
		searchResp, err = NewSearchBuilder(client, indexName).
			Match("title", "默认刷新测试").
			Do(ctx)

		if err != nil {
			t.Fatalf("搜索失败: %v", err)
		}

		if searchResp.Hits.Total.Value < 1 {
			t.Errorf("等待后搜索应该能找到文档")
		} else {
			t.Logf("✓ 等待自动刷新后找到了文档")
		}
	})

	// 测试 4: Update 操作使用 refresh
	t.Run("Update with refresh", func(t *testing.T) {
		// 先创建文档
		_, _ = NewDocumentBuilder(client, indexName).
			ID("update-refresh-test").
			Set("title", "更新前").
			Set("views", 100).
			Refresh("true").
			Do(ctx)

		// 更新文档并立即刷新
		resp, err := NewDocumentBuilder(client, indexName).
			ID("update-refresh-test").
			Set("views", 200).
			Refresh("true").
			Update(ctx)

		if err != nil {
			t.Fatalf("更新文档失败: %v", err)
		}
		t.Logf("✓ 更新文档成功 (refresh=true): Result=%s, Version=%d", resp.Result, resp.Version)

		// 立即获取应该能看到更新
		getResp, _ := NewDocumentBuilder(client, indexName).
			ID("update-refresh-test").
			Get(ctx)

		if getResp != nil && getResp.Source["views"] == float64(200) {
			t.Logf("✓ 立即获取到了更新后的数据")
		}
	})

	// 测试 5: Delete 操作使用 refresh
	t.Run("Delete with refresh", func(t *testing.T) {
		// 先创建文档
		_, _ = NewDocumentBuilder(client, indexName).
			ID("delete-refresh-test").
			Set("title", "待删除").
			Refresh("true").
			Do(ctx)

		// 删除文档并立即刷新
		resp, err := NewDocumentBuilder(client, indexName).
			ID("delete-refresh-test").
			Refresh("true").
			Delete(ctx)

		if err != nil {
			t.Fatalf("删除文档失败: %v", err)
		}
		t.Logf("✓ 删除文档成功 (refresh=true): Result=%s", resp.Result)

		// 立即搜索应该找不到
		searchResp, _ := NewSearchBuilder(client, indexName).
			Match("title", "待删除").
			Do(ctx)

		if searchResp == nil || searchResp.Hits.Total.Value == 0 {
			t.Logf("✓ 使用 refresh=true 删除后立即确认文档已删除")
		}
	})
}
