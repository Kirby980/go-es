# go-es：一个像 GORM 一样优雅的 Elasticsearch Go 客户端

> 厌倦了手写 ES DSL JSON？试试这个链式调用风格的 Go ES 客户端。

---

## 背景：原生 ES SDK 的痛点

在 Go 项目中使用 Elasticsearch，官方 SDK 的写法大概是这样的：

```go
query := map[string]interface{}{
    "query": map[string]interface{}{
        "bool": map[string]interface{}{
            "must": []interface{}{
                map[string]interface{}{
                    "match": map[string]interface{}{
                        "title": "iPhone",
                    },
                },
            },
            "filter": []interface{}{
                map[string]interface{}{
                    "term": map[string]interface{}{
                        "category": "electronics",
                    },
                },
                map[string]interface{}{
                    "range": map[string]interface{}{
                        "price": map[string]interface{}{
                            "gte": 500,
                            "lte": 2000,
                        },
                    },
                },
            },
        },
    },
    "from": 0,
    "size": 10,
    "sort": []interface{}{
        map[string]interface{}{
            "price": map[string]interface{}{"order": "desc"},
        },
    },
}
```

写完这坨 JSON 套娃，你还要手动序列化、构造 HTTP 请求、解析响应……

用 GORM 的人写过 `db.Where("name = ?", "iPhone").Find(&products)` 后，再回来写这个，心里是什么感受，懂的都懂。

**go-es** 就是为了解决这个问题而生的。

---

## go-es 是什么？

`go-es` 是一个轻量级、零第三方 ES SDK 依赖的 Elasticsearch Go 客户端，设计上深度参考 GORM 的链式调用风格，让你用写 ORM 查询的体验来操作 ES。

**同样的查询，用 go-es 写：**

```go
resp, err := builder.NewSearchBuilder(esClient, "products").
    Match("title", "iPhone").
    Term("category", "electronics").
    Range("price", 500, 2000).
    Sort("price", "desc").
    From(0).
    Size(10).
    Do(ctx)
```

**核心特性：**

- **链式调用**：所有操作都支持链式调用，代码可读性极强
- **Builder 模式**：每种操作对应独立的 Builder，职责清晰
- **泛型布尔查询**：`BoolQuery[T]` 泛型设计，must/should/must_not/filter 随意组合
- **GORM 风格 Sugar API**：AutoMigrate、Create、Update、Find 一键操作
- **AutoMigrate**：通过结构体 tag 自动创建/更新索引，再也不用手写 mapping
- **Debug 模式**：一个 `.Debug()` 打印完整请求和响应，调试超方便
- **内置连接池 + 地址轮询**：生产级连接管理，开箱即用
- **完善的错误处理**：ES 错误类型化，`IsNotFound()`、`IsConflict()` 直接判断

---

## 快速上手

### 安装

```bash
go get github.com/Kirby980/go-es
```

Go 版本要求：1.21+

### 创建客户端

```go
import (
    "github.com/Kirby980/go-es/client"
    "github.com/Kirby980/go-es/config"
)

esClient, err := client.New(
    config.WithAddresses("https://localhost:9200"),
    config.WithAuth("elastic", "your-password"),
    config.WithInsecureSkipVerify(true),
    config.WithTimeout(10 * time.Second),
    // 连接池配置
    config.WithMaxConnsPerHost(100),
    config.WithMaxIdleConns(200),
    config.WithMaxIdleConnsPerHost(50),
    config.WithIdleConnTimeout(90 * time.Second),
)
if err != nil {
    panic(err)
}
defer esClient.Close()
```

---

## 核心功能详解

### 1. AutoMigrate —— 告别手写 Mapping

定义好结构体，一行代码搞定索引创建和字段更新：

