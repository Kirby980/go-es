# 搜索 (SearchBuilder)

SearchBuilder 提供了强大的 Elasticsearch 搜索功能。

## 基础搜索

```go
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
```

## 查询类型

### 全文搜索

```go
// 单字段匹配
builder.NewSearchBuilder(esClient, "products").
    Match("name", "iPhone").
    Do(ctx)

// 多字段匹配
builder.NewSearchBuilder(esClient, "products").
    MultiMatch("Apple phone", "name", "description").
    Do(ctx)

// 短语匹配
builder.NewSearchBuilder(esClient, "products").
    MatchPhrase("name", "iPhone Pro").
    Do(ctx)
```

### 精确查询

```go
// 单值精确匹配
builder.NewSearchBuilder(esClient, "products").
    Term("category", "electronics").
    Do(ctx)

// 多值精确匹配
builder.NewSearchBuilder(esClient, "products").
    Terms("category", "electronics", "books").
    Do(ctx)

// ID 查询
builder.NewSearchBuilder(esClient, "products").
    IDs("1", "2", "3").
    Do(ctx)
```

### 范围查询

```go
// 价格范围
builder.NewSearchBuilder(esClient, "products").
    Range("price", 100, 1000).
    Do(ctx)

// 日期范围
builder.NewSearchBuilder(esClient, "orders").
    Range("created_at", "2024-01-01", "2024-12-31").
    Do(ctx)

// 开区间（大于）
builder.NewSearchBuilder(esClient, "products").
    Range("price", 100, nil).
    Do(ctx)
```

### 模糊搜索

```go
// 模糊匹配
builder.NewSearchBuilder(esClient, "products").
    Fuzzy("name", "iPhon", "AUTO").
    Do(ctx)

// 前缀搜索
builder.NewSearchBuilder(esClient, "products").
    Prefix("name", "iPa").
    Do(ctx)

// 通配符搜索
builder.NewSearchBuilder(esClient, "products").
    Wildcard("name", "i*one").
    Do(ctx)

// 正则表达式搜索
builder.NewSearchBuilder(esClient, "products").
    Regexp("name", "i.*one").
    Do(ctx)
```

### QueryString

```go
builder.NewSearchBuilder(esClient, "products").
    QueryString("(iPhone OR iPad) AND electronics", "name", "category").
    Do(ctx)
```

## 布尔查询

### 查询逻辑关系

| 方法类型 | 逻辑关系 | 说明 |
|---------|---------|------|
| `Match()`, `Term()`, `Range()` 等 | **AND（且）** | 所有条件都必须满足 |
| `MatchShould()`, `TermShould()`, `RangeShould()` | **OR（或）** | 至少满足一个条件 |
| `MatchMustNot()`, `TermMustNot()`, `RangeMustNot()` | **NOT（非）** | 必须不匹配 |
| `MinimumShouldMatch()` | **OR 数量控制** | 控制至少匹配几个 should 条件 |

### 复杂组合查询示例

```go
// 需求：搜索符合以下条件的商品
// 1. 必须是 "electronics" 分类 (AND)
// 2. 必须是 "active" 状态 (AND)
// 3. 价格在 100-1000 之间 (AND)
// 4. 品牌是 "Apple" 或 "Samsung" 或 "Huawei"（至少匹配2个）(OR)
// 5. 标题不能包含 "refurbished" (NOT)

resp, err := builder.NewSearchBuilder(esClient, "products").
    // ===== AND 条件（所有都要满足）=====
    Term("category", "electronics").        // AND: 必须是 electronics 分类
    Term("status", "active").               // AND: 必须是 active 状态
    Range("price", 100, 1000).              // AND: 必须在 100-1000 价格区间

    // ===== OR 条件（至少满足2个）=====
    MatchShould("brand", "Apple").          // OR: 可能是 Apple
    MatchShould("brand", "Samsung").        // OR: 可能是 Samsung
    MatchShould("brand", "Huawei").         // OR: 可能是 Huawei
    MinimumShouldMatch(2).                  // 上面3个OR条件至少要满足2个

    // ===== NOT 条件（必须不满足）=====
    MatchMustNot("title", "refurbished").   // NOT: 标题不能包含 refurbished

    From(0).
    Size(20).
    Do(ctx)
```

## 地理位置查询

```go
// 距离查询
builder.NewSearchBuilder(esClient, "stores").
    GeoDistance("location", 37.7749, -122.4194, "10km").
    Do(ctx)

// 边界框查询
builder.NewSearchBuilder(esClient, "stores").
    GeoBoundingBox("location", 40.73, -74.1, 40.01, -71.12).
    Do(ctx)
```

## 其他功能

### 排序

```go
builder.NewSearchBuilder(esClient, "products").
    Match("name", "phone").
    Sort("price", "asc").
    Sort("rating", "desc").
    Do(ctx)
```

### 分页

```go
builder.NewSearchBuilder(esClient, "products").
    Match("category", "electronics").
    From(0).
    Size(20).
    Do(ctx)
```

### 高亮

```go
builder.NewSearchBuilder(esClient, "products").
    Match("name", "iPhone").
    Highlight("name", "description").
    Do(ctx)
```

### 字段过滤

```go
builder.NewSearchBuilder(esClient, "products").
    Match("name", "iPhone").
    Source("name", "price", "category").
    Do(ctx)
```

### 最小评分过滤

```go
builder.NewSearchBuilder(esClient, "products").
    Match("name", "iPhone").
    MinScore(0.5).
    Do(ctx)
```

### 快速计数

```go
count, err := builder.NewSearchBuilder(esClient, "products").
    Match("status", "active").
    Count(ctx)
fmt.Printf("活跃商品数量: %d\n", count)
```

## 支持的功能

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
- ✅ 最小评分 (MinScore)
- ✅ 快速计数 (Count)
