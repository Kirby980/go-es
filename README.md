# Go Elasticsearch 链式调用 API

一个类似 GORM 的 Elasticsearch Go 客户端，提供优雅的链式调用API。

## 特性

- ✅ **链式调用**: 类似 GORM 的优雅 API 设计
- ✅ **完整功能**: 支持索引、文档、搜索、聚合、批量操作、集群管理
- ✅ **类型安全**: 使用 Go 结构体，避免手写 JSON
- ✅ **易于使用**: 简洁的 API，降低学习成本
- ✅ **高性能**: 支持批量操作和连接池
- ✅ **错误处理**: 完善的错误处理和重试机制
- ✅ **链式Debug**: 类似GORM的Debug模式，局部控制日志输出

## 快速开始

### 安装

```bash
go get github.com/Kirby980/go-es
```

### 创建客户端

```go
import (
    "github.com/Kirby980/go-es/builder"
    "github.com/Kirby980/go-es/client"
    "github.com/Kirby980/go-es/config"
)

esClient, err := client.New(
    config.WithAddresses("https://localhost:9200"),
    config.WithAuth("elastic", "password"),
    config.WithTransport(true), // 跳过 SSL 验证
    config.WithTimeout(10*time.Second),
    config.WithMaxConnsPerHost(100),
    config.WithMaxIdConns(200),
    config.WithMaxIdleConnsPerHost(50),
    config.WithIdleConnTimeout(90*time.Second),
)
defer esClient.Close()

ctx := context.Background()
```

## 核心功能示例

### 1. 索引管理

```go
// 创建索引
err := builder.NewIndexBuilder(esClient, "products").
    Shards(1).
    Replicas(0).
    AddProperty("name", builder.FieldTypeText, builder.WithAnalyzer(builder.AnalyzerIKSmart)).
    AddProperty("price", builder.FieldTypeFloat).
    AddProperty("category", builder.FieldTypeKeyword).
    Create(ctx)

// 检查索引是否存在
exists, _ := builder.NewIndexBuilder(esClient, "products").Exists(ctx)

// 删除索引
err = builder.NewIndexBuilder(esClient, "products").Delete(ctx)
```

[查看完整索引管理文档](docs/index.md)

### 2. 文档操作

```go
// 创建文档
resp, err := builder.NewDocumentBuilder(esClient, "products").
    ID("1").
    Set("name", "iPhone 15 Pro").
    Set("price", 999.99).
    Set("category", "electronics").
    Do(ctx)

// 获取文档
getResp, err := builder.NewDocumentBuilder(esClient, "products").
    ID("1").
    Get(ctx)

// 更新文档
updateResp, err := builder.NewDocumentBuilder(esClient, "products").
    ID("1").
    Set("price", 899.99).
    Update(ctx)

// 删除文档
delResp, err := builder.NewDocumentBuilder(esClient, "products").
    ID("1").
    Delete(ctx)
```

[查看完整文档操作文档](docs/document.md)

### 3. 搜索

```go
// 基础搜索
searchResp, err := builder.NewSearchBuilder(esClient, "products").
    Match("name", "iPhone").
    Term("category", "electronics").
    Range("price", 500, 1500).
    Sort("price", "desc").
    From(0).
    Size(10).
    Do(ctx)

// 复杂布尔查询
resp, err := builder.NewSearchBuilder(esClient, "products").
    Term("category", "electronics").           // AND
    Term("status", "active").                   // AND
    Range("price", 100, 1000).                  // AND
    MatchShould("brand", "Apple").              // OR
    MatchShould("brand", "Samsung").            // OR
    MinimumShouldMatch(1).                      // 至少匹配1个OR条件
    MatchMustNot("title", "refurbished").       // NOT
    Do(ctx)

// 快速计数
count, _ := builder.NewSearchBuilder(esClient, "products").
    Match("status", "active").
    Count(ctx)
```

[查看完整搜索文档](docs/search.md)

### 4. 聚合分析

```go
// 统计聚合
aggResp, _ := builder.NewAggregationBuilder(esClient, "products").
    Avg("avg_price", "price").
    Sum("total_quantity", "quantity").
    Min("min_price", "price").
    Max("max_price", "price").
    Do(ctx)

// 分组聚合
termsResp, _ := builder.NewAggregationBuilder(esClient, "products").
    Terms("by_category", "category", 10).
    Do(ctx)

// 日期直方图
dateHistResp, _ := builder.NewAggregationBuilder(esClient, "orders").
    DateHistogram("orders_over_time", "created_at", "1d").
    Do(ctx)
```

