# Go Elasticsearch 链式调用 API

一个类似 GORM 的 Elasticsearch Go 客户端，提供优雅的链式调用API。

## 特性

- ✅ **链式调用**: 链式的API调用
- ✅ **完整功能**: 支持索引、文档、搜索、聚合、批量操作、集群管理
- ✅ **进阶 API**: PIT、KNN、Async Search、SQL、Search Template、Stored Script、ILM、Cat APIs、Snapshot & Restore
- ✅ **运维能力**: Rollover、Field Caps、Rank Eval、Termvectors 等常用诊断/评估接口
- ✅ **类型安全**: 使用 Go 结构体，避免手写 JSON
- ✅ **易于使用**: 简洁的 API，降低学习成本
- ✅ **高性能**: 支持批量操作、连接池，内部重写了流式 JSON 解析与批量 NDJSON 零反射编码，极大地降低了 CPU 和内存分配（避免 OOM）。
- ✅ **错误处理**: 完善的错误处理和重试机制
- ✅ **链式Debug**: Debug 模式，局部控制日志输出（支持持久 Debug）
- ✅ **AutoMigrate**: 自动迁移，通过结构体标签定义映射

## 快速开始

### 版本兼容性

本库设计上兼容 **Elasticsearch 7.x**、**Elasticsearch 8.x** 以及 **Elasticsearch 9.x** 版本。由于底层的 JSON API 和查询 DSL 具有很强的相似性，大部分操作可以无缝兼容。不过，Elasticsearch 8 及以上版本引入了一些新特性（例如默认的安全机制），在创建客户端时请确保提供正确的认证和证书配置（或使用 `config.WithInsecureSkipVerify(true)` 跳过验证）。

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
    config.WithAddresses(
        "https://localhost:9200",
        "https://localhost:9201",
    ),
    config.WithAuth("elastic", "password"),
    config.WithInsecureSkipVerify(true), // 跳过 SSL 验证
    config.WithTimeout(10*time.Second),  // 默认请求超时时间
    config.WithRetry(3, 200*time.Millisecond),
    config.WithExponentialBackoff(true, 30*time.Second),
    config.WithCircuitBreaker(true, 3, 10*time.Second, 5*time.Second),
    config.WithGzip(true),
    config.WithSniff(true, 5*time.Minute),
    config.WithMaxConnsPerHost(100),
    config.WithMaxIdleConns(200),
    config.WithMaxIdleConnsPerHost(50),
    config.WithIdleConnTimeout(90*time.Second),
)
defer esClient.Close()

ctx := context.Background()
```

多地址会默认轮询选择节点；当开启熔断器时，会在节点连续失败后短暂标记为不可用，并在后续请求中自动切换到其他地址。
当开启 Sniff 时，客户端会定时从 `/_nodes/http` 拉取节点列表并更新地址池，以适配集群扩缩容和节点漂移。
Sniff 默认关闭；如需显式关闭可不传该配置，或使用 `config.WithSniff(false, 0)`。

## 核心功能示例

### 1. 索引管理

```go
import esconst "github.com/Kirby980/go-es/const"

// 创建索引
err := builder.NewIndexBuilder(esClient, "products").
    Shards(1).
    Replicas(0).
    AddProperty("name", esconst.FieldTypeText, builder.WithAnalyzer(esconst.AnalyzerIKSmart)).
    AddProperty("price", esconst.FieldTypeFloat).
    AddProperty("category", esconst.FieldTypeKeyword).
    Create(ctx)

// 检查索引是否存在
exists, _ := builder.NewIndexBuilder(esClient, "products").Exists(ctx)

// 删除索引
err = builder.NewIndexBuilder(esClient, "products").Delete(ctx)
```

#### AutoMigrate

```go
// 定义结构体，使用 es 标签定义字段映射
type Product struct {
    Name      string  `es:"type:text;analyzer:ik_smart"`
    Price     float64 `es:"type:float"`
    Category  string  `es:"type:keyword"`
    CreatedAt string  `es:"type:date;format:yyyy-MM-dd HH:mm:ss"`
}

// 可选：实现 IndexName 接口自定义索引名
func (p *Product) IndexName() string {
    return "my_products"
}

// 自动迁移（创建或更新索引）
import "github.com/Kirby980/go-es/sugar"

s := sugar.New(esClient)
err := s.AutoMigrate(&Product{})
```

[查看完整索引管理文档](docs/index.md)

### 2. 文档操作

本库支持两种 API 风格：**简洁风格**（类似 GORM）和 **Builder 风格**（更灵活）。

#### 简洁风格（推荐）

```go
import "github.com/Kirby980/go-es/sugar"

