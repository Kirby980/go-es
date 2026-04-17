# go-es 优化与路线图

本文档用于记录当前仓库的优化状态与后续可演进方向，避免“建议”与主分支实现不一致。

## 已完成

- 重试时请求体重置与超时控制
- 多地址轮询与基础故障转移
- Bulk 构建器的错误传播与内存复用（Buffer Pool）
- 高频查询路径的内存分配优化（Map Pool）
- Builder 侧逐步迁移为流式解码以降低大响应内存峰值
- Debug 行为统一为“默认单次，支持持久 Debug”

## 待优化（建议优先级从高到低）

- 节点嗅探（Sniffer）：定时从 `/_nodes/http` 动态更新节点列表，以适配集群扩缩容与节点漂移
- JSON 性能：在不破坏兼容性的前提下，评估可选高性能 JSON 引擎或为热路径提供可插拔编码器
- DSL 类型安全：为常用 Query/Agg 提供结构化 Query Node，减少 `map[string]any` 误用与运行时错误
- 大结果 Scan：为 hit `_source` 提供 `json.RawMessage`/直解模式，进一步降低二次序列化开销

func (q *queryConditions) buildBoolQuery() map[string]any {
    if len(q.must) == 0 && len(q.filters) == 0 && len(q.should) == 0 && len(q.mustNot) == 0 {
        return nil
    }
    boolQuery := make(map[string]any)
    if len(q.must) > 0 { boolQuery["must"] = q.must }
    if len(q.filters) > 0 { boolQuery["filter"] = q.filters }
    if len(q.should) > 0 { boolQuery["should"] = q.should }
    if len(q.mustNot) > 0 { boolQuery["must_not"] = q.mustNot }
    if q.minimumShouldMatch != nil { boolQuery["minimum_should_match"] = q.minimumShouldMatch }
    return map[string]any{"bool": boolQuery}
}
```

同理，`Match()`、`Term()`、`Terms()`、`Range()` 等方法也可以复用。

### 9. 重试策略缺乏灵活性

**文件**: `client/client.go:77-85`

当前重试采用固定间隔，且仅在网络错误时重试，不处理 5xx 服务端错误：

- 不支持指数退避（Exponential Backoff）
- 不重试 5xx 等可恢复的服务端错误
- 没有可配置的重试条件

**建议**:

```go
type RetryConfig struct {
    MaxRetries     int
    InitialBackoff time.Duration
    MaxBackoff     time.Duration
    BackoffFactor  float64
    RetryableStatus []int // 如 []int{502, 503, 504, 429}
}

func (c *Client) shouldRetry(statusCode int, err error) bool {
    if err != nil { return true }
    for _, code := range c.config.RetryableStatus {
        if statusCode == code { return true }
    }
    return false
}
```

### 10. Debug 输出硬编码 fmt.Printf

**文件**: `builder/debug.go:29-43`

Debug 信息直接使用 `fmt.Printf` 输出，无法自定义日志目标：

**建议**: 引入 Logger 接口，允许用户注入自定义日志实现：

```go
type Logger interface {
    Printf(format string, args ...any)
}

// 在 Config 中添加
type Config struct {
    // ...
    Logger Logger
}
```

### 11. `Close()` 是空操作

**文件**: `client/client.go:55-57`

`Close()` 不做任何清理，未关闭底层 HTTP Transport 的空闲连接：

```go
func (c *Client) Close() error {
    return nil // 什么都没做
}
```

**建议**:

```go
func (c *Client) Close() error {
    if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
        transport.CloseIdleConnections()
    }
    return nil
}
```

---

## 三、健壮性优化

### 12. URL 路径未转义

**文件**: 多个 Builder 的路径构建

索引名和文档 ID 直接拼接到 URL 中，若包含特殊字符（如 `/`、`空格`、`#`）会导致请求失败或安全问题：

```go
path := fmt.Sprintf("/%s/_doc/%s", b.index, b.id) // 未转义
```

**建议**: 使用 `url.PathEscape()` 转义：

```go
path := fmt.Sprintf("/%s/_doc/%s", url.PathEscape(b.index), url.PathEscape(b.id))
```

### 13. `isFlush` 静默吞掉错误并使用 `context.TODO()`

**文件**: `builder/bulk.go:92-99`

```go
func (b *BulkBuilder) isFlush() {
    if b.autoFlushSize > 0 && len(b.operations) >= b.autoFlushSize {
        resp, err := b.flush(context.TODO()) // 无法传递上下文
        if err == nil && b.onFlush != nil {   // 错误被忽略
            b.onFlush(resp)
        }
    }
}
```

**建议**: 将 context 作为参数传入，并将 flush 错误记录到 Builder 的 err 字段中。同时，对外暴露的 `Add()`、`AddDoc()` 等方法也应接受 context 参数。

### 14. Ping 方法不使用重试逻辑

**文件**: `client/client.go:110-131`

`Ping()` 是连接测试的关键方法，但不使用与 `Do()` 相同的重试逻辑：

**建议**: 复用 `Do()` 方法，或将重试逻辑提取为内部公共方法。

### 15. Debug 模式下 `defer SetDebug(false)` 导致一次性行为

**文件**: 多个 Builder 的 Do() 方法

每次 `Do()` 执行后 Debug 模式自动关闭，如果用户期望多次调用都启用 Debug，需要每次重新设置。这一行为虽非 bug，但与直觉不符。

**建议**: 移除 `defer b.SetDebug(false)`，让 Debug 状态持久化，由用户显式控制。或者提供配置项控制此行为。

### 16. ScrollBuilder `Terms` 方法缺失

**文件**: `builder/scroll.go`

`ScrollBuilder` 提供了 `Match`、`Term`、`Range` 方法，但缺少 `Terms` 方法，与其他 Builder 的 API 不一致。

