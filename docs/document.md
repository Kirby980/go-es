# 文档操作 (DocumentBuilder)

DocumentBuilder 提供了完整的 Elasticsearch 文档 CRUD 操作。支持两种 API 风格：**简洁风格**和 **Builder 风格**。

## 简洁风格（推荐）

通过 Client 的方法直接操作结构体，自动推断索引名。

### 定义结构体

```go
type Product struct {
    Name     string  `json:"name" es:"type:text;analyzer:ik_smart"`
    Price    float64 `json:"price" es:"type:float"`
    Category string  `json:"category" es:"type:keyword"`
    Stock    int     `json:"stock" es:"type:integer"`
}

// 实现 IndexName 接口自定义索引名（可选）
func (p *Product) IndexName() string {
    return "products"
}
```

### 创建文档

```go
product := &Product{
    Name:     "iPhone 15 Pro",
    Price:    999.99,
    Category: "electronics",
    Stock:    100,
}

// 自动生成 ID
resp, err := esClient.Create(ctx, product)

// 指定 ID
resp, err := esClient.CreateWithID(ctx, "product-1", product)
```

### 更新文档

```go
product.Price = 899.99
resp, err := esClient.Update(ctx, "product-1", product)
```

### Upsert（存在则更新，不存在则创建）

```go
resp, err := esClient.Upsert(ctx, "product-1", product)
```

### 获取文档

```go
getResp, err := esClient.Get(ctx, "products", "product-1")
if getResp.Found {
    fmt.Println(getResp.Source)
}
```

### 删除文档

```go
delResp, err := esClient.Delete(ctx, "products", "product-1")
```

## Builder 风格（更灵活）

使用 Builder 模式，支持更多细粒度控制。

### 使用 Model 方法

```go
// Model 方法可以自动推断索引名并设置文档数据
product := &Product{Name: "iPhone 15", Price: 999.99}

resp, err := builder.NewDocumentBuilder(esClient, "").
    Model(product).
    ID("1").
    Do(ctx)
```

### 手动设置字段

```go
resp, err := builder.NewDocumentBuilder(esClient, "products").
    ID("1").
    Set("name", "iPhone 15 Pro").
    Set("price", 999.99).
    Set("category", "electronics").
    Set("created_at", time.Now().Format("2006-01-02 15:04:05")).
    Do(ctx)
```

### 从 Map 设置

```go
data := map[string]interface{}{
    "name":  "MacBook Pro",
    "price": 1999.99,
}
resp, err := builder.NewDocumentBuilder(esClient, "products").
    ID("2").
    SetMap(data).
    Do(ctx)
```

### 从结构体设置

```go
resp, err := builder.NewDocumentBuilder(esClient, "products").
    ID("3").
    SetStruct(&Product{Name: "iPad", Price: 599.99}).
    Do(ctx)
```

### 获取文档

```go
getResp, err := builder.NewDocumentBuilder(esClient, "products").
    ID("1").
    Get(ctx)

if getResp.Found {
    fmt.Println(getResp.Source)
}
```

### 更新文档

```go
updateResp, err := builder.NewDocumentBuilder(esClient, "products").
    ID("1").
    Set("price", 899.99).
    Set("on_sale", true).
    Update(ctx)
```

### 使用脚本更新

```go
scriptResp, err := builder.NewDocumentBuilder(esClient, "products").
    ID("1").
    Script("ctx._source.quantity -= params.count",
           map[string]interface{}{"count": 5}).
    Update(ctx)
```

### Upsert (更新或插入)

```go
upsertResp, err := builder.NewDocumentBuilder(esClient, "products").
    ID("2").
    Set("name", "MacBook Pro").
    Set("price", 1999.99).
    Upsert(ctx)
```

### 删除文档

```go
delResp, err := builder.NewDocumentBuilder(esClient, "products").
    ID("1").
    Delete(ctx)
```

### 检查文档是否存在

```go
exists, err := builder.NewDocumentBuilder(esClient, "products").
    ID("1").
    Exists(ctx)
```

### 刷新策略

```go
// 立即刷新（文档立即可搜索）
resp, err := builder.NewDocumentBuilder(esClient, "products").
    ID("1").
    Set("name", "Product").
    Refresh("true").
    Do(ctx)

// 等待刷新完成
resp, err := builder.NewDocumentBuilder(esClient, "products").
    ID("1").
    Refresh("wait_for").
    Do(ctx)
```

### Debug 模式

```go
// 启用调试模式，打印请求和响应
resp, err := builder.NewDocumentBuilder(esClient, "products").
    Debug().
    ID("1").
    Set("name", "Test").
    Do(ctx)
```

## 支持的功能

- ✅ 简洁风格 CRUD (Create, Update, Upsert, Get, Delete)
- ✅ Model 方法（自动推断索引名）
- ✅ 索引文档 (Do)
- ✅ 创建文档 (Create)
- ✅ 更新文档 (Update)
- ✅ 脚本更新 (Script)
- ✅ Upsert (Upsert)
- ✅ 获取文档 (Get)
- ✅ 删除文档 (Delete)
- ✅ 检查存在 (Exists)
- ✅ 批量获取 (MGet)
- ✅ 刷新策略 (Refresh)
- ✅ Debug 模式