```go
import (
    "github.com/Kirby980/go-es/sugar"
    esconst "github.com/Kirby980/go-es/const"
)

type Article struct {
    Title     string    `json:"title"      es:"type:text;analyzer:ik_max_word"`
    Summary   string    `json:"summary"    es:"type:text;analyzer:ik_smart"`
    Author    string    `json:"author"     es:"type:keyword"`
    Tags      []string  `json:"tags"       es:"type:keyword"`
    ViewCount int       `json:"view_count" es:"type:integer"`
    Price     float64   `json:"price"      es:"type:float"`
    Published bool      `json:"published"  es:"type:boolean"`
    CreatedAt string    `json:"created_at" es:"type:date;format:yyyy-MM-dd HH:mm:ss"`
    Location  string    `json:"location"   es:"type:geo_point"`
}

// 自定义索引名（可选）
func (a *Article) IndexName() string {
    return "articles"
}

s := sugar.New(esClient)

// 索引不存在则创建，已存在则追加新字段（ES 不支持删除字段）
err := s.AutoMigrate(&Article{})

// 同时迁移多个模型
err = s.AutoMigrate(&Article{}, &User{}, &Product{})
```

### 2. 文档 CRUD —— Sugar API

```go
s := sugar.New(esClient)

article := &Article{
    Title:     "Go 1.24 新特性详解",
    Author:    "张三",
    ViewCount: 0,
    Published: true,
    CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
}

// 创建文档（自动生成 ID）
resp, err := s.Create(ctx, article)
fmt.Println(resp.ID) // ES 自动生成的 ID

// 创建文档（指定 ID）
resp, err = s.CreateWithID(ctx, "article-001", article)

// 更新文档
article.ViewCount = 100
_, err = s.Update(ctx, "article-001", article)

// Upsert（存在则更新，不存在则创建）
_, err = s.Upsert(ctx, "article-001", article)

// 获取文档
getResp, err := s.Get(ctx, "articles", "article-001")
if getResp.Found {
    fmt.Println(getResp.Source)
}

// 删除文档
_, err = s.Delete(ctx, "articles", "article-001")
```

### 3. 搜索 —— 链式查询的魅力

#### 基础布尔查询

```go
resp, err := builder.NewSearchBuilder(esClient, "articles").
    Match("title", "Go 语言").          // must: 全文搜索
    Term("author", "张三").             // filter: 精确匹配
    Range("view_count", 100, nil).      // filter: 范围查询
    MatchShould("tags", "golang").      // should: OR 条件
    MatchShould("tags", "go").
    MinimumShouldMatch(1).             // 至少匹配 1 个 should
    Sort("view_count", "desc").
    From(0).Size(10).
    Do(ctx)

fmt.Printf("共 %d 篇文章\n", resp.Hits.Total.Value)

// 结果扫描到结构体切片，就像 GORM 的 Find
var articles []Article
resp.Scan(&articles)
for _, a := range articles {
    fmt.Printf("[%s] %s (%d 阅读)\n", a.Author, a.Title, a.ViewCount)
}
```

#### 更多查询类型

```go
builder.NewSearchBuilder(esClient, "articles").
    Wildcard("title", "Go*").           // 通配符
    Prefix("title", "Go ").            // 前缀
    Fuzzy("title", "Golanng", 1).      // 模糊（容错1个字符）
    MultiMatch("Go 语言", "title", "summary").  // 多字段匹配
    IDs("id-1", "id-2", "id-3").       // 按 ID 查询
    Exists("cover_image").             // 字段存在查询
    GeoDistance("location", 39.9, 116.4, "10km"). // 地理距离
    Highlight("title", "summary").     // 高亮关键词
    Source("title", "author").         // 只返回指定字段
    Count(ctx)                         // 只计数，不返回文档
```

#### Sugar Find 快捷搜索

```go
var results []Article
resp, _ := sugar.New(esClient).Find("articles").
    Match("title", "Go 语言").
    Term("published", true).
    Size(20).
    Do(ctx)
resp.Scan(&results)
```

### 4. 聚合分析

