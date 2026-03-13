# Requirements: go-es Quality Fixes & Feature Enhancements

**Defined:** 2026-03-13
**Core Value:** 修复后的库在生产环境下可靠、安全运行——重试不丢数据，错误不被吞没，节点故障自动转移。

## v1 Requirements

Requirements for this milestone. Each maps to roadmap phases.

### Correctness & Reliability

- [x] **RELY-01**: Do() 和 DoRequest() 重试时正确重置请求体（使用 io.Seeker），不再发送空请求
- [x] **RELY-02**: Ping() 尝试所有配置的地址，而非仅 addresses[0]
- [x] **RELY-03**: Exists() 区分 404（返回 false, nil）和真实错误（返回 false, error），保持函数签名不变
- [x] **RELY-04**: BulkBuilder AutoFlush 使用用户传入的 context 而非 context.TODO()
- [x] **RELY-05**: BulkBuilder Set 方法链中的错误通过 Do() 返回值传递给调用者，文档明确说明行为
- [x] **RELY-06**: 所有 error 类型断言统一使用 errors.As() 替代直接类型断言，支持 wrapped errors

### Security

- [x] **SECU-01**: 所有 URL 路径组件（index name, document ID, scroll ID 等）使用 url.PathEscape() 编码

### Code Quality

- [x] **QUAL-01**: toSnakeCase() 正确处理连续大写字母（如 HTTPServer → http_server）
- [x] **QUAL-02**: AutoMigrate 正确处理 []Struct 类型字段标记为 es:"type:nested" 的情况

### Performance

- [ ] **PERF-01**: SearchResponse.Scan() 避免双重序列化，直接从 map[string]any 转换到目标结构体
- [ ] **PERF-02**: DocumentBuilder.Model() 避免 struct→JSON→map 双重序列化
- [ ] **PERF-03**: structToMap() 使用直接反射替代 JSON 序列化中间步骤
- [ ] **PERF-04**: BulkBuilder.Build() 根据 operations 数量预分配缓冲区

### Node Pool & Retry

- [ ] **POOL-01**: 实现连接池，跟踪每个节点的 dead/live 状态和失败次数
- [ ] **POOL-02**: Dead 节点支持指数退避自动恢复检测
- [ ] **POOL-03**: 重试逻辑升级为指数退避 + 随机抖动（full jitter）
- [ ] **POOL-04**: 重试范围扩展到 HTTP 429, 502, 503, 504 状态码

### Circuit Breaker

- [ ] **BRKR-01**: 实现 3 状态熔断器（Closed/Open/Half-Open），零外部依赖
- [ ] **BRKR-02**: 熔断器正确分类错误：context 取消和 4xx 不触发熔断
- [ ] **BRKR-03**: 通过 config.Option 函数配置熔断器参数（阈值、超时、半开请求数）

### Middleware

- [ ] **MIDW-01**: 实现请求/响应中间件链，支持用户注入自定义逻辑（logging、metrics、tracing）
- [ ] **MIDW-02**: 中间件通过 config.Option 配置，不改变 ESClient 接口
- [ ] **MIDW-03**: 提供至少一个示例中间件（如 request logging middleware）

### Cache

- [ ] **CACH-01**: 实现基于接口的查询缓存层，默认使用 sync.Map + TTL 实现
- [ ] **CACH-02**: 仅缓存读操作（GET, _search POST），写操作自动失效相关缓存
- [ ] **CACH-03**: 通过 config.Option 配置缓存开关、TTL、最大条目数

### Index Aliases

- [ ] **ALIA-01**: AutoMigrate 支持创建和管理 Index Aliases
- [ ] **ALIA-02**: IndexBuilder 提供 Alias 操作方法（Add, Remove, Switch）

### Test Coverage