// 定义结构体
type Product struct {
    Name     string  `json:"name" es:"type:text"`
    Price    float64 `json:"price" es:"type:float"`
    Category string  `json:"category" es:"type:keyword"`
}

func (p *Product) IndexName() string { return "products" }

// 创建文档
product := &Product{Name: "iPhone 15", Price: 999.99, Category: "electronics"}

s := sugar.New(esClient)

// 创建文档（自动生成 ID）
resp, err := s.Create(ctx, product)

// 创建文档（指定 ID）
resp, err := s.CreateWithID(ctx, "product-1", product)

// 更新文档
product.Price = 899.99
resp, err := s.Update(ctx, "product-1", product)

// Upsert（存在则更新，不存在则创建）
resp, err := s.Upsert(ctx, "product-1", product)

// 获取文档
getResp, err := s.Get(ctx, "products", "product-1")

// 删除文档
delResp, err := s.Delete(ctx, "products", "product-1")
```

#### Builder 风格（更灵活）

```go
// 创建文档
resp, err := builder.NewDocumentBuilder(esClient, "products").
    ID("1").
    Set("name", "iPhone 15 Pro").
    Set("price", 999.99).
    Set("category", "electronics").
    Do(ctx)

// 使用 Model 方法从结构体创建（自动推断索引名）
resp, err := builder.NewDocumentBuilder(esClient, "").
    Model(&Product{Name: "iPhone", Price: 999.99}).
    ID("1").
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

// 搜索结果扫描到结构体切片
var products []Product
if err := searchResp.Scan(&products); err != nil {
    // handle error
}
for _, p := range products {
    fmt.Printf("%s: $%.2f\n", p.Name, p.Price)
}

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

// 使用 Find 方法（简洁风格）
resp, _ := sugar.New(esClient).Find("products").
    Match("name", "iPhone").
    Size(10).
    Do(ctx)
var results []Product
resp.Scan(&results)
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
    Add("", "1", map[string]any{"name": "iPad Air", "price": 599.99}).
    Add("", "2", map[string]any{"name": "Apple Watch", "price": 399.99}).
    Update("", "3", map[string]any{"price": 349.99}).
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
import (
    "errors"

    eserrors "github.com/Kirby980/go-es/errors"
)