```go
aggResp, err := builder.NewAggregationBuilder(esClient, "articles").
    // 指标聚合
    Avg("avg_views", "view_count").
    Max("max_views", "view_count").
    Cardinality("unique_authors", "author").
    // 桶聚合
    Terms("by_author", "author", 10).
    DateHistogram("posts_by_day", "created_at", "1d").
    Do(ctx)

// 读取结果
avgViews := aggResp.Aggregations["avg_views"].(map[string]any)["value"]
fmt.Printf("平均阅读量: %.0f\n", avgViews)
```

### 5. 批量操作 —— BulkBuilder

适合数据导入、批量更新等场景，支持自动分批提交：

```go
bulkResp, err := builder.NewBulkBuilder(esClient).
    Index("articles").
    AutoFlushSize(500).  // 每 500 条自动提交
    OnFlush(func(resp *builder.BulkResponse) {
        fmt.Printf("已提交 %d 条，成功 %d 条\n",
            len(resp.Items), resp.SuccessCount())
    }).
    Add("", "1", map[string]any{"title": "文章一", "author": "张三"}).
    Add("", "2", map[string]any{"title": "文章二", "author": "李四"}).
    Update("", "3", map[string]any{"view_count": 9999}).
    Delete("", "4").
    Do(ctx)

// 链式 API（对每条数据设置多个字段）
builder.NewBulkBuilder(esClient).
    Index("articles").
    AddDoc("doc-001").
        Set("title", "新文章").
        Set("author", "王五").
        Set("view_count", 0).
    AddDoc("doc-002").
        Set("title", "另一篇").
        Set("author", "赵六").
    Do(ctx)
```

### 6. 深度分页

#### Scroll —— 大数据全量导出

```go
scroll := builder.NewScrollBuilder(esClient, "articles").
    Term("published", true).
    Size(1000).
    KeepAlive("5m")

resp, _ := scroll.Do(ctx)
total := 0
for scroll.HasMore(resp) {
    for _, hit := range resp.Hits.Hits {
        // 处理每条数据
        total++
    }
    resp, _ = scroll.Next(ctx)
}
scroll.Clear(ctx)
fmt.Printf("共导出 %d 条\n", total)
```

#### Search After —— 实时深度翻页

```go
sa := builder.NewSearchAfterBuilder(esClient, "articles").
    Term("published", true).
    Sort("view_count", "desc").
    Sort("_id", "asc").  // 需要唯一排序字段做 tiebreaker
    Size(20)

// 第一页
resp, _ := sa.Do(ctx)

// 后续页（自动携带上页最后一条的 sort 值）
resp, _ = sa.Next(ctx)
resp, _ = sa.Next(ctx)
```

### 7. 按条件批量更新/删除

```go
// 将所有 pending 状态的文章标记为 published
resp, err := builder.NewUpdateByQueryBuilder(esClient, "articles").
    Term("status", "pending").
    Set("status", "published").
    Set("updated_at", time.Now().Unix()).
    Do(ctx)

fmt.Printf("更新了 %d 篇文章\n", resp.Updated)

// 删除三个月前的草稿
resp2, err := builder.NewDeleteByQueryBuilder(esClient, "articles").
    Term("status", "draft").
    Range("created_at", nil, "now-90d/d").
    Do(ctx)

fmt.Printf("删除了 %d 篇文章\n", resp2.Deleted)
```

### 8. 索引管理

#### 手动创建索引（完整控制）

```go
import (
    "github.com/Kirby980/go-es/builder"
    esconst "github.com/Kirby980/go-es/const"
)

err := builder.NewIndexBuilder(esClient, "articles").
    Shards(3).
    Replicas(1).
    RefreshInterval("1s").
    // 自定义分析器：IK 保留大小写
    AddTokenizer("ik_case_sensitive",
        builder.WithTokenizerType(esconst.TokenizerIKSmart),
        builder.WithEnableLowercase(false),
    ).
    AddAnalyzer("ik_no_lowercase",
        builder.WithAnalyzerType(esconst.AnalyzerTypeCustom),
        builder.WithTokenizer("ik_case_sensitive"),
    ).
    // 字段映射
    AddProperty("title", esconst.FieldTypeText,
        builder.WithAnalyzer("ik_no_lowercase"),
        builder.WithSubField("keyword", esconst.FieldTypeKeyword,
            builder.WithIgnoreAbove(256)),
    ).
    AddProperty("author", esconst.FieldTypeKeyword).
    AddProperty("view_count", esconst.FieldTypeLong).
    AddProperty("created_at", esconst.FieldTypeDate,
        builder.WithFormat("yyyy-MM-dd HH:mm:ss")).
    AddAlias("articles-alias", nil).
    Create(ctx)
```

