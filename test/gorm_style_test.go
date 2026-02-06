package builder_test

import (
	"context"
	"testing"
	"time"

	"github.com/Kirby980/go-es/builder"
)

// Product 测试用结构体
type Product struct {
	Name     string  `json:"name" es:"type:text;analyzer:standard"`
	Price    float64 `json:"price" es:"type:float"`
	Category string  `json:"category" es:"type:keyword"`
	Stock    int     `json:"stock" es:"type:integer"`
}

// IndexName 实现 IndexName 接口
func (p *Product) IndexName() string {
	return "test_products"
}

// TestGormStyleCreate 测试 GORM 风格创建文档
func TestGormStyleCreate(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	// 先删除并创建索引
	_ = builder.NewIndexBuilder(client, "test_products").Delete(ctx)
	err := client.AutoMigrate(&Product{})
	if err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// 使用 GORM 风格创建文档
	product := &Product{
		Name:     "iPhone 15",
		Price:    999.99,
		Category: "electronics",
		Stock:    100,
	}

	resp, err := client.Create(ctx, product)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	t.Logf("✓ Create success: ID=%s, Index=%s", resp.ID, resp.Index)

	// 使用指定 ID 创建
	product2 := &Product{
		Name:     "MacBook Pro",
		Price:    1999.99,
		Category: "computers",
		Stock:    50,
	}
	resp2, err := client.CreateWithID(ctx, "macbook-1", product2)
	if err != nil {
		t.Fatalf("CreateWithID failed: %v", err)
	}
	t.Logf("✓ CreateWithID success: ID=%s", resp2.ID)
}

// TestGormStyleUpdate 测试 GORM 风格更新文档
func TestGormStyleUpdate(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	// 先创建文档
	product := &Product{
		Name:     "Test Product",
		Price:    100.00,
		Category: "test",
		Stock:    10,
	}
	resp, err := client.CreateWithID(ctx, "update-test-1", product)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	t.Logf("Created document: %s", resp.ID)

	time.Sleep(500 * time.Millisecond)

	// 更新文档
	product.Price = 150.00
	product.Stock = 20
	_, err = client.Update(ctx, "update-test-1", product)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	t.Logf("✓ Update success")
}

// TestGormStyleUpsert 测试 GORM 风格 Upsert
func TestGormStyleUpsert(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	product := &Product{
		Name:     "Upsert Product",
		Price:    200.00,
		Category: "test",
		Stock:    40,
	}

	// Upsert（不存在则创建）
	resp, err := client.Upsert(ctx, "upsert-test-1", product)
	if err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}
	t.Logf("✓ Upsert success: result=%s", resp.Result)
}

// TestGormStyleFindAndScan 测试 GORM 风格查询并扫描到结构体
func TestGormStyleFindAndScan(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()
	// 准备测试数据
	_ = builder.NewIndexBuilder(client, "test_products").Delete(ctx)
	_ = client.AutoMigrate(&Product{})
	time.Sleep(500 * time.Millisecond)

	products := []*Product{
		{Name: "iPhone 15", Price: 999.99, Category: "electronics", Stock: 100},
		{Name: "iPad Air", Price: 599.99, Category: "electronics", Stock: 80},
		{Name: "MacBook", Price: 1999.99, Category: "computers", Stock: 50},
	}

	for i, p := range products {
		_, err := client.CreateWithID(ctx, string(rune('1'+i)), p)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	time.Sleep(1 * time.Second) // 等待索引刷新

	// 查询并扫描到结构体
	var results []Product
	resp, err := client.Find("test_products").
		Term("category", "electronics").
		Size(10).
		Do(ctx)

	if err != nil {
		t.Fatalf("Find failed: %v", err)
	}

	// 使用 Scan 方法扫描结果
	if err := resp.Scan(&results); err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	t.Logf("✓ Find and Scan success: found %d products", len(results))
	for _, p := range results {
		t.Logf("  - %s: $%.2f", p.Name, p.Price)
	}
}

// TestDocumentBuilderModel 测试 DocumentBuilder.Model 方法
func TestDocumentBuilderModel(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	// 准备索引
	_ = builder.NewIndexBuilder(client, "test_products").Delete(ctx)
	_ = client.AutoMigrate(&Product{})
	time.Sleep(500 * time.Millisecond)

	// 使用 Model 方法
	product := &Product{
		Name:     "Model Test Product",
		Price:    299.99,
		Category: "test",
		Stock:    5,
	}

	resp, err := builder.NewDocumentBuilder(client, "").
		Model(product).
		ID("model-test-1").
		Do(ctx)

	if err != nil {
		t.Fatalf("Model failed: %v", err)
	}

	if resp.Index != "test_products" {
		t.Errorf("Expected index 'test_products', got '%s'", resp.Index)
	}

	t.Logf("✓ Model method success: Index=%s, ID=%s", resp.Index, resp.ID)
}

// TestSearchResponseScan 测试 SearchResponse.Scan 方法
func TestSearchResponseScan(t *testing.T) {
	client := createTestClient(t)
	defer client.Close()
	ctx := context.Background()

	// 准备测试数据
	_ = builder.NewIndexBuilder(client, "test_products").Delete(ctx)
	_ = client.AutoMigrate(&Product{})
	time.Sleep(500 * time.Millisecond)

	// 创建一些测试数据
	for i := 0; i < 5; i++ {
		product := &Product{
			Name:     "Scan Test Product " + string(rune('A'+i)),
			Price:    float64(100 + i*50),
			Category: "scan-test",
			Stock:    i * 10,
		}
		_, _ = client.CreateWithID(ctx, "scan-"+string(rune('1'+i)), product)
	}

	time.Sleep(1 * time.Second)

	// 搜索
	resp, err := builder.NewSearchBuilder(client, "test_products").
		Term("category", "scan-test").
		Sort("price", "asc").
		Do(ctx)

	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// 测试扫描到值类型切片
	var products []Product
	if err := resp.Scan(&products); err != nil {
		t.Fatalf("Scan to []Product failed: %v", err)
	}
	t.Logf("✓ Scan to []Product: %d items", len(products))

	// 测试扫描到指针类型切片
	var productPtrs []*Product
	if err := resp.Scan(&productPtrs); err != nil {
		t.Fatalf("Scan to []*Product failed: %v", err)
	}
	t.Logf("✓ Scan to []*Product: %d items", len(productPtrs))

	// 验证数据
	for _, p := range products {
		t.Logf("  - %s: $%.2f, stock=%d", p.Name, p.Price, p.Stock)
	}
}
