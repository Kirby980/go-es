# 高级功能

## 批量操作 (BulkBuilder)

批量操作可以大幅提高性能，一次请求执行多个操作。

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

### 支持的功能

- ✅ 批量索引 (Add)
- ✅ 批量创建 (Create)
- ✅ 批量更新 (Update)
- ✅ 批量删除 (Delete)
- ✅ 错误处理 (HasErrors, FailedItems, SuccessCount)

## 按条件批量更新 (UpdateByQueryBuilder)

```go
// 按条件批量更新文档
resp, err := builder.NewUpdateByQueryBuilder(esClient, "products").
    Term("status", "pending").
    Set("status", "processed").
    Set("updated_at", time.Now().Unix()).
    Do(ctx)

fmt.Printf("更新了 %d 个文档\n", resp.Updated)

// 使用脚本更新
resp, err := builder.NewUpdateByQueryBuilder(esClient, "products").
    Range("price", nil, 100).
    Script("ctx._source.discount = ctx._source.price * 0.9", nil).
    Do(ctx)
```

### 支持的功能

- ✅ 按条件批量更新 (Term, Range, Match查询)
- ✅ 脚本更新 (Script)
- ✅ 简化字段更新 (Set)

## 按条件批量删除 (DeleteByQueryBuilder)

```go
// 按条件批量删除文档
resp, err := builder.NewDeleteByQueryBuilder(esClient, "products").
    Term("status", "expired").
    Range("created_at", nil, "2020-01-01").
    Do(ctx)

fmt.Printf("删除了 %d 个文档\n", resp.Deleted)

// 删除特定分类的商品
resp, err := builder.NewDeleteByQueryBuilder(esClient, "products").
    Terms("category", "discontinued", "obsolete").
    Do(ctx)
```

### 支持的功能

- ✅ 按条件批量删除 (Term, Range, Match查询)
- ✅ 安全检查 (必须提供查询条件)

## 深度分页遍历 (ScrollBuilder)

Scroll 适合大数据集的顺序遍历和导出。

```go
// 创建scroll查询
scroll := builder.NewScrollBuilder(esClient, "products").
    Match("status", "active").
    Size(1000).
    KeepAlive("5m")

// 第一次查询
resp, err := scroll.Do(ctx)
if err != nil {
    log.Fatal(err)
}

// 处理第一批数据
for _, hit := range resp.Hits.Hits {
    fmt.Printf("处理文档: %s\n", hit.ID)
}

// 持续获取下一批数据
for scroll.HasMore(resp) {
    resp, err = scroll.Next(ctx)
    if err != nil {
        break
    }

    for _, hit := range resp.Hits.Hits {
        fmt.Printf("处理文档: %s\n", hit.ID)
    }
}

// 清理scroll上下文
scroll.Clear(ctx)
```

### 支持的功能

- ✅ 深度分页遍历 (Do, Next)
- ✅ 游标管理 (KeepAlive, Clear)
- ✅ 批量处理 (Size, HasMore)

## 高效深度分页 (SearchAfterBuilder)

Search After 是无状态的深度分页方案，适合实时分页场景。

### Search After vs Scroll 对比

| 特性 | Search After | Scroll |
|------|-------------|--------|
| **性能** | 更轻量，无状态 | 需要维护 scroll context |
| **适用场景** | 实时分页、API分页 | 大数据集顺序遍历、导出 |
| **资源占用** | 低（无需保存上下文） | 高（需要保存快照） |
| **实时性** | 能看到最新数据 | 看不到查询后的新数据 |
| **使用限制** | 必须有排序字段 | 不需要排序 |