#### 在线更新 Mapping（添加字段）

```go
err := builder.NewIndexBuilder(esClient, "articles").
    AddProperty("cover_image", esconst.FieldTypeKeyword).
    AddProperty("word_count", esconst.FieldTypeInteger).
    PutMapping(ctx)
```

### 9. Debug 模式 —— 开发调试利器

只需在链式调用中加一个 `.Debug()`，会自动打印完整的请求 DSL 和响应体：

```go
resp, err := builder.NewSearchBuilder(esClient, "articles").
    Debug().   // 👈 加这一行
    Match("title", "Go 语言").
    Term("published", true).
    Do(ctx)
```

输出（使用 zap 结构化日志）：

```
[ES Debug] 请求  method=POST  path=/articles/_search
  body={"from":0,"query":{"bool":{"filter":[{"term":{"published":true}}],"must":[{"match":{"title":"Go 语言"}}]}},"size":10}

[ES Debug] 响应
  body={"took":3,"hits":{"total":{"value":42,...},...}}
```

所有 Builder 都支持 Debug 模式，且调用一次后自动重置，不会影响后续请求。

### 10. 错误处理

ES 返回的 HTTP 错误会被封装为有类型的 `ESError`：

```go
import "github.com/Kirby980/go-es/errors"

getResp, err := builder.NewDocumentBuilder(esClient, "articles").
    ID("not-exist-id").
    Get(ctx)

if err != nil {
    var esErr *errors.ESError
    if errors.As(err, &esErr) {
        switch {
        case esErr.IsNotFound():
            fmt.Println("文章不存在")
        case esErr.IsConflict():
            fmt.Println("版本冲突，请重试")
        case esErr.IsBadRequest():
            fmt.Printf("请求错误: %s\n", esErr.Reason)
        case esErr.IsTimeout():
            fmt.Println("请求超时")
        }
    }
}
```

---

## 项目架构设计

```
go-es/
├── client/       # HTTP 客户端，连接池，地址轮询
├── config/       # 函数选项模式配置
├── builder/      # 核心：所有 Builder 实现
│   ├── search.go          # 搜索
│   ├── document.go        # 文档 CRUD
│   ├── index.go           # 索引管理
│   ├── bulk.go            # 批量操作
│   ├── aggregation.go     # 聚合分析
│   ├── scroll.go          # Scroll 分页
│   ├── search_after.go    # SearchAfter 分页
│   ├── cluster.go         # 集群管理
│   ├── delete_by_query.go # 按条件删除
│   ├── update_by_query.go # 按条件更新
│   └── query.go           # 泛型 BoolQuery[T]
├── sugar/        # 语法糖：GORM 风格快捷 API
├── const/        # 字段类型、分析器常量
├── errors/       # ES 错误类型
└── logger/       # 日志接口（支持自定义）
```

几个设计亮点：

**1. 泛型 BoolQuery[T]**

所有支持查询的 Builder 都通过嵌入 `BoolQuery[T]` 获得布尔查询能力，复用代码的同时保持链式调用的类型安全：

```go
type SearchBuilder struct {
    BoolQuery[SearchBuilder]  // 泛型嵌入，方法返回 *SearchBuilder
    // ...
}
```

**2. 接口隔离避免循环依赖**

`builder` 包不直接依赖 `client` 包，而是通过 `ESClient` 接口解耦：