- [ ] **TEST-01**: 重试逻辑的单元测试（模拟网络错误、请求体重置验证）
- [ ] **TEST-02**: 错误路径测试（无效输入、空字符串、特殊字符 ID/index name）
- [ ] **TEST-03**: URL 编码边界测试（含 /, ?, #, 空格, 非 ASCII 字符的 index/ID）
- [ ] **TEST-04**: AutoMigrate 边界测试（多层嵌套、Slice of Struct、指针、匿名嵌入）
- [ ] **TEST-05**: 节点故障转移测试（节点宕机、恢复、全部宕机场景）
- [ ] **TEST-06**: 熔断器状态转换测试（Closed→Open→Half-Open→Closed 完整流转）
- [ ] **TEST-07**: 中间件链执行顺序和错误处理测试
- [ ] **TEST-08**: 缓存命中/未命中/失效测试

## v2 Requirements

Deferred to future release. Tracked but not in current roadmap.

### Extended Auth

- **AUTH-01**: OAuth/SAML/Token 认证支持
- **AUTH-02**: API Key 认证支持

### Observability

- **OBSV-01**: 内置 Prometheus metrics 中间件
- **OBSV-02**: OpenTelemetry tracing 中间件

### Advanced Features

- **ADVF-01**: 节点自动发现（Sniffing）
- **ADVF-02**: 并发 Scroll 操作限制器
- **ADVF-03**: Bulk 流式模式（streaming mode）

## Out of Scope

| Feature | Reason |
|---------|--------|
| 第三方 JSON 库替换 | stdlib encoding/json 够用，替换影响面太大 |
| gRPC 支持 | Elasticsearch 主要用 REST API |
| 并发 Builder 安全 | Builder 设计上是非线程安全的，这是设计决策 |
| ESClient 接口修改 | 破坏向后兼容，所有新功能通过 config.Option 注入 |
| 外部依赖引入 | 保持零依赖特性，所有功能用 stdlib 实现 |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| RELY-01 | Phase 1 | Complete |
| RELY-02 | Phase 1 | Complete |
| RELY-03 | Phase 1 | Complete |
| RELY-04 | Phase 1 | Complete |
| RELY-05 | Phase 1 | Complete |
| RELY-06 | Phase 1 | Complete |
| SECU-01 | Phase 1 | Complete |
| QUAL-01 | Phase 1 | Complete |
| QUAL-02 | Phase 1 | Complete |
| PERF-01 | Phase 5 | Pending |
| PERF-02 | Phase 5 | Pending |
| PERF-03 | Phase 5 | Pending |
| PERF-04 | Phase 5 | Pending |
| POOL-01 | Phase 2 | Pending |
| POOL-02 | Phase 2 | Pending |
| POOL-03 | Phase 2 | Pending |
| POOL-04 | Phase 2 | Pending |
| BRKR-01 | Phase 3 | Pending |
| BRKR-02 | Phase 3 | Pending |
| BRKR-03 | Phase 3 | Pending |
| MIDW-01 | Phase 4 | Pending |
| MIDW-02 | Phase 4 | Pending |
| MIDW-03 | Phase 4 | Pending |
| CACH-01 | Phase 6 | Pending |
| CACH-02 | Phase 6 | Pending |
| CACH-03 | Phase 6 | Pending |
| ALIA-01 | Phase 6 | Pending |
| ALIA-02 | Phase 6 | Pending |
| TEST-01 | Phase 2 | Pending |
| TEST-02 | Phase 1 | Pending |
| TEST-03 | Phase 1 | Pending |
| TEST-04 | Phase 1 | Pending |
| TEST-05 | Phase 2 | Pending |
| TEST-06 | Phase 3 | Pending |
| TEST-07 | Phase 4 | Pending |
| TEST-08 | Phase 6 | Pending |

**Coverage:**
- v1 requirements: 36 total
- Mapped to phases: 36
- Unmapped: 0

---
*Requirements defined: 2026-03-13*
*Last updated: 2026-03-13 after roadmap creation*
