# 文档操作 (DocumentBuilder)

DocumentBuilder 提供了完整的 Elasticsearch 文档 CRUD 操作。

## 基础操作

### 创建/索引文档

```go
resp, err := builder.NewDocumentBuilder(esClient, "products").
    ID("1").
    Set("name", "iPhone 15 Pro").
    Set("price", 999.99).
    Set("category", "electronics").
    Set("created_at", time.Now().Format("2006-01-02 15:04:05")).
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

## 支持的功能

- ✅ 索引文档 (Do)
- ✅ 创建文档 (Create)
- ✅ 更新文档 (Update)
- ✅ 脚本更新 (Script)
- ✅ Upsert (Upsert)
- ✅ 获取文档 (Get)
- ✅ 删除文档 (Delete)
- ✅ 检查存在 (Exists)
- ✅ 批量获取 (MGet)