[查看完整聚合文档](docs/aggregation.md)

### 5. 批量操作

```go
bulkResp, err := builder.NewBulkBuilder(esClient).
    Index("products").
    Add("", "1", map[string]interface{}{"name": "iPad Air", "price": 599.99}).
    Add("", "2", map[string]interface{}{"name": "Apple Watch", "price": 399.99}).
    Update("", "3", map[string]interface{}{"price": 349.99}).
    Delete("", "4").
    Do(ctx)

// 检查结果
if bulkResp.HasErrors() {
    for _, item := range bulkResp.FailedItems() {
        fmt.Printf("失败: ID=%s, 错误=%s\n", item.ID, item.Error.Reason)
    }
}
```

### 6. 深度分页

```go
// Scroll 遍历（适合大数据导出）
scroll := builder.NewScrollBuilder(esClient, "products").
    Match("status", "active").
    Size(1000).
    KeepAlive("5m")

resp, _ := scroll.Do(ctx)
for scroll.HasMore(resp) {
    resp, _ = scroll.Next(ctx)
    // 处理数据...
}
scroll.Clear(ctx)

// Search After（适合实时分页）
searchAfter := builder.NewSearchAfterBuilder(esClient, "products").
    Match("status", "active").
    Sort("price", "asc").
    Sort("_id", "asc").
    Size(20)

resp, _ := searchAfter.Do(ctx)
resp, _ = searchAfter.Next(ctx)  // 下一页
```

[查看完整高级功能文档](docs/advanced.md)

### 7. Debug模式

```go
// 启用Debug模式查看请求和响应（类似GORM）
resp, err := builder.NewSearchBuilder(esClient, "products").
    Debug().  // 启用调试
    Match("name", "iPhone").
    Do(ctx)

// 所有Builder都支持Debug
builder.NewDocumentBuilder(esClient, "products").Debug().ID("1").Get(ctx)
builder.NewBulkBuilder(esClient).Debug().Add(...).Do(ctx)
```

### 8. 错误处理

```go
import "github.com/Kirby980/go-es/errors"

resp, err := builder.NewDocumentBuilder(esClient, "products").ID("123").Get(ctx)
if err != nil {
    if esErr, ok := err.(*errors.ESError); ok {
        if esErr.IsNotFound() {
            // 文档不存在 (404)
        } else if esErr.IsConflict() {
            // 版本冲突 (409)
        } else if esErr.IsBadRequest() {
            // 请求错误 (400)
        } else if esErr.IsTimeout() {
            // 请求超时 (408)
        }
        // 详细信息
        fmt.Printf("错误: %s - %s\n", esErr.Type, esErr.Reason)
    }
}
```

[查看完整错误处理文档](docs/errors.md)

## 线程安全说明

### Client 是线程安全的

`*client.Client` 可以在多个 goroutine 中并发使用：

```go
esClient, _ := client.New(...)

// ✅ 安全：多个 goroutine 共享 client
go func() {
    builder.NewSearchBuilder(esClient, "index1").Match(...).Do(ctx)
}()
go func() {
    builder.NewSearchBuilder(esClient, "index2").Match(...).Do(ctx)
}()
```

### Builder 不是线程安全的

所有 Builder（SearchBuilder、DocumentBuilder 等）都**不是线程安全**的，不能在多个 goroutine 中共享使用：

```go
// ❌ 错误：多个 goroutine 共享同一个 Builder
sb := builder.NewSearchBuilder(esClient, "index")
go func() { sb.Match("field1", "value1").Do(ctx) }()  // 数据竞争！
go func() { sb.Match("field2", "value2").Do(ctx) }()  // 数据竞争！

// ✅ 正确：每个 goroutine 创建自己的 Builder
go func() {
    builder.NewSearchBuilder(esClient, "index").Match("field1", "value1").Do(ctx)
}()
go func() {
    builder.NewSearchBuilder(esClient, "index").Match("field2", "value2").Do(ctx)
}()
```

### 并发最佳实践

1. **全局共享 Client**：应用启动时创建一个 Client，全局共享
2. **每次查询创建新 Builder**：不要重复使用 Builder 实例
3. **使用连接池**：配置合适的连接池参数提升并发性能

