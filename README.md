# Go Elasticsearch 链式调用 API

一个类似 GORM 的 Elasticsearch Go 客户端，提供优雅的链式调用API。

## 特性

- ✅ **链式调用**: 类似 GORM 的优雅 API 设计
- ✅ **完整功能**: 支持索引、文档、搜索、聚合、批量操作、集群管理
- ✅ **类型安全**: 使用 Go 结构体，避免手写 JSON
- ✅ **易于使用**: 简洁的 API，降低学习成本
- ✅ **高性能**: 支持批量操作和连接池
- ✅ **错误处理**: 完善的错误处理和重试机制

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
)
defer esClient.Close()

ctx := context.Background()
```

## 核心功能

### 1. 索引管理 (IndexBuilder)

```go
// 创建索引
err := builder.NewIndexBuilder(esClient, "products").
    Shards(1).
    Replicas(0).
    RefreshInterval("1s").
    AddProperty("name", "text", builder.WithAnalyzer("ik_smart")).
    AddProperty("price", "float").
    AddProperty("category", "keyword").
    AddProperty("created_at", "date", builder.WithFormat("yyyy-MM-dd HH:mm:ss")).
    AddAlias("products-alias", nil).
    Do(ctx)

// 检查索引是否存在
exists, _ := builder.NewIndexBuilder(esClient, "products").Exists(ctx)

// 获取索引信息
info, _ := builder.NewIndexBuilder(esClient, "products").Get(ctx)
fmt.Println(info.PrettyJSON())

// 删除索引
err = builder.NewIndexBuilder(esClient, "products").Delete(ctx)
```

### 2. 文档操作 (DocumentBuilder)

```go
// 创建/索引文档
resp, err := builder.NewDocumentBuilder(esClient, "products").
    ID("1").
    Set("name", "iPhone 15 Pro").
    Set("price", 999.99).
    Set("category", "electronics").
    Set("created_at", time.Now().Format("2006-01-02 15:04:05")).
    Do(ctx)

// 获取文档
getResp, err := builder.NewDocumentBuilder(esClient, "products").
    ID("1").
    Get(ctx)
if getResp.Found {
    fmt.Println(getResp.Source)
}

// 更新文档
updateResp, err := builder.NewDocumentBuilder(esClient, "products").
    ID("1").
    Set("price", 899.99).
    Set("on_sale", true).
    Update(ctx)

// 使用脚本更新
scriptResp, err := builder.NewDocumentBuilder(esClient, "products").
    ID("1").
    Script("ctx._source.quantity -= params.count",
           map[string]interface{}{"count": 5}).
    Update(ctx)

// Upsert (更新或插入)
upsertResp, err := builder.NewDocumentBuilder(esClient, "products").
    ID("2").
    Set("name", "MacBook Pro").
    Set("price", 1999.99).
    Upsert(ctx)

// 删除文档
delResp, err := builder.NewDocumentBuilder(esClient, "products").
    ID("1").
    Delete(ctx)

// 检查文档是否存在
exists, _ := builder.NewDocumentBuilder(esClient, "products").
    ID("1").
    Exists(ctx)
```

### 3. 批量操作 (BulkBuilder)

```go
bulkResp, err := builder.NewBulkBuilder(esClient).
    Index("products").
    Add("", "1", map[string]interface{}{
        "name": "iPad Air",
        "price": 599.99,
    }).
    Add("", "2", map[string]interface{}{
        "name": "Apple Watch",
        "price": 399.99,
    }).
    Update("", "3", map[string]interface{}{
        "price": 349.99,
    }).
    Delete("", "4").
    Do(ctx)

// 检查结果
fmt.Printf("成功: %d, 失败: %d\n",
    bulkResp.SuccessCount(),
    len(bulkResp.FailedItems()))

if bulkResp.HasErrors() {
    for _, item := range bulkResp.FailedItems() {
        fmt.Printf("失败: ID=%s, 错误=%s\n", item.ID, item.Error.Reason)
    }
}
```

### 4. 搜索 (SearchBuilder)

```go
// 基础搜索
searchResp, err := builder.NewSearchBuilder(esClient, "products").
    Match("name", "iPhone").
    Term("category", "electronics").
    Range("price", 500, 1500).
    Sort("price", "desc").
    From(0).
    Size(10).
    Highlight("name", "description").
    Source("name", "price", "category").
    Do(ctx)

fmt.Printf("找到 %d 条结果\n", searchResp.Hits.Total.Value)
for _, hit := range searchResp.Hits.Hits {
    fmt.Printf("- [%s] %v (score: %.2f)\n", hit.ID, hit.Source, hit.Score)
}