resp, err := builder.NewDocumentBuilder(esClient, "products").ID("123").Get(ctx)
if err != nil {
    var esErr *eserrors.ESError
    if errors.As(err, &esErr) {
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
- ✅ AutoMigrate（类似 GORM 的自动迁移）

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
- ✅ PointInTime / PIT (现代 ES 推荐的分页方式, 替代 Scroll)
- ✅ KNN 向量搜索 (支持 ES 8.x / 9.x 的 AI 向量检索)
- ✅ Reindex (跨索引重建数据)
- ✅ Aliases (/_aliases 原子切换别名，零停机切流量)
- ✅ Index Template / Component Template (模板与组件模板管理，适配 ES 7/8/9)
- ✅ Data Stream (数据流管理，适配 ES 7/8/9)
- ✅ Ingest Pipeline (写入预处理管道管理与 simulate，适配 ES 7/8/9)
- ✅ EQL (Event Query Language, 支持安全日志事件查询)
- ✅ Multi Search (msearch 批量并发搜索)
- ✅ Explain API (单文档搜索评分详情解释)
- ✅ Field Collapsing (字段折叠去重)
- ✅ Suggest (搜索建议/自动补全)
- ✅ ClusterBuilder (集群管理)
- ✅ Debug模式 (类似GORM)

### Alias 原子切换示例

```go
resp, err := builder.NewAliasesBuilder(esClient).
    Replace("products_write", "products_v1", "products_v2", builder.WithIsWriteIndex(true)).
    Do(ctx)
_ = resp
_ = err
```

### Template / Data Stream / Pipeline 示例

```go
_, _ = builder.NewComponentTemplateBuilder(esClient, "ct_common").
    TemplateSettings(map[string]any{"index.number_of_shards": 1}).
    Put(ctx)

_, _ = builder.NewIndexTemplateBuilder(esClient, "it_logs").
    IndexPatterns("logs-*").
    ComposedOf("ct_common").
    DataStream(true).
    Put(ctx)

_, _ = builder.NewDataStreamBuilder(esClient, "logs-app").Create(ctx)

_, _ = builder.NewIngestPipelineBuilder(esClient, "p1").
    Description("add field").
    AddProcessor("set", map[string]any{"field": "env", "value": "prod"}).
    Put(ctx)
```

## 配置选项

```go
esClient, err := client.New(
    config.WithAddresses("https://localhost:9200"),      // ES 地址
    config.WithAuth("username", "password"),             // 认证
    config.WithInsecureSkipVerify(true),                 // 跳过 SSL 验证
    config.WithTimeout(30*time.Second),                  // 超时时间
    config.WithRetry(3, time.Second),                    // 重试配置
    config.WithDebug(true),                              // 调试模式
    config.WithMaxConnsPerHost(100),                     // 每个 host 的最大连接数
    config.WithMaxIdleConns(200),                        // 最大空闲连接数
    config.WithMaxIdleConnsPerHost(50),                  // 每个 host 的最大空闲连接数
    config.WithIdleConnTimeout(90*time.Second),          // 空闲连接超时时间
    config.WithHooks(hooks.NewMetricsHook()),            // 注册可观测性 Hook
)
```

## 进阶功能示例

### BulkProcessor（高吞吐写入缓冲池）

`BulkBuilder` 原生支持自动分批刷新（Auto Flush），只需设置分批大小即可实现高吞吐写入：

```go
import "github.com/Kirby980/go-es/builder"

// 1. 初始化 BulkBuilder
bulk := builder.NewBulkBuilder(esClient).
    Index("bulk_test_index").
    AutoFlushSize(1000). // 设置每 1000 条自动刷新一次
    AutoFlushContext(ctx).
    OnFlush(func(resp *builder.BulkResponse) {
        fmt.Printf("已刷新 %d 条数据，成功: %d，失败: %d\n", 
            len(resp.Items), resp.SuccessCount(), len(resp.Failed()))
    })

// 2. 持续写入数据
for i := 0; i < 5500; i++ {
    bulk.Add("", fmt.Sprintf("doc_%d", i), map[string]any{"value": i})
}

// 3. 最后手动 Flush 剩余的数据
resp, err := bulk.Flush(ctx)
```

### SQL API

```go
resp, err := builder.NewSQLBuilder(esClient).
    Query("SELECT * FROM products ORDER BY price DESC").
    FetchSize(10).
    Do(ctx)
```

### Async Search

```go
resp, err := builder.NewAsyncSearchBuilder(esClient).
    Index("products").
    Query(map[string]any{"match_all": map[string]any{}}).
    WaitForCompletionTimeout("1s").
    KeepAlive("5m").
    Do(ctx)

if resp.IsRunning {
    _, _ = builder.NewAsyncSearchBuilder(esClient).Get(ctx, resp.ID)
}
```

### Search Template / Stored Script

```go
_, _ = builder.NewScriptBuilder(esClient).
    ID("my_script").
    Script("painless", "return 1", nil).
    Put(ctx)

_, _ = builder.NewSearchTemplateBuilder(esClient).
    Index("products").
    ID("my_search_tpl").
    Params(map[string]any{"q": "iphone"}).
    Do(ctx)
```

### Rollover / Field Caps / Rank Eval / Termvectors

```go
_, _ = builder.NewRolloverBuilder(esClient, "logs_write").
    NewIndex("logs-000002").
    Condition("max_docs", 1000000).
    Do(ctx)

_, _ = builder.NewFieldCapsBuilder(esClient, "products").
    Fields("name", "price").
    IncludeUnmapped(true).
    Do(ctx)

_, _ = builder.NewRankEvalBuilder(esClient, "products").
    Body(map[string]any{"requests": []any{}, "metric": map[string]any{"precision": map[string]any{}}}).
    Do(ctx)

_, _ = builder.NewTermvectorsBuilder(esClient, "products").
    ID("1").
    Body(map[string]any{"fields": []any{"name"}}).
    Do(ctx)
```

### DebugPersistent（持久 Debug）

```go
b := builder.NewSearchBuilder(esClient, "products")
b.DebugPersistent(true)
_, _ = b.Match("name", "iphone").Do(ctx)
_, _ = b.Match("name", "mac").Do(ctx)
```

### 可观测性 Hook（Client Trace/Metrics）

通过 `config.WithHooks` 可以注入自定义的可观测性插件，拦截并记录每一次 HTTP 请求的生命周期。

```go
import (
    "github.com/Kirby980/go-es/client"
    "github.com/Kirby980/go-es/config"
    "github.com/Kirby980/go-es/hooks"
)

// 实例化内置的 Metrics 和 Log Hook
metricsHook := hooks.NewMetricsHook()
logHook := hooks.NewLogHook(loggerInstance)

// 注入客户端
esClient, err := client.New(
    config.WithAddresses("http://localhost:9200"),
    config.WithHooks(metricsHook, logHook),
)

// 发起请求后可以获取指标
reqs, errs, avgLatency := metricsHook.GetMetrics()
fmt.Printf("请求数: %d, 错误数: %d, 平均延迟: %v\n", reqs, errs, avgLatency)
```

## 完整示例

查看 `examples/complete_api_test.go` 获取完整的使用示例。

```bash
# 运行完整示例
go test -v ./examples -run TestCompleteAPI
```