```go
var esClient *client.Client

func init() {
    esClient, _ = client.New(
        config.WithAddresses("https://localhost:9200"),
        config.WithConnectionPool(200, 50, 100), // 高并发配置
    )
}

func SearchProducts(ctx context.Context, keyword string) {
    // 每次查询创建新的 Builder
    resp, _ := builder.NewSearchBuilder(esClient, "products").
        Match("name", keyword).
        Do(ctx)
    // ...
}
```

## 完整文档

- [索引管理](docs/index.md) - 创建、更新、删除索引，自定义分析器
- [文档操作](docs/document.md) - 文档 CRUD 操作
- [搜索](docs/search.md) - 全文搜索、精确查询、布尔查询、地理查询
- [聚合分析](docs/aggregation.md) - 指标聚合、桶聚合、管道聚合
- [高级功能](docs/advanced.md) - 批量操作、深度分页、集群管理
- [错误处理](docs/errors.md) - 错误类型判断和处理

## API 对比表

| 功能 | Elasticsearch REST API | go-es 链式调用 |
|------|------------------------|----------------|
| 创建索引 | PUT /index | `NewIndexBuilder(client, "index").Shards(1).Create(ctx)` |
| 索引文档 | PUT /index/_doc/1 | `NewDocumentBuilder(client, "index").ID("1").Set("field", value).Do(ctx)` |
| 搜索 | POST /index/_search | `NewSearchBuilder(client, "index").Match("field", "value").Do(ctx)` |
| 计数 | POST /index/_count | `NewSearchBuilder(client, "index").Match("field", "value").Count(ctx)` |
| 批量操作 | POST /_bulk | `NewBulkBuilder(client).Add(...).Update(...).Do(ctx)` |
| Scroll遍历 | POST /index/_search?scroll=5m | `NewScrollBuilder(client, "index").Size(1000).Do(ctx)` |
| Search After | POST /index/_search (with search_after) | `NewSearchAfterBuilder(client, "index").Sort("price", "asc").Do(ctx)` |

## 支持的功能

### IndexBuilder
- ✅ 创建/更新/删除索引
- ✅ 自定义分析器
- ✅ 字段映射
- ✅ 别名管理

### DocumentBuilder
- ✅ 文档 CRUD
- ✅ 脚本更新
- ✅ Upsert

### SearchBuilder
- ✅ 全文搜索 (Match, MultiMatch)
- ✅ 精确查询 (Term, Terms)
- ✅ 范围查询 (Range)
- ✅ 模糊查询 (Fuzzy, Wildcard, Prefix, Regexp)
- ✅ 布尔查询 (Must, Should, MustNot)
- ✅ 地理查询 (GeoDistance, GeoBoundingBox)
- ✅ 排序、分页、高亮、字段过滤

### AggregationBuilder
- ✅ 指标聚合 (Avg, Sum, Min, Max, Stats, Cardinality, Percentiles)
- ✅ 桶聚合 (Terms, Histogram, DateHistogram, Range)
- ✅ 管道聚合 (AvgBucket, SumBucket, MovingAvg, Derivative)
- ✅ 地理聚合 (GeoBounds, GeoCentroid, GeoDistance)

### BulkBuilder
- ✅ 批量索引/创建/更新/删除
- ✅ 错误处理

### 其他功能
- ✅ UpdateByQuery (按条件批量更新)
- ✅ DeleteByQuery (按条件批量删除)
- ✅ Scroll (深度分页遍历)
- ✅ SearchAfter (高效深度分页)
- ✅ ClusterBuilder (集群管理)
- ✅ Debug模式 (类似GORM)

## 配置选项

```go
esClient, err := client.New(
    config.WithAddresses("https://localhost:9200"),      // ES 地址
    config.WithAuth("username", "password"),             // 认证
    config.WithTransport(true),                          // 跳过 SSL 验证
    config.WithTimeout(30*time.Second),                  // 超时时间
    config.WithRetry(3, time.Second),                    // 重试配置
    config.WithDebug(true),                              // 调试模式
    config.WithMaxConnsPerHost(100),                     // 每个 host 的最大连接数
    config.WithMaxIdConns(200),                          // 最大空闲连接数
    config.WithMaxIdleConnsPerHost(50),                  // 每个 host 的最大空闲连接数
    config.WithIdleConnTimeout(90*time.Second),          // 空闲连接超时时间
)
```

## 完整示例

查看 `examples/complete_api_test.go` 获取完整的使用示例。

```bash
# 运行完整示例
go test -v ./examples -run TestCompleteAPI
```

## License

MIT License