// 多字段匹配
multiResp, _ := builder.NewSearchBuilder(esClient, "products").
    MultiMatch("Apple phone", "name", "description").
    Size(5).
    Do(ctx)

// 模糊搜索
fuzzyResp, _ := builder.NewSearchBuilder(esClient, "products").
    Fuzzy("name", "iPhon", "AUTO").
    Do(ctx)

// 前缀搜索
prefixResp, _ := builder.NewSearchBuilder(esClient, "products").
    Prefix("name", "iPa").
    Do(ctx)

// 通配符搜索
wildcardResp, _ := builder.NewSearchBuilder(esClient, "products").
    Wildcard("name", "i*one").
    Do(ctx)

// 正则表达式搜索
regexpResp, _ := builder.NewSearchBuilder(esClient, "products").
    Regexp("name", "i.*one").
    Do(ctx)

// QueryString
qsResp, _ := builder.NewSearchBuilder(esClient, "products").
    QueryString("(iPhone OR iPad) AND electronics", "name", "category").
    Do(ctx)

// 复杂组合查询
complexResp, _ := builder.NewSearchBuilder(esClient, "products").
    Should(
        func(b *builder.SearchBuilder) {
            b.Match("name", "iPhone")
        },
        func(b *builder.SearchBuilder) {
            b.Match("name", "Samsung")
        },
    ).
    Range("price", nil, 2000).
    MustNot("category", "refurbished").
    Exists("created_at").
    Sort("price", "asc").
    Sort("rating", "desc").
    Size(20).
    Do(ctx)

// 地理位置查询
geoResp, _ := builder.NewSearchBuilder(esClient, "stores").
    GeoDistance("location", 37.7749, -122.4194, "10km").
    Do(ctx)
```

### 5. 聚合分析 (AggregationBuilder)

#### 指标聚合

```go
// 统计聚合
aggResp, _ := builder.NewAggregationBuilder(esClient, "products").
    Avg("avg_price", "price").
    Sum("total_quantity", "quantity").
    Min("min_price", "price").
    Max("max_price", "price").
    Stats("price_stats", "price").
    Cardinality("unique_categories", "category").
    Do(ctx)

fmt.Println(aggResp.PrettyJSON())
```

#### 桶聚合

```go
// 分组聚合
termsResp, _ := builder.NewAggregationBuilder(esClient, "products").
    Terms("by_category", "category", 10).
    Do(ctx)

// 带排序的分组
termsResp, _ := builder.NewAggregationBuilder(esClient, "products").
    TermsWithOrder("top_categories", "category", 5, "_count", "desc").
    Do(ctx)

// 直方图
histResp, _ := builder.NewAggregationBuilder(esClient, "products").
    Histogram("price_distribution", "price", 100).
    Do(ctx)

// 日期直方图
dateHistResp, _ := builder.NewAggregationBuilder(esClient, "orders").
    DateHistogram("orders_over_time", "created_at", "1d").
    Do(ctx)

// 范围聚合
rangeResp, _ := builder.NewAggregationBuilder(esClient, "products").
    Range("price_ranges", "price", []map[string]interface{}{
        {"key": "cheap", "to": 300},
        {"key": "medium", "from": 300, "to": 1000},
        {"key": "expensive", "from": 1000},
    }).
    Do(ctx)
```

#### 管道聚合

```go
// 平均桶聚合
aggResp, _ := builder.NewAggregationBuilder(esClient, "sales").
    DateHistogram("sales_per_month", "date", "1M").
    AvgBucket("avg_monthly_sales", "sales_per_month>_count").
    Do(ctx)

// 累计求和
cumulativeResp, _ := builder.NewAggregationBuilder(esClient, "sales").
    DateHistogram("daily_sales", "date", "1d").
    CumulativeSum("cumulative_sales", "daily_sales>total_amount").
    Do(ctx)
```

### 6. 集群管理 (ClusterBuilder)

```go
clusterBuilder := builder.NewClusterBuilder(esClient)

// 集群健康
health, _ := clusterBuilder.Health(ctx)
fmt.Printf("集群状态: %s\n", health.Status)
fmt.Printf("节点数: %d\n", health.NumberOfNodes)
fmt.Printf("活跃分片: %d\n", health.ActiveShards)

// 集群统计
stats, _ := clusterBuilder.Stats(ctx)
fmt.Printf("索引数量: %d\n", stats.Indices.Count)

// 节点信息
nodes, _ := clusterBuilder.NodesInfo(ctx)
fmt.Printf("节点数: %d\n", len(nodes.Nodes))

