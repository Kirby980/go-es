# 错误处理

go-es 提供了完善的错误处理机制，可以方便地判断和处理各种 Elasticsearch 错误。

## 错误类型

所有 Elasticsearch 请求错误都会被解析为 `errors.ESError` 类型，包含以下信息：

```go
type ESError struct {
    StatusCode int                      // HTTP状态码
    Type       string                   // 错误类型
    Reason     string                   // 错误原因
    RootCause  []map[string]interface{} // 根本原因
    RawBody    []byte                   // 原始响应体
}
```

## 错误判断方法

提供了便捷的方法来判断常见的错误类型：

```go
import (
    "github.com/Kirby980/go-es/errors"
    "github.com/Kirby980/go-es/builder"
)

resp, err := builder.NewDocumentBuilder(esClient, "products").
    ID("non-existent-id").
    Get(ctx)

if err != nil {
    // 类型断言为 ESError
    if esErr, ok := err.(*errors.ESError); ok {
        // 判断是否为 404 错误
        if esErr.IsNotFound() {
            fmt.Println("文档不存在")
        }

        // 判断是否为冲突错误（版本冲突）
        if esErr.IsConflict() {
            fmt.Println("文档版本冲突")
        }

        // 判断是否为请求错误
        if esErr.IsBadRequest() {
            fmt.Println("请求参数错误")
        }

        // 判断是否为超时错误
        if esErr.IsTimeout() {
            fmt.Println("请求超时")
        }

        // 获取详细错误信息
        fmt.Printf("状态码: %d\n", esErr.StatusCode)
        fmt.Printf("错误类型: %s\n", esErr.Type)
        fmt.Printf("错误原因: %s\n", esErr.Reason)
    }
}
```

## 常见错误处理示例

### 示例1：处理文档不存在的情况

```go
resp, err := builder.NewDocumentBuilder(esClient, "products").
    ID("123").
    Get(ctx)

if err != nil {
    if esErr, ok := err.(*errors.ESError); ok && esErr.IsNotFound() {
        // 文档不存在，创建新文档
        _, err = builder.NewDocumentBuilder(esClient, "products").
            ID("123").
            Set("name", "新商品").
            Set("price", 99.99).
            Do(ctx)
    } else {
        // 其他错误
        return err
    }
}
```

### 示例2：处理索引已存在的冲突

```go
err := builder.NewIndexBuilder(esClient, "products").
    Shards(1).
    Replicas(1).
    Create(ctx)

if err != nil {
    if esErr, ok := err.(*errors.ESError); ok && esErr.IsBadRequest() {
        if esErr.Type == "resource_already_exists_exception" {
            fmt.Println("索引已存在，跳过创建")
        }
    } else {
        return err
    }
}
```

### 示例3：处理搜索请求错误

```go
resp, err := builder.NewSearchBuilder(esClient, "products").
    Match("name", "iPhone").
    Do(ctx)

if err != nil {
    if esErr, ok := err.(*errors.ESError); ok {
        switch {
        case esErr.IsBadRequest():
            fmt.Printf("搜索语法错误: %s\n", esErr.Reason)
        case esErr.IsTimeout():
            fmt.Println("搜索超时，请优化查询条件")
        case esErr.IsNotFound():
            fmt.Println("索引不存在")
        default:
            fmt.Printf("搜索失败: %s\n", esErr.Error())
        }
    }
}
```

### 示例4：处理批量操作错误

```go
bulkResp, err := builder.NewBulkBuilder(esClient).
    Index("products").
    Add("", "1", map[string]interface{}{"name": "商品1"}).
    Add("", "2", map[string]interface{}{"name": "商品2"}).
    Do(ctx)

if err != nil {
    if esErr, ok := err.(*errors.ESError); ok {
        fmt.Printf("批量操作失败: %s\n", esErr.Reason)
    }
}

// 检查部分失败的项
if bulkResp != nil && bulkResp.HasErrors() {
    for _, item := range bulkResp.FailedItems() {
        fmt.Printf("文档 %s 操作失败: %s\n", item.ID, item.Error.Reason)
    }
}
```

## 错误码说明

| HTTP 状态码 | 错误类型 | 判断方法 | 说明 |
|-----------|---------|---------|------|
| 400 | Bad Request | `IsBadRequest()` | 请求参数错误、语法错误 |
| 404 | Not Found | `IsNotFound()` | 索引或文档不存在 |
| 408 | Timeout | `IsTimeout()` | 请求超时 |
| 409 | Conflict | `IsConflict()` | 版本冲突、资源冲突 |
| 429 | Too Many Requests | - | 请求过于频繁 |
| 500+ | Server Error | - | 服务器内部错误 |

## 最佳实践

1. **总是进行错误检查**：所有 Elasticsearch 操作都应该检查错误
2. **使用类型断言**：通过类型断言为 `*errors.ESError` 获取详细错误信息
3. **区分错误类型**：根据不同的错误类型采取不同的处理策略
4. **记录详细信息**：在日志中记录 `Type` 和 `Reason` 以便排查问题
5. **优雅降级**：对于非致命错误（如文档不存在），提供合理的默认行为