---

## 四、命名与 API 规范

### 17. `WithInsecureSkipVerify` 命名误导

**文件**: `config/config.go:56-60`

函数名暗示设置 Transport，但实际控制的是 TLS 证书验证：

```go
func WithInsecureSkipVerify(skip bool) Option {
    return func(c *Config) {
        c.InsecureSkipVerify = skip
    }
}
```

**建议**: 改为 `WithInsecureSkipVerify(skip bool)` 或 `WithTLSSkipVerify(skip bool)`。

### 18. `WithMaxIdleConns` 拼写遗漏

**文件**: `config/config.go:99-104`

函数名 `WithMaxIdleConns` 缺少 `le`，应为 `WithMaxIdleConns`：

```go
func WithMaxIdleConns(maxIdleConns int) Option {  // 缺少 "le"
```

**建议**: 修正为 `WithMaxIdleConns`。若考虑向后兼容，可保留旧名称并标记 `// Deprecated`。

### 19. `AuthFlush` 注释拼写错误

**文件**: `builder/bulk.go:37`

```go
// AuthFlush 设置自动刷新大小
func (b *BulkBuilder) AutoFlushSize(size int) *BulkBuilder {
```

注释写了 `AuthFlush` 但方法名是 `AutoFlushSize`，注释应与方法名一致。

### 20. `WithSubProperties` 参数名 `fileType` 应为 `fieldType`

**文件**: `builder/index.go:382`

```go
func WithSubProperties(name string, fileType string, ...) PropertyOption {
```

`fileType` 是拼写错误，应为 `fieldType`。

---

## 五、性能优化

### 21. `SearchResponse.Scan()` 存在多次 Marshal/Unmarshal

**文件**: `builder/search.go:441-486`

对每一条搜索结果，先 `json.Marshal(hit.Source)` 再 `json.Unmarshal(jsonData, elem)`，这是双重序列化开销。

**建议**: 由于 `hit.Source` 已经是 `map[string]any`，可考虑使用 `mapstructure` 或类似库直接映射，或一次性对整个 hits 数组进行序列化/反序列化。

```go
// 优化方案：一次性序列化所有 source
sources := make([]map[string]any, 0, len(r.Hits.Hits))
for _, hit := range r.Hits.Hits {
    if hit.Source != nil {
        sources = append(sources, hit.Source)
    }
}
jsonData, err := json.Marshal(sources)
if err != nil { return err }
return json.Unmarshal(jsonData, dest)
```

### 22. `BulkBuilder.Build()` 可预分配 Buffer

**文件**: `builder/bulk.go:424-447`

当前每次 Build 都从零容量的 Buffer 开始，对于大批量操作会产生多次扩容：

**建议**: 根据 operations 数量预估大小：

```go
func (b *BulkBuilder) Build() []byte {
    b.commitCurrent()
    // 预估每个操作约 200 字节
    var buf bytes.Buffer
    buf.Grow(len(b.operations) * 200)
    // ...
}
```

### 23. `DocumentBuilder.Model()` 的双重序列化

**文件**: `builder/document.go:49-58`

`Model()` 先 `json.Marshal(model)` 再 `json.Unmarshal(jsonData, &b.doc)`，将 struct 转为 map。

**建议**: 如果可以接受发送原始 struct，则直接使用 `model` 作为请求体传递给 `client.Do()`，跳过 map 转换。

---

## 六、功能增强建议

### 24. 缺少自定义 HTTP Header 支持

无法设置自定义 HTTP 头（如 `X-Request-ID`、`X-Opaque-Id` 等 ES 支持的追踪头）。

**建议**: 在 Config 中增加 `DefaultHeaders map[string]string`，在 Do() 中统一设置。

### 25. 缺少请求/响应中间件（Hooks）

无法在请求发送前或响应接收后执行自定义逻辑（如指标采集、链路追踪、请求日志等）。

**建议**: 引入中间件机制：

```go
type RequestHook func(ctx context.Context, req *http.Request) *http.Request
type ResponseHook func(ctx context.Context, resp *http.Response, err error)

type Config struct {
    // ...
    RequestHooks  []RequestHook
    ResponseHooks []ResponseHook
}
```

### 26. 缺少 Context 超时与取消的内部使用

在 `AutoMigrate`（`client/migrate.go:18`）中使用了 `context.Background()`，外部无法控制其超时/取消行为：

```go
func (c *Client) AutoMigrate(models ...any) error {
    // ...
    ctx := context.Background() // 应该由调用方传入
}
```

**建议**: `AutoMigrate` 方法签名应接受 `ctx context.Context` 参数。

### 27. 缺少 Reindex API 支持

ES 的 `_reindex` API 在索引迁移、数据重建等场景非常常用，当前项目未提供。

### 28. 缺少 Index Template / Component Template 管理

生产环境中索引模板（Index Template）是必备功能，当前项目未提供创建/管理模板的能力。

---

## 七、总结

| 类别 | 数量 | 优先级 |
|------|------|--------|
| Bug 级别 | 6 | P0 - 立即修复 |
| 架构设计 | 5 | P1 - 近期优化 |
| 健壮性 | 5 | P1 - 近期优化 |
| 命名与 API 规范 | 4 | P2 - 常规修复 |
| 性能优化 | 3 | P2 - 按需优化 |
| 功能增强 | 5 | P3 - 版本迭代 |

项目整体代码质量不错，Builder 模式的 API 设计清晰一致，文档和测试覆盖较全面。以上建议主要集中在 **运行时安全性**（重试 bug、panic、错误吞没）和 **架构可扩展性**（负载均衡、中间件、日志接口）两个方面。建议按照优先级由高到低逐步落实。