// 节点统计
nodeStats, _ := clusterBuilder.NodesStats(ctx)

// 获取任务
tasks, _ := clusterBuilder.Tasks(ctx)

// 集群设置
settings, _ := clusterBuilder.GetSettings(ctx)

// 更新集群设置
err := clusterBuilder.UpdateSettings(ctx,
    map[string]interface{}{
        "indices.recovery.max_bytes_per_sec": "50mb",
    }, nil)
```

## API 对比表

| 功能 | Elasticsearch REST API | github.com/Kirby980/go-es 链式调用 |
|------|------------------------|----------------|
| 创建索引 | PUT /index | NewIndexBuilder(client, "index").Shards(1).Do(ctx) |
| 索引文档 | PUT /index/_doc/1 | NewDocumentBuilder(client, "index").ID("1").Set("field", value).Do(ctx) |
| 搜索 | POST /index/_search | NewSearchBuilder(client, "index").Match("field", "value").Do(ctx) |
| 聚合 | POST /index/_search (with aggs) | NewAggregationBuilder(client, "index").Avg("name", "field").Do(ctx) |
| 批量操作 | POST /_bulk | NewBulkBuilder(client).Add(...).Update(...).Do(ctx) |

## 完整示例

查看 `examples/complete_api_test.go` 获取完整的使用示例。

```bash
# 运行完整示例
go test -v ./examples -run TestCompleteAPI
```

## 支持的功能

### IndexBuilder
- ✅ 创建索引 (Shards, Replicas, RefreshInterval)
- ✅ 字段映射 (AddProperty, WithAnalyzer, WithFormat)
- ✅ 别名管理 (AddAlias)
- ✅ 检查存在 (Exists)
- ✅ 获取索引信息 (Get)
- ✅ 删除索引 (Delete)

### DocumentBuilder
- ✅ 索引文档 (Do)
- ✅ 创建文档 (Create)
- ✅ 更新文档 (Update)
- ✅ 脚本更新 (Script)
- ✅ Upsert (Upsert)
- ✅ 获取文档 (Get)
- ✅ 删除文档 (Delete)
- ✅ 检查存在 (Exists)
- ✅ 批量获取 (MGet)

### SearchBuilder
- ✅ 全文搜索 (Match, MatchPhrase, MultiMatch)
- ✅ 精确查询 (Term, Terms, IDs)
- ✅ 范围查询 (Range)
- ✅ 模糊查询 (Fuzzy, Wildcard, Prefix, Regexp)
- ✅ 查询字符串 (QueryString)
- ✅ 布尔查询 (Must, Should, MustNot, Filter)
- ✅ 地理查询 (GeoDistance, GeoBoundingBox)
- ✅ 嵌套查询 (Nested)
- ✅ 排序 (Sort)
- ✅ 分页 (From, Size)
- ✅ 高亮 (Highlight)
- ✅ 字段过滤 (Source)

### AggregationBuilder
- ✅ 指标聚合 (Avg, Sum, Min, Max, Count, Stats, Cardinality, Percentiles)
- ✅ 桶聚合 (Terms, Histogram, DateHistogram, Range, DateRange)
- ✅ 过滤器聚合 (Filter, Filters)
- ✅ 管道聚合 (AvgBucket, SumBucket, MovingAvg, Derivative, CumulativeSum)
- ✅ 地理聚合 (GeoBounds, GeoCentroid, GeoDistance)

### BulkBuilder
- ✅ 批量索引 (Add)
- ✅ 批量创建 (Create)
- ✅ 批量更新 (Update)
- ✅ 批量删除 (Delete)
- ✅ 错误处理 (HasErrors, FailedItems, SuccessCount)

### ClusterBuilder
- ✅ 集群健康 (Health)
- ✅ 集群状态 (State)
- ✅ 集群统计 (Stats)
- ✅ 节点信息 (NodesInfo, NodesStats)
- ✅ 任务管理 (Tasks)
- ✅ 集群设置 (GetSettings, UpdateSettings)
- ✅ 分配解释 (AllocationExplain)

## 配置选项

```go
// 创建客户端时的配置选项
esClient, err := client.New(
    config.WithAddresses("https://localhost:9200"),      // ES 地址
    config.WithAuth("username", "password"),             // 认证
    config.WithTransport(true),                          // 跳过 SSL 验证
    config.WithTimeout(30*time.Second),                  // 超时时间
    config.WithRetry(3, time.Second),                    // 重试配置
    config.WithDebug(true),                              // 调试模式
)
```