```go
type ESClient interface {
    Do(ctx context.Context, method, path string, body any) ([]byte, error)
    GetAddress() string
    DoRequest(ctx context.Context, req *http.Request) ([]byte, error)
    GetLogger() logger.Logger
}
```

**3. 自定义日志**

默认使用 `zap` 生产级 JSON 日志，也可以注入自定义实现：

```go
type MyLogger struct{}
func (l *MyLogger) Info(msg string, kv ...any)  { /* 你的实现 */ }
func (l *MyLogger) Error(msg string, kv ...any) { /* 你的实现 */ }
// ...

esClient, _ := client.New(
    config.WithLogger(&MyLogger{}),
)
```

---

## 与原生写法对比

| 场景 | 原生写法 | go-es |
|------|---------|-------|
| 布尔查询 | 嵌套 `map[string]interface{}` 套娃 | 链式 `.Match().Term().Range()` |
| 创建文档 | 手动序列化 + 构造 HTTP 请求 | `s.Create(ctx, &struct{})` |
| 自动建索引 | 手写 JSON mapping + PUT 请求 | `s.AutoMigrate(&struct{})` |
| 聚合 | 多层嵌套 map | `.Avg().Terms().DateHistogram()` |
| 批量写入 | 手动拼 NDJSON 格式 | `.Add().Update().Delete().Do()` |
| 深度分页 | 手动维护 scroll_id | `.Next()` 自动管理 |
| 调试 DSL | 打 log 手动格式化 | `.Debug()` 一行搞定 |
| 错误判断 | 解析 status code | `esErr.IsNotFound()` |

---

## 当前支持的 ES 功能

- ✅ 索引管理（Create / Delete / Exists / Get / UpdateSettings / PutMapping）
- ✅ AutoMigrate（结构体 tag 自动迁移）
- ✅ 自定义分析器（Analyzer / Tokenizer / CharFilter / TokenFilter）
- ✅ 文档 CRUD（Create / Update / Upsert / Get / Delete / Exists）
- ✅ 脚本更新（Painless script）
- ✅ 批量获取（MGet）
- ✅ 全文搜索（Match / MatchPhrase / MultiMatch / QueryString）
- ✅ 精确查询（Term / Terms / IDs）
- ✅ 范围查询（Range）
- ✅ 模糊查询（Fuzzy / Wildcard / Prefix / Regexp）
- ✅ 布尔组合（Must / Should / MustNot / Filter / MinimumShouldMatch）
- ✅ 地理查询（GeoDistance / GeoBoundingBox）
- ✅ 嵌套查询（Nested）
- ✅ 高亮 / 排序 / 分页 / 字段过滤
- ✅ 聚合：指标（Avg/Sum/Min/Max/Stats/Cardinality/Percentiles）
- ✅ 聚合：桶（Terms/Histogram/DateHistogram/Range/Filter/Missing）
- ✅ 聚合：管道（AvgBucket/MovingAvg/Derivative/CumulativeSum）
- ✅ 地理聚合（GeoBounds/GeoCentroid/GeoDistance）
- ✅ 子聚合（SubAgg）
- ✅ 批量操作（BulkBuilder，支持自动分批 + 回调）
- ✅ Scroll 深度分页
- ✅ SearchAfter 高效深度分页
- ✅ UpdateByQuery / DeleteByQuery
- ✅ 集群管理（Health / State / Stats / Nodes / Settings）

---

## 总结

`go-es` 的目标不是替代官方 SDK，而是在业务开发中提供更友好的抽象层：**能链式调用的，绝不手写 JSON；能自动推断的，绝不重复配置**。

如果你在 Go 项目中用 ES，但又厌倦了繁琐的 map 套 map，不妨试试。

**项目地址：** `github.com/Kirby980/go-es`

```bash
go get github.com/Kirby980/go-es
```

欢迎 Star、Issue、PR！

---

*如有问题或建议，欢迎在评论区交流。*
