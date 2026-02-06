# 索引管理 (IndexBuilder)

IndexBuilder 提供了完整的 Elasticsearch 索引管理功能，包括创建、更新、查询和删除索引。

## AutoMigrate（推荐）

类似 GORM 的自动迁移功能，通过结构体标签自动创建或更新索引映射。

### 基本用法

```go
// 定义结构体，使用 es 标签定义字段映射
type Article struct {
    Title     string    `es:"type:text;analyzer:ik_smart"`
    Content   string    `es:"type:text;analyzer:ik_max_word"`
    Author    string    `es:"type:keyword"`
    Tags      []string  `es:"type:keyword"`
    Views     int       `es:"type:integer"`
    Price     float64   `es:"type:float"`
    Published bool      `es:"type:boolean"`
    CreatedAt time.Time `es:"type:date;format:yyyy-MM-dd HH:mm:ss"`
    Location  string    `es:"type:geo_point"`
}

// 自动迁移（创建或更新索引）
err := esClient.AutoMigrate(&Article{})
```

### 自定义索引名

默认索引名为结构体名称的 snake_case 复数形式（如 `Article` -> `articles`）。

可以实现 `IndexName` 接口自定义索引名：

```go
type Article struct {
    // ...
}

// 实现 IndexName 接口
func (a *Article) IndexName() string {
    return "my_articles"
}

// 迁移时会使用 "my_articles" 作为索引名
err := esClient.AutoMigrate(&Article{})
```

### es 标签语法

标签格式：`es:"key1:value1;key2:value2"`

常用配置：
- `type:xxx` - 字段类型（必填）
- `analyzer:xxx` - 分析器
- `format:xxx` - 日期格式
- `-` - 忽略该字段

```go
type Product struct {
    Name        string  `es:"type:text;analyzer:ik_smart"`     // text + IK分词
    SKU         string  `es:"type:keyword"`                     // 精确匹配
    Price       float64 `es:"type:float"`                       // 浮点数
    Stock       int     `es:"type:integer"`                     // 整数
    Available   bool    `es:"type:boolean"`                     // 布尔
    CreatedAt   string  `es:"type:date;format:yyyy-MM-dd"`      // 日期
    Location    string  `es:"type:geo_point"`                   // 地理位置
    InternalID  string  `es:"-"`                                // 忽略
}
```

### 迁移多个模型

```go
err := esClient.AutoMigrate(
    &Article{},
    &Product{},
    &User{},
)
```

### 注意事项

- ES 不支持删除或修改已有字段类型，AutoMigrate 只能添加新字段
- 如果索引已存在，会调用 PutMapping 添加新字段
- 如果索引不存在，会创建新索引

## 基础操作

### 创建索引

```go
// 基础创建
err := builder.NewIndexBuilder(esClient, "products").
    Shards(1).
    Replicas(0).
    RefreshInterval("1s").
    AddProperty("name", "text", builder.WithAnalyzer("ik_smart")).
    AddProperty("price", "float").
    AddProperty("category", "keyword").
    AddProperty("created_at", "date", builder.WithFormat("yyyy-MM-dd HH:mm:ss")).
    AddAlias("products-alias", nil).
    Create(ctx)
```

### 更新索引设置

```go
err := builder.NewIndexBuilder(esClient, "products").
    Replicas(2).
    RefreshInterval("30s").
    UpdateSettings(ctx)
```

### 更新索引映射

```go
// 添加新字段
err := builder.NewIndexBuilder(esClient, "products").
    AddProperty("description", "text", builder.WithAnalyzer("ik_max_word")).
    AddProperty("stock", "integer").
    PutMapping(ctx)
```

### 检查索引是否存在

```go
exists, err := builder.NewIndexBuilder(esClient, "products").Exists(ctx)
```

### 获取索引信息

```go
info, err := builder.NewIndexBuilder(esClient, "products").Get(ctx)
fmt.Println(info.PrettyJSON())
```

### 删除索引

```go
err := builder.NewIndexBuilder(esClient, "products").Delete(ctx)
```

## 自定义分析器

### 方式1：简化版（基于 tokenizer 快速创建）

```go
err := builder.NewIndexBuilder(esClient, "articles").
    Shards(1).
    Replicas(1).
    // 基于 tokenizer 快速创建（不忽略大小写的 ik_smart）
    AddCustomAnalyzer("ik_case_sensitive", builder.TokenizerIKSmart).
    AddProperty("title", "text", builder.WithAnalyzer("ik_case_sensitive")).
    Create(ctx)
```

### 方式2：完整版（使用 Option 模式自定义配置）

```go
err := builder.NewIndexBuilder(esClient, "articles").
    AddAnalyzer("html_ik_analyzer",
        builder.WithAnalyzerType(builder.AnalyzerTypeCustom),
        builder.WithCharFilters(builder.CharFilterHTMLStrip),  // 去除HTML标签
        builder.WithTokenizer(builder.TokenizerIKSmart),       // IK分词
        builder.WithTokenFilters(builder.TokenFilterLowercase), // 转小写
    ).
    AddProperty("content", "text", builder.WithAnalyzer("html_ik_analyzer")).
    Create(ctx)
```