```go
// 基础用法（自动翻页）
searchAfter := builder.NewSearchAfterBuilder(esClient, "products").
    Match("status", "active").
    Sort("price", "asc").      // 主排序字段
    Sort("_id", "asc").        // tie-breaker（必须）
    Size(20)

// 第一页
resp, err := searchAfter.Do(ctx)
fmt.Printf("第一页: %d 条\n", len(resp.Hits.Hits))

// 第二页（自动使用上一页的最后一个文档的 sort 值）
resp, err = searchAfter.Next(ctx)
fmt.Printf("第二页: %d 条\n", len(resp.Hits.Hits))

// 持续翻页
for searchAfter.HasMore(resp) {
    resp, err = searchAfter.Next(ctx)
    if err != nil {
        break
    }
    for _, hit := range resp.Hits.Hits {
        fmt.Printf("处理文档: %s\n", hit.ID)
    }
}

// 手动指定 search_after 值（适合 API 分页）
lastSort := searchAfter.GetLastSortValues(resp)

// 下次请求时
resp, _ = builder.NewSearchAfterBuilder(esClient, "products").
    Sort("price", "asc").
    Sort("_id", "asc").
    SearchAfter(lastSort...).  // 手动设置上一页的 sort 值
    Size(20).
    Do(ctx)
```

**重要提示：**
- Search After **必须指定排序字段**，建议最后加上 `_id` 作为 tie-breaker
- 适合实时 API 分页，客户端保存上一页的 `sort` 值即可
- 无需清理上下文，比 Scroll 更轻量

### 支持的功能

- ✅ 高效深度分页 (Do, Next)
- ✅ 多字段排序 (Sort, SortBy)
- ✅ 无状态分页 (SearchAfter, GetLastSortValues)
- ✅ 查询条件 (Match, Term, Range, Terms, Exists)
- ✅ 布尔查询 (Should, MustNot, MinimumShouldMatch)
- ✅ 字段过滤 (Source, Highlight, MinScore)
- ✅ 自动/手动翻页 (HasMore)

## Debug调试模式

类似 GORM 的 Debug 模式，局部控制日志输出。

```go
// 启用Debug模式查看请求和响应
resp, err := builder.NewSearchBuilder(esClient, "products").
    Debug().  // 启用调试，会打印请求和响应JSON
    Match("name", "iPhone").
    Do(ctx)

// 不带Debug的查询（不会打印任何东西）
resp2, err := builder.NewSearchBuilder(esClient, "products").
    Match("name", "Samsung").
    Do(ctx)

// 所有Builder都支持Debug
builder.NewDocumentBuilder(esClient, "products").Debug().ID("1").Get(ctx)
builder.NewBulkBuilder(esClient).Debug().Add(...).Do(ctx)
builder.NewIndexBuilder(esClient, "index").Debug().Create(ctx)
builder.NewClusterBuilder(esClient).Debug().Health(ctx)
```

## 集群管理 (ClusterBuilder)

```go
clusterBuilder := builder.NewClusterBuilder(esClient)

// 集群健康
health, err := clusterBuilder.Health(ctx)
fmt.Printf("集群状态: %s\n", health.Status)
fmt.Printf("节点数: %d\n", health.NumberOfNodes)
fmt.Printf("活跃分片: %d\n", health.ActiveShards)

// 集群统计
stats, err := clusterBuilder.Stats(ctx)
fmt.Printf("索引数量: %d\n", stats.Indices.Count)

// 节点信息
nodes, err := clusterBuilder.NodesInfo(ctx)
fmt.Printf("节点数: %d\n", len(nodes.Nodes))

// 节点统计
nodeStats, err := clusterBuilder.NodesStats(ctx)

// 获取任务
tasks, err := clusterBuilder.Tasks(ctx)

// 集群设置
settings, err := clusterBuilder.GetSettings(ctx)

// 更新集群设置
err := clusterBuilder.UpdateSettings(ctx,
    map[string]interface{}{
        "indices.recovery.max_bytes_per_sec": "50mb",
    }, nil)
```

### 支持的功能

- ✅ 集群健康 (Health)
- ✅ 集群状态 (State)
- ✅ 集群统计 (Stats)
- ✅ 节点信息 (NodesInfo, NodesStats)
- ✅ 任务管理 (Tasks)
- ✅ 集群设置 (GetSettings, UpdateSettings)
- ✅ 分配解释 (AllocationExplain)