### 常用分析器示例

```go
err := builder.NewIndexBuilder(esClient, "posts").
    // IK 智能分词（保持大小写）
    AddCustomAnalyzer("ik_case_sensitive", builder.TokenizerIKSmart).

    // IK 最大词数分词（转小写）
    AddCustomAnalyzer("ik_lowercase", builder.TokenizerIKMaxWord,
        builder.TokenFilterLowercase).

    // 标准分词器 + 词干提取 + 停用词过滤
    AddAnalyzer("english_stemmed",
        builder.WithAnalyzerType(builder.AnalyzerTypeCustom),
        builder.WithTokenizer(builder.TokenizerStandard),
        builder.WithTokenFilters(
            builder.TokenFilterLowercase,
            builder.TokenFilterStop,
            builder.TokenFilterPorterStem,
        ),
    ).
    Create(ctx)
```

## 分析器常量说明

所有可用的常量定义在 `builder/analyzer_constants.go` 中：

### 分词器常量 (Tokenizers)

```go
builder.TokenizerIKSmart      // IK智能分词（粗粒度）
builder.TokenizerIKMaxWord    // IK最大词数分词（细粒度）
builder.TokenizerStandard     // 标准分词器
builder.TokenizerWhitespace   // 空格分词器
builder.TokenizerKeyword      // 关键词分词器（不分词）
// 更多见 analyzer_constants.go
```

### 字符过滤器常量 (Char Filters)

```go
builder.CharFilterHTMLStrip   // 去除HTML标签
builder.CharFilterMapping     // 字符映射替换
builder.CharFilterPattern     // 正则替换
```

### Token过滤器常量 (Token Filters)

```go
builder.TokenFilterLowercase  // 转小写
builder.TokenFilterUppercase  // 转大写
builder.TokenFilterStop       // 去除停用词
builder.TokenFilterStemmer    // 词干提取
builder.TokenFilterSynonym    // 同义词替换
builder.TokenFilterUnique     // 去重
// 更多见 analyzer_constants.go
```

### 分析器类型常量 (Analyzer Types)

```go
builder.AnalyzerTypeCustom    // 自定义分析器
builder.AnalyzerTypeStandard  // 标准分析器
builder.AnalyzerIKSmart       // IK智能分词器（内置）
builder.AnalyzerIKMaxWord     // IK最大词数分词器（内置）
```

## 字段类型常量说明

所有可用的字段类型常量定义在 `builder/field_types.go` 中：

### 字符串类型

```go
builder.FieldTypeText      // 全文搜索字段（会分词）
builder.FieldTypeKeyword   // 精确匹配字段（不分词）
```

### 数值类型

```go
builder.FieldTypeInt       // 32位整数
builder.FieldTypeLong      // 64位整数
builder.FieldTypeFloat     // 32位浮点数
builder.FieldTypeDouble    // 64位双精度浮点数
builder.FieldTypeByte      // 8位整数
builder.FieldTypeShort     // 16位整数
```

### 日期/布尔类型

```go
builder.FieldTypeDate      // 日期类型
builder.FieldTypeBoolean   // 布尔类型
```

### 地理位置类型

```go
builder.FieldTypeGeoPoint  // 地理位置点
builder.FieldTypeGeoShape  // 地理形状
```

### 复杂类型

```go
builder.FieldTypeObject    // 对象类型
builder.FieldTypeNested    // 嵌套类型
```

### 其他类型

```go
builder.FieldTypeIP              // IP地址
builder.FieldTypeBinary          // 二进制
builder.FieldTypeCompletion      // 自动补全
builder.FieldTypeDenseVector     // 密集向量（向量搜索）
// 更多见 field_types.go
```

### 使用示例（使用常量避免拼写错误）

```go
err := builder.NewIndexBuilder(esClient, "products").
    // 使用字段类型常量，IDE 会自动提示，避免拼写错误
    AddProperty("name", builder.FieldTypeText, builder.WithAnalyzer(builder.AnalyzerIKSmart)).
    AddProperty("sku", builder.FieldTypeKeyword).
    AddProperty("price", builder.FieldTypeFloat).
    AddProperty("quantity", builder.FieldTypeInt).
    AddProperty("available", builder.FieldTypeBoolean).
    AddProperty("created_at", builder.FieldTypeDate, builder.WithFormat("yyyy-MM-dd HH:mm:ss")).
    AddProperty("location", builder.FieldTypeGeoPoint).
    Create(ctx)
```

## 支持的功能

- ✅ AutoMigrate（自动迁移）
- ✅ 创建索引 (Create, Shards, Replicas, RefreshInterval)
- ✅ 字段映射 (AddProperty, WithAnalyzer, WithFormat)
- ✅ 别名管理 (AddAlias)
- ✅ 更新索引设置 (UpdateSettings)
- ✅ 更新索引映射 (PutMapping)
- ✅ 检查存在 (Exists)
- ✅ 获取索引信息 (Get)
- ✅ 删除索引 (Delete)
